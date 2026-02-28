package panels

import (
	"testing"

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
