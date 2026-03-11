# CLI Contract: SpecKit Modal Keyboard Commands

**Branch**: `009-dashboard-speckit-modal` | **Date**: 2026-03-11

## Global Key Bindings (added)

| Key | Context           | Action                                           |
|-----|-------------------|--------------------------------------------------|
| `S` | Any panel focused, spec selected, no SpecKit running | Open SpecKit actions modal |
| `S` | Any panel focused, SpecKit action running         | Show "SpecKit action in progress" status message  |
| `S` | Any panel focused, no spec selected               | No-op                                            |

## Modal Key Bindings (when modal is open)

| Key         | Action                                                |
|-------------|-------------------------------------------------------|
| `j` / `↓`  | Move cursor down (wraps to top at bottom)             |
| `k` / `↑`  | Move cursor up (wraps to bottom at top)               |
| `enter`     | Select highlighted action → close modal → launch action |
| `esc`       | Close modal without action                             |

All other keys are absorbed (no leak to panels).

## Modal Actions

| Index | Label     | Description              | Subprocess prompt          |
|-------|-----------|--------------------------|----------------------------|
| 0     | Plan      | Generate plan.md         | `/speckit.plan`            |
| 1     | Clarify   | Resolve spec ambiguities | `/speckit.clarify`         |
| 2     | Tasks     | Break down into tasks    | `/speckit.tasks`           |

## Input Prompt (Clarify only)

When Clarify is running and a question is detected:

| Key     | Action                              |
|---------|-------------------------------------|
| typing  | Characters appear in input prompt   |
| `enter` | Submit answer to Clarify subprocess |
| `esc`   | Cancel Clarify execution            |

## Header Status Display

| State             | Display format                                       |
|-------------------|------------------------------------------------------|
| SpecKit running   | `● SPECKIT:PLAN (009-feature-name)`                  |
| SpecKit complete  | Returns to previous state (IDLE/BUILDING/etc.)        |
| SpecKit failed    | `✗ SPECKIT:PLAN FAILED` (clears after 5s or next action) |
