// Package main is the entry point for the Ralph CLI.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("👑 RalphKing — spec-driven AI coding loop")
		fmt.Println("Usage: ralph <command>")
		fmt.Println("Commands: plan, build, run, status, init, spec")
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "ralph: command %q not yet implemented\n", os.Args[1])
	os.Exit(1)
}
