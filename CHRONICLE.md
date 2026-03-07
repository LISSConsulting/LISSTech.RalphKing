
> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: **Specs 001-007 complete. Spec 008 (TUI overhaul) specified — 84 tasks across 13 phases.** All 14 packages pass; go vet clean; ~90% coverage (Windows).

## Completed Work

| Phase | Features | Tags |
|-------|----------|------|
| Foundation & core | Config (TOML, walk-up discovery, init, validation), Git (branch, pull/push, stash, revert, diff), Claude (Agent interface, stream-JSON parser), Loop (iteration cycle, smart run), Cobra CLI, signal handling | 0.0.1-0.0.3 |
| TUI | Bubbletea multi-panel (header/specs/iterations/main/secondary/footer), lipgloss styles, scrollable history, accent color, `--no-tui` | 0.0.4-v0.0.23, spec 003 |
| Regent supervisor | Crash detection + retry/backoff, hang detection, state persistence, test-gated rollback, TUI integration | 0.0.5, 0.0.10 |
| State & status | Formatted status display, running-state detection, stateTracker, Regent context-cancel persistence | 0.0.8-v0.0.25 |
| CI/CD | Go 1.24, version injection, race detection, release workflow, golangci-lint v2 (action v7), coverage badge | 0.0.7-v0.0.88 |
| Spec kit alignment (spec 004) | Directory-based spec discovery, SpecFile.Dir/IsDir, detectDirStatus(), speckit commands (specify/plan/clarify/tasks/run), spec.Resolve(), `ralph loop` parent | 004-speckit-alignment |
| Spec-bounded roam (spec 005) | LogSpecComplete/LogSweepComplete, completion state machine (prevSubtype+commitsProduced), `--roam` flag, augmentPrompt(), spec.Resolve() wiring | 005-spec-bounded-roam |
| Polish & hardening (spec 006) | lineFormatter, `--no-color`, ANTHROPIC_API_KEY warning, loopSetup refactor, README, deps update, CI coverage | 006-polish-and-hardening |
| Worktree support (spec 007) | `internal/worktree/` (Runner, wt CLI adapter), `internal/orchestrator/` (multi-agent, fan-in, auto-merge), `--worktree` flag, WorktreesPanel TUI, per-agent Regent isolation, W/x/M/D keybinds | 007-worktree-support |

## Remaining Work

### Spec 008: TUI Overhaul & UX Fixes (84 tasks)

Phase 1: Setup (branch, glamour dep, config) — 4 tasks
Phase 2: Quick wins — default focus to Specs, git info on init, load iteration history — 10 tasks
Phase 3: `--focus "topic"` flag for roam mode narrowing — 9 tasks (MVP)
Phase 4: Interactive speckit (clarify/specify without `-p` flag) — 5 tasks
Phase 5: Per-tab content buffers (fix output displacing spec/iteration view) — 11 tasks
Phase 6: Panel titles `[N] Title` with accent color — 5 tasks
Phase 7: Layout correctness (fix overflow/misalignment) — 6 tasks
Phase 8: Spec tree view (expandable dirs with child files) — 10 tasks
Phase 9: Glamour markdown rendering for spec viewer — 4 tasks
Phase 10: Spec status detection fix for external projects — 4 tasks
Phase 11: Cost tracking wiring (header total + Cost tab) — 5 tasks
Phase 12: Footer detail view (lazygit-style secondary panel) — 4 tasks
Phase 13: Polish — 7 tasks

See `specs/008-tui-overhaul/tasks.md` for full breakdown.

### Coverage Floors (as of v0.1.56)

`internal/claude` 100%, `internal/notify` 100%, `internal/tui/components` 100%, `internal/tui/panels` 99.8%, `internal/tui` 97.9%, `internal/loop` 99.4%, `internal/spec` 99.0%, `internal/orchestrator` 94.9%, `internal/worktree` 98.0%, `internal/regent` 95.9%, `internal/config` 93.3%, `internal/git` 93.2%, `internal/store` 92.0%, `cmd/ralph` 68.7% (TTY ceiling). Total: ~90.0% (Windows).

### Sweep History

