package regent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

func defaultTestRegentConfig() config.RegentConfig {
	return config.RegentConfig{
		Enabled:               true,
		RollbackOnTestFailure: false,
		TestCommand:           "",
		MaxRetries:            3,
		RetryBackoffSeconds:   0, // no delay in tests
		HangTimeoutSeconds:    0, // disabled in most tests
	}
}

// collectEvents drains events from a channel into a slice.
func collectEvents(ch <-chan loop.LogEntry, done chan<- []loop.LogEntry) {
	var entries []loop.LogEntry
	for e := range ch {
		entries = append(entries, e)
	}
	done <- entries
}

func TestSupervise(t *testing.T) {
	t.Run("successful run completes without retries", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		calls := 0
		run := func(_ context.Context) error {
			calls++
			return nil
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), run)
			close(events)
		}()

		err := <-errCh
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("retries on failure", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.MaxRetries = 2
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		calls := 0
		run := func(_ context.Context) error {
			calls++
			if calls <= 2 {
				return errors.New("fail")
			}
			return nil
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), run)
			close(events)
		}()

		err := <-errCh
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if calls != 3 {
			t.Errorf("expected 3 calls (2 failures + 1 success), got %d", calls)
		}
	})

	t.Run("gives up after max retries", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.MaxRetries = 2
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		run := func(_ context.Context) error {
			return errors.New("persistent failure")
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), run)
			close(events)
		}()

		err := <-errCh
		if err == nil {
			t.Fatal("expected error after max retries")
		}
		if !strings.Contains(err.Error(), "max retries exceeded") {
			t.Errorf("error should mention max retries, got: %v", err)
		}
	})

	t.Run("context cancellation stops supervision", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(ctx, func(_ context.Context) error {
				return nil
			})
			close(events)
		}()

		err := <-errCh
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	})

	t.Run("saves state file", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), func(_ context.Context) error {
				return nil
			})
			close(events)
		}()

		<-errCh
		state, err := LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if state.RalphPID == 0 {
			t.Error("expected non-zero PID in saved state")
		}
	})

	t.Run("successful run sets passed and timestamps", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), func(_ context.Context) error {
				return nil
			})
			close(events)
		}()

		<-errCh
		state, err := LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if !state.Passed {
			t.Error("expected Passed = true after successful run")
		}
		if state.StartedAt.IsZero() {
			t.Error("expected StartedAt to be set")
		}
		if state.FinishedAt.IsZero() {
			t.Error("expected FinishedAt to be set")
		}
		if state.FinishedAt.Before(state.StartedAt) {
			t.Error("expected FinishedAt >= StartedAt")
		}
	})

	t.Run("failed run sets passed false", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.MaxRetries = 0
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), func(_ context.Context) error {
				return errors.New("fail")
			})
			close(events)
		}()

		<-errCh
		state, err := LoadState(dir)
		if err != nil {
			t.Fatalf("LoadState: %v", err)
		}
		if state.Passed {
			t.Error("expected Passed = false after failed run")
		}
		if state.FinishedAt.IsZero() {
			t.Error("expected FinishedAt to be set on failure")
		}
	})

	t.Run("emits regent events", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		go func() {
			rgt.Supervise(context.Background(), func(_ context.Context) error {
				return nil
			})
			close(events)
		}()

		var regentMsgs []string
		for entry := range events {
			if entry.Kind == loop.LogRegent {
				regentMsgs = append(regentMsgs, entry.Message)
			}
		}

		if len(regentMsgs) == 0 {
			t.Error("expected at least one regent event")
		}

		foundStart := false
		for _, msg := range regentMsgs {
			if strings.Contains(msg, "Starting Ralph") {
				foundStart = true
			}
		}
		if !foundStart {
			t.Errorf("expected 'Starting Ralph' message, got: %v", regentMsgs)
		}
	})
}

func TestHangDetection(t *testing.T) {
	t.Run("kills loop after hang timeout", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.HangTimeoutSeconds = 1 // 1 second timeout for test
		cfg.MaxRetries = 0         // don't retry after hang

		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		run := func(ctx context.Context) error {
			// Block until context is cancelled (simulates a hang)
			<-ctx.Done()
			return ctx.Err()
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), run)
			close(events)
		}()

		err := <-errCh
		if err == nil {
			t.Fatal("expected error from hang detection")
		}

		// Check that a hang detection event was emitted
		var foundHang bool
		for entry := range events {
			if entry.Kind == loop.LogRegent && strings.Contains(entry.Message, "Hang detected") {
				foundHang = true
			}
		}
		if !foundHang {
			t.Error("expected hang detection message")
		}
	})

	t.Run("output resets hang timer", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.HangTimeoutSeconds = 2

		events := make(chan loop.LogEntry, 128)
		rgt := New(cfg, dir, &mockGit{branch: "main"}, events)

		run := func(ctx context.Context) error {
			// Produce output periodically to prevent hang
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
					rgt.NotifyOutput()
				}
			}
			return nil
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- rgt.Supervise(context.Background(), run)
			close(events)
		}()

		err := <-errCh
		if err != nil {
			t.Fatalf("should complete without hang, got: %v", err)
		}
	})
}

