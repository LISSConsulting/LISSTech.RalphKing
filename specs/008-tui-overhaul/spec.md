# Spec 008: TUI Overhaul & UX Fixes

## Summary

Comprehensive TUI overhaul addressing layout bugs, broken interactive commands, missing panel titles, content displacement issues, spec tree navigation, markdown rendering, cost tracking, git visibility, footer detail view, and a new `--focus` flag for roam mode.

## Problem Statement

The current TUI has accumulated multiple usability issues:
- `ralph clarify` and `ralph specify` require interactive input but Claude exits after the first question (non-interactive `-p` flag)
- Panels lack titles and number labels, making navigation unclear
- Default focus is Main (panel 3) instead of Specs (panel 1)
- Specs panel is a flat list; should be a collapsible tree showing spec.md/plan.md/tasks.md per spec
- Spec status detection fails when running ralph from a different project directory (shows "not started" for completed specs)
- TUI layout is misaligned and overflows the terminal on many screen sizes
- Opening a spec or iteration in the Main panel gets displaced by new streaming output
- Iteration tracking is inaccurate: doesn't persist/restore previous iterations from the store on startup
- Git info (branch, last commit) should always be visible, not just when events arrive
- Selecting a line in command output should show full detail in the footer (lazygit-style)
- Cost tab in Secondary panel is non-functional (no live cost accumulation, header cost stays $0.00)
- Spec content is rendered as raw text; should use glamour for markdown rendering
- Roam mode lacks focus/narrowing: need `--focus "topic"` flag to constrain roam to a specific area

## User Stories

### US-001: Interactive Speckit Commands in TUI
As a user, I want `clarify` and `specify` to work as interactive sessions rendered in the TUI as a modal/overlay, so I can answer Claude's questions without the process exiting.

### US-002: Panel Titles and Numbers
As a user, I want each panel to display a title (e.g., "Specs", "Iterations", "Output", "Secondary") and its keybinding number (1-4/5), so I can see at a glance which panel is which and how to jump to it.

### US-003: Specs Panel as Traversable Tree
As a user, I want the Specs panel to show a tree structure where each spec directory is expandable, revealing spec.md, plan.md, and tasks.md as child nodes that I can select individually.

### US-004: Correct Spec Status Detection
As a user running ralph from a different project directory, I want spec statuses to correctly reflect their actual state (completed vs not started) based on file content analysis, not just file existence.

### US-005: Stable Content Panels (No Displacement)
As a user, when I open a spec or iteration log in the Main panel, I want it to stay visible in its own tab without being displaced by new streaming output. Each tab should maintain independent content and scroll position.

### US-006: Layout Correctness
As a user, I want the TUI to render correctly within my terminal dimensions without overflow, misalignment, or off-screen content.

### US-007: Persistent Iteration History
As a user, I want the Iterations panel to load previous iterations from the session store on startup, so I can see the full history, not just iterations from the current session.

### US-008: Always-Visible Git Info
As a user, I want the header/footer to always display the current git branch and last commit hash (read from git on startup), not just when log events provide them.

### US-009: Footer Detail View (lazygit-style)
As a user, when I select/highlight a line in the command output or iteration list, I want the footer/secondary area to show full detail for that item (commit details, full log entry, etc.), similar to how lazygit shows commit details.

### US-010: Functional Cost Tracking
As a user, I want the Cost tab to show a live-updating per-iteration cost breakdown and the header's cost display to reflect the running total accurately.

### US-011: Markdown Rendering with Glamour
As a user, I want spec content displayed in the Main panel to be rendered as formatted markdown (syntax highlighting, headers, lists) using glamour, not raw text.

### US-012: Roam Focus Flag
As a user, I want `ralph build --roam --focus "UI/UX"` to constrain roam mode to a specific domain/area/subject, making Claude focus on that topic rather than wandering freely.

### US-013: Default Focus on Panel 1
As a user, I want the TUI to start with focus on the Specs panel (panel 1), not the Main panel.

## Functional Requirements

### FR-001: Interactive Speckit Modal
- `ralph clarify` and `ralph specify` must NOT use `claude -p` (print mode); they need interactive stdio
- When launched from the TUI dashboard (future), render as a modal overlay that captures input
- When launched from CLI, use `claude` without `-p` flag, with inherited stdio for full interactivity
- The modal should show Claude's questions and accept user input inline

### FR-002: Panel Titles
- Each bordered panel must render a title in its top border: "[1] Specs", "[2] Iterations", "[3] Output", "[4] Secondary", "[5] Worktrees"
- The focused panel's title should use the accent color
- Titles should be rendered using lipgloss border title functionality or custom top-border rendering

### FR-003: Spec Tree View
- Each spec directory (specs/NNN-name/) is a tree node
- Expanding a node reveals child files: spec.md, plan.md, tasks.md (only files that exist)
- Enter on a directory node toggles expand/collapse
- Enter on a file node opens that file's content in the Main panel
- j/k navigates the flattened tree (expanded children count as rows)

