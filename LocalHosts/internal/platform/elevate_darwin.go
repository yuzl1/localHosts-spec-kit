//go:build darwin

package platform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ApplyHostsWithAdmin(hostsPath string, content []byte) error {
	if err := applyHostsViaSession(hostsPath, content); err == nil {
		return nil
	} else if err != nil && !errors.Is(err, ErrNoElevatedSession) {
		return err
	}

	tmpFile, err := os.CreateTemp("", "localhosts-hosts-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), content, 0o600); err != nil {
		return err
	}

	return applyHostsViaSudo(hostsPath, tmpFile.Name())
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func applyHostsViaSudo(hostsPath string, tmpFilePath string) error {
	backupPath := BackupPath(hostsPath)
	tempDest := TempSystemPath(hostsPath)

	cmdString := strings.Join([]string{
		"cp " + shellQuote(hostsPath) + " " + shellQuote(backupPath),
		"cp " + shellQuote(tmpFilePath) + " " + shellQuote(tempDest),
		"chmod 0644 " + shellQuote(tempDest),
		"mv -f " + shellQuote(tempDest) + " " + shellQuote(hostsPath),
	}, " && ")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := sudoValidate(ctx); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "sudo", "sh", "-c", cmdString)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("CANCELLED: authorization timeout")
		}
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	return nil
}
