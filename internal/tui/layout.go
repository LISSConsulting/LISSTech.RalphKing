package tui

// Rect represents a rectangular region of the terminal.
type Rect struct {
	X, Y, Width, Height int
}

// Layout holds the computed panel geometry for a given terminal size.
type Layout struct {
	Header, Footer    Rect
	Specs, Iterations Rect
	Main, Secondary   Rect
	TooSmall          bool // true when terminal is below the minimum 80×24
}

// Calculate computes the panel layout for a terminal of the given dimensions.
// Returns a Layout with TooSmall=true if width < 80 or height < 24.
//
// Algorithm:
//   - Header: full width, 1 row at top
//   - Footer: full width, 1 row at bottom
//   - Sidebar: 25% of width, clamped to [24, 35]
//   - Specs: sidebar width × 40% of body height (top of sidebar)
//   - Iterations: sidebar width × remaining body height (bottom of sidebar)
//   - Main: remaining width × 65% of body height (top-right)
//   - Secondary: remaining width × remaining body height (bottom-right)
func Calculate(width, height int) Layout {
	if width < 80 || height < 24 {
		return Layout{TooSmall: true}
	}

	bodyH := height - 2 // subtract header + footer rows

	sidebarW := width * 25 / 100
	if sidebarW < 24 {
		sidebarW = 24
	}
	if sidebarW > 35 {
		sidebarW = 35
	}
	rightW := width - sidebarW

	specsH := bodyH * 40 / 100
	itersH := bodyH - specsH

	mainH := bodyH * 65 / 100
	secH := bodyH - mainH

	return Layout{
		Header:     Rect{X: 0, Y: 0, Width: width, Height: 1},
		Footer:     Rect{X: 0, Y: height - 1, Width: width, Height: 1},
		Specs:      Rect{X: 0, Y: 1, Width: sidebarW, Height: specsH},
		Iterations: Rect{X: 0, Y: 1 + specsH, Width: sidebarW, Height: itersH},
		Main:       Rect{X: sidebarW, Y: 1, Width: rightW, Height: mainH},
		Secondary:  Rect{X: sidebarW, Y: 1 + mainH, Width: rightW, Height: secH},
		TooSmall:   false,
	}
}
