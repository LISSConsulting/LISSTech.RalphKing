# Feature Specification: The Regent — Ralph's Supervisor

**Feature Branch**: `001-the-genesis`
**Created**: 2026-02-23
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Crash Detection and Recovery (Priority: P1)

Ralph crashes mid-iteration (exits non-zero). The Regent detects the crash, logs
the failure, waits a configurable backoff period, and restarts Ralph automatically.
After `max_retries` consecutive failures, the Regent escalates and stops.

**Why this priority**: Autonomous coding loops must survive transient failures;
without crash recovery Ralph is no better than a bash script.

**Independent Test**: Mock Ralph to exit non-zero; verify Regent restarts it after
backoff and stops after max retries.

**Acceptance Scenarios**:

1. **Given** Ralph exits with a non-zero code, **When** the Regent detects the crash, **Then** it logs the failure, waits `retry_backoff_seconds`, and restarts Ralph.
2. **Given** Ralph has crashed `max_retries` consecutive times, **When** the next crash occurs, **Then** the Regent escalates (prints to terminal) and stops supervision.
3. **Given** Ralph completes an iteration successfully after a crash, **When** the success is detected, **Then** the consecutive error counter resets to zero.

---

### User Story 2 - Hang Detection (Priority: P2)

Ralph stops producing output but has not exited (hung process). The Regent detects
the silence, kills Ralph, and restarts it.

**Why this priority**: A hung process blocks the loop indefinitely with no visible
feedback; automatic detection prevents wasted time and compute.

**Independent Test**: Mock Ralph to stop producing output; verify Regent kills it
after `hang_timeout_seconds` and restarts.

**Acceptance Scenarios**:

1. **Given** Ralph has produced no output for `hang_timeout_seconds`, **When** the timeout fires, **Then** the Regent kills the process, logs the timeout, and restarts Ralph.
2. **Given** Ralph is producing output, **When** each new output line arrives, **Then** the hang timer resets.

---

### User Story 3 - Test Regression Detection (Priority: P3)

After a successful iteration and git push, the Regent runs the configured test
command. If tests fail, the Regent reverts the commit and pushes the revert,
keeping the codebase healthy.

**Why this priority**: Autonomous commits can introduce regressions; test gating
is the safety net that keeps the codebase deployable.

**Independent Test**: Configure `test_command` and `rollback_on_test_failure = true`;
mock a failing test after commit; verify Regent reverts HEAD and pushes.

**Acceptance Scenarios**:

1. **Given** `rollback_on_test_failure = true` and `test_command` is set, **When** tests pass after an iteration, **Then** the commit is kept and the Regent logs success.
2. **Given** `rollback_on_test_failure = true` and `test_command` is set, **When** tests fail after an iteration, **Then** the Regent runs `git revert HEAD --no-edit`, pushes the revert, and logs the reason.
3. **Given** `rollback_on_test_failure = false`, **When** an iteration completes, **Then** no test run is performed.

---

### User Story 4 - State Tracking and Status (Priority: P4)

The Regent writes its state to `.ralph/regent-state.json` after each significant
event. `ralph status` reads this file to display the last run summary.

**Why this priority**: Observable state enables debugging, monitoring, and the
`ralph status` command.

**Independent Test**: Run Ralph with Regent; verify `.ralph/regent-state.json`
contains expected fields; run `ralph status` and verify output.

**Acceptance Scenarios**:

1. **Given** the Regent is running, **When** an iteration completes, **Then** `.ralph/regent-state.json` is updated with ralph_pid, iteration count, consecutive_errors, last_output_at, last_commit, and total_cost_usd.
2. **Given** a valid state file exists, **When** `ralph status` is run, **Then** it displays the state summary.

---

### User Story 5 - Graceful Shutdown (Priority: P5)

The user sends a signal to stop the Regent. SIGINT/SIGTERM finishes the current
iteration before stopping; SIGQUIT stops immediately and kills Ralph.

**Why this priority**: Users need predictable shutdown behavior to avoid dirty
git state or lost work.

**Independent Test**: Send SIGINT during an iteration; verify it completes before
exiting. Send SIGQUIT; verify immediate termination.

**Acceptance Scenarios**:

1. **Given** Ralph is mid-iteration, **When** SIGINT is received, **Then** the current iteration completes (including git push), then the Regent stops.
2. **Given** Ralph is mid-iteration, **When** SIGQUIT is received, **Then** the Regent kills Ralph immediately and exits.

---

### Edge Cases

- What happens when `.ralph/` directory does not exist (first run)?
- What happens when the test command itself hangs?
- What happens when `git revert` fails (e.g., merge commit)?
- What happens when the state file is corrupted or has an incompatible schema?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST monitor Ralph as a child process and detect non-zero exits.
- **FR-002**: System MUST restart Ralph after `retry_backoff_seconds` on crash, up to `max_retries` consecutive failures.
- **FR-003**: System MUST track last output timestamp and kill Ralph if no output for `hang_timeout_seconds`.
- **FR-004**: System MUST run `test_command` after each successful iteration when `rollback_on_test_failure = true`.
- **FR-005**: System MUST revert HEAD and push when test command fails.
- **FR-006**: System MUST write state to `.ralph/regent-state.json` with fields: ralph_pid, iteration, consecutive_errors, last_output_at, last_commit, total_cost_usd.
- **FR-007**: System MUST handle SIGINT/SIGTERM by finishing the current iteration then stopping.
- **FR-008**: System MUST handle SIGQUIT by killing Ralph immediately and exiting.
- **FR-009**: System MUST run in embedded mode (same binary as Ralph, separate goroutine) by default.
- **FR-010**: System MUST surface Regent activity in the Ralph TUI with a shield icon prefix and orange color.

### Key Entities

- **Regent**: Supervisor struct; owns Start(), Stop(), and the supervision loop.
- **State**: Serializable struct persisted to `.ralph/regent-state.json`.
- **Tester**: Runs the configured test command and handles revert logic.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Ralph crash triggers Regent restart within `retry_backoff_seconds`.
- **SC-002**: Ralph hang triggers Regent kill-and-restart after `hang_timeout_seconds` of silence.
- **SC-003**: Test failure with `rollback_on_test_failure = true` results in automatic `git revert HEAD` and push.
- **SC-004**: After `max_retries` consecutive failures, Regent stops and reports.
- **SC-005**: `ralph status` displays current Regent state from the state file.
- **SC-006**: SIGINT results in graceful shutdown after current iteration.
- **SC-007**: `.ralph/regent-state.json` is written and readable after every state change.
- **SC-008**: Regent messages are visible in the TUI log with shield prefix.