### FR-004: Spec Status Detection Fix
- `spec.List()` / `detectDirStatus()` must work correctly when `workDir` is a different project
- Ensure file paths are resolved relative to the target project's spec directory, not CWD
- Add integration test: run `spec.List("/path/to/other/project")` and verify correct statuses

### FR-005: Independent Tab Content
- MainView must maintain separate content buffers per tab (Output, Spec, Iteration, Summary)
- Appending to the Output tab must NOT affect the Spec or Iteration tab content
- Currently `logview` is shared across tabs; refactor to one LogView per tab
- Each tab retains its scroll position independently

### FR-006: Layout Fix
- Audit `Calculate()` for off-by-one errors in height/width distribution
- Ensure `lipgloss.JoinVertical/Horizontal` output fits exactly within `width x height`
- Test with common terminal sizes: 80x24, 120x40, 200x60, and verify no overflow
- Account for border characters (2 per dimension) consistently

### FR-007: Load Iterations on Startup
- On TUI init, if `storeReader` is non-nil, call `storeReader.Iterations()` to load past summaries
- Populate `iterationsPanel` with historical data before first event arrives
- Mark none as "running" initially (all are completed)

### FR-008: Git Info on Startup
- On TUI init or first `WindowSizeMsg`, read current branch via `git branch --show-current`
- Read last commit via `git log -1 --format=%h`
- Populate `m.branch` and `m.lastCommit` before any loop events
- These serve as defaults; loop events override them as they arrive

### FR-009: Footer Detail Pane
- When the user highlights a line in the Output logview (cursor mode), show expanded detail in the Secondary panel's active tab or a dedicated "Detail" view
- For iteration list selections: show full iteration summary (cost, duration, commit, subtype, timestamps)
- For git entries: show full commit message, diff stat
- Model after lazygit's bottom detail pane

### FR-010: Cost Wiring
- Verify `entry.CostUSD` and `entry.TotalCost` are populated by the loop's LogIterComplete events
- Ensure `m.totalCost` accumulates correctly and feeds the header's `cost: $X.XX`
- Ensure `secondary.AddIteration()` receives summaries so the Cost tab table populates
- Add a running total row that updates in real time

### FR-011: Glamour Markdown Rendering
- Add `github.com/charmbracelet/glamour` as an approved dependency
- In `handleSpecSelected`, render spec content through glamour before displaying
- Use a dark theme (or auto-detect) with terminal-width-aware wrapping
- Fallback to raw text if glamour rendering fails

### FR-012: --focus Flag
- Add `--focus <topic>` flag to `ralph build` and `ralph run` commands
- Store as `Loop.Focus string`
- When non-empty, append a focus directive to the prompt: "Focus your work on: <topic>. Prioritize changes related to this area."
- Works with or without `--roam`; when combined with roam, constrains the sweep to the topic
- Config file support: `[build] focus = ""`

### FR-013: Default Focus Panel 1
- Change `New()` to initialize `focus: FocusSpecs` instead of `focus: FocusMain`

## Non-Functional Requirements

- NFR-001: All changes must pass `go vet ./...` with zero warnings
- NFR-002: New code must have table-driven tests with `t.Run` subtests
- NFR-003: No new dependencies beyond glamour (which must be justified and added to approved list)
- NFR-004: Layout must render correctly on terminals >= 80x24
- NFR-005: TUI must remain responsive (no blocking calls in Update/View)

## Acceptance Criteria

- AC-001: `ralph clarify` and `ralph specify` complete a full interactive session without premature exit
- AC-002: All panels show "[N] Title" in their border, accent-colored when focused
- AC-003: Specs panel renders as expandable tree; enter on file opens markdown-rendered content
- AC-004: Running ralph against a project with completed specs shows correct status symbols
- AC-005: Opening a spec in Main panel stays visible while Output tab continues receiving new lines
- AC-006: TUI renders within bounds at 80x24, 120x40, 200x60 with no clipping
- AC-007: Previous iterations appear in the panel immediately on TUI startup
- AC-008: Branch and last commit visible in header from the moment TUI opens
- AC-009: Selecting an iteration or output line shows detail in footer/secondary area
- AC-010: Cost tab shows per-iteration breakdown; header shows running total
- AC-011: Spec content renders with markdown formatting (headers, bold, code blocks)
- AC-012: `ralph build --roam --focus "UI/UX"` passes focus string to Claude prompt
- AC-013: TUI starts with panel 1 (Specs) focused

## Dependencies

- `github.com/charmbracelet/glamour` — terminal markdown rendering (new approved dep)

## Out of Scope

- Full TUI redesign / new panel types beyond what's listed
- Mobile/web rendering
- Spec editing within the TUI (beyond $EDITOR delegation)
- Real-time cost from Claude API (we use post-iteration cost from log events)
