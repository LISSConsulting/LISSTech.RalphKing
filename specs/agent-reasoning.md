# Agent Reasoning Display

## Goal

Surface Claude's text output (reasoning, commentary, and explanations) in the
RalphKing TUI and plain-text log. Claude emits `text` content blocks in its
stream-JSON output when it has reasoning or commentary to share between tool
calls. Currently these are parsed by the stream parser but discarded in the
loop's event drain. This feature makes that content visible.

## Background

Claude CLI's `--output-format=stream-json --verbose` output contains messages
of the form:

```json
{
  "type": "assistant",
  "message": {
    "content": [
      { "type": "text", "text": "I'll start by reading the config file to understand the structure." },
      { "type": "tool_use", "name": "Read", "input": { "file_path": "ralph.toml" } }
    ]
  }
}
```

The `text` blocks are Claude's reasoning before and between tool calls. They
reveal *why* the agent is taking each action — valuable for understanding what
the agent is doing and for debugging unexpected behavior.

## Requirements

1. **Emit text events.** In `loop.Run()`, when a `claude.EventText` event is
   received, emit a `LogEntry` with `Kind: LogText` and `Message` set to the
   text content. Empty text is silently ignored.

2. **New log kind.** Add `LogText` to the `LogKind` enumeration in
   `internal/loop/event.go`.

3. **TUI rendering.** In `internal/tui/view.go`, render `LogText` entries with:
   - A 💭 icon to indicate reasoning/commentary
   - A muted (gray) style to visually de-emphasise relative to tool calls and
     results
   - Truncation at 80 display characters (79 runes + `…`) to preserve the
     single-line-per-entry layout

4. **Plain-text rendering.** In `--no-tui` mode, text entries flow through the
   standard `emit()` path and are written verbatim to the log writer with the
   standard `[HH:MM:SS]  message` format. No special handling is needed.

5. **`formatLogLine` compatibility.** `cmd/ralph/execute.go`'s `formatLogLine`
   does not need a special case for `LogText`; the generic path
   (`[%s]  %s`) already formats it correctly.

## Acceptance Criteria

- A `claude.EventText` event produced by the mock agent in `loop_test.go`
  results in a `LogText` entry being emitted on the `Events` channel.
- `tui.renderLine` for a `LogText` entry contains `💭` and the message text.
- Messages longer than 80 characters are truncated to 79 runes + `…` in the
  rendered TUI line.
- All existing tests pass without modification.
- `go vet ./...` passes with zero warnings.
