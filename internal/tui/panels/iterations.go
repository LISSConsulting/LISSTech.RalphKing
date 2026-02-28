package panels

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
)

// IterationSelectedMsg is emitted when the user selects an iteration.
// Defined here (not in parent tui package) to avoid circular imports.
type IterationSelectedMsg struct{ Number int }

// iterItem implements list.Item for an iteration summary.
type iterItem struct {
	summary store.IterationSummary
	running bool // true if this is the currently-running iteration
}

func (i iterItem) Title() string {
	status := "✓"
	if i.running {
		status = "●"
	} else if i.summary.Subtype == "error_max_turns" || i.summary.Subtype == "error" {
		status = "✗"
	}
	return fmt.Sprintf("#%d %s %s", i.summary.Number, i.summary.Mode, status)
}

func (i iterItem) Description() string {
	if i.running {
		return "running…"
	}
	return fmt.Sprintf("$%.3f  %.1fs", i.summary.CostUSD, i.summary.Duration)
}

func (i iterItem) FilterValue() string {
	return fmt.Sprintf("#%d", i.summary.Number)
}

// IterationsPanel displays a list of completed and in-progress iterations.
type IterationsPanel struct {
	list       list.Model
	iterations []store.IterationSummary
	currentNum *int // currently running iteration number (nil if idle)
	width      int
	height     int
}

// iterDelegate is a custom list item delegate with compact rendering.
type iterDelegate struct{}

func (d iterDelegate) Height() int                             { return 1 }
func (d iterDelegate) Spacing() int                            { return 0 }
func (d iterDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d iterDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(iterItem)
	if !ok {
		return
	}
	s := fmt.Sprintf("%s  %s", item.Title(), item.Description())
	if index == m.Index() {
		s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("> " + s)
	} else {
		s = "  " + s
	}
	fmt.Fprint(w, s)
}

// NewIterationsPanel creates an empty iterations panel.
func NewIterationsPanel(w, h int) IterationsPanel {
	delegate := iterDelegate{}
	l := list.New(nil, delegate, w, h)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return IterationsPanel{
		list:   l,
		width:  w,
		height: h,
	}
}

// AddIteration adds a completed iteration summary to the panel.
func (p IterationsPanel) AddIteration(s store.IterationSummary) IterationsPanel {
	p.iterations = append(p.iterations, s)
	p.list.SetItems(p.buildItems())
	return p
}

// SetCurrent marks the given iteration number as currently running.
// Pass 0 to clear the running indicator.
func (p IterationsPanel) SetCurrent(n int) IterationsPanel {
	if n == 0 {
		p.currentNum = nil
	} else {
		p.currentNum = &n
	}
	p.list.SetItems(p.buildItems())
	return p
}

// buildItems rebuilds the list.Item slice from stored summaries.
func (p IterationsPanel) buildItems() []list.Item {
	items := make([]list.Item, len(p.iterations))
	for i, s := range p.iterations {
		running := p.currentNum != nil && *p.currentNum == s.Number
		items[i] = iterItem{summary: s, running: running}
	}
	return items
}

// SelectedIteration returns the currently selected iteration summary, or nil.
func (p IterationsPanel) SelectedIteration() *store.IterationSummary {
	if item, ok := p.list.SelectedItem().(iterItem); ok {
		s := item.summary
		return &s
	}
	return nil
}

// SetSize resizes the panel.
func (p IterationsPanel) SetSize(w, h int) IterationsPanel {
	p.width = w
	p.height = h
	p.list.SetSize(w, h)
	return p
}

// Update handles key/mouse messages for the panel.
func (p IterationsPanel) Update(msg tea.Msg) (IterationsPanel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			p.list, cmd = p.list.Update(tea.KeyMsg{Type: tea.KeyDown})
		case "k", "up":
			p.list, cmd = p.list.Update(tea.KeyMsg{Type: tea.KeyUp})
		case "enter":
			if sel := p.SelectedIteration(); sel != nil {
				return p, func() tea.Msg { return IterationSelectedMsg{Number: sel.Number} }
			}
		default:
			p.list, cmd = p.list.Update(msg)
		}
	default:
		p.list, cmd = p.list.Update(msg)
	}
	return p, cmd
}

// View renders the iterations panel with a border.
func (p IterationsPanel) View() string {
	if len(p.iterations) == 0 && p.currentNum == nil {
		return lipgloss.NewStyle().
			Width(p.width).Height(p.height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render("No iterations yet")
	}
	return p.list.View()
}
