//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func RequestAdminAccess() error {
	if hasElevatedSession() && elevatedSessionHealthy(500*time.Millisecond) {
		return nil
	}
	if hasElevatedSession() {
		clearElevatedSession()
	}

	pkexecPath, err := exec.LookPath("pkexec")
	if err != nil {
		return fmt.Errorf("PERMISSION_DENIED: pkexec not found")
	}

	port, err := pickFreeLocalPort()
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	token, err := newSessionToken()
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}

	baseURL := "http://127.0.0.1:" + strconv.Itoa(port)
	cmdString := shQuoteLinux(exePath) +
		" --elevated-helper --port " + strconv.Itoa(port) +
		" --token " + shQuoteLinux(token) +
		" --parent-pid " + strconv.Itoa(os.Getpid()) +
		" >/dev/null 2>&1 &"

	cmd := exec.Command(pkexecPath, "sh", "-c", cmdString)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("CANCELLED: authorization cancelled")
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

	if err := waitForElevatedSession(baseURL, 4*time.Second); err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	setElevatedSession(baseURL, token)
	return nil
}

func shQuoteLinux(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
