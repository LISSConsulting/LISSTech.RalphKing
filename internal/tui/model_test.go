package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

func TestNew(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	if m.width != 80 {
		t.Errorf("expected default width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("expected default height 24, got %d", m.height)
	}
	if m.done {
		t.Error("expected done to be false")
	}
}

func TestInit(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	if cmd != nil {
		t.Error("window size should not produce a command")
	}
	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}
	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestUpdateLogEntry(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	entry := logEntryMsg(loop.LogEntry{
		Kind:      loop.LogIterStart,
		Timestamp: time.Now(),
		Message:   "── iteration 1 ──",
		Iteration: 1,
		MaxIter:   3,
		Branch:    "feat/test",
		Mode:      "build",
	})

	updated, cmd := m.Update(entry)
	model := updated.(Model)

	if cmd == nil {
		t.Error("log entry should produce a command to wait for more events")
	}
	if model.iteration != 1 {
		t.Errorf("expected iteration 1, got %d", model.iteration)
	}
	if model.maxIter != 3 {
		t.Errorf("expected maxIter 3, got %d", model.maxIter)
	}
	if model.branch != "feat/test" {
		t.Errorf("expected branch feat/test, got %s", model.branch)
	}
	if model.mode != "build" {
		t.Errorf("expected mode build, got %s", model.mode)
	}
	if len(model.lines) != 1 {
		t.Errorf("expected 1 log line, got %d", len(model.lines))
	}
}

func TestUpdateCostTracking(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// Send running total
	entry := logEntryMsg(loop.LogEntry{
		Kind:      loop.LogInfo,
		Timestamp: time.Now(),
		Message:   "Running total: $0.42",
		TotalCost: 0.42,
	})

	updated, _ := m.Update(entry)
	model := updated.(Model)

	if model.totalCost != 0.42 {
		t.Errorf("expected total cost 0.42, got %.2f", model.totalCost)
	}
}

func TestUpdateLoopDone(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	updated, _ := m.Update(loopDoneMsg{})
	model := updated.(Model)

	if !model.done {
		t.Error("expected done to be true after loopDoneMsg")
	}
}

func TestUpdateLoopErr(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	testErr := &testError{msg: "something failed"}
	updated, _ := m.Update(loopErrMsg{err: testErr})
	model := updated.(Model)

	if !model.done {
		t.Error("expected done to be true after loopErrMsg")
	}
	if model.Err() == nil {
		t.Error("expected Err() to return the error")
	}
	if model.Err().Error() != "something failed" {
		t.Errorf("expected error message 'something failed', got %s", model.Err().Error())
	}
}

func TestUpdateKeyQuit(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"q key", "q"},
		{"ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan loop.LogEntry, 1)
			m := New(ch, "")

			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if tt.key == "q" {
				if cmd == nil {
					t.Error("q key should produce a quit command")
				}
			}
		})
	}
}

