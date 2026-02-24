// Package config parses ralph.toml project configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the top-level ralph.toml configuration.
type Config struct {
	Project ProjectConfig `toml:"project"`
	Claude  ClaudeConfig  `toml:"claude"`
	Plan    PlanConfig    `toml:"plan"`
	Build   BuildConfig   `toml:"build"`
	Git     GitConfig     `toml:"git"`
	Regent  RegentConfig  `toml:"regent"`
}

// ProjectConfig identifies the project.
type ProjectConfig struct {
	Name string `toml:"name"`
}

// ClaudeConfig controls the Claude CLI invocation.
type ClaudeConfig struct {
	Model                string `toml:"model"`
	DangerSkipPermissions bool   `toml:"danger_skip_permissions"`
}

// PlanConfig controls the plan loop.
type PlanConfig struct {
	PromptFile    string `toml:"prompt_file"`
	MaxIterations int    `toml:"max_iterations"`
}

// BuildConfig controls the build loop.
type BuildConfig struct {
	PromptFile    string `toml:"prompt_file"`
	MaxIterations int    `toml:"max_iterations"`
}

// GitConfig controls git operations between iterations.
type GitConfig struct {
	AutoPullRebase bool `toml:"auto_pull_rebase"`
	AutoPush       bool `toml:"auto_push"`
}

// RegentConfig controls the Regent supervisor.
type RegentConfig struct {
	Enabled               bool   `toml:"enabled"`
	RollbackOnTestFailure bool   `toml:"rollback_on_test_failure"`
	TestCommand           string `toml:"test_command"`
	MaxRetries            int    `toml:"max_retries"`
	RetryBackoffSeconds   int    `toml:"retry_backoff_seconds"`
	HangTimeoutSeconds    int    `toml:"hang_timeout_seconds"`
}

// Validate checks the configuration for issues that would cause confusing
// runtime failures. It returns all found issues joined together.
func (c *Config) Validate() error {
	var errs []error

	if c.Plan.PromptFile == "" {
		errs = append(errs, fmt.Errorf("plan.prompt_file must not be empty"))
	}
	if c.Build.PromptFile == "" {
		errs = append(errs, fmt.Errorf("build.prompt_file must not be empty"))
	}
	if c.Plan.MaxIterations < 0 {
		errs = append(errs, fmt.Errorf("plan.max_iterations must be >= 0 (0 = unlimited)"))
	}
	if c.Build.MaxIterations < 0 {
		errs = append(errs, fmt.Errorf("build.max_iterations must be >= 0 (0 = unlimited)"))
	}

	if c.Regent.Enabled {
		if c.Regent.MaxRetries < 0 {
			errs = append(errs, fmt.Errorf("regent.max_retries must be >= 0"))
		}
		if c.Regent.RetryBackoffSeconds < 0 {
			errs = append(errs, fmt.Errorf("regent.retry_backoff_seconds must be >= 0"))
		}
		if c.Regent.HangTimeoutSeconds < 0 {
			errs = append(errs, fmt.Errorf("regent.hang_timeout_seconds must be >= 0 (0 = no hang detection)"))
		}
	}

	if c.Regent.RollbackOnTestFailure && c.Regent.TestCommand == "" {
		errs = append(errs, fmt.Errorf("regent.test_command must be set when regent.rollback_on_test_failure is true"))
	}

	return errors.Join(errs...)
}

// Defaults returns a Config with sensible defaults matching the spec.
func Defaults() Config {
	return Config{
		Project: ProjectConfig{Name: ""},
		Claude: ClaudeConfig{
			Model:                "sonnet",
			DangerSkipPermissions: true,
		},
		Plan: PlanConfig{
			PromptFile:    "PROMPT_plan.md",
			MaxIterations: 3,
		},
		Build: BuildConfig{
			PromptFile:    "PROMPT_build.md",
			MaxIterations: 0,
		},
		Git: GitConfig{
			AutoPullRebase: true,
			AutoPush:       true,
		},
		Regent: RegentConfig{
			Enabled:               true,
			RollbackOnTestFailure: false,
			TestCommand:           "",
			MaxRetries:            3,
			RetryBackoffSeconds:   30,
			HangTimeoutSeconds:    300,
		},
	}
}

// Load reads ralph.toml from the given path. If path is empty, it walks up
// from the current working directory looking for ralph.toml.
func Load(path string) (*Config, error) {
	if path == "" {
		found, err := findConfig()
		if err != nil {
			return nil, err
		}
		path = found
	}

	cfg := Defaults()
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("config: decode %s: %w", path, err)
	}
	return &cfg, nil
}

// findConfig walks up from the current directory looking for ralph.toml.
func findConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("config: get working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "ralph.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("config: ralph.toml not found (searched up from %s)", dir)
		}
		dir = parent
	}
}

// InitFile writes a default ralph.toml template to the given directory.
func InitFile(dir string) (string, error) {
	path := filepath.Join(dir, "ralph.toml")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("config: ralph.toml already exists at %s", path)
	}

	content := `# ralph.toml — RalphKing project configuration
# Place this file in the root of your project.

[project]
name = ""

[claude]
model = "sonnet"
danger_skip_permissions = true

[plan]
prompt_file = "PROMPT_plan.md"
max_iterations = 3

[build]
prompt_file = "PROMPT_build.md"
max_iterations = 0  # 0 = unlimited

[git]
auto_pull_rebase = true
auto_push = true

[regent]
enabled = true
rollback_on_test_failure = false
test_command = ""
max_retries = 3
retry_backoff_seconds = 30
hang_timeout_seconds = 300
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("config: write %s: %w", path, err)
	}
	return path, nil
}
