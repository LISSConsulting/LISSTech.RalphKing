// Package main is the entry point for the Ralph CLI.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	registerQuitHandler()
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ralph",
		Short: "RalphKing — spec-driven AI coding loop",
		Long: "RalphKing — spec-driven AI coding loop\n\n" +
			"Spec kit workflow: specify → plan → clarify → tasks → run\n" +
			"Loop commands: ralph loop plan/build/run\n" +
			"Run without a subcommand to enter dashboard mode (TUI idle state).",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeDashboard()
		},
	}

	root.PersistentFlags().Bool("no-tui", false, "disable TUI, use plain text output")

	root.AddCommand(
		// Spec kit workflow commands
		specifyCmd(),
		speckitPlanCmd(),
		clarifyCmd(),
		speckitTasksCmd(),
		speckitRunCmd(),
		// Autonomous loop (build kept at top-level; plan/run moved under loop)
		buildCmd(),
		loopCmd(),
		// Project management
		statusCmd(),
		initCmd(),
		specCmd(),
	)

	return root
}

// signalContext returns a context that is cancelled on SIGINT or SIGTERM,
// and a cancel function for cleanup. The signal goroutine exits when the
// context is cancelled (either by signal or by the returned cancel func),
// preventing goroutine leaks.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigs)
	}()
	return ctx, cancel
}
