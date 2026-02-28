# CLAUDE.md — RalphKing

Spec-driven AI coding loop CLI in Go. Ralph runs Claude Code against specs in a continuous loop; The Regent supervises for crashes, hangs, and test regressions.

## Architecture

```
Regent (supervisor) → Ralph (loop) → Claude (worker)
```

- `cmd/ralph/main.go` — CLI entry point (cobra)
- `internal/config/` — TOML config parsing (`ralph.toml`)
- `internal/claude/` — Claude CLI adapter, stream-JSON event parser
- `internal/git/` — pull, push, branch helpers
- `internal/loop/` — Core iteration: prompt → claude → parse → git
- `internal/spec/` — Spec file discovery and templating
- `internal/tui/` — bubbletea + lipgloss TUI
- `internal/regent/` — Supervisor: crash/hang detection, test-gated rollback

Specs live in `specs/`. Read `specs/ralph-core.md` and `specs/the-regent.md` before implementing anything.

## Build & Test

```sh
go build ./cmd/ralph/         # build
go test ./...                 # test
go vet ./...                  # must pass with zero warnings
```

Cross-compile targets: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`. CI handles this — see `.github/workflows/ci.yml`.

## Rules

- **Specs are law.** Every feature originates from a spec in `specs/`. Read the spec before writing code.
- **Idiomatic Go.** Standard library first. Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`. Justify any new dependency.
- **Table-driven tests.** Use `t.Run` subtests. Target 80% coverage.
- **No global mutable state.** Pass dependencies explicitly. Structs hold state, functions transform it.
- **Errors are values.** Wrap with `fmt.Errorf("context: %w", err)`. Never swallow errors silently.
- **Small packages, clear boundaries.** Each `internal/` package owns one concern. No circular imports.

## Governance

Project constitution at `.specify/memory/constitution.md`. Five principles: spec-driven, supervised autonomy, test-gated commits, idiomatic Go, observable loops.

## Config

`ralph.toml` at repo root is the example config. All config fields documented in `specs/ralph-core.md` under Configuration.
