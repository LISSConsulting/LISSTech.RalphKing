package panels

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

func makeSpec(name, path string, status spec.Status) spec.SpecFile {
	return spec.SpecFile{Name: name, Path: path, Status: status}
}

// keyMsg is defined in secondary_test.go (shared across panel tests in this package).

// ---- Basic panel tests ----

func TestNewSpecsPanel_Empty(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	view := p.View()
	if !strings.Contains(view, "No specs") {
		t.Errorf("empty panel should show 'No specs'; got %q", view)
	}
}

func TestNewSpecsPanel_WithSpecs(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("spec-one", "specs/spec-one.md", spec.StatusDone),
		makeSpec("spec-two", "specs/spec-two.md", spec.StatusInProgress),
	}
	p := NewSpecsPanel(specs, "", 80, 20)
	view := p.View()
	for _, want := range []string{"spec-one", "spec-two"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing spec %q; got %q", want, view)
		}
	}
}

func TestSpecsPanel_SelectedSpec(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("spec-a", "specs/spec-a.md", spec.StatusNotStarted),
	}
	p := NewSpecsPanel(specs, "", 80, 20)
	sel := p.SelectedSpec()
	if sel == nil {
		t.Fatal("SelectedSpec() returned nil for non-empty panel")
	}
	if sel.Name != "spec-a" {
		t.Errorf("SelectedSpec().Name = %q, want %q", sel.Name, "spec-a")
	}
}

func TestSpecsPanel_SetSize(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	p = p.SetSize(100, 30)
	if p.width != 100 || p.height != 30 {
		t.Errorf("SetSize: got %dx%d, want 100x30", p.width, p.height)
	}
}

// ---- Key navigation tests ----

func TestSpecsPanel_EKey_EmitsEditRequest(t *testing.T) {
	specs := []spec.SpecFile{makeSpec("foo", "specs/foo.md", spec.StatusNotStarted)}
	p := NewSpecsPanel(specs, "", 80, 20)
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
	p := NewSpecsPanel(nil, "", 80, 20)
	_, cmd := p.Update(keyMsg("e"))
	if cmd != nil {
		t.Error("'e' on empty panel should return nil cmd")
	}
}

func TestSpecsPanel_NKey_ActivatesInput(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
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
	p := NewSpecsPanel(nil, "", 80, 20)
	p, _ = p.Update(keyMsg("n"))
	p, _ = p.Update(keyMsg("t"))
	p, _ = p.Update(keyMsg("e"))
	p, _ = p.Update(keyMsg("s"))
	p, _ = p.Update(keyMsg("t"))
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
	p := NewSpecsPanel(nil, "", 80, 20)
	p, _ = p.Update(keyMsg("n"))
	p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter with empty name should return nil cmd")
	}
	if !p2.inputActive {
		t.Error("inputActive should remain true after empty submit")
	}
}

func TestSpecsPanel_NKey_Esc_Cancels(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
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

func TestSpecsPanel_JK_Enter_EmitsSpecSelectedMsg(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("spec-a", "specs/spec-a.md", spec.StatusNotStarted),
		makeSpec("spec-b", "specs/spec-b.md", spec.StatusNotStarted),
	}
	p := NewSpecsPanel(specs, "", 80, 20)

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

	_, cmd = p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on non-empty leaf panel should return cmd")
	}
	if _, ok := cmd().(SpecSelectedMsg); !ok {
		t.Errorf("enter should emit SpecSelectedMsg, got %T", cmd())
	}
}

func TestSpecsPanel_NonKeyMsg_Ignored(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	_, cmd := p.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	if cmd != nil {
		t.Error("non-key message on empty panel should return nil cmd")
	}
}

func TestSpecsPanel_EnterOnEmpty_NoCmd(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on empty panel should return nil cmd")
	}
}

// ---- Tree-specific tests (T060) ----

func TestBuildTree_FlatSpec(t *testing.T) {
	specs := []spec.SpecFile{
		makeSpec("flat-spec", "specs/flat.md", spec.StatusDone),
	}
	nodes := buildTree(specs, "")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].sf.Name != "flat-spec" {
		t.Errorf("node.sf.Name = %q, want %q", nodes[0].sf.Name, "flat-spec")
	}
	if len(nodes[0].children) != 0 {
		t.Errorf("flat spec should have 0 children, got %d", len(nodes[0].children))
	}
}

