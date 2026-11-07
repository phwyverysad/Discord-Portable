package rpc

import (
	"os"
	"testing"
)

func TestConfigureRouting(t *testing.T) {
	tests := []struct {
		profileID      string
		expectedEnv    string
		expectedPipe   string
	}{
		{"1", "0", `\\.\pipe\discord-ipc-0`},
		{"profile_1", "0", `\\.\pipe\discord-ipc-0`},
		{"01_yellow", "0", `\\.\pipe\discord-ipc-0`},
		{"2", "1", `\\.\pipe\discord-ipc-1`},
		{"profile_2", "1", `\\.\pipe\discord-ipc-1`},
		{"02_yellow_orange", "1", `\\.\pipe\discord-ipc-1`},
		{"12_yellow_green", "11", `\\.\pipe\discord-ipc-11`},
	}

	for _, tt := range tests {
		ConfigureRouting(tt.profileID)
		actualEnv := os.Getenv("DISCORD_INSTANCE_ID")
		if actualEnv != tt.expectedEnv {
			t.Errorf("profile %s: expected env %s, got %s", tt.profileID, tt.expectedEnv, actualEnv)
		}
		actualPipe := GetPipeName(tt.profileID)
		if actualPipe != tt.expectedPipe {
			t.Errorf("profile %s: expected pipe %s, got %s", tt.profileID, tt.expectedPipe, actualPipe)
		}
	}
}
