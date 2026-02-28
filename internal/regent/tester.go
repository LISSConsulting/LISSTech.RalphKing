package regent

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GitOps defines the git operations the Regent needs for test-gated rollback.
type GitOps interface {
	LastCommit() (string, error)
	Revert(sha string) error
	Push(branch string) error
	CurrentBranch() (string, error)
}

// RunTestResult holds the outcome of a test command execution.
type RunTestResult struct {
	Passed bool
	Output string
}

// RunTests executes the test command in the given directory and returns the result.
// Returns an error only if the command could not be started (not if tests fail).
// On Windows, the command is run via cmd /C; on Unix, via sh -c.
func RunTests(dir, testCommand string) (RunTestResult, error) {
	if testCommand == "" {
		return RunTestResult{Passed: true}, nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", testCommand)
	} else {
		cmd = exec.Command("sh", "-c", testCommand)
	}
	cmd.Dir = dir

	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err := cmd.Run()
	output := strings.TrimSpace(combined.String())

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Test command ran but returned non-zero exit code — test failure.
			return RunTestResult{Passed: false, Output: output}, nil
		}
		// Shell binary could not be started (not found in PATH, permission denied, etc.).
		return RunTestResult{}, fmt.Errorf("regent: run test command: %w", err)
	}
	return RunTestResult{Passed: true, Output: output}, nil
}

// RevertLastCommit reverts HEAD and pushes the revert. Returns the SHA that was reverted.
func RevertLastCommit(gitOps GitOps) (string, error) {
	commit, err := gitOps.LastCommit()
	if err != nil {
		return "", fmt.Errorf("regent: get last commit for revert: %w", err)
	}

	// Extract just the short SHA (first field of "abc1234 commit message")
	sha := commit
	if idx := strings.IndexByte(commit, ' '); idx > 0 {
		sha = commit[:idx]
	}

	if revertErr := gitOps.Revert(sha); revertErr != nil {
		return sha, fmt.Errorf("regent: revert %s: %w", sha, revertErr)
	}

	branch, err := gitOps.CurrentBranch()
	if err != nil {
		return sha, fmt.Errorf("regent: get branch for push after revert: %w", err)
	}

	if pushErr := gitOps.Push(branch); pushErr != nil {
		return sha, fmt.Errorf("regent: push revert: %w", pushErr)
	}

	return sha, nil
}
