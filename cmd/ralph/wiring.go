package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui"
)

// runWithRegent runs the loop under Regent supervision without TUI.
// Events are drained to stdout.
func runWithRegent(ctx context.Context, lp *loop.Loop, cfg *config.Config, gitRunner *git.Runner, dir string, sw store.Writer, run regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	rgt := regent.New(cfg.Regent, dir, gitRunner, events)
	lp.PostIteration = rgt.RunPostIterationTests

	// Drain events to stdout and update regent state
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for entry := range events {
			if sw != nil {
				_ = sw.Append(entry)
			}
			if entry.Kind != loop.LogRegent {
				rgt.UpdateState(entry)
			}
			fmt.Fprintln(os.Stdout, formatLogLine(entry))
		}
	}()

	err := rgt.Supervise(ctx, run)
	close(events)
	<-drainDone
	// Flush persists any UpdateState changes made by the drain goroutine that
	// may have raced with Supervise's own saveState call. After drainDone, all
	// UpdateState calls are complete and a single flush produces the correct
	// final state on disk.
	rgt.FlushState()
	return err
}

// runWithRegentTUI runs the loop under Regent supervision with TUI display.
// Loop events are forwarded through the Regent for state/hang tracking, then
// sent to the TUI. Regent messages are sent directly to the TUI channel.
func runWithRegentTUI(ctx context.Context, lp *loop.Loop, cfg *config.Config, gitRunner *git.Runner, dir string, sw store.Writer, sr store.Reader, run regent.RunFunc) error {
	loopEvents := make(chan loop.LogEntry, 128)
	tuiEvents := make(chan loop.LogEntry, 128)

	// Graceful stop: TUI 's' key closes stopCh; loop checks it after each iteration.
	stopCh := make(chan struct{})
	var stopOnce sync.Once
	requestStop := func() { stopOnce.Do(func() { close(stopCh) }) }
	lp.StopAfter = stopCh

	lp.Events = loopEvents
	rgt := regent.New(cfg.Regent, dir, gitRunner, tuiEvents)
	lp.PostIteration = rgt.RunPostIterationTests

	specFiles, _ := spec.List(dir)
	model := tui.New(tuiEvents, sr, cfg.TUI.AccentColor, cfg.Project.Name, dir, specFiles, requestStop, nil)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Forward loop events → regent state update → TUI
	forwardDone := make(chan struct{})
	go func() {
		defer close(forwardDone)
		for entry := range loopEvents {
			if sw != nil {
				_ = sw.Append(entry)
			}
			rgt.UpdateState(entry)
			select {
			case tuiEvents <- entry:
			default:
			}
		}
	}()

	// Run loop under Regent supervision; close channels when done
	errCh := make(chan error, 1)
	go func() {
		defer close(tuiEvents)
		superviseErr := rgt.Supervise(ctx, run)
		close(loopEvents)
		<-forwardDone
		errCh <- superviseErr
	}()

	tuiErr := finishTUI(program)
	if tuiErr != nil {
		return tuiErr
	}

	// Collect Regent/loop error if available.
	select {
	case loopErr := <-errCh:
		if loopErr != nil && !errors.Is(loopErr, context.Canceled) {
			return loopErr
		}
	default:
	}
	return nil
}

// finishTUI runs the bubbletea program and returns any loop error.
// Context cancellation errors are suppressed since they indicate normal
// shutdown (user quit, signal).
func finishTUI(program *tea.Program) error {
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	if m, ok := finalModel.(tui.Model); ok && m.Err() != nil {
		if errors.Is(m.Err(), context.Canceled) {
			return nil
		}
		return m.Err()
	}

	return nil
}

// runWithStateTracking runs the loop without Regent supervision in no-TUI mode,
// draining events to stdout and persisting state to .ralph/regent-state.json
// so that `ralph status` works even when the Regent is disabled.
func runWithStateTracking(ctx context.Context, lp *loop.Loop, dir string, gitRunner *git.Runner, mode string, sw store.Writer, run regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	st := newStateTracker(dir, mode, gitRunner)
	st.save()

	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for entry := range events {
			if sw != nil {
				_ = sw.Append(entry)
			}
			fmt.Fprintln(os.Stdout, formatLogLine(entry))
			st.trackEntry(entry)
		}
	}()

	runErr := run(ctx)
	close(events)
	<-drainDone

	st.finish(runErr)
	return runErr
}

