package components

import (
	"strings"
	"testing"
)

func TestNewLogView(t *testing.T) {
	lv := NewLogView(80, 24)
	if !lv.Following() {
		t.Error("NewLogView: expected follow mode to be enabled by default")
	}
	if lv.width != 80 || lv.height != 24 {
		t.Errorf("dimensions: got %dx%d, want 80x24", lv.width, lv.height)
	}
}

func TestLogView_AppendLine(t *testing.T) {
	lv := NewLogView(80, 10)
	lv = lv.AppendLine("line 1")
	lv = lv.AppendLine("line 2")
	lv = lv.AppendLine("line 3")

	if len(lv.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lv.lines))
	}

	view := lv.View()
	for _, want := range []string{"line 1", "line 2", "line 3"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q: %q", want, view)
		}
	}
}

func TestLogView_SetContent(t *testing.T) {
	lv := NewLogView(80, 10)
	lv = lv.AppendLine("old line")
	lv = lv.SetContent([]string{"new 1", "new 2"})

	if len(lv.lines) != 2 {
		t.Errorf("expected 2 lines after SetContent, got %d", len(lv.lines))
	}
	if lv.lines[0] != "new 1" || lv.lines[1] != "new 2" {
		t.Errorf("lines mismatch: %v", lv.lines)
	}
}

func TestLogView_SetContent_IndependentCopy(t *testing.T) {
	lv := NewLogView(80, 10)
	original := []string{"a", "b"}
	lv = lv.SetContent(original)
	original[0] = "mutated"
	if lv.lines[0] != "a" {
		t.Error("SetContent should copy the slice, not reference it")
	}
}

func TestLogView_ToggleFollow(t *testing.T) {
	lv := NewLogView(80, 10)
	if !lv.Following() {
		t.Error("initial follow should be true")
	}
	lv = lv.ToggleFollow()
	if lv.Following() {
		t.Error("after first toggle follow should be false")
	}
	lv = lv.ToggleFollow()
	if !lv.Following() {
		t.Error("after second toggle follow should be true")
	}
}

func TestLogView_SetSize(t *testing.T) {
	lv := NewLogView(80, 10)
	lv = lv.SetSize(100, 20)
	if lv.width != 100 || lv.height != 20 {
		t.Errorf("SetSize: got %dx%d, want 100x20", lv.width, lv.height)
	}
	if lv.vp.Width != 100 || lv.vp.Height != 20 {
		t.Errorf("viewport dimensions: got %dx%d, want 100x20", lv.vp.Width, lv.vp.Height)
	}
}

func TestLogView_View_Empty(t *testing.T) {
	lv := NewLogView(80, 10)
	// Should not panic; returns viewport content (may be empty or whitespace).
	_ = lv.View()
}
