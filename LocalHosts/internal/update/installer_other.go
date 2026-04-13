//go:build !darwin && !windows

package update

import (
	"context"
	"fmt"
)

func InstallUpdate(ctx context.Context, installerPath string) error {
	EmitProgress(ctx, Progress{Stage: "error", Message: "unsupported platform"})
	return fmt.Errorf("UPDATE_UNSUPPORTED_PLATFORM")
}
