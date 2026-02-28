package panels

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

// SpecSelectedMsg is emitted when the user selects a spec.
type SpecSelectedMsg struct{ Spec spec.SpecFile }

// specItem wraps a spec.SpecFile as a list.Item.
type specItem struct {
	sf spec.SpecFile
}

func (s specItem) Title() string {
	return fmt.Sprintf("%s  %s", s.sf.Status.Symbol(), s.sf.Name)
}

func (s specItem) Description() string {
	return s.sf.Path
}

func (s specItem) FilterValue() string {
	return s.sf.Name
}

// specDelegate is a custom item delegate for compact single-line spec items.
type specDelegate struct{}

func (d specDelegate) Height() int                             { return 1 }
func (d specDelegate) Spacing() int                            { return 0 }
func (d specDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d specDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	si, ok := item.(specItem)
	if !ok {
		return
	}
	s := si.Title()
	if index == m.Index() {
		s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("> " + s)
	} else {
		s = "  " + s
	}
	fmt.Fprint(w, s)
}

// SpecsPanel displays a navigable list of spec files.
type SpecsPanel struct {
	list   list.Model
	specs  []spec.SpecFile
	width  int
	height int
}

// NewSpecsPanel creates a specs panel with the given spec files.
func NewSpecsPanel(specs []spec.SpecFile, w, h int) SpecsPanel {
	items := make([]list.Item, len(specs))
	for i, sf := range specs {
		items[i] = specItem{sf: sf}
	}
	delegate := specDelegate{}
	l := list.New(items, delegate, w, h)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return SpecsPanel{
		list:   l,
		specs:  specs,
		width:  w,
		height: h,
	}
}

// SelectedSpec returns the currently highlighted spec, or nil.
func (p SpecsPanel) SelectedSpec() *spec.SpecFile {
	if item, ok := p.list.SelectedItem().(specItem); ok {
		sf := item.sf
		return &sf
	}
	return nil
}

// SetSize resizes the panel.
func (p SpecsPanel) SetSize(w, h int) SpecsPanel {
	p.width = w
	p.height = h
	p.list.SetSize(w, h)
	return p
}

// Update handles key/mouse messages for the panel.
func (p SpecsPanel) Update(msg tea.Msg) (SpecsPanel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			p.list, cmd = p.list.Update(tea.KeyMsg{Type: tea.KeyDown})
			if sel := p.SelectedSpec(); sel != nil {
				sf := *sel
				return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
			}
		case "k", "up":
			p.list, cmd = p.list.Update(tea.KeyMsg{Type: tea.KeyUp})
			if sel := p.SelectedSpec(); sel != nil {
				sf := *sel
				return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
			}
		case "enter":
			if sel := p.SelectedSpec(); sel != nil {
				sf := *sel
				return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
			}
		default:
			p.list, cmd = p.list.Update(msg)
		}
	default:
		p.list, cmd = p.list.Update(msg)
	}
	return p, cmd
}

// View renders the specs panel.
func (p SpecsPanel) View() string {
	if len(p.specs) == 0 {
		return lipgloss.NewStyle().
			Width(p.width).Height(p.height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render("No specs")
	}
	return p.list.View()
}
