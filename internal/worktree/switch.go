package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// Switch creates or switches to a worktree for the given branch.
//
// When create is true, it runs `wt switch -c <branch>` to atomically create
// the branch and worktree. When create is false, it runs `wt switch <branch>`
// to reuse an existing worktree.
//
// The worktree path is parsed from worktrunk's stdout: the tool prints a
// success message containing `@ <path>` (e.g. "✓ Created branch foo @ /path").
func (r *Runner) Switch(branch string, create bool) (string, error) {
	var args []string
	if create {
		args = []string{"switch", "-c", branch}
	} else {
		args = []string{"switch", branch}
	}

	cmd := exec.Command(r.exe(), args...)
	cmd.Dir = r.Dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("wt switch %s: %s", branch, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("wt switch %s: %w", branch, err)
	}

	path := parseWorktreePath(string(out))
	if path == "" {
		// Path extraction failed; try to discover it via List().
		infos, listErr := r.List()
		if listErr == nil {
			for _, info := range infos {
				if info.Branch == branch || info.Branch == "refs/heads/"+branch {
					return info.Path, nil
				}
			}
		}
		return "", fmt.Errorf("wt switch %s: could not determine worktree path from output: %s", branch, strings.TrimSpace(string(out)))
	}
	return path, nil
}

// parseWorktreePath extracts the path from worktrunk switch output.
// Worktrunk prints lines like:
//
//	✓ Created branch foo from main and worktree @ /home/user/.worktrees/foo
//
// We look for "@ " and return the remainder of that line trimmed.
func parseWorktreePath(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if idx := strings.Index(line, "@ "); idx >= 0 {
			path := strings.TrimSpace(line[idx+2:])
			if path != "" {
				return path
			}
		}
	}
	return ""
}
