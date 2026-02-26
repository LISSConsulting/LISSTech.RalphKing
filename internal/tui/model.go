package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	lines        []logLine
	width        int
	height       int
	scrollOffset int // 0 = at bottom (auto-scroll), >0 = scrolled up N lines
	newBelow     int // count of new messages that arrived while scrolled up

	// Accent-dependent styles (configured per instance)
	accentHeaderStyle lipgloss.Style
	accentGitStyle    lipgloss.Style

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
// accentColor is a hex color string (e.g. "#7D56F4") for the header and
// accent elements. If empty, the default indigo is used.
func New(events <-chan loop.LogEntry, accentColor string) Model {
	accent := lipgloss.Color(defaultAccentColor)
	if accentColor != "" {
		accent = lipgloss.Color(accentColor)
	}
	return Model{
		events: events,
		width:  80,
		height: 24,
		accentHeaderStyle: lipgloss.NewStyle().
			Background(accent).
			Foreground(colorWhite).
			Bold(true).
			Padding(0, 1),
		accentGitStyle: lipgloss.NewStyle().
			Foreground(accent),
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

// logHeight returns the number of lines available for the log panel.
func (m Model) logHeight() int {
	h := m.height - 2 // 1 header + 1 footer
	if h < 1 {
		h = 1
	}
	return h
}

// maxScrollOffset returns the maximum valid scroll offset.
func (m Model) maxScrollOffset() int {
	max := len(m.lines) - m.logHeight()
	if max < 0 {
		return 0
	}
	return max
}

// clampScroll ensures scrollOffset is within valid bounds.
func (m *Model) clampScroll() {
	max := m.maxScrollOffset()
	if m.scrollOffset > max {
		m.scrollOffset = max
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
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
