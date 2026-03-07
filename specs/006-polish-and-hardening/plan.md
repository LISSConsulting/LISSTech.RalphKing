# Plan 006 — Polish & Hardening

## Architecture

No new packages. Changes touch `cmd/ralph/` (CLI wiring), `internal/tui/styles.go` (export color logic), `internal/store/`, `internal/regent/`, `internal/spec/`, `internal/config/`, and `.github/workflows/ci.yml`.

---

## Phase 1: ANTHROPIC_API_KEY Warning (F2)

### Design
- Add `PersistentPreRunE` on `rootCmd()` in `main.go`.
- Check `os.Getenv("ANTHROPIC_API_KEY")`; if non-empty, print styled warning to stderr.
- Use lipgloss directly (it auto-detects terminal capability; piped/dumb = no ANSI).
- Respect `--no-color` flag (added in Phase 2, but wire it here as a persistent flag first).

### Files
- `cmd/ralph/main.go` — add `--no-color` persistent flag + `PersistentPreRunE`
- `cmd/ralph/main_test.go` or `cmd/ralph/commands_test.go` — test warning appears/absent

### Tests
- `TestAPIKeyWarning_Set`: set env var, capture stderr, assert warning present
- `TestAPIKeyWarning_Unset`: no env var, assert no warning
- `TestAPIKeyWarning_NoColor`: set env var + `--no-color`, assert warning present but no ANSI escapes

---

## Phase 2: Colorful --no-tui Output (F1)

### Design
- Create `lineFormatter` in `cmd/ralph/format.go` (new file):
  ```go
  type lineFormatter struct {
      color bool
  }
  func (f lineFormatter) format(entry loop.LogEntry) string
  ```
- When `color=true`, apply lipgloss styles matching the TUI palette:
  - Timestamp → gray
  - `LogError` → red bold
  - `LogIterComplete` → green bold
  - `LogToolUse` → blue/green/yellow per tool (import style logic from `tui/styles.go`)
  - `LogRegent` → orange + shield
  - `LogDone` → green
  - `LogStopped` → gray
  - `LogInfo`, `LogText` → white
- Extract `toolCategory(toolName string) string` to a shared location or duplicate the small switch (3 categories: read/write/bash).
- Wire `lineFormatter` into all 3 no-TUI drain goroutines in `wiring.go` (lines 43, 157, and runWithRegent line 43).

### Files
- `cmd/ralph/format.go` (new) — `lineFormatter` + tests
- `cmd/ralph/format_test.go` (new) — table-driven tests for each LogKind with color on/off
- `cmd/ralph/execute.go` — delete old `formatLogLine()`, wire new formatter
- `cmd/ralph/wiring.go` — pass formatter to drain goroutines

### Tests
- Table-driven: each `LogKind` × `{color: true, color: false}` → assert ANSI presence/absence
- Regression: existing `formatLogLine` test cases migrate to new formatter

---

## Phase 3: Extract Common Setup (F4)

### Design
- New `setupLoop(noTUI, roam bool) (*loopSetup, error)` in `cmd/ralph/execute.go`.
- `loopSetup` struct:
  ```go
  type loopSetup struct {
      cfg       *config.Config
      dir       string
      ctx       context.Context
      cancel    context.CancelFunc
      stopCh    <-chan struct{}
      lp        *loop.Loop
      sw        store.Writer
      sr        store.Reader
      closer    func() // deferred store close
      formatter lineFormatter
  }
  ```
- `executeLoop()` and `executeSmartRun()` call `setupLoop()` then diverge on run logic only.
- `executeDashboard()` extracts store init subset (or just calls `setupLoop` and ignores unused fields).

### Files
- `cmd/ralph/execute.go` — refactor

### Tests
- All existing `execute_test.go` tests pass unchanged (behavioral equivalence).
- Add `TestSetupLoop_ConfigError` and `TestSetupLoop_PromptMissing` if not already covered.

---

## Phase 4: Test Coverage Easy Wins (F3)

### 4a: spec/resolve.go — `checkDir()` non-directory
- Create a regular file where a spec dir is expected, assert `checkDir` returns error.
- File: `internal/spec/resolve_test.go`

### 4b: regent/state.go — `SaveState()` write errors
- Use a read-only temp dir so `os.CreateTemp` fails.
- File: `internal/regent/state_test.go`

### 4c: config/config.go — `findConfig()` parent search
- Create `ralph.toml` in parent, run `findConfig` from child dir, assert found.
- File: `internal/config/config_test.go`

### 4d: regent/tester.go — `RunTests()` non-ExitError
- Use a command that produces a non-ExitError (e.g., command not found).
- File: `internal/regent/tester_test.go`

### 4e: store/jsonl.go — `NewJSONL` seek, `Append` write, `EnforceRetention` remove errors
- `NewJSONL`: open file with restricted permissions or use a directory as path.
- `Append`: close the underlying file before calling Append, assert error.
- `EnforceRetention`: create read-only log dir so `os.Remove` fails.
- File: `internal/store/jsonl_test.go`
- Note: some of these may not be triggerable on Windows (permission model differs). Use `t.Skip` with build tag or runtime check.

---

## Phase 5: Documentation (F5)

### README.md updates
- Add flags table or expand usage examples:
  - `--no-tui` — headless/CI mode
  - `--no-color` — disable ANSI colors
  - `--max N` — override max iterations
  - `--roam` — ignore spec boundary, search entire codebase
- Add ANTHROPIC_API_KEY warning note
- Keep changes minimal; don't restructure the whole README

### File
- `README.md`

---

## Phase 6: Dependency Updates (F6)

- `go get -u ./...`
- `go mod tidy`
- `go test ./...` + `go vet ./...`
- If any test breaks, pin the offending dep back.

---

## Phase 7: CI Improvements (F7)

### Coverage artifact
- Modify test step: `go test -race -coverprofile=coverage.out ./...`
- Add `actions/upload-artifact@v4` for `coverage.out`

### Release automation
- New job triggered on `tags: ['v*']`
- Uses existing cross-compile matrix
- Runs `gh release create $TAG` with all binaries attached
- Only runs after test+lint pass

### File
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml` (new, keeps CI clean)

---

## Dependency Graph

```
Phase 1 (API key warning + --no-color flag)
  └─ Phase 2 (colorful output, uses --no-color)
       └─ Phase 3 (extract setup, uses lineFormatter)
Phase 4 (test coverage) — independent, can parallel with 1-3
Phase 5 (docs) — after 1-2 (documents new flags)
Phase 6 (deps) — independent
Phase 7 (CI) — independent
```

## Risk

- **Low:** All changes are additive or internal refactors. No behavioral changes to the loop, Regent, or TUI.
- **Platform:** Some Phase 4 tests may need Windows skips for permission-based error paths.
- **Lipgloss in non-TTY:** Already handles this gracefully — verified by existing TUI tests using `tea.WithoutRenderer()`.
