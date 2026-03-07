# Research: Git Worktree Support via Worktrunk

**Feature**: 007-worktree-support
**Date**: 2026-03-07

## R-001: Worktrunk CLI Output Parsing

**Decision**: Use `wt list --json` for reliable worktree path discovery. Parse `wt switch -c` stdout for the path token after `@ ` in the success line.

**Rationale**: `wt list --json` provides structured data (branch, path, status, ahead/behind). Human-readable output from `wt switch` is sufficient for initial path capture since it's a single known format.

**Alternatives considered**:
- Computing paths from worktrunk's path template config — rejected; couples Ralph to worktrunk internals
- Using `git worktree list --porcelain` — rejected; bypasses worktrunk features (hooks, status markers)

## R-002: Worktrunk on Windows

**Decision**: Auto-detect by checking `git-wt` first on Windows (avoids Windows Terminal alias conflict), then fall back to `wt`. Validate with `--version` containing "worktrunk".

**Rationale**: Worktrunk's winget package installs as `git-wt`. Users who disable the WT alias can use `wt` directly. Auto-detection handles both cases.

**Alternatives considered**:
- Config option for binary name — rejected as unnecessary; auto-detection is two lines of code
- Requiring users to alias — rejected; poor UX

## R-003: Subprocess Faking for Tests

**Decision**: `_FAKE_WT=1` env pattern with `init()` test binary registration, identical to existing `_FAKE_CLAUDE=1` in `internal/claude/`.

**Rationale**: Proven pattern in this codebase. Tests exercise real subprocess invocation paths.

**Alternatives considered**:
- Interface mocking — rejected; doesn't test the subprocess boundary
- Exec faking via `os/exec` test helper — same pattern, different name; using project convention

## R-004: Orchestrator Concurrency Model

**Decision**: `sync.Mutex`-guarded `map[string]*WorktreeAgent` in Orchestrator. Each agent's `Loop.Run()` executes in a dedicated goroutine. Fan-in multiplexer tags events with branch name before forwarding to TUI.

**Rationale**: Loop struct is already self-contained (Dir, GitOps, Agent, Events). Parallel agents = N Loop instances with different Dir values. Fan-in is clean for lifecycle management.

**Alternatives considered**:
- Round-robin polling — rejected; adds latency
- Shared channel with branch tags — viable but fan-in gives per-agent goroutine cleanup

## R-005: Log Aggregation Strategy

**Decision**: Hybrid — each agent logs to `<worktree>/.ralph/logs/`. Dashboard aggregates by scanning paths tracked by Orchestrator.

**Rationale**: Per clarification. Self-contained worktrees (inspectable independently) + unified dashboard view.

**Alternatives considered**:
- Centralized in main repo — rejected per clarification
- Distributed without aggregation — rejected; dashboard needs unified view

## R-006: Minimum Worktrunk Version

**Decision**: No hard minimum. Feature-detect at runtime (e.g., `wt list --json` support). Log clear upgrade message if feature missing.

**Rationale**: Worktrunk is new and evolving. Hard pins frustrate users; feature detection is resilient.
