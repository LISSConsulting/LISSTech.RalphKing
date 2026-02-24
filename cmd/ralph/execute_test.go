package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
)

func TestFormatLogLine(t *testing.T) {
	ts := time.Date(2026, 2, 23, 14, 23, 1, 0, time.UTC)

	tests := []struct {
		name string
		entry loop.LogEntry
		want  string
	}{
		{
			name: "info entry — timestamp and message",
			entry: loop.LogEntry{
				Kind:      loop.LogInfo,
				Timestamp: ts,
				Message:   "starting iteration 3",
			},
			want: "[14:23:01]  starting iteration 3",
		},
		{
			name: "tool use entry — no special prefix",
			entry: loop.LogEntry{
				Kind:      loop.LogToolUse,
				Timestamp: ts,
				Message:   "📖  read_file      app/main.go",
			},
			want: "[14:23:01]  📖  read_file      app/main.go",
		},
		{
			name: "regent entry — shield prefix",
			entry: loop.LogEntry{
				Kind:      loop.LogRegent,
				Timestamp: ts,
				Message:   "Ralph exited (exit 1) — retrying in 30s",
			},
			want: "[14:23:01]  🛡️  Regent: Ralph exited (exit 1) — retrying in 30s",
		},
		{
			name: "error entry — no special prefix",
			entry: loop.LogEntry{
				Kind:      loop.LogError,
				Timestamp: ts,
				Message:   "claude exited with error",
			},
			want: "[14:23:01]  claude exited with error",
		},
		{
			name: "git push entry — no special prefix",
			entry: loop.LogEntry{
				Kind:      loop.LogGitPush,
				Timestamp: ts,
				Message:   "⬇ pushed to origin/main",
			},
			want: "[14:23:01]  ⬇ pushed to origin/main",
		},
		{
			name: "done entry — no special prefix",
			entry: loop.LogEntry{
				Kind:      loop.LogDone,
				Timestamp: ts,
				Message:   "loop finished (5 iterations, $1.42)",
			},
			want: "[14:23:01]  loop finished (5 iterations, $1.42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatLogLine(tt.entry)
			if got != tt.want {
				t.Errorf("formatLogLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassifyResult(t *testing.T) {
	now := time.Now()
	past := now.Add(-10 * time.Minute)

	tests := []struct {
		name  string
		state regent.State
		want  statusResult
	}{
		{
			name:  "no state — zero PID and zero iteration",
			state: regent.State{},
			want:  statusNoState,
		},
		{
			name: "running — started but not finished",
			state: regent.State{
				RalphPID:  123,
				Iteration: 3,
				StartedAt: past,
			},
			want: statusRunning,
		},
		{
			name: "pass — finished with Passed true",
			state: regent.State{
				RalphPID:   123,
				Iteration:  5,
				StartedAt:  past,
				FinishedAt: now,
				Passed:     true,
			},
			want: statusPass,
		},
		{
			name: "fail with consecutive errors",
			state: regent.State{
				RalphPID:        123,
				Iteration:       2,
				StartedAt:       past,
				FinishedAt:      now,
				ConsecutiveErrs: 3,
			},
			want: statusFailWithErrors,
		},
		{
			name: "plain fail — finished but not passed, no consecutive errors",
			state: regent.State{
				RalphPID:   123,
				Iteration:  1,
				StartedAt:  past,
				FinishedAt: now,
			},
			want: statusFail,
		},
		{
			name: "passed wins over consecutive errors",
			state: regent.State{
				RalphPID:        123,
				Iteration:       4,
				StartedAt:       past,
				FinishedAt:      now,
				Passed:          true,
				ConsecutiveErrs: 2,
			},
			want: statusPass,
		},
		{
			name: "running wins over passed",
			state: regent.State{
				RalphPID:  123,
				Iteration: 2,
				StartedAt: past,
				Passed:    true,
			},
			want: statusRunning,
		},
		{
			name: "non-zero PID with zero iteration is no-state",
			state: regent.State{
				RalphPID: 123,
			},
			want: statusNoState,
		},
		{
			name: "zero PID with non-zero iteration is no-state",
			state: regent.State{
				Iteration: 5,
			},
			want: statusNoState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyResult(tt.state)
			if got != tt.want {
				t.Errorf("classifyResult() = %d, want %d", got, tt.want)
			}
		})
	}
}

// fakeFileInfo implements fs.FileInfo for testing needsPlanPhase.
type fakeFileInfo struct {
	size int64
}

func (f fakeFileInfo) Name() string       { return "IMPLEMENTATION_PLAN.md" }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() fs.FileMode  { return 0644 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() any           { return nil }

func TestNeedsPlanPhase(t *testing.T) {
	tests := []struct {
		name    string
		info    fs.FileInfo
		statErr error
		want    bool
	}{
		{
			name:    "file does not exist",
			info:    nil,
			statErr: fs.ErrNotExist,
			want:    true,
		},
		{
			name:    "file exists but empty",
			info:    fakeFileInfo{size: 0},
			statErr: nil,
			want:    true,
		},
		{
			name:    "file exists with content",
			info:    fakeFileInfo{size: 1024},
			statErr: nil,
			want:    false,
		},
		{
			name:    "nil info with nil error (defensive)",
			info:    nil,
			statErr: nil,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := needsPlanPhase(tt.info, tt.statErr)
			if got != tt.want {
				t.Errorf("needsPlanPhase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	now := time.Date(2026, 2, 23, 15, 0, 0, 0, time.UTC)
	started := now.Add(-10 * time.Minute)
	finished := now.Add(-1 * time.Minute)
	lastOutput := now.Add(-30 * time.Second)

	tests := []struct {
		name     string
		state    regent.State
		contains []string
		excludes []string
	}{
		{
			name:     "no state — empty state shows prompt",
			state:    regent.State{},
			contains: []string{"No state found"},
			excludes: []string{"Ralph Status"},
		},
		{
			name: "running — shows elapsed duration and last output",
			state: regent.State{
				RalphPID:     123,
				Iteration:    3,
				Branch:       "feat/test",
				Mode:         "build",
				LastCommit:   "abc1234",
				TotalCostUSD: 0.42,
				StartedAt:    started,
				LastOutputAt: lastOutput,
			},
			contains: []string{
				"Ralph Status",
				"Branch:",
				"feat/test",
				"Mode:",
				"build",
				"Last commit:",
				"abc1234",
				"Iteration:",
				"3",
				"$0.42",
				"10m0s (running)",
				"30s ago",
				"Result:",
				"running",
			},
		},
		{
			name: "pass — shows duration and pass result",
			state: regent.State{
				RalphPID:     123,
				Iteration:    5,
				Branch:       "main",
				TotalCostUSD: 1.50,
				StartedAt:    started,
				FinishedAt:   finished,
				Passed:       true,
			},
			contains: []string{
				"Ralph Status",
				"main",
				"Iteration:",
				"5",
				"$1.50",
				"9m0s",
				"Result:",
				"pass",
			},
			excludes: []string{"running", "fail", "Last output:"},
		},
		{
			name: "fail with consecutive errors — shows error count",
			state: regent.State{
				RalphPID:        123,
				Iteration:       2,
				TotalCostUSD:    0.30,
				StartedAt:       started,
				FinishedAt:      finished,
				ConsecutiveErrs: 3,
			},
			contains: []string{
				"Ralph Status",
				"fail (3 consecutive errors)",
			},
			excludes: []string{"pass", "running"},
		},
		{
			name: "plain fail — finished but not passed, no errors",
			state: regent.State{
				RalphPID:   123,
				Iteration:  1,
				StartedAt:  started,
				FinishedAt: finished,
			},
			contains: []string{
				"Ralph Status",
				"Result:",
				"fail",
			},
			excludes: []string{"pass", "running", "consecutive"},
		},
		{
			name: "optional fields omitted when empty",
			state: regent.State{
				RalphPID:   123,
				Iteration:  1,
				StartedAt:  started,
				FinishedAt: finished,
			},
			excludes: []string{"Branch:", "Mode:", "Last commit:", "Last output:"},
		},
		{
			name: "running without last output — omits last output line",
			state: regent.State{
				RalphPID:  123,
				Iteration: 1,
				StartedAt: started,
			},
			contains: []string{"running"},
			excludes: []string{"Last output:"},
		},
		{
			name: "zero cost displays as $0.00",
			state: regent.State{
				RalphPID:   123,
				Iteration:  1,
				StartedAt:  started,
				FinishedAt: finished,
			},
			contains: []string{"$0.00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatus(tt.state, now)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output should contain %q\ngot:\n%s", want, got)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(got, exclude) {
					t.Errorf("output should NOT contain %q\ngot:\n%s", exclude, got)
				}
			}
		})
	}
}

// ---- Integration tests for executeLoop and executeSmartRun ----
//
// These tests exercise the full orchestration path through config loading,
// wiring, and loop setup. They fail before any claude invocation
// (config not found, invalid config, or prompt file missing), so they
// work without a real Claude binary installed.

// testConfigNoRegent returns a minimal ralph.toml with regent disabled and
// git ops turned off so tests don't attempt network operations.
func testConfigNoRegent() string {
	return `[plan]
prompt_file = "PROMPT_plan.md"
max_iterations = 1

[build]
prompt_file = "PROMPT_build.md"
max_iterations = 1

[git]
auto_pull_rebase = false
auto_push = false

[regent]
enabled = false
`
}

// testConfigWithRegent returns a ralph.toml with regent enabled but
// max_retries=0 so it fails fast after one error without backoff.
func testConfigWithRegent() string {
	return `[plan]
prompt_file = "PROMPT_plan.md"
max_iterations = 1

[build]
prompt_file = "PROMPT_build.md"
max_iterations = 1

[git]
auto_pull_rebase = false
auto_push = false

[regent]
enabled = true
max_retries = 0
retry_backoff_seconds = 0
hang_timeout_seconds = 0
rollback_on_test_failure = false
`
}

// writeExecTestFile writes content to dir/name, creating parent directories.
func writeExecTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", name, err)
	}
}

func TestExecuteLoop_ConfigNotFound(t *testing.T) {
	// Isolated temp dir with no ralph.toml anywhere in its ancestor tree.
	t.Chdir(t.TempDir())

	err := executeLoop(loop.ModePlan, 1, true)
	if err == nil {
		t.Fatal("expected error when ralph.toml not found")
	}
	if !strings.Contains(err.Error(), "ralph.toml") {
		t.Errorf("error should mention ralph.toml, got: %v", err)
	}
}

func TestExecuteLoop_ConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// Empty plan.prompt_file fails Validate()
	writeExecTestFile(t, dir, "ralph.toml", "[plan]\nprompt_file = \"\"\n[build]\nprompt_file = \"b.md\"\n")

	err := executeLoop(loop.ModePlan, 1, true)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "config validation") {
		t.Errorf("error should mention config validation, got: %v", err)
	}
}

func TestExecuteLoop_RegentDisabled_PromptMissing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeExecTestFile(t, dir, "ralph.toml", testConfigNoRegent())
	// PROMPT_plan.md intentionally absent — loop.Run fails reading it.

	err := executeLoop(loop.ModePlan, 1, true)
	if err == nil {
		t.Fatal("expected error when prompt file missing")
	}
	if !strings.Contains(err.Error(), "PROMPT_plan.md") {
		t.Errorf("error should mention PROMPT_plan.md, got: %v", err)
	}
}

func TestExecuteLoop_RegentEnabled_PromptMissing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeExecTestFile(t, dir, "ralph.toml", testConfigWithRegent())
	// PROMPT_plan.md intentionally absent.
	// Regent gives up after 0 retries and returns max-retries error.

	err := executeLoop(loop.ModePlan, 1, true)
	if err == nil {
		t.Fatal("expected error — Regent should give up (max_retries=0)")
	}
}

