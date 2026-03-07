> Go CLI: spec-driven AI coding loop with Regent supervisor.
> **Specs 001-008 in progress. Spec 008 (TUI overhaul) phases 1-4 complete (T001-T028).** ~90% coverage.

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

## Remaining Work

Spec 008 phases 5-13 — see `specs/008-tui-overhaul/tasks.md` (T029-T084):
- Phase 5 (US5): Per-tab content buffers in MainView — T029-T039
- Phase 6 (US2): Panel titles and numbers — T040-T044
- Phase 7 (US6): Layout correctness audit — T045-T050
- Phase 8 (US3): Specs as traversable tree — T051-T060
- Phase 9 (US11): Markdown rendering with glamour — T061-T064
- Phase 10 (US4): Correct spec status detection — T065-T068
- Phase 11 (US10): Functional cost tracking — T069-T073
- Phase 12 (US9): Footer detail view — T074-T077
- Phase 13: Polish, vet, lint — T078-T084

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
