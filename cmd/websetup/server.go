package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/portapps/discord-ptb-portable/pkg/backup"
	"github.com/portapps/portapps/v3/pkg/log"
)

//go:embed web/*
var embeddedWebFS embed.FS

type ProfileItem struct {
	ID        string `json:"id"`
	Number    int    `json:"number"`
	ColorKey  string `json:"colorKey"`
	NameTH    string `json:"nameTH"`
	NameEN    string `json:"nameEN"`
	ColorHex  string `json:"colorHex"`
	IconPath  string `json:"iconPath"`
	Selected  bool   `json:"selected"`
	IsPrimary bool   `json:"isPrimary"`
}

var defaultProfiles = []ProfileItem{
	{ID: "01_yellow", Number: 1, ColorKey: "yellow", NameTH: "Discord Yellow", NameEN: "Discord Yellow", ColorHex: "#FFFF00", IconPath: "assets/icons/01_yellow_FFFF00.png", Selected: true, IsPrimary: true},
	{ID: "02_yellow_orange", Number: 2, ColorKey: "yellow_orange", NameTH: "Discord Yellow Orange", NameEN: "Discord Yellow Orange", ColorHex: "#FFC000", IconPath: "assets/icons/02_yellow_orange_FFC000.png", Selected: false},
	{ID: "03_orange", Number: 3, ColorKey: "orange", NameTH: "Discord Orange", NameEN: "Discord Orange", ColorHex: "#F58220", IconPath: "assets/icons/03_orange_F58220.png", Selected: false},
	{ID: "04_red_orange", Number: 4, ColorKey: "red_orange", NameTH: "Discord Red Orange", NameEN: "Discord Red Orange", ColorHex: "#FA6304", IconPath: "assets/icons/04_red_orange_FA6304.png", Selected: false},
	{ID: "05_red", Number: 5, ColorKey: "red", NameTH: "Discord Red", NameEN: "Discord Red", ColorHex: "#ED1C24", IconPath: "assets/icons/05_red_ED1C24.png", Selected: false},
	{ID: "06_red_violet", Number: 6, ColorKey: "red_violet", NameTH: "Discord Red Violet", NameEN: "Discord Red Violet", ColorHex: "#BD0074", IconPath: "assets/icons/06_red_violet_BD0074.png", Selected: false},
	{ID: "07_violet", Number: 7, ColorKey: "violet", NameTH: "Discord Violet", NameEN: "Discord Violet", ColorHex: "#59277A", IconPath: "assets/icons/07_violet_59277A.png", Selected: false},
	{ID: "08_blue_violet", Number: 8, ColorKey: "blue_violet", NameTH: "Discord Blue Violet", NameEN: "Discord Blue Violet", ColorHex: "#4400A3", IconPath: "assets/icons/08_blue_violet_4400A3.png", Selected: false},
	{ID: "09_blue", Number: 9, ColorKey: "blue", NameTH: "Discord Blue", NameEN: "Discord Blue", ColorHex: "#0020C2", IconPath: "assets/icons/09_blue_0020C2.png", Selected: false},
	{ID: "10_blue_green", Number: 10, ColorKey: "blue_green", NameTH: "Discord Blue Green", NameEN: "Discord Blue Green", ColorHex: "#008240", IconPath: "assets/icons/10_blue_green_008240.png", Selected: false},
	{ID: "11_green", Number: 11, ColorKey: "green", NameTH: "Discord Green", NameEN: "Discord Green", ColorHex: "#00B050", IconPath: "assets/icons/11_green_00B050.png", Selected: false},
	{ID: "12_yellow_green", Number: 12, ColorKey: "yellow_green", NameTH: "Discord Yellow Green", NameEN: "Discord Yellow Green", ColorHex: "#8EF226", IconPath: "assets/icons/12_yellow_green_8EF226.png", Selected: false},
}

type SetupServer struct {
	listener net.Listener
	port     int
	rootPath string
	srv      *http.Server
	stopOnce sync.Once
}

func NewSetupServer(rootSourcePath string) (*SetupServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed binding local port: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	s := &SetupServer{
		listener: ln,
		port:     port,
		rootPath: rootSourcePath,
	}

	mux := http.NewServeMux()

	// Static web assets
	webContent, err := fs.Sub(embeddedWebFS, "web")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(webContent)))
	} else {
		// Fallback to local disk
		mux.Handle("/", http.FileServer(http.Dir(filepath.Join(rootSourcePath, "cmd", "websetup", "web"))))
	}

	// API endpoints
	mux.HandleFunc("/api/profiles", s.handleGetProfiles)
	mux.HandleFunc("/api/browse", s.handleBrowse)
	mux.HandleFunc("/api/install", s.handleInstall)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/backup", s.handleBackup)
	mux.HandleFunc("/api/restore", s.handleRestore)
	mux.HandleFunc("/api/launch", s.handleLaunch)
	mux.HandleFunc("/api/exit", s.handleExit)

	s.srv = &http.Server{Handler: mux}
	return s, nil
}

func (s *SetupServer) Start() {
	go func() {
		_ = s.srv.Serve(s.listener)
	}()
}

func (s *SetupServer) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

func (s *SetupServer) Stop() {
	s.stopOnce.Do(func() {
		_ = s.srv.Close()
	})
}

func (s *SetupServer) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(defaultProfiles)
}

func (s *SetupServer) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg InstallConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go func() {
		_ = ExecuteInstallation(cfg, s.rootPath)
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"started": true})
}

func (s *SetupServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(getStatus())
}

func (s *SetupServer) handleBackup(w http.ResponseWriter, r *http.Request) {
	profilesDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Discord-Portable", "data", "profiles")
	outZip, err := backup.CreateBackup(profilesDir, "")
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "file": outZip})
}

func (s *SetupServer) handleRestore(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		ZipPath string `json:"zipPath"`
	}
	var body reqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	targetDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Discord-Portable", "data", "profiles")
	err := backup.RestoreBackup(body.ZipPath, targetDir)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (s *SetupServer) handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	psScript := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = 'เลือกโฟลเดอร์สำหรับติดตั้ง Discord PTB Portable'; if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $f.SelectedPath }`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	out, err := cmd.Output()
	chosenPath := strings.TrimSpace(string(out))
	if err != nil || chosenPath == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"path": ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"path": chosenPath})
}

func (s *SetupServer) handleLaunch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"launched": true})
}

func (s *SetupServer) handleExit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	go func() {
		log.Info().Msg("WebSetup UI closed, shutting down server")
		os.Exit(0)
	}()
}
