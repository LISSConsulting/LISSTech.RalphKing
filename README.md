# 👑 RalphKing

**Spec-driven AI coding loop CLI — with a supervisor that keeps the King honest.**

Ralph runs Claude Code against your specs in a continuous loop: plan → build → commit → push → repeat. The Regent watches Ralph, detects failures, rolls back bad commits, and resurrects him if he crashes.

```
ralph plan          # Run Claude in plan mode against specs/
ralph build         # Run Claude in build mode
ralph run           # Auto: plan if no IMPLEMENTATION_PLAN, then build
ralph status        # Show last run, cost, iteration, branch
ralph init          # Scaffold ralph.toml in current project
ralph spec new      # Create a new spec
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
danger_skip_permissions = true

[plan]
max_iterations = 3

[build]
max_iterations = 0  # 0 = unlimited

[git]
auto_pull_rebase = true
auto_push = true

[regent]
enabled = true
rollback_on_test_failure = true
max_retries = 3
retry_backoff_seconds = 30
```

## Spec Kit Integration

Ralph natively understands `specs/` directories. With `ralph spec new`, Ralph scaffolds a new spec using Spec Kit conventions. `ralph plan` feeds all specs to Claude for gap analysis.

## Supported Agents

- `claude` — Claude Code CLI (default)
- More coming: OpenAI Codex, Gemini, custom

---

*Ralph is King. The Regent keeps him honest.*
