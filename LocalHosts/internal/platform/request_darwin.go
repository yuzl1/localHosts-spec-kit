//go:build darwin

package platform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func RequestAdminAccess() error {
	if hasElevatedSession() && elevatedSessionHealthy(500*time.Millisecond) {
		return nil
	}
	if hasElevatedSession() {
		clearElevatedSession()
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
	if err := startHelperViaSudo(exePath, port, token); err != nil {
		return err
	}
	if err := waitForElevatedSession(baseURL, 60*time.Second); err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	setElevatedSession(baseURL, token)
	return nil
}

func startHelperViaSudo(exePath string, port int, token string) error {
	if _, err := exec.LookPath("script"); err != nil {
		return err
	}

	cmd := exec.Command("script", "-q", "/dev/null",
		"sudo",
		exePath,
		"--elevated-helper",
		"--port", strconv.Itoa(port),
		"--token", token,
		"--parent-pid", strconv.Itoa(os.Getpid()),
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("PERMISSION_DENIED: %s", err.Error())
	}
	return nil
}
