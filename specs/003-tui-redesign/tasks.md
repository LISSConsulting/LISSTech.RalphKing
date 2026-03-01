# Tasks: Panel-Based TUI Redesign

**Input**: Design documents from `/specs/003-tui-redesign/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included per constitution principle III (test-gated commits) and spec target ≥80% coverage. Each implementation task includes its companion `_test.go` with table-driven subtests (TDD: write test, verify fail, implement, verify pass).

**Organization**: Tasks grouped by user story. US1 (Live Loop Monitoring) and US4 (Panel Navigation & Focus Management) are merged into a single MVP phase because they are co-dependent — panel navigation is meaningless without live content, and live content is unusable without panel navigation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US4)
- Include exact file paths in descriptions

## Path Conventions

- Go project: `internal/` for library packages, `cmd/ralph/` for CLI entry
- Tests: `_test.go` companion files in same package (Go convention)

---

## Phase 1: Setup

**Purpose**: Add the new dependency and extend existing infrastructure

- [ ] T001 Add `charmbracelet/bubbles` dependency via `go get github.com/charmbracelet/bubbles` and verify `go build ./cmd/ralph/` succeeds
- [ ] T002 [P] Add `LogRetention` field (default 20) to `TUIConfig` struct in `internal/config/config.go`; add validation (must be ≥ 0); update `Defaults()` and `InitFile()` template; add table-driven test cases to `internal/config/config_test.go`
- [ ] T003 [P] Implement `EnforceRetention(dir string, maxKeep int) error` in `internal/store/jsonl.go` — list `*.jsonl` files in dir, sort by name, delete oldest exceeding maxKeep; add table-driven tests in `internal/store/jsonl_test.go` (0 files, fewer than limit, exactly at limit, over limit, empty dir)
- [ ] T004 [P] Add malformed-line recovery to `IterationLog()` in `internal/store/jsonl.go` — change `json.Unmarshal` error from hard return to `log.Printf` warning + `continue`; add test case in `internal/store/jsonl_test.go` with a truncated JSON line verifying remaining entries still parse
- [ ] T005 Wire retention into startup: call `store.EnforceRetention()` after `store.NewJSONL()` in `cmd/ralph/execute.go`, passing `cfg.TUI.LogRetention`

**Checkpoint**: Store extended, bubbles available, config ready. Existing tests still pass (`go test ./...`).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core TUI infrastructure that ALL panels depend on. Must complete before any user story phase.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T006 [P] Create `internal/tui/msg.go` — define shared message types: `logEntryMsg`, `loopDoneMsg`, `loopErrMsg`, `tickMsg`, `specSelectedMsg`, `iterationSelectedMsg`, `iterationLogLoadedMsg`, `specsRefreshedMsg`, `loopStateTransitionMsg` per plan.md contracts; also define `waitForEvent(ch)` and `tickCmd()` functions carried forward from current TUI
- [ ] T007 [P] Create `internal/tui/focus.go` + `internal/tui/focus_test.go` — implement `FocusTarget` enum (FocusSpecs/FocusIterations/FocusMain/FocusSecondary) with `Next()`, `Prev()`, `String()` methods; implement `LoopState` enum (StateIdle through StateRegentRestart) with `CanTransitionTo()`, `Label()`, `Symbol()` methods and strict transition table from spec; table-driven tests for all transitions (valid and invalid), cycling, labels
- [ ] T008 [P] Create `internal/tui/layout.go` + `internal/tui/layout_test.go` — implement `Rect` struct, `Layout` struct (Header/Footer/Specs/Iterations/Main/Secondary rects + TooSmall flag), pure `Calculate(width, height int) Layout` function using spec algorithm (sidebar 25% clamped 24–35, specs 40% of body, main 65% of body); table-driven tests: 80×24, 120×40, 200×60, 79×24 (TooSmall), 80×23 (TooSmall), edge cases
- [ ] T009 [P] Create `internal/tui/theme.go` + `internal/tui/theme_test.go` — carry forward color palette from current `styles.go` (colorWhite/Gray/Blue/Green/Yellow/Red/Orange, defaultAccentColor); add `NewTheme(accentColor string) Theme` struct with methods: `AccentHeaderStyle()`, `AccentBorderStyle()`, `DimBorderStyle()`, `PanelBorderStyle(focused bool)`, plus static `toolIcon(name)` and `toolStyle(name)` functions; add `RenderLogLine(entry loop.LogEntry, width int, theme Theme) string` function carrying forward all LogKind rendering from current `view.go:renderLine()`; test toolIcon/toolStyle dispatch and RenderLogLine for each LogKind
- [ ] T010 [P] Create `internal/tui/keymap.go` + `internal/tui/keymap_test.go` — define `GlobalKeyBindings` (tab, shift+tab, 1-4, q, ctrl+c, s, ?) and per-panel key sets (specs: j/k/enter/e/n; iterations: j/k/enter; main: f/[/]/ctrl+u/ctrl+d/j/k; secondary: [/]/j/k); implement `IsGlobalKey(key string) bool` and `PanelKeys(focus FocusTarget) []string` for footer hints; table-driven tests
- [ ] T011 Create `internal/tui/components/tabbar.go` + `internal/tui/components/tabbar_test.go` — implement `TabBar` struct with `NewTabBar(tabs []string)`, `View() string`, `Next() TabBar`, `Prev() TabBar`, `Active() int`, `SetWidth(w int) TabBar`; render active tab with accent color+bold, inactive dimmed, separator characters; table-driven tests: render 2/3/4 tabs, cycling wraps, width truncation
- [ ] T012 Create `internal/tui/components/logview.go` + `internal/tui/components/logview_test.go` — implement `LogView` wrapping `bubbles/viewport` with `NewLogView(w, h int)`, `AppendLine(rendered string)`, `SetContent(lines []string)`, `ToggleFollow()`, `SetSize(w, h int)`, `Update(msg) (LogView, tea.Cmd)`, `View() string`; follow mode auto-scrolls via `viewport.GotoBottom()` on new content; table-driven tests: append+follow, append+manual scroll, toggle follow, resize

**Checkpoint**: Foundation ready — all shared types, layout, theme, keybindings, and reusable components exist. `go test ./internal/tui/...` passes.

---

## Phase 3: US1 + US4 — Live Loop Monitoring + Panel Navigation (Priority: P1) 🎯 MVP

**Goal**: Replace the current single-panel TUI with the multi-panel dashboard. Operator sees live output in main panel, iterations populate in sidebar, Regent messages appear in secondary panel, and all panels are navigable via keyboard. This is the minimum viable TUI redesign.

**Independent Test**: Run `ralph build --max 2` against a project with specs. Verify: header shows branch/iteration/cost/state, live output streams in main panel with auto-scroll, iterations panel populates on completion, Regent messages appear in secondary panel, `tab`/`1-4` cycle focus with visual border highlight, footer shows context-sensitive keybindings.

### Implementation

- [ ] T013 [P] [US1] Create `internal/tui/panels/header.go` + `internal/tui/panels/header_test.go` — implement stateless `RenderHeader(props HeaderProps, width int, accentStyle lipgloss.Style) string` per plan contract; display project name, workDir, mode, branch, iter/max, cost, LoopState symbol+label, elapsed, clock; table-driven tests: all LoopState values, various widths, empty fields fallback
- [ ] T014 [P] [US4] Create `internal/tui/panels/footer.go` + `internal/tui/panels/footer_test.go` — implement stateless `RenderFooter(props FooterProps, width int) string` per plan contract; context-sensitive hints per FocusTarget (specs: j/k/enter/e/n; iterations: j/k/enter; main: f/[/]/ctrl+u/d; secondary: [/]/j/k) plus global hints (q/1-4/s); handle stopRequested state; table-driven tests: each focus target, stop requested, narrow width
- [ ] T015 [P] [US1] Create `internal/tui/panels/specs.go` + `internal/tui/panels/specs_test.go` — implement `SpecsPanel` wrapping `bubbles/list` per plan contract; `NewSpecsPanel(specs []spec.SpecFile, w, h int)` with custom item delegate showing name + status symbol (✅/🔄/⬜); `Update` handles j/k via list, emits `specSelectedMsg` on selection change; `View` renders list with border; `SetSize`; `SelectedSpec`; tests: empty list shows "No specs", selection updates, resize
- [ ] T016 [P] [US1] Create `internal/tui/panels/iterations.go` + `internal/tui/panels/iterations_test.go` — implement `IterationsPanel` wrapping `bubbles/list` per plan contract; items show `#N mode ✓/✗/●` with cost and duration; `AddIteration(s)` and `SetCurrent(n)` for live updates from events; `Update` handles j/k, emits `iterationSelectedMsg`; `View` renders with border; tests: add iterations, selection, current running indicator
- [ ] T017 [P] [US1] Create `internal/tui/panels/main_view.go` + `internal/tui/panels/main_view_test.go` — implement `MainView` per plan contract with tabbar + logview; `AppendLogEntry(entry)` renders entry via `theme.RenderLogLine()` and appends to logview; `Update` handles [/] for tab switching, delegates scroll keys to logview; `View` renders tabbar + active tab content; for MVP, only Output tab is functional (TabSpecContent and TabIterationDetail show placeholder); tests: append entries, tab switching, follow mode toggle, resize
- [ ] T018 [P] [US1] Create `internal/tui/panels/secondary.go` + `internal/tui/panels/secondary_test.go` — implement `SecondaryPanel` per plan contract with tabbar + logview for Regent tab; `AppendEntry(entry)` routes LogRegent/LogGitPull/LogGitPush to appropriate tab views; `Update` handles [/] for tab switching; `View` renders tabbar + active tab; for MVP, only Regent tab is fully functional (Git/Tests/Cost show placeholder); tests: regent events route correctly, tab switching, resize
- [ ] T019 [US1+US4] Create `internal/tui/app.go` + `internal/tui/app_test.go` — implement root `Model` composing all sub-models per plan contract; `New()` constructor takes events channel, store.Reader, accentColor, projectName, workDir, specFiles, requestStop; `Init()` batches waitForEvent + tickCmd; `Update()`: (1) global keys → focus cycling/quit/stop, (2) WindowSizeMsg → recalc layout + resize all panels, (3) logEntryMsg → broadcast to iterations+mainView+secondary + update header state + LoopState transitions, (4) tickMsg → update clock, (5) delegate remaining KeyMsg to focused panel; `View()`: if TooSmall → centered message, else compose header+sidebar+main+secondary+footer via lipgloss.JoinHorizontal/Vertical with focused border highlighting; `Err()` method; tests: focus cycling via tab/1-4, window resize layout, logEntry broadcast, too-small message, quit
- [ ] T020 [US1+US4] Delete old TUI files: remove `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/styles.go`, `internal/tui/model_test.go`
- [ ] T021 [US1+US4] Update `cmd/ralph/wiring.go` — change `tui.New()` call signature in `runWithRegentTUI` and `runWithTUIAndState` to pass `store.Reader` (cast store.Store to store.Reader), call `spec.List(dir)` and pass result; update `finishTUI` if needed for new Model type; ensure all 4 wiring paths compile
- [ ] T022 [US1+US4] Verify build and all tests pass: `go build ./cmd/ralph/` && `go test ./...` && `go vet ./...`

