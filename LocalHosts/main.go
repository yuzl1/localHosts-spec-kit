package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
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
