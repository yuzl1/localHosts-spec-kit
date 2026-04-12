package platform

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func SystemHostsPath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts`
	default:
		return "/etc/hosts"
	}
}

func UserAppDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "LocalHosts"), nil
}

func WorkspacePath(appDir string) string {
	return filepath.Join(appDir, "workspace.json")
}

func BackupPath(hostsPath string) string {
	return hostsPath + ".localhosts.bak"
}

func TempSystemPath(hostsPath string) string {
	return hostsPath + ".localhosts.tmp"
}

type elevatedSession struct {
	BaseURL string
	Token   string
}

var (
	elevatedMu         sync.Mutex
	elevatedSessionRef *elevatedSession
)

var ErrNoElevatedSession = errors.New("no elevated session")

func hasElevatedSession() bool {
	elevatedMu.Lock()
	defer elevatedMu.Unlock()
	return elevatedSessionRef != nil && elevatedSessionRef.BaseURL != "" && elevatedSessionRef.Token != ""
}

func setElevatedSession(baseURL string, token string) {
	elevatedMu.Lock()
	defer elevatedMu.Unlock()
	elevatedSessionRef = &elevatedSession{BaseURL: baseURL, Token: token}
}

func getElevatedSession() *elevatedSession {
	elevatedMu.Lock()
	defer elevatedMu.Unlock()
	if elevatedSessionRef == nil {
		return nil
	}
	cp := *elevatedSessionRef
	return &cp
}

func pickFreeLocalPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected addr type")
	}
	return addr.Port, nil
}

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func waitForElevatedSession(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := http.Client{Timeout: 800 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(120 * time.Millisecond)
	}
	return fmt.Errorf("elevated helper not ready")
}

type applyRequest struct {
	Token      string `json:"token"`
	HostsPath  string `json:"hostsPath"`
	ContentB64 string `json:"contentB64"`
}

type applyResponse struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

func applyHostsViaSession(hostsPath string, content []byte) error {
	s := getElevatedSession()
	if s == nil {
		return ErrNoElevatedSession
	}
	reqBody := applyRequest{
		Token:      s.Token,
		HostsPath:  hostsPath,
		ContentB64: base64.StdEncoding.EncodeToString(content),
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %w", err)
	}

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(s.BaseURL+"/apply", "application/json", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("WRITE_FAILED: %w", err)
	}
	var parsed applyResponse
	if err := json.Unmarshal(out, &parsed); err != nil {
		return fmt.Errorf("WRITE_FAILED: %s", string(out))
	}
	if parsed.OK {
		return nil
	}
	code := strings.TrimSpace(parsed.Code)
	if code == "" {
		code = "WRITE_FAILED"
	}
	msg := strings.TrimSpace(parsed.Error)
	if msg == "" {
		msg = "unknown error"
	}
	return fmt.Errorf("%s: %s", code, msg)
}

func ApplyHostsDirect(hostsPath string, content []byte) error {
	backupPath := BackupPath(hostsPath)
	tempDest := TempSystemPath(hostsPath)

	orig, origErr := os.ReadFile(hostsPath)
	if origErr == nil {
		if err := os.WriteFile(backupPath, orig, 0o644); err != nil {
			return err
		}
	}

	if runtime.GOOS == "windows" {
		if err := os.WriteFile(hostsPath, content, 0o644); err != nil {
			if origErr == nil {
				_ = os.WriteFile(hostsPath, orig, 0o644)
			}
			return err
		}
		return nil
	}

	if err := os.WriteFile(tempDest, content, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tempDest, hostsPath); err != nil {
		_ = os.Remove(tempDest)
		if origErr == nil {
			_ = os.WriteFile(hostsPath, orig, 0o644)
		}
		return err
	}
	return nil
}
