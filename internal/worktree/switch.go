package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Switch creates or switches to a worktree for the given branch.
//
// When WorktreeDir is set, Switch uses `git worktree add` directly with the
// custom base directory instead of delegating to worktrunk. This allows Ralph
// to control where worktrees are stored (e.g. ~/.ralph/worktrees).
//
// Otherwise, when create is true, it runs `wt switch -c <branch>` to atomically
// create the branch and worktree. When create is false, it runs `wt switch <branch>`
// to reuse an existing worktree.
//
// The worktree path is parsed from worktrunk's stdout: the tool prints a
// success message containing `@ <path>` (e.g. "✓ Created branch foo @ /path").
func (r *Runner) Switch(branch string, create bool) (string, error) {
	// When a custom worktree directory is configured, use git directly.
	if r.WorktreeDir != "" {
		return r.switchGit(branch, create)
	}

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

// switchGit uses `git worktree add` directly with a custom path under WorktreeDir.
// If the worktree already exists at that path, it returns the path directly.
func (r *Runner) switchGit(branch string, create bool) (string, error) {
	// Sanitise branch name for use as a directory name.
	dirName := strings.ReplaceAll(branch, "/", "-")
	wtPath := filepath.Join(r.WorktreeDir, dirName)

	// If the worktree directory already exists, reuse it.
	if info, err := os.Stat(wtPath); err == nil && info.IsDir() {
		return wtPath, nil
	}

	// Ensure the parent directory exists.
	if err := os.MkdirAll(r.WorktreeDir, 0o755); err != nil {
		return "", fmt.Errorf("create worktree dir %s: %w", r.WorktreeDir, err)
	}

	var cmd *exec.Cmd
	if create {
		// Create new branch and worktree: git worktree add -b <branch> <path>
		cmd = exec.Command("git", "worktree", "add", "-b", branch, wtPath)
	} else {
		// Create worktree for existing branch: git worktree add <path> <branch>
		cmd = exec.Command("git", "worktree", "add", wtPath, branch)
	}
	cmd.Dir = r.Dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add %s: %s", branch, strings.TrimSpace(string(out)))
	}
	return wtPath, nil
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
