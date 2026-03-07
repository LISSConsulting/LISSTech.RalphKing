> Go CLI: spec-driven AI coding loop with Regent supervisor.
> **Specs 001-008 in progress. Spec 008 (TUI overhaul) phases 1-12 complete (T001-T077).** ~90% coverage.

## Completed Work

| Spec | What | Branch |
|------|------|--------|
| Foundation | Config, Git, Claude adapter, Loop, Cobra CLI, signal handling | 0.0.1-0.0.3 |
| TUI (003) | Multi-panel bubbletea (header/specs/iterations/main/secondary/footer), lipgloss, scrollable history | spec 003 |
| Regent | Crash/hang detection, retry/backoff, test-gated rollback, state persistence | 0.0.5+ |
| Spec kit (004) | Directory-based spec discovery, speckit commands, spec.Resolve(), `ralph loop` parent | spec 004 |
| Roam (005) | Completion state machine, `--roam`, augmentPrompt(), LogSpecComplete/LogSweepComplete | spec 005 |
| Polish (006) | lineFormatter, `--no-color`, loopSetup refactor, CI coverage, README | spec 006 |
| Worktrees (007) | `internal/worktree/`, `internal/orchestrator/`, `--worktree`, WorktreesPanel, per-agent Regent | spec 007 |
| TUI Overhaul (008) Phase 1-4 | glamour dep, Focus config/loop field, default FocusSpecs, git info on startup, iterations pre-load, --focus flag, interactive speckit | 008-tui-overhaul |
| TUI Overhaul (008) Phase 5 | Per-tab LogView buffers in MainView: outputLog/specLog/iterationLog/summaryLog independent, AppendLine never displaces spec/iteration content | 008-tui-overhaul |
| TUI Overhaul (008) Phase 6-9 | Panel titles with numbers, layout correctness (ANSI-safe truncation, MaxWidth), specs tree with expand/collapse, glamour markdown rendering | 008-tui-overhaul |
| TUI Overhaul (008) Phase 10-12 | CWD-independent spec status, immediate cost accumulation on LogIterComplete, ShowDetail on IterationSelected, TabBar.SetActive | 008-tui-overhaul |
| TUI Overhaul (008) Phase 13 | go vet clean, all tests pass, golangci-lint clean (gofmt fix in app.go), glamour in CLAUDE.md approved deps | 008-tui-overhaul |

## Remaining Work

Spec 008 ALL PHASES COMPLETE — T001-T084 done. T082-T084 require manual TTY verification.

## Key Learnings

### Architecture
- Module: `github.com/LISSConsulting/LISSTech.RalphKing`, Go 1.24
- Deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`, `bubbles`
- `GitOps` interface at consumer for testability; `RunFunc` lets Regent supervise any loop variant
- `LoopController` decouples TUI from wiring; `Loop.PostIteration` wires Regent per agent

### Loop & Events
- `emit()` non-blocking send prevents deadlock; closures must re-evaluate filesystem state
- Completion: `prevSubtype == "success" && !commitsProduced`
- `augmentPrompt()` appends spec context or roam directive

### TUI
- `waitForEvent` Cmd reads `<-chan LogEntry`, reschedules itself
- `loopDoneMsg` → StateIdle (doesn't quit); clock via `tea.Tick(1s)`
- `innerDims()` subtracts 2 for border; `worktreesSplitDims()` for 3-panel sidebar
- `WithOrchestrator()` builder; `nextFocus()`/`prevFocus()` explicit cycle when orch active

### Testing
- Subprocess fakes: `init()` + `_FAKE_CLAUDE=1` / `_FAKE_WT=1` env guard
- TUI tests: `tea.WithoutRenderer()` + `WithInput("q")` + `WithOutput(Discard)`
- Error injection: file-as-directory, `PATH=""`, null-byte URLs
- Coverage ceilings: `cmd/ralph` 68.7% (TTY), `worktree` 98% (platform), `orchestrator` 94.9% (Windows)

### Lint & Go Gotchas
- golangci-lint v2 + `action@v7`; `.gitattributes` `*.go text eol=lf` for Windows
- `select { case ch <- v: default: }` PANICS on closed channel — use `recover()`
- `SaveState` uses write-then-rename (atomic)

## Out of Scope

- OpenAI / Gemini agents
- Regent daemon mode
