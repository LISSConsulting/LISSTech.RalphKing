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

func TestMainView_SummaryTab(t *testing.T) {
	summaryLines := []string{
		"Iteration:   3",
		"Mode:        build",
		"Cost:        $0.0234",
		"Duration:    45.2s",
		"Exit:        success",
		"Commit:      abc1234",
	}

	mv := NewMainView(80, 20)
	mv = mv.SetIterationSummary(summaryLines)

	// Tab is still on Output after SetIterationSummary (no auto-switch).
	if mv.activeTab != TabOutput {
		t.Errorf("SetIterationSummary should not change activeTab, got %v", mv.activeTab)
	}

	// Cycle ] three times: Output → Spec → Iteration → Summary.
	mv, _ = mv.Update(keyMsg("]"))
	mv, _ = mv.Update(keyMsg("]"))
	mv, _ = mv.Update(keyMsg("]"))
	if mv.activeTab != TabIterationSummary {
		t.Errorf("after 3x ], expected TabIterationSummary, got %v", mv.activeTab)
	}

	view := mv.View()
	for _, want := range summaryLines {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q; got:\n%s", want, view)
		}
	}
}

func TestMainView_SummaryTab_ContainsLabel(t *testing.T) {
	mv := NewMainView(80, 20)
	view := mv.View()
	if !strings.Contains(view, "Summary") {
		t.Errorf("View() should contain tab label 'Summary'; got %q", view)
	}
}

func TestMainView_SummaryTab_ScrollDelegation(t *testing.T) {
	// Verify scroll keys are delegated to summaryLogview when on summary tab.
	mv := NewMainView(80, 20)
	mv = mv.SetIterationSummary([]string{"line1", "line2"})
	// Advance to TabIterationSummary.
	mv, _ = mv.Update(keyMsg("]"))
	mv, _ = mv.Update(keyMsg("]"))
	mv, _ = mv.Update(keyMsg("]"))
	// Sending a scroll key should not panic.
	mv, _ = mv.Update(keyMsg("j"))
	_ = mv.View()
}
