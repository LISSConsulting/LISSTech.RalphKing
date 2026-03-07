package orchestrator

import (
	"sync"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// startFanIn launches a goroutine that subscribes to an agent's Events channel,
// wraps each LogEntry in a TaggedLogEntry with the branch name, and forwards it
// to the shared MergedEvents channel.
//
// When the agent's Events channel is closed (loop exits), the goroutine
// decrements the WaitGroup and exits cleanly. The caller is responsible for
// closing MergedEvents after all fan-in goroutines have finished.
func startFanIn(branch string, events <-chan loop.LogEntry, merged chan<- TaggedLogEntry, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for entry := range events {
			select {
			case merged <- TaggedLogEntry{Branch: branch, Entry: entry}:
			default:
				// Drop if MergedEvents is full to avoid blocking the agent.
			}
		}
	}()
}
