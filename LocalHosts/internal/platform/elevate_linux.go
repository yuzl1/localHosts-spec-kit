//go:build linux

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

	pkexecPath, err := exec.LookPath("pkexec")
	if err != nil {
		return fmt.Errorf("pkexec not found")
	}

	backupPath := BackupPath(hostsPath)
	tempDest := TempSystemPath(hostsPath)

	main := strings.Join([]string{
		"cp " + shellQuote(hostsPath) + " " + shellQuote(backupPath),
		"cp " + shellQuote(tmpFile.Name()) + " " + shellQuote(tempDest),
		"chmod 0644 " + shellQuote(tempDest),
		"mv -f " + shellQuote(tempDest) + " " + shellQuote(hostsPath),
	}, " && ")

	rollback := "cp " + shellQuote(backupPath) + " " + shellQuote(hostsPath)
	cmdString := "(" + main + ") || (" + rollback + " && exit 1)"

	cmd := exec.Command(pkexecPath, "sh", "-c", cmdString)
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
		if strings.Contains(low, "not authorized") || strings.Contains(low, "authentication") || strings.Contains(low, "permission denied") {
			return fmt.Errorf("PERMISSION_DENIED: %s", msg)
		}
		return fmt.Errorf("WRITE_FAILED: %s", msg)
	}

	return nil
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
