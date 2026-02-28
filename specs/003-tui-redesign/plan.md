# Implementation Plan: Panel-Based TUI Redesign

**Branch**: `003-tui-redesign` | **Date**: 2026-02-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-tui-redesign/spec.md`

## Summary

Replace the single-panel log viewer TUI (`internal/tui/`) with a multi-panel, keyboard-driven dashboard inspired by lazygit/lazydocker. The new TUI uses bubbletea's Elm architecture with composed sub-models for four panels (specs, iterations, main view, secondary view), a responsive layout calculator, and a focus management system. The existing `internal/store/` JSONL session log (already implemented) is extended with log retention and malformed-line recovery. The sole new dependency is `charmbracelet/bubbles` for battle-tested viewport, list, textinput, and spinner components.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: `bubbletea` v1.3.10, `lipgloss` v1.1.0, `cobra` v1.10.2, `BurntSushi/toml` v1.6.0, **NEW**: `charmbracelet/bubbles`
**Storage**: Append-only JSONL files in `.ralph/logs/` (existing `internal/store/` package)
**Testing**: `go test ./...` with table-driven subtests; target ≥80% coverage
**Target Platform**: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
**Project Type**: CLI tool
**Performance Goals**: Store write latency <5ms per event (SC-010); past iteration access <1s on launch (SC-009)
**Constraints**: bubbletea + lipgloss only (no gocui, no raw ANSI); minimum terminal 80×24; binary size increase ≤2MB
**Scale/Scope**: Single-user local CLI; session logs of ~100s of iterations; ~20 source files changed/created

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | ✅ PASS | `specs/003-tui-redesign/spec.md` drives all implementation. Clarifications recorded. |
| II. Supervised Autonomy | ✅ PASS | Regent integration preserved. Secondary panel surfaces all Regent activity. JSONL store survives Regent kill/restart cycles. |
| III. Test-Gated Commits | ✅ PASS | Table-driven tests planned for every new package/file. ≥80% coverage target for `tui/`, `store/`. |
| IV. Idiomatic Go | ✅ PASS | bubbletea + lipgloss only. `bubbles` is the sole new dep (official Charm library). Store uses stdlib only. Small packages, clear interfaces. |
| V. Observable Loops | ✅ PASS | Every `LogKind` renders. Header shows state at a glance. Regent has dedicated tab. Session log makes all iteration output reviewable. |

No violations. No complexity justification needed.

## Project Structure

### Documentation (this feature)

```text
specs/003-tui-redesign/
├── plan.md              # This file
├── spec.md              # Feature specification (done)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/store/
├── store.go             # Store/Writer/Reader interfaces (EXISTS — unchanged)
├── store_test.go        # (EXISTS — extend with retention + recovery tests)
├── jsonl.go             # JSONL implementation (EXISTS — add retention + malformed-line skip)
├── jsonl_test.go        # (EXISTS — extend)
└── index.go             # In-memory byte-offset index (EXISTS — unchanged)

internal/tui/            # CLEAN REPLACEMENT — all existing files deleted
├── app.go               # Root Model: Init/Update/View, panel composition, focus management
├── app_test.go
├── focus.go             # FocusTarget enum, focus cycling logic
├── focus_test.go
├── layout.go            # Responsive layout calculator: (w,h) → panel rects
├── layout_test.go
├── keymap.go            # Global keybindings + per-panel dispatch table
├── keymap_test.go
├── theme.go             # Colors, border styles, accent (carries forward styles.go palette)
├── theme_test.go
├── msg.go               # Custom tea.Msg types shared across panels
├── panels/
│   ├── specs.go         # Specs list panel: model, update, view, keybindings
│   ├── specs_test.go
│   ├── iterations.go    # Iterations list panel: model, update, view
│   ├── iterations_test.go
│   ├── main_view.go     # Main panel: tabbed content (output, spec, iteration detail)
│   ├── main_view_test.go
│   ├── secondary.go     # Secondary panel: tabbed content (regent, git, tests, cost)
│   ├── secondary_test.go
│   ├── header.go        # Header bar renderer (stateless — receives props)
│   ├── header_test.go
│   ├── footer.go        # Footer bar renderer (context-sensitive hints)
│   └── footer_test.go
└── components/
    ├── logview.go       # Scrollable log viewport with follow mode (wraps bubbles/viewport)
    ├── logview_test.go
    ├── tabbar.go        # Tab bar component: renders tab titles, handles [/] switching
    └── tabbar_test.go

