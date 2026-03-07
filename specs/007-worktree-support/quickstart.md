# Quickstart: Worktree Support

## Prerequisites

1. Install [worktrunk](https://github.com/max-sixty/worktrunk):
   ```sh
   brew install worktrunk    # macOS/Linux
   winget install max-sixty.worktrunk  # Windows (installs as git-wt)
   ```

2. Set up shell integration:
   ```sh
   wt config shell install
   ```

## Single Worktree Build

Run a build loop in an isolated worktree:

```sh
ralph build --worktree
```

This creates a worktree for your active spec's branch, runs the build loop inside it, and streams results to the TUI. Your main working directory stays clean.

## Parallel Agents via Dashboard

1. Enable worktree support in `ralph.toml`:
   ```toml
   [worktree]
   enabled = true
   ```

2. Launch the dashboard:
   ```sh
   ralph
   ```

3. Select a spec in the Specs panel, press `W` to launch it in a worktree. Repeat for other specs.

4. Monitor all agents in the Worktrees panel. Select one to view its log.

## Merge & Cleanup

After a build completes:

```sh
# Merge a completed worktree
ralph worktree merge 007-worktree-support

# Or from the TUI: select the worktree, press M

# Discard without merging
ralph worktree clean 007-worktree-support

# Clean all completed/failed worktrees
ralph worktree clean --all
```

## Auto-Merge (opt-in)

```toml
[worktree]
enabled = true
auto_merge = true
merge_target = "develop"  # optional: defaults to worktree's base branch
```

With auto-merge, completed builds that pass tests are automatically merged and cleaned up.
