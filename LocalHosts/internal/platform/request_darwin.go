//go:build darwin

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
	if hasElevatedSession() {
		return nil
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
	cmdString := shQuote(exePath) +
		" --elevated-helper --port " + strconv.Itoa(port) +
		" --token " + shQuote(token) +
		" --parent-pid " + strconv.Itoa(os.Getpid()) +
		" >/dev/null 2>&1 &"

	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, escapeOsaString2(cmdString))
	cmd := exec.Command("osascript", "-e", script)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("CANCELLED: authorization cancelled")
		}
		low := strings.ToLower(msg)
		if strings.Contains(low, "user canceled") || strings.Contains(low, "canceled") {
			return fmt.Errorf("CANCELLED: %s", msg)
		}
		if strings.Contains(low, "not authorized") || strings.Contains(low, "authentication failed") {
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

func shQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func escapeOsaString2(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
