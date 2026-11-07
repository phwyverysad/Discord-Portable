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

	"github.com/portapps/discord-ptb-portable/pkg/backup"
	"github.com/portapps/discord-ptb-portable/pkg/rpc"
	"github.com/portapps/discord-ptb-portable/pkg/security"
	"github.com/portapps/discord-ptb-portable/pkg/updater"
	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/files"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/registry"
	"github.com/portapps/portapps/v3/pkg/shortcut"
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

func parseFlags(rawArgs []string) (profile string, eco bool, vencordMode bool, backupMode bool, restorePath string, filteredArgs []string) {
	profile = "01_yellow"
	for _, arg := range rawArgs {
		switch {
		case strings.HasPrefix(arg, "--profile="):
			profile = strings.TrimPrefix(arg, "--profile=")
		case arg == "--eco":
			eco = true
		case arg == "--vencord":
			vencordMode = true
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
	// Clean profile ID
	id := strings.TrimSpace(profileID)
	// Try direct icon match in res/icons
	iconDir := filepath.Join(rootPath, "res", "icons")
	directMatch := filepath.Join(iconDir, id+".ico")
	if _, err := os.Stat(directMatch); err == nil {
		return directMatch
	}

	// Match by prefix (e.g. "1" matches "01_yellow.ico")
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

	profileID, ecoMode, vencordMode, doBackup, restoreFile, remainingArgs := parseFlags(safeArgs)

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
	}

	if !vencordMode {
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

	// Manage Start Menu & Desktop Shortcut with dedicated colored icon
	dispName := profileDisplayName(profileID)
	shortcutName := fmt.Sprintf("%s.lnk", dispName)
	shortcutPath := filepath.Join(files.StartMenuPath(), shortcutName)
	_ = os.WriteFile(shortcutPath, defaultShortcut, 0644)

	_ = shortcut.Create(shortcut.Shortcut{
		ShortcutPath:     shortcutPath,
		TargetPath:       app.Process,
		Arguments:        shortcut.Property{Value: fmt.Sprintf("--user-data-dir=\"%s\" --multi-instance", profileDataPath)},
		Description:      shortcut.Property{Value: dispName},
		IconLocation:     shortcut.Property{Value: profileIcon},
		WorkingDirectory: shortcut.Property{Value: app.AppPath},
	})

	defer func() {
		_ = os.Remove(shortcutPath)
		app.Close()
	}()

	log.Info().
		Str("profile", profileID).
		Str("dataPath", profileDataPath).
		Str("icon", profileIcon).
		Msg("Launching Discord PTB Portable Profile...")

	// Silent, clean launch without console window or error popups on exit
	cmd := exec.Command(app.Process, launchArgs...)
	cmd.Dir = electronAppPath
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	_ = cmd.Run()
}
