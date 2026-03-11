# Implementation Plan: Dashboard SpecKit Modal

**Branch**: `009-dashboard-speckit-modal` | **Date**: 2026-03-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-dashboard-speckit-modal/spec.md`

## Summary

Add a centered modal dialog to the TUI dashboard that lets users launch SpecKit workflows (Plan, Clarify, Tasks) against the selected spec via the global `S` key. The modal displays the target spec name for confirmation and three navigable actions. Clarify runs interactively with inline Q&A in the Output tab. Also adjusts Specs panel layout (55/45 split, inner padding).

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: bubbletea, lipgloss, bubbles (viewport, textinput), cobra
**Storage**: N/A (spec files on disk)
**Testing**: `go test ./...` — table-driven tests with `t.Run` subtests
**Target Platform**: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
**Project Type**: CLI with TUI dashboard
**Performance Goals**: Modal renders instantly (<16ms); subprocess launch <1s
**Constraints**: Minimum terminal 80×24; no new external dependencies
**Scale/Scope**: 3 new TUI components (modal, speckit runner, input overlay); 2 layout adjustments

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | Spec at `specs/009-dashboard-speckit-modal/spec.md` with clarifications complete |
| II. Supervised Autonomy | PASS | SpecKit actions are developer-initiated, not autonomous; Regent not involved |
| III. Test-Gated Commits | PASS | Table-driven tests planned for modal, layout, and runner |
| IV. Idiomatic Go | PASS | No new dependencies; stdlib + approved deps only; small focused packages |
| V. Observable Loops | PASS | SpecKit output streams to Output tab; status shown in header |

No violations. Gate passes.

## Project Structure

### Documentation (this feature)

```text
specs/009-dashboard-speckit-modal/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli-commands.md  # SpecKit modal keyboard contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/tui/
├── app.go                      # MODIFY: add modal state, global 'S' key, modal rendering in View()
├── layout.go                   # MODIFY: change sidebar split from 40/60 to 55/45
├── keymap.go                   # MODIFY: add 'S' to GlobalKeyBindings
├── msg.go                      # MODIFY: add SpecKitActionMsg, SpecKitDoneMsg, SpecKitOutputMsg
├── focus.go                    # MODIFY: add LoopState values for SpecKit execution
├── modal.go                    # NEW: modal overlay component (render, key handling, state)
├── speckit_runner.go           # NEW: SpecKit subprocess launcher + event bridge
├── panels/
│   ├── specs.go                # MODIFY: add inner padding to View()
│   └── main_view.go           # MODIFY: add input prompt mode for Clarify Q&A
└── components/
    └── logview.go              # POSSIBLY MODIFY: support input prompt at bottom

internal/tui/
├── modal_test.go               # NEW: table-driven tests for modal state/navigation/rendering
├── speckit_runner_test.go      # NEW: tests for subprocess launch + event parsing
├── layout_test.go              # MODIFY: update expected ratios
└── panels/
    ├── specs_test.go           # MODIFY: test inner padding
    └── main_view_test.go       # MODIFY: test input prompt mode
