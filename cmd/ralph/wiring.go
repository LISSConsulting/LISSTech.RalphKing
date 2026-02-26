package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui"
)

// runWithRegent runs the loop under Regent supervision without TUI.
// Events are drained to stdout.
func runWithRegent(ctx context.Context, lp *loop.Loop, cfg *config.Config, gitRunner *git.Runner, dir string, run regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	rgt := regent.New(cfg.Regent, dir, gitRunner, events)
	lp.PostIteration = rgt.RunPostIterationTests

	// Drain events to stdout and update regent state
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for entry := range events {
			if entry.Kind != loop.LogRegent {
				rgt.UpdateState(entry)
			}
			fmt.Fprintln(os.Stdout, formatLogLine(entry))
		}
	}()

	err := rgt.Supervise(ctx, run)
	close(events)
	<-drainDone
	return err
}

// runWithRegentTUI runs the loop under Regent supervision with TUI display.
// Loop events are forwarded through the Regent for state/hang tracking, then
// sent to the TUI. Regent messages are sent directly to the TUI channel.
func runWithRegentTUI(ctx context.Context, lp *loop.Loop, cfg *config.Config, gitRunner *git.Runner, dir string, run regent.RunFunc) error {
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

	model := tui.New(tuiEvents, cfg.TUI.AccentColor, cfg.Project.Name, dir, requestStop)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Forward loop events → regent state update → TUI
	forwardDone := make(chan struct{})
	go func() {
		defer close(forwardDone)
		for entry := range loopEvents {
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
func runWithStateTracking(ctx context.Context, lp *loop.Loop, dir string, gitRunner *git.Runner, mode string, run regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	st := newStateTracker(dir, mode, gitRunner)
	st.save()

	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for entry := range events {
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
func runWithTUIAndState(ctx context.Context, lp *loop.Loop, dir string, gitRunner *git.Runner, mode string, accentColor string, projectName string, run regent.RunFunc) error {
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

	model := tui.New(tuiEvents, accentColor, projectName, dir, requestStop)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Forward loop events → state tracking → TUI
	forwardDone := make(chan struct{})
	go func() {
		defer close(forwardDone)
		for entry := range loopEvents {
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
