package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
)

// executeLoop loads config, builds the loop, and runs it in the given mode.
func executeLoop(mode loop.Mode, maxOverride int, noTUI bool) error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation: %w", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	ctx, cancel := signalContext()
	defer cancel()

	gitRunner := git.NewRunner(dir)
	lp := &loop.Loop{
		Agent:  loop.NewClaudeAgent(),
		Git:    gitRunner,
		Config: cfg,
		Dir:    dir,
	}

	runFn := func(ctx context.Context) error {
		return lp.Run(ctx, mode, maxOverride)
	}

	if !cfg.Regent.Enabled {
		if noTUI {
			return runWithStateTracking(ctx, lp, dir, gitRunner, string(mode), runFn)
		}
		return runWithTUIAndState(ctx, lp, dir, gitRunner, string(mode), cfg.TUI.AccentColor, runFn)
	}

	if noTUI {
		return runWithRegent(ctx, lp, cfg, gitRunner, dir, runFn)
	}
	return runWithRegentTUI(ctx, lp, cfg, gitRunner, dir, runFn)
}

// executeSmartRun runs plan if IMPLEMENTATION_PLAN.md doesn't exist, then build.
func executeSmartRun(maxOverride int, noTUI bool) error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation: %w", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	ctx, cancel := signalContext()
	defer cancel()

	gitRunner := git.NewRunner(dir)
	lp := &loop.Loop{
		Agent:  loop.NewClaudeAgent(),
		Git:    gitRunner,
		Config: cfg,
		Dir:    dir,
	}

	smartRunFn := func(ctx context.Context) error {
		// Check inside the closure so Regent retries re-evaluate whether
		// the plan file exists (it may have been created by a prior attempt).
		planPath := filepath.Join(dir, "IMPLEMENTATION_PLAN.md")
		info, statErr := os.Stat(planPath)
		if needsPlanPhase(info, statErr) {
			if planErr := lp.Run(ctx, loop.ModePlan, 0); planErr != nil {
				return fmt.Errorf("plan phase: %w", planErr)
			}
		}
		return lp.Run(ctx, loop.ModeBuild, maxOverride)
	}

	if !cfg.Regent.Enabled {
		if noTUI {
			return runWithStateTracking(ctx, lp, dir, gitRunner, "run", smartRunFn)
		}
		return runWithTUIAndState(ctx, lp, dir, gitRunner, "run", cfg.TUI.AccentColor, smartRunFn)
	}

	if noTUI {
		return runWithRegent(ctx, lp, cfg, gitRunner, dir, smartRunFn)
	}
	return runWithRegentTUI(ctx, lp, cfg, gitRunner, dir, smartRunFn)
}

// showStatus reads .ralph/regent-state.json and prints a formatted summary.
// Per ralph-core.md: branch, last commit, iteration count, total cost, duration, pass/fail.
func showStatus() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	state, err := regent.LoadState(dir)
	if err != nil {
		return err
	}

	result := classifyResult(state)
	if result == statusNoState {
		fmt.Println("No state found. Run 'ralph build' or 'ralph run' first.")
		return nil
	}

	fmt.Println("Ralph Status")
	fmt.Println("────────────")

	if state.Branch != "" {
		fmt.Printf("  %-20s %s\n", "Branch:", state.Branch)
	}
	if state.Mode != "" {
		fmt.Printf("  %-20s %s\n", "Mode:", state.Mode)
	}
	if state.LastCommit != "" {
		fmt.Printf("  %-20s %s\n", "Last commit:", state.LastCommit)
	}
	fmt.Printf("  %-20s %d\n", "Iteration:", state.Iteration)
	fmt.Printf("  %-20s $%.2f\n", "Total cost:", state.TotalCostUSD)

	if result == statusRunning {
		elapsed := time.Since(state.StartedAt).Round(time.Second)
		fmt.Printf("  %-20s %s (running)\n", "Duration:", elapsed)
	} else if !state.StartedAt.IsZero() && !state.FinishedAt.IsZero() {
		dur := state.FinishedAt.Sub(state.StartedAt).Round(time.Second)
		fmt.Printf("  %-20s %s\n", "Duration:", dur)
	}

	if result == statusRunning && !state.LastOutputAt.IsZero() {
		ago := time.Since(state.LastOutputAt).Round(time.Second)
		fmt.Printf("  %-20s %s ago\n", "Last output:", ago)
	}

	switch result {
	case statusRunning:
		fmt.Printf("  %-20s %s\n", "Result:", "running")
	case statusPass:
		fmt.Printf("  %-20s %s\n", "Result:", "pass")
	case statusFailWithErrors:
		fmt.Printf("  %-20s fail (%d consecutive errors)\n", "Result:", state.ConsecutiveErrs)
	case statusFail:
		fmt.Printf("  %-20s %s\n", "Result:", "fail")
	}

	return nil
}

// statusResult represents the outcome classification for ralph status display.
type statusResult int

const (
	statusNoState statusResult = iota
	statusRunning
	statusPass
	statusFailWithErrors
	statusFail
)

// classifyResult determines the result label from a Regent state snapshot.
// The priority order is: no-state, running, pass, fail-with-errors, plain-fail.
func classifyResult(state regent.State) statusResult {
	if state.RalphPID == 0 && state.Iteration == 0 {
		return statusNoState
	}
	running := !state.StartedAt.IsZero() && state.FinishedAt.IsZero()
	switch {
	case running:
		return statusRunning
	case state.Passed:
		return statusPass
	case state.ConsecutiveErrs > 0:
		return statusFailWithErrors
	case !state.FinishedAt.IsZero():
		return statusFail
	default:
		return statusNoState
	}
}

// needsPlanPhase reports whether the plan phase should run based on the
// result of os.Stat on IMPLEMENTATION_PLAN.md. Returns true if the file
// does not exist or is empty.
func needsPlanPhase(info fs.FileInfo, statErr error) bool {
	return statErr != nil || info == nil || info.Size() == 0
}

// openEditor launches the given editor with the file path, connecting stdio.
func openEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
