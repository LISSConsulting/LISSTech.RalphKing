package main

import (
	"context"
	"fmt"
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
			lp.Log = os.Stdout
			return runFn(ctx)
		}
		return runWithTUI(ctx, lp, runFn)
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

	// Check if plan is needed
	planPath := filepath.Join(dir, "IMPLEMENTATION_PLAN.md")
	needsPlan := false
	info, statErr := os.Stat(planPath)
	if statErr != nil || info.Size() == 0 {
		needsPlan = true
	}

	smartRunFn := func(ctx context.Context) error {
		if needsPlan {
			if planErr := lp.Run(ctx, loop.ModePlan, 0); planErr != nil {
				return fmt.Errorf("plan phase: %w", planErr)
			}
		}
		return lp.Run(ctx, loop.ModeBuild, maxOverride)
	}

	if !cfg.Regent.Enabled {
		if noTUI {
			lp.Log = os.Stdout
			return smartRunFn(ctx)
		}
		return runWithTUI(ctx, lp, smartRunFn)
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

	if state.RalphPID == 0 && state.Iteration == 0 {
		fmt.Println("No Regent state found. Run 'ralph build' or 'ralph run' first.")
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

	running := !state.StartedAt.IsZero() && state.FinishedAt.IsZero()

	if running {
		elapsed := time.Since(state.StartedAt).Round(time.Second)
		fmt.Printf("  %-20s %s (running)\n", "Duration:", elapsed)
	} else if !state.StartedAt.IsZero() && !state.FinishedAt.IsZero() {
		dur := state.FinishedAt.Sub(state.StartedAt).Round(time.Second)
		fmt.Printf("  %-20s %s\n", "Duration:", dur)
	}

	if running && !state.LastOutputAt.IsZero() {
		ago := time.Since(state.LastOutputAt).Round(time.Second)
		fmt.Printf("  %-20s %s ago\n", "Last output:", ago)
	}

	if running {
		fmt.Printf("  %-20s %s\n", "Result:", "running")
	} else if state.Passed {
		fmt.Printf("  %-20s %s\n", "Result:", "pass")
	} else if state.ConsecutiveErrs > 0 {
		fmt.Printf("  %-20s fail (%d consecutive errors)\n", "Result:", state.ConsecutiveErrs)
	}

	return nil
}

// openEditor launches the given editor with the file path, connecting stdio.
func openEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
