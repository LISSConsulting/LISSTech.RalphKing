
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All core features complete (P1–P6).** Both specs (`ralph-core.md`, `the-regent.md`) fully implemented. 84–97% test coverage across all packages.

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
| Regent supervisor — crash detection, retry with backoff, max retries | the-regent.md | 0.0.5 |
| Regent hang detection — output timeout tracking, kill and restart | the-regent.md | 0.0.5 |
| Regent state persistence — `.ralph/regent-state.json` read/write | the-regent.md | 0.0.5 |
| Regent test runner — `test_command` execution, revert on failure, push revert | the-regent.md | 0.0.5 |
| Regent TUI integration — `LogRegent` kind, orange `regentStyle`, inline messages | the-regent.md | 0.0.5 |
| Regent CLI wiring — `regent.enabled` toggles supervision for plan/build/run | the-regent.md | 0.0.5 |
| Git package tests — conflict fallback, push rejection, error paths (75.5% → 96.2%) | ralph-core.md | 0.0.6 |
| CI/build hygiene — Go 1.24 in CI, version injection, gitignore binary, go mod tidy, tag normalization | — | 0.0.7 |

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
- Regent uses RunFunc abstraction (`func(ctx) error`) to supervise any loop variant (plan, build, smart run)
- Regent hang detection uses a ticker goroutine checking `lastOutputAt` every `hangTimeout/4`; cancelled when the loop context is done
- Regent TUI wiring uses two channels: loopEvents → forwarding goroutine (updates state) → tuiEvents; Regent emits directly to tuiEvents
- Regent no-TUI wiring uses a single shared channel with a drain goroutine; both loop and Regent write non-blocking

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
- CI release pipeline (exists already in `.github/`)