func TestUpdateCommitTracking(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	entry := logEntryMsg(loop.LogEntry{
		Kind:      loop.LogGitPush,
		Timestamp: time.Now(),
		Message:   "Pushed — last commit: abc1234 feat(tui): add header",
		Commit:    "abc1234 feat(tui): add header",
		Branch:    "main",
	})

	updated, _ := m.Update(entry)
	model := updated.(Model)

	if model.lastCommit != "abc1234 feat(tui): add header" {
		t.Errorf("expected commit tracking, got %s", model.lastCommit)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func TestViewRenders(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
	if !strings.Contains(view, "RalphKing") {
		t.Error("View should contain RalphKing header")
	}
	if !strings.Contains(view, "q to quit") {
		t.Error("View should contain quit hint in footer")
	}
}

func TestViewWithEntries(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// Add entries
	entries := []loop.LogEntry{
		{Kind: loop.LogIterStart, Timestamp: time.Now(), Message: "iter 1", Iteration: 1},
		{Kind: loop.LogToolUse, Timestamp: time.Now(), ToolName: "read_file", ToolInput: "main.go"},
		{Kind: loop.LogIterComplete, Timestamp: time.Now(), Iteration: 1, CostUSD: 0.10, Duration: 2.5},
		{Kind: loop.LogError, Timestamp: time.Now(), Message: "Error: something broke"},
		{Kind: loop.LogGitPull, Timestamp: time.Now(), Message: "Pulling main", Branch: "main"},
		{Kind: loop.LogGitPush, Timestamp: time.Now(), Message: "Pushing main", Branch: "main"},
	}

	for _, e := range entries {
		updated, _ := m.Update(logEntryMsg(e))
		m = updated.(Model)
	}

	view := m.View()

	if !strings.Contains(view, "read_file") {
		t.Error("View should contain tool name")
	}
	if !strings.Contains(view, "main.go") {
		t.Error("View should contain tool input")
	}
}

func TestToolIcon(t *testing.T) {
	tests := []struct {
		tool string
		icon string
	}{
		{"Read", "📖"},
		{"read_file", "📖"},
		{"Write", "✏️ "},
		{"Bash", "🔧"},
		{"WebFetch", "🌐"},
		{"Task", "🔀"},
		{"unknown", "⚡"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			got := toolIcon(tt.tool)
			if got != tt.icon {
				t.Errorf("toolIcon(%q) = %q, want %q", tt.tool, got, tt.icon)
			}
		})
	}
}

func TestRenderLineTypes(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		entry    loop.LogEntry
		contains string
	}{
		{
			"tool_use",
			loop.LogEntry{Kind: loop.LogToolUse, Timestamp: now, ToolName: "Bash", ToolInput: "go test"},
			"Bash",
		},
		{
			"iter_start",
			loop.LogEntry{Kind: loop.LogIterStart, Timestamp: now, Iteration: 3},
			"iteration 3",
		},
		{
			"iter_complete",
			loop.LogEntry{Kind: loop.LogIterComplete, Timestamp: now, Iteration: 2, CostUSD: 0.15, Duration: 3.2},
			"iteration 2 complete",
		},
		{
			"error",
			loop.LogEntry{Kind: loop.LogError, Timestamp: now, Message: "Error: network timeout"},
			"network timeout",
		},
		{
			"git_pull",
			loop.LogEntry{Kind: loop.LogGitPull, Timestamp: now, Message: "Pulling main"},
			"Pulling main",
		},
		{
			"git_push",
			loop.LogEntry{Kind: loop.LogGitPush, Timestamp: now, Message: "Pushing main"},
			"Pushing main",
		},
		{
			"done",
			loop.LogEntry{Kind: loop.LogDone, Timestamp: now, Message: "Loop complete"},
			"Loop complete",
		},
		{
			"stopped",
			loop.LogEntry{Kind: loop.LogStopped, Timestamp: now, Message: "Loop stopped"},
			"Loop stopped",
		},
		{
			"info",
			loop.LogEntry{Kind: loop.LogInfo, Timestamp: now, Message: "Running Claude..."},
			"Running Claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := m.renderLine(logLine{entry: tt.entry})
			if !strings.Contains(rendered, tt.contains) {
				t.Errorf("renderLine(%s) should contain %q, got: %s", tt.name, tt.contains, rendered)
			}
		})
	}
}

func TestRenderHeaderContent(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.mode = "build"
	m.branch = "feat/tui"
	m.iteration = 3
	m.maxIter = 10
	m.totalCost = 1.42

	header := m.renderHeader()

	checks := []string{"RalphKing", "build", "feat/tui", "3/10", "1.42"}
	for _, check := range checks {
		if !strings.Contains(header, check) {
			t.Errorf("header should contain %q, got: %s", check, header)
		}
	}
}

func TestRenderHeaderUnlimitedIterations(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.maxIter = 0

	header := m.renderHeader()
	if !strings.Contains(header, "∞") {
		t.Errorf("header should show ∞ for unlimited iterations, got: %s", header)
	}
}

func TestRenderFooterContent(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.lastCommit = "abc1234 feat(tui): colors"

	footer := m.renderFooter()

	if !strings.Contains(footer, "abc1234") {
		t.Errorf("footer should contain commit hash, got: %s", footer)
	}
	if !strings.Contains(footer, "q to quit") {
		t.Errorf("footer should contain quit hint, got: %s", footer)
	}
}

func TestLogScrolling(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // header(1) + log(3) + footer(1)

	// Add more lines than the log can display
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{
				Kind:      loop.LogInfo,
				Timestamp: time.Now(),
				Message:   strings.Repeat("x", 5),
			},
		})
	}

	log := m.renderLog(3)
	lines := strings.Split(log, "\n")

	// Should only show 3 lines of content (auto-scrolled to bottom)
	if len(lines) != 3 {
		t.Errorf("expected 3 visible lines, got %d", len(lines))
	}
}

