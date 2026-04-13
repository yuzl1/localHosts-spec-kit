package update

import "testing"

func TestSelectBestAssetWindowsAMD64(t *testing.T) {
	rel := GitHubRelease{
		Assets: []GitHubAsset{
			{Name: "LocalHosts-windows-amd64-setup.exe", BrowserDownloadURL: "https://example.com/a", Size: 1},
		},
	}
	a, err := SelectBestAsset(rel, "windows", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "LocalHosts-windows-amd64-setup.exe" {
		t.Fatalf("got %q", a.Name)
	}
}

func TestSelectBestAssetMacOSUniversalPrefersZip(t *testing.T) {
	rel := GitHubRelease{
		Assets: []GitHubAsset{
			{Name: "LocalHosts-macos-universal.dmg", BrowserDownloadURL: "https://example.com/dmg", Size: 2},
			{Name: "LocalHosts-macos-universal.zip", BrowserDownloadURL: "https://example.com/zip", Size: 3},
		},
	}
	a, err := SelectBestAsset(rel, "darwin", "arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "LocalHosts-macos-universal.zip" {
		t.Fatalf("got %q", a.Name)
	}
}
