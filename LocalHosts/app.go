package main

import (
	"context"
	"os"
	"strings"

	"LocalHosts/internal/hosts"
	"LocalHosts/internal/platform"
	"LocalHosts/internal/workspace"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetHosts returns the current workspace and readonly system hosts lines.
func (a *App) GetHosts() (hosts.HostWorkspace, error) {
	state, _, _, err := workspace.LoadOrInit()
	if err != nil {
		return hosts.HostWorkspace{}, err
	}

	hostsPath := platform.SystemHostsPath()
	raw, err := readFile(hostsPath)
	if err != nil {
		return hosts.HostWorkspace{
			ComposeEnabled:      state.ComposeEnabled,
			Groups:              state.Groups,
			Entries:             state.Entries,
			ReadonlySystemLines: []string{},
			SystemHostsLines:    []string{},
		}, nil
	}

	doc, blockLines, hasBlock, _ := hosts.SplitDocument(raw)
	systemHostsLines := splitLines(string(raw), doc.EOL, doc.EndsWithEOL)
	readonly := append([]string{}, doc.PrefixLines...)
	readonly = append(readonly, doc.SuffixLines...)

	parsedGroups := []hosts.HostGroup{}
	parsedEntries := []hosts.HostEntry{}
	if hasBlock {
		parsedGroups, parsedEntries = hosts.ParseManagedBlockWithGroups(blockLines)
	}

	if len(state.Entries) == 0 && len(parsedEntries) > 0 {
		if len(parsedGroups) > 0 {
			state.Groups = parsedGroups
		}
		state.Entries = parsedEntries
		_, _ = workspace.Save(state)
	}

	for i := range state.Entries {
		if state.Entries[i].ID == "" {
			state.Entries[i].ID = uuid.NewString()
		}
		if state.Entries[i].GroupID == "" {
			state.Entries[i].GroupID = "default"
		}
	}

	return hosts.HostWorkspace{
		ComposeEnabled:      state.ComposeEnabled,
		Theme:               state.Theme,
		Groups:              state.Groups,
		Entries:             state.Entries,
		ReadonlySystemLines: readonly,
		SystemHostsLines:    systemHostsLines,
	}, nil
}

func (a *App) CanWriteHosts() bool {
	hostsPath := platform.SystemHostsPath()
	f, err := os.OpenFile(hostsPath, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

func (a *App) EnsureWriteAccess() error {
	return platform.RequestAdminAccess()
}

// SaveHosts persists the workspace to the user directory.
func (a *App) SaveHosts(workspaceInput hosts.HostWorkspace) error {
	state := hosts.WorkspaceState{
		Version:        1,
		ComposeEnabled: workspaceInput.ComposeEnabled,
		Theme:          workspaceInput.Theme,
		Groups:         workspaceInput.Groups,
		Entries:        workspaceInput.Entries,
	}

	for i := range state.Entries {
		if state.Entries[i].ID == "" {
			state.Entries[i].ID = uuid.NewString()
		}
		if state.Entries[i].GroupID == "" {
			state.Entries[i].GroupID = "default"
		}
	}

	_, err := workspace.Save(state)
	return err
}

// ApplyHosts writes the composed workspace into the system hosts file.
func (a *App) ApplyHosts(workspaceInput hosts.HostWorkspace) error {
	if err := a.SaveHosts(workspaceInput); err != nil {
		return err
	}

	hostsPath := platform.SystemHostsPath()
	raw, err := readFile(hostsPath)
	if err != nil {
		return err
	}

	doc, _, hasBlock, _ := hosts.SplitDocument(raw)
	content := hosts.BuildHostsFile(doc, workspaceInput, hasBlock)
	return platform.ApplyHostsWithAdmin(hostsPath, content)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func splitLines(raw string, eol string, endsWithEOL bool) []string {
	lines := strings.Split(raw, eol)
	if endsWithEOL && len(lines) > 0 && lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}
