
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: P1–P4 complete. Spec commands, core loop, ClaudeAgent, and CLI wiring implemented. 95% test coverage on spec package.

## Completed Work

| Feature | Spec | Tag |
|---------|------|-----|
| Repo scaffold (stub CLI, specs, CLAUDE.md, ralph.toml) | — | — |
| Config package — TOML parsing, defaults, walk-up discovery, init | ralph-core.md | 0.0.1 |
| Git package — branch, pull/push, stash, revert, diff helpers | ralph-core.md | 0.0.1 |
| Claude package — Agent interface, event types, stream-JSON parser | ralph-core.md | 0.0.1 |
| Cobra CLI skeleton — root + plan/build/run/status/init/spec commands | ralph-core.md | 0.0.1 |
| Loop package — Loop struct, Run method, iteration cycle (stash/pull/claude/push) | ralph-core.md | 0.0.2 |
| ClaudeAgent — implements claude.Agent, spawns claude -p subprocess | ralph-core.md | 0.0.2 |
| GitOps interface — consumer-side interface for testable git operations | ralph-core.md | 0.0.2 |
| CLI wiring — plan/build/run/status commands connected to loop | ralph-core.md | 0.0.2 |
| Smart run — plan if no IMPLEMENTATION_PLAN.md, then build | ralph-core.md | 0.0.2 |
| Signal handling — SIGINT/SIGTERM graceful shutdown via context | ralph-core.md | 0.0.2 |
| Status command — reads .ralph/regent-state.json, prints summary | the-regent.md | 0.0.2 |
| Spec package — discovery, status detection, template scaffolding | ralph-core.md | 0.0.3 |
| `ralph spec list` — list specs with ✅/🔄/⬜ status indicators | ralph-core.md | 0.0.3 |
| `ralph spec new <name>` — create spec from embedded template, open $EDITOR | ralph-core.md | 0.0.3 |

## Remaining Work (Prioritized)

### P5 — TUI (`internal/tui/`) — ralph-core.md

- `bubbletea` model with header bar, scrollable log, footer bar
- Header: `RalphKing  |  branch  |  iter N/M  |  cost $X.XX`
- Log: timestamped tool events (read, write, bash, result, error)
- Footer: `[pull] [push]  last commit: ...  |  q to quit`
- Color-coded: reads=blue, writes=green, bash=yellow, errors=red, regent=orange
- Replace loop's io.Writer log with TUI event consumption

### P6 — Regent (`internal/regent/`) — the-regent.md

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
- GitOps interface defined at consumer (loop package) for clean testability — *git.Runner satisfies it implicitly
- Spec status detection uses IMPLEMENTATION_PLAN.md cross-referencing — reference spec filenames in remaining work headers for accurate 🔄 detection

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
- CI release pipeline (exists already in `.github/`)
