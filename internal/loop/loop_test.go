package loop

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/claude"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
)

// mockAgent is a test double for claude.Agent.
type mockAgent struct {
	events []claude.Event
	err    error
	calls  int
}

func (m *mockAgent) Run(_ context.Context, _ string, _ claude.RunOptions) (<-chan claude.Event, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	ch := make(chan claude.Event, len(m.events))
	for _, ev := range m.events {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

// mockGit is a test double for GitOps.
type mockGit struct {
	branch         string
	dirty          bool
	diffFromRemote bool
	pullErr        error
	pushErr        error
	stashErr       error
	stashPopErr    error
	dirtyErr       error // error returned by HasUncommittedChanges
	diffErr        error // error returned by DiffFromRemote
	lastCommit     string

	pullCalls     int
	pushCalls     int
	stashCalls    int
	stashPopCalls int
}

func (m *mockGit) CurrentBranch() (string, error)       { return m.branch, nil }
func (m *mockGit) HasUncommittedChanges() (bool, error)  { return m.dirty, m.dirtyErr }
func (m *mockGit) Pull(_ string) error                   { m.pullCalls++; return m.pullErr }
func (m *mockGit) Push(_ string) error                   { m.pushCalls++; return m.pushErr }
func (m *mockGit) Stash() error                          { m.stashCalls++; return m.stashErr }
func (m *mockGit) StashPop() error                       { m.stashPopCalls++; return m.stashPopErr }
func (m *mockGit) LastCommit() (string, error)           { return m.lastCommit, nil }
func (m *mockGit) DiffFromRemote(_ string) (bool, error) { return m.diffFromRemote, m.diffErr }

func defaultTestConfig() *config.Config {
	cfg := config.Defaults()
	return &cfg
}

func setupTestLoop(t *testing.T, agent claude.Agent, git GitOps, cfg *config.Config) (*Loop, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()

	// Write prompt files
	if err := os.WriteFile(filepath.Join(dir, cfg.Plan.PromptFile), []byte("plan prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cfg.Build.PromptFile), []byte("build prompt"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	return &Loop{
		Agent:  agent,
		Git:    git,
		Config: cfg,
		Log:    &buf,
		Dir:    dir,
	}, &buf
}

func TestRun(t *testing.T) {
	t.Run("plan mode runs configured iterations", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 2.5, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc123 initial"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 2

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if agent.calls != 2 {
			t.Errorf("expected 2 agent calls, got %d", agent.calls)
		}
	})

	t.Run("build mode with max override", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.05, 1.0, "success")},
		}
		git := &mockGit{branch: "feat/test", lastCommit: "def456 test"}
		cfg := defaultTestConfig()
		cfg.Build.MaxIterations = 0 // unlimited in config

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModeBuild, 3)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if agent.calls != 3 {
			t.Errorf("expected 3 agent calls (override), got %d", agent.calls)
		}
	})

	t.Run("missing prompt file returns error", func(t *testing.T) {
		agent := &mockAgent{}
		git := &mockGit{branch: "main"}
		cfg := defaultTestConfig()
		cfg.Plan.PromptFile = "nonexistent.md"

		dir := t.TempDir()
		var buf bytes.Buffer
		lp := &Loop{Agent: agent, Git: git, Config: cfg, Log: &buf, Dir: dir}

		err := lp.Run(context.Background(), ModePlan, 0)
		if err == nil {
			t.Fatal("expected error for missing prompt file")
		}
		if !strings.Contains(err.Error(), "nonexistent.md") {
			t.Errorf("error should mention file name, got: %v", err)
		}
	})

	t.Run("context cancellation stops loop", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc123 test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 100

		lp, _ := setupTestLoop(t, agent, git, cfg)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		err := lp.Run(ctx, ModePlan, 0)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
		if agent.calls != 0 {
			t.Errorf("expected 0 agent calls after cancel, got %d", agent.calls)
		}
	})

	t.Run("agent error propagates", func(t *testing.T) {
		agent := &mockAgent{err: errors.New("agent failed")}
		git := &mockGit{branch: "main"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err == nil {
			t.Fatal("expected error from agent failure")
		}
		if !strings.Contains(err.Error(), "agent failed") {
			t.Errorf("error should contain agent message, got: %v", err)
		}
	})
}

