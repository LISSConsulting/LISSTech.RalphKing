# Implementation Plan: TUI Overhaul & UX Fixes

**Branch**: `008-tui-overhaul` | **Date**: 2026-03-07 | **Spec**: `specs/008-tui-overhaul/spec.md`
**Input**: Feature specification from `specs/008-tui-overhaul/spec.md`

## Summary

Comprehensive TUI overhaul addressing 13 user stories: fix interactive speckit commands (clarify/specify), add panel titles with numbers, convert specs to tree view, fix spec status detection for external projects, separate per-tab content buffers to prevent displacement, fix layout overflow, load iteration history on startup, show git info immediately, add lazygit-style footer detail, wire cost tracking, render markdown with glamour, add `--focus` flag for roam mode, and default focus to panel 1.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: cobra, BurntSushi/toml, bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0
**New Dependency**: `github.com/charmbracelet/glamour` (terminal markdown rendering)
**Storage**: JSONL session log (`internal/store`)
**Testing**: `go test ./...` with table-driven tests, `go vet ./...`
**Target Platform**: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
**Project Type**: CLI tool
**Constraints**: TUI must render correctly at >= 80x24; no blocking in Update/View

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | Spec 008 written and filed in `specs/008-tui-overhaul/` |
| II. Supervised Autonomy | PASS | No changes to Regent supervision model |
| III. Test-Gated Commits | PASS | All changes will have table-driven tests |
| IV. Idiomatic Go | PASS | One new dep (glamour) justified — Charm ecosystem, same author as bubbletea/lipgloss |
| V. Observable Loops | PASS | This spec improves observability (cost wiring, git info, iteration history) |

## Project Structure

### Documentation (this feature)

```text
specs/008-tui-overhaul/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (affected files)

```text
cmd/ralph/
├── commands.go          # Add --focus flag to build/run/loop commands
├── execute.go           # Wire focus to Loop struct; fix executeSpeckit()
├── speckit_cmds.go      # Fix clarify/specify to use interactive mode
└── wiring.go            # Pass git info + iteration history to TUI init

internal/loop/
├── loop.go              # Add Focus field; augmentPrompt() focus directive

internal/spec/
├── spec.go              # Fix detectDirStatus() for external projects (no change needed — already uses abs paths)

internal/tui/
├── app.go               # Default focus to FocusSpecs; load iterations on init; pass git info
├── focus.go             # No changes needed
├── layout.go            # Audit/fix off-by-one in Calculate()
├── theme.go             # Add PanelBorderStyleWithTitle() method
├── msg.go               # Add new message types (gitInfoMsg, iterationsLoadedMsg)
└── keymap.go            # No changes needed

internal/tui/components/
├── logview.go           # No changes needed (per-tab instances solve displacement)
└── tabbar.go            # No changes needed

internal/tui/panels/
├── specs.go             # Refactor to tree view with expand/collapse
├── main_view.go         # Separate LogView per tab (output, spec, iteration, summary)
├── secondary.go         # Detail view support for selected items
├── footer.go            # Expand footer for detail view rendering
├── header.go            # No changes needed (git info already wired)
├── iterations.go        # No changes needed (just needs data on init)
└── worktrees.go         # No changes needed

internal/config/
├── config.go            # Add Focus field to BuildConfig
```

**Structure Decision**: Existing `internal/` package layout preserved. No new packages. Changes are surgical modifications to existing files.

## Phase 0: Research

### R-001: Glamour Integration

**Decision**: Use `github.com/charmbracelet/glamour` for markdown rendering in the spec viewer.

**Rationale**: Glamour is the canonical terminal markdown renderer in the Charm ecosystem (same authors as bubbletea, lipgloss, bubbles). It supports:
- Dark/light theme auto-detection
- Width-aware word wrapping
- Syntax highlighting for code blocks
- ANSI output compatible with lipgloss rendering

**Alternatives considered**:
- Raw text display (current): No formatting, poor readability for specs with headers/code
- Custom markdown parser: Over-engineered for this use case
- `github.com/MichaelMure/go-term-markdown`: Less maintained, not Charm ecosystem

**Integration pattern**:
```go
import "github.com/charmbracelet/glamour"

renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(width),
)
rendered, _ := renderer.Render(markdownContent)
```

### R-002: Interactive Speckit Commands

**Decision**: Replace `claude -p "/<skill> <args>" --verbose` with `claude --verbose` using inherited stdio for `clarify` and `specify` commands. Keep `-p` for non-interactive commands (plan, tasks, etc.).

**Rationale**: The `-p` (print) flag makes Claude run non-interactively — it processes the prompt and exits. `clarify` and `specify` require back-and-forth dialogue. Without `-p`, Claude runs interactively with the terminal.

**Alternatives considered**:
- TUI modal with PTY: Too complex for v1; would require pseudo-terminal multiplexing
- Pipe questions through stdin: Claude's interactive protocol doesn't support this cleanly

**Implementation**: In `executeSpeckit()`, check if skill is "clarify" or "specify" and omit the `-p` flag. The skill name is still passed as a prompt argument but Claude stays interactive.

### R-003: Per-Tab Content Buffers (MainView)

**Decision**: Refactor `MainView` to have separate `LogView` instances per tab instead of sharing one.

**Rationale**: Currently `MainView` has a single `logview` field shared across Output/Spec/Iteration tabs. `AppendLine()` writes to this shared buffer, so new output displaces spec/iteration content. Each tab needs its own `LogView` to maintain independent content and scroll state.

**New structure**:
```go
type MainView struct {
    tabbar         components.TabBar
    outputLog      components.LogView  // Tab 0: live loop output
    specLog        components.LogView  // Tab 1: spec content viewer
    iterationLog   components.LogView  // Tab 2: iteration detail
    summaryLog     components.LogView  // Tab 3: iteration summary
    width, height  int
    activeTab      MainTab
}
```

`AppendLine()` always appends to `outputLog` regardless of active tab. `ShowSpec()` writes to `specLog`. `ShowIterationLog()` writes to `iterationLog`.

### R-004: Panel Titles with Border

**Decision**: Use lipgloss `BorderTop` with inline title text, rendered as part of the border decoration.

**Rationale**: lipgloss v1.1.0 doesn't have a built-in border title. The approach is to render the title as a styled string positioned in the top border row, replacing border characters with the title text.

**Implementation pattern**: Modify `PanelBorderStyle()` to accept a title string. Alternatively, render titles inside the panel content area as the first line (simpler, more reliable across terminal emulators).

**Decision**: Render titles as the first line inside the panel border, styled with the accent color when focused. This is simpler and avoids border rendering edge cases.

```
╭──────────────────╮
│ [1] Specs         │
│ > 008-tui-overhaul│
│   007-worktree    │
╰──────────────────╯
```

### R-005: Spec Tree View

**Decision**: Extend `SpecsPanel` to support a tree structure where each spec directory is a collapsible node with child files (spec.md, plan.md, tasks.md).

**Rationale**: Current flat list shows one entry per spec directory. Users need to access individual files within each spec (spec.md for requirements, plan.md for design, tasks.md for progress).

**Data model**:
```go
type specTreeItem struct {
    sf       spec.SpecFile  // The directory-level spec
    children []string       // File basenames: "spec.md", "plan.md", "tasks.md"
    expanded bool
}
```

The tree is rendered as a flat list where expanded items insert child rows. j/k navigates the flattened view. Enter on a directory toggles expand. Enter on a child file opens that file in the Main panel.

### R-006: Layout Audit

**Decision**: The current `Calculate()` function has a structural issue — it doesn't account for borders in the total width/height budget. Each bordered panel adds 2 to width and 2 to height. `lipgloss.JoinVertical/Horizontal` places panels side by side including their borders, so the total rendered size can exceed terminal dimensions.

**Fix**: `Calculate()` should compute panel dimensions such that outer dimensions (content + border) sum to exactly the terminal size. Currently `innerDims()` subtracts 2 from the layout rect, but the layout rect itself doesn't account for the fact that borders add to the total.

The fix is to ensure `Calculate()` distributes the full `width x height` among panels where each panel's rect represents its OUTER size (including border).

### R-007: Focus Flag

**Decision**: Add `--focus <topic>` string flag to build/run commands. Stored in `Loop.Focus`. Appended to prompt via `augmentPrompt()`.

**Directive format**: `"\n\nFocus your work on: <topic>. Prioritize changes related to this area over other improvements."`

Works orthogonally with `--roam`: roam + focus = sweep constrained to a topic.

## Phase 1: Data Model & Contracts

### Data Model

#### Loop.Focus (new field)

| Field | Type | Description |
|-------|------|-------------|
| `Focus` | `string` | Optional focus topic for roam/build narrowing |

Added to `internal/loop/loop.go` `Loop` struct.

#### BuildConfig.Focus (new config field)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Focus` | `string` | `""` | Default focus topic from ralph.toml `[build] focus = ""` |

