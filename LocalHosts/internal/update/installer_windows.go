//go:build windows

package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "preparing update"})

	exePath, err := os.Executable()
	if err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	updaterPath := filepath.Join(os.TempDir(), "LocalHosts-updater-bin.exe")
	_ = os.Remove(updaterPath)
	if err := copyFile(exePath, updaterPath); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}

	cmd := exec.Command(updaterPath,
		"--update-apply",
		"--pid", fmt.Sprintf("%d", os.Getpid()),
		"--target", exePath,
		"--src", installerPath,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
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
