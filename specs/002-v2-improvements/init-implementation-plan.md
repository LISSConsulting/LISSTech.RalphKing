# Feature Specification: ralph init Creates CHRONICLE.md

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fresh Project Gets CHRONICLE.md (Priority: P1)

A developer runs `ralph init` in a new project directory and receives a starter
`CHRONICLE.md` file with structured sections for completed work, remaining work,
and key learnings. This gives the plan phase a clear starting point to build on.

**Why this priority**: Without an initial chronicle file, new projects have no
structure for Claude's plan phase to populate; the starter template ensures
a consistent starting point.

**Independent Test**: Run `ralph init` in an empty directory; verify `CHRONICLE.md`
exists with all required sections.

**Acceptance Scenarios**:

1. **Given** a directory with no existing files, **When** `ralph init` runs, **Then** `CHRONICLE.md` is created with a one-line summary placeholder, a Completed Work table, a Remaining Work table, and a Key Learnings section.
2. **Given** a directory that already contains `CHRONICLE.md`, **When** `ralph init` runs, **Then** the existing file is not modified and is not listed in the created files output.
3. **Given** a fresh directory, **When** `ralph init` runs, **Then** `CHRONICLE.md` appears in the created files list after `.gitignore`.

---

### Edge Cases

- What happens when the directory is read-only?
- What happens when `CHRONICLE.md` exists but is empty?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `ScaffoldProject` MUST create `CHRONICLE.md` in the project root when the file does not exist.
- **FR-002**: The created `CHRONICLE.md` MUST contain: a one-line project summary placeholder, a `## Completed Work` section with a table (Phase, Features, Tags columns), a `## Remaining Work` section with a table (Priority, Item, Location, Notes columns), and a `## Key Learnings` section with a bullet placeholder.
- **FR-003**: `ScaffoldProject` MUST NOT modify or overwrite an existing `CHRONICLE.md` (idempotent).
- **FR-004**: `ScaffoldProject` MUST NOT include `CHRONICLE.md` in the created files list if it already exists.
- **FR-005**: When created, `CHRONICLE.md` MUST appear in the created list after `.gitignore`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `ScaffoldProject` creates `CHRONICLE.md` with all required sections in a fresh directory.
- **SC-002**: Calling `ScaffoldProject` twice results in `CHRONICLE.md` listed in `created` only on the first call.
- **SC-003**: `go test ./internal/config/...` passes with coverage on the new code paths.
