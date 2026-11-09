package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"github.com/portapps/discord-ptb-portable/pkg/security"
	"github.com/portapps/discord-ptb-portable/pkg/shell"
	"github.com/portapps/discord-ptb-portable/pkg/trayanim"
	"golang.org/x/sys/windows"
)

//go:embed web/*
var embeddedWebFS embed.FS

var (
	kernel32DLL            = syscall.NewLazyDLL("kernel32.dll")
	user32DLL              = syscall.NewLazyDLL("user32.dll")
	gdi32DLL               = syscall.NewLazyDLL("gdi32.dll")
	procFreeConsole        = kernel32DLL.NewProc("FreeConsole")
	procRegisterClassExW   = user32DLL.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32DLL.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32DLL.NewProc("DefWindowProcW")
	procUpdateWindowWin    = user32DLL.NewProc("UpdateWindow")
	procShowWindowWin      = user32DLL.NewProc("ShowWindow")
	procGetMessageW        = user32DLL.NewProc("GetMessageW")
	procTranslateMessage   = user32DLL.NewProc("TranslateMessage")
	procDispatchMessageW   = user32DLL.NewProc("DispatchMessageW")
	procPostQuitMessage    = user32DLL.NewProc("PostQuitMessage")
	procGetSystemMetrics   = user32DLL.NewProc("GetSystemMetrics")
	procLoadCursorW        = user32DLL.NewProc("LoadCursorW")
	procSetCursor          = user32DLL.NewProc("SetCursor")
	procCreateRoundRectRgn = gdi32DLL.NewProc("CreateRoundRectRgn")
	procSetWindowRgn       = user32DLL.NewProc("SetWindowRgn")
	procGetWindowRectWin   = user32DLL.NewProc("GetWindowRect")
	procSetWindowPosWin    = user32DLL.NewProc("SetWindowPos")
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       struct{ X, Y int32 }
	LPrivate uint32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

const (
	WS_POPUP             = 0x80000000
	WS_CLIPCHILDREN      = 0x02000000
	WS_EX_APPWINDOW      = 0x00040000
	CS_DROPSHADOW        = 0x00020000
	WM_DESTROY           = 0x0002
	WM_MOVE              = 0x0003
	WM_SIZE              = 0x0005
	WM_CLOSE             = 0x0010
	WM_SETCURSOR         = 0x0020
	SW_SHOW              = 5
	SW_SHOWMINIMIZED_CMD = 2
	SWP_NOSIZE_CMD       = 0x0001
	SWP_NOZORDER_CMD     = 0x0004
	SM_CXSCREEN          = 0
	SM_CYSCREEN          = 1
)

var (
	globalChromium *edge.Chromium
	defaultCursor  uintptr
)

type ProfileItem struct {
	ID       string `json:"id"`
	Number   int    `json:"number"`
	ColorKey string `json:"colorKey"`
	NameTH   string `json:"nameTH"`
	NameEN   string `json:"nameEN"`
	ColorHex string `json:"colorHex"`
	IconPath string `json:"iconPath"`
}

var allProfiles = []ProfileItem{
	{ID: "01_yellow", Number: 1, ColorKey: "yellow", NameTH: "Discord Yellow", NameEN: "Discord Yellow", ColorHex: "#FFFF00", IconPath: "assets/icons/01_yellow_FFFF00.png"},
	{ID: "02_yellow_orange", Number: 2, ColorKey: "yellow_orange", NameTH: "Discord Yellow Orange", NameEN: "Discord Yellow Orange", ColorHex: "#FFC000", IconPath: "assets/icons/02_yellow_orange_FFC000.png"},
	{ID: "03_orange", Number: 3, ColorKey: "orange", NameTH: "Discord Orange", NameEN: "Discord Orange", ColorHex: "#F58220", IconPath: "assets/icons/03_orange_F58220.png"},
	{ID: "04_red_orange", Number: 4, ColorKey: "red_orange", NameTH: "Discord Red Orange", NameEN: "Discord Red Orange", ColorHex: "#FA6304", IconPath: "assets/icons/04_red_orange_FA6304.png"},
	{ID: "05_red", Number: 5, ColorKey: "red", NameTH: "Discord Red", NameEN: "Discord Red", ColorHex: "#ED1C24", IconPath: "assets/icons/05_red_ED1C24.png"},
	{ID: "06_red_violet", Number: 6, ColorKey: "red_violet", NameTH: "Discord Red Violet", NameEN: "Discord Red Violet", ColorHex: "#BD0074", IconPath: "assets/icons/06_red_violet_BD0074.png"},
	{ID: "07_violet", Number: 7, ColorKey: "violet", NameTH: "Discord Violet", NameEN: "Discord Violet", ColorHex: "#59277A", IconPath: "assets/icons/07_violet_59277A.png"},
	{ID: "08_blue_violet", Number: 8, ColorKey: "blue_violet", NameTH: "Discord Blue Violet", NameEN: "Discord Blue Violet", ColorHex: "#4400A3", IconPath: "assets/icons/08_blue_violet_4400A3.png"},
	{ID: "09_blue", Number: 9, ColorKey: "blue", NameTH: "Discord Blue", NameEN: "Discord Blue", ColorHex: "#0020C2", IconPath: "assets/icons/09_blue_0020C2.png"},
	{ID: "10_blue_green", Number: 10, ColorKey: "blue_green", NameTH: "Discord Blue Green", NameEN: "Discord Blue Green", ColorHex: "#008240", IconPath: "assets/icons/10_blue_green_008240.png"},
	{ID: "11_green", Number: 11, ColorKey: "green", NameTH: "Discord Green", NameEN: "Discord Green", ColorHex: "#00B050", IconPath: "assets/icons/11_green_00B050.png"},
	{ID: "12_yellow_green", Number: 12, ColorKey: "yellow_green", NameTH: "Discord Yellow Green", NameEN: "Discord Yellow Green", ColorHex: "#8EF226", IconPath: "assets/icons/12_yellow_green_8EF226.png"},
}

func framelessWndProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case WM_SETCURSOR:
		if defaultCursor != 0 {
			procSetCursor.Call(defaultCursor)
			return 1
		}
	case WM_SIZE:
		if globalChromium != nil {
			globalChromium.Resize()
		}
	case WM_MOVE:
		if globalChromium != nil {
			_ = globalChromium.NotifyParentWindowPositionChanged()
		}
	case WM_CLOSE:
		procPostQuitMessage.Call(0)
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
	return r
}

