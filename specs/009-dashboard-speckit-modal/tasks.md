# Tasks: Dashboard SpecKit Modal

**Input**: Design documents from `/specs/009-dashboard-speckit-modal/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included — constitution principle III (Test-Gated Commits) requires test-first development.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add message types, state extensions, and key bindings shared across all user stories

- [ ] T001 Add SpecKit message types (SpecKitActionMsg, SpecKitOutputMsg, SpecKitInputRequestMsg, SpecKitInputResponseMsg, SpecKitDoneMsg) to internal/tui/msg.go
- [ ] T002 [P] Add StateSpecKit to LoopState enum with Label "SPECKIT", Symbol "●", and valid transitions (Idle↔SpecKit, SpecKit→Failed) in internal/tui/focus.go
- [ ] T003 [P] Add `S` to GlobalKeyBindings slice in internal/tui/keymap.go

---

## Phase 2: Foundational (Layout Improvements)

**Purpose**: Specs panel layout changes (FR-013, FR-014) — standalone, no dependency on modal

**⚠️ CRITICAL**: These are independent of the modal but improve the Specs panel for all subsequent work

- [ ] T004 Change sidebar vertical split from 40% Specs / 60% Iterations to 55% / 45% in Calculate() in internal/tui/layout.go
- [ ] T005 Update layout tests to expect 55/45 split ratios in internal/tui/layout_test.go
- [ ] T006 [P] Add 1-character horizontal inner padding to SpecsPanel.View() so content does not touch the panel border in internal/tui/panels/specs.go
- [ ] T007 [P] Update specs panel tests to verify inner padding in rendered output in internal/tui/panels/specs_test.go

**Checkpoint**: `go test ./internal/tui/...` passes. Specs panel has more room and padding.

---

## Phase 3: User Story 1 + 2 — Open Modal & Select Action (Priority: P1) 🎯 MVP

**Goal**: User presses `S` from any panel → centered modal appears showing spec name and 3 actions → user navigates with j/k → presses enter to select → modal closes and emits SpecKitActionMsg

**Independent Test**: Press `S` with a spec selected → modal appears with spec name, 3 actions. Navigate, press enter → modal closes. Press esc → modal closes without action.

### Tests for US1 + US2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T008 [P] [US1] Test SpecKitModal.Open() sets visible=true, stores specName/specDir, resets cursor to 0 in internal/tui/modal_test.go
- [ ] T009 [P] [US1] Test SpecKitModal.View() renders centered box with spec name in title, 3 action rows with descriptions, cursor indicator on highlighted row in internal/tui/modal_test.go
- [ ] T010 [P] [US2] Test SpecKitModal.Update() key handling: j/k/up/down moves cursor with wrapping at boundaries (0→2→0), esc sets visible=false, enter returns SpecKitActionMsg with correct action/specDir/specName in internal/tui/modal_test.go
- [ ] T011 [P] [US1] Test SpecKitModal.View() renders correctly at minimum terminal size (80×24) — modal fits within available space in internal/tui/modal_test.go

### Implementation for US1 + US2

- [ ] T012 [US1] Create SpecKitModal struct with Open(), Close(), Update(), View() methods following the help overlay pattern (lipgloss.Place for centering, bordered box with accent color) in internal/tui/modal.go
- [ ] T013 [US1] Add `modal SpecKitModal` field to Model struct in internal/tui/app.go
- [ ] T014 [US1] Handle `S` key in handleKey(): when modal not visible and no speckitRunner, get SelectedSpec() from specsPanel, open modal if spec exists; when speckitRunner is active, show "SpecKit action in progress" status in internal/tui/app.go
- [ ] T015 [US1] Add modal key interception in handleKey(): when modal.visible, delegate all keys to modal.Update() before any other key handling (same pattern as helpVisible) in internal/tui/app.go
- [ ] T016 [US1] Add overlayModal() rendering in View(): render base panels first, then overlay modal on top using lipgloss.Place() when modal.visible in internal/tui/app.go
- [ ] T017 [US2] Handle SpecKitActionMsg in Model.Update(): close modal, store action details for runner launch (runner wiring deferred to Phase 4) in internal/tui/app.go
- [ ] T018 [US1] Handle tea.WindowSizeMsg: update modal width/height when terminal is resized in internal/tui/app.go

**Checkpoint**: `go test ./internal/tui/...` passes. Modal opens/closes, navigates, emits action message. No subprocess launched yet.

---

## Phase 4: User Story 3 — Visual Feedback During SpecKit Execution (Priority: P2)

**Goal**: After selecting an action from the modal, a Claude subprocess launches, output streams to the Output tab, and the header shows execution status (SPECKIT:PLAN/CLARIFY/TASKS)

**Independent Test**: Select "Plan" from modal → header shows `● SPECKIT:PLAN (spec-name)` → output appears in Output tab → on completion, header returns to IDLE

### Tests for US3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T019 [P] [US3] Test SpecKitRunner.Start() spawns subprocess with correct arguments (claude -p "/speckit.<action>" --dangerously-skip-permissions) and parses stdout events in internal/tui/speckit_runner_test.go
- [ ] T020 [P] [US3] Test SpecKitRunner emits SpecKitOutputMsg for each parsed event line and SpecKitDoneMsg on process completion in internal/tui/speckit_runner_test.go
- [ ] T021 [P] [US3] Test SpecKitRunner.Stop() cancels context and subprocess exits cleanly in internal/tui/speckit_runner_test.go

### Implementation for US3

- [ ] T022 [US3] Create SpecKitRunner struct with Start(), Stop(), and event bridging goroutine that sends SpecKitOutputMsg/SpecKitDoneMsg via tea.Program.Send() in internal/tui/speckit_runner.go
- [ ] T023 [US3] Add `speckitRunner *SpecKitRunner` and `speckitAction string` fields to Model struct in internal/tui/app.go
- [ ] T024 [US3] Wire SpecKitActionMsg handler: create SpecKitRunner, call Start(), transition LoopState to StateSpecKit, set speckitAction for header display in internal/tui/app.go
- [ ] T025 [US3] Wire SpecKitOutputMsg handler: append line to mainView.outputLog (same as existing logEntryMsg output routing) in internal/tui/app.go
- [ ] T026 [US3] Wire SpecKitDoneMsg handler: set speckitRunner to nil, transition LoopState back to StateIdle (or StateFailed on error), clear speckitAction in internal/tui/app.go
- [ ] T027 [US3] Update header rendering to show `● SPECKIT:PLAN (spec-name)` when LoopState is StateSpecKit — extend HeaderProps or use speckitAction field in internal/tui/panels/header.go

**Checkpoint**: `go test ./internal/tui/...` passes. Full flow works: S → modal → select → subprocess → output streams → header status → completion.

---

## Phase 5: User Story 4 — Interactive Clarify Q&A (Priority: P2)

**Goal**: When Clarify is running and a question is detected in the output, a text input prompt appears at the bottom of the Output tab. User types answer, presses enter, answer is sent to subprocess stdin.

**Independent Test**: Select "Clarify" from modal → question appears in Output tab → input prompt shown → type answer, press enter → answer sent to process → next question appears → session completes → input prompt removed

### Tests for US4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T028 [P] [US4] Test SpecKitRunner detects question patterns ("Your choice:", "Format:") in output and emits SpecKitInputRequestMsg in internal/tui/speckit_runner_test.go
- [ ] T029 [P] [US4] Test SpecKitRunner.WriteAnswer() writes to subprocess stdin pipe in internal/tui/speckit_runner_test.go
- [ ] T030 [P] [US4] Test MainView input prompt mode: SetInputMode(true) shows textinput at bottom of Output tab, SetInputMode(false) removes it, enter key emits SpecKitInputResponseMsg in internal/tui/panels/main_view_test.go

### Implementation for US4

- [ ] T031 [US4] Add stdin pipe creation to SpecKitRunner.Start() when action is "clarify" — use cmd.StdinPipe() in internal/tui/speckit_runner.go
- [ ] T032 [US4] Add question detection logic to SpecKitRunner event bridging goroutine: pattern-match on "Your choice:", "Format:", "**Question" in output lines, emit SpecKitInputRequestMsg when detected in internal/tui/speckit_runner.go
- [ ] T033 [US4] Add WriteAnswer(answer string) method to SpecKitRunner that writes answer + newline to stdin pipe in internal/tui/speckit_runner.go
- [ ] T034 [US4] Add input prompt mode to MainView: add textinput.Model field, SetInputMode(bool) method, render textinput below outputLog viewport when active, handle enter key to emit SpecKitInputResponseMsg in internal/tui/panels/main_view.go
- [ ] T035 [US4] Wire SpecKitInputRequestMsg handler in app.go: call mainView.SetInputMode(true), switch active tab to TabOutput in internal/tui/app.go
- [ ] T036 [US4] Wire SpecKitInputResponseMsg handler in app.go: call speckitRunner.WriteAnswer(), call mainView.SetInputMode(false) in internal/tui/app.go
- [ ] T037 [US4] On SpecKitDoneMsg, ensure mainView.SetInputMode(false) is called to clean up input prompt in internal/tui/app.go

**Checkpoint**: `go test ./internal/tui/...` passes. Full interactive clarify flow works end-to-end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, edge cases, and cleanup

- [ ] T038 Verify modal renders correctly at edge terminal sizes (80×24, 120×40, 200×60) — add table-driven test cases in internal/tui/modal_test.go
- [ ] T039 [P] Add edge case: modal opens while build/plan loop is running (should still work — SpecKit actions are independent) — test in internal/tui/app.go or integration test
- [ ] T040 [P] Verify `go vet ./...` passes with zero warnings
- [ ] T041 Run quickstart.md manual validation steps: launch dashboard, select spec, press S, navigate modal, launch Plan, verify output and header status
- [ ] T042 Verify `go build ./cmd/ralph/` succeeds on all target platforms (cross-compile check)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: No dependencies on Phase 1 — can run in parallel with Setup
- **US1+US2 (Phase 3)**: Depends on Phase 1 (message types, keymap)
- **US3 (Phase 4)**: Depends on Phase 3 (modal emits SpecKitActionMsg that runner handles)
- **US4 (Phase 5)**: Depends on Phase 4 (runner must exist for stdin pipe)
- **Polish (Phase 6)**: Depends on all previous phases

### User Story Dependencies

- **US1+US2 (P1)**: Can start after Phase 1 — no dependency on other stories
- **US3 (P2)**: Depends on US1+US2 (needs modal to emit SpecKitActionMsg)
- **US4 (P2)**: Depends on US3 (needs SpecKitRunner infrastructure)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Struct/type definitions before methods
- Component before wiring into app.go
- Core flow before edge cases

### Parallel Opportunities

- Phase 1 tasks T001, T002, T003 can all run in parallel (different files)
- Phase 2 tasks T004+T005 (layout) and T006+T007 (padding) can run in parallel
- Phase 3 test tasks T008–T011 can all run in parallel (same file but independent test functions)
- Phase 4 test tasks T019–T021 can all run in parallel
- Phase 5 test tasks T028–T030 can all run in parallel (T028–T029 in runner tests, T030 in main_view tests)
- Phase 6 tasks T038–T042 can mostly run in parallel

---

## Parallel Example: Phase 3 (US1 + US2)

```
# Write all tests in parallel:
T008: Test modal Open() state
T009: Test modal View() rendering
T010: Test modal Update() key handling
T011: Test modal minimum size rendering

