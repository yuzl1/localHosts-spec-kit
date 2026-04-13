//go:build darwin

package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "preparing update"})

	appPath, err := currentAppBundlePath()
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	extractDir, err := os.MkdirTemp("", "LocalHosts-update-*")
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	newAppPath, err := extractAppBundle(ctx, installerPath, extractDir)
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	updaterPath := filepath.Join(os.TempDir(), "LocalHosts-updater-bin")
	_ = os.Remove(updaterPath)
	if err := copyFile(exePath, updaterPath); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	cmd := exec.Command(updaterPath,
		"--update-apply",
		"--pid", fmt.Sprintf("%d", os.Getpid()),
		"--target", appPath,
		"--src", newAppPath,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}
	EmitProgress(ctx, Progress{Stage: "done", Message: "update scheduled"})
	return nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

func currentAppBundlePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	path := exe
	for {
		if strings.HasSuffix(path, ".app") {
			return path, nil
		}
		next := filepath.Dir(path)
		if next == path {
			break
		}
		path = next
	}
	return "", fmt.Errorf("UPDATE_INSTALL_FAILED: app bundle not found")
}

func extractAppBundle(ctx context.Context, installerPath string, extractDir string) (string, error) {
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "ditto", "-x", "-k", installerPath, extractDir)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	app := filepath.Join(extractDir, "LocalHosts.app")
	if _, err := os.Stat(app); err == nil {
		return app, nil
	}
	return "", fmt.Errorf("UPDATE_INSTALL_FAILED: extracted app not found")
}
