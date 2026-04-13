//go:build windows

package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "starting installer"})
	cmd := exec.CommandContext(ctx, installerPath)
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}
	EmitProgress(ctx, Progress{Stage: "done", Message: "installer started"})
	return nil
}

func copyFile(src, dst string) error {
	// Not needed for Windows NSIS installer, as we just run the installer directly
	return nil
}
