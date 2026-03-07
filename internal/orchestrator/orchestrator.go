package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/worktree"
)

const mergedEventsBuf = 256

// Orchestrator manages the lifecycle of multiple WorktreeAgents.
type Orchestrator struct {
	mu          sync.Mutex
	agents      map[string]*WorktreeAgent // keyed by branch name
	fanInWg     sync.WaitGroup
	MaxParallel int
	AutoMerge   bool
	MergeTarget string
	WorktreeOps worktree.WorktreeOps
	cfg         *config.Config

	// MergedEvents receives tagged events from all agent fan-in goroutines.
	// Consumers (e.g. TUI) read from this channel.
	MergedEvents chan TaggedLogEntry
}

// New creates an Orchestrator with the given settings.
func New(cfg *config.Config, ops worktree.WorktreeOps) *Orchestrator {
	return &Orchestrator{
		agents:       make(map[string]*WorktreeAgent),
		MaxParallel:  cfg.Worktree.MaxParallel,
		AutoMerge:    cfg.Worktree.AutoMerge,
		MergeTarget:  cfg.Worktree.MergeTarget,
		WorktreeOps:  ops,
		cfg:          cfg,
		MergedEvents: make(chan TaggedLogEntry, mergedEventsBuf),
	}
}

// ActiveAgents returns a snapshot of agents that have not been removed.
func (o *Orchestrator) ActiveAgents() []*WorktreeAgent {
	o.mu.Lock()
	defer o.mu.Unlock()

	var result []*WorktreeAgent
	for _, a := range o.agents {
		if a.State != StateRemoved {
			result = append(result, a)
		}
	}
	return result
}

// RunningCount returns the number of agents currently in StateRunning.
func (o *Orchestrator) RunningCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()

	count := 0
	for _, a := range o.agents {
		if a.State == StateRunning {
			count++
		}
	}
	return count
}

// AgentByBranch returns the agent for the given branch, or nil if not found.
func (o *Orchestrator) AgentByBranch(branch string) *WorktreeAgent {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.agents[branch]
}

// Launch creates a worktree for branch, starts a loop inside it, and registers
// the agent with the fan-in multiplexer. The loop runs in a background
// goroutine.
//
// Returns an error if:
//   - max_parallel is already reached
//   - an agent for branch already exists and is not in a terminal state
//   - WorktreeOps.Switch() fails
func (o *Orchestrator) Launch(ctx context.Context, branch, specName, specDir string, mode loop.Mode, maxOverride int) error {
	o.mu.Lock()

	// Reject duplicate branches (non-terminal state).
	if existing, ok := o.agents[branch]; ok {
		switch existing.State {
		case StateRunning, StateCreating:
			o.mu.Unlock()
			return fmt.Errorf("orchestrator: agent already running on branch %s", branch)
		}
	}

	if o.runningCount() >= o.MaxParallel {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: max parallel agents (%d) reached", o.MaxParallel)
	}

	events := make(chan loop.LogEntry, 128)
	stopCh := make(chan struct{})

	agent := &WorktreeAgent{
		Branch:   branch,
		SpecName: specName,
		SpecDir:  specDir,
		State:    StateCreating,
		Events:   events,
		StopCh:   stopCh,
	}
	o.agents[branch] = agent
	o.mu.Unlock()

	// Create/switch worktree outside the lock (subprocess call).
	wtPath, err := o.WorktreeOps.Switch(branch, true)
	if err != nil {
		o.mu.Lock()
		agent.State = StateFailed
		agent.Error = err
		o.mu.Unlock()
		close(events)
		return fmt.Errorf("orchestrator: create worktree for %s: %w", branch, err)
	}

	o.mu.Lock()
	agent.WorktreePath = wtPath
	agent.State = StateRunning
	o.mu.Unlock()

	// Register with fan-in so events reach MergedEvents.
	startFanIn(branch, events, o.MergedEvents, &o.fanInWg)

	// Build the loop for this worktree.
	lp := &loop.Loop{
		Agent:     &loop.ClaudeAgent{Executable: "claude"},
		Git:       git.NewRunner(wtPath),
		Config:    o.cfg,
		Dir:       wtPath,
		Events:    events,
		Spec:      specName,
		SpecDir:   specDir,
		StopAfter: stopCh,
	}

	go func() {
		runErr := lp.Run(ctx, mode, maxOverride)

		o.mu.Lock()
		if runErr != nil && runErr != context.Canceled {
			agent.State = StateFailed
			agent.Error = runErr
		} else if agent.State == StateRunning {
			agent.State = StateCompleted
		}
		o.mu.Unlock()

		close(events)

		// Auto-merge if configured and agent completed successfully.
		if agent.State == StateCompleted && o.AutoMerge {
			_ = o.Merge(branch)
		}
	}()

	return nil
}

