package panels

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewWorktreesPanel_Empty(t *testing.T) {
	p := NewWorktreesPanel(nil, 40, 10)
	view := p.View()
	if !strings.Contains(strings.ToLower(view), "no worktrees") {
		t.Errorf("empty panel should show 'No worktrees' hint; got: %q", view)
	}
}

func TestNewWorktreesPanel_WithEntries(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/feat-a", State: "running", Iterations: 3, TotalCost: 0.05, SpecName: "feat-a"},
		{Branch: "wt/feat-b", State: "completed", Iterations: 5, TotalCost: 0.12, SpecName: "feat-b"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	view := p.View()
	if !strings.Contains(view, "wt/feat-a") {
		t.Errorf("panel view should contain first branch name; got: %q", view)
	}
}

func TestWorktreesPanel_SelectedBranch_Initial(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
		{Branch: "wt/beta", State: "completed"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	got := p.SelectedBranch()
	if got != "wt/alpha" {
		t.Errorf("initial selection should be first entry %q, got %q", "wt/alpha", got)
	}
}

func TestWorktreesPanel_SelectedBranch_Empty(t *testing.T) {
	p := NewWorktreesPanel(nil, 40, 10)
	if p.SelectedBranch() != "" {
		t.Error("SelectedBranch on empty panel should return empty string")
	}
}

func TestWorktreesPanel_Navigation_J(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
		{Branch: "wt/beta", State: "completed"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	// j moves down to second entry.
	p2, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if p2.SelectedBranch() != "wt/beta" {
		t.Errorf("after j, selected branch should be wt/beta, got %q", p2.SelectedBranch())
	}
}

func TestWorktreesPanel_Navigation_K(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
		{Branch: "wt/beta", State: "completed"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	// Move down first, then k to go back up.
	p2, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	p3, _ := p2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if p3.SelectedBranch() != "wt/alpha" {
		t.Errorf("after j then k, selected branch should be wt/alpha, got %q", p3.SelectedBranch())
	}
}

func TestWorktreesPanel_Enter_EmitsWorktreeSelected(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key should return a cmd")
	}
	msg := cmd()
	sel, ok := msg.(WorktreeSelectedMsg)
	if !ok {
		t.Fatalf("expected WorktreeSelectedMsg, got %T", msg)
	}
	if sel.Branch != "wt/alpha" {
		t.Errorf("WorktreeSelectedMsg.Branch = %q, want %q", sel.Branch, "wt/alpha")
	}
}

func TestWorktreesPanel_KeyX_EmitsStop(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd == nil {
		t.Fatal("x key should return a cmd")
	}
	msg := cmd()
	act, ok := msg.(WorktreeActionMsg)
	if !ok {
		t.Fatalf("expected WorktreeActionMsg, got %T", msg)
	}
	if act.Action != "stop" {
		t.Errorf("Action = %q, want %q", act.Action, "stop")
	}
	if act.Branch != "wt/alpha" {
		t.Errorf("Branch = %q, want %q", act.Branch, "wt/alpha")
	}
}

func TestWorktreesPanel_KeyM_EmitsMerge(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "completed"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	if cmd == nil {
		t.Fatal("M key should return a cmd")
	}
	msg := cmd()
	act, ok := msg.(WorktreeActionMsg)
	if !ok {
		t.Fatalf("expected WorktreeActionMsg, got %T", msg)
	}
	if act.Action != "merge" {
		t.Errorf("Action = %q, want %q", act.Action, "merge")
	}
}

func TestWorktreesPanel_KeyD_EmitsClean(t *testing.T) {
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "stopped"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	if cmd == nil {
		t.Fatal("D key should return a cmd")
	}
	msg := cmd()
	act, ok := msg.(WorktreeActionMsg)
	if !ok {
		t.Fatalf("expected WorktreeActionMsg, got %T", msg)
	}
	if act.Action != "clean" {
		t.Errorf("Action = %q, want %q", act.Action, "clean")
	}
}

func TestWorktreesPanel_ActionKeys_EmptyPanel_NoCmd(t *testing.T) {
	p := NewWorktreesPanel(nil, 40, 10)
	for _, key := range []string{"x", "M", "D", "enter"} {
		_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		// On an empty panel, SelectedBranch() == "" so no cmd should be returned.
		// (The list.Update might return a cmd from internal bubbles state.)
		// We just verify no panic and that no WorktreeActionMsg is emitted.
		if cmd != nil {
			msg := cmd()
			if _, isAction := msg.(WorktreeActionMsg); isAction {
				t.Errorf("key %q on empty panel should not emit WorktreeActionMsg", key)
			}
			if _, isSel := msg.(WorktreeSelectedMsg); isSel {
				t.Errorf("key %q on empty panel should not emit WorktreeSelectedMsg", key)
			}
		}
	}
}

func TestWorktreesPanel_SetSize(t *testing.T) {
	p := NewWorktreesPanel(nil, 40, 10)
	p2 := p.SetSize(80, 20)
	if p2.width != 80 || p2.height != 20 {
		t.Errorf("SetSize: got (%d, %d), want (80, 20)", p2.width, p2.height)
	}
}

func TestWorktreesPanel_SetSize_ZeroHeight(t *testing.T) {
	p := NewWorktreesPanel(nil, 40, 0)
	_ = p.View() // must not panic
}

func TestWorktreesPanel_SetEntries_Replaces(t *testing.T) {
	p := NewWorktreesPanel(nil, 60, 10)
	if p.SelectedBranch() != "" {
		t.Fatal("setup: empty panel should have no selection")
	}
	p2 := p.SetEntries([]WorktreeEntry{
		{Branch: "wt/new", State: "running"},
	})
	if p2.SelectedBranch() != "wt/new" {
		t.Errorf("after SetEntries, SelectedBranch = %q, want %q", p2.SelectedBranch(), "wt/new")
	}
	if len(p2.entries) != 1 {
		t.Errorf("entries length = %d, want 1", len(p2.entries))
	}
}

func TestWorktreeStateIcon_AllStates(t *testing.T) {
	states := []string{"creating", "running", "completed", "failed", "stopped", "merging", "merged", "merge_failed", "removed", "unknown-state"}
	for _, s := range states {
		icon := worktreeStateIcon(s)
		if icon == "" {
			t.Errorf("worktreeStateIcon(%q) returned empty string", s)
		}
	}
}

func TestWorktreesPanel_View_StatusIcons(t *testing.T) {
	// Verify that the running state icon appears in the panel view.
	entries := []WorktreeEntry{
		{Branch: "wt/alpha", State: "running"},
	}
	p := NewWorktreesPanel(entries, 60, 10)
	view := p.View()
	icon := worktreeStateIcon("running")
	if !strings.Contains(view, icon) {
		t.Errorf("running state icon %q not found in view: %q", icon, view)
	}
}

func TestWorktreesPanel_Update_OtherKey_Passthrough(t *testing.T) {
	p := NewWorktreesPanel(nil, 60, 10)
	// An unhandled key should fall through to the underlying list without panicking.
	p2, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	_ = p2.View() // must not panic
}
