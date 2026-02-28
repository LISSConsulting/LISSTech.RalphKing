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
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui/panels"
)

func newTestModel() Model {
	ch := make(chan loop.LogEntry, 1)
	return New(ch, nil, "", "TestProject", "/tmp/proj", nil, nil)
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

func TestUpdate_LoopDone_QuitCmd(t *testing.T) {
	m := newTestModel()
	updated, cmd := m.Update(loopDoneMsg{})
	m2 := updated.(Model)
	if !m2.done {
		t.Error("loopDoneMsg should set done=true")
	}
	if cmd == nil {
		t.Fatal("loopDoneMsg should return a quit cmd")
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
	m := New(ch, nil, "", "TestProject", dir, nil, nil)

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
	m := New(ch, nil, "", "", "", nil, func() { stopped = true })

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