# Then implement sequentially:
T012: Create SpecKitModal struct (modal.go)
T013: Add modal field to Model (app.go)
T014: Handle S key (app.go)
T015: Modal key interception (app.go)
T016: Modal overlay rendering (app.go)
T017: Handle SpecKitActionMsg (app.go)
T018: Handle resize (app.go)
```

---

## Implementation Strategy

### MVP First (US1 + US2 Only)

1. Complete Phase 1: Setup (T001–T003)
2. Complete Phase 2: Layout (T004–T007) — can run in parallel with Phase 1
3. Complete Phase 3: US1+US2 (T008–T018)
4. **STOP and VALIDATE**: Modal opens, navigates, selects action, emits message
5. This is a functional MVP — modal works, actions are triggered (even if subprocess isn't wired yet)

### Incremental Delivery

1. Phase 1+2 → Foundation ready (layout improved, types defined)
2. Phase 3 → Modal works (MVP — press S, see modal, select action)
3. Phase 4 → Actions execute (output streams, header shows status)
4. Phase 5 → Clarify is interactive (Q&A inline in Output tab)
5. Phase 6 → Polish and edge cases

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US1 and US2 are combined in one phase because they're tightly coupled (modal open ↔ action select)
- All tests use table-driven patterns with t.Run subtests per constitution
- SpecKit runner follows the same subprocess pattern as internal/loop/runner.go
- Commit after each task or logical group
