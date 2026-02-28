
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All core features complete + hardened.** Both specs (`ralph-core.md`, `the-regent.md`) fully implemented. 96-99% test coverage across all internal packages; cmd/ralph 72.4%, overall ~89%. Re-audited 2026-02-26 via full code search across all spec requirements. All remaining work items resolved as of v0.0.40. SIGQUIT handling confirmed implemented in `quit_unix.go`. `spec.List()` subdirectory walk fixed in v0.0.39. v2 improvements branch (`002-v2-improvements`) active — bugs fixed in v0.0.42. `ralph init` now writes `.gitignore` with `.ralph/regent-state.json` entry in v0.0.43. Fixed race condition in `regent.SaveState` (concurrent writes via atomic rename) in v0.0.43. Fixed `runWithRegent` final-state race via `FlushState` in v0.0.67. `InitFile` write error path covered in v0.0.70 (config 92.3%→93.1%).

## Completed Work

| Phase | Features | Tags |
|-------|----------|------|
| Foundation & core | Config (TOML, defaults, walk-up discovery, init, validation), Git (branch, pull/push, stash, revert, diff), Claude (Agent interface, events, stream-JSON parser), Loop (iteration cycle, ClaudeAgent subprocess, GitOps, smart run), Cobra CLI (plan/build/run/status/init/spec), signal handling | 0.0.1–0.0.3 |
| TUI | Bubbletea model (header/log/footer), lipgloss styles, `--no-tui`, scrollable history (j/k/pgup/pgdown/g/G), configurable accent color, `↓N new` indicator | 0.0.4, 0.0.11, v0.0.22–v0.0.23 |
| Regent supervisor | Crash detection + retry/backoff, hang detection (output timeout), state persistence, test-gated rollback (per-iteration), TUI integration, CLI wiring, graceful shutdown | 0.0.5, 0.0.10 |
| Hardening | Stream-JSON `is_error`/`scanner.Err()` handling, `DiffFromRemote` error distinction, config validation, ClaudeAgent stderr capture, TUI error propagation, stale closure fix, result subtype surface, unknown TOML key rejection, rebase abort error surfacing, LastCommit error fallback, signal goroutine leak fix, TUI long tool name truncation, fix gocritic `ifElseChain` in `scaffold.go`/`view.go`, fix `gofmt` alignment in `cmd/ralph/*_test.go`, fix `runWithRegent` final-state race (`FlushState` after drain goroutine), fix `regent.emit` panic on closed channel (`recover()` guard) | 0.0.12, 0.0.17–0.0.20, v0.0.27, v0.0.29, v0.0.31, v0.0.32, v0.0.66, v0.0.67, v0.0.68 |
| State & status | Formatted status display, running-state detection, stateTracker live persistence (non-Regent paths), Regent context-cancel persistence, `detectStatus` fallback | 0.0.8, 0.0.13–0.0.16, v0.0.24–v0.0.25 |
| Cost control | `claude.max_turns` config (0 = unlimited), `--max-turns` CLI passthrough | v0.0.26 |
| Scaffolding | `ralph init` creates ralph.toml + PROMPT_plan.md + PROMPT_build.md + specs/ (idempotent) | v0.0.28 |
| CI/CD | Go 1.24, version injection, race detection, release workflow (cross-compiled binaries on tag push), golangci-lint (go-critic + gofmt) in CI & release | 0.0.7, 0.0.19, v0.0.30 |
| Test coverage | Git 93.2%, TUI 99.5%, loop 81.3% (runner.Run skipped on Windows), claude 97.8%, regent 95.9%, config 93.1%, spec 96.2%, notify 95.0%, cmd/ralph 72.4% | 0.0.6, 0.0.14, 0.0.16, v0.0.32, v0.0.36–v0.0.40, v0.0.56, v0.0.58–v0.0.63, v0.0.65, v0.0.68, v0.0.70, v0.0.71 |
| Refactoring | Split `cmd/ralph/main.go` into main/commands/execute/wiring, prompt files, extract `classifyResult`/`needsPlanPhase`/`formatStatus`/`formatLogLine`/`formatSpecList`/`formatScaffoldResult` pure functions with table-driven tests, command tree structure tests, end-to-end command execution tests (cmd/ralph 8.8% → 41.8%); added `runWithStateTracking`/`runWithRegent`/`openEditor` tests (41.8% → 53.4%); added `executeLoop`/`executeSmartRun` integration tests + plan/build/run RunE tests (53.4% → 70.7%); added config-invalid/regent-enabled/corrupted-state-file tests for `executeSmartRun` and `showStatus` (70.7% → 72.0%) | 0.0.9, v0.0.21, v0.0.33–v0.0.38 |

