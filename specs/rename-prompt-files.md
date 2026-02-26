# Spec: Rename Prompt Files and Chronicle

**Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Proposed

## Overview

The default filenames used by `ralph init` for prompt files and the implementation
plan are verbose and inconsistent with the project's naming conventions. This spec
defines cleaner, shorter names for the three key scaffold files:

| Old name | New name |
|----------|----------|
| `PROMPT_plan.md` | `PLAN.md` |
| `PROMPT_build.md` | `BUILD.md` |
| `IMPLEMENTATION_PLAN.md` | `CHRONICLE.md` |

The rename affects: scaffold file creation, default config values, the `ralph run`
smart-run plan-phase detection, and the `ralph spec list` status cross-reference.

## Why

- `PROMPT_plan.md` / `PROMPT_build.md` expose an internal implementation detail
  (`PROMPT_`) that users don't need to see. `PLAN.md` and `BUILD.md` are clearer
  about what the files contain (the plan phase instructions, the build phase
  instructions).
- `IMPLEMENTATION_PLAN.md` is unwieldy. `CHRONICLE.md` is evocative of its purpose:
  a running record of what has been done and what remains.
- Shorter names are easier to type and reference in prompts (e.g. `@CHRONICLE.md`
  vs `@IMPLEMENTATION_PLAN.md`).

## Scope

### New projects (`ralph init`)

`ScaffoldProject` creates `PLAN.md`, `BUILD.md`, and `CHRONICLE.md` instead of the
old names. The ralph.toml template written by `InitFile` references the new names.
The template content inside PLAN.md and BUILD.md refers to `@CHRONICLE.md`.

### Default config values

`config.Defaults()` returns `PLAN.md` and `BUILD.md` as the default prompt file
paths for the plan and build modes respectively. These defaults are used when
ralph.toml does not override the values.

### Smart-run plan-phase detection

`executeSmartRun` checks whether `CHRONICLE.md` exists (and is non-empty) to
decide whether the plan phase must run before the build phase.

### Spec list status cross-reference

`spec.List()` reads `CHRONICLE.md` (instead of `IMPLEMENTATION_PLAN.md`) when
checking whether a spec has been implemented.

## Breaking Change

This is a **breaking change** for existing projects that:

1. Rely on the default `plan.prompt_file = "PROMPT_plan.md"` (without specifying it
   in ralph.toml), **and**
2. Have `PROMPT_plan.md` / `PROMPT_build.md` files under those names.

**Migration path:** Add explicit `prompt_file` entries in ralph.toml pointing to the
old filenames, OR rename the files to `PLAN.md` / `BUILD.md`. Projects that already
set explicit `prompt_file` values in ralph.toml are unaffected.

Similarly, `ralph run` smart-phase detection and `ralph spec list` status display
will look for `CHRONICLE.md` instead of `IMPLEMENTATION_PLAN.md`. Rename existing
`IMPLEMENTATION_PLAN.md` to `CHRONICLE.md` to preserve this behaviour.

## Requirements

### R1 — Scaffold creates new filenames

`ScaffoldProject` must create:
- `PLAN.md` (not `PROMPT_plan.md`)
- `BUILD.md` (not `PROMPT_build.md`)
- `CHRONICLE.md` (not `IMPLEMENTATION_PLAN.md`)

Idempotency rules remain the same: skip files that already exist.

### R2 — Template content references CHRONICLE.md

The content of the created `PLAN.md` and `BUILD.md` files must reference
`CHRONICLE.md` (not `IMPLEMENTATION_PLAN.md`) as the state file.

### R3 — Default config values updated

`config.Defaults()` must return:
- `Plan.PromptFile = "PLAN.md"`
- `Build.PromptFile = "BUILD.md"`

### R4 — InitFile template updated

The ralph.toml template written by `InitFile` must use:
```toml
[plan]
prompt_file = "PLAN.md"

[build]
prompt_file = "BUILD.md"
```

### R5 — Smart-run detects CHRONICLE.md

`executeSmartRun` checks `filepath.Join(dir, "CHRONICLE.md")` for plan-phase
detection.

### R6 — Spec list reads CHRONICLE.md

`spec.List()` reads `CHRONICLE.md` for status cross-reference.

## Acceptance Criteria

- [ ] `ralph init` in a fresh directory creates `PLAN.md`, `BUILD.md`,
      `CHRONICLE.md` (not the old names).
- [ ] The created ralph.toml references `PLAN.md` and `BUILD.md`.
- [ ] `PLAN.md` and `BUILD.md` content references `CHRONICLE.md`.
- [ ] `config.Defaults()` returns `PLAN.md` and `BUILD.md`.
- [ ] `ralph run` in a project without `CHRONICLE.md` triggers the plan phase.
- [ ] `ralph spec list` cross-references `CHRONICLE.md` for status.
- [ ] All existing tests pass with updated file name expectations.
- [ ] New tests cover idempotency for `CHRONICLE.md` (analogous to the old
      `IMPLEMENTATION_PLAN.md` tests).
