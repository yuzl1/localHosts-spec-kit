//go:build darwin

package platform

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
		if err := startHelperViaAppleScript(exePath, port, token); err != nil {
			return err
		}
	}

	if err := waitForElevatedSession(baseURL, 4*time.Second); err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", err.Error())
	}
	setElevatedSession(baseURL, token)
	return nil
}

func startHelperViaSudo(exePath string, port int, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	if err := sudoValidate(ctx); err != nil {
		return err
	}

	args := []string{
		"-n",
		exePath,
		"--elevated-helper",
		"--port", strconv.Itoa(port),
		"--token", token,
		"--parent-pid", strconv.Itoa(os.Getpid()),
	}
	cmd := exec.CommandContext(ctx, "sudo", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
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

func startHelperViaAppleScript(exePath string, port int, token string) error {
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