cmd/ralph/
├── wiring.go            # MODIFY — add store.Reader to TUI constructor, retain all 4 paths
├── execute.go           # MODIFY — add log_retention config passing to store
└── main.go              # MODIFY — if dashboard mode added (P3)

internal/config/
└── config.go            # MODIFY — add LogRetention field to TUIConfig (or new StoreConfig)
```

**Structure Decision**: Existing Go package layout preserved. The `internal/tui/` package is replaced in-place (clean replacement per clarification). No new top-level packages. The `internal/tui/panels/` and `internal/tui/components/` sub-packages are new, following the spec's package structure. The `internal/store/` package is extended minimally.

## Phase 0: Research

### Research Tasks

No NEEDS CLARIFICATION items remain — all unknowns were resolved during `/speckit.clarify`:

1. **Malformed JSONL recovery** → Skip silently, log warning to stderr
2. **Migration strategy** → Clean replacement, delete old files
3. **Session log retention** → Keep last N (default 20), configurable via `ralph.toml`
4. **Below-minimum terminal** → Show centered resize message
5. **LoopState transitions** → Strict state machine with defined transition table

### Technology Best Practices

| Technology | Best Practice | Source |
|------------|---------------|--------|
| `bubbles/viewport` | Use `viewport.New(w, h)`, set content via `SetContent()`, handle `viewport.Update()` for scroll keys | Charm docs |
| `bubbles/list` | Use `list.New(items, delegate, w, h)` with custom `list.ItemDelegate` for rendering. Disable built-in filter/help for panels | Charm docs |
| `bubbles/textinput` | Use for spec name input (US2-AC3). `textinput.New()` with `Focus()` to activate | Charm docs |
| bubbletea sub-models | Each panel is a `tea.Model` with its own `Update`/`View`. Root model dispatches to focused panel. Use custom `tea.Msg` types for inter-panel communication | bubbletea examples |
| lipgloss layout | Use `lipgloss.Place()` for centering. Use `lipgloss.JoinHorizontal`/`JoinVertical` for panel composition. Set border styles per panel | lipgloss docs |
| `tea.Suspend` | Call `tea.ExecProcess()` to launch `$EDITOR` and resume after exit (US2-AC2) | bubbletea docs |

### Decision Log

| Decision | Rationale | Alternatives Rejected |
|----------|-----------|----------------------|
| Use `bubbles/list` for specs and iterations panels | Battle-tested, handles selection, filtering, scrolling. Avoids reimplementation | Custom list from scratch — more work, less reliable |
| Use `bubbles/viewport` for main panel output | Handles content scrolling, page up/down, line-by-line navigation | Custom scroll logic — already exists in current TUI but doesn't support multi-panel |
| Stateless header/footer (pure render functions, not tea.Model) | Header and footer don't process keys or maintain state — they just render props. Making them tea.Models adds unnecessary complexity | Sub-model pattern — overkill for display-only components |
| Single `theme.go` replaces `styles.go` | Consolidates all colors, borders, and tool-specific styles. Carries forward the exact same color palette | Scattered styles per panel — harder to maintain consistency |
| `FocusTarget` enum (not interface) | Only 4 fixed panels. An interface is over-abstraction for a closed set | Interface-based focus — dynamic dispatch unnecessary for 4 panels |
| `LoopState` as strict state machine in `focus.go` (or `state.go`) | Prevents invalid transitions, makes header indicator deterministic, testable via table-driven tests | Open enum — no validation, harder to test, potential for invalid states |

## Phase 1: Design

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Root Model (app.go)                       │
│  Fields: focus, width, height, loopState, specs, iterations,    │
│          mainView, secondary, header props, store.Reader         │
│                                                                  │
│  Init() → batch: waitForEvent, tickCmd, loadSpecs               │
│  Update(msg) →                                                   │
│    1. Global keys (tab, 1-4, q, ?) → handled first              │
│    2. logEntryMsg → broadcast to all panels + update loopState   │
│    3. tea.KeyMsg → dispatch to focused panel                     │
│    4. tea.WindowSizeMsg → recalc layout, resize all panels       │
│  View() →                                                        │
│    1. If below 80×24 → centered resize message                  │
│    2. Else → compose: header + (sidebar | main+secondary) + footer│
└──────────┬────────────┬────────────┬────────────┬───────────────┘
           │            │            │            │
     ┌─────▼─────┐ ┌───▼────┐ ┌────▼─────┐ ┌───▼──────┐
     │  Specs    │ │Iters   │ │MainView  │ │Secondary │
     │  Panel    │ │Panel   │ │  Panel   │ │  Panel   │
     │(list.Model│ │(list)  │ │(viewport │ │(viewport │
     │ +delegate)│ │        │ │ +tabbar) │ │ +tabbar) │
     └───────────┘ └────────┘ └──────────┘ └──────────┘
```

