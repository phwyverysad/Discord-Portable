package sfx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	comctl32             = syscall.NewLazyDLL("comctl32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procSendMessageW     = user32.NewProc("SendMessageW")
	procSetWindowTextW   = user32.NewProc("SetWindowTextW")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procPeekMessageW     = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
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

type ProgressDialog struct {
	hwnd     uintptr
	lblTitle uintptr
	lblDesc  uintptr
	pbar     uintptr
}

func NewProgressDialog(title, initialStatus string) *ProgressDialog {
	runtime.LockOSThread()

	className, _ := syscall.UTF16PtrFromString("DiscordPortableExtractSplash")
	wndProc := syscall.NewCallback(func(hwnd, msg, wparam, lparam uintptr) uintptr {
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wparam, lparam)
		return ret
	})

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         0x0003, // CS_HREDRAW | CS_VREDRAW
		LpfnWndProc:   wndProc,
		HbrBackground: 6, // COLOR_WINDOW
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	const width = 440
	const height = 150
	screenWidth, _, _ := procGetSystemMetrics.Call(0)
	screenHeight, _, _ := procGetSystemMetrics.Call(1)
	x := (int(screenWidth) - width) / 2
	y := (int(screenHeight) - height) / 2

	wTitle, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := procCreateWindowExW.Call(
		0x00000008, // WS_EX_TOPMOST
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(wTitle)),
		0x00C00000|0x00080000|0x10000000, // WS_CAPTION | WS_SYSMENU | WS_VISIBLE
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		0, 0, 0, 0,
	)

	// Create Font
	fontName, _ := syscall.UTF16PtrFromString("Segoe UI")
	hFont, _, _ := procCreateFontW.Call(
		15, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 2, 0, uintptr(unsafe.Pointer(fontName)),
	)
	hFontBold, _, _ := procCreateFontW.Call(
		17, 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 2, 0, uintptr(unsafe.Pointer(fontName)),
	)

	staticClass, _ := syscall.UTF16PtrFromString("STATIC")
	titleText, _ := syscall.UTF16PtrFromString("Discord Portable")
	lblTitle, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(titleText)),
		0x50000000, 24, 16, 380, 24, hwnd, 0, 0, 0,
	)
	procSendMessageW.Call(lblTitle, 0x0030, hFontBold, 1)

	descText, _ := syscall.UTF16PtrFromString(initialStatus)
	lblDesc, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(staticClass)), uintptr(unsafe.Pointer(descText)),
		0x50000000, 24, 44, 380, 20, hwnd, 0, 0, 0,
	)
	procSendMessageW.Call(lblDesc, 0x0030, hFont, 1)

	pbarClass, _ := syscall.UTF16PtrFromString("msctls_progress32")
	pbar, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(pbarClass)), 0,
		0x50000000, 24, 72, 376, 18, hwnd, 0, 0, 0,
	)

	// Set progress bar range 0 - 100
	procSendMessageW.Call(pbar, 0x0401, 0, 100<<16) // PBM_SETRANGE

	procShowWindow.Call(hwnd, 5) // SW_SHOW
	procUpdateWindow.Call(hwnd)

	return &ProgressDialog{
		hwnd:     hwnd,
		lblTitle: lblTitle,
		lblDesc:  lblDesc,
		pbar:     pbar,
	}
}

func (d *ProgressDialog) SetProgress(percent int, status string) {
	if d.hwnd == 0 {
		return
	}
	procSendMessageW.Call(d.pbar, 0x0402, uintptr(percent), 0) // PBM_SETPOS
	if status != "" {
		st, _ := syscall.UTF16PtrFromString(status)
		procSetWindowTextW.Call(d.lblDesc, uintptr(unsafe.Pointer(st)))
	}

	// Message pump to keep window responsive
	var msg MSG
	for {
		ret, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 1)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (d *ProgressDialog) Close() {
	if d.hwnd != 0 {
		procDestroyWindow.Call(d.hwnd)
		d.hwnd = 0
		runtime.UnlockOSThread()
	}
}

// EnsurePortableRuntime verifies if Discord runtime exists. If missing, it unpacks the embedded payload.
func EnsurePortableRuntime(currentRoot string) string {
	// 1. Direct check in current directory
	directApp := filepath.Join(currentRoot, "app", "app-1.0.1231", "DiscordPTB.exe")
	if _, err := os.Stat(directApp); err == nil {
		return currentRoot
	}

	exePath, err := os.Executable()
	if err != nil {
		exePath, _ = filepath.Abs(os.Args[0])
	}
	exeDir := filepath.Dir(exePath)

	// Check if already extracted in exe directory
	if exeDir != currentRoot {
		exeApp := filepath.Join(exeDir, "app", "app-1.0.1231", "DiscordPTB.exe")
		if _, err := os.Stat(exeApp); err == nil {
			return exeDir
		}
	}

	// Determine clean extraction target directory
	targetDir := exeDir
	lowerDir := strings.ToLower(exeDir)
	if strings.Contains(lowerDir, "desktop") || strings.Contains(lowerDir, "downloads") {
		// Avoid cluttering user Desktop/Downloads with hundreds of files
		targetDir = filepath.Join(exeDir, "Discord-PTB-Portable")
	}

	// Check if already extracted in target directory
	targetApp := filepath.Join(targetDir, "app", "app-1.0.1231", "DiscordPTB.exe")
	if _, err := os.Stat(targetApp); err == nil {
		return targetDir
	}

	// Check if the executable contains an embedded / appended zip payload
	r, err := zip.OpenReader(exePath)
	if err != nil {
		// Fallback check: check standard default installation C:\Discord-PTB-Portable
		cApp := filepath.Join(`C:\Discord-PTB-Portable`, "app", "app-1.0.1231", "DiscordPTB.exe")
		if _, err := os.Stat(cApp); err == nil {
			return `C:\Discord-PTB-Portable`
		}
		return currentRoot
	}
	defer r.Close()

	// Show progress window while unpacking
	dlg := NewProgressDialog("Discord Portable", "กำลังเตรียมความพร้อม Discord Portable... กรุณารอสักครู่")
	defer dlg.Close()

	totalFiles := len(r.File)
	buf := make([]byte, 256*1024) // 256KB copy buffer
	lastUpdate := time.Now()

	for idx, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") {
			continue
		}
		destPath := filepath.Join(targetDir, cleanName)

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(destPath, 0755)
			continue
		}

		_ = os.MkdirAll(filepath.Dir(destPath), 0755)
		rc, err := f.Open()
		if err != nil {
			continue
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err == nil {
			_, _ = io.CopyBuffer(outFile, rc, buf)
			outFile.Close()
		}
		rc.Close()

		if time.Since(lastUpdate) > 80*time.Millisecond || idx == totalFiles-1 {
			pct := int((float64(idx+1) / float64(totalFiles)) * 100)
			dlg.SetProgress(pct, fmt.Sprintf("กำลังเตรียมไฟล์ Discord... %d%%", pct))
			lastUpdate = time.Now()
		}
	}

	dlg.SetProgress(100, "พร้อมเปิดใช้งาน...")
	time.Sleep(300 * time.Millisecond)

	return targetDir
}
