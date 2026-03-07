package tui

import "testing"

func TestFocusTarget_Next(t *testing.T) {
	tests := []struct {
		name  string
		input FocusTarget
		want  FocusTarget
	}{
		{"specs → iterations", FocusSpecs, FocusIterations},
		{"iterations → main", FocusIterations, FocusMain},
		{"main → secondary", FocusMain, FocusSecondary},
		{"secondary wraps → specs", FocusSecondary, FocusSpecs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Next()
			if got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFocusTarget_Prev(t *testing.T) {
	tests := []struct {
		name  string
		input FocusTarget
		want  FocusTarget
	}{
		{"specs wraps → secondary", FocusSpecs, FocusSecondary},
		{"iterations → specs", FocusIterations, FocusSpecs},
		{"main → iterations", FocusMain, FocusIterations},
		{"secondary → main", FocusSecondary, FocusMain},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Prev()
			if got != tt.want {
				t.Errorf("Prev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFocusTarget_String(t *testing.T) {
	tests := []struct {
		input FocusTarget
		want  string
	}{
		{FocusSpecs, "specs"},
		{FocusIterations, "iterations"},
		{FocusMain, "main"},
		{FocusSecondary, "secondary"},
		{FocusTarget(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFocusTarget_CycleFullRound(t *testing.T) {
	start := FocusSpecs
	cur := start
	for i := 0; i < 4; i++ {
		cur = cur.Next()
	}
	if cur != start {
		t.Errorf("4 Next() calls did not return to start: got %v", cur)
	}
}

func TestLoopState_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		from LoopState
		to   LoopState
		want bool
	}{
		// StateIdle
		{StateIdle, StatePlanning, true},
		{StateIdle, StateBuilding, true},
		{StateIdle, StateFailed, false},
		{StateIdle, StateRegentRestart, false},
		// StatePlanning
		{StatePlanning, StateBuilding, true},
		{StatePlanning, StateFailed, true},
		{StatePlanning, StateIdle, true},
		{StatePlanning, StateRegentRestart, true},
		// StateBuilding
		{StateBuilding, StateFailed, true},
		{StateBuilding, StateIdle, true},
		{StateBuilding, StateRegentRestart, true},
		{StateBuilding, StatePlanning, false},
		// StateFailed
		{StateFailed, StateIdle, true},
		{StateFailed, StateRegentRestart, true},
		{StateFailed, StateBuilding, true},
		{StateFailed, StatePlanning, true},
		// StateRegentRestart
		{StateRegentRestart, StateBuilding, true},
		{StateRegentRestart, StatePlanning, true},
		{StateRegentRestart, StateFailed, true},
		{StateRegentRestart, StateIdle, true},
		// Invalid same-state transitions
		{StateIdle, StateIdle, false},
		{StateBuilding, StateBuilding, false},
	}

	for _, tt := range tests {
		name := tt.from.Label() + "→" + tt.to.Label()
		t.Run(name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%v) = %v, want %v", tt.to, got, tt.want)
			}
		})
	}
}

func TestLoopState_Label(t *testing.T) {
	tests := []struct {
		input LoopState
		want  string
	}{
		{StateIdle, "IDLE"},
		{StatePlanning, "PLANNING"},
		{StateBuilding, "BUILDING"},
		{StateFailed, "FAILED"},
		{StateRegentRestart, "REGENT RESTART"},
		{LoopState(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.Label()
			if got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoopState_Symbol(t *testing.T) {
	tests := []struct {
		input LoopState
		want  string
	}{
		{StateIdle, "✓"},
		{StatePlanning, "●"},
		{StateBuilding, "●"},
		{StateFailed, "✗"},
		{StateRegentRestart, "⟳"},
		{LoopState(99), "?"},
	}
	for _, tt := range tests {
		t.Run(tt.input.Label(), func(t *testing.T) {
			got := tt.input.Symbol()
			if got != tt.want {
				t.Errorf("Symbol() = %q, want %q", got, tt.want)
			}
		})
	}
}
