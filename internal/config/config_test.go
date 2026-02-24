package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{
			name:   "defaults are valid",
			modify: func(c *Config) {},
		},
		{
			name:   "empty plan.prompt_file",
			modify: func(c *Config) { c.Plan.PromptFile = "" },
			wantErr: "plan.prompt_file must not be empty",
		},
		{
			name:   "empty build.prompt_file",
			modify: func(c *Config) { c.Build.PromptFile = "" },
			wantErr: "build.prompt_file must not be empty",
		},
		{
			name:   "negative plan.max_iterations",
			modify: func(c *Config) { c.Plan.MaxIterations = -1 },
			wantErr: "plan.max_iterations must be >= 0",
		},
		{
			name:   "negative build.max_iterations",
			modify: func(c *Config) { c.Build.MaxIterations = -1 },
			wantErr: "build.max_iterations must be >= 0",
		},
		{
			name: "negative regent.max_retries when enabled",
			modify: func(c *Config) {
				c.Regent.Enabled = true
				c.Regent.MaxRetries = -1
			},
			wantErr: "regent.max_retries must be >= 0",
		},
		{
			name: "negative regent.retry_backoff_seconds when enabled",
			modify: func(c *Config) {
				c.Regent.Enabled = true
				c.Regent.RetryBackoffSeconds = -1
			},
			wantErr: "regent.retry_backoff_seconds must be >= 0",
		},
		{
			name: "negative regent.hang_timeout_seconds when enabled",
			modify: func(c *Config) {
				c.Regent.Enabled = true
				c.Regent.HangTimeoutSeconds = -1
			},
			wantErr: "regent.hang_timeout_seconds must be >= 0",
		},
		{
			name: "regent numeric checks skipped when disabled",
			modify: func(c *Config) {
				c.Regent.Enabled = false
				c.Regent.MaxRetries = -1
				c.Regent.RetryBackoffSeconds = -1
				c.Regent.HangTimeoutSeconds = -1
			},
		},
		{
			name: "rollback_on_test_failure without test_command skipped when disabled",
			modify: func(c *Config) {
				c.Regent.Enabled = false
				c.Regent.RollbackOnTestFailure = true
				c.Regent.TestCommand = ""
			},
		},
		{
			name: "rollback_on_test_failure without test_command",
			modify: func(c *Config) {
				c.Regent.RollbackOnTestFailure = true
				c.Regent.TestCommand = ""
			},
			wantErr: "regent.test_command must be set",
		},
		{
			name: "rollback_on_test_failure with test_command",
			modify: func(c *Config) {
				c.Regent.RollbackOnTestFailure = true
				c.Regent.TestCommand = "go test ./..."
			},
		},
		{
			name: "zero max_iterations is valid (unlimited)",
			modify: func(c *Config) {
				c.Plan.MaxIterations = 0
				c.Build.MaxIterations = 0
			},
		},
		{
			name: "zero hang_timeout is valid (no hang detection)",
			modify: func(c *Config) {
				c.Regent.Enabled = true
				c.Regent.HangTimeoutSeconds = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Defaults()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	cfg := Defaults()
	cfg.Plan.PromptFile = ""
	cfg.Build.PromptFile = ""
	cfg.Plan.MaxIterations = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	expected := []string{
		"plan.prompt_file must not be empty",
		"build.prompt_file must not be empty",
		"plan.max_iterations must be >= 0",
	}
	for _, want := range expected {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not contain %q", msg, want)
		}
	}
}
