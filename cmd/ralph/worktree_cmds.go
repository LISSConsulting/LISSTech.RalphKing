package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/LISSConsulting/RalphSpec/internal/worktree"
)

// worktreeCmd returns the `ralph worktree` parent command.
func worktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktrees for parallel agent workflows",
	}
	cmd.AddCommand(worktreeListCmd(), worktreeMergeCmd(), worktreeCleanCmd())
	return cmd
}

// worktreeListCmd implements `ralph worktree list`.
func worktreeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active git worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			wtr := worktree.NewRunner(dir)
			if err := wtr.Detect(); err != nil {
				return err
			}

			infos, err := wtr.List()
			if err != nil {
				return err
			}

			asJSON, _ := cmd.Flags().GetBool("json")
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(infos)
			}

			fmt.Print(formatWorktreeList(infos))
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "output as JSON")
	return cmd
}

// worktreeMergeCmd implements `ralph worktree merge [branch]`.
func worktreeMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge [branch]",
		Short: "Merge a completed worktree branch and clean up",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			wtr := worktree.NewRunner(dir)
			if err := wtr.Detect(); err != nil {
				return err
			}

			target, _ := cmd.Flags().GetString("target")

			var branch string
			if len(args) > 0 {
				branch = args[0]
			} else {
				// Derive branch from the current worktree's git state.
				infos, listErr := wtr.List()
				if listErr != nil {
					return fmt.Errorf("list worktrees: %w", listErr)
				}
				if len(infos) == 0 {
					return fmt.Errorf("no worktrees found — specify a branch name")
				}
				// Find the worktree whose path matches the current dir.
				for _, info := range infos {
					if info.Path == dir {
						branch = info.Branch
						break
					}
				}
				if branch == "" {
					return fmt.Errorf("could not determine current worktree branch — specify a branch name")
				}
			}

			if err := wtr.Merge(branch, target); err != nil {
				return err
			}
			fmt.Printf("Merged %s into %s\n", branch, target)

			// Remove the worktree unless --no-remove was specified.
			noRemove, _ := cmd.Flags().GetBool("no-remove")
			if !noRemove {
				if rmErr := wtr.Remove(branch); rmErr != nil {
					fmt.Fprintf(os.Stderr, "ralph: remove worktree after merge: %v\n", rmErr)
				}
			}
			return nil
		},
	}
	cmd.Flags().String("target", "", "target branch to merge into (empty = worktrunk default)")
	cmd.Flags().Bool("no-remove", false, "keep worktree after merge")
	return cmd
}

// worktreeCleanCmd implements `ralph worktree clean [branch|--all]`.
func worktreeCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean [branch]",
		Short: "Remove one or more completed/failed worktrees",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			wtr := worktree.NewRunner(dir)
			if err := wtr.Detect(); err != nil {
				return err
			}

			all, _ := cmd.Flags().GetBool("all")

			if all {
				infos, listErr := wtr.List()
				if listErr != nil {
					return fmt.Errorf("list worktrees: %w", listErr)
				}
				var errs []error
				for _, info := range infos {
					if info.Bare {
						continue // never remove bare worktrees
					}
					if rmErr := wtr.Remove(info.Branch); rmErr != nil {
						errs = append(errs, fmt.Errorf("%s: %w", info.Branch, rmErr))
					} else {
						fmt.Printf("Removed worktree %s\n", info.Branch)
					}
				}
				if len(errs) > 0 {
					for _, e := range errs {
						fmt.Fprintf(os.Stderr, "ralph: %v\n", e)
					}
					return fmt.Errorf("some worktrees could not be removed")
				}
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a branch name or use --all")
			}
			branch := args[0]
			if err := wtr.Remove(branch); err != nil {
				return err
			}
			fmt.Printf("Removed worktree %s\n", branch)
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "remove all non-bare worktrees")
	cmd.Flags().Bool("force", false, "force removal even if branch has uncommitted changes")
	return cmd
}

// formatWorktreeList renders a list of WorktreeInfo entries as a table.
func formatWorktreeList(infos []worktree.WorktreeInfo) string {
	if len(infos) == 0 {
		return "No worktrees found.\n"
	}
	var b []byte
	b = append(b, "Worktrees\n─────────\n"...)
	for _, info := range infos {
		branch := info.Branch
		if branch == "" {
			branch = "(detached)"
		}
		kind := ""
		if info.Bare {
			kind = "  [bare]"
		}
		b = fmt.Appendf(b, "  %-30s  %s%s\n", branch, info.Path, kind)
	}
	return string(b)
}
