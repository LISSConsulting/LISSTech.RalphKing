# Quickstart: Spec-Bounded Loop with --roam Flag

## Default Behavior (spec-bounded)

```bash
# Ralph stops when the current spec's work is done
ralph build --max 10 --no-tui
# Output:
# [14:30:15]  Starting build loop on branch 005-spec-bounded-roam (max: 10)
# [14:30:16]  ── iteration 1 ──
# ...
# [14:32:45]  Iteration 4 complete — $0.31 — 42.3s — success
# [14:32:46]  ── iteration 5 ──
# ...
# [14:33:10]  Iteration 5 complete — $0.08 — 12.1s — success
# [14:33:10]  Spec complete — 005-spec-bounded-roam (5 iterations)
```

Ralph detects spec completion when:
1. Claude reports success (result subtype `"success"`)
2. The next iteration produces no new commits (confirmation signal)

If Claude finds more work after reporting success (commits appear), the loop continues normally.

## Roam Mode (improvement sweep)

```bash
# Ralph sweeps all specs on a single sweep branch
ralph build --roam --no-tui
# Output:
# [14:30:15]  Creating sweep branch: sweep/2026-03-02
# [14:30:15]  Starting build loop on branch sweep/2026-03-02 (max: unlimited)
# [14:30:16]  ── iteration 1 ──
# ...
# [14:45:30]  Iteration 8 complete — $0.15 — 18.2s — success
# [14:45:31]  ── iteration 9 ──
# ...
# [14:45:50]  Iteration 9 complete — $0.05 — 10.1s — success
# [14:45:50]  Sweep complete (9 iterations, $2.47)
```

Roam creates a `sweep/YYYY-MM-DD` branch and instructs Claude to check ALL specs for improvements. No branch switching during the sweep.

## With TUI

```bash
ralph build              # default mode with TUI — stops at spec boundary
ralph build --roam       # roam mode with TUI — shows sweep progress
```

## Config Default

```toml
# ralph.toml
[build]
prompt_file = "BUILD.md"
max_iterations = 20
roam = false  # set true to always roam by default
```

## Flag Combinations

| Flags | Behavior |
|-------|----------|
| `ralph build` | Build current spec, stop when complete |
| `ralph build --max 5` | Build current spec, stop when complete or after 5 iterations |
| `ralph build --roam` | Create sweep branch, sweep all specs |
| `ralph build --roam --max 10` | Sweep all specs, hard stop at 10 iterations |
| `ralph build --roam --spec X` | ERROR: mutually exclusive |
| `ralph loop run --roam` | Plan phase (if needed), then roam sweep |

## Stopping

| Signal | Effect |
|--------|--------|
| Spec complete (default mode) | Loop exits early with "spec complete" message |
| Sweep complete (roam mode) | Loop exits early with "sweep complete" message |
| `--max` reached | Hard stop regardless of spec/sweep status |
| Ctrl+C (first, --no-tui) | Finish current iteration, then stop |
| Ctrl+C (second) | Kill immediately |
| `s` key (TUI) | Finish current iteration, then stop |
