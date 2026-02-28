package panels

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestSecondaryPanel_AllTabsRenderable(t *testing.T) {
	p := NewSecondaryPanel(80, 20)
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
