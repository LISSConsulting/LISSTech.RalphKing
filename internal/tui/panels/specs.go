package panels

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

// SpecSelectedMsg is emitted when the user selects a spec or child file.
// Spec.Path holds the specific file path (spec.md, plan.md, tasks.md, or the
// flat-file path for non-directory specs).
type SpecSelectedMsg struct{ Spec spec.SpecFile }

// EditSpecRequestMsg is emitted when the user presses 'e' to open the selected spec in $EDITOR.
type EditSpecRequestMsg struct{ Path string }

// CreateSpecRequestMsg is emitted when the user submits a new spec name via the 'n' overlay.
type CreateSpecRequestMsg struct{ Name string }

// specTreeNode holds a directory-level spec with its discovered child files.
type specTreeNode struct {
	sf       spec.SpecFile // the directory or single-file spec
	children []string      // child file paths that exist (relative to workDir)
	expanded bool          // whether children are currently visible
}

// specRow is one visible row in the flattened tree.
type specRow struct {
	nodeIdx  int  // which node this row belongs to
	isChild  bool // true for child file rows
	childIdx int  // index in node.children (when isChild)
}

// childFileIcon returns a display icon for a known spec child file name.
func childFileIcon(name string) string {
	switch name {
	case "spec.md":
		return "📋"
	case "plan.md":
		return "📐"
	case "tasks.md":
		return "✅"
	default:
		return "📄"
	}
}

// buildTree discovers child files for each directory spec node and returns the tree.
// workDir is used to stat child files; empty string disables file system checks.
func buildTree(specs []spec.SpecFile, workDir string) []specTreeNode {
	nodes := make([]specTreeNode, len(specs))
	for i, sf := range specs {
		if !sf.IsDir {
			nodes[i] = specTreeNode{sf: sf}
			continue
		}
		var children []string
		for _, name := range []string{"spec.md", "plan.md", "tasks.md"} {
			p := filepath.Join(workDir, sf.Dir, name)
			if _, err := os.Stat(p); err == nil {
				children = append(children, filepath.Join(sf.Dir, name))
			}
		}
		nodes[i] = specTreeNode{sf: sf, children: children}
	}
	return nodes
}

// flattenTree returns the visible rows given the current expansion state.
func flattenTree(nodes []specTreeNode) []specRow {
	var rows []specRow
	for i, n := range nodes {
		rows = append(rows, specRow{nodeIdx: i})
		if n.expanded {
			for j := range n.children {
				rows = append(rows, specRow{nodeIdx: i, isChild: true, childIdx: j})
			}
		}
	}
	return rows
}

// SpecsPanel displays a navigable tree of spec files.
// Directory specs can be expanded to reveal spec.md, plan.md, and tasks.md children.
type SpecsPanel struct {
	nodes     []specTreeNode
	flat      []specRow // current flattened view (rebuilt on expand/collapse)
	cursor    int       // cursor position in flat
	scrollTop int       // first visible row index
	workDir   string
	specs     []spec.SpecFile // original spec list preserved for callers
	width     int
	height    int

	input       textinput.Model
	inputActive bool
}

// NewSpecsPanel creates a specs panel. workDir is used to discover child files
// for directory specs; pass "" to skip file system checks (e.g. in tests).
func NewSpecsPanel(specs []spec.SpecFile, workDir string, w, h int) SpecsPanel {
	nodes := buildTree(specs, workDir)
	flat := flattenTree(nodes)

	ti := textinput.New()
	ti.Placeholder = "spec-name"
	ti.CharLimit = 64
	if w > 4 {
		ti.Width = w - 4
	}

	return SpecsPanel{
		nodes:   nodes,
		flat:    flat,
		specs:   specs,
		workDir: workDir,
		width:   w,
		height:  h,
		input:   ti,
	}
}

// SelectedSpec returns the directory-level spec for the current cursor position.
// For child-file rows the parent directory spec is returned so callers can use
// it for operations like launching worktree agents.  Returns nil when empty.
func (p SpecsPanel) SelectedSpec() *spec.SpecFile {
	if len(p.flat) == 0 {
		return nil
	}
	sf := p.nodes[p.flat[p.cursor].nodeIdx].sf
	return &sf
}

// SetSize resizes the panel.
func (p SpecsPanel) SetSize(w, h int) SpecsPanel {
	p.width = w
	p.height = h
	if w > 4 {
		p.input.Width = w - 4
	}
	return p
}