**Checkpoint**: MVP complete. Multi-panel TUI replaces old single-panel TUI. Live loop monitoring works. Panel navigation works. `--no-tui` mode unaffected.

---

## Phase 4: US2 — Spec Navigation (Priority: P2)

**Goal**: Operator can browse specs in the sidebar, view spec content in the main panel, open specs in $EDITOR, and create new specs — all without leaving the TUI.

**Independent Test**: Launch TUI, focus specs panel (press `1`), navigate with j/k, press enter to view spec content in main panel, press `e` to open in $EDITOR, press `n` to create new spec.

### Implementation

- [ ] T023 [US2] Add spec content rendering to main panel: update `MainView.ShowSpec(content string)` in `internal/tui/panels/main_view.go` to load file content into specView viewport and switch to TabSpecContent; in `app.go`, handle `specSelectedMsg` by reading spec file and calling `mainView.ShowSpec()`; add tests for spec tab rendering and content display
- [ ] T024 [US2] Add $EDITOR integration to specs panel in `internal/tui/panels/specs.go` — on `e` key: check `$EDITOR` env var, if unset emit footer message "Set $EDITOR to open specs" via custom msg, if set return `tea.ExecProcess(exec.Command(editor, specPath), callback)` that emits `specsRefreshedMsg` on return; in `app.go` handle `specsRefreshedMsg` by refreshing specs panel list; add test for editor launch command generation and missing-editor fallback
- [ ] T025 [US2] Add spec creation to specs panel in `internal/tui/panels/specs.go` — on `n` key: show `textinput.Model` overlay prompting for spec name; on enter: run spec scaffolding (call `spec.New()` or create file), emit `specsRefreshedMsg`; on escape: cancel; add test for textinput activation, name validation, cancel behavior
- [ ] T026 [US2] Verify spec navigation end-to-end: `go test ./internal/tui/...` passes; verify j/k selection drives main panel content; verify status indicators (✅/🔄/⬜) show correctly