### Component Contracts

#### Root Model (`app.go`)

```go
type Model struct {
    // Sub-models
    specs      panels.SpecsPanel
    iterations panels.IterationsPanel
    mainView   panels.MainView
    secondary  panels.SecondaryPanel

    // Layout
    width, height int
    focus         FocusTarget

    // Loop state (strict state machine)
    loopState LoopState

    // Event source
    events <-chan loop.LogEntry
    store  store.Reader

    // Header props (extracted from events)
    projectName string
    workDir     string
    branch      string
    mode        string
    iteration   int
    maxIter     int
    totalCost   float64
    lastCommit  string

    // Graceful stop
    requestStop   func()
    stopRequested bool

    // Time
    startedAt time.Time
    now       time.Time

    // Terminal
    done bool
    err  error
}

func New(events <-chan loop.LogEntry, storeReader store.Reader,
         accentColor, projectName, workDir string,
         specFiles []spec.SpecFile, requestStop func()) Model
```

**Signature change**: `tui.New()` gains `store.Reader` and `[]spec.SpecFile` parameters. `wiring.go` creates the store, discovers specs, and passes both.

#### FocusTarget (`focus.go`)

```go
type FocusTarget int

const (
    FocusSpecs FocusTarget = iota
    FocusIterations
    FocusMain
    FocusSecondary
)

func (f FocusTarget) Next() FocusTarget  // tab: cycles forward
func (f FocusTarget) Prev() FocusTarget  // shift+tab: cycles backward
func (f FocusTarget) String() string     // "specs", "iterations", "main", "secondary"
```

#### LoopState (`focus.go`)

```go
type LoopState int

const (
    StateIdle LoopState = iota
    StatePlanning
    StateBuilding
    StateFailed
    StateRegentRestart
)

func (s LoopState) CanTransitionTo(next LoopState) bool
func (s LoopState) Label() string   // "IDLE", "PLANNING", "BUILDING", "FAILED", "REGENT RESTART"
func (s LoopState) Symbol() string  // "✓", "●", "●", "✗", "⟳"
```

#### Layout Calculator (`layout.go`)

```go
type Rect struct {
    X, Y, Width, Height int
}

type Layout struct {
    Header, Footer         Rect
    Specs, Iterations      Rect
    Main, Secondary        Rect
    TooSmall               bool // true when terminal < 80×24
}

func Calculate(width, height int) Layout
```

Pure function, no side effects. Extensively table-tested.

#### Specs Panel (`panels/specs.go`)

```go
type SpecsPanel struct {
    list   list.Model       // bubbles/list
    specs  []spec.SpecFile
    width  int
    height int
}

func NewSpecsPanel(specs []spec.SpecFile, w, h int) SpecsPanel
func (p SpecsPanel) Update(msg tea.Msg) (SpecsPanel, tea.Cmd)
func (p SpecsPanel) View() string
func (p SpecsPanel) SelectedSpec() *spec.SpecFile
func (p SpecsPanel) SetSize(w, h int) SpecsPanel
```

Emits `SpecSelectedMsg{spec.SpecFile}` when selection changes. Root model passes this to MainView to show spec content.

#### Iterations Panel (`panels/iterations.go`)

```go
type IterationsPanel struct {
    list       list.Model
    iterations []store.IterationSummary
    current    *int            // currently running iteration number (nil if idle)
    width      int
    height     int
}

func NewIterationsPanel(w, h int) IterationsPanel
func (p IterationsPanel) Update(msg tea.Msg) (IterationsPanel, tea.Cmd)
func (p IterationsPanel) View() string
func (p IterationsPanel) AddIteration(s store.IterationSummary) IterationsPanel
func (p IterationsPanel) SetCurrent(n int) IterationsPanel
func (p IterationsPanel) SelectedIteration() *store.IterationSummary
func (p IterationsPanel) SetSize(w, h int) IterationsPanel
```

Emits `IterationSelectedMsg{Number int}` when selection changes. Root model uses `store.Reader.IterationLog(n)` to load full log and passes to MainView.

#### Main View Panel (`panels/main_view.go`)

