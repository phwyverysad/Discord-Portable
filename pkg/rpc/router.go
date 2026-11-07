package rpc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/portapps/portapps/v3/pkg/log"
)

// ConfigureRouting sets up environment variables to control Discord RPC (Rich Presence) per profile.
func ConfigureRouting(profileID string) {
	// Extract numeric ID from profile string (e.g. "1", "profile_1", "01_yellow")
	numID := extractProfileNumber(profileID)

	if numID <= 1 {
		// Profile 1 is the Primary RPC master
		os.Setenv("DISCORD_INSTANCE_ID", "0")
		log.Info().Str("profile", profileID).Msg("RPC Router: Designated as Primary RPC Master (discord-ipc-0)")
	} else {
		// Secondary profiles offset their IPC pipe to prevent conflict with games / Spotify
		offset := strconv.Itoa(numID - 1)
		os.Setenv("DISCORD_INSTANCE_ID", offset)
		log.Info().
			Str("profile", profileID).
			Str("instance_id", offset).
			Msg("RPC Router: Designated as Secondary instance (avoiding pipe collision)")
	}
}

func extractProfileNumber(profileID string) int {
	clean := strings.TrimSpace(profileID)
	// Try direct conversion
	if n, err := strconv.Atoi(clean); err == nil {
		return n
	}

	// Try extracting leading digits or suffix digits
	clean = strings.TrimPrefix(clean, "profile_")
	clean = strings.TrimPrefix(clean, "profile")
	parts := strings.Split(clean, "_")
	if len(parts) > 0 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			return n
		}
	}

	for i := 0; i < len(clean); i++ {
		if clean[i] >= '0' && clean[i] <= '9' {
			end := i + 1
			for end < len(clean) && clean[end] >= '0' && clean[end] <= '9' {
				end++
			}
			if n, err := strconv.Atoi(clean[i:end]); err == nil {
				return n
			}
		}
	}

	return 1
}

// GetPipeName returns the expected named pipe name for the profile
func GetPipeName(profileID string) string {
	numID := extractProfileNumber(profileID)
	if numID <= 1 {
		return `\\.\pipe\discord-ipc-0`
	}
	return fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, numID-1)
}
