# Research: Spec 008 — TUI Overhaul

## R-001: Glamour Markdown Rendering

**Decision**: Use `github.com/charmbracelet/glamour` v0.8+
**Rationale**: Same Charm ecosystem as bubbletea/lipgloss/bubbles. Purpose-built for terminal markdown.
**Alternatives**: Raw text (current, poor UX), go-term-markdown (less maintained), goldmark+custom renderer (too much work).

Key API:
```go
r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))
out, _ := r.Render(markdown)
```

Glamour output is ANSI-styled text compatible with lipgloss. Can be set as viewport content directly.

**Risk**: Glamour pulls in goldmark + chroma (syntax highlighting). Binary size increase ~2-3MB. Acceptable for a CLI tool.

## R-002: Interactive Speckit Commands

**Decision**: `clarify` and `specify` use `claude --verbose` (no `-p` flag) with inherited stdio.
**Rationale**: `-p` flag is "print mode" — Claude processes prompt once and exits. Interactive skills need conversation.

Current code (`execute.go:411`):
```go
cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--verbose")
```

Fix: Check skill name; for interactive skills, pass the prompt differently:
```go
cmd := exec.CommandContext(ctx, "claude", "--verbose")
// Feed the skill invocation via stdin or as initial message
```

Actually, the cleanest approach: `claude` without `-p` enters interactive mode. We can pass the skill as the first message. But this changes the UX — Claude will show its interactive UI.

**Revised decision**: For `clarify` and `specify`, spawn `claude` in fully interactive mode (no `-p`). The user interacts directly with Claude in their terminal. Ralph resumes after Claude exits.

```go
func executeSpeckit(ctx context.Context, skill string, args []string, interactive bool) error {
    if interactive {
        cmd := exec.CommandContext(ctx, "claude", "--verbose")
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        return cmd.Run()
    }
    // existing -p path for plan/tasks/etc.
    prompt := "/" + skill + " " + strings.Join(args, " ")
    cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--verbose")
    ...
}
```

Wait — the user also needs the skill context. The `-p` approach does pass the skill invocation. The issue is that clarify asks questions and waits for answers, but `-p` mode exits after one turn.

**Final decision**: Use `claude` in interactive mode and pass the initial prompt via `--initial-prompt` or just let the user type `/clarify` themselves. Simplest: just launch `claude --verbose` and let the user invoke the skill manually. Or use `claude -p "/<skill>" --continue` if that flag exists.

Actually the simplest fix: use `claude` without `-p` but with `--prompt` (or `-m`) which sends an initial message but keeps the session open. Checking Claude CLI docs... The flag is `--message` or just positional. Since we can't verify this at build time, the safest approach is:

**Final final decision**: For interactive skills, use `exec.Command("claude")` with inherited stdio and no `-p`. The user types `/clarify` or `/specify` themselves in the Claude session. Ralph's job is just to spawn Claude in the right directory. This is the simplest, most reliable approach.

## R-003: Per-Tab Content Buffers

**Decision**: Replace single `logview` with four LogView instances.
**Rationale**: Current shared logview means AppendLine() (from streaming output) overwrites spec/iteration content.

Migration:
- `logview` → `outputLog` (always receives AppendLine)
- New `specLog` (receives ShowSpec content)
- New `iterationLog` (receives ShowIterationLog content)
- `summaryLogview` stays as `summaryLog`

View() switches which LogView to render based on activeTab.

## R-004: Panel Titles

**Decision**: Render title as first line inside panel content area.
**Rationale**: lipgloss v1.1.0 has no built-in border title API. Rendering inside is simpler and cross-terminal compatible.

Implementation: Each panel's View() prepends a title line. The title is styled with accent color when the panel has focus. Focus state must be passed to each panel.

Alternative considered: Using lipgloss `Border` customization to embed text in the top border line. This is fragile across different border styles and terminal emulators.