func TestBuildTree_DirSpec_WithChildren(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "specs", "001-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	// tasks.md is intentionally absent.

	specs := []spec.SpecFile{{
		Name:  "001-feature",
		Path:  filepath.Join("specs", "001-feature", "spec.md"),
		Dir:   filepath.Join("specs", "001-feature"),
		IsDir: true,
	}}
	nodes := buildTree(specs, tmp)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if len(nodes[0].children) != 2 {
		t.Errorf("expected 2 children (spec.md, plan.md), got %d: %v", len(nodes[0].children), nodes[0].children)
	}
}

func TestBuildTree_DirSpec_NoChildren(t *testing.T) {
	specs := []spec.SpecFile{{
		Name:  "001-feature",
		Path:  "specs/001-feature/spec.md",
		Dir:   "specs/001-feature",
		IsDir: true,
	}}
	nodes := buildTree(specs, "")
	if len(nodes[0].children) != 0 {
		t.Errorf("expected 0 children with empty workDir, got %d", len(nodes[0].children))
	}
}

func TestFlattenTree_CollapsedDir(t *testing.T) {
	nodes := []specTreeNode{
		{sf: makeSpec("a", "a.md", spec.StatusDone), children: []string{"a/spec.md"}, expanded: false},
		{sf: makeSpec("b", "b.md", spec.StatusDone)},
	}
	flat := flattenTree(nodes)
	if len(flat) != 2 {
		t.Fatalf("collapsed dir should yield 2 rows, got %d", len(flat))
	}
}

func TestFlattenTree_ExpandedDir(t *testing.T) {
	nodes := []specTreeNode{
		{sf: makeSpec("a", "a.md", spec.StatusDone), children: []string{"a/spec.md", "a/plan.md"}, expanded: true},
		{sf: makeSpec("b", "b.md", spec.StatusDone)},
	}
	flat := flattenTree(nodes)
	// 1 (dir a) + 2 (children) + 1 (dir b) = 4
	if len(flat) != 4 {
		t.Fatalf("expanded dir should yield 4 rows, got %d", len(flat))
	}
	if !flat[1].isChild || flat[1].nodeIdx != 0 || flat[1].childIdx != 0 {
		t.Errorf("row 1 should be first child of node 0; got %+v", flat[1])
	}
	if !flat[2].isChild || flat[2].childIdx != 1 {
		t.Errorf("row 2 should be second child of node 0; got %+v", flat[2])
	}
	if flat[3].isChild {
		t.Errorf("row 3 should be dir b, not a child")
	}
}

func TestSpecsPanel_EnterOnDir_TogglesExpand(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "specs", "001-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"spec.md", "plan.md"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	specs := []spec.SpecFile{{
		Name:  "001-feature",
		Path:  filepath.Join("specs", "001-feature", "spec.md"),
		Dir:   filepath.Join("specs", "001-feature"),
		IsDir: true,
	}}

	p := NewSpecsPanel(specs, tmp, 80, 20)
	if len(p.flat) != 1 {
		t.Fatalf("before expand: expected 1 flat row, got %d", len(p.flat))
	}

	// Press enter → should expand (no cmd emitted).
	p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on dir row should return nil cmd")
	}
	// 1 dir + 2 children = 3 rows.
	if len(p2.flat) != 3 {
		t.Fatalf("after expand: expected 3 flat rows, got %d", len(p2.flat))
	}

	// Press enter again → collapse.
	p3, _ := p2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(p3.flat) != 1 {
		t.Fatalf("after collapse: expected 1 flat row, got %d", len(p3.flat))
	}
}

func TestSpecsPanel_EnterOnChild_EmitsSpecSelectedMsgWithChildPath(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "specs", "002-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"spec.md", "plan.md"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	specs := []spec.SpecFile{{
		Name:  "002-feature",
		Path:  filepath.Join("specs", "002-feature", "spec.md"),
		Dir:   filepath.Join("specs", "002-feature"),
		IsDir: true,
	}}

	p := NewSpecsPanel(specs, tmp, 80, 20)
	// Expand.
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Move to first child (cursor=1).
	p, _ = p.Update(keyMsg("j"))
	// Enter on child.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on child row should return a cmd")
	}
	msg, ok := cmd().(SpecSelectedMsg)
	if !ok {
		t.Fatalf("expected SpecSelectedMsg, got %T", cmd())
	}
	// The path should be the child file path (ends with spec.md).
	if !strings.HasSuffix(msg.Spec.Path, "spec.md") {
		t.Errorf("SpecSelectedMsg.Spec.Path = %q, expected to end with spec.md", msg.Spec.Path)
	}
}

