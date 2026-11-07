package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/portapps/portapps/v3/pkg/log"
)

var (
	user32Main                = syscall.NewLazyDLL("user32.dll")
	gdi32Main                 = syscall.NewLazyDLL("gdi32.dll")
	dwmapiMain                = syscall.NewLazyDLL("dwmapi.dll")
	procFindWindowW           = user32Main.NewProc("FindWindowW")
	procGetWindowLongW        = user32Main.NewProc("GetWindowLongW")
	procSetWindowLongW        = user32Main.NewProc("SetWindowLongW")
	procSetWindowPos          = user32Main.NewProc("SetWindowPos")
	procSetWindowRgn          = user32Main.NewProc("SetWindowRgn")
	procCreateRoundRectRgn    = gdi32Main.NewProc("CreateRoundRectRgn")
	procDwmSetWindowAttribute = dwmapiMain.NewProc("DwmSetWindowAttribute")
)

const (
	GWL_STYLE                = ^uintptr(15) // -16
	WS_CAPTION               = 0x00C00000
	WS_THICKFRAME            = 0x00040000
	WS_MINIMIZEBOX           = 0x00020000
	WS_MAXIMIZEBOX           = 0x00010000
	WS_SYSMENU               = 0x00080000
	WS_BORDER                = 0x00800000
	WS_DLGFRAME              = 0x00400000
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

	browserExe := findBrowser()
	tempBrowserData := filepath.Join(os.Getenv("TEMP"), "discord_websetup_cache")
	_ = os.MkdirAll(tempBrowserData, 0700)
	defer os.RemoveAll(tempBrowserData)

	if browserExe != "" {
		appArgs := []string{
			fmt.Sprintf("--app=%s", targetURL),
			"--window-size=860,640",
			fmt.Sprintf("--user-data-dir=%s", tempBrowserData),
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-extensions",
		}
		cmd := exec.Command(browserExe, appArgs...)
		log.Info().Str("browser", browserExe).Msg("Spawning WebSetup window in Frameless App mode")

		if err := cmd.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed launching browser window")
		}

		// Asynchronously locate window and apply frameless rounded styling
		go func() {
			titlePtr, _ := syscall.UTF16PtrFromString("Discord PTB Portable - WebSetup")
			var hwnd uintptr
			for i := 0; i < 40; i++ {
				time.Sleep(100 * time.Millisecond)
				r, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
				if r != 0 {
					hwnd = r
					break
				}
			}

			if hwnd != 0 {
				server.SetWindowHwnd(hwnd)

				// Strip title bar, caption, and OS standard borders (Red Box removal)
				curStyle, _, _ := procGetWindowLongW.Call(hwnd, GWL_STYLE)
				newStyle := curStyle &^ (WS_CAPTION | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX | WS_SYSMENU | WS_BORDER | WS_DLGFRAME)
				procSetWindowLongW.Call(hwnd, GWL_STYLE, newStyle)

				// Refresh frame and set exact dimensions
				procSetWindowPos.Call(hwnd, 0, 0, 0, 860, 640, SWP_NOMOVE|SWP_NOZORDER|SWP_FRAMECHANGED)

				// Enable Windows 11 Rounded Corners (DWM)
				pref := uint32(DWMWCP_ROUND)
				procDwmSetWindowAttribute.Call(hwnd, DWMWA_WINDOW_CORNER_PREF, uintptr(unsafe.Pointer(&pref)), 4)

				// Clip to exact smooth rounded rectangle
				rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, 860, 640, 24, 24)
				if rgn != 0 {
					procSetWindowRgn.Call(hwnd, rgn, 1)
				}

				log.Info().Str("hwnd", fmt.Sprintf("0x%X", hwnd)).Msg("Custom frameless curved window applied successfully")
			}
		}()

		if err := cmd.Wait(); err != nil {
			log.Warn().Err(err).Msg("Browser closed or exited")
		}
	} else {
		// Fallback to default browser
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
		_ = cmd.Start()
		// Keep server alive
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
