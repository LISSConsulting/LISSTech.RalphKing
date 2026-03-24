// Package worktree provides a thin adapter around the worktrunk CLI (`wt`).
// Worktrunk is an external binary dependency; Ralph invokes it as a subprocess
// rather than importing it as a Go library.
package worktree

import (
	"runtime"
)

// WorktreeInfo describes a single git worktree managed by worktrunk.
type WorktreeInfo struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	Bare   bool   `json:"bare"`
}

// WorktreeOps is the interface that callers use to interact with worktrees.
// Runner satisfies this interface.
type WorktreeOps interface {
	Detect() error
	Switch(branch string, create bool) (path string, err error)
	List() ([]WorktreeInfo, error)
	Merge(branch, target string) error
	Remove(branch string) error
}

// Runner invokes worktrunk commands as subprocesses from a given working
// directory. Dir must be the root of the git repository.
type Runner struct {
	Dir         string
	WorktreeDir string // custom base directory for worktrees; empty = worktrunk default
	executable  string // cached after Detect()
}

// NewRunner returns a new Runner rooted at dir.
func NewRunner(dir string) *Runner {
	return &Runner{Dir: dir}
}

// wtExecutables returns the list of binary names to try when looking for
// worktrunk. On Windows, `wt` may resolve to Windows Terminal's App Execution
// Alias, so `git-wt` is preferred there.
func wtExecutables() []string {
	if runtime.GOOS == "windows" {
		return []string{"git-wt", "wt"}
	}
	return []string{"wt", "git-wt"}
}
