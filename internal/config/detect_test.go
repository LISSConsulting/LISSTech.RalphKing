package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectName(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string // filename -> content
		want  string
	}{
		{
			name:  "no manifest files returns empty",
			files: map[string]string{},
			want:  "",
		},
		{
			name: "pyproject.toml PEP 621 [project] name",
			files: map[string]string{
				"pyproject.toml": `[project]
name = "my-python-project"
`,
			},
			want: "my-python-project",
		},
		{
			name: "pyproject.toml [tool.poetry] name when [project] absent",
			files: map[string]string{
				"pyproject.toml": `[tool.poetry]
name = "my-poetry-project"
`,
			},
			want: "my-poetry-project",
		},
		{
			name: "pyproject.toml [project] wins over [tool.poetry]",
			files: map[string]string{
				"pyproject.toml": `[project]
name = "pep621-name"

[tool.poetry]
name = "poetry-name"
`,
			},
			want: "pep621-name",
		},
		{
			name: "package.json top-level name",
			files: map[string]string{
				"package.json": `{"name": "my-node-project", "version": "1.0.0"}`,
			},
			want: "my-node-project",
		},
		{
			name: "Cargo.toml [package] name",
			files: map[string]string{
				"Cargo.toml": `[package]
name = "my-rust-project"
version = "0.1.0"
`,
			},
			want: "my-rust-project",
		},
		{
			name: "pyproject.toml wins over package.json",
			files: map[string]string{
				"pyproject.toml": `[project]
name = "python-wins"
`,
				"package.json": `{"name": "node-loses"}`,
			},
			want: "python-wins",
		},
		{
			name: "package.json wins over Cargo.toml",
			files: map[string]string{
				"package.json": `{"name": "node-wins"}`,
				"Cargo.toml": `[package]
name = "rust-loses"
`,
			},
			want: "node-wins",
		},
		{
			name: "malformed pyproject.toml falls through to package.json",
			files: map[string]string{
				"pyproject.toml": `not valid [[[ toml`,
				"package.json":   `{"name": "fallback-node"}`,
			},
			want: "fallback-node",
		},
		{
			name: "malformed package.json falls through to Cargo.toml",
			files: map[string]string{
				"package.json": `not valid json`,
				"Cargo.toml": `[package]
name = "fallback-rust"
`,
			},
			want: "fallback-rust",
		},
		{
			name: "pyproject.toml with empty name falls through to package.json",
			files: map[string]string{
				"pyproject.toml": `[project]
name = ""
`,
				"package.json": `{"name": "node-picks-up"}`,
			},
			want: "node-picks-up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := DetectProjectName(dir)
			if got != tt.want {
				t.Errorf("DetectProjectName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadDetectsProjectName(t *testing.T) {
	t.Run("auto-detects from pyproject.toml when project.name empty", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "ralph.toml"), `[plan]
prompt_file = "PROMPT_plan.md"
[build]
prompt_file = "PROMPT_build.md"
`)
		writeFile(t, filepath.Join(dir, "pyproject.toml"), `[project]
name = "detected-python"
`)

		cfg, err := Load(filepath.Join(dir, "ralph.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Project.Name != "detected-python" {
			t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "detected-python")
		}
	})

	t.Run("explicit project.name in ralph.toml is not overwritten", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "ralph.toml"), `[project]
name = "explicit-name"
[plan]
prompt_file = "PROMPT_plan.md"
[build]
prompt_file = "PROMPT_build.md"
`)
		writeFile(t, filepath.Join(dir, "pyproject.toml"), `[project]
name = "should-not-appear"
`)

		cfg, err := Load(filepath.Join(dir, "ralph.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Project.Name != "explicit-name" {
			t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "explicit-name")
		}
	})

	t.Run("no manifest files leaves project.name empty", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "ralph.toml"), `[plan]
prompt_file = "PROMPT_plan.md"
[build]
prompt_file = "PROMPT_build.md"
`)

		cfg, err := Load(filepath.Join(dir, "ralph.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Project.Name != "" {
			t.Errorf("Project.Name = %q, want empty", cfg.Project.Name)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
