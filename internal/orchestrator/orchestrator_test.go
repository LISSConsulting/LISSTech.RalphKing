package orchestrator

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/worktree"
)

// fakeWorktreeOps is a test double for WorktreeOps.
type fakeWorktreeOps struct {
	switchPath string
	switchErr  error
	mergeErr   error
	removeErr  error
	listResult []worktree.WorktreeInfo
}

func (f *fakeWorktreeOps) Detect() error                                    { return nil }
func (f *fakeWorktreeOps) Switch(_ string, _ bool) (string, error)          { return f.switchPath, f.switchErr }
func (f *fakeWorktreeOps) List() ([]worktree.WorktreeInfo, error)           { return f.listResult, nil }
func (f *fakeWorktreeOps) Merge(_, _ string) error                          { return f.mergeErr }
func (f *fakeWorktreeOps) Remove(_ string) error                            { return f.removeErr }

func defaultCfg() *config.Config {
	cfg := config.Defaults()
	cfg.Worktree.MaxParallel = 2
	return &cfg
}

func newTestOrchestrator(ops worktree.WorktreeOps) *Orchestrator {
	return New(defaultCfg(), ops)
}

// ─── ActiveAgents / AgentByBranch / RunningCount ─────────────────────────────

func TestActiveAgents_FiltersRemoved(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})

	// Manually insert agents in various states.
	o.agents["running"] = &WorktreeAgent{Branch: "running", State: StateRunning}
	o.agents["completed"] = &WorktreeAgent{Branch: "completed", State: StateCompleted}
	o.agents["removed"] = &WorktreeAgent{Branch: "removed", State: StateRemoved}

	active := o.ActiveAgents()
	if len(active) != 2 {
		t.Fatalf("expected 2 active agents, got %d", len(active))
	}
	for _, a := range active {
		if a.State == StateRemoved {
			t.Errorf("removed agent should not appear in ActiveAgents()")
		}
	}
}

func TestAgentByBranch(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["feat/x"] = &WorktreeAgent{Branch: "feat/x", State: StateRunning}

	got := o.AgentByBranch("feat/x")
	if got == nil || got.Branch != "feat/x" {
		t.Errorf("expected agent for feat/x, got %v", got)
	}
	if o.AgentByBranch("nonexistent") != nil {
		t.Error("expected nil for unknown branch")
	}
}

func TestRunningCount(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["a"] = &WorktreeAgent{State: StateRunning}
	o.agents["b"] = &WorktreeAgent{State: StateRunning}
	o.agents["c"] = &WorktreeAgent{State: StateCompleted}

	if got := o.RunningCount(); got != 2 {
		t.Errorf("RunningCount: got %d, want 2", got)
	}
}

// ─── Stop ─────────────────────────────────────────────────────────────────────

func TestStop_RunningAgent(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	stopCh := make(chan struct{})
	o.agents["feat/stop"] = &WorktreeAgent{Branch: "feat/stop", State: StateRunning, StopCh: stopCh}

	if err := o.Stop("feat/stop"); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if o.agents["feat/stop"].State != StateStopped {
		t.Errorf("expected StateStopped, got %v", o.agents["feat/stop"].State)
	}
	// StopCh should be closed.
	select {
	case <-stopCh:
	case <-time.After(100 * time.Millisecond):
		t.Error("StopCh was not closed")
	}
}

func TestStop_NotRunning(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["feat/done"] = &WorktreeAgent{Branch: "feat/done", State: StateCompleted}

	err := o.Stop("feat/done")
	if err == nil {
		t.Fatal("expected error stopping non-running agent")
	}
}

func TestStop_NoAgent(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	if err := o.Stop("nonexistent"); err == nil {
		t.Fatal("expected error for unknown branch")
	}
}

// ─── Merge ────────────────────────────────────────────────────────────────────

func TestMerge_CompletedAgent(t *testing.T) {
	ops := &fakeWorktreeOps{switchPath: "/tmp/wt"}
	o := newTestOrchestrator(ops)
	o.agents["feat/m"] = &WorktreeAgent{Branch: "feat/m", State: StateCompleted, WorktreePath: "/tmp/wt"}

	if err := o.Merge("feat/m"); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if o.agents["feat/m"].State != StateMerged {
		t.Errorf("expected StateMerged, got %v", o.agents["feat/m"].State)
	}
}

func TestMerge_RunningRejected(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["feat/r"] = &WorktreeAgent{Branch: "feat/r", State: StateRunning}

	err := o.Merge("feat/r")
	if err == nil {
		t.Fatal("expected error merging running agent")
	}
}

func TestMerge_ConflictLeavesMergeFailed(t *testing.T) {
	ops := &fakeWorktreeOps{switchPath: "/tmp/wt", mergeErr: errors.New("conflict")}
	o := newTestOrchestrator(ops)
	o.agents["feat/conflict"] = &WorktreeAgent{Branch: "feat/conflict", State: StateCompleted}

	err := o.Merge("feat/conflict")
	if err == nil {
		t.Fatal("expected error from merge conflict")
	}
	if o.agents["feat/conflict"].State != StateMergeFailed {
		t.Errorf("expected StateMergeFailed, got %v", o.agents["feat/conflict"].State)
	}
}

