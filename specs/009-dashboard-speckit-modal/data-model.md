# Data Model: Dashboard SpecKit Modal

**Branch**: `009-dashboard-speckit-modal` | **Date**: 2026-03-11

## Entities

### SpecKitModal

Tracks the modal overlay state within the TUI Model.

| Field    | Type   | Description                                    |
|----------|--------|------------------------------------------------|
| visible  | bool   | Whether the modal is currently displayed        |
| cursor   | int    | Highlighted action index (0=Plan, 1=Clarify, 2=Tasks) |
| specName | string | Display name of the target spec (shown in title) |
| specDir  | string | Absolute path to the spec directory              |
| width    | int    | Current modal render width (recalculated on resize) |
| height   | int    | Current modal render height                      |

**Validation**: `cursor` wraps at boundaries (0–2). `specName` and `specDir` must be non-empty when `visible == true`.

**State transitions**:
- Closed → Open: `S` key pressed with valid spec selected
- Open → Closed (no action): `esc` pressed
- Open → Closed (action): `enter` pressed → emits `SpecKitActionMsg`

### SpecKitAction

Enumeration of the three available actions.

| Value   | Label     | Description                          | Interactive |
|---------|-----------|--------------------------------------|-------------|
| plan    | Plan      | Generate implementation plan (plan.md) | No          |
| clarify | Clarify   | Resolve spec ambiguities via Q&A       | Yes (stdin)  |
| tasks   | Tasks     | Break down plan into task list (tasks.md) | No       |

### SpecKitRunner

Manages the lifecycle of a running SpecKit subprocess.

| Field     | Type              | Description                                    |
|-----------|-------------------|------------------------------------------------|
| action    | string            | Which action is running ("plan"/"clarify"/"tasks") |
| specName  | string            | Spec being processed (for display)              |
| specDir   | string            | Working directory for subprocess                |
| cmd       | *exec.Cmd         | Running subprocess handle                       |
| stdin     | io.WriteCloser    | Stdin pipe (used only for clarify)              |
| cancel    | context.CancelFunc| Cancels subprocess context                      |
| done      | bool              | Whether subprocess has completed                 |
| err       | error             | Subprocess error (nil on success)               |

**Lifecycle**:
1. Created → `Start()` → subprocess spawned, stdout parsing begins
2. Running → events streamed as `SpecKitOutputMsg`
3. Running (clarify) → `SpecKitInputRequestMsg` emitted when question detected
4. Running (clarify) → `WriteAnswer(string)` sends to stdin
5. Completed → `SpecKitDoneMsg` emitted, runner set to nil on Model

## Messages

| Message                  | Emitted by     | Handled by    | Purpose                          |
|--------------------------|----------------|---------------|----------------------------------|
| SpecKitActionMsg         | Modal (enter)  | app.go Update | Start a SpecKit subprocess       |
| SpecKitOutputMsg         | Runner goroutine | app.go Update | Stream output line to Output tab |
| SpecKitInputRequestMsg   | Runner goroutine | app.go Update | Switch MainView to input mode    |
| SpecKitInputResponseMsg  | MainView (enter) | app.go Update | Forward answer to runner stdin   |
| SpecKitDoneMsg           | Runner goroutine | app.go Update | Clean up runner, update state    |

## Relationships

```
Model (app.go)
├── SpecKitModal          1:1  modal overlay state
├── SpecKitRunner          0:1  nil when no action running
├── SpecsPanel.SelectedSpec()  →  provides specName + specDir to modal
└── MainView.outputLog    ←  receives SpecKitOutputMsg lines
```
