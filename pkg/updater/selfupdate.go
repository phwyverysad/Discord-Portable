package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/portapps/portapps/v3/pkg/log"
)

// ReleaseInfo represents the GitHub release response.
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckAndSelfUpdate checks GitHub for a newer version of the launcher and applies an atomic update.
func CheckAndSelfUpdate(repoOwner, repoName, currentVersion string) (bool, error) {
	if repoOwner == "" || repoName == "" {
		return false, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	client := &http.Client{Timeout: 8 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Discord-Portable-SelfUpdater")

	resp, err := client.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("SelfUpdater: Could not reach GitHub releases, skipping update check")
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Msg("SelfUpdater: GitHub returned non-200")
		return false, nil
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, err
	}

	latestTag := strings.TrimPrefix(release.TagName, "v")
	currentVer := strings.TrimPrefix(currentVersion, "v")

	if latestTag == "" || latestTag == currentVer {
		log.Info().Str("version", currentVersion).Msg("SelfUpdater: Launcher is up to date")
		return false, nil
	}

	log.Info().
		Str("current", currentVer).
		Str("latest", latestTag).
		Msg("SelfUpdater: New launcher version found, attempting atomic update")

	// Find matching executable asset
	currentExe, err := os.Executable()
	if err != nil {
		return false, err
	}
	exeName := filepath.Base(currentExe)

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, exeName) || strings.HasSuffix(strings.ToLower(asset.Name), ".exe") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		log.Warn().Msg("SelfUpdater: No matching executable asset found in latest release")
		return false, nil
	}

	// Download to .new file
	newExePath := currentExe + ".new"
	out, err := os.Create(newExePath)
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file for update: %w", err)
	}

	dlResp, err := client.Get(downloadURL)
	if err != nil {
		out.Close()
		os.Remove(newExePath)
		return false, err
	}
	defer dlResp.Body.Close()

	_, err = io.Copy(out, dlResp.Body)
	out.Close()
	if err != nil {
		os.Remove(newExePath)
		return false, err
	}

	// Windows atomic swap: rename current running exe to .old, then move .new to current
	oldExePath := currentExe + ".old"
	os.Remove(oldExePath) // remove previous old file if any

	if err := os.Rename(currentExe, oldExePath); err != nil {
		os.Remove(newExePath)
		return false, fmt.Errorf("failed renaming current executable: %w", err)
	}

	if err := os.Rename(newExePath, currentExe); err != nil {
		// Rollback
		os.Rename(oldExePath, currentExe)
		return false, fmt.Errorf("failed moving new executable to current location: %w", err)
	}

	log.Info().Msg("SelfUpdater: Atomic update succeeded! New version will run on next execution.")
	return true, nil
}
