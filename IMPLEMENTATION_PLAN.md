
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: P1 foundation complete. Config, git, claude packages and cobra CLI skeleton implemented with 86% test coverage.

## Completed Work

| Feature | Spec | Tag |
|---------|------|-----|
| Repo scaffold (stub CLI, specs, CLAUDE.md, ralph.toml) | — | — |
| Config package — TOML parsing, defaults, walk-up discovery, init | ralph-core.md | 0.0.1 |
| Git package — branch, pull/push, stash, revert, diff helpers | ralph-core.md | 0.0.1 |
| Claude package — Agent interface, event types, stream-JSON parser | ralph-core.md | 0.0.1 |
| Cobra CLI skeleton — root + plan/build/run/status/init/spec commands | ralph-core.md | 0.0.1 |

## Remaining Work (Prioritized)

### P2 — Core Loop (`internal/loop/`)

- **`loop.go`** — `Loop` struct, `Run(ctx, config, mode) error`
  - Mode: `plan` or `build`
  - Each iteration: stash if dirty → pull → run Claude → push if new commits
  - Respects `max_iterations` (0 = unlimited)
  - Emits iteration events for TUI consumption

- **`runner.go`** — Claude process management
  - Wraps `internal/claude` agent with `ClaudeAgent` implementation
  - Spawns `claude -p` subprocess with `--output-format=stream-json --verbose`
  - Reads prompt file, feeds to Claude
  - Returns parsed events via channel

### P3 — Commands (wire up loop to CLI)

- **`ralph plan [--max N]`** — run plan loop, feed `plan.prompt_file`
- **`ralph build [--max N]`** — run build loop, feed `build.prompt_file`
- **`ralph run [--max N]`** — smart: plan if no `IMPLEMENTATION_PLAN.md`, then build
- **`ralph init`** — ✅ wired to `config.InitFile`
- **`ralph status`** — read `.ralph/regent-state.json`, print summary

### P4 — Spec Commands

- **`ralph spec list`** — list `specs/*.md` with status (✅ complete per IMPLEMENTATION_PLAN, ⬜ not started)
- **`ralph spec new <name>`** — create `specs/<name>.md` from Spec Kit template, open `$EDITOR`

### P5 — TUI (`internal/tui/`)

- `bubbletea` model with header bar, scrollable log, footer bar
- Header: `👑 RalphKing  │  branch  │  iter N/M  │  cost $X.XX`
- Log: timestamped tool events (📖 read, ✏️ write, 🔧 bash, ✅ result, ❌ error)
- Footer: `[⬆ pull] [⬇ push]  last commit: ...  │  q to quit`
- Color-coded: reads=blue, writes=green, bash=yellow, errors=red, regent=orange
- Initial implementation can use simple `fmt.Println` output (loop works first, pretty TUI second)

### P6 — Regent (`internal/regent/`)

- `regent.go` — supervisor goroutine: crash detection, hang detection, restart with backoff
- `state.go` — read/write `.ralph/regent-state.json`
- `tester.go` — run `test_command`, revert on failure
- Wire into `ralph build` / `ralph run` when `regent.enabled = true`

## Key Learnings

- Go module: `github.com/LISSConsulting/LISSTech.RalphKing`
- `go 1.23` — use modern Go idioms
- Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`
- Build target: `go build ./cmd/ralph/`
- Test: `go test ./...`
- Vet: `go vet ./...`
- Cross-compile: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`
- Start tags at `0.0.1`, increment patch per meaningful milestone

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
- CI release pipeline (exists already in `.github/`)
