# Spec: `ralph init` Creates IMPLEMENTATION_PLAN.md

**Branch**: `002-v2-improvements`
**Created**: 2026-02-26
**Status**: Proposed

## Overview

When `ralph init` scaffolds a new project, it should also create a starter
`IMPLEMENTATION_PLAN.md` file. The plan prompt (`PROMPT_plan.md`) instructs
Claude to create or update `IMPLEMENTATION_PLAN.md`, but without an initial
file, new projects have no structure for Claude to build on. Providing a
starter template ensures the plan phase has a clear starting point.

## User Scenarios & Testing

### Scenario 1 — Fresh project gets IMPLEMENTATION_PLAN.md

Given a directory with no existing files,
when `ralph init` runs,
then `IMPLEMENTATION_PLAN.md` is created with a starter template that includes
all standard sections.

### Scenario 2 — Existing IMPLEMENTATION_PLAN.md is not overwritten

Given a directory that already contains `IMPLEMENTATION_PLAN.md`,
when `ralph init` runs,
then the file is left unchanged (idempotent).

### Scenario 3 — IMPLEMENTATION_PLAN.md listed in created files

Given a fresh project directory,
when `ralph init` runs,
then the returned `created` list includes the path to `IMPLEMENTATION_PLAN.md`.

## Requirements

### R1 — Create IMPLEMENTATION_PLAN.md in ScaffoldProject

`ScaffoldProject` must create `IMPLEMENTATION_PLAN.md` in the project root
when the file does not already exist.

### R2 — Starter template content

The created file must contain the following sections:

1. A one-line project summary placeholder at the top.
2. `## Completed Work` — empty table with `Phase`, `Features`, `Tags` columns.
3. `## Remaining Work` — empty table with `Priority`, `Item`, `Location`,
   `Notes` columns.
4. `## Key Learnings` — empty bullet point placeholder.

### R3 — Idempotent

If `IMPLEMENTATION_PLAN.md` already exists (any content), `ScaffoldProject`
must not modify it and must not include it in the `created` list.

### R4 — Placement in created list

When created, `IMPLEMENTATION_PLAN.md` must appear in the `created` list after
`.gitignore` (i.e., as the last item, consistent with the append-only ordering
of existing scaffolded files).

## Success Criteria

- `ScaffoldProject` creates `IMPLEMENTATION_PLAN.md` with all required
  sections in a fresh directory.
- Calling `ScaffoldProject` twice in the same directory results in
  `IMPLEMENTATION_PLAN.md` being listed in `created` only on the first call.
- `go test ./internal/config/...` passes with coverage on the new code paths.
