package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	prof1 := filepath.Join(profilesDir, "profile_1")
	if err := os.MkdirAll(prof1, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(prof1, "settings.json")
	if err := os.WriteFile(testFile, []byte(`{"theme":"dark"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	zipOut := filepath.Join(tmpDir, "backup.zip")
	_, err := CreateBackup(profilesDir, zipOut)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Restore to new dir
	restoreDir := filepath.Join(tmpDir, "restored_profiles")
	if err := RestoreBackup(zipOut, restoreDir); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	restoredFile := filepath.Join(restoreDir, "profile_1", "settings.json")
	content, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("failed reading restored file: %v", err)
	}

	if string(content) != `{"theme":"dark"}` {
		t.Fatalf("unexpected content: %s", string(content))
	}
}
