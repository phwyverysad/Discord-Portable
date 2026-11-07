package security

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/portapps/portapps/v3/pkg/log"
)

// BlockedFlags defines forbidden Electron/Chrome flags that could be exploited to steal tokens or inspect memory.
var BlockedFlags = []string{
	"--remote-debugging-port",
	"--remote-debugging-pipe",
	"--inspect",
	"--inspect-brk",
	"--enable-logging",
	"--ignore-certificate-errors",
	"--explicitly-allowed-ports",
	"--allow-insecure-localhost",
}

// SanitizeArgs strips out any dangerous or unauthorized flags.
func SanitizeArgs(args []string) (safeArgs []string, stripped []string) {
	for _, arg := range args {
		blocked := false
		lower := strings.ToLower(arg)
		for _, forbidden := range BlockedFlags {
			if strings.HasPrefix(lower, forbidden) {
				blocked = true
				stripped = append(stripped, arg)
				log.Warn().Str("flag", arg).Msg("Security Shield: Blocked malicious/debugging flag")
				break
			}
		}
		if !blocked {
			safeArgs = append(safeArgs, arg)
		}
	}
	return safeArgs, stripped
}

// ComputeFileHash calculates the SHA-256 hash of a file.
func ComputeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyIntegrity verifies that target files match their expected SHA-256 checksums if a registry exists.
func VerifyIntegrity(filePath string, expectedHash string) error {
	if expectedHash == "" {
		return nil
	}
	actual, err := ComputeFileHash(filePath)
	if err != nil {
		return fmt.Errorf("integrity check failed to read %s: %w", filePath, err)
	}
	if !strings.EqualFold(actual, expectedHash) {
		log.Error().
			Str("file", filePath).
			Str("expected", expectedHash).
			Str("actual", actual).
			Msg("Security Shield: Hash integrity mismatch detected! Possible tampering.")
		return errors.New("file integrity verification failed")
	}
	return nil
}

// ProtectPath verifies that the profile user data path is properly isolated.
func ProtectPath(profilePath string) error {
	if err := os.MkdirAll(profilePath, 0o700); err != nil {
		return fmt.Errorf("cannot create secure profile directory: %w", err)
	}

	appData := os.Getenv("APPDATA")
	if appData != "" && strings.HasPrefix(filepath.Clean(profilePath), filepath.Clean(appData)) {
		log.Warn().Str("path", profilePath).Msg("Warning: Profile path is inside %APPDATA%, recommended to use isolated portable directory")
	}

	return nil
}
