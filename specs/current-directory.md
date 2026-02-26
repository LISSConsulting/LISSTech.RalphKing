# Spec: Display Current Working Directory in TUI Header

## Topic of Concern

Show the project's working directory in the TUI header so users know which
repository Ralph is operating on without needing to check the terminal title
or run `pwd` separately.

## Why

When running multiple Ralph instances across different projects, or when launching
Ralph from a script, the terminal title may not identify the active repository.
Displaying the working directory in the header provides unambiguous context.

## Behaviour

The working directory is shown in the TUI header as `dir: <abbreviated-path>`,
immediately after the project name, using `~` to abbreviate the user's home
directory prefix:

```
👑 MyProject  │  dir: ~/Projects/my-project  │  mode: build  │  ...
```

- Backslashes (Windows paths) are normalised to forward slashes for display.
- If the path begins with the user's home directory, that prefix is replaced
  with `~`.
- If the working directory is empty (not provided), the `dir:` field is omitted
  from the header entirely.

## Implementation

### `internal/tui/model.go`

Add `workDir string` field to `Model`:

```go
// Working directory displayed in the header
workDir string
```

Update `New()` to accept a fifth parameter `workDir string` (after `projectName`):

```go
func New(events <-chan loop.LogEntry, accentColor, projectName, workDir string, requestStop func()) Model
```

### `internal/tui/view.go`

Add a helper function `abbreviatePath`:

```go
func abbreviatePath(path string) string {
    if path == "" {
        return ""
    }
    if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(path, home) {
        path = "~" + path[len(home):]
    }
    return strings.ReplaceAll(path, "\\", "/")
}
```

In `renderHeader()`, insert the `dir:` entry immediately after the project name:

```go
parts := []string{"👑 " + name}
if m.workDir != "" {
    parts = append(parts, "dir: "+abbreviatePath(m.workDir))
}
parts = append(parts,
    fmt.Sprintf("mode: %s", mode),
    ...
)
```

### `cmd/ralph/wiring.go`

Pass `dir` to `tui.New()` in both TUI-enabled wiring functions:

- `runWithRegentTUI`: `tui.New(tuiEvents, cfg.TUI.AccentColor, cfg.Project.Name, dir, requestStop)`
- `runWithTUIAndState`: `tui.New(tuiEvents, accentColor, projectName, dir, requestStop)`

Both functions already receive `dir` as a parameter.

## Acceptance Criteria

- [ ] When `workDir` is set, header contains `dir: ~/...` (tilde-abbreviated, forward slashes).
- [ ] When `workDir` is empty, `dir:` field is absent from the header.
- [ ] Backslashes in paths are converted to forward slashes in display.
- [ ] All existing tests continue to pass.
- [ ] New tests cover `abbreviatePath` and the `workDir` header rendering.
