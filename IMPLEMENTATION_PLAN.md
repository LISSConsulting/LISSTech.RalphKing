# Implementation Plan — RalphKing

> Go CLI: spec-driven AI coding loop with Regent supervisor.
> Current state: stub CLI only (`cmd/ralph/main.go`). All internal packages empty.

## Completed Work

| Feature | Spec | Tag |
|---------|------|-----|
| Repo scaffold (stub CLI, specs, CLAUDE.md, ralph.toml) | — | — |

## Remaining Work (Prioritized)

### P1 — Foundation (must come first, everything depends on these)

- **Config package** (`internal/config/config.go`)
  - Parse `ralph.toml` using `BurntSushi/toml`
  - Structs: `Config`, `ProjectConfig`, `ClaudeConfig`, `PlanConfig`, `BuildConfig`, `GitConfig`, `RegentConfig`
  - `Load(path string) (*Config, error)` — walk up from CWD if no path given
  - `Defaults()` — sensible defaults matching `ralph.toml` example in spec

- **Git package** (`internal/git/git.go`)
  - `CurrentBranch() (string, error)`
  - `Pull(branch string) error` — rebase first, fall back to merge on conflict
  - `Push(branch string) error` — with `-u origin` fallback for new branches
  - `HasUncommittedChanges() (bool, error)`
  - `Stash() error`, `StashPop() error`
  - `LastCommit() (string, error)` — short SHA + message
  - `Revert(sha string) error`

- **Claude package** (`internal/claude/`)
  - `claude.go` — `Agent` interface + `ClaudeAgent` struct
  - `events.go` — typed event structs: `ToolUseEvent`, `ResultEvent`, `ErrorEvent`, `TextEvent`
  - Stream-JSON parser: reads `--output-format=stream-json --verbose` output line by line
  - `Run(ctx, promptFile, opts) (<-chan Event, error)` — spawns `claude -p` subprocess

- **Cobra CLI skeleton** (`cmd/ralph/main.go` + subcommands)
  - Root command with version
  - Subcommands registered: `plan`, `build`, `run`, `status`, `init`, `spec`
  - Each subcommand can be a stub that prints "not yet implemented" — wired up properly in later steps

### P2 — Core Loop (`internal/loop/`)

- **`loop.go`** — `Loop` struct, `Run(ctx, config, mode) error`
  - Mode: `plan` or `build`
  - Each iteration: stash if dirty → pull → run Claude → push if new commits
  - Respects `max_iterations` (0 = unlimited)
  - Emits iteration events for TUI consumption

- **`runner.go`** — Claude process management
  - Wraps `internal/claude` agent
  - Reads prompt file, feeds to Claude
  - Returns parsed events

### P3 — Commands (wire up loop to CLI)

- **`ralph plan [--max N]`** — run plan loop, feed `plan.prompt_file`
- **`ralph build [--max N]`** — run build loop, feed `build.prompt_file`
- **`ralph run [--max N]`** — smart: plan if no `IMPLEMENTATION_PLAN.md`, then build
- **`ralph init`** — write `ralph.toml` template to CWD
- **`ralph status`** — read `.ralph/regent-state.json`, print summary

### P4 — Spec Commands

- **`ralph spec list`** — list `specs/*.md` with status (✅ complete per IMPLEMENTATION_PLAN, ⬜ not started)
- **`ralph spec new <name>`** — create `specs/<name>.md` from Spec Kit template, open `$EDITOR`

### P5 — TUI (`internal/tui/`)

- `bubbletea` model with header bar, scrollable log, footer bar
- Header: `👑 RalphKing  │  branch  │  iter N/M  │  cost $X.XX`
- Log: timestamped tool events (📖 read, ✏️ write, 🔧 bash, ✅ result, ❌ error)
- Footer: `[⬆ pull] [⬇ push]  last commit: ...  │  q to quit`
- Color-coded: reads=blue, writes=green, bash=yellow, errors=red, regent=orange
- Initial implementation can use simple `fmt.Println` output (loop works first, pretty TUI second)

### P6 — Regent (`internal/regent/`)

- `regent.go` — supervisor goroutine: crash detection, hang detection, restart with backoff
- `state.go` — read/write `.ralph/regent-state.json`
- `tester.go` — run `test_command`, revert on failure
- Wire into `ralph build` / `ralph run` when `regent.enabled = true`

## Key Learnings

- Go module: `github.com/LISSConsulting/LISSTech.RalphKing`
- `go 1.23` — use modern Go idioms
- Approved deps: `cobra`, `BurntSushi/toml`, `bubbletea`, `lipgloss`
- Build target: `go build ./cmd/ralph/`
- Test: `go test ./...`
- Vet: `go vet ./...`
- Cross-compile: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`
- Start tags at `0.0.1`, increment patch per meaningful milestone

## Out of Scope (for now)

- OpenAI / Gemini agent implementations
- Daemon mode (`ralph regent start`)
- Webhook notifications from Regent
- CI release pipeline (exists already in `.github/`)