```go
type MainView struct {
    tabbar     components.TabBar
    logview    components.LogView   // live output + past iteration output
    specView   viewport.Model       // spec content viewer
    width      int
    height     int
    activeTab  MainTab
}

type MainTab int
const (
    TabOutput MainTab = iota
    TabSpecContent
    TabIterationDetail
)

func NewMainView(w, h int) MainView
func (v MainView) Update(msg tea.Msg) (MainView, tea.Cmd)
func (v MainView) View() string
func (v MainView) AppendLogEntry(entry loop.LogEntry) MainView
func (v MainView) ShowSpec(content string) MainView
func (v MainView) ShowIterationLog(entries []loop.LogEntry) MainView
func (v MainView) SetSize(w, h int) MainView
```

#### Secondary Panel (`panels/secondary.go`)

```go
type SecondaryPanel struct {
    tabbar    components.TabBar
    regent    components.LogView  // Regent messages
    gitLog    viewport.Model      // git operations
    tests     viewport.Model      // test output
    cost      viewport.Model      // cost table
    width     int
    height    int
    activeTab SecondaryTab
}

type SecondaryTab int
const (
    TabRegent SecondaryTab = iota
    TabGit
    TabTests
    TabCost
)

func NewSecondaryPanel(w, h int) SecondaryPanel
func (p SecondaryPanel) Update(msg tea.Msg) (SecondaryPanel, tea.Cmd)
func (p SecondaryPanel) View() string
func (p SecondaryPanel) AppendEntry(entry loop.LogEntry) SecondaryPanel
func (p SecondaryPanel) SetSize(w, h int) SecondaryPanel
```

#### LogView Component (`components/logview.go`)

```go
type LogView struct {
    viewport viewport.Model
    lines    []string        // rendered lines
    follow   bool            // auto-scroll to bottom
    width    int
    height   int
}

func NewLogView(w, h int) LogView
func (v LogView) Update(msg tea.Msg) (LogView, tea.Cmd)
func (v LogView) View() string
func (v LogView) AppendLine(rendered string) LogView
func (v LogView) SetContent(lines []string) LogView
func (v LogView) ToggleFollow() LogView
func (v LogView) SetSize(w, h int) LogView
```

Wraps `bubbles/viewport`. When `follow=true`, auto-scrolls to bottom on new content. `f` key toggles follow mode.

#### TabBar Component (`components/tabbar.go`)

```go
type TabBar struct {
    tabs     []string
    active   int
    width    int
}

func NewTabBar(tabs []string) TabBar
func (t TabBar) View() string
func (t TabBar) Next() TabBar       // ] key
func (t TabBar) Prev() TabBar       // [ key
func (t TabBar) Active() int
func (t TabBar) SetWidth(w int) TabBar
```

Pure renderer. Tab titles are rendered with the active tab highlighted (accent color, bold). Inactive tabs are dimmed.

#### Header (`panels/header.go`)

```go
type HeaderProps struct {
    ProjectName string
    WorkDir     string
    Branch      string
    Mode        string
    Iteration   int
    MaxIter     int
    TotalCost   float64
    State       LoopState
    Elapsed     time.Duration
    Clock       time.Time
}

func RenderHeader(props HeaderProps, width int, accentStyle lipgloss.Style) string
```

Stateless pure function — not a `tea.Model`. Carries forward the existing header layout with the addition of the `LoopState` symbol/label.

#### Footer (`panels/footer.go`)

```go
type FooterProps struct {
    Focus         FocusTarget
    LastCommit    string
    StopRequested bool
    LoopState     LoopState
    ScrollOffset  int
    NewBelow      int
}

func RenderFooter(props FooterProps, width int) string
```

Context-sensitive keybinding hints change per focused panel:
- **Specs focused**: `j/k:navigate  e:edit  n:new  enter:view  tab:next panel`
- **Iterations focused**: `j/k:navigate  enter:view  tab:next panel`
- **Main focused**: `f:follow  [/]:tab  ctrl+u/d:scroll  tab:next panel`
- **Secondary focused**: `[/]:tab  j/k:scroll  tab:next panel`
- **Global** (always shown): `q:quit  1-4:panel  s:stop`

#### Custom Messages (`msg.go`)

