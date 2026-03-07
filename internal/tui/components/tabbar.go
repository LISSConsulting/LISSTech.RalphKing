// Package components provides reusable TUI components for the Ralph multi-panel UI.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabActiveStyle renders the active tab with bold accent-colored text.
var tabActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

// tabInactiveStyle renders inactive tabs in a dimmed style.
var tabInactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

// TabBar is a stateless tab bar component that renders a row of labelled tabs.
// The active tab is highlighted with accent color and bold text.
type TabBar struct {
	tabs   []string
	active int
	width  int
}

// NewTabBar creates a TabBar with the given tab titles. The first tab is active.
func NewTabBar(tabs []string) TabBar {
	return TabBar{tabs: tabs}
}

// Active returns the index of the currently active tab.
func (t TabBar) Active() int {
	return t.active
}

// Next returns a TabBar with the next tab active (wraps around).
func (t TabBar) Next() TabBar {
	if len(t.tabs) == 0 {
		return t
	}
	t.active = (t.active + 1) % len(t.tabs)
	return t
}

// Prev returns a TabBar with the previous tab active (wraps around).
func (t TabBar) Prev() TabBar {
	if len(t.tabs) == 0 {
		return t
	}
	t.active = (t.active + len(t.tabs) - 1) % len(t.tabs)
	return t
}

// SetActive returns a TabBar with the given tab index active.
// If i is out of range it is clamped to valid bounds.
func (t TabBar) SetActive(i int) TabBar {
	if len(t.tabs) == 0 {
		return t
	}
	if i < 0 {
		i = 0
	}
	if i >= len(t.tabs) {
		i = len(t.tabs) - 1
	}
	t.active = i
	return t
}

// SetWidth returns a TabBar configured for the given render width.
func (t TabBar) SetWidth(w int) TabBar {
	t.width = w
	return t
}

// View renders the tab bar as a single line string.
// Active tab: bold + accent color. Inactive tabs: dimmed.
// Tabs are separated by " │ " and the result is padded/truncated to width.
func (t TabBar) View() string {
	if len(t.tabs) == 0 {
		return ""
	}

	var parts []string
	for i, label := range t.tabs {
		var rendered string
		if i == t.active {
			rendered = tabActiveStyle.Render(label)
		} else {
			rendered = tabInactiveStyle.Render(label)
		}
		parts = append(parts, rendered)
	}

	line := strings.Join(parts, "  │  ")
	if t.width > 0 {
		return lipgloss.NewStyle().MaxWidth(t.width).Render(line)
	}
	return line
}
