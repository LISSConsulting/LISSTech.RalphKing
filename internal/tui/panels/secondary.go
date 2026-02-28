package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/components"
)

// SecondaryTab identifies the active content tab in the secondary panel.
type SecondaryTab int

const (
	TabRegent SecondaryTab = iota // Regent supervisor messages
	TabGit                        // Git operation log
	TabTests                      // Test output (MVP: placeholder)
	TabCost                       // Cost breakdown (MVP: placeholder)
)

var secondaryTabLabels = []string{"Regent", "Git", "Tests", "Cost"}

// SecondaryPanel is the secondary (right-bottom) panel with Regent/git/test/cost tabs.
type SecondaryPanel struct {
	tabbar    components.TabBar
	regent    components.LogView // Regent supervisor messages
	gitLog    components.LogView // Git operation messages
	width     int
	height    int
	activeTab SecondaryTab
}

// NewSecondaryPanel creates a secondary panel.
func NewSecondaryPanel(w, h int) SecondaryPanel {
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}
	return SecondaryPanel{
		tabbar:    components.NewTabBar(secondaryTabLabels).SetWidth(w),
		regent:    components.NewLogView(w, contentH),
		gitLog:    components.NewLogView(w, contentH),
		width:     w,
		height:    h,
		activeTab: TabRegent,
	}
}

// AppendLine appends a pre-rendered line routed to the appropriate tab.
// routeTab specifies which tab this line belongs to (use TabRegent, TabGit, etc.).
func (p SecondaryPanel) AppendLine(rendered string, routeTab SecondaryTab) SecondaryPanel {
	switch routeTab {
	case TabRegent:
		p.regent = p.regent.AppendLine(rendered)
	case TabGit:
		p.gitLog = p.gitLog.AppendLine(rendered)
	}
	return p
}

// SetSize resizes all internal viewports.
func (p SecondaryPanel) SetSize(w, h int) SecondaryPanel {
	p.width = w
	p.height = h
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}
	p.tabbar = p.tabbar.SetWidth(w)
	p.regent = p.regent.SetSize(w, contentH)
	p.gitLog = p.gitLog.SetSize(w, contentH)
	return p
}

// Update handles key messages for the secondary panel.
func (p SecondaryPanel) Update(msg tea.Msg) (SecondaryPanel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "]":
			p.tabbar = p.tabbar.Next()
			p.activeTab = SecondaryTab(p.tabbar.Active())
		case "[":
			p.tabbar = p.tabbar.Prev()
			p.activeTab = SecondaryTab(p.tabbar.Active())
		default:
			// Delegate scroll keys to active tab's logview.
			switch p.activeTab {
			case TabRegent:
				p.regent, cmd = p.regent.Update(msg)
			case TabGit:
				p.gitLog, cmd = p.gitLog.Update(msg)
			}
		}
	default:
		switch p.activeTab {
		case TabRegent:
			p.regent, cmd = p.regent.Update(msg)
		case TabGit:
			p.gitLog, cmd = p.gitLog.Update(msg)
		}
	}
	return p, cmd
}

// View renders the secondary panel: tab bar + active tab content.
func (p SecondaryPanel) View() string {
	tabRow := p.tabbar.View()
	var content string
	switch p.activeTab {
	case TabRegent:
		content = p.regent.View()
	case TabGit:
		content = p.gitLog.View()
	default:
		content = lipgloss.NewStyle().
			Width(p.width).Height(p.height-1).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render("Coming soon")
	}
	return lipgloss.JoinVertical(lipgloss.Left, tabRow, content)
}