// runWithTUIAndState runs the loop without Regent supervision with TUI display,
// forwarding events through a state tracker so `ralph status` works.
func runWithTUIAndState(ctx context.Context, lp *loop.Loop, dir string, gitRunner *git.Runner, mode string, accentColor string, projectName string, sw store.Writer, sr store.Reader, run regent.RunFunc) error {
	loopEvents := make(chan loop.LogEntry, 128)
	tuiEvents := make(chan loop.LogEntry, 128)

	// Graceful stop: TUI 's' key closes stopCh; loop checks it after each iteration.
	stopCh := make(chan struct{})
	var stopOnce sync.Once
	requestStop := func() { stopOnce.Do(func() { close(stopCh) }) }
	lp.StopAfter = stopCh

	lp.Events = loopEvents

	st := newStateTracker(dir, mode, gitRunner)
	st.save()

	specFiles, _ := spec.List(dir)
	model := tui.New(tuiEvents, sr, accentColor, projectName, dir, specFiles, requestStop, nil)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Forward loop events → state tracking → TUI
	forwardDone := make(chan struct{})
	go func() {
		defer close(forwardDone)
		for entry := range loopEvents {
			if sw != nil {
				_ = sw.Append(entry)
			}
			st.trackEntry(entry)
			select {
			case tuiEvents <- entry:
			default:
			}
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		defer close(tuiEvents)
		runErr := run(ctx)
		close(loopEvents)
		<-forwardDone
		errCh <- runErr
	}()

	tuiErr := finishTUI(program)
	if tuiErr != nil {
		return tuiErr
	}

	select {
	case loopErr := <-errCh:
		st.finish(loopErr)
		if loopErr != nil && !errors.Is(loopErr, context.Canceled) {
			return loopErr
		}
	default:
		st.finish(nil)
	}
	return nil
}

// stateTracker persists loop state to .ralph/regent-state.json for `ralph status`.
// Used in non-Regent paths where the Regent is not available to track state.
type stateTracker struct {
	state regent.State
	dir   string
}

func newStateTracker(dir, mode string, gitRunner *git.Runner) *stateTracker {
	branch, _ := gitRunner.CurrentBranch()
	now := time.Now()
	return &stateTracker{
		dir: dir,
		state: regent.State{
			RalphPID:     os.Getpid(),
			Branch:       branch,
			Mode:         mode,
			StartedAt:    now,
			LastOutputAt: now,
		},
	}
}

func (s *stateTracker) trackEntry(entry loop.LogEntry) {
	changed := false
	if entry.Iteration > 0 {
		s.state.Iteration = entry.Iteration
		changed = true
	}
	if entry.TotalCost > 0 {
		s.state.TotalCostUSD = entry.TotalCost
		changed = true
	}
	if entry.Commit != "" {
		s.state.LastCommit = entry.Commit
		changed = true
	}
	if entry.Branch != "" {
		s.state.Branch = entry.Branch
		changed = true
	}
	if entry.Mode != "" {
		s.state.Mode = entry.Mode
		changed = true
	}
	s.state.LastOutputAt = time.Now()
	if changed {
		s.save()
	}
}

func (s *stateTracker) save() {
	_ = regent.SaveState(s.dir, s.state)
}

func (s *stateTracker) finish(err error) {
	s.state.FinishedAt = time.Now()
	s.state.Passed = err == nil || errors.Is(err, context.Canceled)
	s.save()
}

// loopController implements tui.LoopController for dashboard mode.
// It starts and stops loop runs in response to TUI key presses (b/p/R/x).
type loopController struct {
	cfg       *config.Config
	dir       string
	gitRunner *git.Runner
	sw        store.Writer
	tuiSend   chan<- loop.LogEntry
	outerCtx  context.Context
	mu        sync.Mutex
	cancel    context.CancelFunc
}

// IsRunning reports whether a loop goroutine is currently active.
func (lc *loopController) IsRunning() bool {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.cancel != nil
}

// StartLoop starts a loop in the given mode ("build", "plan", or "smart").
// A no-op if a loop is already running.
func (lc *loopController) StartLoop(mode string) {
	lc.mu.Lock()
	if lc.cancel != nil {
		lc.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(lc.outerCtx)
	lc.cancel = cancel
	lc.mu.Unlock()

	go lc.runLoop(ctx, mode)
}

// StopLoop immediately cancels the running loop. No-op if idle.
func (lc *loopController) StopLoop() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.cancel != nil {
		lc.cancel()
	}
}

// runLoop executes the loop and forwards events to the TUI channel.
func (lc *loopController) runLoop(ctx context.Context, mode string) {
	lp := &loop.Loop{
		Agent:  loop.NewClaudeAgent(),
		Git:    lc.gitRunner,
		Config: lc.cfg,
		Dir:    lc.dir,
	}
	loopEvents := make(chan loop.LogEntry, 128)
	lp.Events = loopEvents

	forwardDone := make(chan struct{})
	go func() {
		defer close(forwardDone)
		for entry := range loopEvents {
			if lc.sw != nil {
				_ = lc.sw.Append(entry)
			}
			select {
			case lc.tuiSend <- entry:
			default:
			}
		}
	}()

	var runErr error
	switch mode {
	case "plan":
		runErr = lp.Run(ctx, loop.ModePlan, 0)
	case "smart":
		planPath := filepath.Join(lc.dir, "CHRONICLE.md")
		info, statErr := os.Stat(planPath)
		if needsPlanPhase(info, statErr) {
			runErr = lp.Run(ctx, loop.ModePlan, 0)
		}
		if runErr == nil {
			runErr = lp.Run(ctx, loop.ModeBuild, 0)
		}
	default: // "build"
		runErr = lp.Run(ctx, loop.ModeBuild, 0)
	}

	close(loopEvents)
	<-forwardDone

	lc.mu.Lock()
	lc.cancel = nil
	lc.mu.Unlock()

	_ = runErr
}

// runDashboard launches the TUI in idle (dashboard) state with no loop running.
// The user can press b/p/R to start a loop and x to stop it.
func runDashboard(ctx context.Context, cfg *config.Config, dir string, sw store.Writer, sr store.Reader) error {
	tuiEvents := make(chan loop.LogEntry, 128)
	// Note: tuiEvents is intentionally never closed; the TUI exits when user presses q.

	gitRunner := git.NewRunner(dir)
	ctrl := &loopController{
		cfg:       cfg,
		dir:       dir,
		gitRunner: gitRunner,
		sw:        sw,
		tuiSend:   tuiEvents,
		outerCtx:  ctx,
	}

	specFiles, _ := spec.List(dir)
	model := tui.New(tuiEvents, sr, cfg.TUI.AccentColor, cfg.Project.Name, dir, specFiles, nil, ctrl)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return finishTUI(program)
}
