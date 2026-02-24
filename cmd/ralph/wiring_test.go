package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
)

func TestNewStateTracker(t *testing.T) {
	dir := t.TempDir()
	// Create a fake git repo so CurrentBranch doesn't error
	initGitRepo(t, dir)

	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "build", runner)

	if st.state.RalphPID == 0 {
		t.Error("expected non-zero PID")
	}
	if st.state.Mode != "build" {
		t.Errorf("Mode = %q, want %q", st.state.Mode, "build")
	}
	if st.state.Branch == "" {
		t.Error("expected non-empty branch")
	}
	if st.state.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
	if st.state.LastOutputAt.IsZero() {
		t.Error("expected LastOutputAt to be set")
	}
}

func TestStateTrackerTrackEntry(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "plan", runner)

	st.trackEntry(loop.LogEntry{
		Iteration: 3,
		TotalCost: 1.50,
		Commit:    "abc1234 feat: add stuff",
		Branch:    "feat/test",
		Mode:      "build",
	})

	if st.state.Iteration != 3 {
		t.Errorf("Iteration = %d, want 3", st.state.Iteration)
	}
	if st.state.TotalCostUSD != 1.50 {
		t.Errorf("TotalCostUSD = %f, want 1.50", st.state.TotalCostUSD)
	}
	if st.state.LastCommit != "abc1234 feat: add stuff" {
		t.Errorf("LastCommit = %q, want %q", st.state.LastCommit, "abc1234 feat: add stuff")
	}
	if st.state.Branch != "feat/test" {
		t.Errorf("Branch = %q, want %q", st.state.Branch, "feat/test")
	}
	if st.state.Mode != "build" {
		t.Errorf("Mode = %q, want %q", st.state.Mode, "build")
	}
}

func TestStateTrackerZeroValuesPreserved(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "build", runner)

	st.trackEntry(loop.LogEntry{Iteration: 5, Branch: "main"})
	st.trackEntry(loop.LogEntry{Message: "just info"})

	if st.state.Iteration != 5 {
		t.Errorf("Iteration = %d, want 5 (should not be overwritten by zero)", st.state.Iteration)
	}
	if st.state.Branch != "main" {
		t.Errorf("Branch = %q, want %q", st.state.Branch, "main")
	}
}

func TestStateTrackerSave(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "build", runner)

	st.trackEntry(loop.LogEntry{Iteration: 2, TotalCost: 0.75})
	st.save()

	state, err := regent.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.Iteration != 2 {
		t.Errorf("saved Iteration = %d, want 2", state.Iteration)
	}
	if state.TotalCostUSD != 0.75 {
		t.Errorf("saved TotalCostUSD = %f, want 0.75", state.TotalCostUSD)
	}
}

func TestStateTrackerLivePersistence(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "build", runner)
	st.save()

	// trackEntry with meaningful fields auto-saves to disk
	st.trackEntry(loop.LogEntry{Iteration: 3, TotalCost: 1.25, Branch: "feat/live"})

	state, err := regent.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.Iteration != 3 {
		t.Errorf("live Iteration = %d, want 3", state.Iteration)
	}
	if state.TotalCostUSD != 1.25 {
		t.Errorf("live TotalCostUSD = %f, want 1.25", state.TotalCostUSD)
	}
	if state.Branch != "feat/live" {
		t.Errorf("live Branch = %q, want %q", state.Branch, "feat/live")
	}

	// trackEntry with no meaningful changes does not overwrite disk state
	st.trackEntry(loop.LogEntry{Message: "just a log line"})

	state2, err := regent.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState after no-op: %v", err)
	}
	if state2.Iteration != 3 {
		t.Errorf("after no-op Iteration = %d, want 3", state2.Iteration)
	}
}

func TestStateTrackerFinish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		initGitRepo(t, dir)
		runner := git.NewRunner(dir)
		st := newStateTracker(dir, "build", runner)

		st.finish(nil)

		state, err := regent.LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if !state.Passed {
			t.Error("expected Passed = true for nil error")
		}
		if state.FinishedAt.IsZero() {
			t.Error("expected FinishedAt to be set")
		}
	})

	t.Run("context canceled treated as success", func(t *testing.T) {
		dir := t.TempDir()
		initGitRepo(t, dir)
		runner := git.NewRunner(dir)
		st := newStateTracker(dir, "build", runner)

		st.finish(context.Canceled)

		state, err := regent.LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if !state.Passed {
			t.Error("expected Passed = true for context.Canceled (normal shutdown)")
		}
	})

	t.Run("real error treated as failure", func(t *testing.T) {
		dir := t.TempDir()
		initGitRepo(t, dir)
		runner := git.NewRunner(dir)
		st := newStateTracker(dir, "build", runner)

		st.finish(errors.New("something broke"))

		state, err := regent.LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if state.Passed {
			t.Error("expected Passed = false for real error")
		}
	})
}

func TestStateTrackerLastOutputAt(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	st := newStateTracker(dir, "build", runner)

	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	st.trackEntry(loop.LogEntry{Iteration: 1})
	after := time.Now()

	if st.state.LastOutputAt.Before(before) || st.state.LastOutputAt.After(after) {
		t.Errorf("LastOutputAt = %v, expected between %v and %v", st.state.LastOutputAt, before, after)
	}
}

// initGitRepo creates a minimal git repo in dir for tests that need git operations.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "checkout", "-b", "test-branch"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}
}
