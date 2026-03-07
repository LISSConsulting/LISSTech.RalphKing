package worktree

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// List returns all worktrees currently managed by worktrunk.
//
// It first tries `wt list --json` for structured output. If that flag is not
// supported (older worktrunk versions), it falls back to
// `git worktree list --porcelain` and parses the plain-text format.
func (r *Runner) List() ([]WorktreeInfo, error) {
	infos, err := r.listJSON()
	if err == nil {
		return infos, nil
	}
	// Fall back to git worktree list --porcelain.
	return r.listPorcelain()
}

func (r *Runner) listJSON() ([]WorktreeInfo, error) {
	cmd := exec.Command(r.exe(), "list", "--json")
	cmd.Dir = r.Dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("wt list --json: %w", err)
	}
	var infos []WorktreeInfo
	if jsonErr := json.Unmarshal(out, &infos); jsonErr != nil {
		return nil, fmt.Errorf("wt list --json: parse: %w", jsonErr)
	}
	return infos, nil
}

func (r *Runner) listPorcelain() ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = r.Dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	return parsePorcelain(string(out)), nil
}

// parsePorcelain parses the output of `git worktree list --porcelain`.
// Each worktree block is separated by a blank line and starts with:
//
//	worktree <path>
//	HEAD <sha>
//	branch refs/heads/<name>   (or "detached" for detached HEADs)
func parsePorcelain(output string) []WorktreeInfo {
	var result []WorktreeInfo
	var current WorktreeInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				result = append(result, current)
			}
			current = WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			// Trim "refs/heads/" prefix for a cleaner branch name.
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.Bare = true
		}
	}
	if current.Path != "" {
		result = append(result, current)
	}
	return result
}
