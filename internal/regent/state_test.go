package regent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)

	original := State{
		RalphPID:        12345,
		Iteration:       7,
		ConsecutiveErrs: 0,
		LastOutputAt:    now,
		LastCommit:      "abc1234",
		TotalCostUSD:    1.42,
	}

	if err := SaveState(dir, original); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if loaded.RalphPID != original.RalphPID {
		t.Errorf("RalphPID = %d, want %d", loaded.RalphPID, original.RalphPID)
	}
	if loaded.Iteration != original.Iteration {
		t.Errorf("Iteration = %d, want %d", loaded.Iteration, original.Iteration)
	}
	if loaded.ConsecutiveErrs != original.ConsecutiveErrs {
		t.Errorf("ConsecutiveErrs = %d, want %d", loaded.ConsecutiveErrs, original.ConsecutiveErrs)
	}
	if !loaded.LastOutputAt.Equal(original.LastOutputAt) {
		t.Errorf("LastOutputAt = %v, want %v", loaded.LastOutputAt, original.LastOutputAt)
	}
	if loaded.LastCommit != original.LastCommit {
		t.Errorf("LastCommit = %q, want %q", loaded.LastCommit, original.LastCommit)
	}
	if loaded.TotalCostUSD != original.TotalCostUSD {
		t.Errorf("TotalCostUSD = %f, want %f", loaded.TotalCostUSD, original.TotalCostUSD)
	}
}

func TestLoadState_NoFile(t *testing.T) {
	dir := t.TempDir()
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState with no file should not error: %v", err)
	}
	if state.RalphPID != 0 {
		t.Errorf("expected zero state, got PID=%d", state.RalphPID)
	}
}

func TestSaveState_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	ralphDir := filepath.Join(dir, stateDirName)

	// Ensure .ralph does not exist
	if _, err := os.Stat(ralphDir); !os.IsNotExist(err) {
		t.Fatal("expected .ralph to not exist initially")
	}

	if err := SaveState(dir, State{RalphPID: 1}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	if _, err := os.Stat(ralphDir); os.IsNotExist(err) {
		t.Error("expected .ralph directory to be created")
	}
}

func TestLoadState_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, stateDirName)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, stateFileName), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadState(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
