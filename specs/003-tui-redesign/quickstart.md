# Quickstart: Panel-Based TUI Redesign

**Feature**: 003-tui-redesign | **Date**: 2026-02-28

## Prerequisites

- Go 1.24+
- Existing RalphKing repo with `ralph.toml` configured
- Terminal emulator supporting 256 colors (recommended: 80×24 minimum)

## Build & Run

```sh
# Add bubbles dependency
go get github.com/charmbracelet/bubbles

# Build
go build ./cmd/ralph/

# Run with TUI (default)
ralph build

# Run without TUI (headless)
ralph build --no-tui

# Check status
ralph status
```

## TUI Keyboard Reference

### Global (any panel)

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Cycle panel focus |
| `1` `2` `3` `4` | Jump to specs / iterations / main / secondary |
| `q` / `ctrl+c` | Quit |
| `s` | Stop loop after current iteration |

### Specs Panel (1)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate spec list |
| `enter` | View spec content in main panel |
| `e` | Open spec in $EDITOR |
| `n` | Create new spec |

### Iterations Panel (2)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate iteration list |
| `enter` | View iteration output in main panel |

### Main Panel (3)

| Key | Action |
|-----|--------|
| `[` / `]` | Switch tabs (Output / Spec / Iteration) |
| `f` | Toggle follow mode (auto-scroll) |
| `j` / `k` | Scroll line by line |
| `ctrl+u` / `ctrl+d` | Page scroll |

### Secondary Panel (4)

| Key | Action |
|-----|--------|
| `[` / `]` | Switch tabs (Regent / Git / Tests / Cost) |
| `j` / `k` | Scroll |

## Configuration

Add to `ralph.toml`:

```toml
[tui]
accent_color = "#7D56F4"  # hex color for borders and headers
log_retention = 20         # keep last N session logs (0 = unlimited)
```

## Testing

```sh
# Run all tests
go test ./...

# Run TUI-specific tests
go test ./internal/tui/...

# Run store tests
go test ./internal/store/...

# Check coverage
go test ./internal/tui/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Session Logs

Session logs are stored in `.ralph/logs/` as JSONL files:

```
.ralph/logs/1709123456-12345.jsonl
```

Each line is a JSON-serialized `loop.LogEntry`. The TUI reads from these files for iteration drill-down. Old session logs are automatically pruned on startup based on `log_retention`.
