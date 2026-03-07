package worktree

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Detect checks that a worktrunk binary is available on PATH. It tries each
// candidate in wtExecutables() order, checks that --version output contains
// "worktrunk", and caches the successful executable path on the Runner.
//
// On Windows, git-wt is tried before wt to avoid the Windows Terminal alias
// collision.
func (r *Runner) Detect() error {
	for _, name := range wtExecutables() {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		// Validate that it's actually worktrunk (not Windows Terminal or some
		// other wt binary) by checking the --version output.
		cmd := exec.Command(path, "--version")
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(bytes.TrimSpace(out))), "worktrunk") {
			r.executable = path
			return nil
		}
	}

	return fmt.Errorf(
		"worktree: worktrunk not found on PATH\n" +
			"  Install worktrunk: https://github.com/nicholasgasior/worktrunk\n" +
			"  On Windows, if `wt` resolves to Windows Terminal, use `git-wt` instead",
	)
}

// exe returns the cached executable path, falling back to "wt" if Detect was
// not called. Callers that skip Detect() will get a helpful exec error.
func (r *Runner) exe() string {
	if r.executable != "" {
		return r.executable
	}
	if candidates := wtExecutables(); len(candidates) > 0 {
		return candidates[0]
	}
	return "wt"
}
