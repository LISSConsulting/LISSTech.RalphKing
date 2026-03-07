package tui

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		tooSmall bool
		sidebarW int
		rightW   int
		bodyH    int
		specsH   int
		itersH   int
		mainH    int
		secH     int
	}{
		{
			name:  "80x24 minimum viable",
			width: 80, height: 24,
			tooSmall: false,
			sidebarW: 24, // clamp to min 24 (80*25/100=20 → clamped to 24)
			rightW:   56,
			bodyH:    22,
			specsH:   8,  // 22*40/100 = 8
			itersH:   14, // 22 - 8
			mainH:    14, // 22*65/100 = 14
			secH:     8,  // 22 - 14
		},
		{
			name:  "120x40",
			width: 120, height: 40,
			tooSmall: false,
			sidebarW: 30, // 120*25/100=30 (in range)
			rightW:   90,
			bodyH:    38,
			specsH:   15, // 38*40/100=15
			itersH:   23, // 38-15
			mainH:    24, // 38*65/100=24
			secH:     14, // 38-24
		},
		{
			name:  "200x60",
			width: 200, height: 60,
			tooSmall: false,
			sidebarW: 35, // 200*25/100=50 → clamped to max 35
			rightW:   165,
			bodyH:    58,
			specsH:   23, // 58*40/100=23
			itersH:   35, // 58-23
			mainH:    37, // 58*65/100=37
			secH:     21, // 58-37
		},
		{
			name:  "79x24 too small (width)",
			width: 79, height: 24,
			tooSmall: true,
		},
		{
			name:  "80x23 too small (height)",
			width: 80, height: 23,
			tooSmall: true,
		},
		{
			name:  "0x0 too small",
			width: 0, height: 0,
			tooSmall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Calculate(tt.width, tt.height)
			if l.TooSmall != tt.tooSmall {
				t.Errorf("TooSmall: got %v, want %v", l.TooSmall, tt.tooSmall)
				return
			}
			if tt.tooSmall {
				return // no further assertions for too-small layouts
			}

			// Header
			if l.Header.Y != 0 || l.Header.Width != tt.width || l.Header.Height != 1 {
				t.Errorf("Header: got %+v", l.Header)
			}

			// Footer
			if l.Footer.Y != tt.height-1 || l.Footer.Width != tt.width || l.Footer.Height != 1 {
				t.Errorf("Footer: got %+v", l.Footer)
			}

			// Sidebar width
			if l.Specs.Width != tt.sidebarW {
				t.Errorf("Specs.Width: got %d, want %d", l.Specs.Width, tt.sidebarW)
			}
			if l.Iterations.Width != tt.sidebarW {
				t.Errorf("Iterations.Width: got %d, want %d", l.Iterations.Width, tt.sidebarW)
			}

			// Right width
			if l.Main.Width != tt.rightW {
				t.Errorf("Main.Width: got %d, want %d", l.Main.Width, tt.rightW)
			}
			if l.Secondary.Width != tt.rightW {
				t.Errorf("Secondary.Width: got %d, want %d", l.Secondary.Width, tt.rightW)
			}

			// Specs and Iterations heights
			if l.Specs.Height != tt.specsH {
				t.Errorf("Specs.Height: got %d, want %d", l.Specs.Height, tt.specsH)
			}
			if l.Iterations.Height != tt.itersH {
				t.Errorf("Iterations.Height: got %d, want %d", l.Iterations.Height, tt.itersH)
			}

			// Main and Secondary heights
			if l.Main.Height != tt.mainH {
				t.Errorf("Main.Height: got %d, want %d", l.Main.Height, tt.mainH)
			}
			if l.Secondary.Height != tt.secH {
				t.Errorf("Secondary.Height: got %d, want %d", l.Secondary.Height, tt.secH)
			}

			// Y positions
			if l.Specs.Y != 1 {
				t.Errorf("Specs.Y: got %d, want 1", l.Specs.Y)
			}
			if l.Iterations.Y != 1+tt.specsH {
				t.Errorf("Iterations.Y: got %d, want %d", l.Iterations.Y, 1+tt.specsH)
			}
			if l.Main.Y != 1 {
				t.Errorf("Main.Y: got %d, want 1", l.Main.Y)
			}
			if l.Secondary.Y != 1+tt.mainH {
				t.Errorf("Secondary.Y: got %d, want %d", l.Secondary.Y, 1+tt.mainH)
			}

			// X positions
			if l.Specs.X != 0 {
				t.Errorf("Specs.X: got %d, want 0", l.Specs.X)
			}
			if l.Main.X != tt.sidebarW {
				t.Errorf("Main.X: got %d, want %d", l.Main.X, tt.sidebarW)
			}

			// Heights sum to bodyH
			if l.Specs.Height+l.Iterations.Height != tt.bodyH {
				t.Errorf("sidebar heights %d+%d != bodyH %d", l.Specs.Height, l.Iterations.Height, tt.bodyH)
			}
			if l.Main.Height+l.Secondary.Height != tt.bodyH {
				t.Errorf("right heights %d+%d != bodyH %d", l.Main.Height, l.Secondary.Height, tt.bodyH)
			}
		})
	}
}

