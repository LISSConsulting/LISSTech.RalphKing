package loop

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/claude"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/config"
)

// Mode selects which loop configuration to use.
type Mode string

const (
	ModePlan  Mode = "plan"
	ModeBuild Mode = "build"
)

// GitOps defines the git operations the loop needs.
// *git.Runner satisfies this interface.
type GitOps interface {
	CurrentBranch() (string, error)
	HasUncommittedChanges() (bool, error)
	Pull(branch string) error
	Push(branch string) error
	Stash() error
	StashPop() error
	LastCommit() (string, error)
	DiffFromRemote(branch string) (bool, error)
}

// Loop orchestrates the prompt -> claude -> parse -> git iteration cycle.
type Loop struct {
	Agent  claude.Agent
	Git    GitOps
	Config *config.Config
	Log    io.Writer      // output destination; defaults to os.Stdout
	Events chan<- LogEntry // optional: structured event sink for TUI
	Dir    string         // working directory for prompt file resolution
}

// Run executes the loop in the given mode. It runs iterations until the
// configured max is reached, the context is cancelled, or an error occurs.
// If maxOverride > 0, it overrides the config's max_iterations.
func (l *Loop) Run(ctx context.Context, mode Mode, maxOverride int) error {
	promptFile, maxIter := l.modeConfig(mode)
	if maxOverride > 0 {
		maxIter = maxOverride
	}

	promptPath := filepath.Join(l.Dir, promptFile)
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("loop: read prompt %s: %w", promptFile, err)
	}

	branch, err := l.Git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("loop: get branch: %w", err)
	}

	l.emit(LogEntry{
		Kind:    LogInfo,
		Message: fmt.Sprintf("Starting %s loop on branch %s (max: %s)", mode, branch, iterLabel(maxIter)),
		Branch:  branch,
		MaxIter: maxIter,
		Mode:    string(mode),
	})

	var totalCost float64
	for i := 1; maxIter == 0 || i <= maxIter; i++ {
		select {
		case <-ctx.Done():
			l.emit(LogEntry{
				Kind:    LogStopped,
				Message: fmt.Sprintf("Loop stopped: %v", ctx.Err()),
			})
			return ctx.Err()
		default:
		}

		cost, iterErr := l.iteration(ctx, i, maxIter, string(prompt), branch)
		if iterErr != nil {
			return fmt.Errorf("loop: iteration %d: %w", i, iterErr)
		}
		totalCost += cost
		l.emit(LogEntry{
			Kind:      LogInfo,
			Message:   fmt.Sprintf("Running total: $%.2f", totalCost),
			TotalCost: totalCost,
		})
	}

	l.emit(LogEntry{
		Kind:      LogDone,
		Message:   fmt.Sprintf("Loop complete — %s iterations done, total cost: $%.2f", iterLabel(maxIter), totalCost),
		TotalCost: totalCost,
		MaxIter:   maxIter,
	})
	return nil
}

