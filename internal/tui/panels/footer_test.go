package panels

import (
	"strings"
	"testing"
)

func TestRenderFooter_EachFocusTarget(t *testing.T) {
	tests := []struct {
		focus string
		hints []string
	}{
		{"specs", []string{"j/k:navigate", "e:edit", "n:new", "enter:view"}},
		{"iterations", []string{"j/k:navigate", "enter:view"}},
		{"main", []string{"f:follow", "[/]:tab", "ctrl+u/d:scroll"}},
		{"secondary", []string{"[/]:tab", "j/k:scroll"}},
	}

	for _, tt := range tests {
		t.Run(tt.focus, func(t *testing.T) {
			props := FooterProps{Focus: tt.focus, LastCommit: "abc1234"}
			rendered := RenderFooter(props, 200)
			for _, hint := range tt.hints {
				if !strings.Contains(rendered, hint) {
					t.Errorf("RenderFooter(focus=%q) missing hint %q; got %q", tt.focus, hint, rendered)
				}
			}
		})
	}
}

func TestRenderFooter_GlobalHintsAlwaysShown(t *testing.T) {
	props := FooterProps{Focus: "main"}
	rendered := RenderFooter(props, 200)
	for _, global := range []string{"?:help", "q:quit", "1-4:panel", "s:stop"} {
		if !strings.Contains(rendered, global) {
			t.Errorf("global hint %q missing; got %q", global, rendered)
		}
	}
}

func TestRenderFooter_StopRequested(t *testing.T) {
	props := FooterProps{StopRequested: true, Focus: "main"}
	rendered := RenderFooter(props, 200)
	if !strings.Contains(rendered, "stopping after iteration") {
		t.Errorf("stop requested footer missing message; got %q", rendered)
	}
}

func TestRenderFooter_LastCommit(t *testing.T) {
	props := FooterProps{LastCommit: "deadbeef", Focus: "specs"}
	rendered := RenderFooter(props, 200)
	if !strings.Contains(rendered, "deadbeef") {
		t.Errorf("footer missing commit hash; got %q", rendered)
	}
}

func TestRenderFooter_EmptyCommitFallback(t *testing.T) {
	props := FooterProps{LastCommit: ""}
	rendered := RenderFooter(props, 200)
	if !strings.Contains(rendered, "—") {
		t.Errorf("footer should show — for empty commit; got %q", rendered)
	}
}

func TestRenderFooter_NarrowWidth(t *testing.T) {
	props := FooterProps{Focus: "main", LastCommit: "abc"}
	// Should not panic even at very narrow widths
	_ = RenderFooter(props, 30)
}

func TestRenderFooter_ScrollInfo(t *testing.T) {
	t.Run("scrolled up with new below", func(t *testing.T) {
		props := FooterProps{Focus: "main", ScrollOffset: 5, NewBelow: 3}
		rendered := RenderFooter(props, 200)
		if !strings.Contains(rendered, "↓3 new") {
			t.Errorf("expected new-below indicator; got %q", rendered)
		}
		if !strings.Contains(rendered, "↑5") {
			t.Errorf("expected scroll-up indicator; got %q", rendered)
		}
	})

	t.Run("scrolled up no new", func(t *testing.T) {
		props := FooterProps{Focus: "main", ScrollOffset: 2, NewBelow: 0}
		rendered := RenderFooter(props, 200)
		if !strings.Contains(rendered, "↑2") {
			t.Errorf("expected scroll-up indicator; got %q", rendered)
		}
	})
}
