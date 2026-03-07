package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// Merge merges the worktree branch into target by invoking `wt merge <target>`.
// The command is run from the worktree's directory (Dir), which is required by
// worktrunk to identify which worktree to operate on.
//
// If target is empty, worktrunk uses the branch from which the worktree was
// created (its default behaviour).
func (r *Runner) Merge(branch, target string) error {
	args := []string{"merge"}
	if target != "" {
		args = append(args, target)
	}

	cmd := exec.Command(r.exe(), args...)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("wt merge %s: %w", branch, err)
		}
		return fmt.Errorf("wt merge %s: %s", branch, msg)
	}
	return nil
}
