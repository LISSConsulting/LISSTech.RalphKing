package worktree

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Detect checks that a wt binary is available on PATH. It tries each
// candidate in wtExecutables() order, validates --version output to
// distinguish from Windows Terminal, and caches the successful path.
//
// On Windows, git-wt is tried before wt to avoid the Windows Terminal alias
// collision.
func (r *Runner) Detect() error {
	for _, name := range wtExecutables() {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		// Validate that it's the wt CLI (not Windows Terminal or some other
		// binary) by checking the --version output. CombinedOutput is used
		// because Rust/clap CLIs (and scoop shims) may write to stderr.
		cmd := exec.Command(path, "--version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}
		if isWTVersion(string(bytes.TrimSpace(out))) {
			r.executable = path
			return nil
		}
	}

	return fmt.Errorf(
		"worktree: wt not found on PATH\n" +
			"  Install wt: https://github.com/nicholasgasior/worktrunk\n" +
			"  On Windows, if `wt` resolves to Windows Terminal, use `git-wt` instead",
	)
}

// isWTVersion returns true if the --version output looks like the wt CLI.
// Accepted formats: "wt v0.28.2", "worktrunk 1.0.0", etc.
// Windows Terminal's wt.exe produces empty output, so it won't match.
func isWTVersion(s string) bool {
	low := strings.ToLower(s)
	return strings.HasPrefix(low, "wt v") || strings.Contains(low, "worktrunk")
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
