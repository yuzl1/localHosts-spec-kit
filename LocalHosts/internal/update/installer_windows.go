//go:build windows

package update

import (
	"context"
	"os/exec"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "starting installer"})
	// NSIS installer handles process termination and restart
	cmd := exec.Command(installerPath)
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}
	EmitProgress(ctx, Progress{Stage: "done", Message: "installer started"})
	return nil
}

func copyFile(src, dst string) error {
	return nil
}
