package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestContextMenuRegistration(t *testing.T) {
	tempDir := t.TempDir()
	fakeManager := filepath.Join(tempDir, "Discord-ProfileManager.exe")
	_ = os.WriteFile(fakeManager, []byte("fake"), 0755)

	iconDir := filepath.Join(tempDir, "res", "icons")
	_ = os.MkdirAll(iconDir, 0755)
	_ = os.WriteFile(filepath.Join(iconDir, "01_yellow.ico"), []byte("icon"), 0644)

	err := RegisterDiscordContextMenu(fakeManager, iconDir)
	if err != nil {
		t.Fatalf("RegisterDiscordContextMenu failed: %v", err)
	}

	// Verify lnkfile verb exists
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\lnkfile\shell\DiscordInstallMore`, registry.READ)
	if err != nil {
		t.Fatalf("Expected registry key DiscordInstallMore to exist: %v", err)
	}
	defer k.Close()

	verb, _, err := k.GetStringValue("MUIVerb")
	if err != nil || verb != "ติดตั้งเพิ่มเติม" {
		t.Fatalf("Expected MUIVerb 'ติดตั้งเพิ่มเติม', got '%s'", verb)
	}

	appliesTo, _, err := k.GetStringValue("AppliesTo")
	if err != nil || !strings.Contains(appliesTo, "Discord Yellow.lnk") {
		t.Fatalf("Expected AppliesTo to contain 'Discord Yellow.lnk', got '%s'", appliesTo)
	}

	// Unregister
	UnregisterDiscordContextMenu()

	// Verify deleted
	_, err = registry.OpenKey(registry.CURRENT_USER, `Software\Classes\lnkfile\shell\DiscordInstallMore`, registry.READ)
	if err == nil {
		t.Fatalf("Expected registry key DiscordInstallMore to be deleted")
	}

	// Restore real context menu if production app is installed
	prodManager := `C:\Discord-PTB-Portable\Discord-ProfileManager.exe`
	if _, err := os.Stat(prodManager); err == nil {
		_ = RegisterDiscordContextMenu(prodManager, `C:\Discord-PTB-Portable\res\icons`)
	}
}
