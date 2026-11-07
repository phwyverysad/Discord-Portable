package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
	"github.com/portapps/portapps/v3/pkg/log"
)

var (
	kernel32Main              = syscall.NewLazyDLL("kernel32.dll")
	user32Main                = syscall.NewLazyDLL("user32.dll")
	gdi32Main                 = syscall.NewLazyDLL("gdi32.dll")
	dwmapiMain                = syscall.NewLazyDLL("dwmapi.dll")
	procFreeConsole           = kernel32Main.NewProc("FreeConsole")
	procGetWindowLongW        = user32Main.NewProc("GetWindowLongW")
	procSetWindowLongW        = user32Main.NewProc("SetWindowLongW")
	procSetWindowPos          = user32Main.NewProc("SetWindowPos")
	procSetWindowRgn          = user32Main.NewProc("SetWindowRgn")
	procCreateRoundRectRgn    = gdi32Main.NewProc("CreateRoundRectRgn")
	procDwmSetWindowAttribute = dwmapiMain.NewProc("DwmSetWindowAttribute")
	procShowWindow            = user32Main.NewProc("ShowWindow")
)

const (
	GWL_STYLE                = ^uintptr(15) // -16
	WS_POPUP                 = 0x80000000
	WS_VISIBLE               = 0x10000000
	WS_CLIPCHILDREN          = 0x02000000
	SWP_NOMOVE               = 0x0002
	SWP_NOZORDER             = 0x0004
	SWP_FRAMECHANGED         = 0x0020
	DWMWA_WINDOW_CORNER_PREF = 33
	DWMWCP_ROUND             = 2
)

func findBrowser() string {
	candidates := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Microsoft\Edge\Application\msedge.exe`),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func main() {
	// Detach and hide any attached console window
	procFreeConsole.Call()

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot get executable path")
	}

	rootPath := filepath.Dir(exePath)
	if _, err := os.Stat(filepath.Join(rootPath, "res")); os.IsNotExist(err) {
		// If running under cmd/websetup, look two levels up
		rootPath = filepath.Dir(filepath.Dir(rootPath))
	}

	server, err := NewSetupServer(rootPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed initializing WebSetup server")
	}

	server.Start()
	defer server.Stop()

	targetURL := server.URL()
	log.Info().Str("url", targetURL).Msg("WebSetup Server listening")

	// Native frameless WebView2 window
	userDataFolder := filepath.Join(os.Getenv("TEMP"), "discord_websetup_webview2")
	_ = os.MkdirAll(userDataFolder, 0700)
	defer os.RemoveAll(userDataFolder)

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		DataPath:  userDataFolder,
		WindowOptions: webview2.WindowOptions{
			Title:  "Discord PTB Portable - WebSetup",
			Width:  860,
			Height: 640,
			Center: true,
		},
	})

	if w != nil {
		hwnd := uintptr(w.Window())
		server.SetWindowHwnd(hwnd)

		// Set 100% frameless popup style (No OS title bar, no OS borders)
		procSetWindowLongW.Call(hwnd, GWL_STYLE, WS_POPUP|WS_VISIBLE|WS_CLIPCHILDREN)
		procSetWindowPos.Call(hwnd, 0, 0, 0, 860, 640, SWP_NOMOVE|SWP_NOZORDER|SWP_FRAMECHANGED)

		// Enable Windows 11 rounded corners
		pref := uint32(DWMWCP_ROUND)
		procDwmSetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREF, uintptr(unsafe.Pointer(&pref)), 4)

		// Apply exact rounded clipping
		rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, 860, 640, 24, 24)
		if rgn != 0 {
			procSetWindowRgn.Call(hwnd, rgn, 1)
		}

		w.Navigate(targetURL)
		w.Run()
		return
	}

	// Fallback to browser if WebView2 is missing
	browserExe := findBrowser()
	if browserExe != "" {
		appArgs := []string{
			fmt.Sprintf("--app=%s", targetURL),
			"--window-size=860,640",
			fmt.Sprintf("--user-data-dir=%s", userDataFolder),
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-extensions",
		}
		cmd := exec.Command(browserExe, appArgs...)
		_ = cmd.Run()
	} else {
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
		_ = cmd.Start()
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
