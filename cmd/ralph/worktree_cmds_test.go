package main

import (
	"strings"
	"testing"

	"github.com/LISSConsulting/RalphSpec/internal/worktree"
)

// TestWorktreeListCmd_DetectFails covers worktreeListCmd.RunE up to and
// including the Detect() error return, which is always hit in test environments
// where worktrunk is not installed.
func TestWorktreeListCmd_DetectFails(t *testing.T) {
	t.Setenv("PATH", "")
	t.Chdir(t.TempDir())
	cmd := worktreeListCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when worktrunk not on PATH")
	}
}

// TestWorktreeMergeCmd_DetectFails mirrors the above for worktreeMergeCmd.
func TestWorktreeMergeCmd_DetectFails(t *testing.T) {
	t.Setenv("PATH", "")
	t.Chdir(t.TempDir())
	cmd := worktreeMergeCmd()
	err := cmd.RunE(cmd, []string{"feat/x"})
	if err == nil {
		t.Fatal("expected error when worktrunk not on PATH")
	}
}

// TestWorktreeCleanCmd_DetectFails mirrors the above for worktreeCleanCmd.
func TestWorktreeCleanCmd_DetectFails(t *testing.T) {
	t.Setenv("PATH", "")
	t.Chdir(t.TempDir())
	cmd := worktreeCleanCmd()
	err := cmd.RunE(cmd, []string{"feat/x"})
	if err == nil {
		t.Fatal("expected error when worktrunk not on PATH")
	}
}

func TestFormatWorktreeList_Empty(t *testing.T) {
	got := formatWorktreeList(nil)
	if got != "No worktrees found.\n" {
		t.Errorf("empty list: got %q", got)
	}
}

func TestFormatWorktreeList_Entries(t *testing.T) {
	infos := []worktree.WorktreeInfo{
		{Branch: "main", Path: "/repo", Bare: false},
		{Branch: "feat/x", Path: "/worktrees/feat-x", Bare: false},
		{Branch: "", Path: "/worktrees/detached", Bare: false},
		{Branch: "bare-wt", Path: "/bare", Bare: true},
	}
	got := formatWorktreeList(infos)

	if !strings.Contains(got, "Worktrees") {
		t.Errorf("missing header: %q", got)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("missing main branch: %q", got)
	}
	if !strings.Contains(got, "feat/x") {
		t.Errorf("missing feat/x branch: %q", got)
	}
	if !strings.Contains(got, "(detached)") {
		t.Errorf("missing (detached) for empty branch: %q", got)
	}
	if !strings.Contains(got, "[bare]") {
		t.Errorf("missing [bare] annotation: %q", got)
	}
}
