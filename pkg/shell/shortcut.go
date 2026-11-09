package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var (
	ole32DLL             = syscall.NewLazyDLL("ole32.dll")
	procCoInitialize     = ole32DLL.NewProc("CoInitialize")
	procCoUninitialize   = ole32DLL.NewProc("CoUninitialize")
	procCoCreateInstance = ole32DLL.NewProc("CoCreateInstance")
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type PROPERTYKEY struct {
	Fmtid GUID
	Pid   uint32
}

type PROPVARIANT struct {
	Vt         uint16
	WReserved1 uint16
	WReserved2 uint16
	WReserved3 uint16
	Val        uintptr
	Dummy      [8]byte
}

var (
	clsidShellLink     = GUID{0x00021401, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidShellLinkW      = GUID{0x000214F9, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidPersistFile     = GUID{0x0000010B, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidPropertyStore   = GUID{0x886d8eeb, 0x8cf2, 0x4446, [8]byte{0x8d, 0x02, 0xcd, 0xba, 0x1d, 0xbd, 0xcf, 0x99}}
	pkeyAppUserModelID = PROPERTYKEY{
		Fmtid: GUID{0x9F4C2855, 0x9F79, 0x4B39, [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}},
		Pid:   5,
	}
	pkeyRelaunchCommand = PROPERTYKEY{
		Fmtid: GUID{0x9F4C2855, 0x9F79, 0x4B39, [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}},
		Pid:   2,
	}
	pkeyRelaunchIconResource = PROPERTYKEY{
		Fmtid: GUID{0x9F4C2855, 0x9F79, 0x4B39, [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}},
		Pid:   3,
	}
	pkeyRelaunchDisplayName = PROPERTYKEY{
		Fmtid: GUID{0x9F4C2855, 0x9F79, 0x4B39, [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}},
		Pid:   4,
	}
	shell32DLL                       = syscall.NewLazyDLL("shell32.dll")
	procSHGetPropertyStoreForWindow = shell32DLL.NewProc("SHGetPropertyStoreForWindow")
	procSHChangeNotify               = shell32DLL.NewProc("SHChangeNotify")
)

const (
	SHCNE_ASSOCCHANGED = 0x08000000
	SHCNE_UPDATEITEM   = 0x00002000
	SHCNF_IDLIST       = 0x0000
	SHCNF_PATHW        = 0x0005
)

// RefreshShellIcons invalidates the Windows shell icon cache and forces Explorer to redraw all icons.
func RefreshShellIcons() {
	if procSHChangeNotify.Find() == nil {
		procSHChangeNotify.Call(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, 0, 0)
	}
}

// CreateProfileShortcut creates a Windows shell link (.lnk) with explicit AppUserModelID and dedicated icon.
func CreateProfileShortcut(shortcutPath, targetPath, arguments, iconLocation, aumid, description, workingDir string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	_ = os.MkdirAll(filepath.Dir(shortcutPath), 0755)

	procCoInitialize.Call(0)
	defer procCoUninitialize.Call()

	var pUnknown uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidShellLink)),
		0,
		1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&iidShellLinkW)),
		uintptr(unsafe.Pointer(&pUnknown)),
	)
	if hr != 0 || pUnknown == 0 {
		return fmt.Errorf("CoCreateInstance failed: 0x%x", hr)
	}

	vtableLink := *(**[32]uintptr)(unsafe.Pointer(pUnknown))
	defer func() {
		syscall.SyscallN(vtableLink[2], pUnknown) // Release
	}()

	// 20: SetPath, 11: SetArguments, 9: SetWorkingDirectory, 7: SetDescription, 17: SetIconLocation
	if targetPath != "" {
		p, _ := syscall.UTF16PtrFromString(targetPath)
		syscall.SyscallN(vtableLink[20], pUnknown, uintptr(unsafe.Pointer(p)))
	}
	if arguments != "" {
		p, _ := syscall.UTF16PtrFromString(arguments)
		syscall.SyscallN(vtableLink[11], pUnknown, uintptr(unsafe.Pointer(p)))
	}
	if workingDir != "" {
		p, _ := syscall.UTF16PtrFromString(workingDir)
		syscall.SyscallN(vtableLink[9], pUnknown, uintptr(unsafe.Pointer(p)))
	}
	if description != "" {
		p, _ := syscall.UTF16PtrFromString(description)
		syscall.SyscallN(vtableLink[7], pUnknown, uintptr(unsafe.Pointer(p)))
	}
	if iconLocation != "" {
		p, _ := syscall.UTF16PtrFromString(iconLocation)
		syscall.SyscallN(vtableLink[17], pUnknown, uintptr(unsafe.Pointer(p)), 0)
	}

	// Set AppUserModelID on IPropertyStore
	if aumid != "" {
		var pProp uintptr
		hrQI, _, _ := syscall.SyscallN(vtableLink[0], pUnknown, uintptr(unsafe.Pointer(&iidPropertyStore)), uintptr(unsafe.Pointer(&pProp)))
		if hrQI == 0 && pProp != 0 {
			vtableProp := *(**[16]uintptr)(unsafe.Pointer(pProp))
			aumidUTF16, _ := syscall.UTF16PtrFromString(aumid)
			pv := PROPVARIANT{
				Vt:  31, // VT_LPWSTR
				Val: uintptr(unsafe.Pointer(aumidUTF16)),
			}
			syscall.SyscallN(vtableProp[6], pProp, uintptr(unsafe.Pointer(&pkeyAppUserModelID)), uintptr(unsafe.Pointer(&pv)))
			syscall.SyscallN(vtableProp[7], pProp) // Commit
			syscall.SyscallN(vtableProp[2], pProp) // Release
		}
	}

	// Save to disk via IPersistFile
	var pPersist uintptr
	hrPersist, _, _ := syscall.SyscallN(vtableLink[0], pUnknown, uintptr(unsafe.Pointer(&iidPersistFile)), uintptr(unsafe.Pointer(&pPersist)))
	if hrPersist != 0 || pPersist == 0 {
		return fmt.Errorf("QueryInterface(IPersistFile) failed: 0x%x", hrPersist)
	}
	vtablePersist := *(**[16]uintptr)(unsafe.Pointer(pPersist))
	defer syscall.SyscallN(vtablePersist[2], pPersist) // Release

	destUTF16, _ := syscall.UTF16PtrFromString(shortcutPath)
	hrSave, _, _ := syscall.SyscallN(vtablePersist[6], pPersist, uintptr(unsafe.Pointer(destUTF16)), 1)
	if hrSave != 0 {
		return fmt.Errorf("IPersistFile.Save failed: 0x%x", hrSave)
	}

	if procSHChangeNotify.Find() == nil {
		procSHChangeNotify.Call(SHCNE_UPDATEITEM, SHCNF_PATHW, uintptr(unsafe.Pointer(destUTF16)), 0)
	}

	return nil
}

