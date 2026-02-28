package panels

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
)

func makeSummary(n int, mode, subtype string, cost, dur float64) store.IterationSummary {
	return store.IterationSummary{
		Number:   n,
		Mode:     mode,
		CostUSD:  cost,
		Duration: dur,
		Subtype:  subtype,
	}
}

func TestNewIterationsPanel(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	if p.SelectedIteration() != nil {
		t.Error("expected nil selection on empty panel")
	}
}

func TestIterationsPanel_AddIteration(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.AddIteration(makeSummary(1, "build", "success", 0.01, 1.5))
	p = p.AddIteration(makeSummary(2, "build", "success", 0.02, 2.0))

	if len(p.iterations) != 2 {
		t.Errorf("expected 2 iterations, got %d", len(p.iterations))
	}
}

func TestIterationsPanel_SetCurrent(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.AddIteration(makeSummary(1, "build", "success", 0.01, 1.5))
	p = p.SetCurrent(1)

	if p.currentNum == nil || *p.currentNum != 1 {
		t.Error("SetCurrent(1) did not set currentNum")
	}
}

func TestIterationsPanel_SetCurrent_Clear(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.SetCurrent(3)
	p = p.SetCurrent(0)

	if p.currentNum != nil {
		t.Errorf("SetCurrent(0) should clear currentNum, got %v", p.currentNum)
	}
}

func TestIterationsPanel_View_Empty(t *testing.T) {
	p := NewIterationsPanel(80, 5)
	view := p.View()
	if view == "" {
		t.Error("View() should not return empty string even when no iterations")
	}
}

func TestIterationsPanel_SetSize(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.SetSize(100, 30)
	if p.width != 100 || p.height != 30 {
		t.Errorf("SetSize: got %dx%d, want 100x30", p.width, p.height)
	}
}

func TestIterItem_Description(t *testing.T) {
	t.Run("running", func(t *testing.T) {
		item := iterItem{summary: makeSummary(1, "build", "", 0, 0), running: true}
		if got := item.Description(); got != "running…" {
			t.Errorf("Description() = %q, want %q", got, "running…")
		}
	})
	t.Run("completed", func(t *testing.T) {
		item := iterItem{summary: makeSummary(1, "build", "success", 0.01, 1.5)}
		desc := item.Description()
		if !strings.Contains(desc, "$") {
			t.Errorf("Description() = %q, want to contain '$'", desc)
		}
	})
}

func TestIterItem_FilterValue(t *testing.T) {
	item := iterItem{summary: makeSummary(3, "build", "success", 0, 0)}
	if got := item.FilterValue(); got != "#3" {
		t.Errorf("FilterValue() = %q, want %q", got, "#3")
	}
}

func TestIterDelegate_Update(t *testing.T) {
	d := iterDelegate{}
	l := list.New(nil, d, 80, 20)
	cmd := d.Update(nil, &l)
	if cmd != nil {
		t.Error("iterDelegate.Update() should return nil cmd")
	}
}

func TestIterDelegate_Render(t *testing.T) {
	d := iterDelegate{}
	l := list.New(nil, d, 80, 20)
	item := iterItem{summary: makeSummary(1, "build", "success", 0.01, 1.5)}

	// Render at index 1 (non-selected since list.Index() starts at 0).
	var buf bytes.Buffer
	d.Render(&buf, l, 1, item)
	if buf.Len() == 0 {
		t.Error("Render() should write output for a valid iterItem")
	}

	// Render with wrong item type — should write nothing.
	buf.Reset()
	d.Render(&buf, l, 0, specItem{})
	if buf.Len() != 0 {
		t.Error("Render() with wrong item type should not write anything")
	}
}

func TestIterationsPanel_Update_Keys(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.AddIteration(makeSummary(1, "build", "success", 0.01, 1.5))
	p = p.AddIteration(makeSummary(2, "build", "success", 0.02, 2.0))

	// j key — should navigate down.
	p2, _ := p.Update(keyMsg("j"))
	_ = p2

	// k key — should navigate up.
	p3, _ := p.Update(keyMsg("k"))
	_ = p3

	// enter key — should emit IterationSelectedMsg.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on non-empty panel should return cmd")
	}
	if _, ok := cmd().(IterationSelectedMsg); !ok {
		t.Error("enter should emit IterationSelectedMsg")
	}

	// Other key — delegates to list.
	p4, _ := p.Update(keyMsg("x"))
	_ = p4

	// Non-key message.
	p5, _ := p.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	_ = p5
}

func TestIterationsPanel_SelectedIteration_WithItem(t *testing.T) {
	p := NewIterationsPanel(80, 20)
	p = p.AddIteration(makeSummary(7, "build", "success", 0.07, 7.0))
	sel := p.SelectedIteration()
	if sel == nil {
		t.Fatal("SelectedIteration() should return non-nil after AddIteration")
	}
	if sel.Number != 7 {
		t.Errorf("SelectedIteration().Number = %d, want 7", sel.Number)
	}
}

func TestIterationsPanel_View_WithCurrent(t *testing.T) {
	p := NewIterationsPanel(80, 5)
	p = p.SetCurrent(1) // running indicator but no completed iterations
	view := p.View()
	if view == "" {
		t.Error("View() with running iteration should not be empty")
	}
}

func TestIterItem_Title(t *testing.T) {
	tests := []struct {
		name    string
		item    iterItem
		wantSub string
	}{
		{"success", iterItem{summary: makeSummary(1, "build", "success", 0, 0)}, "✓"},
		{"error", iterItem{summary: makeSummary(2, "build", "error", 0, 0)}, "✗"},
		{"running", iterItem{summary: makeSummary(3, "build", "", 0, 0), running: true}, "●"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := tt.item.Title()
			if len(title) == 0 {
				t.Error("Title() returned empty string")
			}
			// The status character should appear somewhere in title
			found := false
			for _, r := range title {
				if string(r) == tt.wantSub {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Title() = %q, want to contain %q", title, tt.wantSub)
			}
		})
	}
}
