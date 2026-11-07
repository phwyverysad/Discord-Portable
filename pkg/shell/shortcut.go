package shell

import (
	"fmt"
	"os"
	"path/filepath"
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
)

// CreateProfileShortcut creates a Windows shell link (.lnk) with explicit AppUserModelID and dedicated icon.
func CreateProfileShortcut(shortcutPath, targetPath, arguments, iconLocation, aumid, description, workingDir string) error {
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

	return nil
}
