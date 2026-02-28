package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/notify"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
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

	// Pre-flight: verify the prompt file exists before launching TUI or Regent.
	// Without this check the TUI initialises, then fails on the first iteration
	// with a confusing "loop: read prompt …: open …: no such file or directory".
	var promptFile string
	switch mode {
	case loop.ModePlan:
		promptFile = cfg.Plan.PromptFile
	default:
		promptFile = cfg.Build.PromptFile
	}
	if _, statErr := os.Stat(filepath.Join(dir, promptFile)); statErr != nil {
		return fmt.Errorf("prompt file %s: %w", promptFile, statErr)
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
	if cfg.Notifications.URL != "" {
		n := notify.New(cfg.Notifications.URL, cfg.Project.Name,
			cfg.Notifications.OnComplete, cfg.Notifications.OnError, cfg.Notifications.OnStop)
		lp.NotificationHook = n.Hook
	}

	var sw store.Writer
	if s, err := store.NewJSONL(filepath.Join(dir, ".ralph", "logs")); err != nil {
		fmt.Fprintf(os.Stderr, "ralph: session log unavailable: %v\n", err)
	} else {
		sw = s
		defer s.Close()
	}

	runFn := func(ctx context.Context) error {
		return lp.Run(ctx, mode, maxOverride)
	}

	if !cfg.Regent.Enabled {
		if noTUI {
			return runWithStateTracking(ctx, lp, dir, gitRunner, string(mode), sw, runFn)
		}
		return runWithTUIAndState(ctx, lp, dir, gitRunner, string(mode), cfg.TUI.AccentColor, cfg.Project.Name, sw, runFn)
	}

	if noTUI {
		return runWithRegent(ctx, lp, cfg, gitRunner, dir, sw, runFn)
	}
	return runWithRegentTUI(ctx, lp, cfg, gitRunner, dir, sw, runFn)
}

// executeSmartRun runs plan if CHRONICLE.md doesn't exist, then build.
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
	if cfg.Notifications.URL != "" {
		n := notify.New(cfg.Notifications.URL, cfg.Project.Name,
			cfg.Notifications.OnComplete, cfg.Notifications.OnError, cfg.Notifications.OnStop)
		lp.NotificationHook = n.Hook
	}

	var sw store.Writer
	if s, err := store.NewJSONL(filepath.Join(dir, ".ralph", "logs")); err != nil {
		fmt.Fprintf(os.Stderr, "ralph: session log unavailable: %v\n", err)
	} else {
		sw = s
		defer s.Close()
	}

	smartRunFn := func(ctx context.Context) error {
		// Check inside the closure so Regent retries re-evaluate whether
		// the plan file exists (it may have been created by a prior attempt).
		planPath := filepath.Join(dir, "CHRONICLE.md")
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
			return runWithStateTracking(ctx, lp, dir, gitRunner, "run", sw, smartRunFn)
		}
		return runWithTUIAndState(ctx, lp, dir, gitRunner, "run", cfg.TUI.AccentColor, cfg.Project.Name, sw, smartRunFn)
	}

	if noTUI {
		return runWithRegent(ctx, lp, cfg, gitRunner, dir, sw, smartRunFn)
	}
	return runWithRegentTUI(ctx, lp, cfg, gitRunner, dir, sw, smartRunFn)
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

	fmt.Print(formatStatus(state, time.Now()))
	return nil
}

// formatStatus renders a Regent state snapshot as a human-readable status
// string. The now parameter pins the current time for deterministic output.
func formatStatus(state regent.State, now time.Time) string {
	result := classifyResult(state)
	if result == statusNoState {
		return "No state found. Run 'ralph build' or 'ralph run' first.\n"
	}

	var b strings.Builder
	b.WriteString("Ralph Status\n")
	b.WriteString("────────────\n")

	if state.Branch != "" {
		fmt.Fprintf(&b, "  %-20s %s\n", "Branch:", state.Branch)
	}
	if state.Mode != "" {
		fmt.Fprintf(&b, "  %-20s %s\n", "Mode:", state.Mode)
	}
	if state.LastCommit != "" {
		fmt.Fprintf(&b, "  %-20s %s\n", "Last commit:", state.LastCommit)
	}
	fmt.Fprintf(&b, "  %-20s %d\n", "Iteration:", state.Iteration)
	fmt.Fprintf(&b, "  %-20s $%.2f\n", "Total cost:", state.TotalCostUSD)

	if result == statusRunning {
		elapsed := now.Sub(state.StartedAt).Round(time.Second)
		fmt.Fprintf(&b, "  %-20s %s (running)\n", "Duration:", elapsed)
	} else if !state.StartedAt.IsZero() && !state.FinishedAt.IsZero() {
		dur := state.FinishedAt.Sub(state.StartedAt).Round(time.Second)
		fmt.Fprintf(&b, "  %-20s %s\n", "Duration:", dur)
	}

	if result == statusRunning && !state.LastOutputAt.IsZero() {
		ago := now.Sub(state.LastOutputAt).Round(time.Second)
		fmt.Fprintf(&b, "  %-20s %s ago\n", "Last output:", ago)
	}

	switch result {
	case statusRunning:
		fmt.Fprintf(&b, "  %-20s %s\n", "Result:", "running")
	case statusPass:
		fmt.Fprintf(&b, "  %-20s %s\n", "Result:", "pass")
	case statusFailWithErrors:
		fmt.Fprintf(&b, "  %-20s fail (%d consecutive errors)\n", "Result:", state.ConsecutiveErrs)
	case statusFail:
		fmt.Fprintf(&b, "  %-20s %s\n", "Result:", "fail")
	}

	return b.String()
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
// result of os.Stat on CHRONICLE.md. Returns true if the file
// does not exist or is empty.
func needsPlanPhase(info fs.FileInfo, statErr error) bool {
	return statErr != nil || info == nil || info.Size() == 0
}

// formatLogLine renders a log entry as a timestamped line for plain-text output.
// Regent entries get a shield prefix; all others display the message directly.
func formatLogLine(entry loop.LogEntry) string {
	ts := entry.Timestamp.Format("15:04:05")
	if entry.Kind == loop.LogRegent {
		return fmt.Sprintf("[%s]  🛡️  Regent: %s", ts, entry.Message)
	}
	return fmt.Sprintf("[%s]  %s", ts, entry.Message)
}

// openEditor launches the given editor with the file path, connecting stdio.
func openEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
