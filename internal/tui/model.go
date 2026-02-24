package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// logLine is a rendered log entry for display.
type logLine struct {
	entry loop.LogEntry
}

// Model is the bubbletea model for the Ralph TUI.
type Model struct {
	events <-chan loop.LogEntry

	// Display state
	lines []logLine
	width int
	height int

	// Loop state
	mode       string
	branch     string
	iteration  int
	maxIter    int
	totalCost  float64
	lastCommit string
	done       bool
	err        error
}

// logEntryMsg wraps a LogEntry as a bubbletea message.
type logEntryMsg loop.LogEntry

// loopDoneMsg signals the event channel has closed.
type loopDoneMsg struct{}

// loopErrMsg carries a loop error back to the TUI.
type loopErrMsg struct{ err error }

// New creates a new TUI Model that consumes events from the given channel.
func New(events <-chan loop.LogEntry) Model {
	return Model{
		events: events,
		width:  80,
		height: 24,
	}
}

// Init returns the initial command: start listening for events.
func (m Model) Init() tea.Cmd {
	return waitForEvent(m.events)
}

// Err returns any error that occurred during the loop.
func (m Model) Err() error {
	return m.err
}

// waitForEvent returns a command that blocks on the event channel.
func waitForEvent(ch <-chan loop.LogEntry) tea.Cmd {
	return func() tea.Msg {
		entry, ok := <-ch
		if !ok {
			return loopDoneMsg{}
		}
		return logEntryMsg(entry)
	}
}
