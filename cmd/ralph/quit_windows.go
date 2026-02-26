//go:build windows

package main

// registerQuitHandler is a no-op on Windows where SIGQUIT is not available.
func registerQuitHandler() {}
