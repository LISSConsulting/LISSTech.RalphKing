# Spec: Ralph Core — The Loop CLI

## Topic of Concern
`ralph` is a Go CLI that orchestrates Claude Code runs against spec-driven projects in a continuous loop. It replaces `loop.sh` with a proper named binary that has rich TUI output, smart retries, git integration, and Spec Kit awareness.

## Why
`loop.sh` is a bash script with no error handling, no visibility, no retry logic, and no identity. Ralph is a first-class tool that development teams can install, configure, and trust.

## Commands

### `ralph plan [--max N]`
Run Claude in plan mode. Feeds `ralph.toml[plan.prompt_file]` to Claude. Stops after N iterations (default: `ralph.toml[plan.max_iterations]`).

### `ralph build [--max N]`
Run Claude in build mode. Feeds `ralph.toml[build.prompt_file]` to Claude. Runs until max iterations or user interrupt.

### `ralph run [--max N]`
Smart mode: if `IMPLEMENTATION_PLAN.md` doesn't exist or is empty, run plan first (up to `plan.max_iterations`), then build.

### `ralph status`
Show last run summary: branch, last commit, iteration count, total cost, duration, pass/fail.

### `ralph init`
Scaffold `ralph.toml` in the current directory with sensible defaults.

### `ralph spec new <name>`
Create a new spec file at `specs/<name>.md` using the Spec Kit template. Opens in `$EDITOR`.

### `ralph spec list`
List all `specs/*.md` files with a status indicator: ✅ (referenced in IMPLEMENTATION_PLAN as complete), 🔄 (in progress), ⬜ (not started).

## TUI Design

Rich terminal output using `bubbletea` + `lipgloss`:

### Header bar (top)
```
👑 RalphKing  │  branch: feat/teams-shift-assistant  │  iter: 3/10  │  cost: $0.42
```
- Color: primary accent (configurable, default indigo)
- Updates live each iteration

### Iteration panel
Each Claude tool call streams in real time:
```
[14:23:01]  📖  read_file      app/domains/teams/agent.py
[14:23:02]  ✏️   write_file     app/domains/teams/service.py
[14:23:04]  🔧  bash           uv run ruff check --fix app/ (exit 0)
[14:23:07]  ✅  iteration 3 complete  —  $0.14  —  4.2s
```
- Timestamps on every line
- Color-coded by tool type (reads=blue, writes=green, bash=yellow, errors=red)
- Scrollable history

### Footer bar (bottom)
```
[⬆ pull] [⬇ push]  last commit: feat(teams): implement P14  │  q to quit
```

### Cost tracker
Running total displayed in header, per-iteration breakdown in scroll log.

## Git Integration

After each Claude iteration:
1. `git pull --rebase origin <branch>` — pick up concurrent changes
2. If rebase conflict: abort rebase, fall back to `git pull --no-rebase`
3. `git push origin <branch>` (or `git push -u origin <branch>` for new branches)
4. Log result to TUI footer

## Stream-JSON Parsing

Parse Claude's `--output-format=stream-json --verbose` output:
- `type=assistant, content[].type=tool_use` → display tool name + key input
- `type=result` → display cost, duration, exit
- `type=system, subtype=error` → display error in red, trigger retry logic

## Configuration (`ralph.toml`)

```toml
[project]
name = "MyProject"

[claude]
model = "sonnet"
max_turns = 0                    # 0 = unlimited agentic turns per iteration
danger_skip_permissions = true
# future: agent = "claude" | "openai" | "gemini"

[plan]
prompt_file = "PROMPT_plan.md"
max_iterations = 3

[build]
prompt_file = "PROMPT_build.md"
max_iterations = 0

[git]
auto_pull_rebase = true
auto_push = true

[regent]
enabled = true
rollback_on_test_failure = false
test_command = ""
max_retries = 3
retry_backoff_seconds = 30
hang_timeout_seconds = 300
```

## Package Structure

```
cmd/ralph/
  main.go              — cobra root command, subcommand registration

internal/
  config/
    config.go          — ralph.toml parsing (BurntSushi/toml)
  loop/
    loop.go            — core iteration loop: prompt → claude → parse → git
    runner.go          — claude process management (exec, stream)
  claude/
    claude.go          — Claude CLI adapter (stream-json parser, tool event types)
    events.go          — typed event structs (ToolUse, Result, Error)
  git/
    git.go             — pull, push, branch, last commit helpers
  spec/
    spec.go            — spec file discovery, status detection
    template.go        — new spec scaffolding
  tui/
    model.go           — bubbletea model
    view.go            — lipgloss rendering (header, log, footer)
    update.go          — message handling
  regent/
    regent.go          — supervisor (see regent spec)
```

## Agent Abstraction (future-proofing)

Define an `Agent` interface in `internal/claude/`:
```go
type Agent interface {
    Run(ctx context.Context, prompt string, opts RunOptions) (<-chan Event, error)
}
```
Claude implementation is the default. OpenAI/Gemini are future implementations.

## Acceptance Criteria
- `ralph plan --max 3` runs plan loop with rich TUI output
- `ralph build` runs build loop, pulls before push each iteration
- `ralph status` shows last run summary
- `ralph init` creates `ralph.toml` in current directory
- `ralph spec new <name>` creates `specs/<name>.md`
- `ralph spec list` lists specs with status indicators
- TUI shows timestamps, colors, cost per iteration, running total
- Git pull before push, graceful conflict handling
- `go build ./...` passes
- `go test ./...` passes
- Binary cross-compiles for darwin/amd64, darwin/arm64, linux/amd64
