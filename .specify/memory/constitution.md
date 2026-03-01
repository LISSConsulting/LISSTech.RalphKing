<!--
Sync Impact Report
- Version change: 0.0.0 → 1.0.0
- Added principles:
  - I. Spec-Driven
  - II. Supervised Autonomy
  - III. Test-Gated Commits
  - IV. Idiomatic Go
  - V. Observable Loops
- Added sections:
  - Technical Constraints
  - Development Workflow
  - Governance
- Removed sections: none (initial constitution)
- Templates requiring updates:
  - .specify/templates/plan-template.md ✅ aligned (Constitution Check section compatible)
  - .specify/templates/spec-template.md ✅ aligned (user stories + acceptance scenarios compatible)
  - .specify/templates/tasks-template.md ✅ aligned (test-first ordering compatible)
- Follow-up TODOs: none
-->

# RalphKing Constitution

## Core Principles

### I. Spec-Driven

Every feature, command, and behaviour MUST originate from a spec in `specs/`.
Code serves specifications — specifications do not serve code.
No implementation work begins without a written spec that defines the "what"
and "why". The spec is the source of truth; divergence between spec and
implementation is a bug.

### II. Supervised Autonomy

Ralph runs autonomously but MUST NOT run unsupervised in production use.
The Regent MUST watch every Ralph session, detect crashes, hangs, and test
regressions, and take corrective action (restart, revert, escalate).
Safety takes precedence over speed — a reverted bad commit is preferable
to a broken main branch.

### III. Test-Gated Commits

When `regent.test_command` is configured and `regent.rollback_on_test_failure`
is enabled, no commit survives a failing test suite. The Regent MUST revert
any commit that causes test regressions and resume the loop. Tests are the
final arbiter of commit quality. New features SHOULD be developed test-first
(red-green-refactor) whenever practical.

### IV. Idiomatic Go

All code MUST follow standard Go conventions:

- Small, focused packages with clear responsibilities
- Exported interfaces, unexported implementations
- `go vet`, `go fmt`, and `go-critic` MUST pass with zero warnings
- Prefer stdlib over third-party dependencies; justify every external import
- Error handling via explicit return values — no panics for recoverable errors
- Table-driven tests as the default test pattern

### V. Observable Loops

Every iteration MUST be visible to the operator. The TUI MUST display:
timestamps, tool calls, cost per iteration, running cost total, git
operations, and Regent activity. Silent failures are forbidden — if
something fails, it MUST appear in the TUI log. Structured state
(`.ralph/regent-state.json`) MUST be written so that `ralph status` can
report on any session, running or completed.

## Technical Constraints

- **Language**: Go 1.24+
- **Build targets**: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
- **Configuration**: TOML (`ralph.toml`) parsed with `BurntSushi/toml`
- **TUI**: `bubbletea` + `lipgloss` — no raw ANSI escape sequences
- **CLI framework**: `cobra` for command routing
- **Agent interface**: `internal/claude/Agent` interface — Claude is the
  default implementation; additional agents (OpenAI, Gemini) are future work
- **Dependencies**: Minimise external deps; every addition MUST be justified
  in the PR description
- **Binary size**: Keep reasonable — no embedded assets beyond templates

## Development Workflow

1. **Specify** — Write or update a spec in `specs/` describing the feature
2. **Plan** — Produce an implementation plan referencing the spec
3. **Implement** — Build against the plan, committing after each logical unit
4. **Test** — Run `go test ./...` and `go vet ./...`; all MUST pass
5. **Review** — Verify implementation matches the spec's acceptance criteria
6. **Ship** — Cross-compile, tag release, update README if commands changed

Git workflow per iteration (automated by Ralph):

- `git pull --rebase` before push (fallback to merge on conflict)
- `git push` after each successful iteration
- Regent reverts on test failure when configured

## Governance

This constitution supersedes all ad-hoc practices. Amendments require:

1. A written proposal describing the change and its rationale
2. Update to this file with version bump (semver)
3. Propagation check across `.specify/templates/` and `README.md`
4. Commit with message format: `docs: amend constitution to vX.Y.Z (<summary>)`

All PRs and code reviews MUST verify compliance with these principles.
Complexity beyond what a principle permits MUST be justified in writing.

**Version**: 1.0.0 | **Ratified**: 2026-02-23 | **Last Amended**: 2026-02-23
