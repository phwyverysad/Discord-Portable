package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerAPIs(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "data", "profiles")
	_ = os.MkdirAll(filepath.Join(profilesDir, "01_yellow"), 0755)

	server := &ManagerServer{
		rootDir: tempDir,
		action:  "add",
	}

	// 1. Test GET /api/state
	reqState := httptest.NewRequest("GET", "/api/state", nil)
	wState := httptest.NewRecorder()
	server.handleState(wState, reqState)

	if wState.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", wState.Code)
	}

	var stateResp map[string]interface{}
	if err := json.Unmarshal(wState.Body.Bytes(), &stateResp); err != nil {
		t.Fatalf("Failed unmarshaling state: %v", err)
	}
	if stateResp["action"] != "add" {
		t.Fatalf("Expected action 'add', got '%v'", stateResp["action"])
	}

	installed := stateResp["installedProfiles"].([]interface{})
	if len(installed) != 1 || installed[0] != "01_yellow" {
		t.Fatalf("Expected installed '01_yellow', got %v", installed)
	}

	// 2. Test POST /api/execute (add profile)
	addPayload := map[string]interface{}{
		"action":   "add",
		"profiles": []string{"05_red"},
		"all":      false,
	}
	bodyAdd, _ := json.Marshal(addPayload)
	reqAdd := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(bodyAdd))
	wAdd := httptest.NewRecorder()
	server.handleExecute(wAdd, reqAdd)

	if wAdd.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on add, got %d", wAdd.Code)
	}

	redDir := filepath.Join(profilesDir, "05_red")
	if _, err := os.Stat(redDir); os.IsNotExist(err) {
		t.Fatalf("Expected 05_red directory to be created at %s", redDir)
	}

	// 3. Test POST /api/execute (uninstall profile)
	server.action = "uninstall"
	delPayload := map[string]interface{}{
		"action":   "uninstall",
		"profiles": []string{"05_red"},
		"all":      false,
	}
	bodyDel, _ := json.Marshal(delPayload)
	reqDel := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(bodyDel))
	wDel := httptest.NewRecorder()
	server.handleExecute(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on uninstall, got %d", wDel.Code)
	}

	if _, err := os.Stat(redDir); !os.IsNotExist(err) {
		t.Fatalf("Expected 05_red directory to be removed")
	}
}
