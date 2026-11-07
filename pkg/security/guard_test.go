package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeArgs(t *testing.T) {
	input := []string{
		"--profile=1",
		"--remote-debugging-port=9222",
		"--inspect",
		"--inspect-brk=9229",
		"--enable-logging",
		"--multi-instance",
		"--ignore-certificate-errors",
		"--normal-flag=test",
	}

	safe, stripped := SanitizeArgs(input)

	if len(stripped) != 5 {
		t.Fatalf("expected 5 stripped flags, got %d: %v", len(stripped), stripped)
	}

	if len(safe) != 3 {
		t.Fatalf("expected 3 safe flags, got %d: %v", len(safe), safe)
	}

	expectedSafe := []string{"--profile=1", "--multi-instance", "--normal-flag=test"}
	for i, arg := range safe {
		if arg != expectedSafe[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedSafe[i], arg)
		}
	}
}

func TestComputeFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("Discord Portable Security Shield Test")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := ComputeFileHash(filePath)
	if err != nil {
		t.Fatalf("unexpected error computing hash: %v", err)
	}

	if hash == "" {
		t.Fatalf("expected non-empty hash")
	}

	// Verify integrity with matching hash
	if err := VerifyIntegrity(filePath, hash); err != nil {
		t.Errorf("expected integrity match, got error: %v", err)
	}

	// Verify integrity with wrong hash
	if err := VerifyIntegrity(filePath, "invalid_hash_12345"); err == nil {
		t.Errorf("expected integrity failure for wrong hash, but got nil")
	}
}