Sweeps with findings: v0.1.56 (gofmt/CI/docs), v0.1.55 (orchestrator +2 tests), v0.1.54 (focus/footer coverage), v0.1.53 (worktree +4 tests), v0.1.52 (+9 tests across 3 pkgs), v0.1.51 (+7 tests), v0.1.45 (.gitignore), v0.1.44 (specifyCmd test), v0.1.41 (store +2), v0.1.40 (spec +3), v0.1.37 (spec 005 drift doc), v0.1.30 (lint), v0.1.29 (task checkboxes), v0.1.23 (TUI bug fix), v0.1.21-v0.1.26 (task checkboxes, README).

Sweeps with no findings: v0.1.48-v0.1.46, v0.1.43-v0.1.42, v0.1.39-v0.1.31, v0.1.27-v0.1.24, v0.1.22-v0.1.20, v0.1.19-v0.1.01, v0.0.98.

## Key Learnings

### Architecture
- Go module: `github.com/LISSConsulting/LISSTech.RalphKing`, Go 1.24
- Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`, `bubbles`
- `GitOps` interface at consumer (loop pkg) for testability; `*git.Runner` satisfies implicitly
- `RunFunc` abstraction (`func(ctx) error`) lets Regent supervise any loop variant
- `LoopController` interface decouples TUI from wiring; pass nil to disable in-TUI loop control
- `Loop.PostIteration` hook wires Regent's `RunPostIterationTests` per agent

### Loop & Events
- `loop.emit()` non-blocking send prevents deadlock when TUI exits early
- Closures passed to Regent must re-evaluate filesystem state inside the closure body
- `augmentPrompt()` appends spec context or roam directive; returns prompt unchanged when neither applies
- Completion detection: `prevSubtype == "success" && !commitsProduced` → spec/sweep complete
- `summarizeInput()` extracts display text from tool inputs via known field name priority list

### TUI Patterns
- Channel pattern: `waitForEvent` Cmd reads from `<-chan LogEntry`; reschedules itself in handler
- `loopDoneMsg` transitions to StateIdle but doesn't quit — user presses q to exit
- Clock ticker: `tea.Tick(time.Second, ...)` → `tickMsg` → update `m.now` → reschedule
- `singleLine()` strips newlines to prevent height overflow from Claude reasoning text
- `innerDims()` subtracts 2 (border) from layout rect; `worktreesSplitDims()` for 3-panel sidebar
- Focus cycling: `nextFocus()`/`prevFocus()` use explicit cycle slice when `m.orch != nil`
- `WithOrchestrator()` builder — all existing `New()` callers unchanged (8 args)

### Testing Patterns
- Cross-platform subprocess tests: `init()` with `_FAKE_CLAUDE=1` / `_FAKE_WT=1` env guard
- `tea.WithoutRenderer()` + `tea.WithInput(strings.NewReader("q"))` + `tea.WithOutput(io.Discard)` for TUI tests
- `store.Writer` nil check: `if sw != nil { _ = sw.Append(entry) }` in forwarding goroutines
- `t.Chdir(t.TempDir())` + `writeExecTestFile` for end-to-end command tests
- `initGitRepoOnBranch(t, dir, branch)` for real git repos in spec tests
- Error injection: file-as-directory trick, `t.Setenv("PATH","")`, null-byte URLs

### Coverage Ceilings
- `cmd/ralph` 68.7%: `runWithRegentTUI`/`runWithTUIAndState`/`runDashboard` require real TTY
- `internal/worktree` 98.0%: `exe()` dead return, `wtExecutables` non-Windows branch
- `internal/orchestrator` 94.9%: `autoMergeIfNeeded` RunTests on Windows (cmd.exe bypasses PATH)

### Lint & CI
- golangci-lint v2 config: `version: "2"`, `gofmt` in `formatters.enable`
- Must use `golangci-lint-action@v7` for golangci-lint v2.x
- `.gitattributes` with `*.go text eol=lf` prevents CRLF/LF gofmt failures on Windows
- `errcheck`: `fmt.Fprint*` needs `_, _ =`; deferred Close needs `defer func() { _ = ... }()`

### Go Gotchas
- `select { case ch <- v: default: }` PANICS if `ch` is closed — use `recover()` guard
- `appendAssign`: `x := append(slice[1:], elem)` flagged by gocritic; use `make`+`copy`+`append`
- `regent.SaveState` uses write-then-rename (atomic) to prevent partial reads

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
