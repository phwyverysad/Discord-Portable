package trayanim

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/*
var embeddedAssets embed.FS

var (
	shell32DLL           = syscall.NewLazyDLL("shell32.dll")
	user32DLL            = syscall.NewLazyDLL("user32.dll")
	procShellNotifyIconW = shell32DLL.NewProc("Shell_NotifyIconW")
	procLoadImageW       = user32DLL.NewProc("LoadImageW")
	procDestroyIcon      = user32DLL.NewProc("DestroyIcon")
	procSendMessageW     = user32DLL.NewProc("SendMessageW")
)

const (
	NIM_ADD         = 0x00000000
	NIM_MODIFY      = 0x00000001
	NIM_DELETE      = 0x00000002
	NIF_MESSAGE     = 0x00000001
	NIF_ICON        = 0x00000002
	NIF_TIP         = 0x00000004
	NIF_INFO        = 0x00000010
	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x00000010
	WM_SETICON      = 0x0080
	ICON_SMALL      = 0
	ICON_BIG        = 1
	WM_USER         = 0x0400
	WM_TRAYICON     = WM_USER + 200
)

type NOTIFYICONDATAW struct {
	CbSize            uint32
	HWnd              uintptr
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             uintptr
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UVersionOrTimeout uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          [16]byte
	HBalloonIcon      uintptr
}

// Frame holds both small (tray/titlebar) and big (taskbar/Alt-Tab) HICON handles.
type Frame struct {
	Small uintptr // 16x16 / 24x24
	Big   uintptr // 32x32 / 48x48
}

// Animator manages animated tray icons and window taskbar/titlebar synchronization.
type Animator struct {
	hwnd           uintptr
	uid            uint32
	defaultIcon    Frame
	currentAnim    string
	frames         map[string][]Frame
	ticker         *time.Ticker
	stopChan       chan struct{}
	mu             sync.Mutex
	isTrayActive   bool
	currentTip     string
	syncWindowIcon bool
}

var (
	GlobalAnimator *Animator
	animatorOnce   sync.Once
)

// GetGlobalAnimator returns the singleton animator instance.
func GetGlobalAnimator() *Animator {
	animatorOnce.Do(func() {
		GlobalAnimator = NewAnimator(1001)
	})
	return GlobalAnimator
}

// NewAnimator creates an Animator with a specific tray UID.
func NewAnimator(uid uint32) *Animator {
	return &Animator{
		uid:            uid,
		frames:         make(map[string][]Frame),
		syncWindowIcon: true,
	}
}

// Init extracts embedded animation frames into a local cache directory and loads HICON handles.
func (a *Animator) Init(hwnd uintptr, defaultTooltip string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.hwnd = hwnd
	a.currentTip = defaultTooltip

	cacheDir := filepath.Join(os.TempDir(), "discord_trayanim_cache")
	_ = os.MkdirAll(cacheDir, 0755)

	categories := []string{"scuba", "sync", "recording"}
	for _, cat := range categories {
		embedSubDir := path.Join("assets", cat)
		entries, err := embeddedAssets.ReadDir(embedSubDir)
		if err != nil {
			continue
		}

		var fileNames []string
		for _, e := range entries {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".ico") {
				fileNames = append(fileNames, e.Name())
			}
		}
		sort.Strings(fileNames)

		var frameList []Frame
		for _, fn := range fileNames {
			data, err := embeddedAssets.ReadFile(path.Join(embedSubDir, fn))
			if err != nil {
				continue
			}
			tempIcoPath := filepath.Join(cacheDir, fmt.Sprintf("%s_%s", cat, fn))
			_ = os.WriteFile(tempIcoPath, data, 0644)

			pPath, _ := syscall.UTF16PtrFromString(tempIcoPath)
			hSmall, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(pPath)), IMAGE_ICON, 16, 16, LR_LOADFROMFILE)
			hBig, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(pPath)), IMAGE_ICON, 32, 32, LR_LOADFROMFILE)

			if hSmall != 0 {
				if hBig == 0 {
					hBig = hSmall
				}
				frameList = append(frameList, Frame{Small: hSmall, Big: hBig})
			}
		}
		a.frames[cat] = frameList
	}

	// Load default master icon
	defData, err := embeddedAssets.ReadFile("assets/default.ico")
	if err == nil {
		defPath := filepath.Join(cacheDir, "default.ico")
		_ = os.WriteFile(defPath, defData, 0644)
		pPath, _ := syscall.UTF16PtrFromString(defPath)
		hSmall, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(pPath)), IMAGE_ICON, 16, 16, LR_LOADFROMFILE)
		hBig, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(pPath)), IMAGE_ICON, 32, 32, LR_LOADFROMFILE)
		if hSmall != 0 {
			if hBig == 0 {
				hBig = hSmall
			}
			a.defaultIcon = Frame{Small: hSmall, Big: hBig}
		}
	}

	if a.defaultIcon.Small == 0 && len(a.frames["scuba"]) > 0 {
		a.defaultIcon = a.frames["scuba"][0]
	}

	return nil
}

// AddTray registers the icon in the Windows Notification Area.
func (a *Animator) AddTray() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isTrayActive || a.hwnd == 0 {
		return nil
	}

	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = a.hwnd
	nid.UID = a.uid
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAYICON
	nid.HIcon = a.defaultIcon.Small

	copyTip(&nid.SzTip, a.currentTip)

	r, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if r != 0 {
		a.isTrayActive = true
		return nil
	}
	return fmt.Errorf("Shell_NotifyIconW(NIM_ADD) failed")
}

// StartAnimation begins cycling icon frames for the specified animation type.
// Supported animType: "scuba" (dance sticker), "sync" (rotating spinner), "recording" (audio wave pulse).
func (a *Animator) StartAnimation(animType string, intervalMs int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	animType = strings.ToLower(strings.TrimSpace(animType))
	frames, ok := a.frames[animType]
	if !ok || len(frames) == 0 {
		animType = "scuba"
		frames = a.frames["scuba"]
		if len(frames) == 0 {
			return
		}
	}

	// Stop any currently running ticker
	if a.ticker != nil {
		a.ticker.Stop()
		close(a.stopChan)
		a.ticker = nil
	}

	if intervalMs <= 0 {
		switch animType {
		case "sync":
			intervalMs = 45 // 22 fps rotating spinner
		case "recording":
			intervalMs = 80 // 12 fps pulsing audio
		default:
			intervalMs = 60 // ~16 fps scuba dance
		}
	}

	a.currentAnim = animType
	a.ticker = time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
	a.stopChan = make(chan struct{})

	ticker := a.ticker
	stopChan := a.stopChan
	hwnd := a.hwnd
	uid := a.uid
	syncWnd := a.syncWindowIcon

	go func(fList []Frame) {
		idx := 0
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				idx = (idx + 1) % len(fList)
				f := fList[idx]

				// 1. Update Tray icon
				var nid NOTIFYICONDATAW
				nid.CbSize = uint32(unsafe.Sizeof(nid))
				nid.HWnd = hwnd
				nid.UID = uid
				nid.UFlags = NIF_ICON
				nid.HIcon = f.Small
				procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))

				// 2. Update Window Taskbar and Titlebar icon if enabled
				if syncWnd && hwnd != 0 {
					procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, f.Big)
					procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, f.Small)
				}
			}
		}
	}(frames)
}

// StopAnimation stops the animation and restores the default idle icon.
func (a *Animator) StopAnimation() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ticker != nil {
		a.ticker.Stop()
		close(a.stopChan)
		a.ticker = nil
	}
	a.currentAnim = ""

	if a.hwnd != 0 && a.defaultIcon.Small != 0 {
		var nid NOTIFYICONDATAW
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		nid.HWnd = a.hwnd
		nid.UID = a.uid
		nid.UFlags = NIF_ICON
		nid.HIcon = a.defaultIcon.Small
		procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))

		if a.syncWindowIcon {
			if a.defaultIcon.Big != 0 {
				procSendMessageW.Call(a.hwnd, WM_SETICON, ICON_BIG, a.defaultIcon.Big)
			}
			procSendMessageW.Call(a.hwnd, WM_SETICON, ICON_SMALL, a.defaultIcon.Small)
		}
	}
}

// SetTooltip updates the tooltip text shown when hovering over the tray icon.
func (a *Animator) SetTooltip(tip string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.currentTip = tip
	if a.hwnd != 0 {
		var nid NOTIFYICONDATAW
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		nid.HWnd = a.hwnd
		nid.UID = a.uid
		nid.UFlags = NIF_TIP
		copyTip(&nid.SzTip, tip)
		procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	}
}

// RemoveTray removes the icon from the Windows system notification area.
func (a *Animator) RemoveTray() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ticker != nil {
		a.ticker.Stop()
		close(a.stopChan)
		a.ticker = nil
	}

	if a.isTrayActive && a.hwnd != 0 {
		var nid NOTIFYICONDATAW
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		nid.HWnd = a.hwnd
		nid.UID = a.uid
		procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
		a.isTrayActive = false
	}
}

// FrameCount returns the number of frames available for a specific animation.
func (a *Animator) FrameCount(animType string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.frames[animType])
}

// DefaultIcon returns the loaded default small HICON handle (for tray & titlebar).
func (a *Animator) DefaultIcon() uintptr {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.defaultIcon.Small
}

// DefaultIconBig returns the loaded default large HICON handle (for taskbar & Alt-Tab).
func (a *Animator) DefaultIconBig() uintptr {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.defaultIcon.Big != 0 {
		return a.defaultIcon.Big
	}
	return a.defaultIcon.Small
}

func copyTip(dest *[128]uint16, src string) {
	utf16, _ := syscall.UTF16FromString(src)
	copy(dest[:], utf16)
	if len(utf16) < len(dest) {
		dest[len(utf16)] = 0
	}
}