**Revised approach**: Pass `focused bool` and `title string` to each panel's render, or have the parent (app.go View()) prepend the title before wrapping in the border style. The parent approach is cleaner — panels don't need to know about titles.

## R-005: Spec Tree View

**Decision**: Custom tree rendering using a flat list with indentation.
**Rationale**: bubbles/list doesn't support tree hierarchy natively. A custom approach using the existing delegate pattern with indentation gives full control.

The tree is stored as `[]specTreeNode` where each node has an `expanded` bool. The flattened view for rendering is computed on each Update: iterate nodes, for expanded nodes insert child items.

Navigation: cursor tracks position in the flattened list. Enter on a directory node toggles expand. Enter on a child opens the file. j/k moves cursor.

Child files are discovered at init time by checking which of spec.md/plan.md/tasks.md exist in each spec directory.

## R-006: Layout Audit

**Decision**: Fix Calculate() to produce panel rects that account for borders in total dimensions.

Current issue: Calculate() distributes `width` and `height` as panel content rects, but View() wraps each panel in a border (adding 2 to each dimension). The total rendered output is wider/taller than the terminal.

Fix: Calculate() should compute rects where Width/Height include borders. innerDims() already subtracts 2 for content, which is correct. But the sum of all panel rects must equal terminal size, not terminal size + borders.

Specifically:
- `sidebarW + rightW` must equal `width` (including all borders)
- `specsH + itersH + 2` (header+footer) must equal `height`
- `mainH + secH + 2` must equal `height`

Current code: `bodyH = height - 2` is correct. `sidebarW + rightW = width` is correct. But each panel adds a 2-char border, so `specsH + itersH` must equal `bodyH` where each height INCLUDES the border. This seems correct already — the border style sets Width/Height which includes borders.

Actually re-reading the View() code: `PanelBorderStyle(focused).Width(specsW).Height(specsH).Render(content)` — lipgloss Width/Height set the CONTENT width, and borders are ADDED on top. So the total rendered width is `specsW + 2` (left+right border). This means `sidebarW` from Calculate is the content width, but the total sidebar output is `sidebarW + 2`. Similarly rightCol is `rightW + 2`. Total = `sidebarW + 2 + rightW + 2 = width + 4`. That's 4 characters too wide.

Wait, looking more carefully: `innerDims()` returns `r.Width - 2, r.Height - 2`. Then View() uses `Width(specsW).Height(specsH)` where these are inner dims. lipgloss adds the border on top, making the outer size `specsW + 2 = r.Width`. So the outer size of each panel equals the original rect dimensions. This is correct!

So `sidebarW + rightW = width` and each panel's outer width is its rect width. The sidebar panels have outer width `layout.Specs.Width = sidebarW`. The right panels have outer width `layout.Main.Width = rightW`. Total = `sidebarW + rightW = width`. Correct!

Same for height: `specsH_outer + itersH_outer = layout.Specs.Height + layout.Iterations.Height = specsH + itersH = bodyH`. And `headerH + bodyH + footerH = 1 + (height-2) + 1 = height`. Correct!

**Revised assessment**: The layout math is structurally correct. The misalignment the user sees may be caused by:
1. Unicode characters in the header (crown emoji, pipes) that have ambiguous width
2. Content within panels exceeding panel width (no truncation)
3. Tab bar line not constrained to panel width
4. lipgloss width calculation disagreeing with terminal emulator on emoji widths

The fix is likely: ensure all content respects panel width constraints. Audit each panel's View() to ensure output is width-constrained.

## R-007: Focus Flag

**Decision**: Add `--focus` string flag. Append to prompt in augmentPrompt().
**Rationale**: Simple, low-risk way to constrain roam mode.

`augmentPrompt()` gains a `focus` parameter. When non-empty, appends: `"\nFocus your work on: <focus>. Prioritize changes related to this area over other improvements."` This is appended after the roam/spec context, so it works with both modes.
