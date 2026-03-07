package components

import (
	"strings"
	"testing"
)

func TestNewTabBar(t *testing.T) {
	tb := NewTabBar([]string{"Output", "Spec", "Iteration"})
	if tb.Active() != 0 {
		t.Errorf("Active: got %d, want 0", tb.Active())
	}
}

func TestTabBar_Next(t *testing.T) {
	tabs := []string{"A", "B", "C"}
	tb := NewTabBar(tabs)

	tests := []struct {
		wantActive int
	}{
		{1}, {2}, {0}, // wrap
	}
	for _, tt := range tests {
		tb = tb.Next()
		if tb.Active() != tt.wantActive {
			t.Errorf("Active after Next: got %d, want %d", tb.Active(), tt.wantActive)
		}
	}
}

func TestTabBar_Prev(t *testing.T) {
	tabs := []string{"A", "B", "C"}
	tb := NewTabBar(tabs)

	tests := []struct {
		wantActive int
	}{
		{2}, {1}, {0}, // wrap
	}
	for _, tt := range tests {
		tb = tb.Prev()
		if tb.Active() != tt.wantActive {
			t.Errorf("Active after Prev: got %d, want %d", tb.Active(), tt.wantActive)
		}
	}
}

func TestTabBar_View_ContainsAllTabs(t *testing.T) {
	labels := []string{"Output", "Spec", "Iteration"}
	tb := NewTabBar(labels)
	view := tb.View()
	for _, label := range labels {
		if !strings.Contains(view, label) {
			t.Errorf("View() missing label %q: got %q", label, view)
		}
	}
}

func TestTabBar_View_TwoTabs(t *testing.T) {
	tb := NewTabBar([]string{"Alpha", "Beta"})
	view := tb.View()
	if !strings.Contains(view, "Alpha") || !strings.Contains(view, "Beta") {
		t.Errorf("View() = %q, want both tabs", view)
	}
}

func TestTabBar_View_FourTabs(t *testing.T) {
	tb := NewTabBar([]string{"A", "B", "C", "D"})
	view := tb.View()
	for _, label := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(view, label) {
			t.Errorf("View() missing %q", label)
		}
	}
}

func TestTabBar_Empty(t *testing.T) {
	tb := NewTabBar(nil)
	view := tb.View()
	if view != "" {
		t.Errorf("empty TabBar View() = %q, want empty string", view)
	}
	// Next/Prev on empty should not panic
	_ = tb.Next()
	_ = tb.Prev()
}

func TestTabBar_SetWidth(t *testing.T) {
	tb := NewTabBar([]string{"Tab1", "Tab2"}).SetWidth(50)
	if tb.width != 50 {
		t.Errorf("width: got %d, want 50", tb.width)
	}
}

func TestTabBar_SetActive(t *testing.T) {
	tb := NewTabBar([]string{"A", "B", "C"})
	tb = tb.Next().Next() // active = 2
	tb = tb.SetActive(0)
	if tb.Active() != 0 {
		t.Errorf("SetActive(0): got %d, want 0", tb.Active())
	}
	// Out of bounds clamped.
	tb = tb.SetActive(99)
	if tb.Active() != 2 {
		t.Errorf("SetActive(99): got %d, want 2 (clamped)", tb.Active())
	}
	tb = tb.SetActive(-1)
	if tb.Active() != 0 {
		t.Errorf("SetActive(-1): got %d, want 0 (clamped)", tb.Active())
	}
	// Empty TabBar.
	empty := NewTabBar(nil)
	_ = empty.SetActive(0) // must not panic
}

func TestTabBar_CycleWraps(t *testing.T) {
	tabs := []string{"A", "B"}
	tb := NewTabBar(tabs)
	// Full forward cycle
	for i := 0; i < len(tabs); i++ {
		tb = tb.Next()
	}
	if tb.Active() != 0 {
		t.Errorf("expected wrap to 0, got %d", tb.Active())
	}
}

func TestTabBar_View_WithWidth(t *testing.T) {
	// Exercises the t.width > 0 branch in View() that applies MaxWidth truncation.
	tb := NewTabBar([]string{"Alpha", "Beta", "Gamma"}).SetWidth(40)
	view := tb.View()
	if !strings.Contains(view, "Alpha") {
		t.Errorf("View() with width should contain first tab label; got %q", view)
	}
}
