# Feature Specification: Rename Prompt Files and Chronicle

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - New Projects Use Clean File Names (Priority: P1)

A developer runs `ralph init` and receives scaffold files with clean, short names:
`PLAN.md`, `BUILD.md`, and `CHRONICLE.md` instead of the verbose `PROMPT_plan.md`,
`PROMPT_build.md`, and `IMPLEMENTATION_PLAN.md`. The names are easier to type,
reference in prompts, and don't expose internal implementation details.

**Why this priority**: File names are the most visible touchpoint for users; clean
names reduce friction and improve the developer experience.

**Independent Test**: Run `ralph init` in an empty directory; verify `PLAN.md`,
`BUILD.md`, and `CHRONICLE.md` are created (not the old names).

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** `ralph init` runs, **Then** `PLAN.md`, `BUILD.md`, and `CHRONICLE.md` are created.
2. **Given** the created `ralph.toml`, **When** its contents are read, **Then** `plan.prompt_file` references `PLAN.md` and `build.prompt_file` references `BUILD.md`.
3. **Given** the created `PLAN.md` and `BUILD.md`, **When** their contents are read, **Then** they reference `CHRONICLE.md` (not `IMPLEMENTATION_PLAN.md`).

---

### User Story 2 - Config Defaults Use New Names (Priority: P2)

A developer does not override `prompt_file` in `ralph.toml`. The system defaults
to `PLAN.md` and `BUILD.md` for the plan and build modes respectively.

**Why this priority**: Defaults must match the scaffold file names so that
zero-config projects work correctly.

**Independent Test**: Load config without `prompt_file` overrides; verify defaults
are `PLAN.md` and `BUILD.md`.

**Acceptance Scenarios**:

1. **Given** `ralph.toml` has no explicit `plan.prompt_file`, **When** config is loaded, **Then** `cfg.Plan.PromptFile` is `"PLAN.md"`.
2. **Given** `ralph.toml` has no explicit `build.prompt_file`, **When** config is loaded, **Then** `cfg.Build.PromptFile` is `"BUILD.md"`.

---

### User Story 3 - Smart Run Detects CHRONICLE.md (Priority: P3)

A developer runs `ralph run` and the system checks for `CHRONICLE.md` (not
`IMPLEMENTATION_PLAN.md`) to decide whether the plan phase needs to run.

**Why this priority**: Smart-run detection must align with the renamed file to
avoid incorrectly triggering the plan phase in projects using new names.

**Independent Test**: Create a project with `CHRONICLE.md`; run `ralph run` and
verify the plan phase is skipped.

**Acceptance Scenarios**:

1. **Given** a project without `CHRONICLE.md`, **When** `ralph run` is executed, **Then** the plan phase runs first.
2. **Given** a project with a non-empty `CHRONICLE.md`, **When** `ralph run` is executed, **Then** the plan phase is skipped and build runs directly.

---

### Edge Cases

- What happens to existing projects with `PROMPT_plan.md` that don't set explicit `prompt_file`?
- What happens when both `IMPLEMENTATION_PLAN.md` and `CHRONICLE.md` exist?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `ScaffoldProject` MUST create `PLAN.md`, `BUILD.md`, and `CHRONICLE.md` (not the old names).
- **FR-002**: The `PLAN.md` and `BUILD.md` template content MUST reference `CHRONICLE.md`.
- **FR-003**: `config.Defaults()` MUST return `Plan.PromptFile = "PLAN.md"` and `Build.PromptFile = "BUILD.md"`.
- **FR-004**: The `ralph.toml` template written by `InitFile` MUST reference `PLAN.md` and `BUILD.md`.
- **FR-005**: `executeSmartRun` MUST check for `CHRONICLE.md` (not `IMPLEMENTATION_PLAN.md`) for plan-phase detection.
- **FR-006**: `spec.List()` MUST read `CHRONICLE.md` for status cross-reference.
- **FR-007**: Scaffold MUST remain idempotent: skip files that already exist.

### Breaking Change Notice

This is a breaking change for existing projects that rely on default prompt file
names without explicit `prompt_file` in ralph.toml. Migration: rename files or add
explicit `prompt_file` entries pointing to the old names.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `ralph init` in a fresh directory creates `PLAN.md`, `BUILD.md`, `CHRONICLE.md`.
- **SC-002**: The created `ralph.toml` references `PLAN.md` and `BUILD.md`.
- **SC-003**: `PLAN.md` and `BUILD.md` content references `CHRONICLE.md`.
- **SC-004**: `config.Defaults()` returns `PLAN.md` and `BUILD.md`.
- **SC-005**: `ralph run` without `CHRONICLE.md` triggers the plan phase.
- **SC-006**: `ralph spec list` cross-references `CHRONICLE.md` for status.
- **SC-007**: All existing tests pass with updated file name expectations.
- **SC-008**: New tests cover `CHRONICLE.md` idempotency.