func TestChildFileIcon(t *testing.T) {
	tests := []struct {
		name string
		icon string
	}{
		{"spec.md", "📋"},
		{"plan.md", "📐"},
		{"tasks.md", "✅"},
		{"other.md", "📄"},
	}
	for _, tt := range tests {
		if got := childFileIcon(tt.name); got != tt.icon {
			t.Errorf("childFileIcon(%q) = %q, want %q", tt.name, got, tt.icon)
		}
	}
}

func TestSpecsPanel_View_ShowsExpandIndicator(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "specs", "003-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	specs := []spec.SpecFile{{
		Name:  "003-feature",
		Path:  filepath.Join("specs", "003-feature", "spec.md"),
		Dir:   filepath.Join("specs", "003-feature"),
		IsDir: true,
	}}

	p := NewSpecsPanel(specs, tmp, 80, 20)
	if !strings.Contains(p.View(), "▶") {
		t.Error("collapsed dir should show ▶ indicator")
	}
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(p.View(), "▼") {
		t.Error("expanded dir should show ▼ indicator")
	}
}

func TestSpecsPanel_SelectedSpec_ReturnsParentForChild(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "specs", "004-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	specs := []spec.SpecFile{{
		Name:  "004-feature",
		Path:  filepath.Join("specs", "004-feature", "spec.md"),
		Dir:   filepath.Join("specs", "004-feature"),
		IsDir: true,
	}}
	p := NewSpecsPanel(specs, tmp, 80, 20)
	p, _ = p.Update(tea.KeyMsg{Type: tea.KeyEnter}) // expand
	p, _ = p.Update(keyMsg("j"))                    // move to child

	sel := p.SelectedSpec()
	if sel == nil {
		t.Fatal("SelectedSpec() returned nil for child row")
	}
	if sel.Name != "004-feature" {
		t.Errorf("SelectedSpec().Name = %q, want %q", sel.Name, "004-feature")
	}
}

func TestSpecsPanel_SelectedSpec_EmptyPanel_ReturnsNil(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	if p.SelectedSpec() != nil {
		t.Error("SelectedSpec() on empty panel should return nil")
	}
}

// buildManySpecs creates n flat specs for scroll tests.
func buildManySpecs(n int) []spec.SpecFile {
	specs := make([]spec.SpecFile, n)
	for i := range specs {
		specs[i] = makeSpec(fmt.Sprintf("spec-%02d", i), fmt.Sprintf("specs/spec-%02d.md", i), spec.StatusNotStarted)
	}
	return specs
}

func TestSpecsPanel_ScrollDown_AdjustsScrollTop(t *testing.T) {
	p := NewSpecsPanel(buildManySpecs(10), "", 80, 3)
	// Navigate to the last item — each j should scroll when we go past the visible window.
	for i := 0; i < 9; i++ {
		p, _ = p.Update(keyMsg("j"))
	}
	if p.cursor != 9 {
		t.Fatalf("cursor = %d, want 9", p.cursor)
	}
	// Cursor must be within the visible window [scrollTop, scrollTop+height).
	if p.cursor < p.scrollTop || p.cursor >= p.scrollTop+p.height {
		t.Errorf("cursor %d not visible: scrollTop=%d height=%d", p.cursor, p.scrollTop, p.height)
	}
	if p.scrollTop <= 0 {
		t.Errorf("scrollTop = %d; expected > 0 after scrolling down 9 rows with height 3", p.scrollTop)
	}
}

func TestSpecsPanel_ScrollUp_AdjustsScrollTop(t *testing.T) {
	p := NewSpecsPanel(buildManySpecs(10), "", 80, 3)
	// Scroll to bottom.
	for i := 0; i < 9; i++ {
		p, _ = p.Update(keyMsg("j"))
	}
	// Now scroll back to top.
	for i := 0; i < 9; i++ {
		p, _ = p.Update(keyMsg("k"))
	}
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0", p.cursor)
	}
	if p.scrollTop != 0 {
		t.Errorf("scrollTop = %d, want 0 after scrolling back to top", p.scrollTop)
	}
}

func TestSpecsPanel_JKOnEmpty_NoCmd(t *testing.T) {
	p := NewSpecsPanel(nil, "", 80, 20)
	_, cmd := p.Update(keyMsg("j"))
	if cmd != nil {
		t.Error("j on empty panel should return nil cmd")
	}
	_, cmd = p.Update(keyMsg("k"))
	if cmd != nil {
		t.Error("k on empty panel should return nil cmd")
	}
}
