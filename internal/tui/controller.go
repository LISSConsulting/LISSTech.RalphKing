package tui

// LoopController allows the TUI to start and stop loop runs without restarting
// the binary. It is passed to New() and used by the b/p/R/x key handlers.
// Pass nil to disable loop-control keys — the existing ralph build/plan/run
// commands manage the loop externally and do not need in-TUI loop control.
type LoopController interface {
	// StartLoop starts a loop in the given mode ("build", "plan", or "smart").
	// A no-op if a loop is already running.
	StartLoop(mode string)

	// StopLoop immediately cancels the running loop. No-op if idle.
	StopLoop()

	// IsRunning reports whether a loop is currently active.
	IsRunning() bool
}
