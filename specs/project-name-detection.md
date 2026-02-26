# Project Name Auto-Detection

## Overview

When `project.name` is not set in `ralph.toml`, Ralph should auto-detect the project name from common manifest files in the same directory as `ralph.toml`.

## User Story

As a developer using Ralph on an existing project, I want the TUI to show my project's name without having to manually set it in `ralph.toml`, so that setup is effortless.

## Functional Requirements

1. **Detection trigger**: Auto-detection runs only when `project.name` is empty (blank string) in `ralph.toml`.
2. **Detection order**: The following files are checked in priority order:
   1. `pyproject.toml` — Python/UV/Poetry projects
   2. `package.json` — Node.js/npm projects
   3. `Cargo.toml` — Rust projects
3. **pyproject.toml name resolution**:
   - Check `[project] name` (PEP 621 standard, used by pip/uv)
   - If absent, check `[tool.poetry] name` (Poetry legacy format)
4. **package.json name resolution**: Use the top-level `"name"` field.
5. **Cargo.toml name resolution**: Use `[package] name`.
6. **Graceful fallback**: If no manifest file is found or parsing fails, `project.name` remains empty (TUI falls back to "RalphKing"). Detection errors are silently ignored.
7. **Explicit wins**: A non-empty `project.name` in `ralph.toml` always takes precedence; detection is never attempted.

## Implementation Location

- New function `DetectProjectName(dir string) string` in `internal/config/`
- Called inside `Load()` after TOML decode, when `cfg.Project.Name == ""`
- `dir` is the directory containing `ralph.toml`

## Acceptance Criteria

- `[ ]` `Load()` sets `cfg.Project.Name` from pyproject.toml `[project] name` when ralph.toml has empty name
- `[ ]` `Load()` sets `cfg.Project.Name` from pyproject.toml `[tool.poetry] name` when `[project] name` is absent
- `[ ]` `Load()` sets `cfg.Project.Name` from `package.json` top-level `name` when pyproject.toml absent
- `[ ]` `Load()` sets `cfg.Project.Name` from `Cargo.toml` `[package] name` when pyproject.toml and package.json absent
- `[ ]` Explicit `project.name` in ralph.toml is never overwritten
- `[ ]` Missing or malformed manifest files produce no error
- `[ ]` Detection is skipped entirely when `project.name` is already set