**Checkpoint**: Spec navigation fully functional. Operator can browse, view, edit, and create specs from within the TUI.

---

## Phase 5: US3 — Iteration Drill-Down (Priority: P2)

**Goal**: Operator can select a past iteration and review its full agent output and cost via the main panel's tabs. Output is read from the JSONL session log, so it survives Regent restarts.

**Independent Test**: After a multi-iteration build, navigate to iterations panel, select a completed iteration, verify main panel shows full tool-use log. Switch to Summary tab, verify cost/duration/commit. Select the running iteration, verify it returns to live output.

### Implementation

- [ ] T027 [US3] Add iteration output loading to root model in `internal/tui/app.go` — handle `iterationSelectedMsg`: if iteration is current+running, switch mainView to live Output tab; else spawn async `tea.Cmd` that calls `store.IterationLog(n)` and returns `iterationLogLoadedMsg`; handle `iterationLogLoadedMsg`: call `mainView.ShowIterationLog(entries)` which renders entries via theme and switches to TabIterationDetail; add tests for both paths (live vs past)
- [ ] T028 [US3] Implement Summary tab in `MainView` in `internal/tui/panels/main_view.go` — when TabIterationDetail is active, `]` cycles to a Summary sub-tab showing iteration number, mode, cost, duration, exit subtype, and commit hash formatted as key-value pairs; add tests for summary rendering with all fields
- [ ] T029 [US3] Verify drill-down end-to-end: add test in `internal/tui/app_test.go` simulating iteration selection → log load → display → tab to summary → select live iteration → return to live output

