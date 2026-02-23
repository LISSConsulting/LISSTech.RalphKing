package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"claude.model", cfg.Claude.Model, "sonnet"},
		{"claude.danger_skip_permissions", cfg.Claude.DangerSkipPermissions, true},
		{"plan.prompt_file", cfg.Plan.PromptFile, "PROMPT_plan.md"},
		{"plan.max_iterations", cfg.Plan.MaxIterations, 3},
		{"build.prompt_file", cfg.Build.PromptFile, "PROMPT_build.md"},
		{"build.max_iterations", cfg.Build.MaxIterations, 0},
		{"git.auto_pull_rebase", cfg.Git.AutoPullRebase, true},
		{"git.auto_push", cfg.Git.AutoPush, true},
		{"regent.enabled", cfg.Regent.Enabled, true},
		{"regent.max_retries", cfg.Regent.MaxRetries, 3},
		{"regent.retry_backoff_seconds", cfg.Regent.RetryBackoffSeconds, 30},
		{"regent.hang_timeout_seconds", cfg.Regent.HangTimeoutSeconds, 300},
		{"regent.rollback_on_test_failure", cfg.Regent.RollbackOnTestFailure, false},
		{"regent.test_command", cfg.Regent.TestCommand, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		dir := t.TempDir()
		content := `
[project]
name = "TestProject"

[claude]
model = "opus"
danger_skip_permissions = false

[plan]
prompt_file = "MY_PLAN.md"
max_iterations = 5

[build]
prompt_file = "MY_BUILD.md"
max_iterations = 10

[git]
auto_pull_rebase = false
auto_push = false

[regent]
enabled = false
rollback_on_test_failure = true
test_command = "go test ./..."
max_retries = 5
retry_backoff_seconds = 60
hang_timeout_seconds = 600
`
		path := filepath.Join(dir, "ralph.toml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			name string
			got  any
			want any
		}{
			{"project.name", cfg.Project.Name, "TestProject"},
			{"claude.model", cfg.Claude.Model, "opus"},
			{"claude.danger_skip_permissions", cfg.Claude.DangerSkipPermissions, false},
			{"plan.prompt_file", cfg.Plan.PromptFile, "MY_PLAN.md"},
			{"plan.max_iterations", cfg.Plan.MaxIterations, 5},
			{"build.prompt_file", cfg.Build.PromptFile, "MY_BUILD.md"},
			{"build.max_iterations", cfg.Build.MaxIterations, 10},
			{"git.auto_pull_rebase", cfg.Git.AutoPullRebase, false},
			{"git.auto_push", cfg.Git.AutoPush, false},
			{"regent.enabled", cfg.Regent.Enabled, false},
			{"regent.rollback_on_test_failure", cfg.Regent.RollbackOnTestFailure, true},
			{"regent.test_command", cfg.Regent.TestCommand, "go test ./..."},
			{"regent.max_retries", cfg.Regent.MaxRetries, 5},
			{"regent.retry_backoff_seconds", cfg.Regent.RetryBackoffSeconds, 60},
			{"regent.hang_timeout_seconds", cfg.Regent.HangTimeoutSeconds, 600},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.got != tt.want {
					t.Errorf("got %v, want %v", tt.got, tt.want)
				}
			})
		}
	})

	t.Run("partial config uses defaults", func(t *testing.T) {
		dir := t.TempDir()
		content := `
[project]
name = "Partial"
`
		path := filepath.Join(dir, "ralph.toml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}

		if cfg.Project.Name != "Partial" {
			t.Errorf("project.name: got %q, want %q", cfg.Project.Name, "Partial")
		}
		if cfg.Claude.Model != "sonnet" {
			t.Errorf("claude.model: got %q, want %q (default)", cfg.Claude.Model, "sonnet")
		}
		if cfg.Regent.MaxRetries != 3 {
			t.Errorf("regent.max_retries: got %d, want %d (default)", cfg.Regent.MaxRetries, 3)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := Load("/nonexistent/ralph.toml")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("invalid toml returns error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "ralph.toml")
		if err := os.WriteFile(path, []byte("not valid [[[ toml"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := Load(path)
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})
}

func TestLoadAutoDiscovery(t *testing.T) {
	t.Run("finds ralph.toml in parent directory", func(t *testing.T) {
		root := t.TempDir()
		child := filepath.Join(root, "sub", "dir")
		if err := os.MkdirAll(child, 0755); err != nil {
			t.Fatal(err)
		}

		// Write ralph.toml in root
		content := `[project]
name = "FoundIt"
`
		if err := os.WriteFile(filepath.Join(root, "ralph.toml"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Change to child directory to test walk-up
		origDir, _ := os.Getwd()
		t.Cleanup(func() { os.Chdir(origDir) })
		if err := os.Chdir(child); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load("")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Project.Name != "FoundIt" {
			t.Errorf("project.name: got %q, want %q", cfg.Project.Name, "FoundIt")
		}
	})

	t.Run("returns error when ralph.toml not found anywhere", func(t *testing.T) {
		dir := t.TempDir()
		origDir, _ := os.Getwd()
		t.Cleanup(func() { os.Chdir(origDir) })
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}

		_, err := Load("")
		if err == nil {
			t.Error("expected error when ralph.toml not found")
		}
	})
}

func TestInitFile(t *testing.T) {
	t.Run("creates ralph.toml", func(t *testing.T) {
		dir := t.TempDir()
		path, err := InitFile(dir)
		if err != nil {
			t.Fatal(err)
		}

		if filepath.Base(path) != "ralph.toml" {
			t.Errorf("expected ralph.toml, got %s", filepath.Base(path))
		}

		// Verify it's valid TOML by loading it
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("generated file is not valid: %v", err)
		}
		if cfg.Claude.Model != "sonnet" {
			t.Errorf("default model: got %q, want %q", cfg.Claude.Model, "sonnet")
		}
	})

	t.Run("refuses to overwrite existing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "ralph.toml")
		if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := InitFile(dir)
		if err == nil {
			t.Error("expected error when ralph.toml already exists")
		}
	})
}
