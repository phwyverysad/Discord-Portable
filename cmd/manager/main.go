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

func getInstalledProfiles(rootDir string) []string {
	profilesDir := filepath.Join(rootDir, "data", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var installed []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.EqualFold(entry.Name(), "discordptb") {
			installed = append(installed, entry.Name())
		}
	}
	return installed
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

func (s *ManagerServer) handleExecute(w http.ResponseWriter, r *http.Request) {
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
	iconDir := filepath.Join(s.rootDir, "res", "icons")
	targetLauncher := filepath.Join(s.rootDir, "DiscordPTB-Portable.exe")
	if _, err := os.Stat(targetLauncher); os.IsNotExist(err) {
		targetLauncher = filepath.Join(s.rootDir, "Discord-Portable.exe")
	}

	if req.Action == "add" {
		// Install additional profiles
		profilesBaseDir := filepath.Join(s.rootDir, "data", "profiles")
		for _, prof := range req.Profiles {
			profDir := filepath.Join(profilesBaseDir, prof)
			_ = security.ProtectPath(profDir)

			// settings.json
			settingsPath := filepath.Join(profDir, "settings.json")
			_ = os.WriteFile(settingsPath, []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}"), 0644)

			dispName := profileDisplayName(prof)
			shortcutName := fmt.Sprintf("%s.lnk", dispName)
			iconFile := resolveIcon(prof, iconDir)
			aumid := fmt.Sprintf("Discord.PTB.Profile.%s", prof)
			args := fmt.Sprintf("--profile=%s", prof)

			// Desktop shortcut
			desktopLnk := filepath.Join(desktopDir, shortcutName)
			_ = shell.CreateProfileShortcut(desktopLnk, targetLauncher, args, iconFile, aumid, dispName, s.rootDir)

			// Start Menu shortcut (for Windows 11 Shell Taskbar AUMID matching)
			_ = os.MkdirAll(startMenuDir, 0755)
			startMenuLnk := filepath.Join(startMenuDir, shortcutName)
			_ = shell.CreateProfileShortcut(startMenuLnk, targetLauncher, args, iconFile, aumid, dispName, s.rootDir)
		}

		// Ensure context menu keys are registered pointing to this manager
		managerExe, _ := os.Executable()
		_ = shell.RegisterDiscordContextMenu(managerExe, iconDir)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	if req.Action == "uninstall" {
		if req.All {
			// Complete Uninstall: kill processes, remove all shortcuts, unregister context menu, delete directory
			terminateDirectoryProcesses(s.rootDir)

			// Remove desktop shortcuts for all 12 profiles
			for _, p := range allProfiles {
				dispName := profileDisplayName(p.ID)
				_ = os.Remove(filepath.Join(desktopDir, fmt.Sprintf("%s.lnk", dispName)))
			}
			_ = os.Remove(filepath.Join(desktopDir, "Discord-Portable.lnk"))
			_ = os.Remove(filepath.Join(desktopDir, "DiscordPTB-Portable.lnk"))

			// Remove Start Menu directory
			_ = os.RemoveAll(startMenuDir)

			// Unregister Context Menu completely from registry
			shell.UnregisterDiscordContextMenu()

			// Launch self-cleanup script that waits for manager to exit and deletes directory
			cleanupScript := fmt.Sprintf(
				"Start-Sleep -Milliseconds 800; Remove-Item -LiteralPath '%s' -Recurse -Force -ErrorAction SilentlyContinue",
				s.rootDir,
			)
			cleanupCmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", cleanupScript)
			cleanupCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
			_ = cleanupCmd.Start()

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
			return
		}

		// Partial Uninstall: remove chosen profiles only
		for _, prof := range req.Profiles {
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

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	http.Error(w, "Invalid action", http.StatusBadRequest)
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
	flag.Parse()

	action := strings.ToLower(strings.TrimSpace(*actionFlag))
	if action != "uninstall" {
		action = "add"
	}

	exePath, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	rootDir := *dirFlag
	if rootDir == "" {
		rootDir = filepath.Dir(exePath)
	}
	if _, err := os.Stat(filepath.Join(rootDir, "res")); os.IsNotExist(err) {
		// If running from cmd/manager during development
		rootDir = filepath.Dir(filepath.Dir(rootDir))
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
