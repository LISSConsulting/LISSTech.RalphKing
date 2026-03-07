> Go CLI: spec-driven AI coding loop with Regent supervisor.
> **Specs 001-007 complete. Spec 008 (TUI overhaul) specified — 84 tasks, 13 phases.** ~90% coverage.

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

## Remaining Work

Spec 008: TUI Overhaul — see `specs/008-tui-overhaul/tasks.md`

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