Specs implemented: `ralph-core.md`, `the-regent.md`.

## Remaining Work

| Priority | Item | Location | Notes |
|----------|------|----------|-------|
| Info | cmd/ralph coverage ceiling at 72.4% | `cmd/ralph/wiring.go` — `runWithRegentTUI`, `finishTUI`, `runWithTUIAndState`; `cmd/ralph/main.go` — `main`; `cmd/ralph/quit_unix.go`/`quit_windows.go` — `registerQuitHandler` | These functions require a real TTY (bubbletea) or are OS-level signal handlers. Not actionable without a bubbletea headless test mode. `internal/loop` runner.Run at 0% on Windows (shell script tests skipped); loop overall 81.3%. Remaining partial-coverage gaps (git.Stash "No local changes" path, regent.SaveState marshal/createTemp/write/close errors) are not reliably testable on modern git/cross-platform. |

## v2 Improvement Backlog (from GitHub Issues #1 and #2)

These items originate from user feedback. Items requiring new specs are noted; bug fixes can be done directly.

### TUI Improvements (Issue #1)
| Priority | Item | Status | Notes |
|----------|------|--------|-------|
| Bug | Stash error when no changes | ✅ Fixed v0.0.42 | `Stash()` now returns nil for "No local changes to save" |
| Bug | Task/TaskOutput tool inputs empty in TUI | ✅ Fixed v0.0.42 | `summarizeInput()` extended with `description`, `prompt`, `query`, `notebook_path`, `task_id` |
| Low | Replace app branding with project name | ✅ Fixed v0.0.48 | `tui.New()` accepts `projectName` param; header shows `👑 <project.name>` when set, falls back to "RalphKing" when empty |
| Low | Display current directory | ✅ Fixed v0.0.52 | `tui.New()` accepts `workDir string` 4th param; `renderHeader()` shows `dir: ~/abbreviated-path` after project name; `abbreviatePath()` replaces home dir with `~` and normalises backslashes; spec at `specs/current-directory.md` |
| Low | Display current time | ✅ Fixed v0.0.47 | Added `now time.Time` field updated by `tickMsg` every second; shown as `HH:MM` in header |
| Low | Display loop elapsed time | ✅ Fixed v0.0.47 | Added `startedAt time.Time` in `New()`; `formatElapsed()` renders compact duration (e.g. `2m35s`, `1h30m`); shown as `elapsed: X` in header |
| Low | Display last response elapsed time | ✅ Fixed v0.0.44 | Added `lastDuration` to TUI model; updated from `LogIterComplete` entries; shown as `last: %.1fs` in header (omitted until first iteration completes) |
| Low | Always display latest commit | ✅ Fixed v0.0.46 | `loop.Run()` now calls `LastCommit()` at startup and includes `Commit` in the initial `LogInfo` event; TUI footer shows HEAD commit from first render instead of `—` |
| High | Show agent's reasoning | ✅ Fixed v0.0.51 | `LogText` kind added; loop emits `LogText` for `claude.EventText` events; TUI renders with 💭 icon in muted gray style (truncated to 80 chars); spec at `specs/agent-reasoning.md` |
| Low | Truncate long commands | ✅ Fixed v0.0.45, improved v0.0.62 | `renderLine` truncates `ToolInput` dynamically: `maxInput = m.width - 32` (min 20); tool names still capped at 14 chars. LogText uses `maxText = m.width - 17` (min 20). Both adapt to terminal width instead of fixed 60/80-char limits. |
| Bug | macOS iTerm scroll issue | ✅ Fixed v0.0.55 | Enabled `tea.WithMouseCellMotion()` on both TUI program paths; added `tea.MouseMsg` handler in `update.go` that maps `MouseButtonWheelUp`/`Down` to scroll actions (same bounds/newBelow logic as keyboard) |
| Bug | Windows WezTerm header disappears after multiline output | ✅ Fixed v0.0.53 | `singleLine()` helper in `view.go` strips `\r\n`, `\r`, `\n` from all text before rendering; applied to `e.Message` and `e.ToolInput` in all `renderLine()` cases; prevents TUI height overflow when Claude outputs multi-paragraph reasoning text |

