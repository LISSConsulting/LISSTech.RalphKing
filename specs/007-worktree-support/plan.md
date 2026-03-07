# Implementation Plan: Git Worktree Support via Worktrunk

**Branch**: `007-worktree-support` | **Date**: 2026-03-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/007-worktree-support/spec.md`

## Summary

Integrate worktrunk (`wt`) into Ralph to enable isolated and parallel AI agent workflows using git worktrees. A new `--worktree` flag runs build loops in separate worktrees. The TUI dashboard gains multi-agent orchestration (launch, monitor, merge, clean). Auto-merge is opt-in. The Regent supervises each agent independently. All existing single-agent behavior is preserved when worktree support is not activated.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: cobra, BurntSushi/toml, bubbletea, lipgloss, bubbles (all existing); worktrunk (`wt`) as external CLI dependency (not a Go import)
**Storage**: JSONL session logs in `<worktree>/.ralph/logs/`; `regent-state.json` per worktree
**Testing**: `go test ./...` with table-driven subtests; subprocess faking via `_FAKE_WT=1` env pattern (same as existing `_FAKE_CLAUDE=1`)
**Target Platform**: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
**Project Type**: CLI tool
**Performance Goals**: Worktree creation/switch adds <2s overhead per agent launch; TUI remains responsive with up to 5 concurrent agents streaming events
**Constraints**: No new Go dependencies; worktrunk is an external binary dependency only
**Scale/Scope**: Up to 5 parallel agents (configurable); each worktree is a full working copy sharing git objects

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | Spec 007 defines all requirements; plan implements spec |
| II. Supervised Autonomy | PASS | Regent extended to per-worktree supervision (FR-019/FR-020) |
| III. Test-Gated Commits | PASS | Per-worktree test-gated rollback; auto-merge gated on test pass |
| IV. Idiomatic Go | PASS | No new Go deps; worktrunk invoked as subprocess; new packages follow existing patterns |
| V. Observable Loops | PASS | Hybrid log storage; TUI aggregates across worktrees; each agent streams events |

No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/007-worktree-support/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: worktrunk integration research
├── data-model.md        # Phase 1: entity model
├── contracts/           # Phase 1: CLI contract
│   └── cli-commands.md  # New worktree subcommands
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/ralph/
├── commands.go          # Add worktree subcommands (worktree merge, worktree clean, worktree list)
├── execute.go           # Add --worktree flag handling to loop execution
├── wiring.go            # Orchestrator wiring for multi-agent TUI
└── worktree_cmds.go     # NEW: ralph worktree merge/clean/list command implementations

internal/
├── config/
│   └── config.go        # Add WorktreeConfig struct and [worktree] TOML section
├── worktree/            # NEW PACKAGE: worktrunk CLI adapter
│   ├── worktree.go      # Runner: detect wt, switch, merge, remove, list
│   └── worktree_test.go # Subprocess faking via _FAKE_WT=1 pattern
├── orchestrator/        # NEW PACKAGE: multi-agent lifecycle management
│   ├── orchestrator.go  # Agent registry, launch, stop, status, merge, clean
│   └── orchestrator_test.go
├── loop/
│   └── loop.go          # No changes — Loop struct already supports arbitrary Dir
├── store/
│   └── jsonl.go         # No changes — store already creates .ralph/logs/ per Dir
├── regent/
│   └── regent.go        # Extend to accept multiple supervised agents
└── tui/
    ├── app.go           # Wire orchestrator events; new keybinds (W, M, X per worktree)
    └── panels/
        └── worktrees.go # NEW: worktree status panel (list of agents + status)
```

**Structure Decision**: Two new packages (`internal/worktree/`, `internal/orchestrator/`) follow the existing pattern of one concern per package. `internal/worktree/` owns the worktrunk CLI adapter (subprocess calls). `internal/orchestrator/` owns multi-agent lifecycle. This keeps the `loop` package unchanged — each agent is just a `Loop` instance with a different `Dir`.

## Phase 0: Research

### R-001: Worktrunk CLI Output Parsing

**Decision**: Parse `wt switch -c <branch>` stdout to extract the worktree path. Worktrunk outputs the path in its success message (e.g., `✓ Created branch <name> from main and worktree @ <path>`). Use `wt list --json` (if available) as the reliable path discovery mechanism for existing worktrees.

**Rationale**: `wt list --json` provides structured output with worktree paths, branch names, and status. This avoids fragile regex parsing of human-readable output. For `wt switch -c`, capture stdout and parse the path after `@ `.

