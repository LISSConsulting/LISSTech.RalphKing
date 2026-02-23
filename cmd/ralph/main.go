// Package main is the entry point for the Ralph CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ralph",
	Short: "👑 RalphKing — spec-driven AI coding loop",
	Long:  "Ralph orchestrates Claude Code runs against your specs in a continuous loop.\nThe Regent watches Ralph and keeps the King honest.",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