#### MainView (refactored)

| Field | Type | Description |
|-------|------|-------------|
| `tabbar` | `TabBar` | Tab bar component |
| `outputLog` | `LogView` | Tab 0: live streaming output |
| `specLog` | `LogView` | Tab 1: spec file content |
| `iterationLog` | `LogView` | Tab 2: past iteration log |
| `summaryLog` | `LogView` | Tab 3: iteration metadata |
| `activeTab` | `MainTab` | Currently active tab |
| `width, height` | `int` | Panel dimensions |

#### specTreeNode (new, replaces specItem)

| Field | Type | Description |
|-------|------|-------------|
| `sf` | `spec.SpecFile` | Directory-level spec info |
| `children` | `[]string` | Existing files: "spec.md", "plan.md", "tasks.md" |
| `expanded` | `bool` | Whether children are visible |

#### tui.Model (new init fields)

| Field | Type | Description |
|-------|------|-------------|
| `gitBranch` | `string` | Pre-populated from git on init (existing `branch` field) |
| `gitLastCommit` | `string` | Pre-populated from git on init (existing `lastCommit` field) |

No new struct; existing fields populated earlier.

### Contracts

#### CLI Interface Changes

```
ralph build [--max N] [--roam] [--focus "topic"] [--worktree]
ralph run   [--max N] [--roam] [--focus "topic"] [--worktree]
ralph loop build [--max N] [--roam] [--focus "topic"] [--worktree]
ralph loop run   [--max N] [--roam] [--focus "topic"] [--worktree]
ralph clarify [args...]   # Now interactive (no -p flag)
ralph specify [args...]   # Now interactive (no -p flag)
```

#### Prompt Augmentation Contract

```
augmentPrompt(prompt, spec, specDir string, roam bool, focus string) string

Cases:
1. roam=true, focus="UI/UX":
   → prompt + "\n\n## Spec Context\n\nRoam mode is active. ... Focus your work on: UI/UX. ..."
2. roam=true, focus="":
   → prompt + "\n\n## Spec Context\n\nRoam mode is active. ..."  (unchanged)
3. roam=false, spec="008", focus="UI/UX":
   → prompt + "\n\n## Spec Context\n\nActive spec: 008\n... Focus your work on: UI/UX. ..."
4. roam=false, spec="", focus="":
   → prompt (unchanged)
```

#### TUI Panel Title Contract

Each panel renders a title line as the first row of content:
- Format: `[N] Title` where N is the panel number (1-5)
- Focused: accent color, bold
- Unfocused: gray, normal weight
- Titles: `[1] Specs`, `[2] Iterations`, `[3] Output`, `[4] Secondary`, `[5] Worktrees`

#### TUI Init Contract (new)

On startup, before any loop events:
1. Read `git branch --show-current` → populate header branch
2. Read `git log -1 --format=%h` → populate header last commit
3. If `storeReader` non-nil, call `storeReader.Iterations()` → populate iterations panel
4. Focus starts on `FocusSpecs` (panel 1)

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| New dep: glamour | Terminal markdown rendering for spec viewer | Raw text is unreadable for structured specs; custom parser is over-engineered |

## Implementation Phases

### Phase 1: Quick Wins (no new deps)
1. Default focus to FocusSpecs (1 line change in `app.go:102`)
2. Fix interactive speckit commands (modify `executeSpeckit()`)
3. Load iterations from store on TUI init
4. Read git info on TUI init
5. Wire `--focus` flag through commands → Loop → augmentPrompt

### Phase 2: Panel Titles & Layout
6. Add panel titles as first content row
7. Audit and fix `Calculate()` layout math
8. Separate per-tab LogView instances in MainView

### Phase 3: Spec Tree & Markdown
9. Refactor SpecsPanel to tree view
10. Add glamour dependency; render spec markdown
11. Fix spec status detection for external projects

### Phase 4: Detail View & Cost
12. Footer/secondary detail view for selected items
13. Wire cost tracking end-to-end