// moveCursor adjusts cursor by delta and updates scrollTop to keep it visible.
func (p SpecsPanel) moveCursor(delta int) SpecsPanel {
	n := len(p.flat)
	if n == 0 {
		return p
	}
	p.cursor += delta
	if p.cursor < 0 {
		p.cursor = 0
	}
	if p.cursor >= n {
		p.cursor = n - 1
	}
	if p.cursor < p.scrollTop {
		p.scrollTop = p.cursor
	}
	if p.cursor >= p.scrollTop+p.height {
		p.scrollTop = p.cursor - p.height + 1
	}
	if p.scrollTop < 0 {
		p.scrollTop = 0
	}
	return p
}

// Update handles key messages for the panel.
func (p SpecsPanel) Update(msg tea.Msg) (SpecsPanel, tea.Cmd) {
	// When the name-input overlay is active, route messages there first.
	if p.inputActive {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				p.inputActive = false
				p.input.Blur()
				p.input.Reset()
				return p, nil
			case "enter":
				name := strings.TrimSpace(p.input.Value())
				if name == "" {
					return p, nil
				}
				p.inputActive = false
				p.input.Blur()
				p.input.Reset()
				return p, func() tea.Msg { return CreateSpecRequestMsg{Name: name} }
			}
		}
		var cmd tea.Cmd
		p.input, cmd = p.input.Update(msg)
		return p, cmd
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return p, nil
	}

	switch keyMsg.String() {
	case "j", "down":
		p = p.moveCursor(1)
		if sel := p.SelectedSpec(); sel != nil {
			sf := *sel
			return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
		}

	case "k", "up":
		p = p.moveCursor(-1)
		if sel := p.SelectedSpec(); sel != nil {
			sf := *sel
			return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
		}

	case "enter":
		if len(p.flat) == 0 {
			return p, nil
		}
		row := p.flat[p.cursor]
		if row.isChild {
			// Emit SpecSelectedMsg with the child file path.
			childPath := p.nodes[row.nodeIdx].children[row.childIdx]
			sf := p.nodes[row.nodeIdx].sf
			sf.Path = childPath
			return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
		}
		// Directory row: toggle expand if children exist.
		node := p.nodes[row.nodeIdx]
		if len(node.children) > 0 {
			p.nodes[row.nodeIdx].expanded = !p.nodes[row.nodeIdx].expanded
			p.flat = flattenTree(p.nodes)
			if p.cursor >= len(p.flat) {
				p.cursor = len(p.flat) - 1
			}
		} else {
			// Leaf spec (no children discovered) — emit SpecSelectedMsg directly.
			sf := node.sf
			return p, func() tea.Msg { return SpecSelectedMsg{Spec: sf} }
		}

	case "e":
		if len(p.flat) == 0 {
			return p, nil
		}
		row := p.flat[p.cursor]
		var path string
		if row.isChild {
			path = p.nodes[row.nodeIdx].children[row.childIdx]
		} else {
			path = p.nodes[row.nodeIdx].sf.Path
		}
		return p, func() tea.Msg { return EditSpecRequestMsg{Path: path} }

	case "n":
		p.inputActive = true
		p.input.Reset()
		p.input.Focus()
		return p, textinput.Blink
	}

	return p, nil
}

// View renders the specs panel as a tree with expand/collapse indicators.
func (p SpecsPanel) View() string {
	if p.inputActive {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Render("New spec name:")
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("Enter to create · Esc to cancel")
		content := lipgloss.JoinVertical(lipgloss.Left,
			prompt,
			p.input.View(),
			hint,
		)
		return lipgloss.NewStyle().
			Width(p.width).Height(p.height).
			Render(content)
	}

	if len(p.nodes) == 0 {
		return lipgloss.NewStyle().
			Width(p.width).Height(p.height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#888888")).
			Render("No specs")
	}

	accentStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	end := p.scrollTop + p.height
	if end > len(p.flat) {
		end = len(p.flat)
	}

	var lines []string
	for i := p.scrollTop; i < end; i++ {
		row := p.flat[i]
		node := p.nodes[row.nodeIdx]
		selected := i == p.cursor

		var line string
		if row.isChild {
			name := filepath.Base(node.children[row.childIdx])
			icon := childFileIcon(name)
			text := fmt.Sprintf("  %s %s", icon, name)
			if selected {
				line = accentStyle.Render(truncateToWidth("> "+strings.TrimLeft(text, " "), p.width))
			} else {
				line = dimStyle.Render(truncateToWidth(text, p.width))
			}
		} else {
			expand := "▶"
			if node.expanded {
				expand = "▼"
			}
			if len(node.children) == 0 {
				expand = " "
			}
			text := fmt.Sprintf("%s %s  %s", expand, node.sf.Status.Symbol(), node.sf.Name)
			if selected {
				line = accentStyle.Render(truncateToWidth("> "+text, p.width))
			} else {
				line = truncateToWidth("  "+text, p.width)
			}
		}
		lines = append(lines, line)
	}

	// Pad to height with blank lines.
	for len(lines) < p.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}
