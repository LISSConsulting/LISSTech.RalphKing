package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/regent"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/spec"
)

// findNoop returns a no-op command that accepts any args and exits 0.
// Returns ("", false) if no such command is available (e.g. Windows).
func findNoop() (string, bool) {
	path, err := exec.LookPath("true")
	if err != nil {
		return "", false
	}
	return path, true
}

func TestFormatSpecList(t *testing.T) {
	tests := []struct {
		name     string
		specs    []spec.SpecFile
		contains []string
		excludes []string
	}{
		{
			name:     "empty — no specs message",
			specs:    nil,
			contains: []string{"No specs found"},
		},
		{
			name: "single done spec",
			specs: []spec.SpecFile{
				{Name: "ralph-core", Path: "specs/ralph-core.md", Status: spec.StatusDone},
			},
			contains: []string{"Specs", "─────", "✅", "specs/ralph-core.md", "done"},
		},
		{
			name: "multiple specs with mixed statuses",
			specs: []spec.SpecFile{
				{Name: "ralph-core", Path: "specs/ralph-core.md", Status: spec.StatusDone},
				{Name: "the-regent", Path: "specs/the-regent.md", Status: spec.StatusInProgress},
				{Name: "new-feature", Path: "specs/new-feature.md", Status: spec.StatusNotStarted},
			},
			contains: []string{
				"Specs", "─────",
				"✅", "specs/ralph-core.md", "done",
				"🔄", "specs/the-regent.md", "in progress",
				"⬜", "specs/new-feature.md", "not started",
			},
		},
		{
			name:     "empty slice — same as nil",
			specs:    []spec.SpecFile{},
			contains: []string{"No specs found"},
			excludes: []string{"Specs", "─────"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSpecList(tt.specs)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output should contain %q\ngot:\n%s", want, got)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(got, exclude) {
					t.Errorf("output should NOT contain %q\ngot:\n%s", exclude, got)
				}
			}
		})
	}
}

func TestFormatScaffoldResult(t *testing.T) {
	tests := []struct {
		name     string
		created  []string
		contains []string
		excludes []string
	}{
		{
			name:     "nothing created — already exists message",
			created:  nil,
			contains: []string{"All files already exist"},
			excludes: []string{"Created"},
		},
		{
			name:     "empty slice — same as nil",
			created:  []string{},
			contains: []string{"All files already exist"},
		},
		{
			name:    "single file created",
			created: []string{"ralph.toml"},
			contains: []string{"Created ralph.toml"},
			excludes: []string{"already exist"},
		},
		{
			name:    "multiple files created",
			created: []string{"ralph.toml", "PLAN.md", "BUILD.md", "specs/"},
			contains: []string{
				"Created ralph.toml",
				"Created PLAN.md",
				"Created BUILD.md",
				"Created specs/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatScaffoldResult(tt.created)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output should contain %q\ngot:\n%s", want, got)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(got, exclude) {
					t.Errorf("output should NOT contain %q\ngot:\n%s", exclude, got)
				}
			}
		})
	}
}

func TestRootCmdStructure(t *testing.T) {
	root := rootCmd()

	if root.Use != "ralph" {
		t.Errorf("root Use = %q, want %q", root.Use, "ralph")
	}

	// --no-tui persistent flag must exist
	noTUI := root.PersistentFlags().Lookup("no-tui")
	if noTUI == nil {
		t.Fatal("missing --no-tui persistent flag")
	}

	// Verify all expected subcommands
	subs := map[string]bool{}
	for _, sub := range root.Commands() {
		subs[sub.Name()] = true
	}
	for _, want := range []string{"plan", "build", "run", "status", "init", "spec"} {
		if !subs[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestLoopCmdsHaveMaxFlag(t *testing.T) {
	root := rootCmd()

	for _, name := range []string{"plan", "build", "run"} {
		t.Run(name, func(t *testing.T) {
			for _, sub := range root.Commands() {
				if sub.Name() == name {
					if sub.Flags().Lookup("max") == nil {
						t.Errorf("%s: missing --max flag", name)
					}
					return
				}
			}
			t.Fatalf("subcommand %q not found", name)
		})
	}
}

func TestSpecCmdSubcommands(t *testing.T) {
	root := rootCmd()

	for _, sub := range root.Commands() {
		if sub.Name() != "spec" {
			continue
		}
		specSubs := map[string]bool{}
		for _, child := range sub.Commands() {
			specSubs[child.Name()] = true
		}
		for _, want := range []string{"list", "new"} {
			if !specSubs[want] {
				t.Errorf("spec: missing subcommand %q", want)
			}
		}
		return
	}
	t.Fatal("missing spec subcommand")
}

// --- End-to-end command execution tests ---
// These use t.Chdir (Go 1.24) to test command RunE handlers with real I/O.

func TestInitCmdExecution(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd := initCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("initCmd RunE: %v", err)
	}

	// Verify core files were created
	for _, name := range []string{"ralph.toml", "PLAN.md", "BUILD.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestInitCmdIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd1 := initCmd()
	if err := cmd1.RunE(cmd1, nil); err != nil {
		t.Fatalf("first initCmd RunE: %v", err)
	}

	// Second run should succeed without creating anything
	cmd2 := initCmd()
	if err := cmd2.RunE(cmd2, nil); err != nil {
		t.Fatalf("second initCmd RunE: %v", err)
	}
}

func TestInitCmd_ScaffoldError(t *testing.T) {
	// Trigger ScaffoldProject returning an error by creating .gitignore as a
	// directory. Pre-create all files that scaffold checks before .gitignore so
	// the function progresses past them and reaches the .gitignore read step.
	dir := t.TempDir()
	t.Chdir(dir)
	for name, content := range map[string]string{
		"ralph.toml": "x",
		"PLAN.md":    "x",
		"BUILD.md":   "x",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s: %v", name, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "specs"), 0755); err != nil {
		t.Fatalf("MkdirAll specs: %v", err)
	}
	// .gitignore as a directory → os.ReadFile returns a non-IsNotExist error.
	if err := os.MkdirAll(filepath.Join(dir, ".gitignore"), 0755); err != nil {
		t.Fatalf("MkdirAll .gitignore: %v", err)
	}

	cmd := initCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when ScaffoldProject fails")
	}
	if !strings.Contains(err.Error(), ".gitignore") {
		t.Errorf("error should mention .gitignore, got: %v", err)
	}
}

func TestStatusCmdExecution(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	now := time.Now()
	state := regent.State{
		RalphPID:     123,
		Iteration:    5,
		Branch:       "main",
		TotalCostUSD: 1.50,
		StartedAt:    now.Add(-10 * time.Minute),
		FinishedAt:   now,
		Passed:       true,
	}
	if err := regent.SaveState(dir, state); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	cmd := statusCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("statusCmd RunE: %v", err)
	}
}

func TestStatusCmdNoState(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// No state file — should succeed (LoadState returns empty state)
	cmd := statusCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("statusCmd RunE: %v", err)
	}
}

