# Feature Specification: Graceful Stop — Stop After Current Iteration

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Request Graceful Stop via TUI (Priority: P1)

A developer watching the TUI decides to stop the loop after the current iteration
completes. They press `s` and the footer immediately shows a stopping indicator.
The loop finishes its current iteration (including git push), then exits cleanly.

**Why this priority**: `ctrl+c` kills mid-iteration, potentially leaving dirty git
state; a graceful stop ensures the iteration completes and the codebase is clean.

**Independent Test**: Start a build loop, press `s` during an iteration; verify the
footer updates and the loop exits after the iteration completes.

**Acceptance Scenarios**:

1. **Given** the TUI is active and a loop is running, **When** the user presses `s`, **Then** the footer changes to show a stop-pending indicator and the stop-requested flag is set.
2. **Given** a stop has been requested, **When** the current iteration completes (including git push), **Then** the loop emits a `LogStopped` event and exits with nil error.
3. **Given** a stop has already been requested, **When** the user presses `s` again, **Then** nothing happens (no-op).
4. **Given** `--no-tui` mode, **When** the user wants to stop, **Then** they use `ctrl+c` or SIGQUIT as before; no `s` key binding is available.

---

### Edge Cases

- What happens if `s` is pressed between iterations (during git operations)?
- What happens if the loop has zero iterations remaining when `s` is pressed?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept an `s` key press in the TUI to request a graceful stop.
- **FR-002**: System MUST update the TUI footer to show a stop-pending message when stop is requested.
- **FR-003**: System MUST complete the current iteration (including git push) before stopping.
- **FR-004**: System MUST emit a `LogStopped` event and return nil error when stopping gracefully.
- **FR-005**: System MUST treat subsequent `s` presses as no-ops after the first.
- **FR-006**: System MUST accept a `StopAfter <-chan struct{}` field on the Loop struct, checked after each iteration.
- **FR-007**: System MUST NOT affect `--no-tui` mode behavior.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Pressing `s` in the TUI footer shows the stop-pending message.
- **SC-002**: The loop completes its current iteration and exits with `LogStopped` after stop is requested.
- **SC-003**: Second `s` press is a no-op (no duplicate stop signals).
- **SC-004**: `--no-tui` mode is unaffected by the graceful stop feature.
- **SC-005**: All existing tests pass without modification.
- **SC-006**: New tests cover the `s` key handler and the `StopAfter` channel check in the loop.
