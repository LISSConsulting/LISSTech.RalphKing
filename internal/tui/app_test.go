package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/panels"
)

// mockStoreReader is a minimal store.Reader for unit tests.
type mockStoreReader struct {
	iterations []store.IterationSummary
	entries    []loop.LogEntry
}

func (r *mockStoreReader) Iterations() ([]store.IterationSummary, error) {
	return r.iterations, nil
}

func (r *mockStoreReader) IterationLog(_ int) ([]loop.LogEntry, error) {
	return r.entries, nil
}

func (r *mockStoreReader) SessionSummary() (store.SessionSummary, error) {
	return store.SessionSummary{}, nil
}

// mockLoopController is a test double for LoopController.
type mockLoopController struct {
	startCalled string
	stopCalled  bool
	running     bool
}

func (c *mockLoopController) StartLoop(mode string) {
	c.startCalled = mode
	c.running = true
}

func (c *mockLoopController) StopLoop() {
	c.stopCalled = true
	c.running = false
}

func (c *mockLoopController) IsRunning() bool { return c.running }

func newTestModel() Model {
	ch := make(chan loop.LogEntry, 1)
	return New(ch, nil, "", "TestProject", "/tmp/proj", nil, nil, nil)
}

func TestNew_Defaults(t *testing.T) {
	m := newTestModel()
	if m.width != 80 {
		t.Errorf("expected default width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("expected default height 24, got %d", m.height)
	}
	if m.focus != FocusMain {
		t.Errorf("expected default focus FocusMain, got %v", m.focus)
	}
	if m.loopState != StateIdle {
		t.Errorf("expected initial loopState StateIdle, got %v", m.loopState)
	}
	if m.done {
		t.Error("expected done=false at init")
	}
}

func TestInit_ReturnsCmd(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil command")
	}
}

func TestErr_NoError(t *testing.T) {
	m := newTestModel()
	if m.Err() != nil {
		t.Errorf("Err() should be nil at init, got %v", m.Err())
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	m := newTestModel()
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Error("WindowSizeMsg should return nil cmd")
	}
	m2 := updated.(Model)
	if m2.width != 120 || m2.height != 40 {
		t.Errorf("got dimensions %dx%d, want 120x40", m2.width, m2.height)
	}
	if m2.layout.TooSmall {
		t.Error("120x40 should not be TooSmall")
	}
}

func TestUpdate_WindowSize_TooSmall(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m2 := updated.(Model)
	if !m2.layout.TooSmall {
		t.Error("60x20 should be TooSmall")
	}
}

func TestUpdate_Key_Quit(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q key should return a quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("q key cmd should produce tea.QuitMsg, got %T", msg)
	}
}

func TestUpdate_Key_Tab_CyclesFocus(t *testing.T) {
	m := newTestModel()
	// Start at FocusMain (index 2), tab should go to FocusSecondary (3)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(Model)
	if m2.focus != FocusMain.Next() {
		t.Errorf("tab should advance focus from %v to %v, got %v",
			FocusMain, FocusMain.Next(), m2.focus)
	}
}

func TestUpdate_Key_DirectFocus(t *testing.T) {
	tests := []struct {
		key  string
		want FocusTarget
	}{
		{"1", FocusSpecs},
		{"2", FocusIterations},
		{"3", FocusMain},
		{"4", FocusSecondary},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := newTestModel()
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			m2 := updated.(Model)
			if m2.focus != tt.want {
				t.Errorf("key %q: focus = %v, want %v", tt.key, m2.focus, tt.want)
			}
		})
	}
}

func TestUpdate_LogEntry_StateTransition(t *testing.T) {
	tests := []struct {
		name      string
		entry     loop.LogEntry
		wantState LoopState
	}{
		{
			name:      "LogIterStart build → StateBuilding",
			entry:     loop.LogEntry{Kind: loop.LogIterStart, Mode: "build", Iteration: 1},
			wantState: StateBuilding,
		},
		{
			name:      "LogIterStart plan → StatePlanning",
			entry:     loop.LogEntry{Kind: loop.LogIterStart, Mode: "plan", Iteration: 1},
			wantState: StatePlanning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			msg := logEntryMsg(tt.entry)
			updated, _ := m.Update(msg)
			m2 := updated.(Model)
			if m2.loopState != tt.wantState {
				t.Errorf("loopState = %v, want %v", m2.loopState, tt.wantState)
			}
		})
	}
}

func TestUpdate_LogEntry_MetadataExtracted(t *testing.T) {
	m := newTestModel()
	entry := loop.LogEntry{
		Kind:      loop.LogIterStart,
		Timestamp: time.Now(),
		Iteration: 3,
		MaxIter:   10,
		Branch:    "feat/test",
		Mode:      "build",
		TotalCost: 0.05,
	}
	updated, _ := m.Update(logEntryMsg(entry))
	m2 := updated.(Model)

	if m2.iteration != 3 {
		t.Errorf("iteration = %d, want 3", m2.iteration)
	}
	if m2.branch != "feat/test" {
		t.Errorf("branch = %q, want \"feat/test\"", m2.branch)
	}
	if m2.mode != "build" {
		t.Errorf("mode = %q, want \"build\"", m2.mode)
	}
}

