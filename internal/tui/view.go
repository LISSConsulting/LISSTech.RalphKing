package tui

import (
	"fmt"
	"strings"
	"time"

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
	if m.lastDuration > 0 {
		parts = append(parts, fmt.Sprintf("last: %.1fs", m.lastDuration))
	}
	if !m.startedAt.IsZero() {
		elapsed := m.now.Sub(m.startedAt)
		parts = append(parts, fmt.Sprintf("elapsed: %s", formatElapsed(elapsed)))
	}
	if !m.now.IsZero() {
		parts = append(parts, m.now.Format("15:04"))
	}

	content := strings.Join(parts, "  │  ")
	return m.accentHeaderStyle.Width(m.width).Render(content)
}

func (m Model) renderFooter() string {
	commit := m.lastCommit
	if commit == "" {
		commit = "—"
	}

	left := fmt.Sprintf("[⬆ pull] [⬇ push]  last commit: %s", commit)
	right := "q to quit"
	if m.scrollOffset > 0 && m.newBelow > 0 {
		right = fmt.Sprintf("↓%d new  ↑%d  j/k scroll  q to quit", m.newBelow, m.scrollOffset)
	} else if m.scrollOffset > 0 {
		right = fmt.Sprintf("↑%d  j/k scroll  q to quit", m.scrollOffset)
	}

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

	// Calculate visible window based on scroll offset.
	// scrollOffset 0 = bottom (latest lines), >0 = scrolled up.
	end := len(m.lines) - m.scrollOffset
	if end < 0 {
		end = 0
	}
	start := end - height
	if start < 0 {
		start = 0
	}
	visible := m.lines[start:end]

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

// formatElapsed renders a duration as a compact human-readable string.
// Examples: "5s", "2m30s", "1h15m"
func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func (m Model) renderLine(line logLine) string {
	e := line.entry
	ts := timestampStyle.Render(fmt.Sprintf("[%s]", e.Timestamp.Format("15:04:05")))

	switch e.Kind {
	case loop.LogToolUse:
		icon := toolIcon(e.ToolName)
		style := toolStyle(e.ToolName)
		displayName := e.ToolName
		if len(displayName) > 14 {
			displayName = displayName[:13] + "…"
		}
		name := style.Render(fmt.Sprintf("%-14s", displayName))
		input := e.ToolInput
		if len(input) > 60 {
			input = input[:59] + "…"
		}
		return fmt.Sprintf("%s  %s %s %s", ts, icon, name, input)

	case loop.LogIterStart:
		return fmt.Sprintf("%s  ── iteration %d ──", ts, e.Iteration)

	case loop.LogIterComplete:
		iterMsg := fmt.Sprintf("✅ iteration %d complete  —  $%.2f  —  %.1fs", e.Iteration, e.CostUSD, e.Duration)
		if e.Subtype != "" {
			iterMsg += fmt.Sprintf("  —  %s", e.Subtype)
		}
		return fmt.Sprintf("%s  %s", ts, resultStyle.Render(iterMsg))

	case loop.LogError:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("❌ "+e.Message))

	case loop.LogGitPull:
		return fmt.Sprintf("%s  %s", ts, m.accentGitStyle.Render("⬆ "+e.Message))

	case loop.LogGitPush:
		return fmt.Sprintf("%s  %s", ts, m.accentGitStyle.Render("⬇ "+e.Message))

	case loop.LogDone:
		return fmt.Sprintf("%s  %s", ts, resultStyle.Render("✅ "+e.Message))

	case loop.LogStopped:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("⏹ "+e.Message))

	case loop.LogRegent:
		return fmt.Sprintf("%s  %s", ts, regentStyle.Render("🛡️  Regent: "+e.Message))

	default:
		return fmt.Sprintf("%s  %s", ts, infoStyle.Render(e.Message))
	}
}
