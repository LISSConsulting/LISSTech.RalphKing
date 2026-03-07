# Data Model: Spec 008 — TUI Overhaul

## Modified Entities

### Loop (internal/loop/loop.go)

```go
type Loop struct {
    // ... existing fields ...
    Focus string // NEW: optional focus topic for prompt narrowing
}
```

### BuildConfig (internal/config/config.go)

```go
type BuildConfig struct {
    // ... existing fields ...
    Focus string `toml:"focus"` // NEW: default focus topic
}
```

TOML: `[build] focus = ""`

### MainView (internal/tui/panels/main_view.go)

```go
// BEFORE (shared logview):
type MainView struct {
    tabbar         components.TabBar
    logview        components.LogView     // shared across tabs
    summaryLogview components.LogView
    width, height  int
    activeTab      MainTab
}

// AFTER (per-tab logviews):
type MainView struct {
    tabbar       components.TabBar
    outputLog    components.LogView  // Tab 0: live loop output
    specLog      components.LogView  // Tab 1: spec content
    iterationLog components.LogView  // Tab 2: iteration detail
    summaryLog   components.LogView  // Tab 3: iteration summary
    width, height int
    activeTab    MainTab
}
```

### SpecsPanel (internal/tui/panels/specs.go)

```go
// BEFORE: flat list of specItem via bubbles/list
// AFTER: custom tree with expand/collapse

type specTreeNode struct {
    sf       spec.SpecFile
    children []string  // basenames: "spec.md", "plan.md", "tasks.md"
    expanded bool
}

type SpecsPanel struct {
    nodes       []specTreeNode
    cursor      int    // position in flattened view
    width       int
    height      int
    scrollOffset int
    input       textinput.Model
    inputActive bool
}
```

### tui.Model (internal/tui/app.go)

No new fields. Existing fields populated differently:

| Field | Before | After |
|-------|--------|-------|
| `focus` | `FocusMain` | `FocusSpecs` |
| `branch` | `""` (wait for events) | Pre-populated from `git branch --show-current` |
| `lastCommit` | `""` (wait for events) | Pre-populated from `git log -1 --format=%h` |
| `iterationsPanel` | Empty | Pre-loaded from `storeReader.Iterations()` |

## New Message Types (internal/tui/msg.go)

```go
// gitInfoMsg carries pre-read git state for init.
type gitInfoMsg struct {
    Branch     string
    LastCommit string
}

// iterationsLoadedMsg carries historical iterations loaded from store.
type iterationsLoadedMsg struct {
    Summaries []store.IterationSummary
}
```

## Function Signature Changes

### augmentPrompt (internal/loop/loop.go)

```go
// BEFORE:
func augmentPrompt(prompt, spec, specDir string, roam bool) string

// AFTER:
func augmentPrompt(prompt, spec, specDir string, roam bool, focus string) string
```

### executeSpeckit (cmd/ralph/execute.go)

```go
// BEFORE:
func executeSpeckit(ctx context.Context, skill string, args []string) error

// AFTER:
func executeSpeckit(ctx context.Context, skill string, args []string, interactive bool) error
```

### tui.New (internal/tui/app.go)

No signature change. Git info and iterations loaded via Init() commands.

## State Transitions

### Spec Tree Navigation

```
State: cursor on directory node
  Enter → expanded = !expanded (toggle)

State: cursor on child file node
  Enter → emit SpecSelectedMsg with file path

State: cursor anywhere
  j/down → cursor++, skip past collapsed children
  k/up   → cursor--, skip past collapsed children
```

### Tab Independence (MainView)

```
Event: logEntryMsg arrives
  → Always append to outputLog (regardless of activeTab)

Event: SpecSelectedMsg
  → Set specLog content, switch to TabSpecContent

Event: IterationSelectedMsg
  → Set iterationLog content, switch to TabIterationDetail

User: press [ or ]
  → Switch activeTab, View() renders the corresponding LogView
  → Each LogView retains its own scroll position
```
