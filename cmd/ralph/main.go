// Package main is the entry point for the Ralph CLI.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
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
			return executeLoop(loop.ModePlan, max)
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
			return executeLoop(loop.ModeBuild, max)
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
			return executeSmartRun(max)
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
			fmt.Println("ralph spec list: not yet implemented")
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
			fmt.Printf("ralph spec new %q: not yet implemented\n", args[0])
			return nil
		},
	}
}

// executeLoop loads config, builds the loop, and runs it in the given mode.
func executeLoop(mode loop.Mode, maxOverride int) error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	ctx := signalContext()

	lp := &loop.Loop{
		Agent:  loop.NewClaudeAgent(),
		Git:    git.NewRunner(dir),
		Config: cfg,
		Log:    os.Stdout,
		Dir:    dir,
	}

	return lp.Run(ctx, mode, maxOverride)
}

// executeSmartRun runs plan if IMPLEMENTATION_PLAN.md doesn't exist, then build.
func executeSmartRun(maxOverride int) error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	ctx := signalContext()

	lp := &loop.Loop{
		Agent:  loop.NewClaudeAgent(),
		Git:    git.NewRunner(dir),
		Config: cfg,
		Log:    os.Stdout,
		Dir:    dir,
	}

	// Check if plan is needed
	planPath := filepath.Join(dir, "IMPLEMENTATION_PLAN.md")
	needsPlan := false
	info, err := os.Stat(planPath)
	if err != nil || info.Size() == 0 {
		needsPlan = true
	}

	if needsPlan {
		fmt.Println("No IMPLEMENTATION_PLAN.md found — running plan first")
		if planErr := lp.Run(ctx, loop.ModePlan, 0); planErr != nil {
			return fmt.Errorf("plan phase: %w", planErr)
		}
	}

	fmt.Println("Starting build phase")
	return lp.Run(ctx, loop.ModeBuild, maxOverride)
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

// signalContext returns a context that is cancelled on SIGINT or SIGTERM.
func signalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	return ctx
}
