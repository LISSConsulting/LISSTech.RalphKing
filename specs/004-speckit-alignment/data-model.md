# Data Model: Spec Kit Alignment

**Feature**: 004-speckit-alignment
**Date**: 2026-03-01

## Entities

### SpecFeature (refactored from SpecFile)

Represents a single feature in the spec kit workflow. Can be either a directory-based spec kit feature or a legacy flat `.md` file.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Feature identifier (directory name for spec kit, filename without `.md` for legacy) |
| Dir | string | Relative path to feature directory (e.g., `specs/004-speckit-alignment`). Empty for legacy flat files. |
| Path | string | Relative path to primary artifact (`spec.md` for directories, the `.md` file for legacy) |
| Status | FeatureStatus | Current workflow phase, derived from artifact presence |
| IsDir | bool | True if directory-based spec kit feature; false if legacy flat file |

### FeatureStatus (refactored from Status)

Represents the workflow phase of a spec feature, derived from which artifacts exist.

| Value | Condition | Symbol | Label |
|-------|-----------|--------|-------|
| `specified` | `spec.md` exists, no `plan.md` | `📋` | "specified" |
| `planned` | `plan.md` exists, no `tasks.md` | `📐` | "planned" |
| `tasked` | `tasks.md` exists | `✅` | "tasked" |
| `not_started` | Directory exists but no `spec.md` (or legacy: not in CHRONICLE.md) | `⬜` | "not started" |
| `done` | Legacy only: appears in CHRONICLE.md "Completed Work" | `✅` | "done" |
| `in_progress` | Legacy only: appears in CHRONICLE.md "Remaining Work" | `🔄` | "in progress" |

### ActiveSpec

Resolved context for speckit command execution.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Feature name (e.g., `004-speckit-alignment`) |
| Dir | string | Absolute path to spec directory |
| Branch | string | Git branch name that resolved to this spec |
| Explicit | bool | True if resolved via `--spec` flag, false if via branch |

### SpecKitArtifact (informational — not a Go type)

Files within a spec feature directory. Not a runtime struct — just documentation of the canonical layout.

| Artifact | Required | Created By |
|----------|----------|------------|
| `spec.md` | Yes | `ralph specify` / `/speckit.specify` |
| `plan.md` | No | `ralph plan` / `/speckit.plan` |
| `tasks.md` | No | `ralph tasks` / `/speckit.tasks` |
| `research.md` | No | `/speckit.plan` (Phase 0) |
| `data-model.md` | No | `/speckit.plan` (Phase 1) |
| `quickstart.md` | No | `/speckit.plan` (Phase 1) |
| `checklists/requirements.md` | No | `/speckit.specify` |
| `contracts/*.md` | No | `/speckit.plan` (Phase 1) |

## Relationships

```
SpecFeature 1──1 FeatureStatus    (derived from artifacts)
SpecFeature 1──* SpecKitArtifact  (files in directory)
ActiveSpec  1──1 SpecFeature      (resolved at command time)
```

## State Transitions

```
[empty dir] ──specify──→ specified (spec.md created)
specified   ──plan────→ planned   (plan.md created)
planned     ──tasks───→ tasked    (tasks.md created)
tasked      ──run─────→ tasked    (tasks executed, status stays tasked)
```

Note: There is no "done" status for directory-based specs in v1. Task completion tracking (parsing `tasks.md` checkboxes) is deferred to a future iteration.
