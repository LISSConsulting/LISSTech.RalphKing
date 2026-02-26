package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// Update handles incoming messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case logEntryMsg:
		return m.handleLogEntry(msg)

	case tickMsg:
		m.now = time.Time(msg)
		return m, tickCmd()

	case loopDoneMsg:
		m.done = true
		return m, tea.Quit

	case loopErrMsg:
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "s":
		if m.requestStop != nil && !m.stopRequested {
			m.stopRequested = true
			m.requestStop()
		}
	case "up", "k":
		if m.scrollOffset < m.maxScrollOffset() {
			m.scrollOffset++
		}
	case "down", "j":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "pgup":
		m.scrollOffset += m.logHeight()
		m.clampScroll()
	case "pgdown":
		m.scrollOffset -= m.logHeight()
		m.clampScroll()
	case "home", "g":
		m.scrollOffset = m.maxScrollOffset()
	case "end", "G":
		m.scrollOffset = 0
	}
	// Reset new-messages counter when scrolled back to bottom
	if m.scrollOffset == 0 {
		m.newBelow = 0
	}
	return m, nil
}

func (m Model) handleLogEntry(msg logEntryMsg) (tea.Model, tea.Cmd) {
	entry := (loop.LogEntry)(msg)

	// Update state from entry metadata
	if entry.Branch != "" {
		m.branch = entry.Branch
	}
	if entry.Mode != "" {
		m.mode = entry.Mode
	}
	if entry.MaxIter > 0 {
		m.maxIter = entry.MaxIter
	}
	if entry.Iteration > 0 {
		m.iteration = entry.Iteration
	}
	if entry.TotalCost > 0 {
		m.totalCost = entry.TotalCost
	}
	if entry.Kind == loop.LogIterComplete && entry.Duration > 0 {
		m.lastDuration = entry.Duration
	}
	if entry.Commit != "" {
		m.lastCommit = entry.Commit
	}

	// Add to visible log; track if user is scrolled up
	if m.scrollOffset > 0 {
		m.newBelow++
	}
	m.lines = append(m.lines, logLine{entry: entry})

	// Continue listening
	return m, waitForEvent(m.events)
}
