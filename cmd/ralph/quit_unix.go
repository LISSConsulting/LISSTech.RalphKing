//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// registerQuitHandler registers a SIGQUIT handler that exits immediately
// without waiting for the current iteration to finish.
// On SIGQUIT: stop immediately, kill Ralph child process.
func registerQuitHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	go func() {
		<-sigs
		fmt.Fprintln(os.Stderr, "SIGQUIT — stopping immediately")
		os.Exit(1)
	}()
}
