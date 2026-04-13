package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

func FetchLatestGitHubRelease(ctx context.Context, owner string, repo string) (GitHubRelease, error) {
	return FetchLatestGitHubReleaseWithProxy(ctx, owner, repo, "")
}

func FetchLatestGitHubReleaseWithProxy(ctx context.Context, owner string, repo string, proxyPrefix string) (GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	url = applyProxyPrefix(url, proxyPrefix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GitHubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "LocalHosts")

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("UPDATE_CHECK_FAILED: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(resp.Status)
		if msg == "" {
			msg = fmt.Sprintf("http status %d", resp.StatusCode)
		}
		return GitHubRelease{}, fmt.Errorf("UPDATE_CHECK_FAILED: %s", msg)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubRelease{}, err
	}
	var out GitHubRelease
	if err := json.Unmarshal(raw, &out); err != nil {
		return GitHubRelease{}, fmt.Errorf("UPDATE_CHECK_FAILED: %s", string(raw))
	}
	return out, nil
}

func applyProxyPrefix(url string, proxyPrefix string) string {
	p := strings.TrimSpace(proxyPrefix)
	if p == "" {
		return url
	}
	if strings.Contains(p, "{url}") {
		return strings.ReplaceAll(p, "{url}", url)
	}
	return p + url
}
