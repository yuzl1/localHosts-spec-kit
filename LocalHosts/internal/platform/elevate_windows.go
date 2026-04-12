//go:build windows

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	backupPath := BackupPath(hostsPath)

	psCommand := strings.Join([]string{
		"$ErrorActionPreference = 'Stop';",
		"try {",
		"Copy-Item -LiteralPath " + psQuote(hostsPath) + " -Destination " + psQuote(backupPath) + " -Force;",
		"Copy-Item -LiteralPath " + psQuote(tmpFile.Name()) + " -Destination " + psQuote(hostsPath) + " -Force;",
		"} catch {",
		"try { Copy-Item -LiteralPath " + psQuote(backupPath) + " -Destination " + psQuote(hostsPath) + " -Force } catch {}",
		"throw",
		"}",
	}, " ")

	inner := "powershell.exe -NoProfile -ExecutionPolicy Bypass -Command " + psQuote(psCommand)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "Start-Process -Verb RunAs -Wait -FilePath 'cmd.exe' -ArgumentList '/c',"+psQuote(inner))
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return runErr
		}
		low := strings.ToLower(msg)
		if strings.Contains(low, "canceled") || strings.Contains(low, "cancelled") {
			return fmt.Errorf("CANCELLED: %s", msg)
		}
		if strings.Contains(low, "denied") || strings.Contains(low, "not authorized") || strings.Contains(low, "unauthorized") {
			return fmt.Errorf("PERMISSION_DENIED: %s", msg)
		}
		return fmt.Errorf("WRITE_FAILED: %s", msg)
	}

	return nil
}

func psQuote(s string) string {
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, `"`, "`\"")
	return `"` + s + `"`
}