// SetWindowProfileProperties assigns AUMID, RelaunchIconResource, RelaunchCommand and DisplayName directly to an HWND.
func SetWindowProfileProperties(hwnd uintptr, aumid, iconLocation, relaunchCmd, displayName string) error {
	if procSHGetPropertyStoreForWindow.Find() != nil {
		return nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procCoInitialize.Call(0)
	defer procCoUninitialize.Call()

	var pProp uintptr
	hr, _, _ := procSHGetPropertyStoreForWindow.Call(
		hwnd,
		uintptr(unsafe.Pointer(&iidPropertyStore)),
		uintptr(unsafe.Pointer(&pProp)),
	)
	if hr != 0 || pProp == 0 {
		return fmt.Errorf("SHGetPropertyStoreForWindow failed: 0x%x", hr)
	}

	vtableProp := *(**[16]uintptr)(unsafe.Pointer(pProp))
	defer syscall.SyscallN(vtableProp[2], pProp) // Release

	if aumid != "" {
		aumidUTF16, _ := syscall.UTF16PtrFromString(aumid)
		pv := PROPVARIANT{
			Vt:  31, // VT_LPWSTR
			Val: uintptr(unsafe.Pointer(aumidUTF16)),
		}
		syscall.SyscallN(vtableProp[6], pProp, uintptr(unsafe.Pointer(&pkeyAppUserModelID)), uintptr(unsafe.Pointer(&pv)))
	}

	if iconLocation != "" {
		iconStr := fmt.Sprintf("%s,0", iconLocation)
		iconUTF16, _ := syscall.UTF16PtrFromString(iconStr)
		pv := PROPVARIANT{
			Vt:  31, // VT_LPWSTR
			Val: uintptr(unsafe.Pointer(iconUTF16)),
		}
		syscall.SyscallN(vtableProp[6], pProp, uintptr(unsafe.Pointer(&pkeyRelaunchIconResource)), uintptr(unsafe.Pointer(&pv)))
	}

	if relaunchCmd != "" {
		cmdUTF16, _ := syscall.UTF16PtrFromString(relaunchCmd)
		pv := PROPVARIANT{
			Vt:  31, // VT_LPWSTR
			Val: uintptr(unsafe.Pointer(cmdUTF16)),
		}
		syscall.SyscallN(vtableProp[6], pProp, uintptr(unsafe.Pointer(&pkeyRelaunchCommand)), uintptr(unsafe.Pointer(&pv)))
	}

	if displayName != "" {
		dnUTF16, _ := syscall.UTF16PtrFromString(displayName)
		pv := PROPVARIANT{
			Vt:  31, // VT_LPWSTR
			Val: uintptr(unsafe.Pointer(dnUTF16)),
		}
		syscall.SyscallN(vtableProp[6], pProp, uintptr(unsafe.Pointer(&pkeyRelaunchDisplayName)), uintptr(unsafe.Pointer(&pv)))
	}

	syscall.SyscallN(vtableProp[7], pProp) // Commit
	return nil
}


var canonicalProfiles = map[string]string{
	"1": "01_yellow", "01": "01_yellow", "yellow": "01_yellow", "01_yellow": "01_yellow", "profile_1": "01_yellow", "profile_01": "01_yellow",
	"2": "02_yellow_orange", "02": "02_yellow_orange", "yellow_orange": "02_yellow_orange", "02_yellow_orange": "02_yellow_orange", "profile_2": "02_yellow_orange", "profile_02": "02_yellow_orange",
	"3": "03_orange", "03": "03_orange", "orange": "03_orange", "03_orange": "03_orange", "profile_3": "03_orange", "profile_03": "03_orange",
	"4": "04_red_orange", "04": "04_red_orange", "red_orange": "04_red_orange", "04_red_orange": "04_red_orange", "profile_4": "04_red_orange", "profile_04": "04_red_orange",
	"5": "05_red", "05": "05_red", "red": "05_red", "05_red": "05_red", "profile_5": "05_red", "profile_05": "05_red",
	"6": "06_red_violet", "06": "06_red_violet", "red_violet": "06_red_violet", "06_red_violet": "06_red_violet", "profile_6": "06_red_violet", "profile_06": "06_red_violet",
	"7": "07_violet", "07": "07_violet", "violet": "07_violet", "07_violet": "07_violet", "profile_7": "07_violet", "profile_07": "07_violet",
	"8": "08_blue_violet", "08": "08_blue_violet", "blue_violet": "08_blue_violet", "08_blue_violet": "08_blue_violet", "profile_8": "08_blue_violet", "profile_08": "08_blue_violet",
	"9": "09_blue", "09": "09_blue", "blue": "09_blue", "09_blue": "09_blue", "profile_9": "09_blue", "profile_09": "09_blue",
	"10": "10_blue_green", "blue_green": "10_blue_green", "10_blue_green": "10_blue_green", "profile_10": "10_blue_green",
	"11": "11_green", "green": "11_green", "11_green": "11_green", "profile_11": "11_green",
	"12": "12_yellow_green", "yellow_green": "12_yellow_green", "12_yellow_green": "12_yellow_green", "profile_12": "12_yellow_green",
}

// CanonicalProfile returns the canonical profile ID (e.g. "04_red_orange") from any input variant.
func CanonicalProfile(id string) string {
	cleaned := strings.ToLower(strings.TrimSpace(id))
	if canon, ok := canonicalProfiles[cleaned]; ok {
		return canon
	}
	return "01_yellow"
}
