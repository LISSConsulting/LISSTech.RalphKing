# Spec: The Regent — Ralph's Supervisor

## Topic of Concern
The Regent is a supervisor process that watches Ralph, detects failures, rolls back bad commits, and resurrects Ralph if he crashes, hangs, or produces regressions.

## Why
Ralph runs autonomously and makes commits. Without a supervisor, a crash loses the iteration, a bad commit stays in history, and a hang blocks forever. The Regent is the safety layer that keeps Ralph honest and the codebase healthy.

## Responsibilities

### 1. Crash detection & resurrection
- Ralph is run as a child process by the Regent (or the Regent monitors Ralph's PID)
- If Ralph exits non-zero: log the failure, wait `retry_backoff_seconds`, restart Ralph
- After `max_retries` consecutive failures: escalate (print to terminal, optionally send webhook notification) and stop

### 2. Hang detection
- Track last output timestamp from Ralph's stdout/stderr
- If no output for `hang_timeout_seconds`: kill Ralph, log timeout, restart
- Reset hang timer on each new line of output

### 3. Test regression detection (optional, config-gated)
- After each successful iteration (Ralph completes + pushes), if `test_command` is set:
  - Run `test_command` in the repo directory
  - If it fails: run `git revert HEAD --no-edit` and `git push` to undo the bad commit
  - Log the revert with reason
  - Resume Ralph on next iteration
- If `rollback_on_test_failure = false`: skip test run entirely

### 4. State tracking
- Write Regent state to `.ralph/regent-state.json`:
  ```json
  {
    "ralph_pid": 12345,
    "iteration": 7,
    "consecutive_errors": 0,
    "last_output_at": "2026-02-23T20:00:00Z",
    "last_commit": "abc1234",
    "total_cost_usd": 1.42
  }
  ```
- `ralph status` reads this file

### 5. Graceful shutdown
- On SIGINT/SIGTERM: finish current iteration if in progress, then stop
- On SIGQUIT: stop immediately, kill Ralph child process

## Modes

### Embedded (default)
The Regent runs inside the same `ralph` binary. `ralph build` → spawns the loop as a goroutine, Regent runs in a separate goroutine monitoring it.

### Daemon (future)
`ralph regent start` — runs as a background daemon, `ralph regent stop`, `ralph regent logs`.

## Configuration (from `ralph.toml`)
```toml
[regent]
enabled = true                    # false = Ralph runs unsupervised (loop.sh behaviour)
rollback_on_test_failure = false  # true = run tests after each commit
test_command = "go test ./..."    # command to run for regression check
max_retries = 3                   # consecutive failures before giving up
retry_backoff_seconds = 30        # wait between retries
hang_timeout_seconds = 300        # kill Ralph if silent for this long
```

## TUI Integration
The Regent surfaces its activity in the Ralph TUI:

```
[14:25:01]  🛡️  Regent: Ralph exited (exit 1) — retrying in 30s (attempt 2/3)
[14:25:31]  🛡️  Regent: Restarting Ralph...
[14:26:44]  🛡️  Regent: Tests passed ✅ — commit abc1234 kept
[14:27:01]  🛡️  Regent: Tests failed ❌ — reverting commit abc1234
```

Regent messages appear inline in the log with a 🛡️ prefix and orange color.

## Package Structure

```
internal/regent/
  regent.go      — Regent struct, Start(), Stop(), supervision loop
  state.go       — state file read/write (.ralph/regent-state.json)
  tester.go      — test runner, revert logic
```

## Acceptance Criteria
- Ralph crash → Regent restarts after backoff
- Ralph hang → Regent kills and restarts after timeout
- Test failure (when enabled) → Regent reverts commit and continues
- After `max_retries` failures → Regent stops and reports
- `ralph status` shows Regent state
- SIGINT stops gracefully after current iteration
- `.ralph/regent-state.json` written and readable
- Regent messages visible in TUI log
