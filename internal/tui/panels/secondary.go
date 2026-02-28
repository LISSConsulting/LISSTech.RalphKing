package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/components"
)

// SecondaryTab identifies the active content tab in the secondary panel.
type SecondaryTab int

const (
	TabRegent SecondaryTab = iota // Regent supervisor messages
	TabGit                        // Git operation log
	TabTests                      // Test output (from Regent entries)
	TabCost                       // Cost breakdown
)

var secondaryTabLabels = []string{"Regent", "Git", "Tests", "Cost"}

// SecondaryPanel is the secondary (right-bottom) panel with Regent/git/test/cost tabs.
type SecondaryPanel struct {
	tabbar    components.TabBar
	regent    components.LogView         // Regent supervisor messages
	gitLog    components.LogView         // Git operation messages
	tests     components.LogView         // Test output from Regent entries
	costData  []store.IterationSummary   // Per-iteration cost accumulator
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
		tests:     components.NewLogView(w, contentH),
		width:     w,
		height:    h,
		activeTab: TabRegent,
	}
}

// AppendLine appends a pre-rendered line routed to the appropriate tab.
// routeTab specifies which tab this line belongs to (use TabRegent, TabGit, TabTests, etc.).
func (p SecondaryPanel) AppendLine(rendered string, routeTab SecondaryTab) SecondaryPanel {
	switch routeTab {
	case TabRegent:
		p.regent = p.regent.AppendLine(rendered)
	case TabGit:
		p.gitLog = p.gitLog.AppendLine(rendered)
	case TabTests:
		p.tests = p.tests.AppendLine(rendered)
	}
	return p
}

// AddIteration records a completed iteration's summary for the cost tab.
func (p SecondaryPanel) AddIteration(s store.IterationSummary) SecondaryPanel {
	p.costData = append(p.costData, s)
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
	p.tests = p.tests.SetSize(w, contentH)
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
			case TabTests:
				p.tests, cmd = p.tests.Update(msg)
			}
		}
	default:
		switch p.activeTab {
		case TabRegent:
			p.regent, cmd = p.regent.Update(msg)
		case TabGit:
			p.gitLog, cmd = p.gitLog.Update(msg)
		case TabTests:
			p.tests, cmd = p.tests.Update(msg)
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
	case TabTests:
		content = p.tests.View()
	case TabCost:
		content = p.renderCostTable()
	}
	return lipgloss.JoinVertical(lipgloss.Left, tabRow, content)
}

// renderCostTable renders the per-iteration cost table for the Cost tab.
func (p SecondaryPanel) renderCostTable() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	contentH := p.height - 1
	if contentH < 1 {
		contentH = 1
	}
	if len(p.costData) == 0 {
		return lipgloss.NewStyle().
			Width(p.width).Height(contentH).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render("No iterations yet")
	}

	var sb strings.Builder
	header := fmt.Sprintf("  %-4s %-8s %8s %10s", "#", "Mode", "Cost", "Duration")
	divider := strings.Repeat("─", min(p.width, 38))
	sb.WriteString(dim.Render(header))
	sb.WriteString("\n")
	sb.WriteString(dim.Render(divider))
	sb.WriteString("\n")

	var totalCost, totalDur float64
	for _, s := range p.costData {
		cost := fmt.Sprintf("$%.3f", s.CostUSD)
		dur := fmt.Sprintf("%.1fs", s.Duration)
		line := fmt.Sprintf("  %-4d %-8s %8s %10s", s.Number, s.Mode, cost, dur)
		sb.WriteString(line)
		sb.WriteString("\n")
		totalCost += s.CostUSD
		totalDur += s.Duration
	}

	sb.WriteString(dim.Render(divider))
	sb.WriteString("\n")
	totalLine := fmt.Sprintf("  %-13s %8s %10s", "Total", fmt.Sprintf("$%.3f", totalCost), fmt.Sprintf("%.1fs", totalDur))
	sb.WriteString(dim.Render(totalLine))

	return lipgloss.NewStyle().
		Width(p.width).Height(contentH).
		Render(sb.String())
}