func isKnownProfile(id string) bool {
	for _, p := range allProfiles {
		if strings.EqualFold(p.ID, id) {
			return true
		}
	}
	return false
}

func getInstalledProfiles(rootDir string) []string {
	profilesDir := filepath.Join(rootDir, "data", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var installed []string
	for _, entry := range entries {
		if entry.IsDir() && isKnownProfile(entry.Name()) {
			installed = append(installed, entry.Name())
		}
	}
	return installed
}

func resolveLauncher(rootDir string) string {
	candidates := []string{
		filepath.Join(rootDir, "DiscordPTB-Portable.exe"),
		filepath.Join(rootDir, "Discord-Portable.exe"),
	}
	exePath, err := os.Executable()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exePath), "DiscordPTB-Portable.exe"),
			filepath.Join(filepath.Dir(exePath), "Discord-Portable.exe"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return filepath.Join(rootDir, "DiscordPTB-Portable.exe")
}

func resolveIconDir(rootDir string) string {
	candidate := filepath.Join(rootDir, "res", "icons")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	exePath, err := os.Executable()
	if err == nil {
		cand2 := filepath.Join(filepath.Dir(exePath), "res", "icons")
		if _, err := os.Stat(cand2); err == nil {
			return cand2
		}
	}
	return candidate
}

