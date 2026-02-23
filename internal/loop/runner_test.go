package loop

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/claude"
)

func TestNewClaudeAgent(t *testing.T) {
	agent := NewClaudeAgent()
	if agent.Executable != "claude" {
		t.Errorf("expected executable %q, got %q", "claude", agent.Executable)
	}
}

func TestBuildArgs(t *testing.T) {
	agent := &ClaudeAgent{}

	tests := []struct {
		name     string
		prompt   string
		opts     claude.RunOptions
		contains []string
		excludes []string
	}{
		{
			name:   "basic prompt",
			prompt: "test prompt",
			opts:   claude.RunOptions{},
			contains: []string{
				"-p", "test prompt",
				"--output-format", "stream-json",
				"--verbose",
			},
			excludes: []string{"--model", "--dangerously-skip-permissions"},
		},
		{
			name:   "with model",
			prompt: "test",
			opts:   claude.RunOptions{Model: "opus"},
			contains: []string{
				"--model", "opus",
			},
		},
		{
			name:   "with danger skip permissions",
			prompt: "test",
			opts:   claude.RunOptions{DangerSkipPermissions: true},
			contains: []string{
				"--dangerously-skip-permissions",
			},
		},
		{
			name:   "all options",
			prompt: "full test",
			opts:   claude.RunOptions{Model: "sonnet", DangerSkipPermissions: true},
			contains: []string{
				"-p", "full test",
				"--output-format", "stream-json",
				"--verbose",
				"--model", "sonnet",
				"--dangerously-skip-permissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := agent.buildArgs(tt.prompt, tt.opts)

			for _, want := range tt.contains {
				if !containsArg(args, want) {
					t.Errorf("args %v missing expected %q", args, want)
				}
			}
			for _, unwanted := range tt.excludes {
				if containsArg(args, unwanted) {
					t.Errorf("args %v should not contain %q", args, unwanted)
				}
			}
		})
	}
}

// TestClaudeAgentRun tests the full subprocess lifecycle using a fake script.
func TestClaudeAgentRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script tests not supported on Windows")
	}

	t.Run("streams events from subprocess", func(t *testing.T) {
		output := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"read_file","input":{"file_path":"main.go"}}]}}
{"type":"result","cost_usd":0.10,"duration_ms":2500}`

		script := fakeClaudeScript(t, 0, output)
		agent := &ClaudeAgent{Executable: script}

		ch, err := agent.Run(context.Background(), "test prompt", claude.RunOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var events []claude.Event
		for ev := range ch {
			events = append(events, ev)
		}

		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if events[0].Type != claude.EventToolUse {
			t.Errorf("event[0] type: expected %q, got %q", claude.EventToolUse, events[0].Type)
		}
		if events[0].ToolName != "read_file" {
			t.Errorf("event[0] tool name: expected %q, got %q", "read_file", events[0].ToolName)
		}
		if events[1].Type != claude.EventResult {
			t.Errorf("event[1] type: expected %q, got %q", claude.EventResult, events[1].Type)
		}
		if events[1].CostUSD != 0.10 {
			t.Errorf("event[1] cost: expected 0.10, got %.2f", events[1].CostUSD)
		}
	})

	t.Run("non-zero exit sends error event", func(t *testing.T) {
		output := `{"type":"result","cost_usd":0.05,"duration_ms":1000}`
		script := fakeClaudeScript(t, 1, output)
		agent := &ClaudeAgent{Executable: script}

		ch, err := agent.Run(context.Background(), "test", claude.RunOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var events []claude.Event
		for ev := range ch {
			events = append(events, ev)
		}

		// Should have result event + error event from non-zero exit
		if len(events) < 2 {
			t.Fatalf("expected at least 2 events, got %d", len(events))
		}
		last := events[len(events)-1]
		if last.Type != claude.EventError {
			t.Errorf("last event type: expected %q, got %q", claude.EventError, last.Type)
		}
	})

	t.Run("context cancellation stops process", func(t *testing.T) {
		// Script that sleeps forever
		script := fakeClaudeScript(t, 0, "")
		// Rewrite script to sleep
		sleepScript := fmt.Sprintf("#!/bin/sh\nsleep 60\n")
		if err := os.WriteFile(script, []byte(sleepScript), 0755); err != nil {
			t.Fatal(err)
		}

		agent := &ClaudeAgent{Executable: script}
		ctx, cancel := context.WithCancel(context.Background())

		ch, err := agent.Run(ctx, "test", claude.RunOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cancel()

		// Drain channel — should close without hanging
		for range ch {
		}
	})

	t.Run("invalid executable returns error", func(t *testing.T) {
		agent := &ClaudeAgent{Executable: "/nonexistent/binary"}
		_, err := agent.Run(context.Background(), "test", claude.RunOptions{})
		if err == nil {
			t.Fatal("expected error for invalid executable")
		}
	})
}

// Verify ClaudeAgent satisfies claude.Agent at compile time.
var _ claude.Agent = (*ClaudeAgent)(nil)

// fakeClaudeScript creates a shell script that outputs the given content and
// exits with the given code. Returns the path to the script.
func fakeClaudeScript(t *testing.T, exitCode int, output string) string {
	t.Helper()
	dir := t.TempDir()

	// Write output to a data file so we don't have to escape JSON in shell
	dataPath := filepath.Join(dir, "output.txt")
	if err := os.WriteFile(dataPath, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(dir, "claude")
	var buf bytes.Buffer
	buf.WriteString("#!/bin/sh\n")
	buf.WriteString(fmt.Sprintf("cat %s\n", dataPath))
	buf.WriteString(fmt.Sprintf("exit %d\n", exitCode))
	if err := os.WriteFile(script, buf.Bytes(), 0755); err != nil {
		t.Fatal(err)
	}
	return script
}

func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}
