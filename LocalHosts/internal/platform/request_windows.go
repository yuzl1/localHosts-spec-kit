//go:build windows

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

	psCommand := strings.Join([]string{
		"$ErrorActionPreference = 'Stop';",
		"Start-Process -Verb RunAs -WindowStyle Hidden -FilePath " + psQuoteHelper(exePath) + " -ArgumentList " + psArgs([]string{
			"--elevated-helper",
			"--port", strconv.Itoa(port),
			"--token", token,
			"--parent-pid", strconv.Itoa(os.Getpid()),
		}) + ";",
	}, " ")

	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
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
		if strings.Contains(low, "denied") || strings.Contains(low, "unauthorized") {
			return fmt.Errorf("PERMISSION_DENIED: %s", msg)
		}
		return fmt.Errorf("WRITE_FAILED: %s", msg)
	}

	if err := waitForElevatedSession(baseURL, 6*time.Second); err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	setElevatedSession(baseURL, token)
	return nil
}

func psQuoteHelper(s string) string {
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, `"`, "`\"")
	return `"` + s + `"`
}

func psArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, a := range args {
		quoted = append(quoted, psQuoteHelper(a))
	}
	return strings.Join(quoted, ",")
}