func (l *Loop) iteration(ctx context.Context, n, maxIter int, prompt, branch string) (float64, error) {
	l.emit(LogEntry{
		Kind:      LogIterStart,
		Message:   fmt.Sprintf("── iteration %d ──", n),
		Iteration: n,
		MaxIter:   maxIter,
		Branch:    branch,
	})

	// Stash uncommitted changes before pulling
	stashed, err := l.stashIfDirty()
	if err != nil {
		return 0, err
	}

	// Pull latest from remote
	if l.Config.Git.AutoPullRebase {
		l.emit(LogEntry{
			Kind:    LogGitPull,
			Message: fmt.Sprintf("Pulling %s", branch),
			Branch:  branch,
		})
		if pullErr := l.Git.Pull(branch); pullErr != nil {
			l.emit(LogEntry{
				Kind:    LogInfo,
				Message: fmt.Sprintf("Pull failed: %v (continuing)", pullErr),
			})
		}
	}

	// Restore stashed changes
	if stashed {
		if popErr := l.Git.StashPop(); popErr != nil {
			l.emit(LogEntry{
				Kind:    LogInfo,
				Message: fmt.Sprintf("Stash pop failed: %v", popErr),
			})
		}
	}

	// Run Claude
	l.emit(LogEntry{
		Kind:    LogInfo,
		Message: "Running Claude...",
	})
	events, err := l.Agent.Run(ctx, prompt, claude.RunOptions{
		Model:                 l.Config.Claude.Model,
		DangerSkipPermissions: l.Config.Claude.DangerSkipPermissions,
	})
	if err != nil {
		return 0, fmt.Errorf("start claude: %w", err)
	}

	// Drain events
	var cost float64
	for ev := range events {
		switch ev.Type {
		case claude.EventToolUse:
			l.emit(LogEntry{
				Kind:      LogToolUse,
				Message:   fmt.Sprintf("tool: %s  %s", ev.ToolName, summarizeInput(ev.ToolInput)),
				ToolName:  ev.ToolName,
				ToolInput: summarizeInput(ev.ToolInput),
			})
		case claude.EventText:
			// Skip text events in log output (verbose)
		case claude.EventResult:
			cost = ev.CostUSD
			l.emit(LogEntry{
				Kind:      LogIterComplete,
				Message:   fmt.Sprintf("Iteration %d complete — $%.2f — %.1fs", n, ev.CostUSD, ev.Duration),
				Iteration: n,
				CostUSD:   ev.CostUSD,
				Duration:  ev.Duration,
			})
		case claude.EventError:
			l.emit(LogEntry{
				Kind:    LogError,
				Message: fmt.Sprintf("Error: %s", ev.Error),
			})
		}
	}

	// Push if there are new local commits
	if l.Config.Git.AutoPush {
		if pushErr := l.pushIfNeeded(branch); pushErr != nil {
			l.emit(LogEntry{
				Kind:    LogError,
				Message: fmt.Sprintf("Push error: %v", pushErr),
			})
		}
	}

	return cost, nil
}

func (l *Loop) stashIfDirty() (bool, error) {
	dirty, err := l.Git.HasUncommittedChanges()
	if err != nil {
		return false, fmt.Errorf("check changes: %w", err)
	}
	if dirty {
		l.emit(LogEntry{
			Kind:    LogInfo,
			Message: "Stashing uncommitted changes",
		})
		if stashErr := l.Git.Stash(); stashErr != nil {
			return false, fmt.Errorf("stash: %w", stashErr)
		}
		return true, nil
	}
	return false, nil
}

func (l *Loop) pushIfNeeded(branch string) error {
	hasChanges, err := l.Git.DiffFromRemote(branch)
	if err != nil {
		l.emit(LogEntry{
			Kind:    LogInfo,
			Message: fmt.Sprintf("Diff check failed: %v (skipping push)", err),
		})
		return nil
	}
	if !hasChanges {
		return nil
	}
	l.emit(LogEntry{
		Kind:    LogGitPush,
		Message: fmt.Sprintf("Pushing %s", branch),
		Branch:  branch,
	})
	if pushErr := l.Git.Push(branch); pushErr != nil {
		return pushErr
	}
	commit, _ := l.Git.LastCommit()
	l.emit(LogEntry{
		Kind:    LogGitPush,
		Message: fmt.Sprintf("Pushed — last commit: %s", commit),
		Commit:  commit,
		Branch:  branch,
	})
	return nil
}

// emit sends a structured log entry. When Events is set, it sends to the
// channel for TUI consumption. Otherwise, it writes formatted text to Log.
// The channel send is non-blocking to prevent deadlock if the TUI exits
// while the loop is still draining events.
func (l *Loop) emit(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if l.Events != nil {
		select {
		case l.Events <- entry:
		default:
		}
		return
	}
	w := l.Log
	if w == nil {
		w = os.Stdout
	}
	ts := entry.Timestamp.Format("15:04:05")
	fmt.Fprintf(w, "[%s]  %s\n", ts, entry.Message)
}

func (l *Loop) modeConfig(mode Mode) (promptFile string, maxIter int) {
	switch mode {
	case ModePlan:
		return l.Config.Plan.PromptFile, l.Config.Plan.MaxIterations
	default:
		return l.Config.Build.PromptFile, l.Config.Build.MaxIterations
	}
}

func iterLabel(max int) string {
	if max == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", max)
}

func summarizeInput(input map[string]any) string {
	for _, key := range []string{"file_path", "command", "path", "url", "pattern"} {
		if v, ok := input[key]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}
