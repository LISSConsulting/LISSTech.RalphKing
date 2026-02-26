package regent

import (
	"bytes"
	"fmt"
	"os/exec"
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
func RunTests(dir, testCommand string) (RunTestResult, error) {
	if testCommand == "" {
		return RunTestResult{Passed: true}, nil
	}

	cmd := exec.Command("sh", "-c", testCommand)
	cmd.Dir = dir

	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err := cmd.Run()
	output := strings.TrimSpace(combined.String())

	if err != nil {
		return RunTestResult{Passed: false, Output: output}, nil
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
