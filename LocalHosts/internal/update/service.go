package update

import (
	"context"
	goruntime "runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const ProgressEventName = "update:progress"

func EmitProgress(ctx context.Context, p Progress) {
	wailsruntime.EventsEmit(ctx, ProgressEventName, p)
}

func CheckForUpdate(ctx context.Context, owner string, repo string, currentVersion string) (Info, error) {
	cv, err := NormalizeSemverString(currentVersion)
	if err != nil {
		return Info{}, err
	}

	EmitProgress(ctx, Progress{Stage: "checking", Message: "checking for updates"})
	rel, err := FetchLatestGitHubRelease(ctx, owner, repo)
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "check failed"})
		return Info{}, err
	}

	lv, err := NormalizeSemverString(strings.TrimSpace(rel.TagName))
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "invalid latest version"})
		return Info{}, err
	}

	hasUpdate := false
	a, err := ParseSemver(cv)
	if err != nil {
		return Info{}, err
	}
	b, err := ParseSemver(lv)
	if err != nil {
		return Info{}, err
	}
	if CompareSemver(a, b) < 0 {
		hasUpdate = true
	}

	asset, err := SelectBestAsset(rel, goruntime.GOOS, goruntime.GOARCH)
	if err != nil {
		if hasUpdate {
			EmitProgress(ctx, Progress{Stage: "error", Message: "no installer found"})
			return Info{}, err
		}
		EmitProgress(ctx, Progress{Stage: "idle", Message: "up to date"})
		return Info{
			CurrentVersion: cv,
			LatestVersion:  lv,
			HasUpdate:      false,
			ReleaseNotes:   strings.TrimSpace(rel.Body),
			PublishedAt:    strings.TrimSpace(rel.PublishedAt),
		}, nil
	}

	EmitProgress(ctx, Progress{Stage: "available", Message: "update checked"})
	return Info{
		CurrentVersion: cv,
		LatestVersion:  lv,
		HasUpdate:      hasUpdate,
		ReleaseNotes:   strings.TrimSpace(rel.Body),
		PublishedAt:    strings.TrimSpace(rel.PublishedAt),
		AssetName:      asset.Name,
		AssetURL:       asset.BrowserDownloadURL,
		AssetSize:      asset.Size,
	}, nil
}
