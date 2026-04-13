package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
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
	return FetchLatestGitHubReleaseWithProxy(ctx, owner, repo, "", "", 0)
}

func FetchLatestGitHubReleaseWithProxy(ctx context.Context, owner string, repo string, proxyType string, proxyHost string, proxyPort int) (GitHubRelease, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return GitHubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "LocalHosts")

	client := http.Client{Timeout: 15 * time.Second, Transport: newTransport(proxyType, proxyHost, proxyPort)}
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

func newTransport(proxyType string, proxyHost string, proxyPort int) *http.Transport {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	typ := strings.ToLower(strings.TrimSpace(proxyType))
	host := strings.TrimSpace(proxyHost)
	if host == "" || proxyPort <= 0 {
		return t
	}
	port := strconv.Itoa(proxyPort)

	switch typ {
	case "http", "https":
		scheme := typ
		if scheme == "" {
			scheme = "http"
		}
		if pu, err := url.Parse(scheme + "://" + net.JoinHostPort(host, port)); err == nil {
			t.Proxy = http.ProxyURL(pu)
		}
		return t
	case "socks", "socks5":
		dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(host, port), nil, proxy.Direct)
		if err != nil {
			return t
		}
		if cd, ok := dialer.(proxy.ContextDialer); ok {
			t.Proxy = nil
			t.DialContext = cd.DialContext
			return t
		}
		t.Proxy = nil
		t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		return t
	default:
		return t
	}
}