func TestUpdateState(t *testing.T) {
	cfg := defaultTestRegentConfig()
	events := make(chan loop.LogEntry, 128)
	rgt := New(cfg, t.TempDir(), &mockGit{}, events)

	tests := []struct {
		name  string
		entry loop.LogEntry
		check func(t *testing.T, r *Regent)
	}{
		{
			name:  "updates iteration",
			entry: loop.LogEntry{Iteration: 5},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.Iteration != 5 {
					t.Errorf("Iteration = %d, want 5", r.state.Iteration)
				}
			},
		},
		{
			name:  "updates total cost",
			entry: loop.LogEntry{TotalCost: 2.50},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.TotalCostUSD != 2.50 {
					t.Errorf("TotalCostUSD = %f, want 2.50", r.state.TotalCostUSD)
				}
			},
		},
		{
			name:  "updates commit",
			entry: loop.LogEntry{Commit: "def456 new feature"},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.LastCommit != "def456 new feature" {
					t.Errorf("LastCommit = %q, want %q", r.state.LastCommit, "def456 new feature")
				}
			},
		},
		{
			name:  "updates branch",
			entry: loop.LogEntry{Branch: "feat/new-feature"},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.Branch != "feat/new-feature" {
					t.Errorf("Branch = %q, want %q", r.state.Branch, "feat/new-feature")
				}
			},
		},
		{
			name:  "updates mode",
			entry: loop.LogEntry{Mode: "build"},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.Mode != "build" {
					t.Errorf("Mode = %q, want %q", r.state.Mode, "build")
				}
			},
		},
		{
			name:  "zero values do not overwrite",
			entry: loop.LogEntry{Message: "just a message"},
			check: func(t *testing.T, r *Regent) {
				r.mu.Lock()
				defer r.mu.Unlock()
				if r.state.Iteration != 5 {
					t.Errorf("Iteration should remain 5, got %d", r.state.Iteration)
				}
				if r.state.Branch != "feat/new-feature" {
					t.Errorf("Branch should remain feat/new-feature, got %q", r.state.Branch)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rgt.UpdateState(tt.entry)
			tt.check(t, rgt)
		})
	}
}

func TestRunPostIterationTests(t *testing.T) {
	t.Run("skipped when rollback disabled", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.RollbackOnTestFailure = false
		events := make(chan loop.LogEntry, 128)
		g := &mockGit{branch: "main", lastCommit: "abc test"}
		rgt := New(cfg, dir, g, events)

		rgt.RunPostIterationTests()
		if len(g.revertCalls) != 0 {
			t.Error("should not revert when rollback is disabled")
		}
	})

	t.Run("skipped when test command is empty", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.RollbackOnTestFailure = true
		cfg.TestCommand = ""
		events := make(chan loop.LogEntry, 128)
		g := &mockGit{branch: "main", lastCommit: "abc test"}
		rgt := New(cfg, dir, g, events)

		rgt.RunPostIterationTests()
		// No panic, no revert = success
		if len(g.revertCalls) != 0 {
			t.Error("should not revert when test command is empty")
		}
	})

	t.Run("passing tests keep commit", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.RollbackOnTestFailure = true
		cfg.TestCommand = "true"
		events := make(chan loop.LogEntry, 128)
		g := &mockGit{branch: "main", lastCommit: "abc good commit"}
		rgt := New(cfg, dir, g, events)

		go func() {
			for range events {
			}
		}()

		rgt.RunPostIterationTests()
		if len(g.revertCalls) != 0 {
			t.Error("should not revert when tests pass")
		}
	})

	t.Run("failing tests trigger revert", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.RollbackOnTestFailure = true
		cfg.TestCommand = "false"
		events := make(chan loop.LogEntry, 128)
		g := &mockGit{branch: "main", lastCommit: "abc1234 bad commit"}
		rgt := New(cfg, dir, g, events)

		go func() {
			for range events {
			}
		}()

		rgt.RunPostIterationTests()
		if len(g.revertCalls) != 1 {
			t.Fatalf("expected 1 revert call, got %d", len(g.revertCalls))
		}
		if g.revertCalls[0] != "abc1234" {
			t.Errorf("reverted %q, want %q", g.revertCalls[0], "abc1234")
		}
		if len(g.pushCalls) != 1 {
			t.Errorf("expected 1 push call after revert, got %d", len(g.pushCalls))
		}
	})

	t.Run("emits events instead of returning errors", func(t *testing.T) {
		dir := t.TempDir()
		cfg := defaultTestRegentConfig()
		cfg.RollbackOnTestFailure = true
		cfg.TestCommand = "true"
		events := make(chan loop.LogEntry, 128)
		g := &mockGit{branch: "main", lastCommit: "abc commit"}
		rgt := New(cfg, dir, g, events)

		rgt.RunPostIterationTests()

		// Drain events and verify regent messages were emitted
		close(events)
		var regentMsgs []string
		for entry := range events {
			if entry.Kind == loop.LogRegent {
				regentMsgs = append(regentMsgs, entry.Message)
			}
		}
		if len(regentMsgs) == 0 {
			t.Error("expected at least one regent event from RunPostIterationTests")
		}

		foundTests := false
		for _, msg := range regentMsgs {
			if strings.Contains(msg, "Tests passed") {
				foundTests = true
			}
		}
		if !foundTests {
			t.Errorf("expected 'Tests passed' message, got: %v", regentMsgs)
		}
	})
}
