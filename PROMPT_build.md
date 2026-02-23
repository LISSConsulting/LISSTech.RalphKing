You are building **RalphKing** — a spec-driven AI coding loop CLI in Go.

## What you are building

Ralph is a Go CLI that replaces `loop.sh` with a proper binary. It runs Claude Code against specs in a continuous loop: plan → build → commit → push → repeat. The Regent watches Ralph, detects crashes/hangs, rolls back bad commits, and restarts him.

Read `specs/ralph-core.md` and `specs/the-regent.md` before doing anything else.
Read `CLAUDE.md` for architecture, build commands, and rules.
Read `IMPLEMENTATION_PLAN.md` for current status and what to work on next.

## Your task

1. Study the specs and implementation plan. Pick the highest-priority incomplete item.
2. Before writing anything, search the codebase to understand what already exists.
3. Implement the feature fully — no placeholders, no stubs.
4. Run `go build ./...` and `go test ./...` — both must pass before committing.
5. Also run `go vet ./...` — must be clean.
6. Commit with a descriptive message: `feat(scope): what you did`.
7. After committing, update `IMPLEMENTATION_PLAN.md` to reflect progress.
8. If you discover bugs or spec inconsistencies, fix or document them in `IMPLEMENTATION_PLAN.md`.

## Rules (from CLAUDE.md)

- **Specs are law.** Every feature must trace to a spec.
- **Idiomatic Go.** Stdlib first. Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`. Justify any new dependency.
- **Table-driven tests.** Use `t.Run`. Target 80% coverage.
- **No global mutable state.** Pass deps explicitly.
- **Errors are values.** Wrap with `fmt.Errorf("context: %w", err)`. Never swallow.
- **Small packages, clear boundaries.** No circular imports.

## Important

- Keep `IMPLEMENTATION_PLAN.md` current — future loops depend on it.
- When `go test ./...` passes and something meaningful is done, create/increment a git tag (start at `0.0.1`).
- If `IMPLEMENTATION_PLAN.md` gets large, prune completed items.
- Think carefully before adding dependencies — justify each one.