// ---- executeSmartRun integration tests ----

func TestExecuteSmartRun_ConfigNotFound(t *testing.T) {
	t.Chdir(t.TempDir())

	err := executeSmartRun(1, true)
	if err == nil {
		t.Fatal("expected error when ralph.toml not found")
	}
	if !strings.Contains(err.Error(), "ralph.toml") {
		t.Errorf("error should mention ralph.toml, got: %v", err)
	}
}

func TestExecuteSmartRun_NeedsPlan_PromptMissing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeExecTestFile(t, dir, "ralph.toml", testConfigNoRegent())
	// No IMPLEMENTATION_PLAN.md → needsPlanPhase returns true.
	// No PROMPT_plan.md → plan phase fails reading it.

	err := executeSmartRun(1, true)
	if err == nil {
		t.Fatal("expected error when plan prompt file missing")
	}
	if !strings.Contains(err.Error(), "PROMPT_plan.md") {
		t.Errorf("error should mention PROMPT_plan.md, got: %v", err)
	}
}

func TestExecuteSmartRun_SkipPlan_BuildPromptMissing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeExecTestFile(t, dir, "ralph.toml", testConfigNoRegent())
	// Non-empty IMPLEMENTATION_PLAN.md → needsPlanPhase returns false.
	writeExecTestFile(t, dir, "IMPLEMENTATION_PLAN.md", "# Plan\n\nSome content.\n")
	// PROMPT_build.md absent → build loop fails reading it.

	err := executeSmartRun(1, true)
	if err == nil {
		t.Fatal("expected error when build prompt file missing")
	}
	if !strings.Contains(err.Error(), "PROMPT_build.md") {
		t.Errorf("error should mention PROMPT_build.md, got: %v", err)
	}
}