func TestIteration(t *testing.T) {
	t.Run("pulls before running claude", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Git.AutoPullRebase = true
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.pullCalls != 1 {
			t.Errorf("expected 1 pull call, got %d", git.pullCalls)
		}
	})

	t.Run("skips pull when auto_pull_rebase is false", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Git.AutoPullRebase = false
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.pullCalls != 0 {
			t.Errorf("expected 0 pull calls, got %d", git.pullCalls)
		}
	})

	t.Run("stashes dirty working tree", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", dirty: true, lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.stashCalls != 1 {
			t.Errorf("expected 1 stash call, got %d", git.stashCalls)
		}
		if git.stashPopCalls != 1 {
			t.Errorf("expected 1 stash pop call, got %d", git.stashPopCalls)
		}
	})

	t.Run("skips stash when working tree is clean", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", dirty: false, lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.stashCalls != 0 {
			t.Errorf("expected 0 stash calls, got %d", git.stashCalls)
		}
		if git.stashPopCalls != 0 {
			t.Errorf("expected 0 stash pop calls, got %d", git.stashPopCalls)
		}
	})

	t.Run("pushes when there are new local commits", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:         "main",
			diffFromRemote: true,
			lastCommit:     "abc new commit",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = true
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.pushCalls != 1 {
			t.Errorf("expected 1 push call, got %d", git.pushCalls)
		}
	})

	t.Run("skips push when no new commits", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:         "main",
			diffFromRemote: false,
			lastCommit:     "abc same",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = true
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.pushCalls != 0 {
			t.Errorf("expected 0 push calls, got %d", git.pushCalls)
		}
	})

	t.Run("skips push when auto_push is false", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:         "main",
			diffFromRemote: true,
			lastCommit:     "abc new commit",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = false
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if git.pushCalls != 0 {
			t.Errorf("expected 0 push calls when auto_push=false, got %d", git.pushCalls)
		}
	})
}

func TestLogOutput(t *testing.T) {
	t.Run("logs tool use events", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{
				claude.ToolUseEvent("read_file", map[string]any{"file_path": "main.go"}),
				claude.ResultEvent(0.10, 1.0, "success"),
			},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "read_file") {
			t.Error("log should contain tool name")
		}
		if !strings.Contains(output, "main.go") {
			t.Error("log should contain tool input summary")
		}
	})

	t.Run("logs error events", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{
				claude.ErrorEvent("something went wrong"),
				claude.ResultEvent(0.00, 0.5, "success"),
			},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "something went wrong") {
			t.Error("log should contain error message")
		}
	})
}

func TestSubtypeInLogOutput(t *testing.T) {
	t.Run("subtype included in iteration complete message", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.14, 4.2, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "success") {
			t.Errorf("log should contain subtype 'success', got: %s", buf.String())
		}
	})

	t.Run("error_max_turns subtype in log", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.30, 5.0, "error_max_turns")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "error_max_turns") {
			t.Errorf("log should contain subtype 'error_max_turns', got: %s", buf.String())
		}
	})

	t.Run("empty subtype omitted from message", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Iteration 1 complete") {
			t.Errorf("should contain iteration complete, got: %s", output)
		}
		// Should not have trailing separator for empty subtype
		if strings.Contains(output, "1.0s —") {
			t.Errorf("should not show trailing separator for empty subtype, got: %s", output)
		}
	})
}