// Stop requests a graceful stop for the agent on branch by closing its StopCh.
func (o *Orchestrator) Stop(branch string) error {
	o.mu.Lock()
	agent, ok := o.agents[branch]
	if !ok {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: no agent for branch %s", branch)
	}
	if agent.State != StateRunning {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: agent %s is not running (state: %s)", branch, agent.State)
	}
	o.mu.Unlock()

	// Close stop channel; loop checks it after each iteration.
	func() {
		defer func() { _ = recover() }()
		close(agent.StopCh)
	}()

	o.mu.Lock()
	agent.State = StateStopped
	o.mu.Unlock()
	return nil
}

// StopAll stops every currently running agent.
func (o *Orchestrator) StopAll() {
	o.mu.Lock()
	branches := make([]string, 0, len(o.agents))
	for b, a := range o.agents {
		if a.State == StateRunning {
			branches = append(branches, b)
		}
	}
	o.mu.Unlock()

	for _, b := range branches {
		_ = o.Stop(b)
	}
}

// Merge merges a completed/stopped worktree branch into the merge target.
func (o *Orchestrator) Merge(branch string) error {
	o.mu.Lock()
	agent, ok := o.agents[branch]
	if !ok {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: no agent for branch %s", branch)
	}
	if agent.State == StateRunning || agent.State == StateCreating {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: cannot merge running agent %s — stop it first", branch)
	}
	agent.State = StateMerging
	o.mu.Unlock()

	if err := o.WorktreeOps.Merge(branch, o.MergeTarget); err != nil {
		o.mu.Lock()
		agent.State = StateMergeFailed
		agent.Error = err
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: merge %s: %w", branch, err)
	}

	o.mu.Lock()
	agent.State = StateMerged
	o.mu.Unlock()
	return nil
}

// Clean removes a non-running worktree agent via worktrunk.
func (o *Orchestrator) Clean(branch string) error {
	o.mu.Lock()
	agent, ok := o.agents[branch]
	if !ok {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: no agent for branch %s", branch)
	}
	if agent.State == StateRunning || agent.State == StateCreating {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator: cannot clean running agent %s — stop it first", branch)
	}
	o.mu.Unlock()

	if err := o.WorktreeOps.Remove(branch); err != nil {
		return fmt.Errorf("orchestrator: clean %s: %w", branch, err)
	}

	o.mu.Lock()
	agent.State = StateRemoved
	o.mu.Unlock()
	return nil
}

// WorktreePaths returns the working directory path of every non-removed agent.
func (o *Orchestrator) WorktreePaths() []string {
	o.mu.Lock()
	defer o.mu.Unlock()

	var paths []string
	for _, a := range o.agents {
		if a.State != StateRemoved && a.WorktreePath != "" {
			paths = append(paths, a.WorktreePath)
		}
	}
	return paths
}

// runningCount returns the count of StateRunning agents; callers must hold o.mu.
func (o *Orchestrator) runningCount() int {
	count := 0
	for _, a := range o.agents {
		if a.State == StateRunning {
			count++
		}
	}
	return count
}

