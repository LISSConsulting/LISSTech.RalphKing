package tui

import (
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
)

// logEntryMsg wraps a LogEntry for broadcasting to all panels.
type logEntryMsg loop.LogEntry

// loopDoneMsg signals the event channel closed.
type loopDoneMsg struct{}

// loopErrMsg carries an error from the event loop.
type loopErrMsg struct{ err error }

// tickMsg is sent every second for the clock.
type tickMsg time.Time

// iterationLogLoadedMsg carries loaded iteration log data.
type iterationLogLoadedMsg struct {
	Number  int
	Entries []loop.LogEntry
	Summary store.IterationSummary
	Err     error
}

// specsRefreshedMsg carries refreshed spec list after creation/edit.
type specsRefreshedMsg struct{ Specs []spec.SpecFile }
