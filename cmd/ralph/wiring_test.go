package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
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

// --- Tests for runWithStateTracking ---

func TestRunWithStateTracking_Success(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	err := runWithStateTracking(context.Background(), lp, dir, runner, "build", nil, func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, err := regent.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if !state.Passed {
		t.Error("expected Passed = true")
	}
	if state.FinishedAt.IsZero() {
		t.Error("expected FinishedAt to be set")
	}
}

func TestRunWithStateTracking_Error(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	want := errors.New("build failed")
	got := runWithStateTracking(context.Background(), lp, dir, runner, "build", nil, func(_ context.Context) error {
		return want
	})
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	state, err := regent.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.Passed {
		t.Error("expected Passed = false for real error")
	}
}

func TestRunWithStateTracking_ContextCanceled(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	err := runWithStateTracking(context.Background(), lp, dir, runner, "build", nil, func(_ context.Context) error {
		return context.Canceled
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	state, loadErr := regent.LoadState(dir)
	if loadErr != nil {
		t.Fatalf("LoadState: %v", loadErr)
	}
	if !state.Passed {
		t.Error("expected Passed = true for context.Canceled (graceful shutdown)")
	}
}

func TestRunWithStateTracking_EventsForwarded(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	// The run func sends events through lp.Events, which is set by runWithStateTracking
	// before calling run. The drain goroutine processes all events before returning.
	err := runWithStateTracking(context.Background(), lp, dir, runner, "plan", nil, func(_ context.Context) error {
		lp.Events <- loop.LogEntry{Iteration: 3, TotalCost: 1.50, Branch: "feat/test", Mode: "build"}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, loadErr := regent.LoadState(dir)
	if loadErr != nil {
		t.Fatalf("LoadState: %v", loadErr)
	}
	if state.Iteration != 3 {
		t.Errorf("Iteration = %d, want 3", state.Iteration)
	}
	if state.TotalCostUSD != 1.50 {
		t.Errorf("TotalCostUSD = %f, want 1.50", state.TotalCostUSD)
	}
	if state.Branch != "feat/test" {
		t.Errorf("Branch = %q, want %q", state.Branch, "feat/test")
	}
	if state.Mode != "build" {
		t.Errorf("Mode = %q, want %q", state.Mode, "build")
	}
}

// --- Tests for runWithRegent ---

// regentTestConfig returns a Config with Regent settings tuned for fast tests:
// hang detection disabled, no retry backoff, fail fast after 0 retries.
func regentTestConfig(maxRetries int) *config.Config {
	cfg := config.Defaults()
	cfg.Regent.HangTimeoutSeconds = 0  // disable hang detection ticker
	cfg.Regent.RetryBackoffSeconds = 0 // no wait between retries
	cfg.Regent.MaxRetries = maxRetries
	cfg.Regent.RollbackOnTestFailure = false // no git ops in tests
	return &cfg
}

func TestRunWithRegent_Success(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	err := runWithRegent(context.Background(), lp, regentTestConfig(1), runner, dir, nil, func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, loadErr := regent.LoadState(dir)
	if loadErr != nil {
		t.Fatalf("LoadState: %v", loadErr)
	}
	if !state.Passed {
		t.Error("expected Passed = true after successful run")
	}
}

func TestRunWithRegent_MaxRetriesExceeded(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	runErr := errors.New("loop crashed")
	err := runWithRegent(context.Background(), lp, regentTestConfig(0), runner, dir, nil, func(_ context.Context) error {
		return runErr
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "max retries") {
		t.Errorf("expected 'max retries' in error message, got: %v", err)
	}
}

func TestRunWithRegent_ContextCanceled(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Regent exits on first check

	err := runWithRegent(ctx, lp, regentTestConfig(1), runner, dir, nil, func(_ context.Context) error {
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRunWithRegent_LoopEventsUpdateState(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	runner := git.NewRunner(dir)
	lp := &loop.Loop{}

	// Non-Regent events emitted by the run func flow through the drain goroutine's
	// rgt.UpdateState path. Verify they are captured in the persisted state.
	err := runWithRegent(context.Background(), lp, regentTestConfig(1), runner, dir, nil, func(_ context.Context) error {
		lp.Events <- loop.LogEntry{Iteration: 5, TotalCost: 2.50, Branch: "main"}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, loadErr := regent.LoadState(dir)
	if loadErr != nil {
		t.Fatalf("LoadState: %v", loadErr)
	}
	if state.Iteration != 5 {
		t.Errorf("Iteration = %d, want 5", state.Iteration)
	}
	if state.TotalCostUSD != 2.50 {
		t.Errorf("TotalCostUSD = %f, want 2.50", state.TotalCostUSD)
	}
}

// --- Tests for loopController ---

func TestLoopController_IsRunning_InitiallyFalse(t *testing.T) {
	ctrl := &loopController{
		outerCtx: context.Background(),
	}
	if ctrl.IsRunning() {
		t.Error("new loopController should not be running")
	}
}

func TestLoopController_StopLoop_NoopWhenIdle(t *testing.T) {
	ctrl := &loopController{
		outerCtx: context.Background(),
	}
	// Should not panic
	ctrl.StopLoop()
	if ctrl.IsRunning() {
		t.Error("StopLoop on idle controller should not set running=true")
	}
}

func TestLoopController_StartLoop_NoopWhenRunning(t *testing.T) {
	// Set cancel directly to simulate a running loop.
	cancelCalled := false
	ctrl := &loopController{
		outerCtx: context.Background(),
		cancel:   func() { cancelCalled = true },
	}
	ctrl.StartLoop("build") // should be a no-op since cancel != nil
	// cancelCalled should still be false — we didn't call cancel, just skipped StartLoop
	if cancelCalled {
		t.Error("StartLoop should not call cancel when already running")
	}
}

func TestLoopController_StopLoop_CancelsWhenRunning(t *testing.T) {
	cancelCalled := false
	ctrl := &loopController{
		outerCtx: context.Background(),
		cancel:   func() { cancelCalled = true },
	}
	ctrl.StopLoop()
	if !cancelCalled {
		t.Error("StopLoop should call cancel when a loop is running")
	}
}

func TestLoopController_StartLoop_WhenIdle(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// Config is valid but BUILD.md is absent — runLoop fails fast on os.ReadFile.
	writeExecTestFile(t, dir, "ralph.toml", testConfigNoRegent())

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	ctrl := &loopController{
		cfg:       cfg,
		dir:       dir,
		gitRunner: git.NewRunner(dir),
		outerCtx:  context.Background(),
	}

	if ctrl.IsRunning() {
		t.Fatal("should not be running initially")
	}

	ctrl.StartLoop("build")

	// runLoop fails fast (BUILD.md missing) — wait for the goroutine to clean up.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && ctrl.IsRunning() {
		time.Sleep(5 * time.Millisecond)
	}
	if ctrl.IsRunning() {
		t.Error("runLoop goroutine should finish quickly when prompt file is missing")
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
