package hosts

import (
	"sort"
	"strings"
)

func RenderManagedBlock(workspace HostWorkspace) []string {
	lines := []string{StartMarker, HeaderMarker}

	groups := includedGroups(workspace)
	if len(groups) == 0 {
		lines = append(lines, EndMarker)
		return lines
	}

	for _, group := range groups {
		entries := entriesForGroup(workspace.Entries, group.ID)
		if len(entries) == 0 {
			continue
		}
		lines = append(lines, groupStartMarker(group.Name))
		for _, entry := range entries {
			lines = append(lines, renderEntryLine(entry))
		}
		lines = append(lines, groupEndMarker(group.Name))
	}

	lines = append(lines, EndMarker)
	return lines
}

func BuildHostsFile(doc HostDocument, workspace HostWorkspace, hadValidBlock bool) []byte {
	prefix := append([]string{}, doc.PrefixLines...)
	suffix := append([]string{}, doc.SuffixLines...)

	blockLines := RenderManagedBlock(workspace)

	lines := []string{}
	lines = append(lines, prefix...)

	if !hadValidBlock {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
	}

	lines = append(lines, blockLines...)
	lines = append(lines, suffix...)

	content := strings.Join(lines, doc.EOL)
	if doc.EndsWithEOL {
		content += doc.EOL
	}

	return []byte(content)
}

func includedGroups(workspace HostWorkspace) []HostGroup {
	var groups []HostGroup
	for _, group := range workspace.Groups {
		if workspace.ComposeEnabled && !group.Enabled {
			continue
		}
		groups = append(groups, group)
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Order == groups[j].Order {
			return groups[i].Name < groups[j].Name
		}
		return groups[i].Order < groups[j].Order
	})

	return groups
}

func entriesForGroup(entries []HostEntry, groupID string) []HostEntry {
	var out []HostEntry
	for _, entry := range entries {
		if entry.GroupID == groupID {
			out = append(out, entry)
		}
	}
	return out
}

func renderEntryLine(entry HostEntry) string {
	body := entry.IP + " " + entry.Domain
	if entry.Comment != "" {
		body += " # " + entry.Comment
	}

	if entry.Enabled {
		return body
	}
	return "# " + body
}

func groupStartMarker(name string) string {
	return "#######" + name + "开始####"
}

func groupEndMarker(name string) string {
	return "#######" + name + "结束####"
}
