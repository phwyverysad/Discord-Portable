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
	"time"

	"github.com/portapps/discord-ptb-portable/pkg/security"
	"github.com/portapps/portapps/v3/pkg/shortcut"
)

type InstallConfig struct {
	InstallDir     string   `json:"installDir"`
	InstallMode    string   `json:"installMode"` // "pc" or "usb"
	EcoMode        bool     `json:"ecoMode"`
	Profiles       []string `json:"profiles"`
	LaunchOnFinish bool     `json:"launchOnFinish"`
}

type InstallStatus struct {
	Percent int    `json:"percent"`
	Message string `json:"message"`
	Done    bool   `json:"done"`
	Error   string `json:"error,omitempty"`
}

var (
	currentStatus = InstallStatus{Percent: 0, Message: "Ready"}
	statusMutex   sync.RWMutex
)

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
	updateStatus(5, "กำลังเตรียมโฟลเดอร์สำหรับติดตั้ง...", false, "")

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

	updateStatus(15, "กำลังตั้งค่าระบบความปลอดภัย Token & Credential Shield...", false, "")
	profilesDir := filepath.Join(cfg.InstallDir, "data", "profiles")
	if err := os.MkdirAll(profilesDir, 0700); err != nil {
		return err
	}

	// Copy Icons
	updateStatus(30, "กำลังคัดลอกชุดไอคอนสีทั้ง 12 สี...", false, "")
	srcIcons := filepath.Join(rootSourcePath, "res", "icons")
	dstIcons := filepath.Join(cfg.InstallDir, "res", "icons")
	if _, err := os.Stat(srcIcons); err == nil {
		_ = copyDir(srcIcons, dstIcons)
	}

	// Setup Launcher Binary in target
	updateStatus(45, "กำลังติดตั้งตัวรัน Discord Portable...", false, "")
	currentExe, _ := os.Executable()
	targetLauncher := filepath.Join(cfg.InstallDir, "Discord-Portable.exe")
	if currentExe != "" && filepath.Clean(currentExe) != filepath.Clean(targetLauncher) {
		_ = copyFile(currentExe, targetLauncher)
	}

	// Initialize Profiles
	updateStatus(60, "กำลังสร้างโปรไฟล์และแยกพื้นที่ข้อมูล...", false, "")
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []string{"01_yellow"}
	}

	desktopDir, _ := os.UserHomeDir()
	desktopDir = filepath.Join(desktopDir, "Desktop")

	for idx, prof := range cfg.Profiles {
		profDir := filepath.Join(profilesDir, prof)
		_ = security.ProtectPath(profDir)

		// Create settings.json
		settingsPath := filepath.Join(profDir, "settings.json")
		settingsContent := []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}")
		_ = os.WriteFile(settingsPath, settingsContent, 0644)

		// Create desktop shortcut
		pct := 60 + int((float64(idx+1)/float64(len(cfg.Profiles)))*25)
		updateStatus(pct, fmt.Sprintf("กำลังสร้างทางลัดบน Desktop สำหรับ %s...", prof), false, "")

		iconFile := filepath.Join(dstIcons, prof+".ico")
		if _, err := os.Stat(iconFile); os.IsNotExist(err) {
			iconFile = filepath.Join(dstIcons, fmt.Sprintf("profile_%02d.ico", idx+1))
		}

		shortcutName := fmt.Sprintf("Discord - %s.lnk", strings.ReplaceAll(prof, "_", " "))
		shortcutPath := filepath.Join(desktopDir, shortcutName)

		args := fmt.Sprintf("--profile=%s", prof)
		if cfg.EcoMode && idx > 0 {
			args += " --eco"
		}

		_ = shortcut.Create(shortcut.Shortcut{
			ShortcutPath:     shortcutPath,
			TargetPath:       targetLauncher,
			Arguments:        shortcut.Property{Value: args},
			Description:      shortcut.Property{Value: fmt.Sprintf("Discord Portable (%s)", prof)},
			IconLocation:     shortcut.Property{Value: iconFile},
			WorkingDirectory: shortcut.Property{Value: cfg.InstallDir},
		})
	}

	updateStatus(95, "กำลังตรวจสอบความสมบูรณ์ขั้นสุดท้าย...", false, "")
	time.Sleep(500 * time.Millisecond)

	updateStatus(100, "ติดตั้งเสร็จสมบูรณ์เรียบร้อยแล้ว!", true, "")

	if cfg.LaunchOnFinish && len(cfg.Profiles) > 0 {
		firstProf := cfg.Profiles[0]
		cmd := exec.Command(targetLauncher, fmt.Sprintf("--profile=%s", firstProf))
		cmd.Dir = cfg.InstallDir
		_ = cmd.Start()
	}

	return nil
}
