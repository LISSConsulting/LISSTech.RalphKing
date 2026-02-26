# Spec: Graceful Stop — Stop After Current Iteration

## Topic of Concern
Allow the user to request that the loop stops cleanly after the current iteration
completes, rather than killing the process mid-iteration or waiting indefinitely.

## Why
`ctrl+c` kills the loop immediately, potentially mid-commit or mid-push. Users need
a way to say "finish what you're doing, then stop" so git state is never left dirty.

## Behaviour

### Key binding
In the TUI, pressing `s` requests a graceful stop.

- The first `s` press sets a "stop requested" flag; subsequent presses are no-ops.
- The TUI footer immediately reflects the pending stop: the right section changes
  from `q to quit` to `⏹ stopping after iteration…  q to force quit`.
- The loop continues the current iteration to completion (including git push), then
  exits with a `LogStopped` event and nil error.

### Non-TUI (`--no-tui`) mode
No key binding is available. Graceful stop is not supported in `--no-tui` mode;
users should use `ctrl+c` or SIGQUIT as before.

## Implementation

### `internal/loop/loop.go`
Add `StopAfter <-chan struct{}` field to `Loop`. After each iteration (after the
`PostIteration` hook and the running-total emit), check:

```go
if l.StopAfter != nil {
    select {
    case <-l.StopAfter:
        l.emit(LogEntry{Kind: LogStopped, Message: "Stop requested — exiting after this iteration"})
        return nil
    default:
    }
}
```

### `internal/tui/model.go`
Add fields to `Model`:
- `requestStop func()` — called once when `s` is pressed; provided by wiring.
- `stopRequested bool` — set to true after first `s` press.

Update `New()` to accept a fourth parameter: `requestStop func()`.

### `internal/tui/update.go`
In `handleKey`, add:

```go
case "s":
    if m.requestStop != nil && !m.stopRequested {
        m.stopRequested = true
        m.requestStop()
    }
```

### `internal/tui/view.go`
In `renderFooter`, when `m.stopRequested` is true, change the right section:

```
⏹ stopping after iteration…  q to force quit
```

This replaces (not appends to) the existing scroll/quit hint.

### `cmd/ralph/wiring.go`
In both `runWithRegentTUI` and `runWithTUIAndState`:

1. Create the stop channel: `stopCh := make(chan struct{})`.
2. Create a one-shot close using `sync.Once`.
3. Pass the close func to `tui.New()` as the fourth argument.
4. Assign `stopCh` to `lp.StopAfter`.

## Acceptance Criteria

- [ ] Pressing `s` in TUI footer shows `⏹ stopping after iteration…  q to force quit`.
- [ ] Loop completes current iteration (git push included), then emits `LogStopped` and exits.
- [ ] Second `s` press is a no-op.
- [ ] `--no-tui` mode is unaffected.
- [ ] All existing tests continue to pass.
- [ ] New tests cover: `s` key sets `stopRequested`; `StopAfter` check exits loop after iteration.
