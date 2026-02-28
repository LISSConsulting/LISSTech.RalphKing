package tui

// FocusTarget identifies which panel currently holds keyboard focus.
type FocusTarget int

const (
	FocusSpecs      FocusTarget = iota // Left sidebar — specs list
	FocusIterations                    // Left sidebar — iterations list
	FocusMain                          // Right top — main content panel
	FocusSecondary                     // Right bottom — secondary panel
)

// Next returns the next focus target in forward tab order.
func (f FocusTarget) Next() FocusTarget {
	return (f + 1) % 4
}

// Prev returns the previous focus target in reverse tab order.
func (f FocusTarget) Prev() FocusTarget {
	return (f + 3) % 4 // equivalent to (f - 1 + 4) % 4
}

// String returns the human-readable name of the focus target.
func (f FocusTarget) String() string {
	switch f {
	case FocusSpecs:
		return "specs"
	case FocusIterations:
		return "iterations"
	case FocusMain:
		return "main"
	case FocusSecondary:
		return "secondary"
	default:
		return "unknown"
	}
}

// LoopState represents the current state of the Ralph loop.
type LoopState int

const (
	StateIdle          LoopState = iota // No loop running
	StatePlanning                       // Plan loop running
	StateBuilding                       // Build loop running
	StateFailed                         // Last run ended in failure
	StateRegentRestart                  // Regent is restarting the loop
)

// validTransitions defines the allowed LoopState transitions.
var validTransitions = map[LoopState][]LoopState{
	StateIdle:          {StatePlanning, StateBuilding},
	StatePlanning:      {StateBuilding, StateFailed, StateIdle, StateRegentRestart},
	StateBuilding:      {StateFailed, StateIdle, StateRegentRestart},
	StateFailed:        {StateIdle, StateRegentRestart, StateBuilding, StatePlanning},
	StateRegentRestart: {StateBuilding, StatePlanning, StateFailed, StateIdle},
}

// CanTransitionTo reports whether transitioning from s to next is valid.
func (s LoopState) CanTransitionTo(next LoopState) bool {
	for _, valid := range validTransitions[s] {
		if valid == next {
			return true
		}
	}
	return false
}

// Label returns a short uppercase label for the state.
func (s LoopState) Label() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StatePlanning:
		return "PLANNING"
	case StateBuilding:
		return "BUILDING"
	case StateFailed:
		return "FAILED"
	case StateRegentRestart:
		return "REGENT RESTART"
	default:
		return "UNKNOWN"
	}
}

// Symbol returns a single-character symbol representing the state.
func (s LoopState) Symbol() string {
	switch s {
	case StateIdle:
		return "✓"
	case StatePlanning:
		return "●"
	case StateBuilding:
		return "●"
	case StateFailed:
		return "✗"
	case StateRegentRestart:
		return "⟳"
	default:
		return "?"
	}
}
