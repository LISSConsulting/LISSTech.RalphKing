package main

import (
	"io/fs"
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
