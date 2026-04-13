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

	// Create shell script for update
	scriptPath := filepath.Join(os.TempDir(), "LocalHosts-updater.sh")
	logPath := "/tmp/LocalHosts-update.log"
	
	script := fmt.Sprintf(`#!/bin/bash
(
  echo "--- Start Update $(date) ---"
  echo "PID: %d"
  echo "Target: %s"
  echo "Src: %s"
  
  # Wait for main process to exit
  while kill -0 %d 2>/dev/null; do
    sleep 0.5
  done
  
  # Perform replacement
  rm -rf %s
  mv %s %s
  
  # Clean quarantine
  xattr -dr com.apple.quarantine %s >/dev/null 2>&1 || true
  
  # Reopen
  open -n %s
  echo "--- Done ---"
) >> %s 2>&1 &
`, os.Getpid(), appPath, newAppPath, os.Getpid(), 
shQuote(appPath), shQuote(newAppPath), shQuote(appPath), shQuote(appPath), shQuote(appPath), logPath)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	EmitProgress(ctx, Progress{Stage: "done", Message: "update scheduled"})
	return nil
}

func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
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
