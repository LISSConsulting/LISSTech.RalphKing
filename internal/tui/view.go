package tui

import (
	"fmt"
	"strings"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// View renders the TUI: header bar, scrollable log, footer bar.
func (m Model) View() string {
	header := m.renderHeader()
	footer := m.renderFooter()

	// Log fills the space between header and footer
	logHeight := m.height - 2 // 1 header + 1 footer
	if logHeight < 1 {
		logHeight = 1
	}
	logView := m.renderLog(logHeight)

	return header + "\n" + logView + "\n" + footer
}

func (m Model) renderHeader() string {
	iter := fmt.Sprintf("%d", m.iteration)
	maxLabel := "∞"
	if m.maxIter > 0 {
		maxLabel = fmt.Sprintf("%d", m.maxIter)
	}

	branch := m.branch
	if branch == "" {
		branch = "—"
	}

	mode := m.mode
	if mode == "" {
		mode = "—"
	}

	parts := []string{
		"👑 RalphKing",
		fmt.Sprintf("mode: %s", mode),
		fmt.Sprintf("branch: %s", branch),
		fmt.Sprintf("iter: %s/%s", iter, maxLabel),
		fmt.Sprintf("cost: $%.2f", m.totalCost),
	}

	content := strings.Join(parts, "  │  ")
	return headerStyle.Width(m.width).Render(content)
}

func (m Model) renderFooter() string {
	commit := m.lastCommit
	if commit == "" {
		commit = "—"
	}

	left := fmt.Sprintf("[⬆ pull] [⬇ push]  last commit: %s", commit)
	right := "q to quit"

	gap := m.width - len(left) - len(right)
	if gap < 2 {
		gap = 2
	}

	return footerStyle.Width(m.width).Render(
		left + strings.Repeat(" ", gap) + right,
	)
}

func (m Model) renderLog(height int) string {
	if len(m.lines) == 0 {
		return strings.Repeat("\n", height-1)
	}

	// Show the last `height` lines (auto-scroll to bottom)
	start := 0
	if len(m.lines) > height {
		start = len(m.lines) - height
	}
	visible := m.lines[start:]

	var b strings.Builder
	for i, line := range visible {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(m.renderLine(line))
	}

	// Pad remaining lines if fewer than height
	remaining := height - len(visible)
	for i := 0; i < remaining; i++ {
		b.WriteByte('\n')
	}

	return b.String()
}

func (m Model) renderLine(line logLine) string {
	e := line.entry
	ts := timestampStyle.Render(fmt.Sprintf("[%s]", e.Timestamp.Format("15:04:05")))

	switch e.Kind {
	case loop.LogToolUse:
		icon := toolIcon(e.ToolName)
		style := toolStyle(e.ToolName)
		name := style.Render(fmt.Sprintf("%-14s", e.ToolName))
		return fmt.Sprintf("%s  %s %s %s", ts, icon, name, e.ToolInput)

	case loop.LogIterStart:
		return fmt.Sprintf("%s  ── iteration %d ──", ts, e.Iteration)

	case loop.LogIterComplete:
		return fmt.Sprintf("%s  %s", ts,
			resultStyle.Render(fmt.Sprintf("✅ iteration %d complete  —  $%.2f  —  %.1fs",
				e.Iteration, e.CostUSD, e.Duration)))

	case loop.LogError:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("❌ "+e.Message))

	case loop.LogGitPull:
		return fmt.Sprintf("%s  %s", ts, gitStyle.Render("⬆ "+e.Message))

	case loop.LogGitPush:
		return fmt.Sprintf("%s  %s", ts, gitStyle.Render("⬇ "+e.Message))

	case loop.LogDone:
		return fmt.Sprintf("%s  %s", ts, resultStyle.Render("✅ "+e.Message))

	case loop.LogStopped:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("⏹ "+e.Message))

	default:
		return fmt.Sprintf("%s  %s", ts, infoStyle.Render(e.Message))
	}
}
