package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/git"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

// resolveSpec resolves the active spec directory from --spec flag or git branch.
func resolveSpec(dir, specFlag string) (spec.ActiveSpec, error) {
	branch, _ := git.NewRunner(dir).CurrentBranch()
	return spec.Resolve(dir, specFlag, branch)
}

// specifyCmd creates a new spec feature via the speckit.specify skill.
// Usage: ralph specify <description> [--spec <name>]
func specifyCmd() *cobra.Command {
	var specFlag string
	cmd := &cobra.Command{
		Use:   "specify <description...>",
		Short: "Create or update a feature spec via speckit",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			var activeSpec spec.ActiveSpec
			if specFlag != "" {
				// Create the spec directory if it doesn't exist yet.
				specDir := filepath.Join(dir, "specs", specFlag)
				if mkErr := os.MkdirAll(specDir, 0o755); mkErr != nil {
					return fmt.Errorf("create spec directory: %w", mkErr)
				}
				activeSpec = spec.ActiveSpec{
					Name:     specFlag,
					Dir:      specDir,
					Explicit: true,
				}
			} else {
				activeSpec, err = resolveSpec(dir, "")
				if err != nil {
					return err
				}
			}

			ctx, cancel := signalContext()
			defer cancel()

			_ = activeSpec // spec directory created; claude operates in CWD
			return executeSpeckit(ctx, "speckit.specify", args)
		},
	}
	cmd.Flags().StringVar(&specFlag, "spec", "", "spec directory name (e.g. 004-my-feature)")
	return cmd
}

// speckitPlanCmd invokes the speckit.plan skill for the active spec.
// Requires spec.md to exist in the spec directory.
func speckitPlanCmd() *cobra.Command {
	var specFlag string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generate an implementation plan for the active spec via speckit",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			activeSpec, err := resolveSpec(dir, specFlag)
			if err != nil {
				return err
			}

			specFile := filepath.Join(activeSpec.Dir, "spec.md")
			if _, statErr := os.Stat(specFile); statErr != nil {
				return fmt.Errorf("spec.md not found in %s; run 'ralph specify' first", activeSpec.Dir)
			}

			ctx, cancel := signalContext()
			defer cancel()

			return executeSpeckit(ctx, "speckit.plan", nil)
		},
	}
	cmd.Flags().StringVar(&specFlag, "spec", "", "spec directory name (e.g. 004-my-feature)")
	return cmd
}

// clarifyCmd invokes the speckit.clarify skill for the active spec.
// Requires spec.md to exist in the spec directory.
func clarifyCmd() *cobra.Command {
	var specFlag string
	cmd := &cobra.Command{
		Use:   "clarify",
		Short: "Clarify requirements for the active spec via speckit",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			activeSpec, err := resolveSpec(dir, specFlag)
			if err != nil {
				return err
			}

			specFile := filepath.Join(activeSpec.Dir, "spec.md")
			if _, statErr := os.Stat(specFile); statErr != nil {
				return fmt.Errorf("spec.md not found in %s; run 'ralph specify' first", activeSpec.Dir)
			}

			ctx, cancel := signalContext()
			defer cancel()

			return executeSpeckit(ctx, "speckit.clarify", nil)
		},
	}
	cmd.Flags().StringVar(&specFlag, "spec", "", "spec directory name (e.g. 004-my-feature)")
	return cmd
}

// speckitTasksCmd invokes the speckit.tasks skill for the active spec.
// Requires plan.md to exist in the spec directory.
func speckitTasksCmd() *cobra.Command {
	var specFlag string
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Generate a task breakdown for the active spec via speckit",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			activeSpec, err := resolveSpec(dir, specFlag)
			if err != nil {
				return err
			}

			planFile := filepath.Join(activeSpec.Dir, "plan.md")
			if _, statErr := os.Stat(planFile); statErr != nil {
				return fmt.Errorf("plan.md not found in %s; run 'ralph plan' first", activeSpec.Dir)
			}

			ctx, cancel := signalContext()
			defer cancel()

			return executeSpeckit(ctx, "speckit.tasks", nil)
		},
	}
	cmd.Flags().StringVar(&specFlag, "spec", "", "spec directory name (e.g. 004-my-feature)")
	return cmd
}

// speckitRunCmd invokes the speckit.implement skill for the active spec.
// Requires tasks.md to exist in the spec directory.
func speckitRunCmd() *cobra.Command {
	var specFlag string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Implement the active spec via speckit (requires tasks.md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			activeSpec, err := resolveSpec(dir, specFlag)
			if err != nil {
				return err
			}

			tasksFile := filepath.Join(activeSpec.Dir, "tasks.md")
			if _, statErr := os.Stat(tasksFile); statErr != nil {
				return fmt.Errorf("tasks.md not found in %s; run 'ralph tasks' first", activeSpec.Dir)
			}

			ctx, cancel := signalContext()
			defer cancel()

			return executeSpeckit(ctx, "speckit.implement", nil)
		},
	}
	cmd.Flags().StringVar(&specFlag, "spec", "", "spec directory name (e.g. 004-my-feature)")
	return cmd
}
