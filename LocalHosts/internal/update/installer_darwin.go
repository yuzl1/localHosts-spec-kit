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
	"time"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "preparing update"})

	appPath, err := currentAppBundlePath()
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

	needsSudo := strings.HasPrefix(appPath, "/Applications/")
	script := `set -e
pid="$1"
target="$2"
src="$3"
while kill -0 "$pid" 2>/dev/null; do
  sleep 0.2
done
rm -rf "$target.tmp"
ditto "$src" "$target.tmp"
rm -rf "$target"
mv "$target.tmp" "$target"
open "$target"
`

	runCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if needsSudo {
		if err := sudoValidate(runCtx); err != nil {
			EmitProgress(ctx, Progress{Stage: "error", Message: "permission denied"})
			return err
		}
		cmd := exec.CommandContext(runCtx, "sudo", "-n", "sh", "-c", script, "sh", fmt.Sprintf("%d", os.Getpid()), appPath, newAppPath)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
			return err
		}
		EmitProgress(ctx, Progress{Stage: "done", Message: "update scheduled"})
		return nil
	}

	cmd := exec.CommandContext(runCtx, "sh", "-c", script, "sh", fmt.Sprintf("%d", os.Getpid()), appPath, newAppPath)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}
	EmitProgress(ctx, Progress{Stage: "done", Message: "update scheduled"})
	return nil
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

func sudoValidate(ctx context.Context) error {
	if _, err := exec.LookPath("script"); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "script", "-q", "/dev/null", "sudo", "-v")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("CANCELLED: authorization timeout")
		}
		return fmt.Errorf("PERMISSION_DENIED: %s", err.Error())
	}
	return nil
}
