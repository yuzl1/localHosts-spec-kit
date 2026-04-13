//go:build windows

package update

import (
	"context"
	"os/exec"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "installing", Message: "starting installer"})
	cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "", installerPath)
	if err := cmd.Start(); err != nil {
		EmitProgress(ctx, Progress{Stage: "error", Message: "install failed"})
		return err
	}
	EmitProgress(ctx, Progress{Stage: "done", Message: "installer started"})
	return nil
}
