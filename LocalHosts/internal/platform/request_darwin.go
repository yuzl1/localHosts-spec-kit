//go:build darwin

package platform

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

func RequestAdminAccess() error {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	return sudoValidate(ctx)
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