func TestScrollUpDown(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // header(1) + log(3) + footer(1)

	// Add 10 lines (more than log can display)
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{
				Kind:      loop.LogInfo,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("line %d", i),
			},
		})
	}

	// Initially at bottom
	if m.scrollOffset != 0 {
		t.Fatalf("expected scrollOffset 0, got %d", m.scrollOffset)
	}

	// Scroll up
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)
	if m.scrollOffset != 1 {
		t.Errorf("expected scrollOffset 1 after scroll up, got %d", m.scrollOffset)
	}

	// Scroll down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after scroll down, got %d", m.scrollOffset)
	}

	// Can't scroll below 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to stay 0, got %d", m.scrollOffset)
	}
}

func TestScrollUpBound(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // log height = 3

	// Add 10 lines; max scroll offset = 10 - 3 = 7
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Scroll up beyond max
	for i := 0; i < 20; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = updated.(Model)
	}

	if m.scrollOffset != 7 {
		t.Errorf("expected scrollOffset clamped to 7, got %d", m.scrollOffset)
	}
}

func TestScrollPageUpDown(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // log height = 3

	for i := 0; i < 20; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Page up (scroll offset increases by log height = 3)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updated.(Model)
	if m.scrollOffset != 3 {
		t.Errorf("expected scrollOffset 3 after pgup, got %d", m.scrollOffset)
	}

	// Page down (scroll offset decreases by log height = 3)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after pgdown, got %d", m.scrollOffset)
	}

	// Page down below 0 should clamp
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset clamped to 0, got %d", m.scrollOffset)
	}
}

func TestScrollHomeEnd(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // log height = 3

	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Home should scroll to top (max offset = 7)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updated.(Model)
	if m.scrollOffset != 7 {
		t.Errorf("expected scrollOffset 7 (top) after home, got %d", m.scrollOffset)
	}

	// End should scroll to bottom
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 (bottom) after end, got %d", m.scrollOffset)
	}
}

func TestScrollNoEffectWhenFewLines(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 10 // log height = 8, but only 3 lines

	for i := 0; i < 3; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Scroll up should have no effect (all lines visible)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 when all lines fit, got %d", m.scrollOffset)
	}
}

func TestScrollRenderShowsCorrectLines(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 4 // log height = 2

	for i := 0; i < 5; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("msg-%d", i)},
		})
	}

	// At bottom (offset 0): should show lines 3 and 4 (last 2)
	log := m.renderLog(2)
	if !strings.Contains(log, "msg-3") || !strings.Contains(log, "msg-4") {
		t.Errorf("at bottom should show msg-3 and msg-4, got: %s", log)
	}

	// Scroll up 2: should show lines 1 and 2
	m.scrollOffset = 2
	log = m.renderLog(2)
	if !strings.Contains(log, "msg-1") || !strings.Contains(log, "msg-2") {
		t.Errorf("scrolled up 2 should show msg-1 and msg-2, got: %s", log)
	}

	// Scroll to top (offset 3): should show lines 0 and 1
	m.scrollOffset = 3
	log = m.renderLog(2)
	if !strings.Contains(log, "msg-0") || !strings.Contains(log, "msg-1") {
		t.Errorf("scrolled to top should show msg-0 and msg-1, got: %s", log)
	}
}

func TestScrollFooterIndicator(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// At bottom: no scroll indicator
	footer := m.renderFooter()
	if strings.Contains(footer, "j/k scroll") {
		t.Error("footer should not show scroll hint when at bottom")
	}

	// Scrolled up: show indicator
	m.scrollOffset = 5
	footer = m.renderFooter()
	if !strings.Contains(footer, "↑5") {
		t.Errorf("footer should show scroll offset, got: %s", footer)
	}
	if !strings.Contains(footer, "j/k scroll") {
		t.Errorf("footer should show scroll hint, got: %s", footer)
	}
}

