# Research: Panel-Based TUI Redesign

**Feature**: 003-tui-redesign | **Date**: 2026-02-28

## Resolved Unknowns

All unknowns were resolved during `/speckit.clarify`. No additional research was needed.

### 1. Malformed JSONL Recovery

- **Decision**: Skip malformed lines silently, log warning to stderr
- **Rationale**: Only the in-flight write at kill time can be partial — `file.Sync()` after each complete line guarantees all prior lines are intact. Silent skip + warning is the simplest approach that maintains store reliability.
- **Alternatives considered**: (a) Write repair marker to file — adds complexity, callers must understand marker format; (b) Return error to caller — unnecessarily disruptive for a single bad line among thousands of good ones.

### 2. Migration Strategy

- **Decision**: Clean replacement — delete old TUI files, write new from scratch
- **Rationale**: Current TUI is ~300 lines across 4 files with a single-panel architecture fundamentally incompatible with the multi-panel design. Incremental migration would mean maintaining two architectures simultaneously with no safety benefit given the small codebase. Git history preserves the old code.
- **Alternatives considered**: (a) Incremental refactor — adds complexity during transition with two working codepaths; (b) Parallel `tui2/` package — creates import confusion and doubles maintenance until swap.

### 3. Session Log Retention

- **Decision**: Keep last N session logs (default 20), delete oldest on startup. Configurable via `log_retention` in `ralph.toml`.
- **Rationale**: Simple count-based cap avoids unbounded disk growth. 20 sessions is enough for practical debugging. Timestamp-prefixed filenames sort chronologically, making oldest-first deletion trivial. Configurability via TOML follows existing config patterns.
- **Alternatives considered**: (a) No cleanup — unbounded growth; (b) Time-based expiry — requires parsing timestamps, more complex for same result.

### 4. Below-Minimum Terminal Behavior

- **Decision**: Show a centered "Terminal too small — resize to at least 80×24" message
- **Rationale**: Standard TUI practice (lazygit does this). Prevents rendering artifacts and broken layouts. Clear guidance for the user. Layout calculator returns `TooSmall: true` and `View()` short-circuits to the message.
- **Alternatives considered**: (a) Best-effort clipped layout — looks broken, confusing; (b) Force single-column minimal layout — still requires layout logic for small sizes, incomplete content.

### 5. LoopState Transitions

- **Decision**: Strict state machine with explicitly defined transitions
- **Rationale**: 5 states with 9 valid transitions. Invalid transitions are no-ops (logged as warnings). Makes header display deterministic and testable via table-driven tests. Prevents impossible states (e.g., `failed → regentRestart`).
- **Alternatives considered**: (a) Open transitions — no validation, risk of invalid state combinations, harder to reason about and test.

## Technology Best Practices

### charmbracelet/bubbles

- **list.Model**: Use custom `ItemDelegate` for panel rendering. Disable built-in help/filter via `list.SetShowHelp(false)`, `list.SetShowFilter(false)`. Set `list.SetShowStatusBar(false)` for minimal chrome.
- **viewport.Model**: `viewport.New(w, h)` creates viewport. Set content via `SetContent(string)`. Handles `pgup/pgdown`, mouse scroll natively. Use `GotoBottom()` for follow mode.
- **textinput.Model**: `textinput.New()` with `Focus()` to activate for spec name input.
- **spinner.Model**: Optional for loading states when reading store.

### bubbletea Sub-Model Pattern

Each panel implements its own `Update(msg) (Panel, tea.Cmd)` and `View() string`. The root model:
1. Handles global keys first (tab, 1-4, q)
2. Broadcasts `logEntryMsg` to all panels that need it
3. Dispatches remaining `tea.KeyMsg` only to the focused panel
4. Composes all `View()` outputs using lipgloss layout

### lipgloss Panel Composition

```go
// Compose panels using JoinHorizontal/JoinVertical
sidebar := lipgloss.JoinVertical(lipgloss.Left, specsView, itersView)
rightSide := lipgloss.JoinVertical(lipgloss.Left, mainView, secondaryView)
body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, rightSide)
full := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
```

Border styles: focused panel gets accent-colored border, unfocused panels get dim gray border.

### tea.Suspend for $EDITOR

```go
// In specs panel Update, on 'e' key:
cmd := tea.ExecProcess(exec.Command(editor, specPath), func(err error) tea.Msg {
    return specsRefreshedMsg{specs: refreshSpecs()}
})
return panel, cmd
```

`tea.ExecProcess` suspends the TUI, runs the command, and resumes with a callback message.
