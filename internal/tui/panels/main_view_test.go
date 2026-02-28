package panels

import (
	"strings"
	"testing"
)

func TestNewMainView(t *testing.T) {
	mv := NewMainView(80, 20)
	if mv.activeTab != TabOutput {
		t.Errorf("activeTab: got %v, want TabOutput", mv.activeTab)
	}
}

func TestMainView_AppendLine(t *testing.T) {
	mv := NewMainView(80, 20)
	mv = mv.AppendLine("line 1")
	mv = mv.AppendLine("line 2")

	view := mv.View()
	for _, want := range []string{"line 1", "line 2"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; got %q", want, view)
		}
	}
}

func TestMainView_ShowSpec(t *testing.T) {
	mv := NewMainView(80, 20)
	mv = mv.ShowSpec("# Spec Title\n\nSome content.")
	if mv.activeTab != TabSpecContent {
		t.Errorf("ShowSpec should switch to TabSpecContent, got %v", mv.activeTab)
	}
	view := mv.View()
	if !strings.Contains(view, "Spec Title") {
		t.Errorf("View() after ShowSpec missing content; got %q", view)
	}
}

func TestMainView_ShowIterationLog(t *testing.T) {
	mv := NewMainView(80, 20)
	mv = mv.ShowIterationLog([]string{"[12:00:00]  tool call", "[12:00:01]  result"})
	if mv.activeTab != TabIterationDetail {
		t.Errorf("ShowIterationLog should switch to TabIterationDetail, got %v", mv.activeTab)
	}
	view := mv.View()
	if !strings.Contains(view, "tool call") {
		t.Errorf("View() after ShowIterationLog missing content; got %q", view)
	}
}

func TestMainView_SwitchToOutput(t *testing.T) {
	mv := NewMainView(80, 20)
	mv = mv.ShowSpec("content")
	mv = mv.SwitchToOutput()
	if mv.activeTab != TabOutput {
		t.Errorf("SwitchToOutput should return to TabOutput, got %v", mv.activeTab)
	}
}

func TestMainView_TabSwitching(t *testing.T) {
	mv := NewMainView(80, 20)
	// ] should cycle forward
	mv2, _ := mv.Update(keyMsg("]"))
	if mv2.activeTab == mv.activeTab {
		t.Error("'] key should change active tab")
	}
}

func TestMainView_SetSize(t *testing.T) {
	mv := NewMainView(80, 20)
	mv = mv.SetSize(100, 30)
	if mv.width != 100 || mv.height != 30 {
		t.Errorf("SetSize: got %dx%d, want 100x30", mv.width, mv.height)
	}
}

func TestMainView_View_ContainsTabBar(t *testing.T) {
	mv := NewMainView(80, 20)
	view := mv.View()
	if !strings.Contains(view, "Output") {
		t.Errorf("View() should contain tab label 'Output'; got %q", view)
	}
}
