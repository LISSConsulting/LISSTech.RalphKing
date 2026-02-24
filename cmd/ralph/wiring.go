package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui"
)

// runWithTUI creates an event channel, wires it to the loop and TUI, and
// runs the bubbletea program without Regent supervision.
func runWithTUI(ctx context.Context, lp *loop.Loop, runFn regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	model := tui.New(events)
	program := tea.NewProgram(model, tea.WithAltScreen())

	go func() {
		defer close(events)
		runFn(ctx)
	}()

	return finishTUI(program)
}

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
			ts := entry.Timestamp.Format("15:04:05")
			if entry.Kind == loop.LogRegent {
				fmt.Fprintf(os.Stdout, "[%s]  🛡️  Regent: %s\n", ts, entry.Message)
			} else {
				fmt.Fprintf(os.Stdout, "[%s]  %s\n", ts, entry.Message)
			}
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

	lp.Events = loopEvents
	rgt := regent.New(cfg.Regent, dir, gitRunner, tuiEvents)
	lp.PostIteration = rgt.RunPostIterationTests

	model := tui.New(tuiEvents)
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
	go func() {
		defer close(tuiEvents)
		rgt.Supervise(ctx, run)
		close(loopEvents)
		<-forwardDone
	}()

	return finishTUI(program)
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
