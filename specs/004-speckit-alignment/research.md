# Research: Spec Kit Alignment

**Feature**: 004-speckit-alignment
**Date**: 2026-03-01

## R1: Claude Code Slash Command Invocation

### Decision
Spawn `claude` in interactive terminal mode for speckit commands — not stream-JSON mode.

### Rationale
Speckit skills are interactive by design. `/speckit.clarify` asks the user questions one at a time. `/speckit.specify` may need clarification input. These require stdin passthrough. The existing `ClaudeAgent.Run()` in `internal/loop/runner.go` uses:
```
claude -p <prompt> --output-format stream-json --verbose
```
with no stdin connection — unsuitable for interactive skills.

Speckit commands are fundamentally different from loop iterations:
- **Loop**: automated, multi-iteration, event-parsed, TUI-displayed, Regent-supervised
- **Speckit**: user-initiated, single execution, interactive, terminal-displayed

### Implementation Pattern
```go
func runClaude(ctx context.Context, skill string, args ...string) error {
    prompt := fmt.Sprintf("/%s %s", skill, strings.Join(args, " "))
    cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--verbose")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### Alternatives Considered
1. **Extend Agent interface** — Over-engineering. Speckit invocation is a CLI concern, not reused across agent implementations.
2. **Stream-JSON mode** — Interactive skills require stdin. Stream-JSON mode treats Claude as a batch processor.
3. **Embed speckit logic in Go** — Skills are maintained as Claude Code skills. Embedding creates a maintenance burden and divergence risk.

---

## R2: Active Spec Resolution Strategy

### Decision
Exact match: current git branch name → `specs/<branch-name>/` directory.

### Rationale
The spec kit convention uses identical names for branches and spec directories (`004-speckit-alignment`). This makes resolution a single `os.Stat()` call after `git branch --show-current`. No configuration needed.

### Resolution Order
1. `--spec <name>` flag → `specs/<name>/` (explicit override)
2. `git branch --show-current` → `specs/<branch>/` (convention)
3. Error with guidance message

### Edge Cases
- Detached HEAD: `git branch --show-current` returns empty → fall through to error
- Branch with no matching spec dir: error unless command is `specify` (which creates the dir)
- Branch `main`/`master`: no match → error

### Alternatives Considered
1. **Fuzzy matching** — Ambiguous when `004-auth` and `004-auth-v2` both exist.
2. **Config field** — Unnecessary indirection; convention is deterministic.
3. **Environment variable** — The `.specify` scripts set `SPECIFY_FEATURE` but Ralph should be self-sufficient.

---

## R3: Status Model Transition

### Decision
Directory-based specs use artifact-presence detection. Flat `.md` files retain CHRONICLE.md heuristic.

### New Status Types
| Status | Condition | Symbol | Meaning |
|--------|-----------|--------|---------|
| `specified` | `spec.md` exists, no `plan.md` | `📋` | Spec written, needs planning |
| `planned` | `plan.md` exists, no `tasks.md` | `📐` | Plan written, needs task breakdown |
| `tasked` | `tasks.md` exists | `✅` | Ready for implementation |

### Detection Logic
```
if dir has tasks.md → StatusTasked
else if dir has plan.md → StatusPlanned
else if dir has spec.md → StatusSpecified
else → StatusNotStarted (warn: empty spec directory)
```

Flat `.md` files: existing CHRONICLE.md heuristic (unchanged).

### Alternatives Considered
1. **Tasks completion tracking** — Parse `tasks.md` for `- [x]` checkboxes to detect "done". Rejected: overly complex for v1; can be added later.
2. **Unified CHRONICLE.md for everything** — Rejected: spec kit artifacts are self-describing; external state file is redundant.
3. **Drop flat file support entirely** — Rejected: backward compatibility for `specs/*.md` costs almost nothing to maintain.