### RK Improvements (Issue #2)
| Priority | Item | Status | Notes |
|----------|------|--------|-------|
| Low | `ralph init` adds `.ralph/regent-state.json` to `.gitignore` | ✅ Fixed v0.0.43 | `ScaffoldProject` creates/appends `.gitignore` with `.ralph/regent-state.json` entry; idempotent |
| Low | Read project name from pyproject.toml/package.json/cargo.toml | ✅ Fixed v0.0.50, fallback v0.0.64 | `DetectProjectName(dir)` in `internal/config/detect.go`; checks pyproject.toml ([project] name or [tool.poetry] name), package.json (name), Cargo.toml ([package] name) in priority order; falls back to `filepath.Base(dir)` when no manifest provides a name; called in `Load()` when project.name is empty; spec at `specs/project-name-detection.md` |
| Low | Webhooks / ntfy.sh notifications | ✅ Fixed v0.0.65 | `[notifications]` section in ralph.toml with `url`, `on_complete`, `on_error`, `on_stop`; `internal/notify` package with `Notifier.Hook()`; `Loop.NotificationHook` func field called from `emit()`; plain-text POST with `X-Title` header (ntfy.sh compatible); fire-and-forget (goroutine); spec at `specs/002-v2-improvements/notifications.md` |
| Low | Allow user to stop after current iteration | ✅ Fixed v0.0.49 | `s` key in TUI closes `Loop.StopAfter` channel; loop exits after current iteration with `LogStopped`; footer shows `⏹ stopping after iteration…  q to force quit`; spec at `specs/graceful-stop.md` |
| Info | Work trees per iteration | Pending | High effort; needs spec; would require major loop refactor |
| Info | Rename PROMPT_plan.md → PLAN.md, PROMPT_build.md → BUILD.md, IMPLEMENTATION_PLAN.md → CHRONICLE.md | ✅ Fixed v0.0.57 | Scaffold creates PLAN.md, BUILD.md, CHRONICLE.md; defaults updated; spec at `specs/rename-prompt-files.md`; project files renamed |
| Low | `ralph init` write IMPLEMENTATION_PLAN.md | ✅ Fixed v0.0.54 | `ScaffoldProject` creates `IMPLEMENTATION_PLAN.md` with starter template (Completed Work, Remaining Work, Key Learnings sections); idempotent; spec at `specs/init-implementation-plan.md` |
| Info | Webhooks / ntfy.sh notifications | ✅ Fixed v0.0.65 | See RK Improvements table |
| Info | Regent daemon mode | Pending | Explicitly out of scope in current specs |

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
- `Config.Validate()` is pure (no I/O) — prompt file existence pre-flighted in `executeLoop` via `os.Stat` before TUI/Regent start; `loop.Run()` still reads the file at runtime
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
- TUI truncates tool names >14 chars with `"…"` to preserve columnar log layout; tool input and LogText truncation adapts to `m.width` (`maxInput = m.width - 32`, `maxText = m.width - 17`, min 20) so lines always fit the terminal without a fixed character ceiling; clamp paths (width < 52 for ToolUse, width < 37 for LogText) tested via `m.width = 30` directly on the model struct before calling `renderLine`
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
- `executeLoop`/`executeSmartRun` integration tests: use `t.Chdir(t.TempDir())` + `writeExecTestFile` helpers; test error paths (no ralph.toml, invalid config, prompt missing, regent-enabled path) without needing a real Claude binary; `signalContext` goroutine exits cleanly via `defer cancel()` even in error paths; `showStatus` corrupted-state-file test covers `regent.LoadState()` parse-error path
- `planCmd`/`buildCmd`/`runCmd` RunE closures tested by calling `cmd.RunE(cmd, nil)` in no-config temp dir; `--no-tui` persistent flag not inherited when calling RunE directly but doesn't matter since config.Load() fails first
- Remaining 0% functions are OS-level (`main`, `registerQuitHandler`) or TUI-required (`runWithRegentTUI`, `finishTUI`, `runWithTUIAndState`) — not worth testing without a real terminal
- `spec.List()` walks one level of subdirectories (e.g. `specs/001-the-genesis/`); `ralph spec new` still creates flat `specs/name.md`; two-levels-deep and hidden files are ignored; `Path` field is relative to project root (`specs/subdir/name.md`)
- `RunTests()` in tester.go uses `runtime.GOOS` to select `cmd /C` (Windows) or `sh -c` (Unix); `errors.As(err, &exitErr)` distinguishes test failure (`*exec.ExitError` → `Passed: false`) from shell-not-found (other errors → return error); `TestRunTests_ShellNotFound` covers the error path via `t.Setenv("PATH", "")`
- SIGQUIT handling: `quit_unix.go` registers `syscall.SIGQUIT` via `signal.Notify`; goroutine prints "SIGQUIT — stopping immediately" to stderr and calls `os.Exit(1)`; `quit_windows.go` is a no-op (SIGQUIT is Unix-only); satisfies the-regent.md "On SIGQUIT: stop immediately, kill Ralph child process"

