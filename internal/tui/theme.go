package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// Theme holds accent-color-derived styles for the multi-panel TUI.
// Non-accent styles (toolIcon, toolStyle, color vars) are package-level
// and shared with the existing single-panel TUI via styles.go.
type Theme struct {
	accentStyle     lipgloss.Style // for header background / focused elements
	gitStyle        lipgloss.Style // for git operation messages
	borderFocused   lipgloss.Style // focused panel border
	borderUnfocused lipgloss.Style // unfocused panel border
}

// NewTheme creates a Theme from a hex accent color string (e.g. "#7D56F4").
// If accentColor is empty, the default accent color is used.
func NewTheme(accentColor string) Theme {
	color := defaultAccentColor
	if accentColor != "" {
		color = accentColor
	}
	c := lipgloss.Color(color)
	return Theme{
		accentStyle: lipgloss.NewStyle().
			Background(c).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true),
		gitStyle: lipgloss.NewStyle().
			Foreground(c),
		borderFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c),
		borderUnfocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray),
	}
}

// AccentHeaderStyle returns the style for the header bar.
func (t Theme) AccentHeaderStyle() lipgloss.Style {
	return t.accentStyle
}

// AccentBorderStyle returns a focused-panel border style using the accent color.
func (t Theme) AccentBorderStyle() lipgloss.Style {
	return t.borderFocused
}

// DimBorderStyle returns an unfocused-panel border style using gray.
func (t Theme) DimBorderStyle() lipgloss.Style {
	return t.borderUnfocused
}

// PanelBorderStyle returns the appropriate border style for a panel based on
// whether it currently holds keyboard focus.
func (t Theme) PanelBorderStyle(focused bool) lipgloss.Style {
	if focused {
		return t.borderFocused
	}
	return t.borderUnfocused
}

// RenderLogLine renders a loop.LogEntry as a single terminal line.
// Carries forward all LogKind rendering from view.go:renderLine(), but
// accepts width and theme as explicit parameters instead of model fields.
func (t Theme) RenderLogLine(entry loop.LogEntry, width int) string {
	ts := timestampStyle.Render(fmt.Sprintf("[%s]", entry.Timestamp.Format("15:04:05")))

	switch entry.Kind {
	case loop.LogToolUse:
		icon := toolIcon(entry.ToolName)
		style := toolStyle(entry.ToolName)
		displayName := entry.ToolName
		if len(displayName) > 14 {
			displayName = displayName[:13] + "…"
		}
		name := style.Render(fmt.Sprintf("%-14s", displayName))
		input := singleLine(entry.ToolInput)
		maxInput := width - 32
		if maxInput < 20 {
			maxInput = 20
		}
		if len(input) > maxInput {
			input = input[:maxInput-1] + "…"
		}
		return fmt.Sprintf("%s  %s %s %s", ts, icon, name, input)

	case loop.LogText:
		text := singleLine(entry.Message)
		maxText := width - 17
		if maxText < 20 {
			maxText = 20
		}
		if len([]rune(text)) > maxText {
			runes := []rune(text)
			text = string(runes[:maxText-1]) + "…"
		}
		return fmt.Sprintf("%s  %s", ts, reasoningStyle.Render("💭 "+text))

	case loop.LogIterStart:
		return fmt.Sprintf("%s  ── iteration %d ──", ts, entry.Iteration)

	case loop.LogIterComplete:
		iterMsg := fmt.Sprintf("✅ iteration %d complete  —  $%.2f  —  %.1fs", entry.Iteration, entry.CostUSD, entry.Duration)
		if entry.Subtype != "" {
			iterMsg += fmt.Sprintf("  —  %s", singleLine(entry.Subtype))
		}
		return fmt.Sprintf("%s  %s", ts, resultStyle.Render(iterMsg))

	case loop.LogError:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("❌ "+singleLine(entry.Message)))

	case loop.LogGitPull:
		return fmt.Sprintf("%s  %s", ts, t.gitStyle.Render("⬆ "+singleLine(entry.Message)))

	case loop.LogGitPush:
		return fmt.Sprintf("%s  %s", ts, t.gitStyle.Render("⬇ "+singleLine(entry.Message)))

	case loop.LogDone:
		return fmt.Sprintf("%s  %s", ts, resultStyle.Render("✅ "+singleLine(entry.Message)))

	case loop.LogStopped:
		return fmt.Sprintf("%s  %s", ts, errorStyle.Render("⏹ "+singleLine(entry.Message)))

	case loop.LogRegent:
		return fmt.Sprintf("%s  %s", ts, regentStyle.Render("🛡️  Regent: "+singleLine(entry.Message)))

	default:
		return fmt.Sprintf("%s  %s", ts, infoStyle.Render(singleLine(entry.Message)))
	}
}

// RenderLogLine is also exported as a package-level function for convenience.
// It delegates to theme.RenderLogLine.
func RenderLogLine(entry loop.LogEntry, width int, theme Theme) string {
	return theme.RenderLogLine(entry, width)
}
