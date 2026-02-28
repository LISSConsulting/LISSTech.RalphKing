# Feature Specification: Project Name Auto-Detection

**Feature Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Implemented

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Auto-Detect Name from Manifest (Priority: P1)

A developer uses Ralph on an existing project without setting `project.name` in
`ralph.toml`. Ralph automatically detects the project name from common manifest
files (pyproject.toml, package.json, Cargo.toml) and displays it in the TUI
header. Setup is effortless.

**Why this priority**: Most projects already have manifest files with the project
name; requiring manual configuration in `ralph.toml` is unnecessary friction.

**Independent Test**: Create a `ralph.toml` with empty `project.name` alongside
a `package.json` with `"name": "my-app"`; run `ralph status` and verify the
project name is detected.

**Acceptance Scenarios**:

1. **Given** `ralph.toml` has an empty `project.name` and `pyproject.toml` has `[project] name = "myapp"`, **When** config is loaded, **Then** `cfg.Project.Name` is set to `"myapp"`.
2. **Given** `pyproject.toml` has no `[project] name` but has `[tool.poetry] name = "myapp"`, **When** config is loaded, **Then** `cfg.Project.Name` is set to `"myapp"`.
3. **Given** no `pyproject.toml` exists but `package.json` has `"name": "myapp"`, **When** config is loaded, **Then** `cfg.Project.Name` is set to `"myapp"`.
4. **Given** no `pyproject.toml` or `package.json` exists but `Cargo.toml` has `[package] name = "myapp"`, **When** config is loaded, **Then** `cfg.Project.Name` is set to `"myapp"`.
5. **Given** `ralph.toml` has `project.name = "ExplicitName"`, **When** config is loaded, **Then** `cfg.Project.Name` remains `"ExplicitName"` and no detection is attempted.

---

### Edge Cases

- What happens when a manifest file exists but is malformed?
- What happens when multiple manifest files exist (which wins)?
- What happens when the manifest name field is present but empty?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST auto-detect the project name only when `project.name` is empty in `ralph.toml`.
- **FR-002**: System MUST check manifest files in priority order: `pyproject.toml`, `package.json`, `Cargo.toml`.
- **FR-003**: For `pyproject.toml`, system MUST check `[project] name` first, then fall back to `[tool.poetry] name`.
- **FR-004**: For `package.json`, system MUST use the top-level `"name"` field.
- **FR-005**: For `Cargo.toml`, system MUST use `[package] name`.
- **FR-006**: System MUST silently ignore missing or malformed manifest files (no error).
- **FR-007**: An explicit `project.name` in `ralph.toml` MUST always take precedence; detection MUST NOT be attempted when set.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `Load()` sets `cfg.Project.Name` from pyproject.toml `[project] name` when ralph.toml has empty name.
- **SC-002**: `Load()` falls back to `[tool.poetry] name` when `[project] name` is absent.
- **SC-003**: `Load()` falls back to `package.json` when pyproject.toml is absent.
- **SC-004**: `Load()` falls back to `Cargo.toml` when pyproject.toml and package.json are absent.
- **SC-005**: Explicit `project.name` in ralph.toml is never overwritten.
- **SC-006**: Missing or malformed manifest files produce no error.
- **SC-007**: Detection is skipped entirely when `project.name` is already set.
