package update

type Settings struct {
	AutoCheckOnStartup bool `json:"autoCheckOnStartup"`
}

type Info struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseNotes   string `json:"releaseNotes"`
	PublishedAt    string `json:"publishedAt"`
	AssetName      string `json:"assetName"`
	AssetURL       string `json:"assetURL"`
	AssetSize      int64  `json:"assetSize"`
}

type Progress struct {
	Stage           string `json:"stage"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Percent         int    `json:"percent"`
	Message         string `json:"message"`
}