**Checkpoint**: Iteration drill-down fully functional. Past iteration data accessible from JSONL store.

---

## Phase 6: US5 — Loop Control from TUI (Priority: P3)

**Goal**: Operator can start, stop, and restart the loop from within the TUI without restarting the binary.

**Independent Test**: Launch `ralph` in idle/dashboard mode. Press `b` to start build. Press `x` to stop. Press `R` to restart. Verify state transitions in header.

### Implementation

- [ ] T030 [US5] Add loop control keybindings to root model in `internal/tui/app.go` — handle `b` (start build), `x` (stop/cancel context), `p` (start plan), `R` (smart run) as global keys when not captured by focused panel; `b`/`p`/`R` only valid in StateIdle or StateFailed; `x` only valid in StateBuilding or StatePlanning; trigger LoopState transitions via `CanTransitionTo()`; emit appropriate signals (requestStop for `x`, new loop launch for `b`/`p`/`R`)
- [ ] T031 [US5] Update `cmd/ralph/wiring.go` to support TUI-initiated loop starts — expose a mechanism for the TUI to signal "start a new loop run" (e.g., a channel or callback passed to `tui.New()`); this enables dashboard mode where the TUI starts before the loop
- [ ] T032 [US5] Update footer hints for loop control in `internal/tui/panels/footer.go` — when StateIdle: show `b:build p:plan R:run`; when building/planning: show `x:stop`; add tests for each state
- [ ] T033 [US5] Add table-driven tests for all LoopState transitions triggered by key presses in `internal/tui/app_test.go` — test valid transitions (idle→building on `b`, building→failed on `x`) and invalid transitions (building→building on `b` = no-op)

