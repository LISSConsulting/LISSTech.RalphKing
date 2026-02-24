package main

import (
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
)

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
