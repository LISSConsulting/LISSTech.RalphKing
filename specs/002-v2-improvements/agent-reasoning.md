# Feature Specification: Agent Reasoning Display

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Claude's Reasoning in TUI (Priority: P1)

A developer watches the TUI during a build loop and sees Claude's reasoning text
(commentary between tool calls) displayed inline with a thought-bubble icon and
muted style. This reveals *why* the agent is taking each action, aiding
comprehension and debugging.

**Why this priority**: Understanding agent intent is critical for trusting and
debugging autonomous coding loops; without visible reasoning, users see only
tool calls with no context.

**Independent Test**: Run a build loop where Claude emits text content blocks;
verify the TUI displays them with the thought-bubble icon.

**Acceptance Scenarios**:

1. **Given** Claude emits a `text` content block in stream-JSON, **When** the loop processes the event, **Then** a `LogText` entry is emitted on the Events channel with the text content as the message.
2. **Given** a `LogText` entry, **When** the TUI renders it, **Then** the line shows a thought-bubble icon and the message in a muted (gray) style.
3. **Given** a text message longer than 80 characters, **When** the TUI renders it, **Then** the display is truncated to 79 runes followed by an ellipsis character.
4. **Given** an empty text content block, **When** the loop processes the event, **Then** no `LogText` entry is emitted (silently ignored).

---

### User Story 2 - See Claude's Reasoning in Plain-Text Mode (Priority: P2)

A developer running with `--no-tui` sees Claude's reasoning text in the standard
log output with the same `[HH:MM:SS]  message` format as other entries.

**Why this priority**: Plain-text mode must have feature parity for reasoning
visibility.

**Independent Test**: Run with `--no-tui` and verify text entries appear in
standard log format.

**Acceptance Scenarios**:

1. **Given** `--no-tui` mode, **When** a `LogText` entry is emitted, **Then** it is written to the log with `[HH:MM:SS]  message` format via the standard emit path.

---

### Edge Cases

- What happens when Claude emits multiple consecutive text blocks with no tool calls between them?
- What happens when the text content contains newlines or control characters?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST emit a `LogText` log entry when a `claude.EventText` event is received in the loop.
- **FR-002**: System MUST add `LogText` to the `LogKind` enumeration in `internal/loop/event.go`.
- **FR-003**: System MUST render `LogText` entries in the TUI with a thought-bubble icon and muted (gray) style.
- **FR-004**: System MUST truncate TUI-rendered text entries at 80 display characters (79 runes + ellipsis).
- **FR-005**: System MUST silently ignore empty text content blocks.
- **FR-006**: System MUST NOT require special handling in the plain-text `formatLogLine` path; the generic format applies.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A `claude.EventText` event in `loop_test.go` results in a `LogText` entry on the Events channel.
- **SC-002**: `tui.renderLine` for a `LogText` entry contains the thought-bubble icon and message text.
- **SC-003**: Messages longer than 80 characters are truncated to 79 runes + ellipsis in rendered TUI output.
- **SC-004**: All existing tests pass without modification.
- **SC-005**: `go vet ./...` passes with zero warnings.
