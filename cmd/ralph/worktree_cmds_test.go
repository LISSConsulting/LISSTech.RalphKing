package main

import (
	"strings"
	"testing"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/worktree"
)

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
