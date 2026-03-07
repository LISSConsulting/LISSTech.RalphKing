package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/components"
)

// MainTab identifies the active content tab in the main view.
type MainTab int

const (
	TabOutput           MainTab = iota // Live loop output log
	TabSpecContent                     // Spec file content viewer (US2)
	TabIterationDetail                 // Past iteration log drill-down (US3)
	TabIterationSummary                // Iteration metadata summary (US3)
)

// MainView is the main (right-top) panel showing loop output and spec/iteration content.
// Each tab owns its own LogView so content is never displaced by output from another tab.
type MainView struct {
	tabbar       components.TabBar
	outputLog    components.LogView // Tab 0: live streaming output
	specLog      components.LogView // Tab 1: spec file content
	iterationLog components.LogView // Tab 2: past iteration log
	summaryLog   components.LogView // Tab 3: iteration metadata summary
	width        int
	height       int
	activeTab    MainTab
}

var mainTabLabels = []string{"Output", "Spec", "Iteration", "Summary"}

// NewMainView creates a MainView with the output tab active.
func NewMainView(w, h int) MainView {
	contentH := h - 1 // subtract tab bar row
	if contentH < 1 {
		contentH = 1
	}
	return MainView{
		tabbar:       components.NewTabBar(mainTabLabels).SetWidth(w),
		outputLog:    components.NewLogView(w, contentH),
		specLog:      components.NewLogView(w, contentH),
		iterationLog: components.NewLogView(w, contentH),
		summaryLog:   components.NewLogView(w, contentH),
		width:        w,
		height:       h,
	}
}

// AppendLine appends a pre-rendered (styled) line to the output log only.
// It never touches specLog, iterationLog, or summaryLog so those tabs
// retain their content independently of live streaming output.
func (v MainView) AppendLine(rendered string) MainView {
	v.outputLog = v.outputLog.AppendLine(rendered)
	return v
}

// ShowSpec loads spec content into the spec viewer and switches to TabSpecContent.
func (v MainView) ShowSpec(content string) MainView {
	v.specLog = v.specLog.SetContent(splitLines(content))
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
	v.iterationLog = v.iterationLog.SetContent(rendered)
	v.activeTab = TabIterationDetail
	v.tabbar = components.NewTabBar(mainTabLabels).SetWidth(v.width)
	for i := 0; i < int(TabIterationDetail); i++ {
		v.tabbar = v.tabbar.Next()
	}
	return v
}

// SetIterationSummary loads summary key-value lines into the summary viewport.
// The tab is not switched; the user navigates to Summary with ].
func (v MainView) SetIterationSummary(lines []string) MainView {
	v.summaryLog = v.summaryLog.SetContent(lines)
	return v
}

// SwitchToOutput returns to the live output tab.
func (v MainView) SwitchToOutput() MainView {
	v.activeTab = TabOutput
	v.tabbar = components.NewTabBar(mainTabLabels).SetWidth(v.width)
	return v
}

// ShowWorktreeLog loads a worktree agent's accumulated log entries into the
// Output tab so the user can review a specific agent's activity.
// lines are pre-rendered strings (app.go renders via theme.RenderLogLine).
func (v MainView) ShowWorktreeLog(lines []string) MainView {
	v.outputLog = v.outputLog.SetContent(lines)
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
	v.outputLog = v.outputLog.SetSize(w, contentH)
	v.specLog = v.specLog.SetSize(w, contentH)
	v.iterationLog = v.iterationLog.SetSize(w, contentH)
	v.summaryLog = v.summaryLog.SetSize(w, contentH)
	return v
}

// activeLogView returns a pointer to the LogView for the current tab so that
// Update and View can dispatch to the right buffer without a switch per call-site.
func (v *MainView) activeLogView() *components.LogView {
	switch v.activeTab {
	case TabSpecContent:
		return &v.specLog
	case TabIterationDetail:
		return &v.iterationLog
	case TabIterationSummary:
		return &v.summaryLog
	default:
		return &v.outputLog
	}
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
			if v.activeTab != TabIterationSummary {
				lv := v.activeLogView()
				*lv = lv.ToggleFollow()
			}
		default:
			lv := v.activeLogView()
			*lv, cmd = lv.Update(msg)
		}
	default:
		lv := v.activeLogView()
		*lv, cmd = lv.Update(msg)
	}
	return v, cmd
}

// View renders the main panel: tab bar + content area.
func (v MainView) View() string {
	tabRow := v.tabbar.View()
	lv := v.activeLogView()
	return lipgloss.JoinVertical(lipgloss.Left, tabRow, lv.View())
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