**Checkpoint**: Loop control from TUI functional. Operator can manage loop lifecycle without CLI restarts.

---

## Phase 7: US6 — Secondary Panel Tabs (Priority: P3)

**Goal**: Secondary panel provides tabbed access to Regent log, git log, test output, and cost breakdown.

**Independent Test**: During or after a build, focus secondary panel, switch tabs with `]`, verify each tab shows correct content.

### Implementation

- [ ] T034 [P] [US6] Implement Git tab in `internal/tui/panels/secondary.go` — route `LogGitPull` and `LogGitPush` events to a dedicated logview; render with git-specific formatting (⬆/⬇ icons, accent color); add test for git event routing
- [ ] T035 [P] [US6] Implement Tests tab in `internal/tui/panels/secondary.go` — route test-related Regent events (test pass/fail/rollback messages) to a dedicated viewport; parse test output from `LogRegent` entries that contain "Tests" or "Reverted" keywords; add test for test event routing
- [ ] T036 [P] [US6] Implement Cost tab in `internal/tui/panels/secondary.go` — render a per-iteration cost table from accumulated `IterationSummary` data: columns for iteration number, mode, cost, duration; running total at bottom; `AddIteration(s)` updates the table; add test for table rendering with 0/1/5 iterations
- [ ] T037 [US6] Verify all four secondary tabs render correctly: add test in `internal/tui/panels/secondary_test.go` cycling through all tabs and verifying each has content after events are routed

**Checkpoint**: All secondary panel tabs functional. Regent, git, tests, and cost are each visible in their own tab.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Quality, coverage, and final validation

- [ ] T038 Run `go vet ./...` and `golangci-lint run` — fix any warnings (gocritic ifElseChain, gofmt alignment)
- [ ] T039 Run `go test ./internal/tui/... -coverprofile=tui.cov` — verify ≥80% coverage on layout, focus, event dispatch, panel rendering; identify and fill gaps
- [ ] T040 Run `go test ./internal/store/... -coverprofile=store.cov` — verify ≥80% coverage on retention, malformed-line recovery, existing write/read/index paths
- [ ] T041 [P] Verify `--no-tui` mode unchanged: run `ralph build --no-tui --max 1` and confirm plain-text output matches pre-redesign format (regression check per SC-004)
- [ ] T042 [P] Verify minimum terminal size behavior: test that `Calculate(79, 24)` and `Calculate(80, 23)` return `TooSmall=true` and that `app.View()` renders the resize message (SC-005)
- [ ] T043 [P] Verify binary size: `go build -o ralph ./cmd/ralph/` and check size increase is ≤2MB from adding bubbles (SC-008)
- [ ] T044 Run `go build ./cmd/ralph/` for all cross-compile targets: `GOOS=darwin GOARCH=arm64`, `GOOS=darwin GOARCH=amd64`, `GOOS=linux GOARCH=amd64`, `GOOS=windows GOARCH=amd64`
- [ ] T045 Run quickstart.md validation: follow the keyboard reference and verify all documented keybindings match implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on T001 (bubbles dep). BLOCKS all user stories
- **US1+US4 (Phase 3)**: Depends on Phase 2 completion — this is the MVP
- **US2 (Phase 4)**: Depends on Phase 3 (needs working specs panel and main view)
- **US3 (Phase 5)**: Depends on Phase 3 (needs working iterations panel and main view)
- **US5 (Phase 6)**: Depends on Phase 3 (needs working root model and state machine)
- **US6 (Phase 7)**: Depends on Phase 3 (needs working secondary panel skeleton)
- **Polish (Phase 8)**: Depends on all desired phases being complete

