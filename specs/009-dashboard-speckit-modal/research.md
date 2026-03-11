# Research: Dashboard SpecKit Modal

**Branch**: `009-dashboard-speckit-modal` | **Date**: 2026-03-11

## R1: Modal Overlay Pattern in Bubbletea

**Decision**: Follow the existing help overlay pattern in `app.go` — boolean flag on Model, key interception in `handleKey()`, render via `lipgloss.Place()` centering.

**Rationale**: The codebase already has a working overlay (help screen) that absorbs all input and renders centered. Reusing this pattern keeps the implementation consistent and avoids introducing a new abstraction layer.

**Alternatives considered**:
- Dedicated modal framework (e.g., `charmbracelet/huh`) — rejected: adds a new dependency; the modal is simple enough (3 items, no form fields) to build inline.
- Panel-level modal (render within SpecsPanel) — rejected: modal needs to overlay all panels and capture global input, which is a root Model concern.

## R2: Subprocess Communication for Interactive Clarify

**Decision**: Use `exec.CommandContext()` with stdin pipe for bidirectional communication. Runner writes prompts to stdout (parsed via `claude.ParseStream()`), reads user answers from stdin pipe.

**Rationale**: The existing Claude adapter (`internal/loop/runner.go`) already spawns subprocesses with stdout parsing. Adding stdin pipe for Clarify follows the same pattern with minimal new code.

**Alternatives considered**:
- PTY-based communication — rejected: cross-platform complexity (Windows pty support); overkill for line-based Q&A.
- Non-interactive mode with post-hoc editing — rejected: user explicitly chose interactive inline Q&A (clarification Q2 answer).

## R3: Question Detection in Clarify Output

**Decision**: Detect question boundaries by pattern-matching Claude's structured output. Look for lines matching `**Question N**` or `Your choice:` or `Format:` patterns that the clarify workflow consistently emits.

**Rationale**: The clarify workflow follows a strict format (defined in the speckit.clarify skill). Pattern matching on known markers is reliable and doesn't require protocol changes.

**Alternatives considered**:
- Structured JSON protocol between runner and clarify — rejected: would require changes to the clarify skill itself; pattern matching on existing output is sufficient.
- Always show input prompt — rejected: would confuse users during non-question output sections.

## R4: Layout Ratio Change Impact

**Decision**: Change sidebar split from 40% Specs / 60% Iterations to 55% / 45%. No other layout ratios change.

**Rationale**: User requested more room for Specs panel. At minimum terminal (80×24), body height = 22. Old split: Specs=8, Iterations=14. New split: Specs=12, Iterations=10. Both remain usable.

**Alternatives considered**:
- Configurable ratio via `ralph.toml` — rejected: over-engineering for a single ratio; can be added later if needed.
- Collapsible Iterations panel — rejected: out of scope; the 55/45 split addresses the immediate need.

## R5: SpecKit Action Subprocess Invocation

**Decision**: Invoke SpecKit actions via `claude -p "/speckit.<action>" --dangerously-skip-permissions` in the spec directory, matching how Ralph's build loop invokes Claude.

**Rationale**: SpecKit skills are Claude Code slash commands. The existing Claude adapter pattern (subprocess + stream parsing) applies directly. The `--dangerously-skip-permissions` flag is needed because the TUI can't interact with Claude's permission prompts.

**Alternatives considered**:
- Direct Go implementation of speckit logic — rejected: speckit workflows are defined as Claude Code skills, not Go code. Invoking via Claude is the correct integration point.
- Shell script wrapper — rejected: unnecessary indirection; direct subprocess call is simpler.
