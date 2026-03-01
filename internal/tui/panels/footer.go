package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

// FooterProps holds all data needed to render the footer bar.
type FooterProps struct {
	Focus         string // "specs", "iterations", "main", "secondary"
	LastCommit    string
	StopRequested bool
	StateLabel    string // "BUILDING", "IDLE", etc.
	ScrollOffset  int
	NewBelow      int
}

// RenderFooter renders the context-sensitive footer bar.
// Left side: last commit. Right side: keybinding hints for current focus + global.
func RenderFooter(props FooterProps, width int) string {
	commit := props.LastCommit
	if commit == "" {
		commit = "—"
	}
	left := fmt.Sprintf("last commit: %s", commit)

	var right string
	switch {
	case props.StopRequested:
		right = "⏹ stopping after iteration…  q to force quit"
	default:
		panelHints := panelHints(props.Focus)
		scrollHints := ""
		if props.ScrollOffset > 0 && props.NewBelow > 0 {
			scrollHints = fmt.Sprintf("  ↓%d new  ↑%d", props.NewBelow, props.ScrollOffset)
		} else if props.ScrollOffset > 0 {
			scrollHints = fmt.Sprintf("  ↑%d", props.ScrollOffset)
		}
		right = panelHints + scrollHints + "  ?:help  q:quit  1-4:panel  s:stop"
	}

	gap := width - len(left) - len(right)
	if gap < 2 {
		gap = 2
	}

	return footerStyle.Width(width).Render(left + strings.Repeat(" ", gap) + right)
}

// panelHints returns the context-sensitive keybinding hints for a given focus.
func panelHints(focus string) string {
	switch focus {
	case "specs":
		return "j/k:navigate  e:edit  n:new  enter:view  tab:next panel"
	case "iterations":
		return "j/k:navigate  enter:view  tab:next panel"
	case "main":
		return "f:follow  [/]:tab  ctrl+u/d:scroll  tab:next panel"
	case "secondary":
		return "[/]:tab  j/k:scroll  tab:next panel"
	default:
		return "tab:next panel"
	}
}
