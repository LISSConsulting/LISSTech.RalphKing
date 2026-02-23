// Package main is the entry point for the Ralph CLI.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/tui"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "ralph",
		Short:   "RalphKing — spec-driven AI coding loop",
		Version: version,
	}

	root.PersistentFlags().Bool("no-tui", false, "disable TUI, use plain text output")

	root.AddCommand(
		planCmd(),
		buildCmd(),
		runCmd(),
		statusCmd(),
		initCmd(),
		specCmd(),
	)

	return root
}

func planCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Run Claude in plan mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			max, _ := cmd.Flags().GetInt("max")
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			return executeLoop(loop.ModePlan, max, noTUI)
		},
	}
	cmd.Flags().Int("max", 0, "override max iterations (0 = use config)")
	return cmd
}

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run Claude in build mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			max, _ := cmd.Flags().GetInt("max")
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			return executeLoop(loop.ModeBuild, max, noTUI)
		},
	}
	cmd.Flags().Int("max", 0, "override max iterations (0 = use config)")
	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Smart mode: plan if needed, then build",
		RunE: func(cmd *cobra.Command, args []string) error {
			max, _ := cmd.Flags().GetInt("max")
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			return executeSmartRun(max, noTUI)
		},
	}
	cmd.Flags().Int("max", 0, "override max iterations (0 = use config)")
	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show last run summary from Regent state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStatus()
		},
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create ralph.toml in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			path, err := config.InitFile(dir)
			if err != nil {
				return err
			}
			fmt.Printf("Created %s\n", path)
			return nil
		},
	}
}

func specCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage spec files",
	}

	cmd.AddCommand(specListCmd(), specNewCmd())
	return cmd
}

func specListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all spec files with status",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			specs, err := spec.List(dir)
			if err != nil {
				return err
			}

			if len(specs) == 0 {
				fmt.Println("No specs found in specs/")
				return nil
			}

			fmt.Println("Specs")
			fmt.Println("─────")
			for _, s := range specs {
				fmt.Printf("  %s  %-30s  %s\n", s.Status.Symbol(), s.Path, s.Status)
			}
			return nil
		},
	}
}

func specNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new spec file from template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			path, err := spec.New(dir, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Created %s\n", path)

			editor := os.Getenv("EDITOR")
			if editor == "" {
				return nil
			}

			return openEditor(editor, path)
		},
	}
}

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
		return runWithTUI(lp, func() tea.Cmd {
			return tui.RunLoop(ctx, lp, mode, maxOverride)
		})
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
		return runWithTUI(lp, func() tea.Cmd {
			return tui.RunSmartLoop(ctx, lp, maxOverride, needsPlan)
		})
	}

	if noTUI {
		return runWithRegent(ctx, lp, cfg, gitRunner, dir, smartRunFn)
	}
	return runWithRegentTUI(ctx, lp, cfg, gitRunner, dir, smartRunFn)
}

// runWithTUI creates an event channel, wires it to the loop and TUI, and
// runs the bubbletea program without Regent supervision.
func runWithTUI(lp *loop.Loop, loopCmdFn func() tea.Cmd) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	model := tui.New(events)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Start the loop in a goroutine; close channel when done.
	loopCmd := loopCmdFn()
	go func() {
		defer close(events)
		loopCmd()
	}()

	return finishTUI(program)
}

// runWithRegent runs the loop under Regent supervision without TUI.
// Events are drained to stdout.
func runWithRegent(ctx context.Context, lp *loop.Loop, cfg *config.Config, gitRunner *git.Runner, dir string, run regent.RunFunc) error {
	events := make(chan loop.LogEntry, 128)
	lp.Events = events

	rgt := regent.New(cfg.Regent, dir, gitRunner, events)

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

// showStatus reads .ralph/regent-state.json and prints a summary.
func showStatus() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	statePath := filepath.Join(dir, ".ralph", "regent-state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No Regent state found. Run 'ralph build' or 'ralph run' first.")
			return nil
		}
		return fmt.Errorf("read state: %w", err)
	}

	var state map[string]any
	if jsonErr := json.Unmarshal(data, &state); jsonErr != nil {
		return fmt.Errorf("parse state: %w", jsonErr)
	}

	fmt.Println("Ralph Status")
	fmt.Println("────────────")
	for k, v := range state {
		fmt.Printf("  %-20s %v\n", k+":", v)
	}
	return nil
}

// signalContext returns a context that is cancelled on SIGINT or SIGTERM,
// and a cancel function for cleanup.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	return ctx, cancel
}

// openEditor launches the given editor with the file path, connecting stdio.
func openEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
