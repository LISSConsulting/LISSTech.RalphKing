# Feature Specification: Dashboard SpecKit Modal

**Feature Branch**: `009-dashboard-speckit-modal`
**Created**: 2026-03-11
**Status**: Draft
**Input**: User description: "In dashboard mode, add a modal dialog for starting speckit.plan, speckit.clarify, and speckit.tasks"

## Clarifications

### Session 2026-03-11

- Q: How should the interactive clarify workflow (which asks user questions) work inside the TUI? → A: Show clarify Q&A inline in the Output tab with a text input prompt for each question.
- Q: Should the modal trigger key work only from the Specs panel or globally from any panel? → A: Global — `S` works from any panel. The modal MUST prominently display the target spec name so the user can confirm before selecting an action. Actions are not destructive (they create/update spec files), so confirmation via spec name display is sufficient.
- Q: How should the Specs panel vertical space be split with Iterations? → A: 55% Specs / 45% Iterations (up from 40/60).
- Q: What should happen when `S` is pressed while a SpecKit action is already running? → A: Block — modal does not open; show a brief status message ("SpecKit action in progress").

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open SpecKit Actions Modal (Priority: P1)

A developer is viewing the dashboard with a spec selected in the Specs panel. They want to run a SpecKit workflow (plan, clarify, or tasks) against that spec without leaving the TUI. They press a key to open a modal dialog that presents the available SpecKit actions.

**Why this priority**: This is the core interaction — without the modal, no SpecKit actions can be triggered from the dashboard.

**Independent Test**: Can be fully tested by pressing the trigger key with a spec selected and verifying the modal appears with three action options.

**Acceptance Scenarios**:

1. **Given** any panel is focused and a spec is selected in the Specs panel, **When** the user presses the SpecKit action key, **Then** a modal dialog appears centered on screen displaying the target spec name and listing three actions: Plan, Clarify, and Tasks.
2. **Given** no spec is selected (empty specs list), **When** the user presses the SpecKit action key, **Then** nothing happens (no modal appears).
3. **Given** the modal is open, **When** the user presses `esc`, **Then** the modal closes and focus returns to the previously focused panel.

---

### User Story 2 - Select and Launch a SpecKit Action (Priority: P1)

A developer has the SpecKit modal open and wants to pick an action. They navigate the list using `j`/`k` or arrow keys, then press `enter` to launch the selected action against the currently selected spec.

**Why this priority**: Equally critical — the modal is useless if actions can't be launched from it.

**Independent Test**: Can be tested by opening the modal, selecting an action, pressing `enter`, and verifying the corresponding SpecKit workflow starts for the selected spec.

**Acceptance Scenarios**:

1. **Given** the modal is open, **When** the user highlights "Plan" and presses `enter`, **Then** the modal closes and the plan workflow starts for the selected spec.
2. **Given** the modal is open, **When** the user highlights "Clarify" and presses `enter`, **Then** the modal closes and the clarify workflow starts for the selected spec.
3. **Given** the modal is open, **When** the user highlights "Tasks" and presses `enter`, **Then** the modal closes and the tasks workflow starts for the selected spec.
4. **Given** the modal is open, **When** the user navigates with `j`/`k` or arrow keys, **Then** the highlighted action changes accordingly, wrapping at boundaries.

---

### User Story 3 - Visual Feedback During SpecKit Execution (Priority: P2)

After launching a SpecKit action from the modal, the developer sees feedback in the dashboard indicating the action is running and can view its output.

**Why this priority**: Important for usability but the core trigger-and-launch flow works without it.

**Independent Test**: Can be tested by launching a SpecKit action and verifying the dashboard shows execution status and output in the appropriate panel.

**Acceptance Scenarios**:

1. **Given** a SpecKit action has been launched, **When** execution begins, **Then** the header or status area indicates which SpecKit action is running and for which spec.
2. **Given** a SpecKit action is running, **When** the action produces output, **Then** the output is streamed to the Output tab in the Main panel.
3. **Given** a SpecKit action completes, **When** it succeeds, **Then** the status updates to reflect completion.

---

### User Story 4 - Interactive Clarify Q&A in Output Tab (Priority: P2)

A developer launches the Clarify action from the modal. The clarify workflow asks questions one at a time. Each question and its options appear in the Output tab, and a text input prompt appears at the bottom of the Output tab for the user to type their answer. The Q&A continues inline until the clarify session completes.

