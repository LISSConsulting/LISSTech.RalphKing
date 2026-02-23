package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

func TestNew(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch)

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
	m := New(ch)
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch)

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
	m := New(ch)

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
	m := New(ch)

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
	m := New(ch)

	updated, _ := m.Update(loopDoneMsg{})
	model := updated.(Model)

	if !model.done {
		t.Error("expected done to be true after loopDoneMsg")
	}
}

func TestUpdateLoopErr(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch)

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
			m := New(ch)

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
	m := New(ch)

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
	m := New(ch)

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
	m := New(ch)

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
	m := New(ch)
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
	m := New(ch)
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
	m := New(ch)
	m.maxIter = 0

	header := m.renderHeader()
	if !strings.Contains(header, "∞") {
		t.Errorf("header should show ∞ for unlimited iterations, got: %s", header)
	}
}

func TestRenderFooterContent(t *testing.T) {
	ch := make(chan loop.LogEntry, 1)
	m := New(ch)
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
	m := New(ch)
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