func TestUpdate_LoopDone_TransitionsToIdle(t *testing.T) {
	m := newTestModel()
	// Put model in building state to verify transition.
	m.loopState = StateBuilding
	updated, cmd := m.Update(loopDoneMsg{})
	m2 := updated.(Model)
	if m2.done {
		t.Error("loopDoneMsg should NOT set done=true; TUI stays open after loop finishes")
	}
	if cmd != nil {
		t.Error("loopDoneMsg should not return a quit cmd; user presses q to exit")
	}
	if m2.loopState != StateIdle {
		t.Errorf("loopDoneMsg should transition to StateIdle, got %v", m2.loopState)
	}
}

func TestUpdate_LoopErr_SetsErr(t *testing.T) {
	m := newTestModel()
	testErr := fmt.Errorf("loop failure")
	updated, _ := m.Update(loopErrMsg{err: testErr})
	m2 := updated.(Model)
	if m2.Err() != testErr {
		t.Errorf("Err() = %v, want %v", m2.Err(), testErr)
	}
	if !m2.done {
		t.Error("loopErrMsg should set done=true")
	}
}

func TestView_TooSmall(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m2 := updated.(Model)
	view := m2.View()
	lower := strings.ToLower(view)
	if !strings.Contains(lower, "too small") {
		t.Errorf("View() for small terminal should contain 'too small', got: %q", view)
	}
}

func TestView_Normal_DoesNotPanic(t *testing.T) {
	m := newTestModel()
	// Resize to large terminal so all panels have space
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := updated.(Model)
	// Should not panic
	view := m2.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestUpdate_EditSpecRequest_NoEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	m := newTestModel()
	_, cmd := m.Update(panels.EditSpecRequestMsg{Path: "specs/foo.md"})
	if cmd != nil {
		t.Error("EditSpecRequestMsg with no EDITOR should return nil cmd")
	}
}

func TestUpdate_EditSpecRequest_WithEditor(t *testing.T) {
	t.Setenv("EDITOR", "true")
	m := newTestModel()
	_, cmd := m.Update(panels.EditSpecRequestMsg{Path: "specs/foo.md"})
	if cmd == nil {
		t.Error("EditSpecRequestMsg with EDITOR set should return a non-nil cmd")
	}
}

func TestUpdate_CreateSpecRequest_ReturnsRefresh(t *testing.T) {
	dir := t.TempDir()
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "TestProject", dir, nil, nil, nil)

	_, cmd := m.Update(panels.CreateSpecRequestMsg{Name: "my-spec"})
	if cmd == nil {
		t.Fatal("CreateSpecRequestMsg should return a cmd")
	}
	msg := cmd()
	refreshed, ok := msg.(specsRefreshedMsg)
	if !ok {
		t.Fatalf("expected specsRefreshedMsg, got %T", msg)
	}
	// The spec file should have been created.
	found := false
	for _, s := range refreshed.Specs {
		if s.Name == "my-spec" {
			found = true
		}
	}
	if !found {
		t.Errorf("refreshed spec list should contain 'my-spec'; got %v", refreshed.Specs)
	}
	// File should exist on disk.
	if _, err := filepath.Glob(filepath.Join(dir, "specs", "my-spec.md")); err != nil {
		t.Errorf("spec file not created: %v", err)
	}
}

func TestUpdate_SpecsRefreshed_UpdatesPanel(t *testing.T) {
	m := newTestModel()
	// Panel starts with no specs.
	if m.specsPanel.SelectedSpec() != nil {
		t.Fatal("setup: expected empty specs panel")
	}
	newSpecs := []spec.SpecFile{
		{Name: "alpha", Path: "specs/alpha.md"},
	}
	updated, _ := m.Update(specsRefreshedMsg{Specs: newSpecs})
	m2 := updated.(Model)
	sel := m2.specsPanel.SelectedSpec()
	if sel == nil {
		t.Fatal("after specsRefreshedMsg, panel should have a selected spec")
	}
	if sel.Name != "alpha" {
		t.Errorf("selected spec name = %q, want %q", sel.Name, "alpha")
	}
}

func TestUpdate_StopRequested(t *testing.T) {
	stopped := false
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "", "", nil, func() { stopped = true }, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m2 := updated.(Model)
	if !stopped {
		t.Error("'s' key should have called requestStop")
	}
	if !m2.stopRequested {
		t.Error("stopRequested should be true after 's'")
	}

	// Second 's' press should not call requestStop again
	stopped = false
	m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if stopped {
		t.Error("second 's' key should not call requestStop again")
	}
}

