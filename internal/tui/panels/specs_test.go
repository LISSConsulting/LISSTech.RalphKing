package panels

import (
	"strings"
	"testing"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

func makeSpec(name, path string, status spec.Status) spec.SpecFile {
	return spec.SpecFile{Name: name, Path: path, Status: status}
}

func TestNewSpecsPanel_Empty(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	view := p.View()
	if !strings.Contains(view, "No specs") {
		t.Errorf("empty panel should show 'No specs'; got %q", view)
	}
}

func TestNewSpecsPanel_WithSpecs(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("ralph-core", "specs/ralph-core.md", spec.StatusDone),
		makeSpec("the-regent", "specs/the-regent.md", spec.StatusInProgress),
	}
	p := NewSpecsPanel(specs, 80, 20)
	view := p.View()
	for _, want := range []string{"ralph-core", "the-regent"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing spec %q; got %q", want, view)
		}
	}
}

func TestSpecsPanel_SelectedSpec(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("spec-a", "specs/spec-a.md", spec.StatusNotStarted),
	}
	p := NewSpecsPanel(specs, 80, 20)
	sel := p.SelectedSpec()
	if sel == nil {
		t.Fatal("SelectedSpec() returned nil for non-empty panel")
	}
	if sel.Name != "spec-a" {
		t.Errorf("SelectedSpec().Name = %q, want %q", sel.Name, "spec-a")
	}
}

func TestSpecsPanel_SetSize(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	p = p.SetSize(100, 30)
	if p.width != 100 || p.height != 30 {
		t.Errorf("SetSize: got %dx%d, want 100x30", p.width, p.height)
	}
}

func TestSpecItem_Title_ContainsStatusSymbol(t *testing.T) {
	tests := []struct {
		status spec.Status
		symbol string
	}{
		{spec.StatusDone, "✅"},
		{spec.StatusInProgress, "🔄"},
		{spec.StatusNotStarted, "⬜"},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			item := specItem{sf: makeSpec("test", "specs/test.md", tt.status)}
			title := item.Title()
			if !strings.Contains(title, tt.symbol) {
				t.Errorf("specItem.Title() = %q, want to contain %q", title, tt.symbol)
			}
		})
	}
}
