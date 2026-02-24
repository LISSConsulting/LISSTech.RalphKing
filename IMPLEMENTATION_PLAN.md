
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All core features complete (P1–P6) + production polish.** Both specs (`ralph-core.md`, `the-regent.md`) fully implemented. 96–99% test coverage across all internal packages.

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
| `ralph status` — proper formatted display (branch, commit, iteration, cost, duration, pass/fail) | ralph-core.md | 0.0.8 |
| Regent state — added Branch, Mode, StartedAt, FinishedAt, Passed fields | the-regent.md | 0.0.8 |
| SIGQUIT immediate kill — platform-specific handler (unix/windows build tags) | the-regent.md | 0.0.8 |
| Refactor `cmd/ralph/main.go` — split 468-line monolith into `main.go` (55), `commands.go` (151), `execute.go` (166), `wiring.go` (120) | — | 0.0.9 |
| Remove dead code — `tui.RunLoop`, `tui.RunSmartLoop` replaced by direct `RunFunc` wiring | — | 0.0.9 |
| Per-iteration test-gated rollback — `Loop.PostIteration` hook, `Regent.RunPostIterationTests` called after each iteration instead of after loop completion | the-regent.md | 0.0.10 |
| TUI scrollable log history — j/k, pgup/pgdown, g/G, arrow keys; footer scroll indicator; auto-scroll at bottom | ralph-core.md | 0.0.11 |
| Regent live state persistence — `UpdateState()` persists to disk on meaningful changes so `ralph status` is accurate mid-loop | the-regent.md | 0.0.11 |
| Fix TUI error propagation — `runWithTUI` and `runWithRegentTUI` now capture and return loop/Regent errors via buffered error channel instead of silently swallowing them | ralph-core.md, the-regent.md | 0.0.12 |
| ClaudeAgent stderr capture — includes stderr text in error events when Claude CLI exits non-zero, replacing opaque "exit status 1" | ralph-core.md | 0.0.13 |
| `ralph status` running state — detects mid-run processes (StartedAt set, FinishedAt zero), shows elapsed time, last output, and "running" result | ralph-core.md, the-regent.md | 0.0.13 |
| `stateTracker` for non-Regent paths — `runWithStateTracking` / `runWithTUIAndState` persist `.ralph/regent-state.json` so `ralph status` works when `regent.enabled = false` | ralph-core.md | 0.0.14 |
| Test coverage push — tui 94%→99%, loop 90%→98%, regent 90%→94%; added mockGit error injection, stash/push/pull/revert error paths, renderLine LogRegent, toolStyle all branches | — | 0.0.14 |
| `showStatus` fail result fallback — non-Regent runs that fail now show "fail" instead of empty result (ConsecutiveErrs=0 path) | ralph-core.md | 0.0.15 |
| `stateTracker` unit tests — table-driven tests for init, trackEntry, zero-value preservation, save, finish (success/cancel/error), lastOutputAt | — | 0.0.15 |
| Regent context-cancel state persistence — `finishGraceful()` sets `FinishedAt`/`Passed=true` on all three cancel paths (pre-loop, post-failure, backoff) so `ralph status` no longer shows stale "running" after SIGINT | the-regent.md | 0.0.16 |
| `RevertLastCommit` error path tests — `LastCommit` failure, `CurrentBranch` failure; `mockGit` gains `currentBranchErr` field; regent coverage 94% → 96% | the-regent.md | 0.0.16 |
| Fix `ralph run` stale closure bug — `needsPlan` check moved inside `smartRunFn` closure so Regent retries re-evaluate whether `IMPLEMENTATION_PLAN.md` exists instead of using a stale captured value from startup | ralph-core.md | 0.0.17 |
| Config validation — `Config.Validate()` catches empty prompt files, negative iteration counts, invalid Regent settings (max_retries, backoff, hang_timeout), and `rollback_on_test_failure` without `test_command` before loop starts; reports all issues joined; config coverage 92.9% → 95.7% | ralph-core.md, the-regent.md | 0.0.18 |
| CI race detection — `-race` flag added to CI test step; catches concurrency bugs in Regent/loop goroutine code | — | 0.0.19 |
| Release workflow — `release.yml` creates GitHub Releases with cross-compiled binaries when version tags (`v*`) are pushed; runs tests with race detection before building | — | 0.0.19 |
| Stream-JSON parser `is_error` handling — result events with `is_error: true` now emit `ErrorEvent` before `ResultEvent`, so failed Claude runs are logged as errors while still tracking cost; fallback message when `result` field is empty | ralph-core.md | 0.0.20 |
| `DiffFromRemote` error distinction — `git diff --quiet` fatal errors (e.g., missing remote ref) now return errors instead of being silently promoted to "has changes"; `pushIfNeeded` pushes anyway on diff errors (handles new branches) | ralph-core.md | 0.0.20 |
| Config validation gating — `rollback_on_test_failure` without `test_command` check now gated on `regent.enabled`, preventing spurious validation errors for disabled Regent configs | the-regent.md | 0.0.20 |

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
- Per-iteration test-gated rollback uses `Loop.PostIteration` hook (wired to `Regent.RunPostIterationTests` in CLI wiring layer); errors are emitted as events, not returned, so the loop continues to the next iteration per spec
- TUI scroll uses `scrollOffset` (0 = bottom, >0 = lines from end); `renderLog` calculates `end = len(lines) - offset`, `start = end - height`; new entries auto-scroll only when offset is 0
- Regent `UpdateState()` persists to disk only when meaningful state fields change (iteration, cost, commit, branch, mode) — avoids unnecessary I/O for info-only events
- TUI wiring error propagation uses buffered error channel (`errCh := make(chan error, 1)`) + non-blocking `select/default` to safely collect errors without blocking when user quits early; `context.Canceled` is suppressed as normal shutdown
- ClaudeAgent captures stderr via `bytes.Buffer` on `cmd.Stderr`; on non-zero exit, stderr text is appended to the error event for diagnostics (e.g. "claude exited: exit status 1: API rate limit exceeded")
- `ralph status` running detection uses `!StartedAt.IsZero() && FinishedAt.IsZero()` — when the process crashes without cleanup, `FinishedAt` stays zero, so status shows "running" until the next loop overwrites state
- `stateTracker` pattern: a lightweight struct in `cmd/ralph/wiring.go` that tracks loop state and persists to `.ralph/regent-state.json` — used in non-Regent paths; mirrors what Regent.UpdateState() does for Regent paths
- Coverage gaps hard to test: `renderLog` defensive guards (`end < 0`, `end > len`) require out-of-bounds scrollOffset; CLI command handlers in `cmd/ralph` are integration-only (cobra + real deps); `stateTracker` now has unit tests
- `showStatus` result display has four tiers: running → pass → fail (N errors) → fail; the last fallback handles non-Regent runs where `Passed=false` and `ConsecutiveErrs=0`
- Regent `finishGraceful()` mirrors `stateTracker.finish()` semantics: context cancellation = user-initiated stop = `Passed=true`. All three cancel paths (pre-loop select, post-run ctx check, backoff select) now persist state before returning
- **Closures passed to Regent must re-evaluate state**: any `RunFunc` closure that checks filesystem state (e.g., file existence) must do so *inside* the closure body, not capture a variable computed outside. The Regent calls the closure multiple times on retry, so stale captured values cause incorrect behavior
- `Config.Validate()` is pure (no I/O) — checks structural correctness of values. Prompt file existence is still checked at runtime by `os.ReadFile` in `loop.Run`, which gives a clear error. All Regent checks (numeric + rollback) gated on `Regent.Enabled` since disabled Regent values are never used. `errors.Join` collects all issues into a single error for user-friendly reporting
- Claude CLI result events include `is_error` (bool) and `result` (string) fields — when `is_error` is true, the parser emits `ErrorEvent` (with result text) followed by `ResultEvent` (to preserve cost tracking). This handles rate limits, auth failures, and other structured error results that the CLI wraps in a result event
- `git diff --quiet` returns exit 1 for real diffs, but exit 128 + "fatal:" stderr for errors like missing remote refs. `DiffFromRemote` now checks for "fatal:" to distinguish the two; `pushIfNeeded` pushes on error (safe: `Push` handles `-u` fallback for new branches)

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
