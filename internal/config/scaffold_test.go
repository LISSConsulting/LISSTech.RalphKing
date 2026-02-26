package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldProject(t *testing.T) {
	t.Run("creates all files in empty directory", func(t *testing.T) {
		dir := t.TempDir()

		created, err := ScaffoldProject(dir)
		if err != nil {
			t.Fatal(err)
		}

		expected := []string{
			filepath.Join(dir, "ralph.toml"),
			filepath.Join(dir, "PROMPT_plan.md"),
			filepath.Join(dir, "PROMPT_build.md"),
			filepath.Join(dir, "specs"),
		}

		if len(created) != len(expected) {
			t.Fatalf("created %d files, want %d: %v", len(created), len(expected), created)
		}
		for i, want := range expected {
			if created[i] != want {
				t.Errorf("created[%d] = %q, want %q", i, created[i], want)
			}
		}

		// Verify files exist and are non-empty
		for _, path := range expected[:3] {
			info, err := os.Stat(path)
			if err != nil {
				t.Errorf("expected file %s to exist: %v", path, err)
				continue
			}
			if info.Size() == 0 {
				t.Errorf("expected file %s to be non-empty", path)
			}
		}

		// Verify specs/ is a directory
		info, err := os.Stat(filepath.Join(dir, "specs"))
		if err != nil {
			t.Fatalf("specs dir: %v", err)
		}
		if !info.IsDir() {
			t.Error("specs should be a directory")
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		dir := t.TempDir()

		// Pre-create ralph.toml and PROMPT_build.md
		if err := os.WriteFile(filepath.Join(dir, "ralph.toml"), []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "PROMPT_build.md"), []byte("custom build prompt"), 0644); err != nil {
			t.Fatal(err)
		}

		created, err := ScaffoldProject(dir)
		if err != nil {
			t.Fatal(err)
		}

		// Should only create the missing files
		expected := []string{
			filepath.Join(dir, "PROMPT_plan.md"),
			filepath.Join(dir, "specs"),
		}
		if len(created) != len(expected) {
			t.Fatalf("created %d files, want %d: %v", len(created), len(expected), created)
		}
		for i, want := range expected {
			if created[i] != want {
				t.Errorf("created[%d] = %q, want %q", i, created[i], want)
			}
		}

		// Verify pre-existing file was not overwritten
		content, err := os.ReadFile(filepath.Join(dir, "PROMPT_build.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "custom build prompt" {
			t.Error("PROMPT_build.md was overwritten")
		}
	})

	t.Run("all files exist returns empty list", func(t *testing.T) {
		dir := t.TempDir()

		// Create all files
		if err := os.WriteFile(filepath.Join(dir, "ralph.toml"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "PROMPT_plan.md"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "PROMPT_build.md"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "specs"), 0755); err != nil {
			t.Fatal(err)
		}

		created, err := ScaffoldProject(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(created) != 0 {
			t.Errorf("expected empty list, got %v", created)
		}
	})

	t.Run("plan prompt template contains key instructions", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := ScaffoldProject(dir); err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "PROMPT_plan.md"))
		if err != nil {
			t.Fatal(err)
		}

		for _, want := range []string{"specs/", "IMPLEMENTATION_PLAN.md", "planning phase"} {
			if !strings.Contains(string(content), want) {
				t.Errorf("plan prompt should contain %q", want)
			}
		}
	})

	t.Run("build prompt template contains key instructions", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := ScaffoldProject(dir); err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "PROMPT_build.md"))
		if err != nil {
			t.Fatal(err)
		}

		for _, want := range []string{"specs/", "IMPLEMENTATION_PLAN.md", "Implement"} {
			if !strings.Contains(string(content), want) {
				t.Errorf("build prompt should contain %q", want)
			}
		}
	})

	t.Run("ralph.toml created by scaffold is loadable", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := ScaffoldProject(dir); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(filepath.Join(dir, "ralph.toml"))
		if err != nil {
			t.Fatalf("scaffold ralph.toml is not valid: %v", err)
		}
		if cfg.Claude.Model != "sonnet" {
			t.Errorf("default model: got %q, want %q", cfg.Claude.Model, "sonnet")
		}
	})
}
