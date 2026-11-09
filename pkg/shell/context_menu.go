package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var profileShortcutNames = []string{
	"Discord Yellow.lnk",
	"Discord Yellow Orange.lnk",
	"Discord Orange.lnk",
	"Discord Red Orange.lnk",
	"Discord Red.lnk",
	"Discord Red Violet.lnk",
	"Discord Violet.lnk",
	"Discord Blue Violet.lnk",
	"Discord Blue.lnk",
	"Discord Blue Green.lnk",
	"Discord Green.lnk",
	"Discord Yellow Green.lnk",
}

// BuildLnkAppliesToFilter builds an AQS filter that matches ONLY our installed Discord shortcuts.
func BuildLnkAppliesToFilter() string {
	parts := make([]string, 0, len(profileShortcutNames)+2)
	for _, name := range profileShortcutNames {
		parts = append(parts, fmt.Sprintf("System.FileName:=\"%s\"", name))
	}
	parts = append(parts, "System.FileName:=\"Discord-Portable.lnk\"", "System.FileName:=\"DiscordPTB-Portable.lnk\"")
	return strings.Join(parts, " OR ")
}

// BuildExeAppliesToFilter builds an AQS filter that matches ONLY our portable Discord executables.
func BuildExeAppliesToFilter() string {
	return "System.FileName:=\"DiscordPTB-Portable.exe\" OR System.FileName:=\"Discord-Portable.exe\""
}

func fileExists(p string) bool {
	if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
		return true
	}
	return false
}

// RegisterDiscordContextMenu registers the 2 context menu verbs ("ติดตั้งเพิ่มเติม" and "ถอนการติดตั้ง")
// in HKCU\Software\Classes for both .lnk shortcuts and .exe files, guarded by strict AppliesTo filters
// so they ONLY show when right-clicking our installed Discord instances.
func RegisterDiscordContextMenu(managerExePath, iconDir string) error {
	absManager, err := filepath.Abs(managerExePath)
	if err != nil {
		absManager = managerExePath
	}

	targetDir := filepath.Dir(absManager)

	// App icon resolution for verbs:
	// For "ติดตั้งเพิ่มเติม": use Discord-WebSetup.exe or 01_yellow.ico or absManager
	addIcon := absManager
	setupExe := filepath.Join(targetDir, "Discord-WebSetup.exe")
	if fileExists(setupExe) {
		addIcon = setupExe
	} else if fileExists(filepath.Join(iconDir, "01_yellow.ico")) {
		addIcon = filepath.Join(iconDir, "01_yellow.ico")
	}

	// For "ถอนการติดตั้ง": use 05_red.ico (Red Discord icon) or absManager
	uninstallIcon := absManager
	if fileExists(filepath.Join(iconDir, "05_red.ico")) {
		uninstallIcon = filepath.Join(iconDir, "05_red.ico")
	}

	lnkFilter := BuildLnkAppliesToFilter()
	exeFilter := BuildExeAppliesToFilter()

	// 1. lnkfile: "ติดตั้งเพิ่มเติม"
	if err := registerVerb(
		`Software\Classes\lnkfile\shell\DiscordInstallMore`,
		"ติดตั้งเพิ่มเติม",
		lnkFilter,
		addIcon,
		fmt.Sprintf("\"%s\" --action=add --dir=\"%s\"", absManager, targetDir),
	); err != nil {
		return err
	}

	// 2. lnkfile: "ถอนการติดตั้ง"
	if err := registerVerb(
		`Software\Classes\lnkfile\shell\DiscordUninstall`,
		"ถอนการติดตั้ง",
		lnkFilter,
		uninstallIcon,
		fmt.Sprintf("\"%s\" --action=uninstall --dir=\"%s\"", absManager, targetDir),
	); err != nil {
		return err
	}

	// 3. exefile: "ติดตั้งเพิ่มเติม"
	if err := registerVerb(
		`Software\Classes\exefile\shell\DiscordInstallMore`,
		"ติดตั้งเพิ่มเติม",
		exeFilter,
		addIcon,
		fmt.Sprintf("\"%s\" --action=add --dir=\"%s\"", absManager, targetDir),
	); err != nil {
		return err
	}

	// 4. exefile: "ถอนการติดตั้ง"
	if err := registerVerb(
		`Software\Classes\exefile\shell\DiscordUninstall`,
		"ถอนการติดตั้ง",
		exeFilter,
		uninstallIcon,
		fmt.Sprintf("\"%s\" --action=uninstall --dir=\"%s\"", absManager, targetDir),
	); err != nil {
		return err
	}

	return nil
}

// UnregisterDiscordContextMenu completely removes all context menu entries from HKCU.
func UnregisterDiscordContextMenu() {
	deleteVerbTree(`Software\Classes\lnkfile\shell\DiscordInstallMore`)
	deleteVerbTree(`Software\Classes\lnkfile\shell\DiscordUninstall`)
	deleteVerbTree(`Software\Classes\exefile\shell\DiscordInstallMore`)
	deleteVerbTree(`Software\Classes\exefile\shell\DiscordUninstall`)
}

func registerVerb(keyPath, title, appliesTo, icon, command string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	_ = k.SetStringValue("", title)
	_ = k.SetStringValue("MUIVerb", title)
	_ = k.SetStringValue("AppliesTo", appliesTo)
	if icon != "" {
		_ = k.SetStringValue("Icon", icon)
	}

	cmdKey, _, err := registry.CreateKey(k, "command", registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer cmdKey.Close()

	return cmdKey.SetStringValue("", command)
}

func deleteVerbTree(keyPath string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err == nil {
		_ = registry.DeleteKey(k, "command")
		k.Close()
	}
	_ = registry.DeleteKey(registry.CURRENT_USER, keyPath)
}