func profileDisplayName(prof string) string {
	switch strings.ToLower(strings.TrimSpace(prof)) {
	case "01_yellow", "yellow":
		return "Discord Yellow"
	case "02_yellow_orange", "yellow_orange":
		return "Discord Yellow Orange"
	case "03_orange", "orange":
		return "Discord Orange"
	case "04_red_orange", "red_orange":
		return "Discord Red Orange"
	case "05_red", "red":
		return "Discord Red"
	case "06_red_violet", "red_violet":
		return "Discord Red Violet"
	case "07_violet", "violet":
		return "Discord Violet"
	case "08_blue_violet", "blue_violet":
		return "Discord Blue Violet"
	case "09_blue", "blue":
		return "Discord Blue"
	case "10_blue_green", "blue_green":
		return "Discord Blue Green"
	case "11_green", "green":
		return "Discord Green"
	case "12_yellow_green", "yellow_green":
		return "Discord Yellow Green"
	default:
		cleaned := strings.TrimLeft(prof, "0123456789_")
		words := strings.Split(cleaned, "_")
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
		}
		return "Discord " + strings.Join(words, " ")
	}
}

func resolveIcon(prof, iconDir string) string {
	target := filepath.Join(iconDir, prof+".ico")
	if _, err := os.Stat(target); err == nil {
		return target
	}
	entries, err := os.ReadDir(iconDir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(strings.ToLower(name), ".ico") && strings.HasPrefix(name, prof) {
				return filepath.Join(iconDir, name)
			}
		}
	}
	return filepath.Join(filepath.Dir(iconDir), "papp.ico")
}

func terminateDirectoryProcesses(dir string) {
	if dir == "" {
		return
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	psCmd := fmt.Sprintf(
		"Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -like '%s*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }",
		absDir,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	_ = cmd.Run()
}

type ManagerServer struct {
	listener    net.Listener
	port        int
	rootDir     string
	action      string
	srv         *http.Server
	windowHwnd  uintptr
	windowMutex sync.RWMutex
}

func (s *ManagerServer) handleState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	installed := getInstalledProfiles(s.rootDir)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"action":            s.action,
		"installedProfiles": installed,
		"allProfiles":       allProfiles,
	})
}