### User Story Dependencies

- **US1+US4 (P1)**: MVP — no dependencies on other stories
- **US2 (P2)**: Independent of US3. Requires specs panel from Phase 3 as starting point
- **US3 (P2)**: Independent of US2. Requires iterations panel + store.Reader from Phase 3
- **US5 (P3)**: Independent. Extends root model's key handling and state machine
- **US6 (P3)**: Independent. Extends secondary panel's tab content

### Within Each Phase

- Tasks marked [P] can run in parallel (different files, no data dependencies)
- Non-[P] tasks have sequential dependencies on prior tasks in the same phase
- Tests are included within each task (TDD: write test → verify fail → implement → verify pass)

### Parallel Opportunities

**Phase 1** (3 parallel groups):
```
T001 (sequential — needed first)
Then: T002 ∥ T003 ∥ T004 (all different files)
Then: T005 (depends on T002, T003)
```

**Phase 2** (2 parallel waves):
```
Wave 1: T006 ∥ T007 ∥ T008 ∥ T009 ∥ T010 (all different files)
Wave 2: T011 ∥ T012 (depend on T006 for msg types, T009 for theme)
```

**Phase 3** (2 parallel waves):
```
Wave 1: T013 ∥ T014 ∥ T015 ∥ T016 ∥ T017 ∥ T018 (all different files)
Wave 2: T019 (app.go — depends on all panels) → T020 → T021 → T022
```

**Phases 4+5 can run in parallel** (different panels, different store features):
```
Phase 4 (T023-T026) ∥ Phase 5 (T027-T029)
```

**Phase 7 tasks are all parallel** (different tabs, same file but independent sections):
```
T034 ∥ T035 ∥ T036
Then: T037
```

---

## Parallel Example: Phase 3 Wave 1

```
Launch together (all create new files, no shared dependencies):
  T013: Create panels/header.go + header_test.go
  T014: Create panels/footer.go + footer_test.go
  T015: Create panels/specs.go + specs_test.go
  T016: Create panels/iterations.go + iterations_test.go
  T017: Create panels/main_view.go + main_view_test.go
  T018: Create panels/secondary.go + secondary_test.go
```

---

## Implementation Strategy

### MVP First (US1 + US4 Only)

1. Complete Phase 1: Setup (T001–T005)
2. Complete Phase 2: Foundational (T006–T012)
3. Complete Phase 3: US1+US4 MVP (T013–T022)
4. **STOP and VALIDATE**: Run `ralph build --max 2` — verify multi-panel TUI works end-to-end
5. Run `go test ./...` — all tests pass
6. This alone is a shippable improvement over the current single-panel TUI

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1+US4 → Test independently → Ship (MVP!)
3. US2 (spec navigation) → Test independently → Ship
4. US3 (iteration drill-down) → Test independently → Ship
5. US5 (loop control) → Test independently → Ship
6. US6 (secondary tabs) → Test independently → Ship
7. Each phase adds value without breaking previous phases

### Suggested Execution Order (single developer)

```
Phase 1 → Phase 2 → Phase 3 (MVP) → Phase 4 ∥ Phase 5 → Phase 6 → Phase 7 → Phase 8
```

---

## Notes

- [P] tasks = different files, no dependencies — safe for parallel agent execution
- Tests are TDD: write failing test first, then implement to make it pass
- Commit after each task or logical group (e.g., one commit per panel)
- Constitution requires `go vet` + `go fmt` + `gocritic` clean before push
- The old TUI files are deleted in T020 — git history preserves them
- `--no-tui` mode is never broken because it uses `formatLogLine()` in `execute.go`, not the TUI package