// ─── Clean ────────────────────────────────────────────────────────────────────

func TestClean_CompletedAgent(t *testing.T) {
	ops := &fakeWorktreeOps{switchPath: "/tmp/wt"}
	o := newTestOrchestrator(ops)
	o.agents["feat/done"] = &WorktreeAgent{Branch: "feat/done", State: StateCompleted}

	if err := o.Clean("feat/done"); err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if o.agents["feat/done"].State != StateRemoved {
		t.Errorf("expected StateRemoved, got %v", o.agents["feat/done"].State)
	}
}

func TestClean_RunningRejected(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["feat/busy"] = &WorktreeAgent{Branch: "feat/busy", State: StateRunning}

	if err := o.Clean("feat/busy"); err == nil {
		t.Fatal("expected error cleaning running agent")
	}
}

// ─── Launch ───────────────────────────────────────────────────────────────────

func TestLaunch_MaxParallelRejected(t *testing.T) {
	ops := &fakeWorktreeOps{switchPath: "/tmp/wt"}
	o := newTestOrchestrator(ops)
	// Pre-fill with running agents up to MaxParallel.
	o.agents["a"] = &WorktreeAgent{State: StateRunning}
	o.agents["b"] = &WorktreeAgent{State: StateRunning}

	ctx := context.Background()
	err := o.Launch(ctx, "feat/new", "", "", loop.ModeBuild, 0)
	if err == nil {
		t.Fatal("expected error when at max parallel")
	}
	if o.AgentByBranch("feat/new") != nil {
		t.Error("agent should not be created when at max parallel")
	}
}

func TestLaunch_DuplicateBranchRejected(t *testing.T) {
	ops := &fakeWorktreeOps{switchPath: "/tmp/wt"}
	o := newTestOrchestrator(ops)
	o.agents["feat/dupe"] = &WorktreeAgent{Branch: "feat/dupe", State: StateRunning}

	ctx := context.Background()
	err := o.Launch(ctx, "feat/dupe", "", "", loop.ModeBuild, 0)
	if err == nil {
		t.Fatal("expected error for duplicate running branch")
	}
}

func TestLaunch_SwitchError(t *testing.T) {
	ops := &fakeWorktreeOps{switchErr: errors.New("wt switch failed")}
	o := newTestOrchestrator(ops)

	ctx := context.Background()
	err := o.Launch(ctx, "feat/err", "", "", loop.ModeBuild, 0)
	if err == nil {
		t.Fatal("expected error when wt switch fails")
	}
	agent := o.AgentByBranch("feat/err")
	if agent == nil {
		t.Fatal("agent should be present after failed launch")
	}
	if agent.State != StateFailed {
		t.Errorf("expected StateFailed, got %v", agent.State)
	}
}

// ─── Fan-in ───────────────────────────────────────────────────────────────────

func TestFanIn_EventsTaggedCorrectly(t *testing.T) {
	merged := make(chan TaggedLogEntry, 16)
	events := make(chan loop.LogEntry, 4)
	var wg sync.WaitGroup

	startFanIn("feat/fan", events, merged, &wg)

	events <- loop.LogEntry{Kind: loop.LogInfo, Message: "hello from fan-in"}
	close(events)

	wg.Wait()
	close(merged)

	var got []TaggedLogEntry
	for e := range merged {
		got = append(got, e)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 tagged entry, got %d", len(got))
	}
	if got[0].Branch != "feat/fan" {
		t.Errorf("branch: got %q, want %q", got[0].Branch, "feat/fan")
	}
	if got[0].Entry.Message != "hello from fan-in" {
		t.Errorf("message: got %q", got[0].Entry.Message)
	}
}

func TestFanIn_ChannelCloseHandled(t *testing.T) {
	merged := make(chan TaggedLogEntry, 16)
	events := make(chan loop.LogEntry)
	var wg sync.WaitGroup

	startFanIn("feat/close", events, merged, &wg)
	close(events) // goroutine should exit cleanly
	wg.Wait()
}

// ─── WorktreePaths ────────────────────────────────────────────────────────────

func TestWorktreePaths(t *testing.T) {
	o := newTestOrchestrator(&fakeWorktreeOps{switchPath: "/tmp/wt"})
	o.agents["a"] = &WorktreeAgent{WorktreePath: "/tmp/a", State: StateRunning}
	o.agents["b"] = &WorktreeAgent{WorktreePath: "/tmp/b", State: StateCompleted}
	o.agents["c"] = &WorktreeAgent{WorktreePath: "/tmp/c", State: StateRemoved}

	paths := o.WorktreePaths()
	if len(paths) != 2 {
		t.Errorf("expected 2 paths (non-removed), got %d: %v", len(paths), paths)
	}
	for _, p := range paths {
		if p == "/tmp/c" {
			t.Errorf("removed agent path should not appear: %v", paths)
		}
	}
}