func TestCalculate_SidebarClamp(t *testing.T) {
	t.Run("narrow terminal clamps sidebar to min 24", func(t *testing.T) {
		l := Calculate(80, 24)
		if l.Specs.Width < 24 {
			t.Errorf("sidebar width %d is below minimum 24", l.Specs.Width)
		}
	})

	t.Run("wide terminal clamps sidebar to max 35", func(t *testing.T) {
		l := Calculate(200, 30)
		if l.Specs.Width > 35 {
			t.Errorf("sidebar width %d exceeds maximum 35", l.Specs.Width)
		}
	})
}

func TestTitleContentDims(t *testing.T) {
	tests := []struct {
		name  string
		rect  Rect
		wantW int
		wantH int
	}{
		{
			// innerDims: 28x8; titleContentDims: 28x7
			name:  "normal rect subtracts 1 row for title",
			rect:  Rect{Width: 30, Height: 10},
			wantW: 28,
			wantH: 7,
		},
		{
			// innerDims: 8x2; titleContentDims: 8x1 (clamped)
			name:  "small height clamps to 1",
			rect:  Rect{Width: 10, Height: 4},
			wantW: 8,
			wantH: 1,
		},
		{
			// innerDims: 1x1; titleContentDims: 1x1 (clamped, not 0)
			name:  "minimum dims stay at 1",
			rect:  Rect{Width: 2, Height: 3},
			wantW: 1,
			wantH: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := titleContentDims(tt.rect)
			if w != tt.wantW {
				t.Errorf("titleContentDims(%+v) width = %d, want %d", tt.rect, w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("titleContentDims(%+v) height = %d, want %d", tt.rect, h, tt.wantH)
			}
		})
	}
}

// TestCalculate_LayoutSumInvariant verifies that sidebar and right-column
// panel heights always sum to bodyH for the three canonical terminal sizes.
func TestCalculate_LayoutSumInvariant(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
		w, h := size[0], size[1]
		l := Calculate(w, h)
		if l.TooSmall {
			t.Errorf("Calculate(%d,%d) unexpectedly TooSmall", w, h)
			continue
		}
		bodyH := h - 2
		if got := l.Specs.Height + l.Iterations.Height; got != bodyH {
			t.Errorf("Calculate(%d,%d): sidebar heights %d+%d=%d, want %d",
				w, h, l.Specs.Height, l.Iterations.Height, got, bodyH)
		}
		if got := l.Main.Height + l.Secondary.Height; got != bodyH {
			t.Errorf("Calculate(%d,%d): right heights %d+%d=%d, want %d",
				w, h, l.Main.Height, l.Secondary.Height, got, bodyH)
		}
		if got := l.Specs.Width + l.Main.Width; got != w {
			t.Errorf("Calculate(%d,%d): total width %d+%d=%d, want %d",
				w, h, l.Specs.Width, l.Main.Width, got, w)
		}
	}
}