func TestUpdate_IterationLogLoaded_SetsIterationAndSummary(t *testing.T) {
	m := newTestModel()

	// Resize to a standard size so layout is not TooSmall.
	updated0, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated0.(Model)

	summary := store.IterationSummary{
		Number:   2,
		Mode:     "build",
		CostUSD:  0.0234,
		Duration: 45.2,
		Subtype:  "success",
		Commit:   "abc1234",
	}
	msg := iterationLogLoadedMsg{
		Number:  2,
		Entries: []loop.LogEntry{{Kind: loop.LogInfo, Message: "agent output"}},
		Summary: summary,
		Err:     nil,
	}
	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	// After iterationLogLoadedMsg the main view is on TabIterationDetail (2).
	// One ] advances to TabIterationSummary (3).
	updated1, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	m3 := updated1.(Model)

	view := m3.View()
	// The Summary tab should show iteration number, cost, and commit.
	for _, want := range []string{"2", "build", "0.0234", "success", "abc1234"} {
		if !strings.Contains(view, want) {
			t.Errorf("Summary tab view missing %q; got:\n%s", want, view)
		}
	}
}

func TestHandleIterationSelected_NilReader(t *testing.T) {
	m := newTestModel() // storeReader is nil
	_, cmd := m.Update(panels.IterationSelectedMsg{Number: 1})
	if cmd != nil {
		t.Error("nil storeReader should return nil cmd")
	}
}

func TestHandleIterationSelected_WithReader(t *testing.T) {
	reader := &mockStoreReader{
		iterations: []store.IterationSummary{
			{Number: 1, Mode: "build", CostUSD: 0.01, Duration: 10, Subtype: "success", Commit: "deadbeef"},
		},
		entries: []loop.LogEntry{
			{Kind: loop.LogInfo, Message: "hello"},
		},
	}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, reader, "", "Proj", "/tmp", nil, nil, nil)

	_, cmd := m.Update(panels.IterationSelectedMsg{Number: 1})
	if cmd == nil {
		t.Fatal("with storeReader set, should return a non-nil cmd")
	}
	resultMsg := cmd()
	loaded, ok := resultMsg.(iterationLogLoadedMsg)
	if !ok {
		t.Fatalf("expected iterationLogLoadedMsg, got %T", resultMsg)
	}
	if loaded.Err != nil {
		t.Errorf("expected no error, got %v", loaded.Err)
	}
	if loaded.Summary.Commit != "deadbeef" {
		t.Errorf("expected commit deadbeef, got %q", loaded.Summary.Commit)
	}
}

func TestUpdate_Key_LoopControl_NoController(t *testing.T) {
	// Without a controller, b/p/R/x should be no-ops.
	keys := []string{"b", "p", "R", "x"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			m := newTestModel()
			updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			m2 := updated.(Model)
			if m2.loopState != StateIdle {
				t.Errorf("key %q without controller should not change state, got %v", key, m2.loopState)
			}
			if cmd != nil {
				t.Errorf("key %q without controller should return nil cmd", key)
			}
		})
	}
}

func TestUpdate_Key_Build_StartsLoop(t *testing.T) {
	ctrl := &mockLoopController{}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "Proj", "", nil, nil, ctrl)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m2 := updated.(Model)

	if ctrl.startCalled != "build" {
		t.Errorf("expected StartLoop(\"build\"), got %q", ctrl.startCalled)
	}
	if m2.loopState != StateBuilding {
		t.Errorf("expected StateBuilding, got %v", m2.loopState)
	}
}

func TestUpdate_Key_Plan_StartsLoop(t *testing.T) {
	ctrl := &mockLoopController{}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "Proj", "", nil, nil, ctrl)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m2 := updated.(Model)

	if ctrl.startCalled != "plan" {
		t.Errorf("expected StartLoop(\"plan\"), got %q", ctrl.startCalled)
	}
	if m2.loopState != StatePlanning {
		t.Errorf("expected StatePlanning, got %v", m2.loopState)
	}
}

func TestUpdate_Key_SmartRun_StartsLoop(t *testing.T) {
	ctrl := &mockLoopController{}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "Proj", "", nil, nil, ctrl)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	m2 := updated.(Model)

	if ctrl.startCalled != "smart" {
		t.Errorf("expected StartLoop(\"smart\"), got %q", ctrl.startCalled)
	}
	if m2.loopState != StateBuilding {
		t.Errorf("expected StateBuilding, got %v", m2.loopState)
	}
}

func TestUpdate_Key_Stop_StopsLoop(t *testing.T) {
	ctrl := &mockLoopController{running: true}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "Proj", "", nil, nil, ctrl)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if !ctrl.stopCalled {
		t.Error("x key should call StopLoop()")
	}
}

func TestUpdate_Key_Build_NoopWhenRunning(t *testing.T) {
	ctrl := &mockLoopController{running: true}
	ch := make(chan loop.LogEntry, 1)
	m := New(ch, nil, "", "Proj", "", nil, nil, ctrl)
	m.loopState = StateBuilding

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	if ctrl.startCalled != "" {
		t.Errorf("b key should not start loop when already running, startCalled=%q", ctrl.startCalled)
	}
}
