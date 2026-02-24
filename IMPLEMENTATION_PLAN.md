
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All core features complete + production polish.** Both specs (`ralph-core.md`, `the-regent.md`) fully implemented. 96-99% test coverage across all internal packages.

## Completed Work

| Milestone | Features | Tags |
|-----------|----------|------|
| Foundation | Config (TOML parsing, defaults, walk-up discovery, init, validation), Git (branch, pull/push, stash, revert, diff), Claude (Agent interface, events, stream-JSON parser), Cobra CLI skeleton | 0.0.1 |
| Core loop | Loop (Run, iteration cycle: stash/pull/claude/push), ClaudeAgent (spawns `claude -p`), GitOps interface, CLI wiring, smart run, signal handling, status command | 0.0.2 |
| Spec management | Spec discovery, status detection, `ralph spec list`, `ralph spec new <name>` | 0.0.3 |
| TUI | Bubbletea model (header, scrollable log, footer), event system, lipgloss styles, `--no-tui` flag, scrollable history (j/k, pgup/pgdown, g/G) | 0.0.4, 0.0.11 |
| Regent supervisor | Crash detection + retry with backoff, hang detection (output timeout), state persistence, test-gated rollback (per-iteration), TUI integration, CLI wiring | 0.0.5, 0.0.10 |
| Test coverage | Git error paths (96.2%), TUI (99.3%), loop (97.6%), regent (96.0%), stateTracker unit tests | 0.0.6, 0.0.14, 0.0.16 |
| CI/CD | Go 1.24, version injection, race detection, release workflow (cross-compiled binaries on tag push) | 0.0.7, 0.0.19 |
| Status & state | Formatted status display, running-state detection, stateTracker for non-Regent paths, Regent context-cancel persistence | 0.0.8, 0.0.13-0.0.16 |
| Hardening | Stream-JSON `is_error` handling, `DiffFromRemote` error distinction, config validation gating, stale closure fix, ClaudeAgent stderr capture, TUI error propagation | 0.0.12, 0.0.17-0.0.20 |
| Prompt files | `PROMPT_build.md` (build loop instructions), `PROMPT_plan.md` (plan loop instructions) | 0.0.21 |
| TUI config | Configurable accent color via `[tui] accent_color` in ralph.toml (spec: "configurable, default indigo") | v0.0.22 |
| TUI polish | "New messages below" indicator (`↓N new`) in footer when scrolled up and events arrive | v0.0.23 |
| State tracking | stateTracker live persistence: save to disk on meaningful state changes so `ralph status` works mid-loop without Regent | v0.0.24 |
| Refactoring | Split `cmd/ralph/main.go` into main/commands/execute/wiring, removed dead TUI code | 0.0.9 |

## Key Learnings

- Go module: `github.com/LISSConsulting/LISSTech.RalphKing`, Go 1.24
- Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`
- Build: `go build ./cmd/ralph/` | Test: `go test ./...` | Vet: `go vet ./...`
- Cross-compile: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`
- Tags: start at 0.0.1, increment patch per milestone, `v` prefix from v0.0.10+
- GitOps interface at consumer (loop package) for testability — `*git.Runner` satisfies implicitly
- Loop `emit()` non-blocking send prevents deadlock when TUI exits early
- TUI uses bubbletea channel pattern: `waitForEvent` Cmd reads from `<-chan LogEntry`
- Regent uses `RunFunc` abstraction (`func(ctx) error`) to supervise any loop variant
- Regent hang detection: ticker goroutine checks `lastOutputAt` every `hangTimeout/4`
- Per-iteration rollback via `Loop.PostIteration` hook wired to `Regent.RunPostIterationTests`
- TUI scroll: `scrollOffset` 0 = bottom; auto-scroll only when at bottom
- `stateTracker` mirrors Regent.UpdateState() for non-Regent paths, including live persistence on meaningful changes
- Closures passed to Regent must re-evaluate filesystem state inside the closure body (not capture stale values)
- `Config.Validate()` is pure (no I/O) — prompt file existence checked at runtime by `os.ReadFile`
- Claude result events with `is_error: true` emit ErrorEvent then ResultEvent (preserves cost tracking)
- `git diff --quiet` exit 128 + "fatal:" = error; exit 1 = real diff; `pushIfNeeded` pushes on error
- Accent-dependent TUI styles (header, git) live on Model as instance fields; non-accent styles remain package vars
- TUI `newBelow` counter tracks events arriving while scrolled up; resets when `scrollOffset` returns to 0

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