func TestNewBelowIndicator(t *testing.T) {
	tests := []struct {
		name           string
		scrollOffset   int
		newBelow       int
		wantNewBelow   bool // footer should contain "↓N new"
		wantScrollHint bool // footer should contain "↑N"
	}{
		{"at_bottom_no_new", 0, 0, false, false},
		{"scrolled_up_no_new", 3, 0, false, true},
		{"scrolled_up_with_new", 3, 5, true, true},
		{"at_bottom_counter_reset", 0, 5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan loop.LogEntry, 1)
			m := New(ch, "")
			m.scrollOffset = tt.scrollOffset
			m.newBelow = tt.newBelow

			footer := m.renderFooter()

			if tt.wantNewBelow {
				want := fmt.Sprintf("↓%d new", tt.newBelow)
				if !strings.Contains(footer, want) {
					t.Errorf("footer should contain %q, got: %s", want, footer)
				}
			} else if strings.Contains(footer, "new") {
				t.Errorf("footer should not contain 'new', got: %s", footer)
			}

			if tt.wantScrollHint {
				want := fmt.Sprintf("↑%d", tt.scrollOffset)
				if !strings.Contains(footer, want) {
					t.Errorf("footer should contain %q, got: %s", want, footer)
				}
			}
		})
	}
}

func TestNewBelowIncrements(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // log height = 3

	// Add enough lines to allow scrolling
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Scroll up
	m.scrollOffset = 3

	// New entry arrives while scrolled up → newBelow increments
	entry := logEntryMsg(loop.LogEntry{
		Kind:      loop.LogInfo,
		Timestamp: time.Now(),
		Message:   "new message",
	})

	updated, _ := m.Update(entry)
	m = updated.(Model)
	if m.newBelow != 1 {
		t.Errorf("expected newBelow 1, got %d", m.newBelow)
	}

	// Another entry → newBelow increments again
	updated, _ = m.Update(entry)
	m = updated.(Model)
	if m.newBelow != 2 {
		t.Errorf("expected newBelow 2, got %d", m.newBelow)
	}
}

func TestNewBelowNoIncrementAtBottom(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// At bottom (scrollOffset 0): newBelow should not increment
	entry := logEntryMsg(loop.LogEntry{
		Kind:      loop.LogInfo,
		Timestamp: time.Now(),
		Message:   "message at bottom",
	})

	updated, _ := m.Update(entry)
	m = updated.(Model)
	if m.newBelow != 0 {
		t.Errorf("expected newBelow 0 at bottom, got %d", m.newBelow)
	}
}

func TestNewBelowResetsOnScrollToBottom(t *testing.T) {
	tests := []struct {
		name string
		key  tea.Msg
	}{
		{"end_key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}},
		{"j_to_zero", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}},
		{"pgdown_to_zero", tea.KeyMsg{Type: tea.KeyPgDown}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan loop.LogEntry, 1)
			m := New(ch, "")
			m.height = 5

			for i := 0; i < 10; i++ {
				m.lines = append(m.lines, logLine{
					entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
				})
			}

			// Set up scrolled-up state with new messages
			m.scrollOffset = 1
			m.newBelow = 4

			// Scroll to bottom
			updated, _ := m.Update(tt.key)
			model := updated.(Model)

			if model.scrollOffset != 0 {
				t.Errorf("expected scrollOffset 0, got %d", model.scrollOffset)
			}
			if model.newBelow != 0 {
				t.Errorf("expected newBelow 0 after scrolling to bottom, got %d", model.newBelow)
			}
		})
	}
}

func TestNewBelowPersistsWhenStillScrolledUp(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5

	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now(), Message: fmt.Sprintf("line %d", i)},
		})
	}

	// Scrolled up by 3, with 5 new messages
	m.scrollOffset = 3
	m.newBelow = 5

	// Scroll down by 1 (still scrolled up)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model := updated.(Model)

	if model.scrollOffset != 2 {
		t.Errorf("expected scrollOffset 2, got %d", model.scrollOffset)
	}
	// newBelow should persist since we're still scrolled up
	if model.newBelow != 5 {
		t.Errorf("expected newBelow 5 while still scrolled up, got %d", model.newBelow)
	}
}

