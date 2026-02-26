package tui

import (
	"time"

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

	// Project identity
	projectName string // from ralph.toml [project].name; empty falls back to "RalphKing"
	workDir     string // working directory; shown in header when non-empty

	// Graceful stop support
	requestStop  func() // called once when user presses 's'; provided by wiring
	stopRequested bool  // true after first 's' press

	// Loop state
	mode         string
	branch       string
	iteration    int
	maxIter      int
	totalCost    float64
	lastDuration float64 // seconds; 0 until first iteration completes
	lastCommit   string
	done         bool
	err          error

	// Time tracking
	startedAt time.Time // when the TUI was initialized
	now       time.Time // updated every second by tickMsg
}

// logEntryMsg wraps a LogEntry as a bubbletea message.
type logEntryMsg loop.LogEntry

// loopDoneMsg signals the event channel has closed.
type loopDoneMsg struct{}

// loopErrMsg carries a loop error back to the TUI.
type loopErrMsg struct{ err error }

// tickMsg is sent every second to update the clock display.
type tickMsg time.Time

// New creates a new TUI Model that consumes events from the given channel.
// accentColor is a hex color string (e.g. "#7D56F4") for the header and
// accent elements. If empty, the default indigo is used.
// projectName is displayed in the header; if empty, "RalphKing" is shown.
// workDir is the working directory shown in the header; omitted when empty.
// requestStop, if non-nil, is called once when the user presses 's' to request
// a graceful stop after the current iteration.
func New(events <-chan loop.LogEntry, accentColor, projectName, workDir string, requestStop func()) Model {
	accent := lipgloss.Color(defaultAccentColor)
	if accentColor != "" {
		accent = lipgloss.Color(accentColor)
	}
	now := time.Now()
	return Model{
		events:      events,
		width:       80,
		height:      24,
		startedAt:   now,
		now:         now,
		projectName: projectName,
		workDir:     workDir,
		requestStop: requestStop,
		accentHeaderStyle: lipgloss.NewStyle().
			Background(accent).
			Foreground(colorWhite).
			Bold(true).
			Padding(0, 1),
		accentGitStyle: lipgloss.NewStyle().
			Foreground(accent),
	}
}

// Init returns the initial commands: start listening for events and start the clock ticker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.events), tickCmd())
}

// tickCmd schedules the next one-second clock tick.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
