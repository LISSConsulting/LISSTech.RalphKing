package regent

import (
	"errors"
	"strings"
	"testing"
)

// mockGit is a test double for GitOps.
type mockGit struct {
	lastCommit       string
	lastCommitErr    error // error returned by LastCommit
	branch           string
	currentBranchErr error // error returned by CurrentBranch
	revertErr        error
	pushErr          error

	revertCalls []string
	pushCalls   []string
}

func (m *mockGit) LastCommit() (string, error)    { return m.lastCommit, m.lastCommitErr }
func (m *mockGit) CurrentBranch() (string, error) { return m.branch, m.currentBranchErr }

func (m *mockGit) Revert(sha string) error {
	m.revertCalls = append(m.revertCalls, sha)
	return m.revertErr
}

func (m *mockGit) Push(branch string) error {
	m.pushCalls = append(m.pushCalls, branch)
	return m.pushErr
}

func TestRunTests(t *testing.T) {
	tests := []struct {
		name    string
		command string
		passed  bool
	}{
		// Commands that work on both Unix (sh -c) and Windows (cmd /C).
		{"empty command passes", "", true},
		{"exit 0 passes", "exit 0", true},
		{"exit 1 fails", "exit 1", false},
		{"echo passes", "echo hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RunTests(t.TempDir(), tt.command)
			if err != nil {
				t.Fatalf("RunTests error: %v", err)
			}
			if result.Passed != tt.passed {
				t.Errorf("Passed = %v, want %v (output: %s)", result.Passed, tt.passed, result.Output)
			}
		})
	}
}

func TestRunTests_CapturesOutput(t *testing.T) {
	result, err := RunTests(t.TempDir(), "echo test-output-marker")
	if err != nil {
		t.Fatalf("RunTests error: %v", err)
	}
	if !strings.Contains(result.Output, "test-output-marker") {
		t.Errorf("expected output to contain marker, got: %s", result.Output)
	}
}

func TestRunTests_ShellNotFound(t *testing.T) {
	// Clearing PATH prevents exec.LookPath from finding the shell binary (sh or cmd).
	// RunTests must return a real error, not treat this as a test failure.
	t.Setenv("PATH", "")

	_, err := RunTests(t.TempDir(), "exit 0")
	if err == nil {
		t.Error("expected error when shell is not found in PATH, got nil")
	}
}

func TestRevertLastCommit(t *testing.T) {
	t.Run("reverts and pushes", func(t *testing.T) {
		g := &mockGit{lastCommit: "abc1234 bad commit", branch: "feat/test"}
		sha, err := RevertLastCommit(g)
		if err != nil {
			t.Fatalf("RevertLastCommit error: %v", err)
		}
		if sha != "abc1234" {
			t.Errorf("sha = %q, want %q", sha, "abc1234")
		}
		if len(g.revertCalls) != 1 || g.revertCalls[0] != "abc1234" {
			t.Errorf("revert calls = %v, want [abc1234]", g.revertCalls)
		}
		if len(g.pushCalls) != 1 || g.pushCalls[0] != "feat/test" {
			t.Errorf("push calls = %v, want [feat/test]", g.pushCalls)
		}
	})

	t.Run("handles commit with no space (SHA only)", func(t *testing.T) {
		g := &mockGit{lastCommit: "abc1234", branch: "main"}
		sha, err := RevertLastCommit(g)
		if err != nil {
			t.Fatalf("RevertLastCommit error: %v", err)
		}
		if sha != "abc1234" {
			t.Errorf("sha = %q, want %q", sha, "abc1234")
		}
	})

	t.Run("revert error propagates", func(t *testing.T) {
		g := &mockGit{
			lastCommit: "abc1234 bad commit",
			branch:     "main",
			revertErr:  errors.New("revert conflict"),
		}
		sha, err := RevertLastCommit(g)
		if err == nil {
			t.Fatal("expected error when Revert fails")
		}
		if !strings.Contains(err.Error(), "revert conflict") {
			t.Errorf("error should contain cause, got: %v", err)
		}
		// sha is still returned even on revert error
		if sha != "abc1234" {
			t.Errorf("sha = %q, want %q", sha, "abc1234")
		}
	})

	t.Run("push error propagates", func(t *testing.T) {
		g := &mockGit{
			lastCommit: "abc1234 bad commit",
			branch:     "main",
			pushErr:    errors.New("push rejected"),
		}
		sha, err := RevertLastCommit(g)
		if err == nil {
			t.Fatal("expected error when Push fails")
		}
		if !strings.Contains(err.Error(), "push rejected") {
			t.Errorf("error should contain cause, got: %v", err)
		}
		if sha != "abc1234" {
			t.Errorf("sha = %q, want %q", sha, "abc1234")
		}
	})

	t.Run("LastCommit error propagates", func(t *testing.T) {
		g := &mockGit{
			lastCommitErr: errors.New("git log failed"),
			branch:        "main",
		}
		sha, err := RevertLastCommit(g)
		if err == nil {
			t.Fatal("expected error when LastCommit fails")
		}
		if !strings.Contains(err.Error(), "git log failed") {
			t.Errorf("error should contain cause, got: %v", err)
		}
		if sha != "" {
			t.Errorf("sha should be empty on LastCommit error, got %q", sha)
		}
		if len(g.revertCalls) != 0 {
			t.Error("should not attempt revert when LastCommit fails")
		}
	})

	t.Run("CurrentBranch error propagates", func(t *testing.T) {
		g := &mockGit{
			lastCommit:       "abc1234 bad commit",
			branch:           "main",
			currentBranchErr: errors.New("detached HEAD"),
		}
		sha, err := RevertLastCommit(g)
		if err == nil {
			t.Fatal("expected error when CurrentBranch fails")
		}
		if !strings.Contains(err.Error(), "detached HEAD") {
			t.Errorf("error should contain cause, got: %v", err)
		}
		if sha != "abc1234" {
			t.Errorf("sha = %q, want %q", sha, "abc1234")
		}
		// Revert should have been called, but push should not
		if len(g.revertCalls) != 1 {
			t.Errorf("expected 1 revert call, got %d", len(g.revertCalls))
		}
		if len(g.pushCalls) != 0 {
			t.Error("should not push when CurrentBranch fails")
		}
	})
}