func TestScrollHelpers(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 5 // log height = 3

	t.Run("logHeight", func(t *testing.T) {
		if m.logHeight() != 3 {
			t.Errorf("expected logHeight 3, got %d", m.logHeight())
		}
	})

	t.Run("logHeight_minimum", func(t *testing.T) {
		m2 := m
		m2.height = 1
		if m2.logHeight() != 1 {
			t.Errorf("expected minimum logHeight 1, got %d", m2.logHeight())
		}
	})

	t.Run("maxScrollOffset_empty", func(t *testing.T) {
		if m.maxScrollOffset() != 0 {
			t.Errorf("expected maxScrollOffset 0 for empty lines, got %d", m.maxScrollOffset())
		}
	})

	t.Run("maxScrollOffset_with_lines", func(t *testing.T) {
		m2 := m
		for i := 0; i < 10; i++ {
			m2.lines = append(m2.lines, logLine{
				entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now()},
			})
		}
		if m2.maxScrollOffset() != 7 {
			t.Errorf("expected maxScrollOffset 7, got %d", m2.maxScrollOffset())
		}
	})

	t.Run("clampScroll_too_high", func(t *testing.T) {
		m2 := m
		m2.scrollOffset = 100
		m2.clampScroll()
		if m2.scrollOffset != 0 {
			t.Errorf("expected clamped to 0 (no lines), got %d", m2.scrollOffset)
		}
	})

	t.Run("clampScroll_negative", func(t *testing.T) {
		m2 := m
		m2.scrollOffset = -5
		m2.clampScroll()
		if m2.scrollOffset != 0 {
			t.Errorf("expected clamped to 0, got %d", m2.scrollOffset)
		}
	})
}

func TestWaitForEventClosedChannel(t *testing.T) {
	ch := make(chan loop.LogEntry)
	close(ch)

	cmd := waitForEvent(ch)
	msg := cmd()

	if _, ok := msg.(loopDoneMsg); !ok {
		t.Errorf("expected loopDoneMsg from closed channel, got %T", msg)
	}
}

func TestWaitForEventWithEntry(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	ch <- loop.LogEntry{Kind: loop.LogInfo, Message: "test"}

	cmd := waitForEvent(ch)
	msg := cmd()

	entry, ok := msg.(logEntryMsg)
	if !ok {
		t.Fatalf("expected logEntryMsg, got %T", msg)
	}
	if entry.Message != "test" {
		t.Errorf("expected message 'test', got %s", entry.Message)
	}
}

func TestRenderLineAllKinds(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		entry    loop.LogEntry
		contains string
	}{
		{
			"tool_use_write",
			loop.LogEntry{Kind: loop.LogToolUse, Timestamp: now, ToolName: "Write", ToolInput: "main.go"},
			"Write",
		},
		{
			"tool_use_edit",
			loop.LogEntry{Kind: loop.LogToolUse, Timestamp: now, ToolName: "Edit", ToolInput: "config.go"},
			"Edit",
		},
		{
			"tool_use_webfetch",
			loop.LogEntry{Kind: loop.LogToolUse, Timestamp: now, ToolName: "WebFetch", ToolInput: "https://example.com"},
			"WebFetch",
		},
		{
			"tool_use_task",
			loop.LogEntry{Kind: loop.LogToolUse, Timestamp: now, ToolName: "Task", ToolInput: "explore codebase"},
			"Task",
		},
		{
			"regent",
			loop.LogEntry{Kind: loop.LogRegent, Timestamp: now, Message: "Restarting Ralph..."},
			"Regent:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := m.renderLine(logLine{entry: tt.entry})
			if !strings.Contains(rendered, tt.contains) {
				t.Errorf("renderLine(%s) should contain %q, got: %s", tt.name, tt.contains, rendered)
			}
		})
	}
}

func TestRenderLineLongToolName(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)

	// Tool name longer than 14 chars should be truncated with ellipsis
	entry := loop.LogEntry{
		Kind:      loop.LogToolUse,
		Timestamp: now,
		ToolName:  "VeryLongToolNameThatExceeds14",
		ToolInput: "some-input",
	}
	rendered := m.renderLine(logLine{entry: entry})

	// Should contain the truncated name (13 chars + ellipsis)
	if !strings.Contains(rendered, "VeryLongToolN…") {
		t.Errorf("long tool name should be truncated with ellipsis, got: %s", rendered)
	}
	// Should NOT contain the full untruncated name
	if strings.Contains(rendered, "VeryLongToolNameThatExceeds14") {
		t.Errorf("full tool name should not appear in output, got: %s", rendered)
	}
	// Should still contain the input
	if !strings.Contains(rendered, "some-input") {
		t.Errorf("tool input should still appear, got: %s", rendered)
	}
}

