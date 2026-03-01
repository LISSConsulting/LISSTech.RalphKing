# Data Model: Panel-Based TUI Redesign

**Feature**: 003-tui-redesign | **Date**: 2026-02-28

## Entities

### Existing (unchanged)

#### LogEntry (`internal/loop/event.go`)

The event primitive. No schema changes.

| Field | Type | Description |
|-------|------|-------------|
| Kind | LogKind (int) | Event type: LogInfo, LogIterStart, LogToolUse, LogText, LogIterComplete, LogError, LogGitPull, LogGitPush, LogDone, LogStopped, LogRegent |
| Timestamp | time.Time | When the event was emitted |
| Message | string | Human-readable message |
| ToolName | string | Tool name (for LogToolUse) |
| ToolInput | string | Tool input (for LogToolUse) |
| CostUSD | float64 | Iteration cost (for LogIterComplete) |
| Duration | float64 | Iteration duration seconds (for LogIterComplete) |
| TotalCost | float64 | Running total cost |
| Subtype | string | Exit subtype: "success", "error_max_turns", etc. |
| Iteration | int | Current iteration number |
| MaxIter | int | Max iterations configured |
| Branch | string | Git branch |
| Commit | string | Git commit hash |
| Mode | string | "plan" or "build" |

#### IterationSummary (`internal/store/store.go`)

Per-iteration metadata extracted from LogIterStart/LogIterComplete events.

| Field | Type | Description |
|-------|------|-------------|
| Number | int | Iteration sequence number |
| Mode | string | "plan" or "build" |
| CostUSD | float64 | Iteration cost |
| Duration | float64 | Seconds elapsed |
| Subtype | string | Exit subtype |
| Commit | string | Commit hash at iteration end |
| StartAt | time.Time | Iteration start time |
| EndAt | time.Time | Iteration end time |

#### SessionSummary (`internal/store/store.go`)

Session-level aggregate.

| Field | Type | Description |
|-------|------|-------------|
| SessionID | string | `<unix-ts>-<pid>` |
| StartedAt | time.Time | Session start |
| TotalCost | float64 | Sum of iteration costs |
| Iterations | int | Count of completed iterations |
| LastCommit | string | Most recent commit hash |
| Branch | string | Git branch |

#### SpecFile (`internal/spec/spec.go`)

Discovered spec with status.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Filename without extension |
| Path | string | Relative path from project root |
| Status | Status | done, in_progress, not_started |

### New Entities

#### FocusTarget (`internal/tui/focus.go`)

Enum identifying the focused panel.

| Value | Int | Description |
|-------|-----|-------------|
| FocusSpecs | 0 | Specs list panel |
| FocusIterations | 1 | Iterations list panel |
| FocusMain | 2 | Main content panel |
| FocusSecondary | 3 | Secondary info panel |

**Lifecycle**: Set to `FocusSpecs` on init. Cycled by `tab`/`shift+tab`, jumped by `1`-`4`.

#### LoopState (`internal/tui/focus.go`)

Strict state machine for the loop's current operational state.

| Value | Int | Label | Symbol |
|-------|-----|-------|--------|
| StateIdle | 0 | IDLE | ✓ |
| StatePlanning | 1 | PLANNING | ● |
| StateBuilding | 2 | BUILDING | ● |
| StateFailed | 3 | FAILED | ✗ |
| StateRegentRestart | 4 | REGENT RESTART | ⟳ |

**Transitions**:

```
idle → planning        (user presses p / smart-run starts planning)
idle → building        (user presses b / ralph build starts)
planning → building    (plan phase completes, build begins)
planning → failed      (plan phase errors)
building → idle        (all iterations complete successfully)
building → failed      (unrecoverable error or user stops)
building → regentRestart (Regent kills loop)
regentRestart → building (Regent relaunches loop)
failed → idle          (user acknowledges or restarts)
```

Invalid transitions are no-ops (logged as warning).

#### Layout / Rect (`internal/tui/layout.go`)

Panel rectangle dimensions computed from terminal size.

| Field | Type | Description |
|-------|------|-------------|
| X | int | Left column |
| Y | int | Top row |
| Width | int | Panel width |
| Height | int | Panel height |

**Layout** aggregates all panel rects plus a `TooSmall bool` flag.

#### MainTab (`internal/tui/panels/main_view.go`)

Active tab in the main panel.

| Value | Int | Description |
|-------|-----|-------------|
| TabOutput | 0 | Live agent output (auto-scroll) |
| TabSpecContent | 1 | Selected spec rendered as text |
| TabIterationDetail | 2 | Past iteration log from store |

#### SecondaryTab (`internal/tui/panels/secondary.go`)

Active tab in the secondary panel.

| Value | Int | Description |
|-------|-----|-------------|
| TabRegent | 0 | Regent messages |
| TabGit | 1 | Git operations log |
| TabTests | 2 | Test output |
| TabCost | 3 | Per-iteration cost table |

## Relationships

```
Store (JSONL file)
  └── has many → IterationSummary (indexed by Number)
  └── has many → LogEntry (per iteration, retrieved by byte offset)

Root Model
  ├── has one → FocusTarget (current focus)
  ├── has one → LoopState (strict state machine)
  ├── has one → Layout (recalculated on WindowSizeMsg)
  ├── has one → SpecsPanel
  │     └── has many → SpecFile (from spec.List())
  ├── has one → IterationsPanel
  │     └── has many → IterationSummary (from store + live events)
  ├── has one → MainView
  │     ├── has one → LogView (live output)
  │     ├── has one → viewport (spec content)
  │     └── has one → TabBar
  └── has one → SecondaryPanel
        ├── has one → LogView (regent messages)
        ├── has one → viewport (git log)
        ├── has one → viewport (tests)
        ├── has one → viewport (cost table)
        └── has one → TabBar
```

## Config Extension

```toml
[tui]
accent_color = "#7D56F4"
log_retention = 20  # keep last N session logs; 0 = unlimited
```

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| accent_color | string | "#7D56F4" | Must match `^#[0-9A-Fa-f]{6}$` |
| log_retention | int | 20 | Must be ≥ 0 |
