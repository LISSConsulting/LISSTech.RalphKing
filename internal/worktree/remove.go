package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// Remove removes the worktree and its branch via `wt remove <branch>`.
// Returns an error if worktrunk reports that an agent is still running in the
// worktree or if the removal otherwise fails.
func (r *Runner) Remove(branch string) error {
	cmd := exec.Command(r.exe(), "remove", branch)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("wt remove %s: %w", branch, err)
		}
		return fmt.Errorf("wt remove %s: %s", branch, msg)
	}
	return nil
}
