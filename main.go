package main

//go:generate go tool goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/portapps/discord-ptb-portable/pkg/backup"
	"github.com/portapps/discord-ptb-portable/pkg/rpc"
	"github.com/portapps/discord-ptb-portable/pkg/security"
	"github.com/portapps/discord-ptb-portable/pkg/shell"
	"github.com/portapps/discord-ptb-portable/pkg/updater"
	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/files"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/registry"
)

//go:embed res/DiscordPTB.lnk
var defaultShortcut []byte

//go:embed res/pinned_update.json
var pinnedUpdate []byte

type config struct {
	Cleanup bool `yaml:"cleanup" mapstructure:"cleanup"`
}

var (
	app *portapps.App
	cfg *config
	kernel32DLL    = syscall.NewLazyDLL("kernel32.dll")
	procFreeConsole = kernel32DLL.NewProc("FreeConsole")
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup: false,
	}

	// Init app
	if app, err = portapps.NewWithCfg("discord-ptb-portable", "DiscordPTB", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func parseFlags(rawArgs []string) (profile string, eco bool, vanillaMode bool, backupMode bool, restorePath string, filteredArgs []string) {
	profile = "01_yellow"
	for _, arg := range rawArgs {
		switch {
		case strings.HasPrefix(arg, "--profile="):
			profile = strings.TrimPrefix(arg, "--profile=")
		case arg == "--eco":
			eco = true
		case arg == "--vanilla":
			vanillaMode = true
		case arg == "--vencord":
			vanillaMode = false
		case arg == "--backup":
			backupMode = true
		case strings.HasPrefix(arg, "--restore="):
			restorePath = strings.TrimPrefix(arg, "--restore=")
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}
	return
}

func resolveIcon(profileID string, rootPath string) string {
	iconDir := filepath.Join(rootPath, "res", "icons")
	canon := shell.CanonicalProfile(profileID)
	target := filepath.Join(iconDir, canon+".ico")
	if _, err := os.Stat(target); err == nil {
		return target
	}

	directMatch := filepath.Join(iconDir, strings.TrimSpace(profileID)+".ico")
	if _, err := os.Stat(directMatch); err == nil {
		return directMatch
	}

	return filepath.Join(rootPath, "res", "papp.ico")
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

func main() {
	procFreeConsole.Call()
	rawArgs := os.Args[1:]

	// 1. Security Shield: Sanitize input arguments against debug/memory grabbers
	safeArgs, stripped := security.SanitizeArgs(rawArgs)
	if len(stripped) > 0 {
		log.Warn().Strs("stripped_flags", stripped).Msg("Security Shield: Intercepted and blocked suspicious arguments")
	}

	profileID, ecoMode, vanillaMode, doBackup, restoreFile, remainingArgs := parseFlags(safeArgs)

	profilesBaseDir := filepath.Join(app.RootPath, "data", "profiles")

	// 2. Pro Feature: One-Click Backup & Restore
	if doBackup {
		outZip, err := backup.CreateBackup(profilesBaseDir, "")
		if err != nil {
			log.Fatal().Err(err).Msg("Backup failed")
		}
		log.Info().Str("file", outZip).Msg("Profiles backup completed successfully!")
		return
	}

	if restoreFile != "" {
		if err := backup.RestoreBackup(restoreFile, profilesBaseDir); err != nil {
			log.Fatal().Err(err).Msg("Restore failed")
		}
		log.Info().Msg("Profiles restored successfully!")
		return
	}

	// 3. Pro Feature: Self-Updating Launcher Check (Non-blocking)
	go func() {
		updated, err := updater.CheckAndSelfUpdate("phwyverysad", "Discord-Portable", "1.0.1231")
		if err != nil {
			log.Debug().Err(err).Msg("Self-update check finished with notice")
		} else if updated {
			log.Info().Msg("Self-updater: Downloaded new launcher build.")
		}
	}()

	// 4. Multi-Profile Data Isolation Engine
	profileDataPath := filepath.Join(profilesBaseDir, profileID)
	if err := security.ProtectPath(profileDataPath); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize secure profile directory")
	}

	// Capture original host user roaming APPDATA for Windows Shell Start Menu registration
	origUserAppData := os.Getenv("APPDATA")

	// Portable APPDATA isolation: redirect Electron's roaming appdata to portable profiles base
	_ = os.Setenv("APPDATA", profilesBaseDir)
	roamingDiscordPath := filepath.Join(profilesBaseDir, "discordptb")
	_ = os.MkdirAll(roamingDiscordPath, 0700)
	roamingSettingsPath := filepath.Join(roamingDiscordPath, "settings.json")
	_ = os.WriteFile(roamingSettingsPath, []byte("{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}"), 0644)

	// 5. Pro Feature: Smart RPC Router
	rpc.ConfigureRouting(profileID)

	electronAppPath := app.ElectronAppPath()
	app.Process = filepath.Join(electronAppPath, "DiscordPTB.exe")

	launchArgs := []string{
		"--user-data-dir=" + profileDataPath,
		"--multi-instance",
		"--profile=" + profileID,
	}

	if vanillaMode {
		launchArgs = append(launchArgs, "--vanilla")
	}

	// 6. Pro Feature: Eco / Low-RAM Mode
	if ecoMode {
		launchArgs = append(launchArgs,
			"--disable-renderer-backgrounding",
			"--disable-background-timer-throttling=false",
			"--disable-gpu-compositing",
		)
		log.Info().Msg("Eco Mode enabled: Resource throttling active for background tabs")
	}

	launchArgs = append(launchArgs, remainingArgs...)
	app.Args = launchArgs
	app.WorkingDir = electronAppPath

	// Cleanup on exit if configured
	if cfg.Cleanup {
		defer func() {
			regKey := registry.Key{
				Key:  `HKCU\SOFTWARE\DiscordPTB`,
				Arch: "32",
			}
			if regKey.Exists() {
				if err := regKey.Delete(true); err != nil {
					log.Error().Err(err).Msg("Cannot remove registry key")
				}
			}
			files.Cleanup(
				filepath.Join(os.Getenv("APPDATA"), "discordptb"),
				filepath.Join(os.Getenv("TEMP"), "Discord Crashes"),
			)
		}()
	}

	// Setup profile-specific settings.json
	settingsPath := filepath.Join(profileDataPath, "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		rawSettings, err := os.ReadFile(settingsPath)
		if err == nil {
			jsonMapSettings := make(map[string]interface{})
			if err = json.Unmarshal(rawSettings, &jsonMapSettings); err == nil {
				jsonMapSettings["SKIP_HOST_UPDATE"] = true
				jsonMapSettings["USE_PINNED_UPDATE_MANIFEST"] = true
				if jsonSettings, err := json.Marshal(jsonMapSettings); err == nil {
					_ = os.WriteFile(settingsPath, jsonSettings, 0644)
				}
			}
		}
	} else {
		fo, err := os.Create(settingsPath)
		if err == nil {
			defer fo.Close()
			_, _ = io.WriteString(fo, "{\n  \"SKIP_HOST_UPDATE\": true,\n  \"USE_PINNED_UPDATE_MANIFEST\": true\n}")
		}
	}

	// Copy pinned_update.json into profile path
	_ = os.WriteFile(filepath.Join(profileDataPath, "pinned_update.json"), pinnedUpdate, 0644)

	// Resolve custom colored icon for this profile
	profileIcon := resolveIcon(profileID, app.RootPath)
	_ = os.Setenv("DISCORD_PORTABLE_PROFILE", profileID)
	_ = os.Setenv("DISCORD_PORTABLE_ICON", profileIcon)

	// Manage Start Menu Shortcut with dedicated colored icon and explicit AUMID
	dispName := profileDisplayName(profileID)
	shortcutName := fmt.Sprintf("%s.lnk", dispName)
	realStartMenu := filepath.Join(origUserAppData, "Microsoft", "Windows", "Start Menu", "Programs", "Discord Portable")
	_ = os.MkdirAll(realStartMenu, 0755)
	shortcutPath := filepath.Join(realStartMenu, shortcutName)
	aumid := fmt.Sprintf("Discord.PTB.Profile.%s", profileID)

	programsRoot := filepath.Join(origUserAppData, "Microsoft", "Windows", "Start Menu", "Programs")
	if _, err := os.Stat(shortcutPath); os.IsNotExist(err) {
		_ = shell.CreateProfileShortcut(
			shortcutPath,
			app.Process,
			fmt.Sprintf("--user-data-dir=\"%s\" --multi-instance", profileDataPath),
			profileIcon,
			aumid,
			dispName,
			app.AppPath,
		)
	}
	rootLnk := filepath.Join(programsRoot, shortcutName)
	if _, err := os.Stat(rootLnk); os.IsNotExist(err) {
		_ = shell.CreateProfileShortcut(
			rootLnk,
			app.Process,
			fmt.Sprintf("--user-data-dir=\"%s\" --multi-instance", profileDataPath),
			profileIcon,
			aumid,
			dispName,
			app.AppPath,
		)
	}

	defer func() {
		app.Close()
	}()

	log.Info().
		Str("profile", profileID).
		Str("dataPath", profileDataPath).
		Str("icon", profileIcon).
		Msg("Launching Discord PTB Portable Profile...")

	// Clean launch Discord PTB GUI
	cmd := exec.Command(app.Process, launchArgs...)
	cmd.Dir = electronAppPath
	if err := cmd.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start Discord PTB")
		return
	}

	// Dynamic Taskbar Icon Sync: Watch for Discord window and apply profile icon via Shell PropertyStore & WM_SETICON
	go func(targetPid int, iconFile string, targetAumid string, relaunchCmd string, displayName string) {
		if iconFile == "" {
			return
		}
		user32 := syscall.NewLazyDLL("user32.dll")
		procLoadImageW := user32.NewProc("LoadImageW")
		procSendMessageW := user32.NewProc("SendMessageW")
		procEnumWindows := user32.NewProc("EnumWindows")
		procGetWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")
		procSetClassLongPtrW := user32.NewProc("SetClassLongPtrW")
		if procSetClassLongPtrW.Find() != nil {
			procSetClassLongPtrW = user32.NewProc("SetClassLongW")
		}

		iconPtr, err := syscall.UTF16PtrFromString(iconFile)
		if err != nil {
			return
		}

		hBig, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(iconPtr)), 1, 32, 32, 0x0010)
		hSmall, _, _ := procLoadImageW.Call(0, uintptr(unsafe.Pointer(iconPtr)), 1, 16, 16, 0x0010)

		for i := 0; i < 50; i++ {
			time.Sleep(400 * time.Millisecond)
			cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
				var pid uint32
				procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
				if int(pid) == targetPid {
					_ = shell.SetWindowProfileProperties(hwnd, targetAumid, iconFile, relaunchCmd, displayName)
					if hSmall != 0 {
						procSendMessageW.Call(hwnd, 0x0080, 0, hSmall)
						procSetClassLongPtrW.Call(hwnd, ^uintptr(33), hSmall) // -34: GCLP_HICONSM
					}
					if hBig != 0 {
						procSendMessageW.Call(hwnd, 0x0080, 1, hBig)
						procSetClassLongPtrW.Call(hwnd, ^uintptr(13), hBig) // -14: GCLP_HICON
					}
				}
				return 1
			})
			procEnumWindows.Call(cb, 0)
		}
	}(cmd.Process.Pid, profileIcon, aumid, fmt.Sprintf("\"%s\" --profile=%s", filepath.Join(app.RootPath, "DiscordPTB-Portable.exe"), profileID), dispName)

	_ = cmd.Wait()
}
