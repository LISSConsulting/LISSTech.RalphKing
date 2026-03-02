# 👑 RalphKing

**Spec-driven AI coding loop CLI — with a supervisor that keeps the King honest.**

Ralph runs Claude Code against your specs in a continuous loop: plan → build → commit → push → repeat. The Regent watches Ralph, detects failures, rolls back bad commits, and resurrects him if he crashes.

```
ralph               # Dashboard mode — interactive TUI with loop control

# Spec kit workflow (drives Claude through spec kit skills)
ralph specify       # Create a new spec from a description
ralph plan          # Generate implementation plan for active spec
ralph clarify       # Resolve ambiguities in active spec
ralph tasks         # Break plan into actionable task list
ralph run           # Execute spec kit run against active spec

# Autonomous loop (continuous Claude iterations)
ralph build         # Run Claude in build mode (alias for ralph loop build)
ralph build --roam  # Cross-spec improvement sweep on a sweep/YYYY-MM-DD branch
ralph loop plan     # Run Claude in plan mode
ralph loop build    # Run Claude in build mode
ralph loop run      # Smart mode: plan if needed, then build

# Project management
ralph status        # Show last run, cost, iteration, branch
ralph init          # Scaffold ralph project (config, prompts, specs dir)
ralph spec list     # List specs and their status
```

## Architecture

```
┌─────────────┐     watches      ┌─────────────┐     runs      ┌─────────────┐
│   Regent    │ ───────────────> │    Ralph    │ ────────────> │   Claude    │
│ (supervisor)│ <─── reports ─── │   (loop)    │ <── output ── │  (worker)   │
└─────────────┘                  └─────────────┘               └─────────────┘
       │
       ├── detects crash → restarts Ralph
       ├── detects test regression → rolls back commit
       └── detects hang → kills and retries with backoff
```

## Installation

```bash
go install github.com/LISSConsulting/LISSTech.RalphKing/cmd/ralph@latest
```

Or download a pre-built binary from [Releases](https://github.com/LISSConsulting/LISSTech.RalphKing/releases).

## Configuration

Place `ralph.toml` in your project root:

```toml
[project]
name = "MyProject"

[claude]
model = "sonnet"
max_turns = 0  # 0 = unlimited agentic turns per iteration
danger_skip_permissions = true

[plan]
prompt_file = "PLAN.md"
max_iterations = 3

[build]
prompt_file = "BUILD.md"
max_iterations = 0  # 0 = unlimited
roam = false        # enable cross-spec improvement sweep (--roam flag overrides)

[git]
auto_pull_rebase = true
auto_push = true

[regent]
enabled = true
rollback_on_test_failure = false  # set true once test command configured
test_command = "go test ./..."    # command run by Regent after each iteration
max_retries = 3
retry_backoff_seconds = 30
hang_timeout_seconds = 300        # kill Ralph if no output for 5 min

[tui]
accent_color = "#7D56F4"  # hex color for header/accent elements
log_retention = 20        # number of session logs to keep; 0 = unlimited
```

## TUI Keyboard Reference

The four-panel TUI is available for all loop commands and via `ralph` (dashboard mode).

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Cycle panel focus |
| `1` `2` `3` `4` | Jump to Specs / Iterations / Main / Secondary panel |
| `b` | Start build loop |
| `p` | Start plan loop |
| `R` | Smart run (plan if needed, then build) |
| `x` | Cancel running loop immediately (dashboard mode) |
| `s` | Graceful stop after current iteration |
| `?` | Toggle help overlay |
| `q` / `ctrl+c` | Quit |

**Specs panel (`1`):** `j`/`k` navigate · `enter` view · `e` open in `$EDITOR` · `n` create new

**Iterations panel (`2`):** `j`/`k` navigate · `enter` view log (loads in Main panel; use `]` there for summary)

**Main panel (`3`):** `[`/`]` switch tabs · `f` toggle follow · `j`/`k` scroll · `ctrl+u`/`ctrl+d` page

**Secondary panel (`4`):** `[`/`]` switch tabs (Regent / Git / Tests / Cost) · `j`/`k` scroll

Minimum terminal size: 80×24. Set accent color via `[tui] accent_color = "#7D56F4"` in `ralph.toml`.

## Spec Kit Integration

Ralph natively understands `specs/NNN-name/` directories. The spec kit workflow drives Claude through sequential phases: `ralph specify` creates a new spec, `ralph plan` generates a technical plan, `ralph clarify` resolves ambiguities, `ralph tasks` breaks the plan into actionable tasks, and `ralph run` executes the implementation. Use `ralph loop build` for continuous autonomous iterations once the spec is ready.

## Supported Agents

- `claude` — Claude Code CLI (default)
- More coming: OpenAI Codex, Gemini, custom

---

*Ralph is King. The Regent keeps him honest.*