func TestSpecListCmdExecution(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Create specs directory with a test spec
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "test-spec.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cmd := specListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("specListCmd RunE: %v", err)
	}
}

func TestSpecListCmdNoSpecs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// No specs directory — should succeed
	cmd := specListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("specListCmd RunE: %v", err)
	}
}

func TestSpecNewCmdExecution(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("EDITOR", "")

	cmd := specNewCmd()
	if err := cmd.RunE(cmd, []string{"my-feature"}); err != nil {
		t.Fatalf("specNewCmd RunE: %v", err)
	}

	path := filepath.Join(dir, "specs", "my-feature.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected spec file at %s: %v", path, err)
	}
}

func TestSpecNewCmdWithEditor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping editor launch in short mode")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	// Use "true" which accepts any args and exits 0 (Unix only; skip on Windows).
	editor, ok := findNoop()
	if !ok {
		t.Skip("no no-op editor command available on this platform")
	}
	t.Setenv("EDITOR", editor)

	cmd := specNewCmd()
	if err := cmd.RunE(cmd, []string{"editor-test"}); err != nil {
		t.Fatalf("specNewCmd RunE with EDITOR=%q: %v", editor, err)
	}

	path := filepath.Join(dir, "specs", "editor-test.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected spec file at %s: %v", path, err)
	}
}

func TestSpecNewCmdExisting(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("EDITOR", "")

	// Create spec first.
	cmd1 := specNewCmd()
	if err := cmd1.RunE(cmd1, []string{"my-feature"}); err != nil {
		t.Fatalf("first specNewCmd RunE: %v", err)
	}

	// Second creation with same name should fail with "already exists" error.
	cmd2 := specNewCmd()
	err := cmd2.RunE(cmd2, []string{"my-feature"})
	if err == nil {
		t.Fatal("expected error when spec already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestSpecListCmd_SpecsNotDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows, os.ReadDir on a regular file returns an IsNotExist-like error,
		// so spec.List returns nil rather than propagating the error. The error path
		// covered by this test is only reachable on Unix.
		t.Skip("ReadDir on a regular file returns IsNotExist on Windows")
	}
	dir := t.TempDir()
	t.Chdir(dir)

	// Create a regular file named "specs" so ReadDir returns a non-IsNotExist error.
	if err := os.WriteFile(filepath.Join(dir, "specs"), []byte("not a dir"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cmd := specListCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when specs/ is a regular file, not a directory")
	}
}

// ---- RunE handler tests for plan / build / run commands ----
//
// These exercise the RunE closures (currently at 50%) by running them
// in a temp dir with no ralph.toml, which causes config.Load to fail
// immediately before any TUI or Claude invocation.

func TestPlanCmdRunE_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := planCmd()
	if err := cmd.Flags().Set("max", "1"); err != nil {
		t.Fatalf("set --max flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when ralph.toml not found")
	}
	if !strings.Contains(err.Error(), "ralph.toml") {
		t.Errorf("error should mention ralph.toml, got: %v", err)
	}
}

func TestBuildCmdRunE_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := buildCmd()
	if err := cmd.Flags().Set("max", "1"); err != nil {
		t.Fatalf("set --max flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when ralph.toml not found")
	}
	if !strings.Contains(err.Error(), "ralph.toml") {
		t.Errorf("error should mention ralph.toml, got: %v", err)
	}
}

func TestRunCmdRunE_NoConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := runCmd()
	if err := cmd.Flags().Set("max", "1"); err != nil {
		t.Fatalf("set --max flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when ralph.toml not found")
	}
	if !strings.Contains(err.Error(), "ralph.toml") {
		t.Errorf("error should mention ralph.toml, got: %v", err)
	}
}
