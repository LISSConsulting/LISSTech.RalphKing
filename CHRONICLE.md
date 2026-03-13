> Go CLI: spec-driven AI coding loop with Regent supervisor.
> **Specs 001-008 ALL COMPLETE. Tagged v0.1.57. tui/components at 100% coverage.**

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

## Improvement Sweep (2026-03-13, third pass) — v0.1.63+

- **Test coverage**: Added `TestResolvedWorktreeDir` (2 subtests); `WorktreeConfig.ResolvedWorktreeDir` 0%→83.3% (remaining 16.7% is `os.UserHomeDir()` error path, a platform ceiling); `internal/config` 90.0%→93.6%
- **Dead code**: none found
- **Code hygiene**: zero TODO/FIXME/HACK/XXX in non-test Go source
- **Stale references**: README and CLAUDE.md current; all 11 internal packages documented
- **CI health**: all actions current (checkout@v4, setup-go@v5, golangci-lint-action@v7, upload-artifact@v4)
- **Coverage ceilings confirmed**: `ScaffoldProject` 83.3% (write-error paths require OS tricks); `findConfig` 90.9% (`os.Getwd()` failure); `orchestrator.Launch` StateCompleted path (requires real Claude binary); `autoMergeIfNeeded` RunTests error path (PATH clearing doesn't affect cmd.exe on Windows)

## Improvement Sweep (2026-03-13, second pass) — v0.1.62+

- **Test coverage**: Added `TestSwitchGit_{ReuseExisting,CreateNew,ReuseExistingBranch,MkdirAllFails,GitFails}`; covers all `switchGit()` branches (0%→100%); `internal/worktree` 85.0%→99.2%
- **Dead code**: Simplified `exe()` in `detect.go` — removed unreachable `len(candidates) > 0` guard and `return "wt"` fallback; `wtExecutables()` always returns a non-empty slice
- **Code hygiene**: zero TODO/FIXME/HACK/XXX in non-test Go source
- **Coverage ceiling**: `wtExecutables()` at 66.7% on Windows — non-Windows `return []string{"wt", "git-wt"}` branch is a platform ceiling, cannot be covered in Windows CI

## Improvement Sweep (2026-03-13) — v0.1.61+

- **Test reliability**: Fixed `TestLoopController_StartLoop_ForwardGoroutine` flakiness on machines with claude installed; injected `errAgent` stub via new `loopController.agent` field so the test no longer relies on claude being absent from PATH
- **TUI feedback**: W key now shows error message in Main panel when worktree mode is unavailable (`m.orch == nil`) and logs success/failure of each agent launch; previously silent
- **Code hygiene**: zero TODO/FIXME/HACK/XXX in non-test Go source
- **Note**: `TestSpeckitTasksCmd_AllPrereqs_ReachesSpeckit` and `TestExecuteLoop_Roam_*` are slow integration tests (~50s each) that invoke the real claude binary; they pass but are expected to be slow on developer machines — CI should run fine since claude is absent there

## Improvement Sweep (2026-03-11) — v0.1.60+

- **Test coverage**: Added `TestSecondaryPanel_WorktreesTab_ViewAndUpdate` + `TestSecondaryPanel_SetSize_WithWorktrees`; covers TabWorktrees branches in Update/View/SetSize; panels 98.4%→99.1%
- **Test coverage**: Added `TestScaffoldProject/appends_entry_to_gitignore_that_has_no_trailing_newline`; covers no-trailing-newline branch in ScaffoldProject; config 93.3%→94.0%
- **Code hygiene**: zero TODO/FIXME/HACK/XXX in non-test Go source
- **Dead code**: none found

## Improvement Sweep (2026-03-08) — v0.1.59+

- **Docs**: Added `internal/orchestrator/` and `internal/worktree/` to CLAUDE.md and README.md (both missing since spec 007)
- **Test coverage**: Added `TestSecondaryPanel_EnableWorktrees` + `TestSecondaryPanel_SetWorktreeEntries`; panels package 95.6%→~97%
- **Test coverage**: Added `TestTruncateToWidth` (6 subtests); `truncateToWidth` 0%→100%; panels package 97.6%→98.4%
- **Code hygiene**: zero TODO/FIXME/HACK/XXX in non-test Go source
- **Dead code**: none found
- **CI health**: all actions current (same as v0.1.59 sweep)

## Improvement Sweep (2026-03-07, updated 2026-03-09) — v0.1.59 (no new findings)

Full sweep completed with zero actionable findings:
- **Test coverage**: total 90.4% — all remaining gaps confirmed platform ceilings (git.Pull rebase-fails-merge-succeeds hard to engineer; git.Push -u fallback unreachable with local git; git.Stash "No local changes" modern git exits 0; store.NewJSONL OpenFile/Seek errors; regent.SaveState Write/Close errors require disk-full; config.findConfig os.Getwd impossible; regent.RunTests non-Windows sh branch)
- **Code hygiene**: no TODO/FIXME/HACK/XXX in non-test Go source
- **Stale references**: README current — all 8 specs documented, --focus flag present, CI badges accurate
- **Spec consistency**: specs 001-008 fully implemented, T001-T084 complete, all ACs verified
- **CI health**: all actions current (checkout@v4, setup-go@v5, golangci-lint-action@v7 v2.1.6, upload-artifact@v4)
- **Dead code**: no unexported functions with zero callers found

## Improvement Sweep (2026-03-07, updated 2026-03-08) — v0.1.58

- **Loop fixes committed**: isolateProcess (cross-platform subprocess isolation via procattr_unix/windows.go); HasRemoteBranch guard on auto-pull-rebase (skip when no remote tracking branch yet); StashPop silences "No stash entries found" for no-op stash scenarios
- **git coverage**: 88.9% → 93.7% via TestHasRemoteBranch + TestStashPopNoEntries
- **No TODOs/FIXMEs/HAXes** found in non-test Go source
- Tagged v0.1.58

## Improvement Sweep (2026-03-07, updated 2026-03-07) — v0.1.57+

- **Test coverage**: panels 97.9%→98.7% (SelectedSpec nil, moveCursor scroll branches); tui/components 98.5%→100% (TabBar.View width>0 branch); total 89.8%→90.3% (2026-03-07 second sweep: tui 95.3%→98.5% via Cmd-closure direct tests; tui/panels 98.7%→98.9% via worktreeDelegate non-item guard; orchestrator correctness test for TestCommand+pass→merge)
- **Code hygiene**: no TODO/FIXME/HACK/XXX found; `nextFocus`/`prevFocus` unreachable fallbacks changed from silent focus-advance to no-op with comment
- **Stale references**: README updated — added specs 007/008 to project structure, added `--focus` to CLI reference
- **CI health**: all actions current (checkout@v4, setup-go@v5, golangci-lint-action@v7 v2.1.6, upload-artifact@v4)
- **Dead code**: `nextFocus`/`prevFocus` post-cycle fallbacks are unreachable (cycle covers all 5 valid FocusTarget values); `moveCursor` `scrollTop<0` guard is unreachable (cursor clamping proves scrollTop always ≥ 1 when set by the preceding clause)
- **Spec consistency**: spec 008 all ACs implemented; `--focus` flag registered on build/loop-build/loop-run
- Remaining Windows coverage ceilings: store.NewJSONL OpenFile/Seek errors, wtExecutables non-Windows path, regent.SaveState chmod paths, orchestrator autoMergeIfNeeded RunTests error path (PATH clearing doesn't affect cmd.exe on Windows), orchestrator Launch StateCompleted path (requires real Claude binary), tui renderMarkdown glamour error paths (glamour rarely fails), tui tickCmd inner closure (fires after 1s in bubbletea runtime), tui handleEditSpecRequest ExecProcess callback (fires after editor exits), nextFocus/prevFocus post-cycle fallbacks (unreachable dead code)

## Key Learnings

### Architecture
- Module: `github.com/LISSConsulting/RalphSpec`, Go 1.24
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
