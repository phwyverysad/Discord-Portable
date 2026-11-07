package vencord

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsPatched(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vencord-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	if IsPatched(tempDir) {
		t.Errorf("expected IsPatched to be false on empty dir")
	}

	// Create _app.asar
	origAsar := filepath.Join(tempDir, "_app.asar")
	if err := os.WriteFile(origAsar, []byte("fake-asar"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsPatched(tempDir) {
		t.Errorf("expected IsPatched to be true after creating _app.asar")
	}
}

func TestInstallShim(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vencord-shim-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	resourcesDir := filepath.Join(tempDir, "resources")
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Fake app.asar
	appAsar := filepath.Join(resourcesDir, "app.asar")
	if err := os.WriteFile(appAsar, []byte("original-discord-asar"), 0644); err != nil {
		t.Fatal(err)
	}

	// Fake vencord source
	vencordSource := filepath.Join(tempDir, "vencord_source")
	if err := os.MkdirAll(vencordSource, 0755); err != nil {
		t.Fatal(err)
	}

	for _, f := range append([]string{"package.json", "index.js"}, RequiredFiles...) {
		if err := os.WriteFile(filepath.Join(vencordSource, f), []byte("content of "+f), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run InstallShim
	if err := InstallShim(resourcesDir, vencordSource); err != nil {
		t.Fatalf("InstallShim failed: %v", err)
	}

	// Verify _app.asar exists and contains original content
	origAsar := filepath.Join(resourcesDir, "_app.asar")
	origContent, err := os.ReadFile(origAsar)
	if err != nil {
		t.Fatalf("_app.asar was not created: %v", err)
	}
	if string(origContent) != "original-discord-asar" {
		t.Errorf("unexpected _app.asar content: %s", string(origContent))
	}

	// Verify shim files in resources/app
	appIndex := filepath.Join(resourcesDir, "app", "index.js")
	indexContent, err := os.ReadFile(appIndex)
	if err != nil {
		t.Fatalf("app/index.js missing: %v", err)
	}
	if !strings.Contains(string(indexContent), "index.js") {
		t.Errorf("unexpected app/index.js content: %s", string(indexContent))
	}

	appPatcher := filepath.Join(resourcesDir, "app", "patcher.js")
	if _, err := os.Stat(appPatcher); err != nil {
		t.Errorf("app/patcher.js missing: %v", err)
	}

	if !IsPatched(resourcesDir) {
		t.Errorf("IsPatched should return true after InstallShim")
	}
}
