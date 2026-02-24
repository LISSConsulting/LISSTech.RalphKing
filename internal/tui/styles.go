// Package tui provides a bubbletea + lipgloss terminal UI for the Ralph loop.
package tui

import "github.com/charmbracelet/lipgloss"

// defaultAccentColor is the default accent color (indigo).
const defaultAccentColor = "#7D56F4"

// Color palette matching the spec.
var (
	colorWhite  = lipgloss.Color("#FAFAFA")
	colorGray   = lipgloss.Color("#888888")
	colorBlue   = lipgloss.Color("#5B9BD5")
	colorGreen  = lipgloss.Color("#6BCB77")
	colorYellow = lipgloss.Color("#FFD93D")
	colorRed    = lipgloss.Color("#FF6B6B")
	colorOrange = lipgloss.Color("#FFA54F")
)

// Styles used across the TUI. Accent-dependent styles (header, git) live
// on the Model and are computed from the configured accent color at creation.
var (
	footerStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	timestampStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	readStyle = lipgloss.NewStyle().
			Foreground(colorBlue)

	writeStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	bashStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	resultStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	regentStyle = lipgloss.NewStyle().
			Foreground(colorOrange)

	infoStyle = lipgloss.NewStyle().
			Foreground(colorWhite)
)

// toolIcon returns the emoji icon for a given tool name.
func toolIcon(toolName string) string {
	switch toolName {
	case "Read", "read_file", "Glob", "Grep":
		return "📖"
	case "Write", "write_file", "Edit", "NotebookEdit":
		return "✏️ "
	case "Bash":
		return "🔧"
	case "WebFetch", "WebSearch":
		return "🌐"
	case "Task":
		return "🔀"
	default:
		return "⚡"
	}
}

// toolStyle returns the lipgloss style for a given tool name.
func toolStyle(toolName string) lipgloss.Style {
	switch toolName {
	case "Read", "read_file", "Glob", "Grep":
		return readStyle
	case "Write", "write_file", "Edit", "NotebookEdit":
		return writeStyle
	case "Bash":
		return bashStyle
	default:
		return infoStyle
	}
}
