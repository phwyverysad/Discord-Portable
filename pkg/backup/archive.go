package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/portapps/portapps/v3/pkg/log"
)

// CreateBackup zips the profiles directory to a timestamped backup archive.
func CreateBackup(profilesDir string, outputZipPath string) (string, error) {
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return "", fmt.Errorf("profiles directory does not exist: %s", profilesDir)
	}

	if outputZipPath == "" {
		timestamp := time.Now().Format("20060102_150405")
		outputZipPath = filepath.Join(filepath.Dir(profilesDir), fmt.Sprintf("discord_profiles_backup_%s.zip", timestamp))
	}

	outFile, err := os.Create(outputZipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	err = filepath.Walk(profilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(profilesDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		zipEntryPath := filepath.ToSlash(relPath)

		if info.IsDir() {
			_, err = w.Create(zipEntryPath + "/")
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipEntryPath
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return "", fmt.Errorf("failed walking profiles folder: %w", err)
	}

	log.Info().Str("destination", outputZipPath).Msg("Backup completed successfully")
	return outputZipPath, nil
}

// RestoreBackup restores a zipped profile archive into target profiles directory.
func RestoreBackup(zipPath string, targetProfilesDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed opening zip file: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(targetProfilesDir, 0o755); err != nil {
		return fmt.Errorf("failed creating target dir: %w", err)
	}

	for _, f := range r.File {
		// Prevent ZipSlip vulnerability
		cleanedPath := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanedPath, "..") || filepath.IsAbs(cleanedPath) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		destPath := filepath.Join(targetProfilesDir, filepath.FromSlash(f.Name))

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	log.Info().Str("source", zipPath).Str("target", targetProfilesDir).Msg("Restore completed successfully")
	return nil
}