func TestSubtypeInEventChannel(t *testing.T) {
	ch := make(chan LogEntry, 16)
	agent := &mockAgent{
		events: []claude.Event{claude.ResultEvent(0.14, 4.2, "success")},
	}
	git := &mockGit{branch: "main", lastCommit: "abc test"}
	cfg := defaultTestConfig()
	cfg.Plan.MaxIterations = 1

	lp, _ := setupTestLoop(t, agent, git, cfg)
	lp.Events = ch

	err := lp.Run(context.Background(), ModePlan, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	close(ch)
	var iterComplete *LogEntry
	for e := range ch {
		if e.Kind == LogIterComplete {
			iterComplete = &e
			break
		}
	}

	if iterComplete == nil {
		t.Fatal("expected an IterComplete event")
	}
	if iterComplete.Subtype != "success" {
		t.Errorf("expected Subtype 'success', got %q", iterComplete.Subtype)
	}
}

func TestModeConfig(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Plan.PromptFile = "PLAN.md"
	cfg.Plan.MaxIterations = 5
	cfg.Build.PromptFile = "BUILD.md"
	cfg.Build.MaxIterations = 10

	lp := &Loop{Config: cfg}

	t.Run("plan mode", func(t *testing.T) {
		file, max := lp.modeConfig(ModePlan)
		if file != "PLAN.md" {
			t.Errorf("expected PLAN.md, got %s", file)
		}
		if max != 5 {
			t.Errorf("expected 5, got %d", max)
		}
	})

	t.Run("build mode", func(t *testing.T) {
		file, max := lp.modeConfig(ModeBuild)
		if file != "BUILD.md" {
			t.Errorf("expected BUILD.md, got %s", file)
		}
		if max != 10 {
			t.Errorf("expected 10, got %d", max)
		}
	})
}

func TestSummarizeInput(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{"file_path", map[string]any{"file_path": "main.go"}, "main.go"},
		{"command", map[string]any{"command": "go build"}, "go build"},
		{"path", map[string]any{"path": "/tmp"}, "/tmp"},
		{"empty", map[string]any{"other": "value"}, ""},
		{"nil", nil, ""},
		{"prefers file_path", map[string]any{"file_path": "a.go", "command": "ls"}, "a.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeInput(tt.input)
			if got != tt.want {
				t.Errorf("summarizeInput(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPostIteration(t *testing.T) {
	t.Run("hook called after each iteration", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 3

		lp, _ := setupTestLoop(t, agent, git, cfg)
		var hookCalls int
		lp.PostIteration = func() { hookCalls++ }

		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hookCalls != 3 {
			t.Errorf("expected PostIteration called 3 times, got %d", hookCalls)
		}
	})

	t.Run("hook not called when not set", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{branch: "main", lastCommit: "abc test"}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		// PostIteration is nil by default

		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// No panic = success
	})

	t.Run("hook called after push", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:         "main",
			diffFromRemote: true,
			lastCommit:     "abc new",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = true
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		var hookCalled bool
		var pushCountAtHook int
		lp.PostIteration = func() {
			hookCalled = true
			pushCountAtHook = git.pushCalls
		}

		err := lp.Run(context.Background(), ModePlan, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hookCalled {
			t.Fatal("PostIteration was not called")
		}
		if pushCountAtHook != 1 {
			t.Errorf("expected push to happen before hook, pushCalls at hook time = %d", pushCountAtHook)
		}
	})
}

func TestIterLabel(t *testing.T) {
	tests := []struct {
		max  int
		want string
	}{
		{0, "unlimited"},
		{1, "1"},
		{10, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := iterLabel(tt.max); got != tt.want {
				t.Errorf("iterLabel(%d) = %q, want %q", tt.max, got, tt.want)
			}
		})
	}
}

func TestStashIfDirtyErrors(t *testing.T) {
	t.Run("HasUncommittedChanges error propagates", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:   "main",
			dirtyErr: errors.New("git status failed"),
		}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err == nil {
			t.Fatal("expected error when HasUncommittedChanges fails")
		}
		if !strings.Contains(err.Error(), "git status failed") {
			t.Errorf("error should contain cause, got: %v", err)
		}
	})

	t.Run("Stash error propagates", func(t *testing.T) {
		agent := &mockAgent{}
		git := &mockGit{
			branch:   "main",
			dirty:    true,
			stashErr: errors.New("stash failed"),
		}
		cfg := defaultTestConfig()
		cfg.Plan.MaxIterations = 1

		lp, _ := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err == nil {
			t.Fatal("expected error when Stash fails")
		}
		if !strings.Contains(err.Error(), "stash failed") {
			t.Errorf("error should contain cause, got: %v", err)
		}
	})
}

