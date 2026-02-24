
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All core features complete + hardened.** Both specs (`ralph-core.md`, `the-regent.md`) fully implemented. 96-99% test coverage across all internal packages.

## Completed Work

| Phase | Features | Tags |
|-------|----------|------|
| Foundation & core | Config (TOML, defaults, walk-up discovery, init, validation), Git (branch, pull/push, stash, revert, diff), Claude (Agent interface, events, stream-JSON parser), Loop (iteration cycle, ClaudeAgent subprocess, GitOps, smart run), Cobra CLI (plan/build/run/status/init/spec), signal handling | 0.0.1–0.0.3 |
| TUI | Bubbletea model (header/log/footer), lipgloss styles, `--no-tui`, scrollable history (j/k/pgup/pgdown/g/G), configurable accent color, `↓N new` indicator | 0.0.4, 0.0.11, v0.0.22–v0.0.23 |
| Regent supervisor | Crash detection + retry/backoff, hang detection (output timeout), state persistence, test-gated rollback (per-iteration), TUI integration, CLI wiring, graceful shutdown | 0.0.5, 0.0.10 |
| Hardening | Stream-JSON `is_error`/`scanner.Err()` handling, `DiffFromRemote` error distinction, config validation, ClaudeAgent stderr capture, TUI error propagation, stale closure fix, result subtype surface, unknown TOML key rejection, rebase abort error surfacing, LastCommit error fallback, signal goroutine leak fix, TUI long tool name truncation | 0.0.12, 0.0.17–0.0.20, v0.0.27, v0.0.29, v0.0.31, v0.0.32 |
| State & status | Formatted status display, running-state detection, stateTracker live persistence (non-Regent paths), Regent context-cancel persistence, `detectStatus` fallback | 0.0.8, 0.0.13–0.0.16, v0.0.24–v0.0.25 |
| Cost control | `claude.max_turns` config (0 = unlimited), `--max-turns` CLI passthrough | v0.0.26 |
| Scaffolding | `ralph init` creates ralph.toml + PROMPT_plan.md + PROMPT_build.md + specs/ (idempotent) | v0.0.28 |
| CI/CD | Go 1.24, version injection, race detection, release workflow (cross-compiled binaries on tag push), golangci-lint (go-critic + gofmt) in CI & release | 0.0.7, 0.0.19, v0.0.30 |
| Test coverage | Git 94.7%, TUI 100%, loop 97.7%, claude 97.8%, regent 96.0%, config 92.5%, spec 95.5%, cmd/ralph 53.4% | 0.0.6, 0.0.14, 0.0.16, v0.0.32, v0.0.36 |
| Refactoring | Split `cmd/ralph/main.go` into main/commands/execute/wiring, prompt files, extract `classifyResult`/`needsPlanPhase`/`formatStatus`/`formatLogLine`/`formatSpecList`/`formatScaffoldResult` pure functions with table-driven tests, command tree structure tests, end-to-end command execution tests (cmd/ralph 8.8% → 41.8%); added `runWithStateTracking`/`runWithRegent`/`openEditor` tests (41.8% → 53.4%) | 0.0.9, v0.0.21, v0.0.33–v0.0.36 |

Specs implemented: `ralph-core.md`, `the-regent.md`.

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
- `claude.max_turns` (0 = unlimited) passes `--max-turns N` to Claude CLI; complements Regent hang detection with explicit turn limits
- Result `subtype` (success, error_max_turns, etc.) threads through `Event.Subtype` → `LogEntry.Subtype` → TUI/log display; empty subtype omitted from output
- `ScaffoldProject` creates all files referenced by ralph.toml defaults (prompt files, specs dir); `InitFile` still available for ralph.toml-only creation
- `ParseStream` checks `scanner.Err()` after scan loop — surfaces I/O or buffer-overflow errors as error events rather than silently closing the channel
- `.golangci.yml` enables gocritic, gofmt, gosimple, govet, ineffassign, unused, errcheck — CI enforces constitution's "go vet, go fmt, go-critic MUST pass" rule
- `config.Load()` uses `toml.MetaData.Undecoded()` to reject unknown keys in ralph.toml — catches typos like `promptfile` instead of `prompt_file`
- `Pull()` surfaces `rebase --abort` errors in the merge-failure message for better diagnostics; if abort fails AND merge fails, both errors are reported
- `pushIfNeeded` handles `LastCommit()` errors with `"(unknown)"` fallback instead of showing empty commit info
- `signalContext()` selects on both signal channel and `ctx.Done()` to prevent goroutine leaks; calls `signal.Stop()` on exit
- TUI truncates tool names >14 chars with `"…"` to preserve columnar log layout
- `classifyResult(state)` is a pure function returning `statusResult` enum — six-state classification (no-state, running, pass, fail-with-errors, plain-fail) with documented priority order; `showStatus` delegates to it
- `needsPlanPhase(info, statErr)` is a pure function encoding the plan-skip condition: file missing OR empty; used by `executeSmartRun`'s closure
- `formatStatus(state, now)` is a pure function rendering status output as a string; `now` parameter pins time for deterministic tests; `showStatus` delegates to it
- `formatLogLine(entry)` renders a LogEntry as `[HH:MM:SS]  message` (with `🛡️  Regent:` prefix for LogRegent entries); replaces inline formatting in `runWithRegent` and `runWithStateTracking`
- `formatSpecList(specs)` and `formatScaffoldResult(created)` are pure functions extracted from `specListCmd` and `initCmd` RunE closures
- End-to-end command tests use Go 1.24 `t.Chdir()` to test RunE handlers (init, status, spec list, spec new) with real I/O against temp dirs
- Command tree structure tests verify rootCmd subcommands, --max/--no-tui flags, and spec subcommands by calling constructors
- `runWithStateTracking` tested via black-box: success/error/context.Canceled/events-forwarded — all 4 paths; `lp.Events` is set before `run` is called so closures can send safely; drain goroutine processes events before `<-drainDone` unblocks
- `runWithRegent` tested via black-box: success/max-retries-exceeded/context-canceled/loop-events-update-state — 100% coverage; `HangTimeoutSeconds=0` disables ticker, `RetryBackoffSeconds=0` skips backoff wait; non-Regent events in `run` func flow through drain goroutine's `rgt.UpdateState` branch
- `openEditor` tested via `specNewCmd` with `EDITOR="true"` (Unix no-op that exits 0); `findNoop()` helper uses `exec.LookPath("true")` so test skips gracefully on Windows
- Remaining 0% functions are OS-level (`main`, `signalContext`, `registerQuitHandler`) or TUI-required (`runWithRegentTUI`, `finishTUI`, `runWithTUIAndState`) — not worth testing without a real terminal

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
