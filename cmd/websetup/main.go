package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"github.com/portapps/discord-ptb-portable/pkg/trayanim"
	"github.com/portapps/portapps/v3/pkg/log"
	"golang.org/x/sys/windows"
)

var (
	kernel32Main           = syscall.NewLazyDLL("kernel32.dll")
	user32Main             = syscall.NewLazyDLL("user32.dll")
	gdi32Main              = syscall.NewLazyDLL("gdi32.dll")
	procFreeConsole        = kernel32Main.NewProc("FreeConsole")
	procRegisterClassExW   = user32Main.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32Main.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32Main.NewProc("DefWindowProcW")
	procUpdateWindowWin    = user32Main.NewProc("UpdateWindow")
	procGetMessageW        = user32Main.NewProc("GetMessageW")
	procTranslateMessage   = user32Main.NewProc("TranslateMessage")
	procDispatchMessageW   = user32Main.NewProc("DispatchMessageW")
	procPostQuitMessage    = user32Main.NewProc("PostQuitMessage")
	procGetSystemMetrics   = user32Main.NewProc("GetSystemMetrics")
	procLoadCursorW        = user32Main.NewProc("LoadCursorW")
	procSetCursor          = user32Main.NewProc("SetCursor")
	procDestroyWindow      = user32Main.NewProc("DestroyWindow")
	procCreateRoundRectRgn = gdi32Main.NewProc("CreateRoundRectRgn")
	procSetWindowRgn        = user32Main.NewProc("SetWindowRgn")
	procFindWindowW         = user32Main.NewProc("FindWindowW")
	procSetForegroundWindow = user32Main.NewProc("SetForegroundWindow")
	procSendMessageW        = user32Main.NewProc("SendMessageW")
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

const (
	WS_POPUP        = 0x80000000
	WS_CLIPCHILDREN = 0x02000000
	WS_EX_APPWINDOW = 0x00040000
	CS_DROPSHADOW   = 0x00020000
	WM_DESTROY      = 0x0002
	WM_MOVE         = 0x0003
	WM_SIZE         = 0x0005
	WM_CLOSE        = 0x0010
	WM_SETCURSOR    = 0x0020
	SW_SHOW         = 5
	SM_CXSCREEN     = 0
	SM_CYSCREEN     = 1
)

var (
	globalChromium *edge.Chromium
	defaultCursor  uintptr
)

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
	// 1. Immediately detach from console if any attached
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

	// If invoked with profile flags, forward to DiscordPTB-Portable.exe
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--profile=") || arg == "--multi-instance" {
			launcherPath := filepath.Join(rootPath, "DiscordPTB-Portable.exe")
			if _, err := os.Stat(launcherPath); err == nil {
				cmd := exec.Command(launcherPath, os.Args[1:]...)
				cmd.Dir = rootPath
				_ = cmd.Start()
				return
			}
		}
	}

	// 1.1 Single-instance protection: prevent duplicate installer windows
	mutexName, _ := windows.UTF16PtrFromString("Global\\DiscordPTBPortable_WebSetup_SingleInstance")
	hMutex, errMutex := windows.CreateMutex(nil, false, mutexName)
	if errMutex != nil || windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		className, _ := windows.UTF16PtrFromString("DiscordFramelessWebSetup")
		existingHwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
		if existingHwnd != 0 {
			procShowWindowWin.Call(existingHwnd, SW_SHOW)
			procSetForegroundWindow.Call(existingHwnd)
		}
		if hMutex != 0 {
			windows.CloseHandle(hMutex)
		}
		return
	}
	defer windows.CloseHandle(hMutex)

	server, err := NewSetupServer(rootPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed initializing WebSetup server")
	}

	server.Start()
	defer server.Stop()

	targetURL := server.URL()
	log.Info().Str("url", targetURL).Msg("WebSetup Server listening")

	// 2. Register native frameless window class
	var hinst windows.Handle
	windows.GetModuleHandleEx(0, nil, &hinst)

	cur, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	defaultCursor = cur

	className, _ := windows.UTF16PtrFromString("DiscordFramelessWebSetup")
	titleName, _ := windows.UTF16PtrFromString("Discord PTB Portable - WebSetup")
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_DROPSHADOW,
		LpfnWndProc:   windows.NewCallback(framelessWndProc),
		HInstance:     hinst,
		HCursor:       windows.Handle(cur),
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// 3. Compact window dimensions: 800 x 570 (centered)
	width := int32(800)
	height := int32(615)
	scrW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	scrH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	x := (int32(scrW) - width) / 2
	y := (int32(scrH) - height) / 2

	// 4. Create window as WS_POPUP with WS_EX_APPWINDOW (born frameless, never flashes standard title bar)
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

	if hwnd != 0 {
		// Apply smooth rounded window region (radius 14px -> diameter 28)
		rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, uintptr(width+1), uintptr(height+1), 28, 28)
		if rgn != 0 {
			procSetWindowRgn.Call(hwnd, rgn, 1)
		}

		server.SetWindowHwnd(hwnd)

		animator := trayanim.GetGlobalAnimator()
		_ = animator.Init(hwnd, "Discord PTB Portable - Setup")
		_ = animator.AddTray()
		defer animator.RemoveTray()

		if defIcon := animator.DefaultIcon(); defIcon != 0 {
			procSendMessageW.Call(hwnd, 0x0080, 0, defIcon)
		}
		if defIconBig := animator.DefaultIconBig(); defIconBig != 0 {
			procSendMessageW.Call(hwnd, 0x0080, 1, defIconBig)
		}

		// Start continuous taskbar & tray icon animation on launch (calm, smooth 85ms timing)
		animator.StartAnimation("scuba", 85)

		userDataFolder := filepath.Join(os.Getenv("TEMP"), "discord_websetup_wv2")
		_ = os.MkdirAll(userDataFolder, 0700)

		chromium := edge.NewChromium()
		chromium.DataPath = userDataFolder
		globalChromium = chromium

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

			// Show frameless window only when fully initialized
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
			return
		}
		// Destroy unembedded window before fallback
		procDestroyWindow.Call(hwnd)
	}

	// Fallback to browser if WebView2 runtime is completely missing
	browserExe := findBrowser()
	if browserExe != "" {
		tempBrowserData := filepath.Join(os.Getenv("TEMP"), "discord_websetup_cache")
		_ = os.MkdirAll(tempBrowserData, 0700)
		defer os.RemoveAll(tempBrowserData)

		appArgs := []string{
			fmt.Sprintf("--app=%s", targetURL),
			"--window-size=800,570",
			fmt.Sprintf("--user-data-dir=%s", tempBrowserData),
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
