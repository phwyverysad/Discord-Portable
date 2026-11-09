package main

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestEmbeddedWebFS(t *testing.T) {
	sub, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		t.Fatalf("failed fs.Sub: %v", err)
	}

	filesToCheck := []string{
		"index.html",
		"style.css",
		"app.js",
		"i18n.js",
		"assets/bg.jpeg",
		"assets/icons/01_yellow_FFFF00.png",
	}

	for _, f := range filesToCheck {
		if _, err := fs.Stat(sub, f); err != nil {
			t.Errorf("expected embedded file %s, but stat failed: %v", f, err)
		}
	}
}

func TestExecuteFullInstallation(t *testing.T) {
	rootPath, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cfg := InstallConfig{
		InstallDir:      `C:\Discord-PTB-Portable`,
		InstallMode:     "pc",
		EcoMode:         false,
		Profiles: []string{
			"01_yellow", "02_yellow_orange", "03_orange", "04_red_orange",
			"05_red", "06_red_violet", "07_violet", "08_blue_violet",
			"09_blue", "10_blue_green", "11_green", "12_yellow_green",
		},
		InstallVencord:  true,
		LaunchOnFinish:  false,
		CreateShortcuts: true,
	}
	err = ExecuteInstallation(cfg, rootPath)
	if err != nil {
		t.Fatalf("ExecuteInstallation failed: %v", err)
	}
}
