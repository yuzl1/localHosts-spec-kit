package hosts

import (
	"net"
	"strings"

	"github.com/google/uuid"
)

func ParseManagedBlockWithGroups(blockLines []string) ([]HostGroup, []HostEntry) {
	groupsByID := map[string]HostGroup{}
	groupsOrder := []string{}
	currentGroupID := "default"

	ensureGroup := func(id, name string) {
		if _, ok := groupsByID[id]; ok {
			return
		}
		groupsByID[id] = HostGroup{
			ID:      id,
			Name:    name,
			Enabled: true,
			Order:   len(groupsOrder),
		}
		groupsOrder = append(groupsOrder, id)
	}

	ensureGroup("default", "Default")

	var entries []HostEntry

	for _, line := range blockLines {
		if line == StartMarker || line == EndMarker || line == HeaderMarker {
			continue
		}

		if name, ok := parseGroupHeader(line); ok {
			currentGroupID = "group:" + name
			ensureGroup(currentGroupID, name)
			continue
		}

		if name, ok := parseGroupStartMarker(line); ok {
			currentGroupID = "group:" + name
			ensureGroup(currentGroupID, name)
			continue
		}

		if _, ok := parseGroupEndMarker(line); ok {
			currentGroupID = "default"
			continue
		}

		parsed := ParseEntryLine(line)
		for i := range parsed {
			parsed[i].GroupID = currentGroupID
		}
		entries = append(entries, parsed...)
	}

	var groups []HostGroup
	for _, id := range groupsOrder {
		groups = append(groups, groupsByID[id])
	}

	if len(groups) == 0 {
		groups = append(groups, HostGroup{
			ID:      "default",
			Name:    "Default",
			Enabled: true,
			Order:   0,
		})
	}

	return groups, entries
}

func ParseManagedBlock(blockLines []string) []HostEntry {
	var entries []HostEntry

	for _, line := range blockLines {
		if line == StartMarker || line == EndMarker || line == HeaderMarker {
			continue
		}
		parsed := ParseEntryLine(line)
		entries = append(entries, parsed...)
	}

	return entries
}

func parseGroupStartMarker(line string) (string, bool) {
	const prefix = "#######"
	const suffix = "开始####"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix))
	if name == "" {
		return "", false
	}
	return name, true
}

func parseGroupEndMarker(line string) (string, bool) {
	const prefix = "#######"
	const suffix = "结束####"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix))
	if name == "" {
		return "", false
	}
	return name, true
}

func parseGroupHeader(line string) (string, bool) {
	s := strings.TrimSpace(line)
	const prefix = "# --- "
	const suffix = " ---"
	if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, suffix) {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, prefix), suffix))
	if name == "" {
		return "", false
	}
	return name, true
}

func ParseEntryLine(line string) []HostEntry {
	original := line
	trimmedLeft := strings.TrimLeft(line, " \t")

	enabled := true
	if strings.HasPrefix(trimmedLeft, "#") {
		enabled = false
		trimmedLeft = strings.TrimLeft(strings.TrimPrefix(trimmedLeft, "#"), " \t")
	}

	body, comment := splitInlineComment(trimmedLeft)
	fields := strings.Fields(body)
	if len(fields) < 2 {
		return []HostEntry{}
	}

	ip := fields[0]
	if net.ParseIP(ip) == nil {
		return []HostEntry{}
	}

	domains := fields[1:]
	var entries []HostEntry
	for _, domain := range domains {
		if domain == "" || strings.Contains(domain, "#") {
			continue
		}
		entries = append(entries, HostEntry{
			ID:      uuid.NewString(),
			IP:      ip,
			Domain:  domain,
			Comment: comment,
			Enabled: enabled,
			Source:  "managed-block",
			Raw:     original,
		})
	}

	return entries
}

func splitInlineComment(s string) (string, string) {
	index := strings.Index(s, "#")
	if index == -1 {
		return strings.TrimRight(s, " \t"), ""
	}

	body := strings.TrimRight(s[:index], " \t")
	comment := strings.TrimSpace(s[index+1:])
	return body, comment
}
