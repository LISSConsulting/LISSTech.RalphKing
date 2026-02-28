package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/components"
)

// MainTab identifies the active content tab in the main view.
type MainTab int

const (
	TabOutput        MainTab = iota // Live loop output log
	TabSpecContent                  // Spec file content viewer (US2)
	TabIterationDetail              // Past iteration drill-down (US3)
)

// MainView is the main (right-top) panel showing loop output and spec/iteration content.
type MainView struct {
	tabbar    components.TabBar
	logview   components.LogView
	width     int
	height    int
	activeTab MainTab
}

var mainTabLabels = []string{"Output", "Spec", "Iteration"}

// NewMainView creates a MainView with the output tab active.
func NewMainView(w, h int) MainView {
	contentH := h - 1 // subtract tab bar row
	if contentH < 1 {
		contentH = 1
	}
	return MainView{
		tabbar:  components.NewTabBar(mainTabLabels).SetWidth(w),
		logview: components.NewLogView(w, contentH),
		width:   w,
		height:  h,
	}
}

// AppendLine appends a pre-rendered (styled) line to the output log.
func (v MainView) AppendLine(rendered string) MainView {
	v.logview = v.logview.AppendLine(rendered)
	return v
}

// ShowSpec loads spec content into the spec viewer and switches to TabSpecContent.
// For the MVP this simply shows the content in the logview (placeholder).
func (v MainView) ShowSpec(content string) MainView {
	// MVP placeholder: show spec content as plain text in the logview
	v.logview = v.logview.SetContent(splitLines(content))
	v.activeTab = TabSpecContent
	v.tabbar = components.NewTabBar(mainTabLabels).SetWidth(v.width)
	for i := 0; i < int(TabSpecContent); i++ {
		v.tabbar = v.tabbar.Next()
	}
	return v
}

// ShowIterationLog loads a past iteration's log entries and switches to TabIterationDetail.
// entries are pre-rendered strings (app.go renders via theme.RenderLogLine before passing).
func (v MainView) ShowIterationLog(rendered []string) MainView {
	v.logview = v.logview.SetContent(rendered)
	v.activeTab = TabIterationDetail
	v.tabbar = components.NewTabBar(mainTabLabels).SetWidth(v.width)
	for i := 0; i < int(TabIterationDetail); i++ {
		v.tabbar = v.tabbar.Next()
	}
	return v
}

// SwitchToOutput returns to the live output tab.
func (v MainView) SwitchToOutput() MainView {
	v.activeTab = TabOutput
	v.tabbar = components.NewTabBar(mainTabLabels).SetWidth(v.width)
	return v
}

// SetSize resizes the main view.
func (v MainView) SetSize(w, h int) MainView {
	v.width = w
	v.height = h
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}
	v.tabbar = v.tabbar.SetWidth(w)
	v.logview = v.logview.SetSize(w, contentH)
	return v
}

// Update handles key messages for the main panel.
func (v MainView) Update(msg tea.Msg) (MainView, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "]":
			v.tabbar = v.tabbar.Next()
			v.activeTab = MainTab(v.tabbar.Active())
		case "[":
			v.tabbar = v.tabbar.Prev()
			v.activeTab = MainTab(v.tabbar.Active())
		case "f":
			v.logview = v.logview.ToggleFollow()
		default:
			v.logview, cmd = v.logview.Update(msg)
		}
	default:
		v.logview, cmd = v.logview.Update(msg)
	}
	return v, cmd
}

// View renders the main panel: tab bar + content area.
func (v MainView) View() string {
	tabRow := v.tabbar.View()
	content := v.logview.View()
	return lipgloss.JoinVertical(lipgloss.Left, tabRow, content)
}

// splitLines splits a string into lines for SetContent.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
