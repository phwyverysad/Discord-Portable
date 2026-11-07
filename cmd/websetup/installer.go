package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/portapps/discord-ptb-portable/pkg/security"
	"github.com/portapps/portapps/v3/pkg/shortcut"
)

type InstallConfig struct {
	InstallDir      string   `json:"installDir"`
	InstallMode     string   `json:"installMode"` // "pc" or "usb"
	EcoMode         bool     `json:"ecoMode"`
	Profiles        []string `json:"profiles"`
	LaunchOnFinish  bool     `json:"launchOnFinish"`
	CreateShortcuts bool     `json:"createShortcuts"`
}

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}

type InstallStatus struct {
	Percent int    `json:"percent"`
	Message string `json:"message"`
	Done    bool   `json:"done"`
	Error   string `json:"error,omitempty"`
}

var (
	currentStatus       = InstallStatus{Percent: 0, Message: "Ready"}
	statusMutex         sync.RWMutex
	installInProgress   bool
	installRunningMutex sync.Mutex
)

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

func resolveSetupIcon(profileID, iconDir string) string {
	id := strings.TrimSpace(profileID)
	directMatch := filepath.Join(iconDir, id+".ico")
	if _, err := os.Stat(directMatch); err == nil {
		return directMatch
	}

	entries, err := os.ReadDir(iconDir)
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(strings.ToLower(name), ".ico") {
				if strings.HasPrefix(name, fmt.Sprintf("%02s_", id)) || strings.HasPrefix(name, id+"_") {
					return filepath.Join(iconDir, name)
				}
			}
		}
	}
	return filepath.Join(iconDir, "01_yellow.ico")
}

func updateStatus(percent int, msg string, done bool, errStr string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	currentStatus = InstallStatus{
		Percent: percent,
		Message: msg,
		Done:    done,
		Error:   errStr,
	}
}

func getStatus() InstallStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()
	return currentStatus
}

