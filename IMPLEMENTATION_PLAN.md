
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: P1–P5 complete. TUI with bubbletea/lipgloss, structured event system, color-coded log, header/footer bars. 85%+ test coverage across all packages.

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
| `ralph spec list` — list specs with status indicators | ralph-core.md | 0.0.3 |
| `ralph spec new <name>` — create spec from embedded template, open $EDITOR | ralph-core.md | 0.0.3 |
| TUI — bubbletea model with header, scrollable log, footer | ralph-core.md | 0.0.4 |
| Loop event system — LogEntry/LogKind types, emit() replaces logf() | ralph-core.md | 0.0.4 |
| TUI styles — lipgloss color-coded tool display (reads=blue, writes=green, bash=yellow, errors=red) | ralph-core.md | 0.0.4 |
| TUI CLI wiring — `--no-tui` flag, alt-screen mode, event channel bridge | ralph-core.md | 0.0.4 |

## Remaining Work (Prioritized)

### P6 — Regent (`internal/regent/`) — the-regent.md

- `regent.go` — supervisor goroutine: crash detection, hang detection, restart with backoff
- `state.go` — read/write `.ralph/regent-state.json`
- `tester.go` — run `test_command`, revert on failure
- Wire into `ralph build` / `ralph run` when `regent.enabled = true`

## Key Learnings

- Go module: `github.com/LISSConsulting/LISSTech.RalphKing`
- `go 1.24` — bumped from 1.23 by bubbletea dependency
- Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`
- Build target: `go build ./cmd/ralph/`
- Test: `go test ./...`
- Vet: `go vet ./...`
- Cross-compile: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`
- Start tags at `0.0.1`, increment patch per meaningful milestone
- GitOps interface defined at consumer (loop package) for clean testability — *git.Runner satisfies it implicitly
- Spec status detection uses IMPLEMENTATION_PLAN.md cross-referencing — reference spec filenames in remaining work headers for accurate status detection
- Loop emit() is non-blocking on the event channel to prevent deadlock when TUI exits before loop finishes
- TUI uses bubbletea channel pattern: `waitForEvent` Cmd reads from `<-chan LogEntry`, re-schedules itself after each message

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
- CI release pipeline (exists already in `.github/`)
