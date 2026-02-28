package panels

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderHeader_BasicFields(t *testing.T) {
	accent := lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
	now := time.Date(2026, 1, 1, 15, 30, 0, 0, time.UTC)

	props := HeaderProps{
		ProjectName: "MyProject",
		Branch:      "main",
		Mode:        "build",
		Iteration:   3,
		MaxIter:     10,
		TotalCost:   1.23,
		StateSymbol: "●",
		StateLabel:  "BUILDING",
		Clock:       now,
	}

	rendered := RenderHeader(props, 200, accent)

	for _, want := range []string{"MyProject", "main", "build", "3/10", "$1.23", "● BUILDING", "15:30"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("RenderHeader() missing %q; output: %q", want, rendered)
		}
	}
}

func TestRenderHeader_EmptyFieldFallbacks(t *testing.T) {
	accent := lipgloss.NewStyle()
	props := HeaderProps{} // all zero values

	rendered := RenderHeader(props, 200, accent)

	// Fallbacks: no project name → "RalphKing", no branch → "—", no mode → "—"
	for _, want := range []string{"RalphKing", "—"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("RenderHeader() with empty props missing %q; got %q", want, rendered)
		}
	}
}

func TestRenderHeader_WorkDir(t *testing.T) {
	accent := lipgloss.NewStyle()
	props := HeaderProps{WorkDir: "/home/user/myproject"}

	rendered := RenderHeader(props, 200, accent)
	if !strings.Contains(rendered, "dir:") {
		t.Errorf("RenderHeader() missing 'dir:' with non-empty WorkDir; got %q", rendered)
	}
}

func TestRenderHeader_NoWorkDir(t *testing.T) {
	accent := lipgloss.NewStyle()
	props := HeaderProps{WorkDir: ""}

	rendered := RenderHeader(props, 200, accent)
	if strings.Contains(rendered, "dir:") {
		t.Errorf("RenderHeader() should omit 'dir:' when WorkDir is empty; got %q", rendered)
	}
}

func TestRenderHeader_UnlimitedIter(t *testing.T) {
	accent := lipgloss.NewStyle()
	props := HeaderProps{Iteration: 5, MaxIter: 0}

	rendered := RenderHeader(props, 200, accent)
	if !strings.Contains(rendered, "5/∞") {
		t.Errorf("RenderHeader() MaxIter=0 should show ∞; got %q", rendered)
	}
}

func TestRenderHeader_AllLoopStates(t *testing.T) {
	accent := lipgloss.NewStyle()
	states := []struct{ symbol, label string }{
		{"✓", "IDLE"},
		{"●", "PLANNING"},
		{"●", "BUILDING"},
		{"✗", "FAILED"},
		{"⟳", "REGENT RESTART"},
	}

	for _, s := range states {
		props := HeaderProps{StateSymbol: s.symbol, StateLabel: s.label}
		rendered := RenderHeader(props, 200, accent)
		if !strings.Contains(rendered, s.symbol) {
			t.Errorf("RenderHeader() missing symbol %q for state %s; got %q", s.symbol, s.label, rendered)
		}
		if !strings.Contains(rendered, s.label) {
			t.Errorf("RenderHeader() missing label %q; got %q", s.label, rendered)
		}
	}
}

func TestRenderHeader_Elapsed(t *testing.T) {
	accent := lipgloss.NewStyle()
	props := HeaderProps{Elapsed: 95 * time.Second}

	rendered := RenderHeader(props, 200, accent)
	if !strings.Contains(rendered, "elapsed:") {
		t.Errorf("RenderHeader() should show elapsed time; got %q", rendered)
	}
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{5 * time.Second, "5s"},
		{90 * time.Second, "1m30s"},
		{3*time.Hour + 15*time.Minute, "3h15m"},
		{0, "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatElapsed(tt.d)
			if got != tt.want {
				t.Errorf("FormatElapsed(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
