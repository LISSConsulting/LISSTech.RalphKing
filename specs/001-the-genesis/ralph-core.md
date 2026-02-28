# Feature Specification: Ralph Core — The Loop CLI

**Feature Branch**: `001-the-genesis`
**Created**: 2026-02-23
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Build Loop with TUI (Priority: P1)

A developer launches `ralph build` to run Claude Code in build mode against their
project. Ralph reads the prompt file, invokes Claude CLI with `--output-format=stream-json --verbose`,
parses tool call events in real time, and displays them in a rich TUI with timestamps,
color-coded tool types, cost tracking, and git status. After each iteration Ralph
pulls with rebase, then pushes.

**Why this priority**: The build loop is Ralph's core value proposition — replacing
`loop.sh` with a visible, reliable, automated coding loop.

**Independent Test**: Run `ralph build --max 1` against a project with `ralph.toml`
and a prompt file; observe TUI output, git operations, and cost summary.

**Acceptance Scenarios**:

1. **Given** a project with `ralph.toml` and a build prompt file, **When** `ralph build --max 3` is run, **Then** Ralph executes up to 3 Claude iterations with TUI output showing timestamps, tool calls, and cost per iteration.
2. **Given** a running build loop, **When** Claude emits tool_use events, **Then** the TUI displays each tool call with an icon (read=blue, write=green, bash=yellow) and the tool's key input.
3. **Given** a completed iteration, **When** the iteration finishes, **Then** Ralph runs `git pull --rebase` then `git push` and displays git status in the footer.

---

### User Story 2 - Run Plan Loop (Priority: P2)

A developer launches `ralph plan` to have Claude create or update an implementation
plan. Ralph feeds the plan prompt file to Claude and stops after a configured number
of iterations.

**Why this priority**: Planning is the first step in spec-driven development; the
plan loop structures Claude's work before the build phase.

**Independent Test**: Run `ralph plan --max 1` and verify Claude receives the plan
prompt and the TUI shows iteration output.

**Acceptance Scenarios**:

1. **Given** a project with a plan prompt file configured, **When** `ralph plan --max 2` is run, **Then** Ralph runs at most 2 Claude iterations using the plan prompt.
2. **Given** no `--max` flag, **When** `ralph plan` is run, **Then** Ralph uses the default `plan.max_iterations` from `ralph.toml`.

---

### User Story 3 - Smart Run Mode (Priority: P3)

A developer runs `ralph run` for an automatic plan-then-build workflow. If
CHRONICLE.md does not exist or is empty, Ralph runs the plan phase first, then
switches to the build phase.

**Why this priority**: Reduces cognitive overhead — one command handles the full
spec-driven workflow.

**Independent Test**: Run `ralph run` in a project without CHRONICLE.md; verify
plan phase runs first, then build phase starts.

**Acceptance Scenarios**:

1. **Given** a project without CHRONICLE.md, **When** `ralph run` is executed, **Then** Ralph runs the plan phase first (up to `plan.max_iterations`), then the build phase.
2. **Given** a project with a non-empty CHRONICLE.md, **When** `ralph run` is executed, **Then** Ralph skips the plan phase and runs build directly.

---

### User Story 4 - Project Initialization (Priority: P4)

A developer runs `ralph init` to scaffold a new project with `ralph.toml`, prompt
files, and a chronicle file with sensible defaults.

**Why this priority**: Lowers the barrier to getting started; provides a working
default configuration.

**Independent Test**: Run `ralph init` in an empty directory; verify all scaffold
files are created.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** `ralph init` is run, **Then** `ralph.toml`, `PLAN.md`, `BUILD.md`, `CHRONICLE.md`, and `.gitignore` are created.
2. **Given** a directory with existing `ralph.toml`, **When** `ralph init` is run, **Then** existing files are not overwritten (idempotent).

---

### User Story 5 - Spec Management (Priority: P5)

A developer uses `ralph spec new <name>` to create a new spec file and
`ralph spec list` to see all specs with their implementation status.

**Why this priority**: Enforces spec-driven workflow by making spec creation and
tracking first-class CLI operations.

**Independent Test**: Run `ralph spec new my-feature` and verify file creation;
run `ralph spec list` and verify status indicators.

**Acceptance Scenarios**:

1. **Given** a project, **When** `ralph spec new my-feature` is run, **Then** `specs/my-feature.md` is created from the spec template.
2. **Given** specs in the `specs/` directory, **When** `ralph spec list` is run, **Then** each spec is listed with a status indicator (done/in-progress/not-started) based on CHRONICLE.md cross-reference.

---

### Edge Cases

- What happens when `ralph.toml` is missing or malformed?
- What happens when the configured prompt file does not exist?
- What happens when `git pull --rebase` encounters a conflict?
- What happens when Claude CLI is not installed or not on PATH?
- What happens when `--max 0` is passed (unlimited iterations)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse configuration from `ralph.toml` using TOML format with sections: `[project]`, `[claude]`, `[plan]`, `[build]`, `[git]`, `[regent]`.
- **FR-002**: System MUST invoke Claude CLI with `--output-format=stream-json --verbose` and parse the streaming JSON events.
- **FR-003**: System MUST display a rich TUI using bubbletea + lipgloss with a header bar (project name, branch, iteration count, cost), scrollable iteration panel, and footer bar (git status, quit hint).
- **FR-004**: System MUST track cost per iteration and running total from Claude's result events.
- **FR-005**: System MUST run `git pull --rebase` before `git push` after each iteration; fall back to merge on rebase conflict.
- **FR-006**: System MUST support `plan`, `build`, `run`, `status`, `init`, `spec new`, and `spec list` subcommands via cobra.
- **FR-007**: System MUST support a `--no-tui` mode that writes plain-text log output.
- **FR-008**: System MUST support a `--max N` flag to limit iteration count on `plan`, `build`, and `run` commands.
- **FR-009**: System MUST color-code TUI tool call entries by type: reads=blue, writes=green, bash=yellow, errors=red.
- **FR-010**: System MUST timestamp every TUI log line with `[HH:MM:SS]` format.

### Key Entities

- **Config**: Parsed from `ralph.toml`; holds project, claude, plan, build, git, and regent settings.
- **Loop**: Core iteration engine; drives prompt → Claude → parse → git cycle.
- **LogEntry**: Structured event emitted by the loop for TUI consumption (tool use, result, error, git ops).
- **Agent**: Interface for Claude CLI adapter; `Run()` returns a channel of typed events.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `ralph build --max 1` completes one Claude iteration with visible TUI output and git push.
- **SC-002**: `ralph plan --max 1` completes one plan iteration using the configured plan prompt.
- **SC-003**: `ralph init` creates all scaffold files in an empty directory.
- **SC-004**: `ralph spec list` shows correct status indicators for all specs.
- **SC-005**: `go build ./...` and `go test ./...` pass with zero errors.
- **SC-006**: `go vet ./...` passes with zero warnings.
- **SC-007**: Binary cross-compiles for darwin/arm64, darwin/amd64, linux/amd64, windows/amd64.