- `summarizeInput()` extracts display text from tool inputs by checking known field names in priority order: `file_path`, `command`, `path`, `url`, `pattern`, `description`, `prompt`, `query`, `notebook_path`, `task_id`; unknown tool types show no input (empty string is valid)
- TUI `lastDuration` tracks seconds for the last completed iteration via `LogIterComplete` events; rendered as `last: %.1fs` in header; zero-value means no iteration completed yet so the field is omitted; TUI header test must set `Width: 200` via `WindowSizeMsg` to prevent lipgloss wrapping when asserting on header content
- `Stash()` handles "No local changes to save" from `git stash push` as success — some git versions exit non-zero even when nothing is stashed; `stashIfDirty()` already pre-guards via `HasUncommittedChanges()` but defensive handling prevents errors if called directly
- `ScaffoldProject` creates/appends `.gitignore` with `.ralph/regent-state.json` entry; file is created if absent, entry is appended if missing, no-op if already present; `.gitignore` path is included in the `created` list whenever the file was created or the entry was added
- `loop.Run()` calls `LastCommit()` at startup and sets `Commit` on the initial `LogInfo` event; TUI footer updates via `handleLogEntry`'s `if entry.Commit != ""` guard — gracefully ignores empty-repo case
- `regent.SaveState` uses write-then-rename (atomic): writes JSON to a temp file in the `.ralph/` dir, then renames to `regent-state.json`; prevents partial reads when `Supervise` and the drain goroutine in `runWithRegent` call `saveState` concurrently
- TUI clock ticker: `Init()` returns `tea.Batch(waitForEvent, tickCmd())`; `tickCmd()` uses `tea.Tick(time.Second, ...)` to fire `tickMsg` each second; handler in `Update()` updates `m.now` and reschedules with `tickCmd()`; `startedAt` set once in `New()` for elapsed computation; `formatElapsed(d)` renders compact duration (Xs, Xm Ys, Xh Ym)
- TUI `renderLine` truncates `ToolInput` at 60 chars (59 + `…`) to match the tool-name truncation pattern (14 chars); truncation happens at display time in `view.go`, not at source in `loop.go`, keeping `LogEntry.ToolInput` intact for any non-TUI consumers
- `tui.New()` accepts a `projectName` third parameter and a `requestStop func()` fourth parameter; `renderHeader()` shows `👑 <projectName>` when set, falls back to `👑 RalphKing` when empty; both `runWithRegentTUI` and `runWithTUIAndState` pass `cfg.Project.Name` and a `sync.Once`-guarded channel close through
- Graceful stop: wiring creates `stopCh chan struct{}` + `sync.Once`-guarded close; assigns `stopCh` to `Loop.StopAfter` and close func to TUI's `requestStop`; loop checks channel after each iteration via non-blocking `select`; TUI `s` key handler guards on `!m.stopRequested` to make repeat presses no-ops; footer switches to `⏹ stopping after iteration…  q to force quit` when stop is requested
- `DetectProjectName(dir)` tries pyproject.toml → package.json → Cargo.toml → `filepath.Base(dir)` in priority order; pyproject.toml checks `[project] name` (PEP 621) first, then `[tool.poetry] name` (Poetry); all parse errors are silently ignored; called in `Load()` only when `cfg.Project.Name == ""`; BurntSushi/toml used for TOML manifests, encoding/json for package.json; directory fallback ensures TUI always shows a meaningful project name
- `LogText` kind surfaces `claude.EventText` (agent reasoning/commentary between tool calls) in the TUI with 💭 icon and muted gray style; text is truncated at 80 runes (79 + `…`) to preserve single-line layout; empty text events are silently ignored; `formatLogLine` in `cmd/ralph/execute.go` handles it via the generic path
- `tui.New()` `workDir` param (5th, after `projectName`) displays abbreviated working directory as `dir: ~/path` in header; `abbreviatePath()` in `view.go` replaces home prefix with `~` and converts backslashes to forward slashes; omitted from header when empty; both `runWithRegentTUI` and `runWithTUIAndState` pass `dir` through
- `singleLine(s string) string` in `view.go` strips `\r\n`, `\r`, `\n` with space replacement; applied to all text content in `renderLine()` (`e.Message` and `e.ToolInput`); prevents embedded newlines in Claude reasoning text from causing TUI height overflow and header disappearance on Windows WezTerm
- Mouse wheel scroll in TUI: `tea.WithMouseCellMotion()` must be passed to `tea.NewProgram()` to capture wheel events; handle `tea.MouseMsg` in `Update()` with `msg.Button == tea.MouseButtonWheelUp/Down`; without this, iTerm2 and other terminals route wheel events to their own scrollback buffer instead of the application
- `abbreviatePath` home-dir substitution branch covered by calling `os.UserHomeDir()` in the test and constructing a path with `filepath.Join(home, "projects", "myapp")` — skip via `t.Skip` if `UserHomeDir` errors; TUI coverage 99.1% → 99.5%
- `spec.New` MkdirAll error covered by creating a regular file named `specs` in the temp dir (prevents `os.MkdirAll` from creating the directory; `os.MkdirAll` returns ENOTDIR when a path component is a file); cross-platform (Unix and Windows)
- `ScaffoldProject` `.gitignore` read error (non-IsNotExist) covered by creating `.gitignore` as a directory: on Unix `os.ReadFile` opens the dir then `Read` returns EISDIR; on Windows `os.Open` returns ERROR_ACCESS_DENIED; both are non-IsNotExist errors that trigger the `else if err != nil` branch
- `ScaffoldProject` creates `CHRONICLE.md` with a starter template containing `## Completed Work`, `## Remaining Work`, and `## Key Learnings` sections; idempotent — existing files are never overwritten; listed last in `created` slice (after `.gitignore`)
- `internal/loop/runner.Run` is 0% coverage on Windows because all tests use shell scripts (`#!/bin/sh`) which are skipped on Windows; this is the primary cause of loop package dropping from 97.7% to 81.0% on Windows
- `SaveState` rename error tested by creating a directory at the state file path — `os.Rename(tempFile, directory)` fails on all platforms; `regent.saveState` error emit tested by blocking `.ralph` with a regular file
- `RunPostIterationTests` "Failed to start tests" path tested via `t.Setenv("PATH", "")` to prevent shell binary lookup (same pattern as `TestRunTests_ShellNotFound`)
- Scaffold file rename (PROMPT_plan.md→PLAN.md, PROMPT_build.md→BUILD.md, IMPLEMENTATION_PLAN.md→CHRONICLE.md): changing defaults and scaffold is all that's needed for new projects; existing projects with explicit `prompt_file` in ralph.toml are unaffected; `executeSmartRun` and `spec.List()` look for `CHRONICLE.md`; `internal/spec/spec_test.go` references must match the filename `spec.go` reads
- `internal/notify` package: `Notifier.Hook(entry loop.LogEntry)` is a `Loop.NotificationHook`-compatible func; fires `go n.post(message)` on `LogIterComplete` (on_complete), `LogError` (on_error), `LogDone`/`LogStopped` (on_stop); POST is plain text to configured URL with `X-Title` header; all errors silently discarded; wired in `executeLoop` and `executeSmartRun` via `lp.NotificationHook = n.Hook` when `cfg.Notifications.URL != ""`; `Loop.NotificationHook` is always called before the Events channel send so notifications work in both TUI and non-TUI modes
- `initCmd` ScaffoldProject error path covered by `TestInitCmd_ScaffoldError`: pre-create ralph.toml, PLAN.md, BUILD.md, specs/ then create `.gitignore` as a directory — `ScaffoldProject` fails at the .gitignore step and the error propagates through `initCmd.RunE`; same setup as `scaffold_test.go` "gitignore is a directory" test; cross-platform; `initCmd` 77.8%→88.9%, cmd/ralph 71.5%→71.8%
- `runWithRegent` final-state race: `Supervise`'s internal `saveState` (called on success before returning) runs concurrently with the drain goroutine's `UpdateState` calls. If the drain goroutine renames its temp file first and Supervise renames second, the disk has Iteration=0 instead of 5. Fix: call `rgt.FlushState()` in `runWithRegent` after `<-drainDone` — at that point all `UpdateState` calls are complete and the flush saves the authoritative final state.
- `FlushState` tested via `TestFlushState`: calls `UpdateState` then `FlushState`, verifies persisted state matches; 0%→100%. `Supervise` backoff-select branch (`case <-ctx.Done()` during retry wait) tested via event-watcher pattern: run fails with no context cancel → "Ralph exited with error" event triggers cancel from watcher goroutine → `ctx.Done()` fires inside backoff select → 95.6%→100%.
- Residual coverage gaps (all floors, not actionable): `git.Stash` "No local changes to save" return path requires old git that exits non-zero on empty stash (modern git exits 0); `regent.SaveState` marshal error is impossible (State struct always marshals); CreateTemp/Write/Close error paths require OS permission manipulation; `git.Push` fallback `return nil` (line after `-u` push) requires first push failing but second succeeding — no git scenario achieves this; `git.Pull` abort-fails+merge-fails requires fetch-level failure (rebase never starts) which is git-version-dependent
- `regent.emit` panic: Go non-blocking `select { case ch <- v: default: }` still panics if `ch` is closed — the `default` only fires when the channel is full, NOT when it is closed. Fix: `defer func() { _ = recover() }()` inside `emit`. Root cause: `runWithRegent` drain goroutine processes buffered entries after `close(events)`, and `UpdateState` → `saveState` → `emit` runs on the already-closed channel when `SaveState` fails (e.g. concurrent rename contention on Windows).
- `InitFile` write error path covered by passing `filepath.Join(t.TempDir(), "nonexistent")` as dir — `os.WriteFile` fails because the parent directory doesn't exist; this is cross-platform and doesn't require permission manipulation; `InitFile` 85.7%→100%, config 92.3%→93.1%
- `loop.emit` `NotificationHook` path covered by `TestEmitCallsNotificationHook` in `event_test.go`; verifies hook is called for every emitted entry regardless of Events/Log configuration; loop coverage 80.6%→81.3%
- `ScaffoldProject` write error paths (PLAN.md, BUILD.md, CHRONICLE.md, .gitignore create/append) require non-writable parent directory; `os.chmod` approach is Unix-only and not reliable in CI; these paths remain at 0% as residual gaps

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
