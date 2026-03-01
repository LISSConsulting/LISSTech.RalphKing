
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **All specs (001–004) fully implemented. No remaining work items.** All 12 packages pass; go vet clean; golangci-lint 0 issues; tests green. `internal/tui/components` 100%, `internal/tui/panels` 100%. cmd/ralph 76.2% — confirmed ceiling: `runWithRegentTUI`/`runWithTUIAndState`/`runDashboard` (0%) require real TTY; `os.Getwd()` error paths untriggerable in tests; `TestSpecListCmd_SpecsNotDir` skipped on Windows.

## Completed Work

| Phase | Features | Tags |
|-------|----------|------|
| Foundation & core | Config (TOML, defaults, walk-up discovery, init, validation), Git (branch, pull/push, stash, revert, diff), Claude (Agent interface, events, stream-JSON parser), Loop (iteration cycle, ClaudeAgent subprocess, GitOps, smart run), Cobra CLI (plan/build/run/status/init/spec), signal handling | 0.0.1–0.0.3 |
| TUI | Bubbletea model (header/log/footer), lipgloss styles, `--no-tui`, scrollable history (j/k/pgup/pgdown/g/G), configurable accent color, `↓N new` indicator | 0.0.4, 0.0.11, v0.0.22–v0.0.23 |
| Multi-panel TUI (spec 003, Phases 1–3 + T023/T024/T025/T027/T028) | `charmbracelet/bubbles` dep; `store.EnforceRetention()`; `TUIConfig.LogRetention`; `internal/tui/components` (TabBar, LogView); `internal/tui/panels` (header, footer, specs, iterations, main_view, secondary); `internal/tui/app.go` root Model composing all panels; old single-panel TUI files deleted; `wiring.go` updated for new `tui.New()` signature; `store.Reader` passed through to iteration drill-down; T023: `panels/main_view.go::ShowSpec()` reads spec file and displays in logview (TabSpecContent); T024: `e` key on selected spec emits `EditSpecRequestMsg`; `app.go::handleEditSpecRequest()` resolves `$EDITOR`, calls `tea.ExecProcess()`, reloads spec list via `specsRefreshedMsg` on return; T025: `n` key activates `textinput` overlay in `SpecsPanel`; enter emits `CreateSpecRequestMsg{Name}`; `app.go::handleCreateSpecRequest()` calls `spec.New()` then reloads spec list; T027: `app.go::handleIterationSelected()` async-loads JSONL iteration log via `storeReader.IterationLog(n)`, renders with theme, calls `mainView.ShowIterationLog()`; secondary panel Regent+Git tabs fully functional; T028: `TabIterationSummary` (4th tab), `summaryLogview`, `SetIterationSummary()`, `renderIterationSummary()` — cost/duration/subtype/commit as key-value pairs; `]` from Iteration tab shows Summary | 003-tui-redesign (T001–T028) |
| Regent supervisor | Crash detection + retry/backoff, hang detection (output timeout), state persistence, test-gated rollback (per-iteration), TUI integration, CLI wiring, graceful shutdown | 0.0.5, 0.0.10 |
| Hardening | Stream-JSON `is_error`/`scanner.Err()` handling, `DiffFromRemote` error distinction, config validation, ClaudeAgent stderr capture, TUI error propagation, stale closure fix, result subtype surface, unknown TOML key rejection, rebase abort error surfacing, LastCommit error fallback, signal goroutine leak fix, TUI long tool name truncation, fix gocritic `ifElseChain` in `scaffold.go`/`view.go`, fix `gofmt` alignment in `cmd/ralph/*_test.go`, fix `runWithRegent` final-state race (`FlushState` after drain goroutine), fix `regent.emit` panic on closed channel (`recover()` guard), fix `gofmt` alignment in `loop.go`/`tui/model.go` (struct field spacing, comment alignment in `summarizeInput`) | 0.0.12, 0.0.17–0.0.20, v0.0.27, v0.0.29, v0.0.31, v0.0.32, v0.0.66, v0.0.67, v0.0.68, v0.0.74 |
| State & status | Formatted status display, running-state detection, stateTracker live persistence (non-Regent paths), Regent context-cancel persistence, `detectStatus` fallback | 0.0.8, 0.0.13–0.0.16, v0.0.24–v0.0.25 |
| Cost control | `claude.max_turns` config (0 = unlimited), `--max-turns` CLI passthrough | v0.0.26 |
| Scaffolding | `ralph init` creates ralph.toml + PROMPT_plan.md + PROMPT_build.md + specs/ (idempotent) | v0.0.28 |
| CI/CD | Go 1.24, version injection, race detection, release workflow (cross-compiled binaries on tag push), golangci-lint (go-critic + gofmt) in CI & release; CI lint action pinned to golangci-lint v2.1.6 to fix config-verify failure (`version: latest` resolved to v1.64.8 which rejected v2 config schema); both workflows upgraded to `golangci-lint-action@v7` (v6 rejects golangci-lint v2.x with "not supported" error) | 0.0.7, 0.0.19, v0.0.30, v0.0.87, v0.0.88 |
| Test coverage | Git 93.2%, TUI 97.0% (internal/tui; +9.5pp via comprehensive new tests v0.0.86), loop 98.6% (cross-platform runner.Run tests via self-exec init()), claude 97.8%, regent 95.9%, config 93.1%, spec 96.2%, notify 100.0%, **cmd/ralph 72.5%** (was 71.6% → +0.9pp via `TestFinishTUI_Success` using `tea.WithoutRenderer()` + test I/O — `finishTUI` 0%→50%), **rootCmd 100%** (was 80% — added TestRootCmd_NoSubcommand_CallsDashboard), **components 100%** (was 89.1% — added LogView.Update tests covering KeyMsg/MouseMsg/non-scroll paths), **store 91.0%** (was 90.0% — added onAppend commit-from-complete branch test), **panels 100%** (was 94.7% — added AbbreviatePath home-dir, zero-height constructors/SetSize, Update f/[/default-key/non-key branches, splitLines trailing-newline, iterDelegate.Render selected, SpecsPanel j/k cmd invocation) | 0.0.6, 0.0.14, 0.0.16, v0.0.32, v0.0.36–v0.0.40, v0.0.56, v0.0.58–v0.0.63, v0.0.65, v0.0.68, v0.0.70, v0.0.71, v0.0.72, v0.0.75, v0.0.83, v0.0.84, v0.0.86, v0.0.89, v0.0.90, v0.0.91, v0.0.92 |
| Refactoring | Split `cmd/ralph/main.go` into main/commands/execute/wiring, prompt files, extract `classifyResult`/`needsPlanPhase`/`formatStatus`/`formatLogLine`/`formatSpecList`/`formatScaffoldResult` pure functions with table-driven tests, command tree structure tests, end-to-end command execution tests (cmd/ralph 8.8% → 41.8%); added `runWithStateTracking`/`runWithRegent`/`openEditor` tests (41.8% → 53.4%); added `executeLoop`/`executeSmartRun` integration tests + plan/build/run RunE tests (53.4% → 70.7%); added config-invalid/regent-enabled/corrupted-state-file tests for `executeSmartRun` and `showStatus` (70.7% → 72.0%) | 0.0.9, v0.0.21, v0.0.33–v0.0.38 |

Specs implemented: `ralph-core.md`, `the-regent.md`, all `002-v2-improvements/` specs. Spec `003-tui-redesign/spec.md` **fully complete** (T001–T045, all phases including Phase 8 polish). Spec `004-speckit-alignment/` **fully complete** (T001–T034, all phases).

| Phase | Features | Branch |
|-------|----------|--------|
| Spec kit alignment (spec 004, all phases) | Directory-based spec discovery (`List()` emits one SpecFile per `specs/NNN-name/` dir with artifact-presence status: specified→planned→tasked); new status constants (StatusSpecified 📋, StatusPlanned 📐, StatusTasked ✅); SpecFile.Dir + SpecFile.IsDir fields; `detectDirStatus()` checks tasks.md>plan.md>spec.md; `spec.New()` removed (template.go, spec-template.md deleted); `specNewCmd()` removed from CLI; `formatSpecList()` shows Dir for directory specs; TUI specItem.Description() shows Dir for IsDir specs; TUI `n` key creates directory (not flat file); `ralph loop` parent with plan/build/run subcommands (old top-level plan/run freed for speckit); top-level `build` preserved as alias; speckit commands: `specify`/`plan`/`clarify`/`tasks`/`run` at top level; `executeSpeckit()` spawns `claude -p "/<skill>" --verbose`; `spec.Resolve()` maps branch name or --spec flag to spec directory; `internal/spec/resolve.go` with ActiveSpec struct; PLAN.md + BUILD.md updated for spec kit directory awareness | 004-speckit-alignment |

## Remaining Work

None. All code, tests, CI, and documentation are clean.

### Improvement Sweep (v0.1.04, 2026-03-01)

Full sweep completed — no actionable findings:
- **Test coverage**: All packages confirmed at established floors. `internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 100%, `internal/tui` 99.1%, `internal/loop` 99.3%, `internal/spec` 98.0%, `internal/regent` 95.9%, `internal/config` 93.2%, `internal/git` 93.2%, `internal/store` 91.0%, `cmd/ralph` 76.2% (confirmed ceiling). `go vet ./...` clean.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: None found. README.md, CLAUDE.md all current; all commands, flags, and file names match the implementation.
- **Spec consistency**: Full compliance check against spec 004 (T001–T034), all 12 FRs verified — zero drift found.
- **CI health**: Both `ci.yml` and `release.yml` clean — `golangci-lint-action@v7` with `v2.1.6` pinned. `ci.yml` push triggers for `develop` and `feat/**` remain non-functional (sixth consecutive confirmation); no action required.
- **Dead code**: None found. All unexported helper functions verified in active use.

### Improvement Sweep (v0.1.03, 2026-03-01)

Full sweep completed — one stale-reference finding resolved:
- **Test coverage**: All packages confirmed at established floors. `internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 100%, `internal/tui` 99.1%, `internal/loop` 99.3%, `internal/spec` 98.0%, `internal/regent` 95.9%, `internal/config` 93.2%, `internal/git` 93.2%, `internal/store` 91.0%, `cmd/ralph` 76.2% (confirmed ceiling). `executeDashboard` 22.7% is a TTY ceiling (same as runWithRegentTUI/runDashboard — path through store setup + runDashboard is blocked by TTY requirement). `go vet ./...` clean.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: Fixed CLAUDE.md — removed references to deleted `specs/ralph-core.md` and `specs/the-regent.md` (replaced with current spec kit layout description); removed `specs/ralph-core.md` reference from Config section; added `internal/store/` and `internal/notify/` to Architecture section; added TUI subdirs note.
- **Spec consistency**: Full compliance check against spec 004 (T001–T034) — zero drift found. All 12 FRs implemented and verified.
- **CI health**: Both `ci.yml` and `release.yml` clean — `golangci-lint-action@v7` with `v2.1.6` pinned.
- **Dead code**: None found. All 26 unexported helper functions verified in active use.

### Improvement Sweep (v0.1.02, 2026-03-01)

Full sweep completed — one actionable finding resolved:
- **Test coverage**: Added `TestClaudeAgentRun/empty_executable_defaults_to_claude_binary_name` covering the `if exe == "" { exe = "claude" }` default branch in `ClaudeAgent.Run` (`internal/loop/runner.go:32`). Approach: clear PATH via `t.Setenv("PATH", "")` then create `&ClaudeAgent{}` with empty Executable — `cmd.Start()` fails with "claude agent: start: not found". `internal/loop` package: 98.6%→99.3%; `ClaudeAgent.Run` function: 92.0%→96.0%. Remaining 4% (1 stmt) is `StdoutPipe()` failure — OS-level impossibility, confirmed ceiling.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found.
- **Stale references**: None found. README.md, CLAUDE.md, all CI workflows current.
- **CI health**: Both `ci.yml` and `release.yml` clean — `golangci-lint-action@v7` with `v2.1.6` pinned.
- **Dead code**: None found.
- **Spec compliance**: Full compliance check against spec 003 (T001–T045) and spec 004 (T001–T034) — zero implementation drift found.

### Improvement Sweep (v0.1.00, 2026-03-01)

Full sweep completed — one documentation-only carry-forward finding, no functional changes:
- **Test coverage**: All packages confirmed at or above established floors. Numbers unchanged from v0.0.99: `internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 100%, `internal/tui` 99.1%, `internal/loop` 98.6%, `internal/spec` 98.0%, `internal/regent` 95.9%, `internal/config` 93.2%, `internal/git` 93.2%, `internal/store` 91.0%, `cmd/ralph` 76.2% (confirmed ceiling — `runWithRegentTUI`/`runWithTUIAndState`/`runDashboard` require real TTY; no further improvement without programFactory injection). `go vet ./...` clean.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere. No dead code. All unexported functions verified in active use. 8 legitimately skipped tests (platform/environment conditions).
- **Documentation drift (resolved v0.1.01)**: Marked all 45 task checkboxes `[x]` in `specs/003-tui-redesign/tasks.md` and all 34 task checkboxes `[x]` in `specs/004-speckit-alignment/tasks.md`. Both specs were 100% implemented; runtime was unaffected throughout.
- **Stale references**: None found. README.md, CLAUDE.md, .golangci.yml all current.
- **CI health**: Both `ci.yml` and `release.yml` clean — `golangci-lint-action@v7` with `v2.1.6` pinned. `ci.yml` push triggers for `develop` and `feat/**` remain non-functional (fifth consecutive confirmation); no action required.
- **Dead code**: None found.
- **Spec compliance**: Full compliance check against spec 003 (T001–T045) and spec 004 (T001–T034) — zero implementation drift found.

### Improvement Sweep (v0.0.99, 2026-03-01)

Full sweep completed — one documentation drift finding:
- **Test coverage**: All packages confirmed at or above 90% floor. Current numbers unchanged from v0.0.98: `internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 100%, `internal/tui` 99.1%, `internal/loop` 98.6%, `internal/spec` 98.0%, `internal/regent` 95.9%, `internal/config` 93.2%, `internal/git` 93.2%, `internal/store` 91.0%, `cmd/ralph` 76.2% (confirmed ceiling). `go vet ./...` clean.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere. No dead code. All unexported functions verified in active use. 8 skipped tests, all legitimately conditional (platform/environment reasons).
- **Documentation drift (low priority)**: `specs/003-tui-redesign/tasks.md` has all 45 task checkboxes as `[ ]` despite spec being fully complete (T001–T045); `specs/004-speckit-alignment/tasks.md` has all 34 task checkboxes as `[ ]` despite spec being fully complete (T001–T034). These are documentation-only artifacts — `detectDirStatus()` uses file presence not checkbox state, so runtime behavior is unaffected. Update: mark all tasks `[x]` in both files.
- **Stale references**: None found. README.md, CLAUDE.md all current. No deleted commands/features referenced.
- **CI health**: `ci.yml` push triggers still list `develop` (does not exist) and `feat/**` (naming convention not used) — confirmed non-functional-gap for the fourth consecutive sweep. No action required.
- **Dead code**: None found.

### Improvement Sweep (v0.0.98, 2026-03-01)

Full sweep completed — no actionable findings:
- **Test coverage**: All packages confirmed at or above 90% floor. Current numbers: `internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 100%, `internal/tui` 99.1%, `internal/loop` 98.6%, `internal/spec` 98.0%, `internal/regent` 95.9%, `internal/config` 93.2%, `internal/git` 93.2%, `internal/store` 91.0%, `cmd/ralph` 76.2% (confirmed ceiling — unchanged). Updated header from `~72%` to `76.2%` to reflect actual measured ceiling.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere. `ti.Placeholder` in specs.go and `constitution.md` mentions of "placeholder" are false positives (UI property name and documentation wording). No dead code. All unexported functions verified in active use.
- **Stale references**: No drift found. CLAUDE.md, README.md all current.
- **Spec consistency**: Full compliance check against spec 003 (T001–T045) and spec 004 (T001–T034) acceptance criteria — zero drift found.
- **CI health**: One low-priority informational finding — `ci.yml` push triggers list `develop` (branch does not exist) and `feat/**` (naming convention not used; actual branches are `NNN-feature-name`). Feature branches only receive CI via PR-to-main triggers, which is the correct merge gate. Not a functional gap; no action required unless branch naming conventions change.
- **Dead code**: None found.

### Improvement Sweep (v0.0.97, 2026-03-01)

Full sweep completed — actionable findings resolved:
- **Test coverage**: Added `TestHandleLogEntry_LogIterComplete` covering the `LogIterComplete` case in `handleLogEntry` (app.go:309-319, 3 statements — iterationsPanel.AddIteration + secondary.AddIteration); added `TestParseStream_AssistantNoMessage` covering the `msg.Message == nil` early-return in `parseAssistantMessage` (parser.go:94-96). `internal/claude`: 97.8%→100%; `internal/tui`: 98.2%→99.1%. Remaining gaps: `tickCmd()` (tea.Tick callback unreachable), `tea.ExecProcess` closure body in `handleEditSpecRequest` (requires running editor subprocess; not reliably testable on Windows), all previously confirmed residuals.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: Updated CLAUDE.md — corrected `internal/spec/` description ("templating"→"active spec resolution") and added `charmbracelet/bubbles` to approved deps list.
- **Spec consistency**: No drift found (same as v0.0.96 sweep).
- **CI health**: Clean (same as v0.0.96 sweep).
- **Dead code**: None found.

### Improvement Sweep (v0.0.96, 2026-03-01)

Full sweep completed — actionable findings resolved:
- **Test coverage**: Added `TestRenderLogLine_NarrowWidth_ToolUse` and `TestRenderLogLine_NarrowWidth_LogText` covering the `maxInput < 20` (width=30 gives -2 → clamped to 20) and `maxText < 20` (width=30 gives 13 → clamped to 20) branches in `RenderLogLine`. `internal/tui/theme.go:RenderLogLine`: 88.6%→100%; `internal/tui` package: 97.0%→98.2%. Remaining gaps are confirmed residuals: `tickCmd()` (tea.Tick callback unreachable in tests), `os.Getwd()` errors throughout cmd/ralph (OS-level impossibility), `runWithRegentTUI`/`runWithTUIAndState`/`runDashboard` (0%, TTY required), git Pull/Push/Stash edge cases (old-git or impossible two-step failures), ScaffoldProject write errors (chmod Unix-only/unreliable), SaveState CreateTemp/Write/Close errors (OS permission manipulation), store.NewJSONL OpenFile/Seek errors (OS manipulation), EnforceRetention ReadDir non-IsNotExist error (behaves like IsNotExist on Windows, per TestSpecListCmd_SpecsNotDir skip pattern).
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: No drift found.
- **Spec consistency**: Full compliance check against spec 004 acceptance criteria — zero drift found.
- **CI health**: Both `ci.yml` and `release.yml` are clean — `golangci-lint-action@v7` with `v2.1.6` pinned, all other actions at v4/v5.
- **Dead code**: No orphaned functions or unused constants found.

### Improvement Sweep (v0.0.95, 2026-03-01)

Full sweep completed — actionable findings resolved:
- **Test coverage**: Added 10 tests targeting speckit command coverage gaps. `specifyCmd`: 75%→90%; `speckitPlanCmd`/`clarifyCmd`/`speckitTasksCmd`/`speckitRunCmd`: 68.8%→93.8% each (resolve-error path + all-prereqs-reach-executeSpeckit path). Added `TestCheckDir_PathIsFile` covering `!info.IsDir()` branch in `checkDir`: 75%→87.5%. `internal/spec`: 96.9%→98.0%. `cmd/ralph`: 72.6%→76.2%. Remaining gaps are confirmed residuals: `checkDir()` non-IsNotExist stat error requires platform-specific null-byte path; `specifyCmd`/speckit cmds `os.Getwd()` error + `os.MkdirAll` error paths are OS-level impossibilities in tests.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: No drift found.
- **Spec consistency**: Full compliance check against spec 004 acceptance criteria — zero drift found.
- **CI health**: Both `ci.yml` and `release.yml` are clean — `golangci-lint-action@v7` with `v2.1.6` pinned, all other actions at v4/v5.
- **Dead code**: No orphaned functions or unused constants found.

### Improvement Sweep (v0.0.94, 2026-03-01)

Full sweep completed — no actionable findings:
- **Test coverage**: Added `TestIsNumeric` (empty-string branch in `isNumeric()`) and `TestSpecItem_Description_IsDir` (IsDir=true path in `specItem.Description()`). panels: 99.7%→100%; spec: 95.9%→96.9%. Remaining gaps are confirmed residuals: `List()` non-IsNotExist ReadDir error unreachable on Windows (`TestSpecListCmd_SpecsNotDir` Windows skip pattern), `checkDir()` non-IsNotExist stat error requires platform-specific null-byte path.
- **Code hygiene**: No TODO/FIXME/HACK/XXX found anywhere.
- **Stale references**: Fixed README.md — removed deleted `ralph spec new` command, added speckit workflow commands (`specify`/`plan`/`clarify`/`tasks`/`run`), moved loop commands under `ralph loop` in docs. Updated Spec Kit Integration section.
- **Spec consistency**: Full compliance check against spec 004 acceptance criteria — zero drift found. All 12 FRs implemented and tested.
- **CI health**: Both `ci.yml` and `release.yml` are clean — `golangci-lint-action@v7` with `v2.1.6` pinned, all other actions at v4/v5.
- **Dead code**: No orphaned functions or unused constants found.

## Future Backlog (from GitHub Issues #1 and #2)

All items from Issues #1 (TUI) and #2 (RK) are resolved. Two items remain pending but are out of scope for current specs:

| Priority | Item | Notes |
|----------|------|-------|
| Info | Work trees per iteration (`~/.ralph/worktrees`) | High effort; needs new spec; would require major loop refactor |
| Info | Regent daemon mode | Explicitly out of scope in current specs |

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
- `loop.Run` `CurrentBranch` error path covered by `TestRunCurrentBranchError`; `mockGit` gained `branchErr` field so `CurrentBranch()` can return an error; loop coverage 81.3%→82.0%; `loop.go:Run` 96.7%→100%
- `notify/post` `http.NewRequest` error path covered by `TestPost_InvalidURL`; null byte in URL (`"http://host\x00/path"`) triggers `NewRequest` failure; since `notifier_test.go` is in `package notify`, the `Notifier` struct is directly instantiated with the invalid URL bypassing the validated `New()` constructor; notify package 95.0%→100%
- `signalContext` signal-triggered `cancel()` path (the `case <-sigs:` branch) cannot be covered on Windows (SIGTERM maps to TerminateProcess); `TestSignalContext_SIGTERMCancelsContext` sends SIGTERM to self on Linux/CI — safe because `signal.Notify` suppresses the default termination behavior while the channel is registered; test is skipped on Windows; improves `signalContext` to 100% on Linux CI
- golangci-lint v2 migration: config needs `version: "2"`; `gosimple` merged into `staticcheck`; `gofmt` moves from `linters.enable` to `formatters.enable`; `linters-settings` becomes `linters.settings`. On Windows with `autocrlf=true`, golangci-lint v2 gofmt check flags ALL files (CRLF vs LF). Fix: add `.gitattributes` with `*.go text eol=lf` — this overrides autocrlf for Go files, ensuring LF in both working tree and index, so gofmt sees no formatting difference.
- errcheck in golangci-lint v2 is stricter than v1: `fmt.Fprintln`/`fmt.Fprint`/`fmt.Fprintf` unchecked returns now flagged. Fix with `_, _ = fmt.Fprintln(...)`. For fire-and-forget closures (like defers), use `defer func() { _ = s.Close() }()` instead of `defer s.Close()`.
- gocritic `appendAssign`: `x := append(slice[1:], elem)` is flagged even when intentional (creating new variable). Fix by making a new slice explicitly: `x := make([]T, len(slice)-1, len(slice)); copy(x, slice[1:]); x = append(x, elem)`. Or for `data = append(data, '\n')` where original data is no longer needed, just reuse the variable.
- `LoopController` interface (`internal/tui/controller.go`) decouples TUI from wiring: `b`/`p`/`R` call `StartLoop(mode)`, `x` calls `StopLoop()`. Pass `nil` from existing run paths to disable in-TUI loop control. `runDashboard()` in wiring.go creates a `loopController`, passes it to `tui.New()`, and starts TUI in idle state (channel never closed). `loopDoneMsg` no longer triggers `tea.Quit` — TUI stays open after loop finishes, user presses `q` to exit.
- Cross-platform subprocess tests via self-exec `init()` pattern: place `if os.Getenv("_FAKE_CLAUDE") == "1" { ... os.Exit(code) }` in `init()` (not `TestMain`) so the fake subprocess logic runs BEFORE `flag.Parse()` — otherwise unrecognised Claude CLI flags like `--output-format` cause `flag.ExitOnError` to kill the subprocess before it outputs JSON. Tests use `t.Setenv` to set env vars (auto-restored per subtest), write stdout JSON to a temp file (path passed via `_FAKE_CLAUDE_STDOUT_FILE`), and point `ClaudeAgent.Executable` at `os.Executable()` (the test binary itself). This pattern eliminates shell scripts entirely and works on all platforms; loop coverage 82.0%→98.6% on Windows.
- TUI help overlay: `?` in `GlobalKeyBindings` was a dangling declaration (mentioned in spec.md as a global key but never handled). Fix: `helpVisible bool` field in Model; `handleKey()` checks `helpVisible` first — any key dismisses; `case "?"` sets `helpVisible=true`; `renderHelp()` returns `lipgloss.Place(w, h, Center, Center, box)` where box is `AccentBorderStyle().Padding(1,3).Render(content)`; `View()` returns `renderHelp()` when visible, skipping normal render. Footer updated with `?:help` hint.

- `sw != nil` branch in drain goroutines (`runWithStateTracking`, `runWithRegent`) covered by passing a real `store.NewJSONL` instance constructed in a temp dir; all WithStore tests follow the same pattern: `initGitRepo → NewJSONL → pass to function → send events in run func → close store with defer`
- `loopController.runLoop` goroutine body covered by `TestLoopController_StartLoop_ForwardGoroutine`: `initGitRepo` + create PLAN.md so `loop.Run` reads the prompt and calls `git.CurrentBranch()` before emitting `LogInfo` (which exercises sw.Append + tuiSend channel send); without a real Claude binary, the loop fails at iteration start after emitting the initial event; plan/smart/build modes each have their own test (plan: no git/prompt → fast fail; smart: tests both needsPlanPhase paths via presence/absence of CHRONICLE.md)
- Store-unavailable path in `executeLoop`/`executeSmartRun` covered by creating `.ralph` as a regular file before calling the function; `os.MkdirAll(".ralph/logs")` fails because `.ralph` is not a directory on both Unix and Windows; `store.NewJSONL` returns error → `fmt.Fprintf(os.Stderr, "session log unavailable")` executes → loop continues with `sw=nil`; loop then fails at git `CurrentBranch` (no git repo in temp dir) providing the expected non-nil error
- `golangci-lint-action@v6` with `version: latest` resolves to v1.x (v1.64.8) which does not support the v2 config schema (`version: "2"`, `formatters:`, `linters.settings`). Fix: pin `version: v2.1.6` (or any v2.x.x) in `ci.yml`. Never use `version: latest` when the repo has committed to a v2 config format.
- `golangci-lint-action@v6` explicitly refuses golangci-lint v2.x versions with `"invalid version string 'v2.1.6', golangci-lint v2 is not supported by golangci-lint-action v6, you must update to golangci-lint-action v7."` Fix: use `golangci/golangci-lint-action@v7` in both `ci.yml` and `release.yml`. Keep `version: v2.1.6` pinned.
- `LogView.Update` test pattern: to trigger the `follow=false` branch, populate a small-height viewport (e.g. `height=2`) with ≥20 lines so content exceeds the view, then set `lv.vp.YOffset = 0` (direct field access from within `package components`) to force off-bottom state. Then send `tea.KeyMsg` or `tea.MouseMsg{Button: tea.MouseButtonWheelUp}`. The `switch msg.(type) { case tea.KeyMsg, tea.MouseMsg: }` in `Update` then sets `follow=false`. `store.onAppend` commit-from-complete branch: pass `Commit: "hash"` on the `LogIterComplete` entry — the index copies it to the summary's `Commit` field; covered by `TestIterationSummary_CommitFromComplete`.
- `panels` coverage patterns: (a) `contentH < 1` clamp in `NewMainView`/`NewSecondaryPanel`/`SetSize` triggered by passing `h=0`; (b) `renderCostTable` clamp covered by `NewSecondaryPanel(80, 0)` + navigate to Cost tab; (c) `iterDelegate.Render` selected branch covered by `d.Render(&buf, l, 0, item)` where `0 == l.Index()`; (d) `MainView.Update` `else` branch in default-key handler on non-summary tab covered by sending arbitrary key (`"g"`) to output tab; (e) `SpecsPanel` j/k closure bodies covered by calling `cmd()` after the update returns; (f) `AbbreviatePath` home-dir substitution covered by constructing a path under `os.UserHomeDir()` and calling `AbbreviatePath` on it (skip if UserHomeDir errors).
- `finishTUI` happy-path test: bubbletea v1.3.10 supports `tea.WithoutRenderer()` (sets `nilRenderer` — all methods no-op) which allows `program.Run()` to execute without a real TTY. Combined with `tea.WithInput(strings.NewReader("q"))` (non-TTY reader uses `cancelreader.NewReader` fallback on both Unix and Windows) and `tea.WithOutput(io.Discard)`, the test starts the TUI, receives `loopDoneMsg` (from pre-closed events channel), receives `q` key → `tea.Quit`, and exits cleanly. This covers the happy path of `finishTUI` (50% coverage). The error paths (`program.Run()` returning error, `m.Err()` non-nil) are not triggerable without real terminal failures or editor subprocess failures.
- cmd/ralph coverage ceiling analysis (confirmed 2026-03-01): `runWithRegentTUI`/`runWithTUIAndState`/`runDashboard` are 0% because they create `tea.NewProgram` internally with `tea.WithAltScreen()` + `tea.WithMouseCellMotion()` — test options cannot be injected without refactoring. `os.Getwd()` error branches throughout cmd/ralph are impossible to trigger in tests. `TestSpecListCmd_SpecsNotDir` is skipped on Windows because `os.ReadDir` on a regular file returns `IsNotExist` on Windows (path `specs\*` doesn't exist), not a non-IsNotExist error; verified empirically: `err` message is "The system cannot find the path specified", `os.IsNotExist(err)` is `true`. The 72.5% ceiling is definitive without introducing TTY infrastructure or `programFactory` injection patterns.
- Spec kit directory discovery (spec 004): `List()` now treats each `specs/NNN-name/` directory as a single SpecFile (IsDir=true, Dir=relative path, Path=dir/spec.md); status determined by `detectDirStatus()` checking artifact presence: tasks.md→StatusTasked, plan.md→StatusPlanned, spec.md→StatusSpecified. Flat .md files in specs/ still use the legacy CHRONICLE.md-based `detectStatus()` for backward compatibility.
- `spec.Resolve()` maps a branch name or --spec flag to an active spec directory. Resolution: specFlag→exact match; branch→exact match, then try stripping leading NNN- numeric prefix (so "004-speckit-alignment" also matches "speckit-alignment" dir). main/master and empty-branch cases return descriptive errors suggesting `--spec`.
- `executeSpeckit()` spawns `claude -p "/<skill> <args>" --verbose` with inherited stdio; `execkit_cmds.go` contains `specifyCmd`/`speckitPlanCmd`/`clarifyCmd`/`speckitTasksCmd`/`speckitRunCmd` — each validates prerequisite artifacts before calling claude. TUI `n` key now creates spec directories (not flat files) via `os.MkdirAll`.
- Speckit command tests use `initGitRepoOnBranch(t, dir, branch)` to create a real git repo on a named branch — fake `.git/HEAD` files don't work because `git branch --show-current` requires a proper git repo with at least one commit.

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
