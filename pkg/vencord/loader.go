package vencord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/portapps/portapps/v3/pkg/log"
)

const (
	VencordRepo   = "Vendicated/Vencord"
	DefaultApiURL = "https://api.github.com/repos/" + VencordRepo + "/releases"
)

var RequiredFiles = []string{
	"patcher.js",
	"preload.js",
	"renderer.js",
	"renderer.css",
}

// Release represents a Vencord release on GitHub.
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// FetchLatestRelease queries GitHub API for the latest Vencord release info.
func FetchLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// Try /releases/latest first
	latestURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", VencordRepo)
	req, err := http.NewRequest("GET", latestURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "DiscordPTB-Portable-Vencord")
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var rel Release
			if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && len(rel.Assets) > 0 {
				return &rel, nil
			}
		}
	}

	// Fallback to /releases
	req, err = http.NewRequest("GET", DefaultApiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DiscordPTB-Portable-Vencord")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Vencord releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases JSON: %w", err)
	}

	if len(releases) == 0 {
		return nil, errors.New("no releases found for Vencord")
	}

	for _, r := range releases {
		if len(r.Assets) > 0 {
			return &r, nil
		}
	}

	return &releases[0], nil
}

// DownloadFile downloads a file from URL to local target path with redirect following.
func DownloadFile(url, targetPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "DiscordPTB-Portable-Vencord")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d for %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadVencord downloads the necessary Vencord assets to destDir and prepares index.js & package.json.
func DownloadVencord(destDir string) (*Release, error) {
	rel, err := FetchLatestRelease()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}

	assetMap := make(map[string]string)
	for _, a := range rel.Assets {
		assetMap[a.Name] = a.BrowserDownloadURL
	}

	for _, file := range RequiredFiles {
		url, ok := assetMap[file]
		if !ok {
			url = fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", VencordRepo, rel.TagName, file)
		}
		target := filepath.Join(destDir, file)
		log.Info().Str("file", file).Str("url", url).Msg("Downloading Vencord component")
		if err := DownloadFile(url, target); err != nil {
			return nil, fmt.Errorf("failed downloading %s: %w", file, err)
		}
	}

	// Create package.json
	pkgJSON := `{"name":"discord","main":"index.js"}` + "\n"
	if err := os.WriteFile(filepath.Join(destDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		return nil, err
	}

	// Create index.js shim loader
	shimJS := "// Vencord Shim Loader for Discord PTB Portable\nrequire(\"./patcher.js\");\n"
	if err := os.WriteFile(filepath.Join(destDir, "index.js"), []byte(shimJS), 0644); err != nil {
		return nil, err
	}

	return rel, nil
}

// IsPatched checks if the Discord installation in resourcesDir has already been patched with Vencord.
func IsPatched(resourcesDir string) bool {
	origAsar := filepath.Join(resourcesDir, "_app.asar")
	_, err := os.Stat(origAsar)
	return err == nil
}

// InstallShim patches Discord Electron in resourcesDir:
// 1. Renames resources/app.asar -> resources/_app.asar
// 2. Installs shim and Vencord files to resources/app (or resources/app.asar)
func InstallShim(resourcesDir string, vencordSourceDir string) error {
	appAsar := filepath.Join(resourcesDir, "app.asar")
	origAsar := filepath.Join(resourcesDir, "_app.asar")

	// Step 1: Rename app.asar to _app.asar if not already done
	if _, err := os.Stat(origAsar); os.IsNotExist(err) {
		if _, err := os.Stat(appAsar); err == nil {
			log.Info().Msg("Renaming original app.asar to _app.asar")
			if err := os.Rename(appAsar, origAsar); err != nil {
				return fmt.Errorf("failed to rename %s to _app.asar: %w", appAsar, err)
			}
		}
	}

	// Step 2: Copy shim and Vencord files into resources/app directory
	appDir := filepath.Join(resourcesDir, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", appDir, err)
	}

	filesToCopy := append([]string{"package.json", "index.js"}, RequiredFiles...)
	for _, f := range filesToCopy {
		src := filepath.Join(vencordSourceDir, f)
		dst := filepath.Join(appDir, f)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("missing Vencord file %s: %w", src, err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dst, err)
		}
	}

	log.Info().Str("dir", appDir).Msg("Successfully installed Vencord Shim Loader")
	return nil
}
