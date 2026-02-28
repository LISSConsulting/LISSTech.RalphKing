# Feature Specification: Display Current Working Directory in TUI Header

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Working Directory in Header (Priority: P1)

A developer running multiple Ralph instances across different projects sees the
working directory in the TUI header, providing unambiguous context about which
repository Ralph is operating on.

**Why this priority**: When running multiple instances or launching from scripts,
the terminal title may not identify the active repository; the header directory
eliminates ambiguity.

**Independent Test**: Launch Ralph in a project directory; verify the TUI header
displays `dir: ~/Projects/my-project` (tilde-abbreviated, forward slashes).

**Acceptance Scenarios**:

1. **Given** Ralph is launched in `/home/user/Projects/my-project`, **When** the TUI renders the header, **Then** it shows `dir: ~/Projects/my-project` with the home directory abbreviated to `~`.
2. **Given** Ralph is launched in `C:\Users\dev\Projects\my-project` on Windows, **When** the TUI renders the header, **Then** backslashes are converted to forward slashes in the display.
3. **Given** no working directory is provided (empty string), **When** the TUI renders the header, **Then** the `dir:` field is omitted entirely.

---

### Edge Cases

- What happens when the path does not start with the home directory?
- What happens when `os.UserHomeDir()` returns an error?
- What happens when the path is the home directory itself (`~`)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display the working directory in the TUI header as `dir: <abbreviated-path>` immediately after the project name.
- **FR-002**: System MUST replace the user's home directory prefix with `~` in the displayed path.
- **FR-003**: System MUST convert backslashes to forward slashes in the displayed path.
- **FR-004**: System MUST omit the `dir:` field entirely when the working directory is empty.
- **FR-005**: System MUST accept the working directory as a parameter to the TUI model constructor.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: TUI header shows `dir: ~/...` when a working directory is set.
- **SC-002**: TUI header omits `dir:` when working directory is empty.
- **SC-003**: Windows-style backslash paths display as forward slashes.
- **SC-004**: All existing tests pass without modification.
- **SC-005**: New tests cover the `abbreviatePath` helper and header rendering with/without `workDir`.
