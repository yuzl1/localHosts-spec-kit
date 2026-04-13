package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"LocalHosts/internal/platform"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if runElevatedHelperIfNeeded() {
		return
	}
	if runUpdateApplyIfNeeded() {
		return
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "LocalHosts",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runUpdateApplyIfNeeded() bool {
	fs := flag.NewFlagSet("update-apply", flag.ContinueOnError)
	enabled := fs.Bool("update-apply", false, "")
	pid := fs.Int("pid", 0, "")
	target := fs.String("target", "", "")
	src := fs.String("src", "", "")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return false
	}
	if !*enabled {
		return false
	}
	if *pid <= 0 || strings.TrimSpace(*target) == "" || strings.TrimSpace(*src) == "" {
		os.Exit(2)
	}

	logPath := "/tmp/LocalHosts-update-apply.log"
	appendLog := func(msg string) {
		line := time.Now().Format(time.RFC3339) + " " + msg + "\n"
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open update log file: %v\n", err)
			return
		}
		_, _ = f.WriteString(line)
		_ = f.Close()
	}

	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		if !processAlive(*pid) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	t := strings.TrimSpace(*target)
	s := strings.TrimSpace(*src)
	isWin := runtime.GOOS == "windows"
	needsAdmin := !isWin && strings.HasPrefix(t, "/Applications/")

	newPath := t + ".new"
	if isWin {
		newPath = t + ".new.exe"
	}
	backupPath := t + ".bak"
	if isWin {
		backupPath = t + ".bak.exe"
	}
	_ = os.RemoveAll(newPath)

	appendLog("start target=" + t + " src=" + s)

	if isWin {
		// On Windows, s is the setup.exe, we just run it and exit
		appendLog("windows: running setup.exe")
		cmd := exec.Command(s)
		if err := cmd.Start(); err != nil {
			appendLog("start setup failed: " + err.Error())
			os.Exit(1)
		}
		appendLog("setup started, exiting")
		os.Exit(0)
	}

	if needsAdmin {
		cmdString := strings.Join([]string{
			`TARGET=` + shQuote(t),
			`SRC=` + shQuote(s),
			`NEW=` + shQuote(newPath),
			`BAK=` + shQuote(backupPath),
			`/bin/rm -rf "$NEW" "$BAK"`,
			`/usr/bin/ditto "$SRC" "$NEW"`,
			`/usr/bin/xattr -dr com.apple.quarantine "$NEW" >/dev/null 2>&1 || true`,
			`if [ -d "$TARGET" ]; then /bin/mv "$TARGET" "$BAK"; fi`,
			`/bin/mv "$NEW" "$TARGET" || ( [ -d "$BAK" ] && /bin/mv "$BAK" "$TARGET" ; exit 1 )`,
			`/usr/bin/open -n "$TARGET"`,
		}, " ; ")

		script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, escapeOsaString(cmdString))
		cmd := exec.Command("osascript", "-e", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			appendLog("osascript failed: " + strings.TrimSpace(string(out)))
			os.Exit(1)
		}
		appendLog("osascript ok")
		os.Exit(0)
	}

	if err := exec.Command("/usr/bin/ditto", s, newPath).Run(); err != nil {
		appendLog("ditto failed: " + err.Error())
		os.Exit(1)
	}
	_ = exec.Command("/usr/bin/xattr", "-dr", "com.apple.quarantine", newPath).Run()
	_ = os.RemoveAll(backupPath)
	if err := os.Rename(t, backupPath); err != nil {
		appendLog("backup failed: " + err.Error())
	}
	if err := os.Rename(newPath, t); err != nil {
		appendLog("move new failed: " + err.Error())
		_ = os.Rename(backupPath, t)
		os.Exit(1)
	}
	_ = exec.Command("/usr/bin/open", "-n", t).Start()
	appendLog("done")
	os.Exit(0)
	return true
}

func shQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func escapeOsaString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

type elevatedApplyRequest struct {
	Token      string `json:"token"`
	HostsPath  string `json:"hostsPath"`
	ContentB64 string `json:"contentB64"`
}

type elevatedApplyResponse struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

func runElevatedHelperIfNeeded() bool {
	fs := flag.NewFlagSet("elevated-helper", flag.ContinueOnError)
	isHelper := fs.Bool("elevated-helper", false, "")
	port := fs.Int("port", 0, "")
	token := fs.String("token", "", "")
	parentPID := fs.Int("parent-pid", 0, "")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return false
	}

	if !*isHelper {
		return false
	}

	if *port == 0 || *token == "" {
		os.Exit(2)
	}

	if *parentPID > 0 && runtime.GOOS != "windows" {
		go func() {
			for {
				if !processAlive(*parentPID) {
					os.Exit(0)
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}

	addr := "127.0.0.1:" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(1)
		return true
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req elevatedApplyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, elevatedApplyResponse{OK: false, Code: "WRITE_FAILED", Error: err.Error()})
			return
		}
		if req.Token != *token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		raw, err := base64.StdEncoding.DecodeString(req.ContentB64)
		if err != nil {
			writeJSON(w, elevatedApplyResponse{OK: false, Code: "WRITE_FAILED", Error: err.Error()})
			return
		}
		if err := platform.ApplyHostsDirect(req.HostsPath, raw); err != nil {
			writeJSON(w, elevatedApplyResponse{OK: false, Code: "WRITE_FAILED", Error: err.Error()})
			return
		}
		writeJSON(w, elevatedApplyResponse{OK: true})
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req elevatedApplyRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Token != *token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		writeJSON(w, elevatedApplyResponse{OK: true})
		go func() {
			time.Sleep(50 * time.Millisecond)
			os.Exit(0)
		}()
	})

	srv := http.Server{
		Handler: mux,
	}
	_ = srv.Serve(ln)
	return true
}

func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
