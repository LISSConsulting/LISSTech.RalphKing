package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

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
		Short: "Scaffold ralph project (config, prompts, specs dir)",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			created, err := config.ScaffoldProject(dir)
			if err != nil {
				return err
			}
			if len(created) == 0 {
				fmt.Println("All files already exist — nothing to create.")
				return nil
			}
			for _, path := range created {
				fmt.Printf("Created %s\n", path)
			}
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
