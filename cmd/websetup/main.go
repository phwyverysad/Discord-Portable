package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/portapps/portapps/v3/pkg/log"
)

func findBrowser() string {
	candidates := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Microsoft\Edge\Application\msedge.exe`),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot get executable path")
	}

	rootPath := filepath.Dir(exePath)
	if _, err := os.Stat(filepath.Join(rootPath, "res")); os.IsNotExist(err) {
		// If running under cmd/websetup, look two levels up
		rootPath = filepath.Dir(filepath.Dir(rootPath))
	}

	server, err := NewSetupServer(rootPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed initializing WebSetup server")
	}

	server.Start()
	defer server.Stop()

	targetURL := server.URL()
	log.Info().Str("url", targetURL).Msg("WebSetup Server listening")

	browserExe := findBrowser()
	tempBrowserData := filepath.Join(os.Getenv("TEMP"), "discord_websetup_cache")
	_ = os.MkdirAll(tempBrowserData, 0700)
	defer os.RemoveAll(tempBrowserData)

	if browserExe != "" {
		appArgs := []string{
			fmt.Sprintf("--app=%s", targetURL),
			"--window-size=940,700",
			fmt.Sprintf("--user-data-dir=%s", tempBrowserData),
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-extensions",
		}
		cmd := exec.Command(browserExe, appArgs...)
		log.Info().Str("browser", browserExe).Msg("Spawning WebSetup window in App mode")

		if err := cmd.Run(); err != nil {
			log.Warn().Err(err).Msg("Browser closed or error")
		}
	} else {
		// Fallback to default browser
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
		_ = cmd.Start()
		// Keep server alive
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
