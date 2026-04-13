package workspace

import (
	"encoding/json"
	"os"
	"time"

	"LocalHosts/internal/hosts"
	"LocalHosts/internal/platform"
)

func LoadOrInit() (hosts.WorkspaceState, string, bool, error) {
	appDir, err := platform.UserAppDir()
	if err != nil {
		return hosts.WorkspaceState{}, "", false, err
	}

	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return hosts.WorkspaceState{}, "", false, err
	}

	path := platform.WorkspacePath(appDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			state := defaultState()
			return state, path, false, nil
		}
		return hosts.WorkspaceState{}, "", false, err
	}

	var state hosts.WorkspaceState
	if err := json.Unmarshal(data, &state); err != nil {
		state = defaultState()
		return state, path, true, nil
	}

	state = normalizeState(state)
	return state, path, true, nil
}

func Save(state hosts.WorkspaceState) (string, error) {
	appDir, err := platform.UserAppDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}

	state = normalizeState(state)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", err
	}

	path := platform.WorkspacePath(appDir)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}

	return path, nil
}

func defaultState() hosts.WorkspaceState {
	return hosts.WorkspaceState{
		Version:            2,
		ComposeEnabled:     true,
		Theme:              "auto",
		AutoCheckOnStartup: true,
		Groups: []hosts.HostGroup{
			{
				ID:      "default",
				Name:    "Default",
				Enabled: true,
				Order:   0,
			},
		},
		Entries:   []hosts.HostEntry{},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func normalizeState(state hosts.WorkspaceState) hosts.WorkspaceState {
	if state.Version == 0 {
		state.Version = 1
	}

	if state.Version < 2 {
		state.AutoCheckOnStartup = true
		state.Version = 2
	}

	if state.Theme == "" {
		state.Theme = "auto"
	}

	if len(state.Groups) == 0 {
		state.Groups = defaultState().Groups
	}

	if state.UpdatedAt == "" {
		state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return state
}
