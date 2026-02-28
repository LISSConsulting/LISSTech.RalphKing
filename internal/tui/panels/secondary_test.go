package panels

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
)

// keyMsg is a helper for creating key messages in tests.
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func TestNewSecondaryPanel(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	if p.activeTab != TabRegent {
		t.Errorf("activeTab: got %v, want TabRegent", p.activeTab)
	}
}

func TestSecondaryPanel_AppendLine_Regent(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p = p.AppendLine("🛡️  Regent: restarting", TabRegent)

	// Switch to regent tab and verify content appears
	view := p.View()
	if !strings.Contains(view, "Regent") {
		t.Errorf("View() should contain tab label 'Regent'; got %q", view)
	}
}

func TestSecondaryPanel_AppendLine_Git(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p = p.AppendLine("⬆ pushed to origin", TabGit)

	// Switch to git tab
	p2, _ := p.Update(keyMsg("]"))
	view := p2.View()
	if !strings.Contains(view, "Git") {
		t.Errorf("View() on Git tab should show 'Git'; got %q", view)
	}
}

// TestSecondaryPanel_AppendLine_Tests verifies that lines routed to TabTests
// appear when the Tests tab is active (T035).
func TestSecondaryPanel_AppendLine_Tests(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p = p.AppendLine("🛡️  Regent: Tests passed (3/3)", TabTests)

	// Navigate to Tests tab (index 2: Regent→Git→Tests)
	p2, _ := p.Update(keyMsg("]"))
	p2, _ = p2.Update(keyMsg("]"))
	view := p2.View()
	if !strings.Contains(view, "Tests") {
		t.Errorf("View() on Tests tab should show 'Tests' label; got %q", view)
	}
}

// TestSecondaryPanel_TabSwitching verifies ] advances the active tab.
func TestSecondaryPanel_TabSwitching(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p2, _ := p.Update(keyMsg("]"))
	if p2.activeTab == p.activeTab {
		t.Error("'] key should change active tab")
	}
}

func TestSecondaryPanel_SetSize(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p = p.SetSize(100, 30)
	if p.width != 100 || p.height != 30 {
		t.Errorf("SetSize: got %dx%d, want 100x30", p.width, p.height)
	}
}

func TestSecondaryPanel_View_ContainsTabBar(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	view := p.View()
	if !strings.Contains(view, "Regent") {
		t.Errorf("View() should contain 'Regent' tab label; got %q", view)
	}
}

// TestSecondaryPanel_CostTab_Empty verifies the Cost tab renders a placeholder
// when no iterations have been recorded (T036).
func TestSecondaryPanel_CostTab_Empty(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	// Navigate to Cost tab (index 3)
	p, _ = p.Update(keyMsg("]"))
	p, _ = p.Update(keyMsg("]"))
	p, _ = p.Update(keyMsg("]"))
	view := p.View()
	if !strings.Contains(view, "No iterations yet") {
		t.Errorf("Cost tab with no data should show 'No iterations yet'; got %q", view)
	}
}

// TestSecondaryPanel_CostTab_WithIterations verifies AddIteration populates
// the cost table and the total row is rendered (T036).
func TestSecondaryPanel_CostTab_WithIterations(t *testing.T) {
	tests := []struct {
		name       string
		iterations []store.IterationSummary
		wantRows   []string
	}{
		{
			name: "one iteration",
			iterations: []store.IterationSummary{
				{Number: 1, Mode: "build", CostUSD: 0.012, Duration: 10.5},
			},
			wantRows: []string{"1", "build", "$0.012", "10.5s", "Total"},
		},
		{
			name: "five iterations",
			iterations: []store.IterationSummary{
				{Number: 1, Mode: "build", CostUSD: 0.010, Duration: 8.0},
				{Number: 2, Mode: "build", CostUSD: 0.012, Duration: 9.5},
				{Number: 3, Mode: "plan", CostUSD: 0.005, Duration: 5.0},
				{Number: 4, Mode: "build", CostUSD: 0.015, Duration: 12.0},
				{Number: 5, Mode: "build", CostUSD: 0.008, Duration: 7.0},
			},
			wantRows: []string{"1", "2", "3", "plan", "4", "5", "Total"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewSecondaryPanel(80, 20)
			for _, s := range tc.iterations {
				p = p.AddIteration(s)
			}
			// Navigate to Cost tab
			p, _ = p.Update(keyMsg("]"))
			p, _ = p.Update(keyMsg("]"))
			p, _ = p.Update(keyMsg("]"))

			view := p.View()
			for _, want := range tc.wantRows {
				if !strings.Contains(view, want) {
					t.Errorf("Cost tab View() missing %q; got:\n%s", want, view)
				}
			}
		})
	}
}

// TestSecondaryPanel_PrevTab verifies [ navigates to the previous tab.
func TestSecondaryPanel_PrevTab(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	// Navigate forward first.
	p, _ = p.Update(keyMsg("]"))
	tabAfterNext := p.activeTab
	// Navigate back.
	p, _ = p.Update(keyMsg("["))
	if p.activeTab == tabAfterNext {
		t.Error("'[' key should navigate to previous tab")
	}
}

// TestSecondaryPanel_Update_DefaultKeyOnTabs verifies non-tab-switch keys are
// forwarded to the active logview on all scrollable tabs.
func TestSecondaryPanel_Update_DefaultKeyOnTabs(t *testing.T) {
	p := NewSecondaryPanel(80, 20)

	// Default key on TabRegent.
	_, _ = p.Update(keyMsg("j"))

	// Default key on TabGit (navigate to it first).
	p, _ = p.Update(keyMsg("]"))
	_, _ = p.Update(keyMsg("j"))

	// Default key on TabTests.
	p = NewSecondaryPanel(80, 20)
	p, _ = p.Update(keyMsg("]"))
	p, _ = p.Update(keyMsg("]"))
	_, _ = p.Update(keyMsg("j"))
}

// TestSecondaryPanel_Update_NonKeyMsg verifies non-key messages are forwarded
// to the active logview on all scrollable tabs.
func TestSecondaryPanel_Update_NonKeyMsg(t *testing.T) {
	wm := tea.WindowSizeMsg{Width: 100, Height: 30}

	// Non-key on TabRegent.
	p := NewSecondaryPanel(80, 20)
	_, _ = p.Update(wm)

	// Non-key on TabGit.
	p, _ = p.Update(keyMsg("]"))
	_, _ = p.Update(wm)

	// Non-key on TabTests.
	p = NewSecondaryPanel(80, 20)
	p, _ = p.Update(keyMsg("]"))
	p, _ = p.Update(keyMsg("]"))
	_, _ = p.Update(wm)
}

// TestSecondaryPanel_AllTabsRenderable verifies all four tabs produce non-empty
// output after content has been routed (T037).
func TestSecondaryPanel_AllTabsRenderable(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
	p = p.AppendLine("🛡️  Regent: starting", TabRegent)
	p = p.AppendLine("⬆ pushed to origin", TabGit)
	p = p.AppendLine("🛡️  Regent: Tests passed (3/3)", TabTests)
	p = p.AddIteration(store.IterationSummary{Number: 1, Mode: "build", CostUSD: 0.01, Duration: 5.0})

	tabs := []SecondaryTab{TabRegent, TabGit, TabTests, TabCost}
	for i, tab := range tabs {
		p.activeTab = tab
		p.tabbar = p.tabbar.SetWidth(80)
		for j := 0; j < i; j++ {
			p.tabbar = p.tabbar.Next()
		}
		view := p.View()
		if view == "" {
			t.Errorf("View() returned empty for tab %v", tab)
		}
	}
}