func TestPushIfNeededErrors(t *testing.T) {
	t.Run("DiffFromRemote error pushes anyway and logs", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:     "main",
			diffErr:    errors.New("diff failed"),
			lastCommit: "abc new",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = true
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("DiffFromRemote error should not abort loop, got: %v", err)
		}
		if git.pushCalls != 1 {
			t.Errorf("should push when diff check fails (e.g., new branch), got %d push calls", git.pushCalls)
		}
		if !strings.Contains(buf.String(), "Diff check failed") {
			t.Errorf("should log diff failure, got: %s", buf.String())
		}
	})

	t.Run("Push error logged but does not abort loop", func(t *testing.T) {
		agent := &mockAgent{
			events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
		}
		git := &mockGit{
			branch:         "main",
			diffFromRemote: true,
			pushErr:        errors.New("push rejected"),
			lastCommit:     "abc new",
		}
		cfg := defaultTestConfig()
		cfg.Git.AutoPush = true
		cfg.Plan.MaxIterations = 1

		lp, buf := setupTestLoop(t, agent, git, cfg)
		err := lp.Run(context.Background(), ModePlan, 0)

		if err != nil {
			t.Fatalf("push error should not abort loop, got: %v", err)
		}
		if !strings.Contains(buf.String(), "Push error") {
			t.Errorf("should log push error, got: %s", buf.String())
		}
	})
}

func TestIterationContinuesOnPullError(t *testing.T) {
	agent := &mockAgent{
		events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
	}
	git := &mockGit{
		branch:     "main",
		lastCommit: "abc test",
		pullErr:    errors.New("network error"),
	}
	cfg := defaultTestConfig()
	cfg.Git.AutoPullRebase = true
	cfg.Plan.MaxIterations = 1

	lp, buf := setupTestLoop(t, agent, git, cfg)
	err := lp.Run(context.Background(), ModePlan, 0)

	// Pull error is logged but loop continues
	if err != nil {
		t.Fatalf("pull error should not abort loop, got: %v", err)
	}
	if agent.calls != 1 {
		t.Errorf("expected 1 agent call despite pull error, got %d", agent.calls)
	}
	if !strings.Contains(buf.String(), "Pull failed") {
		t.Errorf("log should mention pull failure, got: %s", buf.String())
	}
}

func TestIterationContinuesOnStashPopError(t *testing.T) {
	agent := &mockAgent{
		events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
	}
	git := &mockGit{
		branch:      "main",
		dirty:       true,
		lastCommit:  "abc test",
		stashPopErr: errors.New("stash pop conflict"),
	}
	cfg := defaultTestConfig()
	cfg.Plan.MaxIterations = 1

	lp, buf := setupTestLoop(t, agent, git, cfg)
	err := lp.Run(context.Background(), ModePlan, 0)

	// Stash pop error is logged but loop continues
	if err != nil {
		t.Fatalf("stash pop error should not abort loop, got: %v", err)
	}
	if agent.calls != 1 {
		t.Errorf("expected 1 agent call despite stash pop error, got %d", agent.calls)
	}
	if !strings.Contains(buf.String(), "Stash pop failed") {
		t.Errorf("log should mention stash pop failure, got: %s", buf.String())
	}
}

func TestEmitNilLog(t *testing.T) {
	// When Events is nil and Log is nil, emit should fall back to os.Stdout without panic.
	agent := &mockAgent{
		events: []claude.Event{claude.ResultEvent(0.10, 1.0, "success")},
	}
	git := &mockGit{branch: "main", lastCommit: "abc test"}
	cfg := defaultTestConfig()
	cfg.Plan.MaxIterations = 1

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, cfg.Plan.PromptFile), []byte("plan prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cfg.Build.PromptFile), []byte("build prompt"), 0644); err != nil {
		t.Fatal(err)
	}

	lp := &Loop{
		Agent:  agent,
		Git:    git,
		Config: cfg,
		Dir:    dir,
		// Log is nil, Events is nil — emit should use os.Stdout
	}

	err := lp.Run(context.Background(), ModePlan, 0)
	if err != nil {
		t.Fatalf("nil Log should fall back to stdout without error, got: %v", err)
	}
}