```

**Structure Decision**: All changes within the existing `internal/tui/` package. Two new files (`modal.go`, `speckit_runner.go`) follow the existing pattern of one-concern-per-file. No new packages needed.

## Design

### Modal Component (`internal/tui/modal.go`)

The modal follows the existing help overlay pattern in `app.go`:

```
┌─────────────────────────────────┐
│    SpecKit: 009-dashboard-...   │  ← spec name in title
│                                 │
│  ▸ Plan      Generate plan.md   │  ← highlighted (cursor=0)
│    Clarify   Resolve ambiguity  │
│    Tasks     Break down tasks   │
│                                 │
│         esc to cancel           │
└─────────────────────────────────┘
```

**State**:

```go
type SpecKitModal struct {
    visible  bool
    cursor   int           // 0=Plan, 1=Clarify, 2=Tasks
    specName string        // displayed in title for confirmation
    specDir  string        // passed to runner on action selection
    width    int           // recalculated on resize
    height   int
}
```

**Key handling**: When `visible == true`, the modal intercepts all keys in `app.go:handleKey()` (same pattern as `helpVisible`). Keys: `j`/`k`/up/down = navigate, `enter` = select + emit message, `esc` = close.

**Rendering**: Uses `lipgloss.Place()` for centering (same as help overlay). Bordered box with accent color. Cursor indicator (`▸`) on highlighted row.

### SpecKit Runner (`internal/tui/speckit_runner.go`)

Launches SpecKit workflows as Claude Code subprocesses. Follows the same pattern as `internal/loop/runner.go`:

- Spawns `claude -p "<speckit prompt>" --dangerously-skip-permissions` targeting the spec directory
- Streams stdout via `claude.ParseStream()` into `claude.Event` channel
- Bridges events into bubbletea messages (`SpecKitOutputMsg`) via `tea.Program.Send()`
- For Clarify: detects question prompts in output and switches MainView to input mode

**SpecKit action → prompt mapping**:

| Action  | Claude prompt                                                                    |
|---------|----------------------------------------------------------------------------------|
| Plan    | `/speckit.plan` run against the spec directory                                   |
| Clarify | `/speckit.clarify` run against the spec directory (interactive — needs stdin)     |
| Tasks   | `/speckit.tasks` run against the spec directory                                  |

**Interactive Clarify flow**:
1. Runner detects question pattern in Claude output (e.g., lines ending with "Your choice:")
2. Emits `SpecKitInputRequestMsg` → MainView shows text input prompt
3. User types answer, presses enter → `SpecKitInputResponseMsg` sent to runner
4. Runner writes answer to subprocess stdin
5. On completion → `SpecKitDoneMsg` emitted, input prompt removed

### Layout Changes (`internal/tui/layout.go`)

In `Calculate()`:
- Change `specsH = bodyH * 40 / 100` → `specsH = bodyH * 55 / 100`
- Iterations gets remaining `bodyH - specsH` (45%)

### Specs Panel Padding (`internal/tui/panels/specs.go`)

In `View()`:
- Apply 1-character horizontal padding to rendered content before returning
- Use `lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)` or manual space prefix
- Reduce effective content width by 2 to account for padding

### Message Types (`internal/tui/msg.go`)

```go
type SpecKitActionMsg struct {
    Action  string         // "plan", "clarify", "tasks"
    SpecDir string         // spec directory path
    SpecName string        // display name
}

type SpecKitOutputMsg struct {
    Line string            // rendered output line
}

type SpecKitInputRequestMsg struct{}  // signals MainView to show input prompt

type SpecKitInputResponseMsg struct {
    Answer string          // user's typed response
}

type SpecKitDoneMsg struct {
    Action string
    Err    error           // nil on success
}
```

### App State Changes (`internal/tui/app.go`)

New fields on `Model`:
```go
modal          SpecKitModal        // modal overlay state
speckitRunner  *SpecKitRunner      // nil when no action running
speckitAction  string              // current running action name (for header display)
```

**Update flow**:
1. `handleKey()`: if `modal.visible` → delegate to modal → returns `SpecKitActionMsg` on enter
2. `handleKey()`: if key is `S` and `modal.visible == false` and `speckitRunner == nil`:
   - Get `SelectedSpec()` from specsPanel
   - If nil → no-op
   - Else → open modal with spec name/dir
3. `handleKey()`: if key is `S` and `speckitRunner != nil` → show "SpecKit action in progress" in footer/status
4. On `SpecKitActionMsg` → create `SpecKitRunner`, start subprocess, update header state
5. On `SpecKitOutputMsg` → append to mainView.outputLog
6. On `SpecKitInputRequestMsg` → switch mainView to input mode
7. On `SpecKitInputResponseMsg` → forward to runner's stdin
8. On `SpecKitDoneMsg` → clear runner, update header state

**View flow**:
```go
func (m Model) View() string {
    if m.layout.TooSmall { return tooSmallMsg }
    if m.helpVisible { return m.renderHelp() }
    if m.modal.visible { return m.renderModal() }  // NEW — before panel render
    return m.renderPanels()
}
```

Wait — the modal should overlay panels, not replace them. Revised approach:

```go
func (m Model) View() string {
    if m.layout.TooSmall { return tooSmallMsg }
    if m.helpVisible { return m.renderHelp() }
    base := m.renderPanels()
    if m.modal.visible {
        return m.overlayModal(base)  // NEW — lipgloss.Place over base
    }
    return base
}
```

### LoopState Extension (`internal/tui/focus.go`)

Add SpecKit-specific states:

```go
const (
    StateSpecKit   LoopState = ...  // "SPECKIT" — running plan/clarify/tasks
)
```

With valid transitions: `StateIdle → StateSpecKit`, `StateSpecKit → StateIdle`, `StateSpecKit → StateFailed`.

The header will display: `● SPECKIT:PLAN (009-dashboard-speckit-modal)` during execution.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