```go
// logEntryMsg wraps a LogEntry for broadcasting to all panels.
type logEntryMsg loop.LogEntry

// loopDoneMsg signals the event channel closed.
type loopDoneMsg struct{}

// tickMsg is sent every second for the clock.
type tickMsg time.Time

// specSelectedMsg is emitted by the specs panel on selection change.
type specSelectedMsg struct{ Spec spec.SpecFile }

// iterationSelectedMsg is emitted by the iterations panel on selection change.
type iterationSelectedMsg struct{ Number int }

// iterationLogLoadedMsg carries loaded iteration log data.
type iterationLogLoadedMsg struct {
    Number  int
    Entries []loop.LogEntry
    Err     error
}

// specsRefreshedMsg carries refreshed spec list after creation/edit.
type specsRefreshedMsg struct{ Specs []spec.SpecFile }

// loopStateTransitionMsg requests a state transition.
type loopStateTransitionMsg struct{ To LoopState }
```

### Store Extensions

#### Retention (`jsonl.go`)

```go
// EnforceRetention removes the oldest session log files in dir, keeping
// at most maxKeep files. Called once on store creation (NewJSONL).
func EnforceRetention(dir string, maxKeep int) error
```

Implementation: `os.ReadDir(dir)`, filter `*.jsonl`, sort by name (timestamp-prefixed → chronological), remove `len - maxKeep` oldest.

#### Malformed Line Recovery (`jsonl.go`)

In `IterationLog()`, change `json.Unmarshal` error handling from `return nil, err` to `log.Printf("store: skipping malformed line: %v", err); continue`.

#### Config Extension (`config.go`)

```go
type TUIConfig struct {
    AccentColor  string `toml:"accent_color"`
    LogRetention int    `toml:"log_retention"` // default 20; 0 = unlimited
}
```

### Wiring Changes (`wiring.go`)

All four wiring functions (`runWithRegent`, `runWithRegentTUI`, `runWithStateTracking`, `runWithTUIAndState`) already pass `store.Writer` to the forwarding goroutine. Changes:

1. **TUI constructor**: `tui.New()` gains `store.Reader` parameter. The `store.Store` returned by `NewJSONL` satisfies both `Writer` and `Reader`, so pass the same instance.
2. **Spec discovery**: Call `spec.List(dir)` before creating the TUI model, pass result to `tui.New()`.
3. **Retention**: `execute.go` passes `cfg.TUI.LogRetention` to `store.EnforceRetention()` after creating the store.

### Implementation Phases (for task breakdown)

**Phase A: Foundation** (no visual output yet)
1. Add `bubbles` dependency
2. Extend store: retention, malformed-line recovery
3. Add `log_retention` config field
4. Create `focus.go` + `LoopState` with tests
5. Create `layout.go` with tests
6. Create `theme.go` (carry forward palette from `styles.go`)
7. Create `msg.go` (message types)

**Phase B: Components**
8. Create `components/tabbar.go` with tests
9. Create `components/logview.go` with tests

**Phase C: Panels**
10. Create `panels/header.go` with tests
11. Create `panels/footer.go` with tests
12. Create `panels/specs.go` with tests
13. Create `panels/iterations.go` with tests
14. Create `panels/main_view.go` with tests
15. Create `panels/secondary.go` with tests

**Phase D: Integration**
16. Create `app.go` — root model composing all panels
17. Delete old TUI files (`model.go`, `update.go`, `view.go`, `styles.go`, `model_test.go`)
18. Update `wiring.go` — new `tui.New()` signature, pass `store.Reader` + specs
19. Update `execute.go` — pass retention config
20. Integration tests: keyboard navigation, event flow, layout

**Phase E: Polish (P2/P3 features)**
21. Spec content rendering in main panel (P2)
22. Iteration drill-down from store (P2)
23. `$EDITOR` integration via `tea.Suspend` (P2)
24. Loop control keybindings b/x/R (P3)
25. Secondary panel tabs: git, tests, cost (P3)

### Post-Design Constitution Re-Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | ✅ PASS | Every component traces to a spec requirement (FR-001 through FR-023). |
| II. Supervised Autonomy | ✅ PASS | Regent wiring unchanged. Secondary panel has dedicated Regent tab. Store persists across restarts. |
| III. Test-Gated Commits | ✅ PASS | Every new file has a `_test.go` companion. Layout, focus, state machine, tabbar, logview all table-tested. |
| IV. Idiomatic Go | ✅ PASS | `bubbles` is the sole new dep. Sub-packages follow single-responsibility. Interfaces used at boundaries (store.Reader). No circular imports (panels → components, app → panels). |
| V. Observable Loops | ✅ PASS | All LogKind types rendered via theme.go. Header shows LoopState at a glance. Session log makes history reviewable. |

## Complexity Tracking

> No violations found. No justification needed.