**Alternatives considered**:
- Computing path from worktrunk's path template — rejected because it couples Ralph to worktrunk's internal path logic.
- Using `git worktree list --porcelain` directly — rejected because it bypasses worktrunk and loses its features (hooks, status tracking).

### R-002: Worktrunk on Windows

**Decision**: On Windows, check for both `wt` and `git-wt` executables. Prefer `git-wt` if `wt` resolves to Windows Terminal. Detect by checking if `wt --version` output contains "worktrunk" or similar identifier.

**Rationale**: Windows Terminal registers `wt.exe` as an App Execution Alias. Worktrunk's winget package installs as `git-wt` to avoid this conflict. Ralph must handle both.

**Alternatives considered**:
- Requiring users to configure the binary name in ralph.toml — rejected as unnecessary complexity; auto-detection is straightforward.

### R-003: Subprocess Faking for Tests

**Decision**: Use the established `_FAKE_WT=1` environment variable pattern (same as `_FAKE_CLAUDE=1` in `internal/claude/`). Test binaries intercept `wt` commands and return canned responses. Use `init()` function for test binary registration.

**Rationale**: Proven pattern in this codebase. Avoids interface mocking overhead while testing real subprocess invocation paths.

### R-004: Orchestrator Concurrency Model

**Decision**: The Orchestrator holds a `sync.Mutex`-guarded map of `WorktreeAgent` structs. Each agent runs its `Loop.Run()` in a dedicated goroutine. Events flow from each agent's `chan LogEntry` to the TUI via a fan-in multiplexer that tags each event with the source branch name.

**Rationale**: The existing `Loop` struct is already self-contained — it takes a `Dir`, `GitOps`, `Agent`, and `Events` channel. Launching parallel agents means creating N `Loop` instances with different `Dir` values, each in its own goroutine. The fan-in pattern lets the TUI receive a single merged event stream.

**Alternatives considered**:
- Single goroutine with round-robin polling — rejected because it adds latency and complexity.
- Shared `Events` channel with branch-tagged entries — considered viable but fan-in is cleaner for per-agent lifecycle management.

### R-005: Log Aggregation Strategy

**Decision**: Hybrid storage. Each agent writes to `<worktree>/.ralph/logs/`. The dashboard's store reader scans all known worktree paths (tracked by the Orchestrator) to aggregate logs for the TUI. The Orchestrator maintains a `[]string` of active worktree paths.

**Rationale**: Per-clarification Q1. Keeps each worktree self-contained (can be inspected independently) while the dashboard aggregates for the unified view.

### R-006: Minimum Worktrunk Version

**Decision**: No minimum version enforced at launch. Ralph will check that `wt` is available and responds to `--version`. If a required feature (e.g., `wt list --json`) is missing in an older version, Ralph will log a clear error suggesting an upgrade.

**Rationale**: Worktrunk is new (Feb 2026) and evolving rapidly. Hard version pins would frustrate users. Feature-detection is more resilient.

## Phase 1: Design

### Data Model

See [data-model.md](data-model.md) for full entity definitions.

Key entities:
- **WorktreeAgent** — tracks one agent's branch, path, state, cost, iterations, spec
- **Orchestrator** — manages agent map, enforces limits, routes events, triggers merge/clean
- **WorktreeConfig** — TOML config: enabled, max_parallel, auto_merge, merge_target, path_template

### Contracts

See [contracts/cli-commands.md](contracts/cli-commands.md) for CLI interface.

New commands:
- `ralph worktree list` — show active worktrees and their status
- `ralph worktree merge [branch]` — merge and clean up a completed worktree
- `ralph worktree clean [branch|--all]` — remove completed/failed worktrees

Modified commands:
- `ralph build --worktree` / `-w` — run loop in isolated worktree
- `ralph loop build/plan/run --worktree` — same flag on all loop commands
- Dashboard keybinds: `W` (launch in worktree), `M` (merge selected), `D` (clean/discard selected)

### Post-Design Constitution Re-check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | All new packages and commands trace to spec FR-001 through FR-029 |
| II. Supervised Autonomy | PASS | Regent extended per-worktree; no unsupervised agent runs |
| III. Test-Gated Commits | PASS | Auto-merge gated on test pass (FR-012); per-worktree rollback (FR-019) |
| IV. Idiomatic Go | PASS | Two new packages, clear boundaries; no new Go deps; subprocess adapter pattern |
| V. Observable Loops | PASS | Fan-in event stream; worktree panel; hybrid log aggregation |

No violations. Plan is ready for task breakdown.
