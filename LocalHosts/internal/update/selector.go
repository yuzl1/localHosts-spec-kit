package update

import "fmt"

func SelectBestAsset(release GitHubRelease, goos string, goarch string) (GitHubAsset, error) {
	candidates := []string{}
	switch goos {
	case "darwin":
		candidates = []string{
			"LocalHosts-macos-universal.zip",
			"LocalHosts-macos-universal.dmg",
		}
	case "windows":
		switch goarch {
		case "amd64":
			candidates = []string{"LocalHosts-windows-amd64-setup.exe"}
		case "arm64":
			candidates = []string{"LocalHosts-windows-arm64-setup.exe"}
		default:
			return GitHubAsset{}, fmt.Errorf("UPDATE_UNSUPPORTED_PLATFORM: %s/%s", goos, goarch)
		}
	default:
		return GitHubAsset{}, fmt.Errorf("UPDATE_UNSUPPORTED_PLATFORM: %s/%s", goos, goarch)
	}

	for _, name := range candidates {
		for _, a := range release.Assets {
			if a.Name == name && a.BrowserDownloadURL != "" {
				return a, nil
			}
		}
	}
	return GitHubAsset{}, fmt.Errorf("UPDATE_ASSET_NOT_FOUND")
}
