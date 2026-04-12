package hosts

type HostEntry struct {
	ID      string `json:"id"`
	IP      string `json:"ip"`
	Domain  string `json:"domain"`
	Comment string `json:"comment"`
	Enabled bool   `json:"enabled"`
	GroupID string `json:"groupId"`
	Source  string `json:"source"`
	Raw     string `json:"raw"`
}

type HostGroup struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Order   int    `json:"order"`
}

type WorkspaceState struct {
	Version        int         `json:"version"`
	ComposeEnabled bool        `json:"composeEnabled"`
	Theme          string      `json:"theme"`
	Groups         []HostGroup `json:"groups"`
	Entries        []HostEntry `json:"entries"`
	UpdatedAt      string      `json:"updatedAt"`
}

type HostWorkspace struct {
	ComposeEnabled      bool        `json:"composeEnabled"`
	Theme               string      `json:"theme"`
	Groups              []HostGroup `json:"groups"`
	Entries             []HostEntry `json:"entries"`
	ReadonlySystemLines []string    `json:"readonlySystemLines"`
	SystemHostsLines    []string    `json:"systemHostsLines"`
}

type HostDocument struct {
	PrefixLines []string
	SuffixLines []string
	EOL         string
	EndsWithEOL bool
}