**Why this priority**: Clarify is interactive by nature; without inline input, it would need to run non-interactively or in an external terminal, reducing its value.

**Independent Test**: Can be tested by launching Clarify, verifying a question appears in the Output tab with an input prompt, typing an answer, and verifying the next question appears.

**Acceptance Scenarios**:

1. **Given** the clarify workflow is running, **When** it presents a question, **Then** the question and options appear in the Output tab and a text input prompt is shown for the user to type their answer.
2. **Given** a clarify question is displayed with an input prompt, **When** the user types an answer and presses `enter`, **Then** the answer is sent to the clarify process and the next question appears.
3. **Given** the clarify workflow completes all questions, **When** the session ends, **Then** the input prompt is removed and the Output tab returns to read-only streaming mode.

---

### Edge Cases

- What happens when the user opens the modal while a loop is already running? The modal should still open — SpecKit actions are independent of build/plan loops.
- What happens when a SpecKit action fails mid-execution? The error output appears in the Output tab, and the status reflects failure.
- What happens if the selected spec directory is missing `spec.md`? The Clarify and Plan actions should still attempt to run (they may create the file or report their own error).
- What happens when the modal is open and the terminal is resized? The modal repositions to remain centered.
- What happens when `S` is pressed while a SpecKit action is already running? The modal does not open; a brief status message ("SpecKit action in progress") is shown instead.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a global keyboard shortcut (`S`) to open the SpecKit actions modal when a spec is selected, regardless of which panel is focused.
- **FR-002**: The modal MUST display three selectable actions: Plan, Clarify, and Tasks, each with a brief description.
- **FR-003**: The modal MUST support keyboard navigation (`j`/`k`, up/down arrows) to highlight actions, and `enter` to confirm selection.
- **FR-004**: The modal MUST close on `esc` without triggering any action, restoring prior focus.
- **FR-005**: The modal MUST close on action selection and emit a message to start the corresponding SpecKit workflow for the selected spec.
- **FR-006**: The modal MUST capture all keyboard input while open (no key events leak to underlying panels).
- **FR-007**: The modal MUST render as a centered overlay with a visible border, distinct from the background panels. The modal MUST prominently display the name of the target spec so the user can confirm they are acting on the correct spec.
- **FR-008**: The system MUST display execution status (running/completed/failed) for SpecKit actions in the dashboard header or status area.
- **FR-009**: SpecKit action output MUST stream to the Output tab in the Main panel, consistent with how build/plan loop output is displayed.
- **FR-010**: The modal trigger key MUST be a no-op when no spec is selected.
- **FR-010a**: The modal trigger key MUST NOT open the modal while a SpecKit action is already running; instead, a brief status message MUST be displayed.
- **FR-011**: When the Clarify workflow presents a question, the Output tab MUST show a text input prompt allowing the user to type and submit an answer inline.
- **FR-012**: The text input prompt MUST be removed when the Clarify workflow completes or is cancelled.
- **FR-013**: The Specs panel MUST have inner padding so that content does not touch the panel border.
- **FR-014**: The left sidebar vertical split MUST be 55% Specs / 45% Iterations (changed from 40/60) to give the Specs panel more room.

### Key Entities

- **SpecKit Action**: One of Plan, Clarify, or Tasks — a workflow that processes spec files (spec.md, plan.md, tasks.md) for the selected spec directory.
- **Modal State**: Tracks whether the modal is open, which action is highlighted, and which spec it was invoked for.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can open the SpecKit modal, select an action, and launch it in under 3 seconds (two keypresses: open + confirm).
- **SC-002**: All three SpecKit actions (Plan, Clarify, Tasks) are launchable from the modal without leaving the dashboard.
- **SC-003**: The modal renders correctly at all supported terminal sizes (minimum 80x24).
- **SC-004**: SpecKit action output is visible in the dashboard within 1 second of execution starting.

## Assumptions

- The SpecKit workflows (plan, clarify, tasks) are invoked as Claude Code subprocess commands, consistent with how build/plan loops currently shell out to Claude.
- The modal uses the same styling/theming system as the existing TUI (lipgloss theme).
- Only one SpecKit action can run at a time; launching a second while one is running is out of scope for this feature (existing loop-busy guards apply).
- The trigger key will be `S` (capital S for "SpecKit"), chosen to avoid conflicts with existing bindings (`s` = stop, `n` = new spec, `e` = edit, `W` = worktree).
