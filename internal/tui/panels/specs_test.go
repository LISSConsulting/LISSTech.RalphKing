package panels

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

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

func TestSpecsPanel_EKey_EmitsEditRequest(t *testing.T) {
	specs := []spec.SpecFile{makeSpec("foo", "specs/foo.md", spec.StatusNotStarted)}
	p := NewSpecsPanel(specs, 80, 20)
	_, cmd := p.Update(keyMsg("e"))
	if cmd == nil {
		t.Fatal("'e' key on selected spec should return a cmd")
	}
	msg := cmd()
	req, ok := msg.(EditSpecRequestMsg)
	if !ok {
		t.Fatalf("expected EditSpecRequestMsg, got %T", msg)
	}
	if req.Path != "specs/foo.md" {
		t.Errorf("EditSpecRequestMsg.Path = %q, want %q", req.Path, "specs/foo.md")
	}
}

func TestSpecsPanel_EKey_EmptyPanel_NoCmd(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	_, cmd := p.Update(keyMsg("e"))
	if cmd != nil {
		t.Error("'e' on empty panel should return nil cmd")
	}
}

func TestSpecsPanel_NKey_ActivatesInput(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	p2, _ := p.Update(keyMsg("n"))
	if !p2.inputActive {
		t.Error("'n' key should activate inputActive")
	}
	view := p2.View()
	if !strings.Contains(view, "New spec name:") {
		t.Errorf("View() when inputActive should show prompt; got %q", view)
	}
}

func TestSpecsPanel_NKey_Submit_EmitsCreateRequest(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	// Activate input.
	p, _ = p.Update(keyMsg("n"))
	// Type a name character by character.
	p, _ = p.Update(keyMsg("t"))
	p, _ = p.Update(keyMsg("e"))
	p, _ = p.Update(keyMsg("s"))
	p, _ = p.Update(keyMsg("t"))
	// Submit.
	p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter with typed name should return a cmd")
	}
	msg := cmd()
	req, ok := msg.(CreateSpecRequestMsg)
	if !ok {
		t.Fatalf("expected CreateSpecRequestMsg, got %T", msg)
	}
	if req.Name != "test" {
		t.Errorf("CreateSpecRequestMsg.Name = %q, want %q", req.Name, "test")
	}
	if p2.inputActive {
		t.Error("inputActive should be false after submission")
	}
}

func TestSpecsPanel_NKey_EmptySubmit_NoCmd(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	p, _ = p.Update(keyMsg("n"))
	// Submit immediately without typing anything.
	p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter with empty name should return nil cmd")
	}
	if !p2.inputActive {
		t.Error("inputActive should remain true after empty submit")
	}
}

func TestSpecsPanel_NKey_Esc_Cancels(t *testing.T) {
	p := NewSpecsPanel(nil, 80, 20)
	p, _ = p.Update(keyMsg("n"))
	if !p.inputActive {
		t.Fatal("setup: inputActive should be true after 'n'")
	}
	p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("esc should return nil cmd")
	}
	if p2.inputActive {
		t.Error("inputActive should be false after esc")
	}
}

func TestSpecItem_Description(t *testing.T) {
	item := specItem{sf: makeSpec("foo", "specs/foo.md", spec.StatusDone)}
	if got := item.Description(); got != "specs/foo.md" {
		t.Errorf("Description() = %q, want %q", got, "specs/foo.md")
	}
}

func TestSpecItem_FilterValue(t *testing.T) {
	item := specItem{sf: makeSpec("bar", "specs/bar.md", spec.StatusDone)}
	if got := item.FilterValue(); got != "bar" {
		t.Errorf("FilterValue() = %q, want %q", got, "bar")
	}
}

func TestSpecDelegate_Update(t *testing.T) {
	d := specDelegate{}
	l := list.New(nil, d, 80, 20)
	cmd := d.Update(nil, &l)
	if cmd != nil {
		t.Error("specDelegate.Update() should return nil cmd")
	}
}

func TestSpecDelegate_Render_WrongType(t *testing.T) {
	d := specDelegate{}
	l := list.New(nil, d, 80, 20)
	var buf bytes.Buffer
	d.Render(&buf, l, 0, iterItem{})
	if buf.Len() != 0 {
		t.Error("Render() with wrong item type should not write anything")
	}
}

func TestSpecsPanel_Update_JK_Enter(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("spec-a", "specs/spec-a.md", spec.StatusNotStarted),
		makeSpec("spec-b", "specs/spec-b.md", spec.StatusNotStarted),
	}
	p := NewSpecsPanel(specs, 80, 20)

	// j key — may emit SpecSelectedMsg.
	_, _ = p.Update(keyMsg("j"))

	// k key — may emit SpecSelectedMsg.
	_, _ = p.Update(keyMsg("k"))

	// enter key — should emit SpecSelectedMsg.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on non-empty panel should return cmd")
	}
	if _, ok := cmd().(SpecSelectedMsg); !ok {
		t.Errorf("enter should emit SpecSelectedMsg, got %T", cmd())
	}

	// Default key in normal mode — delegates to list.
	_, _ = p.Update(keyMsg("x"))

	// Non-key message — delegates to list.
	_, _ = p.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
}

func TestSpecsPanel_JK_EmitsSpecSelectedMsg(t *testing.T) {
	// The j and k cmds must be invoked to cover the closure body.
	specs := []spec.SpecFile{
		makeSpec("spec-a", "specs/spec-a.md", spec.StatusNotStarted),
		makeSpec("spec-b", "specs/spec-b.md", spec.StatusNotStarted),
	}
	p := NewSpecsPanel(specs, 80, 20)

	_, cmd := p.Update(keyMsg("j"))
	if cmd == nil {
		t.Fatal("'j' on non-empty panel should return a cmd")
	}
	if _, ok := cmd().(SpecSelectedMsg); !ok {
		t.Errorf("j cmd should emit SpecSelectedMsg, got %T", cmd())
	}

	_, cmd = p.Update(keyMsg("k"))
	if cmd == nil {
		t.Fatal("'k' on non-empty panel should return a cmd")
	}
	if _, ok := cmd().(SpecSelectedMsg); !ok {
		t.Errorf("k cmd should emit SpecSelectedMsg, got %T", cmd())
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