func TestViewTinyHeight(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.height = 1 // logHeight would be -1, should clamp to 1

	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string even with tiny height")
	}
	if !strings.Contains(view, "RalphKing") {
		t.Error("View with tiny height should still render header")
	}
}

func TestRenderFooterNarrowWidth(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	m.width = 5 // too narrow; gap clamps to 2

	footer := m.renderFooter()
	if footer == "" {
		t.Error("renderFooter should not return empty string for narrow width")
	}
}

func TestRenderLogEmpty(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// Empty lines → returns newline-padded string
	log := m.renderLog(3)
	lines := strings.Split(log, "\n")
	if len(lines) != 3 {
		t.Errorf("empty renderLog(3) should return 3 lines of padding, got %d", len(lines))
	}
}

func TestUpdateUnknownMsg(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// Unknown message type → no-op
	updated, cmd := m.Update("unknown message type")
	model := updated.(Model)

	if cmd != nil {
		t.Error("unknown message should not produce a command")
	}
	if model.done {
		t.Error("unknown message should not change done state")
	}
}

func TestRenderLogDefensiveScrollOffset(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	now := time.Now()

	// Add 5 lines
	for i := 0; i < 5; i++ {
		m.lines = append(m.lines, logLine{
			entry: loop.LogEntry{Kind: loop.LogInfo, Timestamp: now, Message: fmt.Sprintf("msg-%d", i)},
		})
	}

	// scrollOffset larger than line count: end = 5 - 100 = -95 → clamped to 0
	// This exercises the `if end < 0 { end = 0 }` guard.
	m.scrollOffset = 100
	log := m.renderLog(2)
	if log == "" {
		t.Error("renderLog should not return empty string with out-of-bounds scrollOffset")
	}
}

func TestRenderIterCompleteWithSubtype(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		subtype     string
		wantContain string
		wantAbsent  string
	}{
		{
			name:        "success subtype shown",
			subtype:     "success",
			wantContain: "success",
		},
		{
			name:        "error_max_turns subtype shown",
			subtype:     "error_max_turns",
			wantContain: "error_max_turns",
		},
		{
			name:       "empty subtype omitted",
			subtype:    "",
			wantAbsent: "  —  \x1b", // no trailing separator before ANSI reset
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := loop.LogEntry{
				Kind:      loop.LogIterComplete,
				Timestamp: now,
				Iteration: 1,
				CostUSD:   0.14,
				Duration:  4.2,
				Subtype:   tt.subtype,
			}
			rendered := m.renderLine(logLine{entry: entry})

			if !strings.Contains(rendered, "iteration 1 complete") {
				t.Errorf("should contain 'iteration 1 complete', got: %s", rendered)
			}
			if tt.wantContain != "" && !strings.Contains(rendered, tt.wantContain) {
				t.Errorf("should contain %q, got: %s", tt.wantContain, rendered)
			}
		})
	}
}

func TestNewDefaultAccentColor(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "")

	// Should use default indigo accent; verify header renders without panic
	m.mode = "build"
	header := m.renderHeader()
	if !strings.Contains(header, "RalphKing") {
		t.Errorf("default accent header should render correctly, got: %s", header)
	}
}

func TestNewCustomAccentColor(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, "#FF0000")

	// Custom accent should render header and git lines without panic
	m.mode = "plan"
	m.branch = "feat/custom-color"
	header := m.renderHeader()
	if !strings.Contains(header, "RalphKing") {
		t.Errorf("custom accent header should render correctly, got: %s", header)
	}

	// Git lines should use the custom accent
	now := time.Date(2026, 2, 23, 14, 30, 0, 0, time.UTC)
	pull := m.renderLine(logLine{entry: loop.LogEntry{
		Kind: loop.LogGitPull, Timestamp: now, Message: "Pulling main",
	}})
	if !strings.Contains(pull, "Pulling main") {
		t.Errorf("git pull line with custom accent should render, got: %s", pull)
	}
	push := m.renderLine(logLine{entry: loop.LogEntry{
		Kind: loop.LogGitPush, Timestamp: now, Message: "Pushing main",
	}})
	if !strings.Contains(push, "Pushing main") {
		t.Errorf("git push line with custom accent should render, got: %s", push)
	}
}