// DownloadFile downloads a URL to a local destination file with progress.
func downloadFile(url, destPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// ExecuteInstallation runs the full installation flow based on InstallConfig.
func ExecuteInstallation(cfg InstallConfig, rootSourcePath string) error {
	installRunningMutex.Lock()
	if installInProgress {
		installRunningMutex.Unlock()
		return fmt.Errorf("installation already in progress")
	}
	installInProgress = true
	installRunningMutex.Unlock()
	defer func() {
		installRunningMutex.Lock()
		installInProgress = false
		installRunningMutex.Unlock()
	}()

	updateStatus(5, "กำลังเตรียมโฟลเดอร์ติดตั้ง...", false, "")

	// 1. Terminate any running Discord instances to prevent file locking (hidden in background)
	_ = hiddenCommand("taskkill", "/F", "/IM", "DiscordPTB.exe").Run()
	_ = hiddenCommand("taskkill", "/F", "/IM", "DiscordPTB-Portable.exe").Run()
	_ = hiddenCommand("taskkill", "/F", "/IM", "Discord-Portable.exe").Run()

	if cfg.InstallDir == "" {
		if cfg.InstallMode == "usb" {
			cfg.InstallDir = filepath.Join(".", "Discord-Portable")
		} else {
			cfg.InstallDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Discord-Portable")
		}
	}

	if err := os.MkdirAll(cfg.InstallDir, 0755); err != nil {
		updateStatus(0, "Error", false, err.Error())
		return err
	}

	updateStatus(15, "กำลังตั้งค่าความปลอดภัย...", false, "")
	profilesDir := filepath.Join(cfg.InstallDir, "data", "profiles")
	if err := os.MkdirAll(profilesDir, 0700); err != nil {
		return err
	}

	// Copy Icons (all 12 colors)
	updateStatus(25, "กำลังคัดลอกชุดไอคอน...", false, "")
	srcIcons := filepath.Join(rootSourcePath, "res", "icons")
	dstIcons := filepath.Join(cfg.InstallDir, "res", "icons")
	if _, err := os.Stat(srcIcons); err == nil {
		_ = copyDir(srcIcons, dstIcons)
	}

	// Setup Launcher Binary in target
	updateStatus(35, "กำลังติดตั้งตัวเปิดโปรแกรม...", false, "")
	srcLauncher := filepath.Join(rootSourcePath, "DiscordPTB-Portable.exe")
	if _, err := os.Stat(srcLauncher); os.IsNotExist(err) {
		srcLauncher = filepath.Join(rootSourcePath, "Discord-Portable.exe")
	}
	targetLauncher := filepath.Join(cfg.InstallDir, "DiscordPTB-Portable.exe")
	if _, err := os.Stat(srcLauncher); err == nil {
		_ = copyFile(srcLauncher, targetLauncher)
		_ = copyFile(srcLauncher, filepath.Join(cfg.InstallDir, "Discord-Portable.exe"))
	}

	// Copy portapp.json configuration
	srcPortapp := filepath.Join(rootSourcePath, "portapp.json")
	if _, err := os.Stat(srcPortapp); err == nil {
		_ = copyFile(srcPortapp, filepath.Join(cfg.InstallDir, "portapp.json"))
	}

	// Deploy Discord App Binaries into app/app-1.0.1231
	updateStatus(45, "กำลังติดตั้ง Discord PTB...", false, "")
	dstAppDir := filepath.Join(cfg.InstallDir, "app", "app-1.0.1231")
	srcAppDir := filepath.Join(rootSourcePath, "app", "app-1.0.1231")
	if _, err := os.Stat(filepath.Join(srcAppDir, "DiscordPTB.exe")); err == nil {
		_ = copyDir(srcAppDir, dstAppDir)
	} else {
		// Fallback from LocalAppData if available
		localAppDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "DiscordPTB", "app-1.0.1231")
		if _, err := os.Stat(filepath.Join(localAppDir, "DiscordPTB.exe")); err == nil {
			_ = copyDir(localAppDir, dstAppDir)
		}
	}

	// Deploy Discord Updater database and packages into app/ (CRITICAL: prevents InconsistentInstallerState)
	updateStatus(55, "กำลังติดตั้งฐานข้อมูลระบบ...", false, "")
	dstAppBase := filepath.Join(cfg.InstallDir, "app")
	_ = os.MkdirAll(dstAppBase, 0755)

	srcInstallerDB := filepath.Join(rootSourcePath, "app", "installer.db")
	dstInstallerDB := filepath.Join(dstAppBase, "installer.db")
	if _, err := os.Stat(srcInstallerDB); err == nil {
		_ = copyFile(srcInstallerDB, dstInstallerDB)
	}

	srcPackagesDir := filepath.Join(rootSourcePath, "app", "packages")
	dstPackagesDir := filepath.Join(dstAppBase, "packages")
	if _, err := os.Stat(srcPackagesDir); err == nil {
		_ = copyDir(srcPackagesDir, dstPackagesDir)
	}

	// Setup Vencord Client Mod
	dstVencordApp := filepath.Join(dstAppDir, "resources", "app")
	srcVencord := filepath.Join(rootSourcePath, "res", "vencord")
	if _, err := os.Stat(srcVencord); err == nil {
		_ = copyDir(srcVencord, dstVencordApp)
	}

	// Initialize Profiles
	updateStatus(70, "กำลังสร้างพื้นที่โปรไฟล์...", false, "")
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []string{"01_yellow"}
	}

	// Setup roaming APPDATA settings.json for profile
	roamingDir := filepath.Join(profilesDir, "discordptb")
	_ = os.MkdirAll(roamingDir, 0700)
	_ = os.WriteFile(filepath.Join(roamingDir, "settings.json"), []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}"), 0644)

	desktopDir, _ := os.UserHomeDir()
	desktopDir = filepath.Join(desktopDir, "Desktop")

	for idx, prof := range cfg.Profiles {
		profDir := filepath.Join(profilesDir, prof)
		_ = security.ProtectPath(profDir)

		// Create settings.json
		settingsPath := filepath.Join(profDir, "settings.json")
		settingsContent := []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}")
		_ = os.WriteFile(settingsPath, settingsContent, 0644)

		// Create desktop shortcut if enabled
		if cfg.CreateShortcuts {
			displayName := profileDisplayName(prof)
			pct := 70 + int((float64(idx+1)/float64(len(cfg.Profiles)))*20)
			updateStatus(pct, fmt.Sprintf("กำลังสร้างทางลัด %s...", displayName), false, "")

			iconFile := resolveSetupIcon(prof, dstIcons)

			shortcutName := fmt.Sprintf("%s.lnk", displayName)
			shortcutPath := filepath.Join(desktopDir, shortcutName)

			args := fmt.Sprintf("--profile=%s", prof)
			if cfg.EcoMode && idx > 0 {
				args += " --eco"
			}

			_ = shortcut.Create(shortcut.Shortcut{
				ShortcutPath:     shortcutPath,
				TargetPath:       targetLauncher,
				Arguments:        shortcut.Property{Value: args},
				Description:      shortcut.Property{Value: displayName},
				IconLocation:     shortcut.Property{Value: iconFile},
				WorkingDirectory: shortcut.Property{Value: cfg.InstallDir},
			})
		}
	}

	updateStatus(95, "กำลังตรวจสอบความพร้อม...", false, "")
	time.Sleep(500 * time.Millisecond)

	updateStatus(100, "ติดตั้งเสร็จสมบูรณ์", true, "")

	// Note: Auto-launching is handled gracefully by user clicking "Finish" on Step 4
	return nil
}