func terminateProfileProcesses(rootDir, profileID string) {
	if rootDir == "" || profileID == "" {
		return
	}
	absDir, err := filepath.Abs(rootDir)
	if err != nil {
		absDir = rootDir
	}
	currentPID := os.Getpid()
	psCmd := fmt.Sprintf(
		"Get-CimInstance Win32_Process | Where-Object { $_.ProcessId -ne %d -and $_.Name -ne 'Discord-ProfileManager.exe' -and ($_.CommandLine -like '*--profile=%s*' -or $_.CommandLine -like '*profiles\\%s*') -and $_.ExecutablePath -like '%s*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }",
		currentPID,
		profileID,
		profileID,
		absDir,
	)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

func executeFullUninstall(rootDir, desktopDir, startMenuDir string, parentPid int) error {
	// 1. Unregister Context Menu completely from registry
	shell.UnregisterDiscordContextMenu()

	// 2. Remove desktop shortcuts for all 12 profiles and generic shortcuts
	for _, p := range allProfiles {
		dispName := profileDisplayName(p.ID)
		_ = os.Remove(filepath.Join(desktopDir, fmt.Sprintf("%s.lnk", dispName)))
	}
	_ = os.Remove(filepath.Join(desktopDir, "Discord-Portable.lnk"))
	_ = os.Remove(filepath.Join(desktopDir, "DiscordPTB-Portable.lnk"))

	// Clean up any other Discord *.lnk shortcuts from desktop
	if entries, err := os.ReadDir(desktopDir); err == nil {
		for _, e := range entries {
			lowerName := strings.ToLower(e.Name())
			if strings.HasPrefix(lowerName, "discord") && strings.HasSuffix(lowerName, ".lnk") {
				_ = os.Remove(filepath.Join(desktopDir, e.Name()))
			}
		}
	}

	// 3. Remove Start Menu directory
	_ = os.RemoveAll(startMenuDir)

	// 4. Create detached cleanup batch file in %TEMP%
	batchScriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("discord_clean_%d.cmd", time.Now().UnixNano()))
	cleanCmd := fmt.Sprintf(`@echo off
set "PARENT_PID=%%~1"
set "TARGET_DIR=%%~2"

:wait_pid
if not "%%PARENT_PID%%"=="" if not "%%PARENT_PID%%"=="0" (
    tasklist /fi "PID eq %%PARENT_PID%%" 2>nul | findstr /i "%%PARENT_PID%%" >nul
    if not errorlevel 1 (
        timeout /t 1 /nobreak >nul
        goto wait_pid
    )
)

powershell.exe -NoProfile -NonInteractive -Command "Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -like '%%TARGET_DIR%%\*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }" >nul 2>&1
taskkill /f /im DiscordPTB.exe >nul 2>&1
taskkill /f /im DiscordPTB-Portable.exe >nul 2>&1
taskkill /f /im Discord-ProfileManager.exe >nul 2>&1

for /l %%%%i in (1, 1, 30) do (
    if not exist "%%TARGET_DIR%%" goto done
    rd /s /q "%%TARGET_DIR%%" >nul 2>&1
    if not exist "%%TARGET_DIR%%" goto done
    timeout /t 1 /nobreak >nul
)

:done
del "%%~f0" >nul 2>&1
`)

	_ = os.WriteFile(batchScriptPath, []byte(cleanCmd), 0644)

	cmd := exec.Command("cmd.exe", "/C", "start", "", "/B", batchScriptPath, strconv.Itoa(parentPid), rootDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

func (s *ManagerServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	trayanim.GetGlobalAnimator().StartAnimation("scuba", 60)
	defer trayanim.GetGlobalAnimator().StopAnimation()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type reqPayload struct {
		Action   string   `json:"action"`
		Profiles []string `json:"profiles"`
		All      bool     `json:"all"`
	}

	var req reqPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desktopDir := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
	startMenuDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Discord Portable")
	iconDirectory := resolveIconDir(s.rootDir)
	launcherPath := resolveLauncher(s.rootDir)

	if req.Action == "add" {
		profilesBaseDir := filepath.Join(s.rootDir, "data", "profiles")
		for _, prof := range req.Profiles {
			profDir := filepath.Join(profilesBaseDir, prof)
			_ = security.ProtectPath(profDir)

			// settings.json
			settingsPath := filepath.Join(profDir, "settings.json")
			_ = os.WriteFile(settingsPath, []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}"), 0644)

			dispName := profileDisplayName(prof)
			shortcutName := fmt.Sprintf("%s.lnk", dispName)
			iconFile := resolveIcon(prof, iconDirectory)
			aumid := fmt.Sprintf("Discord.PTB.Profile.%s", prof)
			args := fmt.Sprintf("--profile=%s", prof)

			// Desktop shortcut
			desktopLnk := filepath.Join(desktopDir, shortcutName)
			_ = shell.CreateProfileShortcut(desktopLnk, launcherPath, args, iconFile, aumid, dispName, s.rootDir)

			// Start Menu shortcut (for Windows 11 Shell Taskbar AUMID matching)
			_ = os.MkdirAll(startMenuDir, 0755)
			startMenuLnk := filepath.Join(startMenuDir, shortcutName)
			_ = shell.CreateProfileShortcut(startMenuLnk, launcherPath, args, iconFile, aumid, dispName, s.rootDir)
		}

		// Ensure context menu keys are registered pointing to this manager
		managerExe := filepath.Join(s.rootDir, "Discord-ProfileManager.exe")
		if _, err := os.Stat(managerExe); os.IsNotExist(err) {
			managerExe, _ = os.Executable()
		}
		if !strings.HasSuffix(strings.ToLower(managerExe), ".test.exe") && !strings.Contains(strings.ToLower(managerExe), "go-build") {
			_ = shell.RegisterDiscordContextMenu(managerExe, iconDirectory)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	if req.Action == "uninstall" {
		if req.All {
			_ = executeFullUninstall(s.rootDir, desktopDir, startMenuDir, os.Getpid())
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
			return
		}

		// Partial Uninstall: remove chosen profiles only
		for _, prof := range req.Profiles {
			terminateProfileProcesses(s.rootDir, prof)

			dispName := profileDisplayName(prof)
			shortcutName := fmt.Sprintf("%s.lnk", dispName)

			// Remove desktop shortcut
			_ = os.Remove(filepath.Join(desktopDir, shortcutName))

			// Remove Start Menu shortcut
			_ = os.Remove(filepath.Join(startMenuDir, shortcutName))

			// Remove profile directory
			profDir := filepath.Join(s.rootDir, "data", "profiles", prof)
			_ = os.RemoveAll(profDir)
		}

		// If no profiles left after deletion, perform full uninstall
		remaining := getInstalledProfiles(s.rootDir)
		if len(remaining) == 0 {
			_ = executeFullUninstall(s.rootDir, desktopDir, startMenuDir, os.Getpid())
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	http.Error(w, "Invalid action", http.StatusBadRequest)
}


func (s *ManagerServer) handleTrayAnimate(w http.ResponseWriter, r *http.Request) {
	animType := r.URL.Query().Get("type")
	intervalStr := r.URL.Query().Get("interval")
	interval, _ := strconv.Atoi(intervalStr)
	trayanim.GetGlobalAnimator().StartAnimation(animType, interval)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *ManagerServer) handleTrayStop(w http.ResponseWriter, r *http.Request) {
	trayanim.GetGlobalAnimator().StopAnimation()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *ManagerServer) handleExit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	go func() {
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()
}

func (s *ManagerServer) handleWindowMinimize(w http.ResponseWriter, r *http.Request) {
	s.windowMutex.RLock()
	hwnd := s.windowHwnd
	s.windowMutex.RUnlock()
	if hwnd != 0 {
		procShowWindowWin.Call(hwnd, SW_SHOWMINIMIZED_CMD)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *ManagerServer) handleWindowMove(w http.ResponseWriter, r *http.Request) {
	s.windowMutex.RLock()
	hwnd := s.windowHwnd
	s.windowMutex.RUnlock()
	dxStr := r.URL.Query().Get("dx")
	dyStr := r.URL.Query().Get("dy")
	dx, _ := strconv.Atoi(dxStr)
	dy, _ := strconv.Atoi(dyStr)

	if hwnd != 0 && (dx != 0 || dy != 0) {
		var rect RECT
		procGetWindowRectWin.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
		procSetWindowPosWin.Call(hwnd, 0, uintptr(rect.Left+int32(dx)), uintptr(rect.Top+int32(dy)), 0, 0, SWP_NOSIZE_CMD|SWP_NOZORDER_CMD)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func main() {

	procFreeConsole.Call()

	actionFlag := flag.String("action", "add", "Action mode: add or uninstall")
	dirFlag := flag.String("dir", "", "Target installation directory")
	silentFlag := flag.Bool("silent", false, "Run silently without UI")
	allFlag := flag.Bool("all", false, "Execute for all profiles")
	profilesFlag := flag.String("profiles", "", "Comma-separated list of profile IDs")
	flag.Parse()

	action := strings.ToLower(strings.TrimSpace(*actionFlag))
	if action != "uninstall" {
		action = "add"
	}

	exePath, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	rootDirRaw := strings.Trim(strings.TrimSpace(*dirFlag), "\"")
	var rootDir string
	if rootDirRaw == "" {
		rootDir = filepath.Dir(exePath)
	} else {
		rootDir = filepath.Clean(rootDirRaw)
	}
	if _, err := os.Stat(filepath.Join(rootDir, "res")); os.IsNotExist(err) {
		// If running from cmd/manager during development
		rootDir = filepath.Dir(filepath.Dir(rootDir))
	}

	if *silentFlag {
		desktopDir := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
		startMenuDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Discord Portable")
		launcherPath := resolveLauncher(rootDir)
		iconDirectory := resolveIconDir(rootDir)

		if action == "uninstall" {
			if *allFlag || len(getInstalledProfiles(rootDir)) == 0 {
				_ = executeFullUninstall(rootDir, desktopDir, startMenuDir, os.Getpid())
			} else if *profilesFlag != "" {
				for _, p := range strings.Split(*profilesFlag, ",") {
					p = strings.TrimSpace(p)
					if isKnownProfile(p) {
						terminateProfileProcesses(rootDir, p)
						dispName := profileDisplayName(p)
						_ = os.Remove(filepath.Join(desktopDir, fmt.Sprintf("%s.lnk", dispName)))
						_ = os.Remove(filepath.Join(startMenuDir, fmt.Sprintf("%s.lnk", dispName)))
						_ = os.RemoveAll(filepath.Join(rootDir, "data", "profiles", p))
					}
				}
				if len(getInstalledProfiles(rootDir)) == 0 {
					_ = executeFullUninstall(rootDir, desktopDir, startMenuDir, os.Getpid())
				}
			}
		} else if action == "add" {
			var targets []string
			if *profilesFlag != "" {
				for _, p := range strings.Split(*profilesFlag, ",") {
					p = strings.TrimSpace(p)
					if isKnownProfile(p) {
						targets = append(targets, p)
					}
				}
			} else if *allFlag {
				inst := make(map[string]bool)
				for _, p := range getInstalledProfiles(rootDir) {
					inst[p] = true
				}
				for _, p := range allProfiles {
					if !inst[p.ID] {
						targets = append(targets, p.ID)
					}
				}
			}
			profilesBaseDir := filepath.Join(rootDir, "data", "profiles")
			for _, prof := range targets {
				profDir := filepath.Join(profilesBaseDir, prof)
				_ = security.ProtectPath(profDir)
				settingsPath := filepath.Join(profDir, "settings.json")
				_ = os.WriteFile(settingsPath, []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}"), 0644)

				dispName := profileDisplayName(prof)
				shortcutName := fmt.Sprintf("%s.lnk", dispName)
				iconFile := resolveIcon(prof, iconDirectory)
				aumid := fmt.Sprintf("Discord.PTB.Profile.%s", prof)
				args := fmt.Sprintf("--profile=%s", prof)

				desktopLnk := filepath.Join(desktopDir, shortcutName)
				_ = shell.CreateProfileShortcut(desktopLnk, launcherPath, args, iconFile, aumid, dispName, rootDir)

				_ = os.MkdirAll(startMenuDir, 0755)
				startMenuLnk := filepath.Join(startMenuDir, shortcutName)
				_ = shell.CreateProfileShortcut(startMenuLnk, launcherPath, args, iconFile, aumid, dispName, rootDir)
			}
			managerExe := filepath.Join(rootDir, "Discord-ProfileManager.exe")
			if _, err := os.Stat(managerExe); os.IsNotExist(err) {
				managerExe, _ = os.Executable()
			}
			if !strings.HasSuffix(strings.ToLower(managerExe), ".test.exe") && !strings.Contains(strings.ToLower(managerExe), "go-build") {
				_ = shell.RegisterDiscordContextMenu(managerExe, iconDirectory)
			}
		}
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		os.Exit(1)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	server := &ManagerServer{
		listener: ln,
		port:     port,
		rootDir:  rootDir,
		action:   action,
	}

	mux := http.NewServeMux()
	webContent, err := fs.Sub(embeddedWebFS, "web")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(webContent)))
	} else {
		mux.Handle("/", http.FileServer(http.Dir(filepath.Join(rootDir, "cmd", "manager", "web"))))
	}

	mux.HandleFunc("/api/state", server.handleState)
	mux.HandleFunc("/api/execute", server.handleExecute)
	mux.HandleFunc("/api/exit", server.handleExit)
	mux.HandleFunc("/api/window/minimize", server.handleWindowMinimize)
	mux.HandleFunc("/api/window/move", server.handleWindowMove)
	mux.HandleFunc("/api/tray/animate", server.handleTrayAnimate)
	mux.HandleFunc("/api/tray/stop", server.handleTrayStop)

	server.srv = &http.Server{Handler: mux}
	go func() {
		_ = server.srv.Serve(server.listener)
	}()

	var hinst windows.Handle
	windows.GetModuleHandleEx(0, nil, &hinst)

	className, _ := windows.UTF16PtrFromString("DiscordProfileManagerWindow")
	titleName, _ := windows.UTF16PtrFromString("Discord PTB - Profile Manager")

	arrowCursor, _, _ := procLoadCursorW.Call(0, 32512)
	defaultCursor = arrowCursor

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_DROPSHADOW,
		LpfnWndProc:   windows.NewCallback(framelessWndProc),
		HInstance:     hinst,
		HCursor:       windows.Handle(arrowCursor),
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	width := int32(800)
	height := int32(500)
	scrW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	scrH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	x := (int32(scrW) - width) / 2
	y := (int32(scrH) - height) / 2

	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_APPWINDOW,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titleName)),
		WS_POPUP|WS_CLIPCHILDREN,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		0, 0,
		uintptr(hinst),
		0,
	)

	if hwnd == 0 {
		os.Exit(1)
	}

	// Rounded corners (radius 14px -> diameter 28)
	rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, uintptr(width+1), uintptr(height+1), 28, 28)
	if rgn != 0 {
		procSetWindowRgn.Call(hwnd, rgn, 1)
	}

	server.windowMutex.Lock()
	server.windowHwnd = hwnd
	server.windowMutex.Unlock()

	animator := trayanim.GetGlobalAnimator()
	_ = animator.Init(hwnd, "Discord PTB - Profile Manager")
	_ = animator.AddTray()
	defer animator.RemoveTray()

	if defIcon := animator.DefaultIcon(); defIcon != 0 {
		user32DLL.NewProc("SendMessageW").Call(hwnd, 0x0080, 0, defIcon)
	}
	if defIconBig := animator.DefaultIconBig(); defIconBig != 0 {
		user32DLL.NewProc("SendMessageW").Call(hwnd, 0x0080, 1, defIconBig)
	}

	// Start continuous taskbar & tray icon animation on launch (calm, smooth 85ms timing)
	animator.StartAnimation("scuba", 85)

	userDataFolder := filepath.Join(os.TempDir(), "DiscordProfileManager_WebView2")
	_ = os.MkdirAll(userDataFolder, 0700)

	chromium := edge.NewChromium()
	chromium.DataPath = userDataFolder
	globalChromium = chromium

	targetURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if chromium.Embed(hwnd) {
		_ = chromium.Show()
		chromium.Focus()

		settings, _ := chromium.GetSettings()
		if settings != nil {
			_ = settings.PutAreDefaultContextMenusEnabled(false)
			_ = settings.PutAreDevToolsEnabled(false)
		}

		chromium.Navigate(targetURL)
		chromium.Resize()

		procShowWindowWin.Call(hwnd, SW_SHOW)
		procUpdateWindowWin.Call(hwnd)

		var msg MSG
		for {
			r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if r == 0 || int32(r) == -1 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}
