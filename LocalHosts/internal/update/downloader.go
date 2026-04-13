package update

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadInstaller(ctx context.Context, url string, filename string, expectedSize int64, proxyPrefix string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, "LocalHosts", "updates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	destPath := filepath.Join(dir, filename)
	tmpPath := destPath + ".part"
	_ = os.Remove(tmpPath)

	EmitProgress(ctx, Progress{Stage: "downloading", Message: "starting download"})

	url = applyProxyPrefix(url, proxyPrefix)

	transport := &http.Transport{
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
	client := http.Client{Transport: transport}

	var resp *http.Response
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
			return "", err
		}
		req.Header.Set("User-Agent", "LocalHosts")

		resp, err = client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
		if !isTimeoutError(err) {
			break
		}
		time.Sleep(time.Duration(500+attempt*700) * time.Millisecond)
	}
	if resp == nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		if lastErr != nil {
			return "", fmt.Errorf("UPDATE_DOWNLOAD_FAILED: %s", lastErr.Error())
		}
		return "", fmt.Errorf("UPDATE_DOWNLOAD_FAILED")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		return "", fmt.Errorf("UPDATE_DOWNLOAD_FAILED: %s", resp.Status)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = expectedSize
	}

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		return "", err
	}

	defer func() {
		_ = f.Close()
	}()

	buf := make([]byte, 128*1024)
	var downloaded int64
	lastEmit := time.Now()
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
				return "", werr
			}
			downloaded += int64(n)
		}
		if time.Since(lastEmit) >= 200*time.Millisecond || rerr == io.EOF {
			percent := 0
			if total > 0 {
				p := int(downloaded * 100 / total)
				if p > 100 {
					p = 100
				}
				percent = p
			}
			EmitProgress(ctx, Progress{
				Stage:           "downloading",
				DownloadedBytes: downloaded,
				TotalBytes:      total,
				Percent:         percent,
				Message:         "downloading",
			})
			lastEmit = time.Now()
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
			return "", rerr
		}
	}

	if err := f.Sync(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		return "", err
	}
	if err := f.Close(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		return "", err
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "download failed"})
		return "", err
	}

	EmitProgress(ctx, Progress{Stage: "downloaded", DownloadedBytes: downloaded, TotalBytes: total, Percent: 100, Message: "downloaded"})
	return destPath, nil
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return true
	}
	return false
}
