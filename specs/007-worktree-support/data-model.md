# Data Model: Git Worktree Support via Worktrunk

**Feature**: 007-worktree-support
**Date**: 2026-03-07

## Entities

### WorktreeAgent

Represents a single Claude agent running inside a git worktree.

| Field | Type | Description |
|-------|------|-------------|
| Branch | string | Git branch name (unique identifier per agent) |
| WorktreePath | string | Absolute path to the worktree directory |
| SpecName | string | Name of the spec being worked on (e.g., "007-worktree-support") |
| SpecDir | string | Relative path to spec directory (e.g., "specs/007-worktree-support") |
| State | AgentState | Current lifecycle state |
| Iterations | int | Number of completed loop iterations |
| TotalCost | float64 | Accumulated cost across all iterations |
| Events | chan LogEntry | Agent's event stream (owned by this agent) |
| StopCh | chan struct{} | Channel to signal graceful stop |
| Error | error | Last error if state is Failed |

**State transitions (AgentState)**:

```
Creating → Running → Completed
                  → Failed
                  → Stopped (graceful stop via StopCh)
Completed → Merging → Merged (worktree removed)
                    → MergeFailed (worktree preserved)
Completed → Removed (via clean command)
Failed    → Removed (via clean command)
```

**AgentState values**: `Creating`, `Running`, `Completed`, `Failed`, `Stopped`, `Merging`, `Merged`, `MergeFailed`, `Removed`

**Identity**: Branch name is the unique key. Only one agent per branch (FR-025).

### Orchestrator

Manages the lifecycle of multiple WorktreeAgents.

| Field | Type | Description |
|-------|------|-------------|
| Agents | map[string]*WorktreeAgent | Branch name → agent; mutex-guarded |
| MaxParallel | int | Maximum concurrent agents (from config) |
| AutoMerge | bool | Whether to auto-merge on completion (from config) |
| MergeTarget | string | Default merge target branch (from config) |
| WorktreeRunner | WorktreeOps | Interface to worktrunk CLI adapter |
| MergedEvents | chan TaggedLogEntry | Fan-in channel for TUI consumption |

**Relationships**:
- Owns 0..N WorktreeAgents
- Delegates worktree operations to WorktreeOps (internal/worktree.Runner)
- Delegates supervision to Regent (one Regent per agent)
- Feeds merged event stream to TUI

### TaggedLogEntry

A log entry annotated with its source agent for the fan-in stream.

| Field | Type | Description |
|-------|------|-------------|
| Branch | string | Source agent's branch name |
| Entry | LogEntry | The original log entry from loop.LogEntry |

### WorktreeConfig

The `[worktree]` section of `ralph.toml`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Enabled | bool | false | Enable worktree support in dashboard mode |
| MaxParallel | int | 5 | Maximum concurrent worktree agents |
| AutoMerge | bool | false | Auto-merge on successful completion + test pass |
| MergeTarget | string | "" | Target branch for merge (empty = worktree's base branch) |
| PathTemplate | string | "" | Worktrunk path template (empty = worktrunk default) |

### WorktreeInfo

Parsed output from `wt list --json` for a single worktree.

| Field | Type | Description |
|-------|------|-------------|
| Branch | string | Branch name |
| Path | string | Absolute path to worktree directory |
| IsDefault | bool | Whether this is the main worktree |
| Dirty | bool | Whether worktree has uncommitted changes |
| AheadBehind | string | Commits ahead/behind remote |

## Interfaces

### WorktreeOps

Interface for worktrunk CLI operations (satisfied by `internal/worktree.Runner`).

```
Detect() error                              // Check wt is available
Switch(branch string, create bool) (path string, err error)  // Create or switch to worktree
List() ([]WorktreeInfo, error)              // List all worktrees
Merge(branch, target string) error          // Merge and clean up
Remove(branch string) error                 // Remove worktree + branch
```

### Orchestrator interface (for TUI)

```
Launch(spec SpecFile, mode Mode) error      // Create worktree + start agent
Stop(branch string) error                   // Graceful stop
StopAll() error                             // Graceful stop all agents
Merge(branch string) error                  // Merge completed agent
Clean(branch string) error                  // Remove worktree
ActiveAgents() []*WorktreeAgent             // List all agents
MergedEvents() <-chan TaggedLogEntry         // Fan-in event stream
```
