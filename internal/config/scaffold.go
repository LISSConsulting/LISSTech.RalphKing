package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScaffoldProject creates the full ralph project structure in the given
// directory. It creates ralph.toml, prompt files for plan and build modes,
// the specs/ directory, .gitignore, and CHRONICLE.md. Files that
// already exist are left untouched. Returns the list of created paths.
func ScaffoldProject(dir string) ([]string, error) {
	var created []string

	// ralph.toml
	tomlPath := filepath.Join(dir, "ralph.toml")
	if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
		if _, initErr := InitFile(dir); initErr != nil {
			return created, initErr
		}
		created = append(created, tomlPath)
	}

	// PLAN.md
	planPath := filepath.Join(dir, "PLAN.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(planPath, []byte(planPromptTemplate), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", planPath, writeErr)
		}
		created = append(created, planPath)
	}

	// BUILD.md
	buildPath := filepath.Join(dir, "BUILD.md")
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(buildPath, []byte(buildPromptTemplate), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", buildPath, writeErr)
		}
		created = append(created, buildPath)
	}

	// specs/ directory
	specsDir := filepath.Join(dir, "specs")
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(specsDir, 0755); mkErr != nil {
			return created, fmt.Errorf("scaffold: create %s: %w", specsDir, mkErr)
		}
		created = append(created, specsDir)
	}

	// .gitignore — ensure the regent state file is excluded from version control
	const gitignoreEntry = ".ralph/regent-state.json"
	gitignorePath := filepath.Join(dir, ".gitignore")
	existing, err := os.ReadFile(gitignorePath)
	switch {
	case os.IsNotExist(err):
		if writeErr := os.WriteFile(gitignorePath, []byte(gitignoreEntry+"\n"), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", gitignorePath, writeErr)
		}
		created = append(created, gitignorePath)
	case err != nil:
		return created, fmt.Errorf("scaffold: read %s: %w", gitignorePath, err)
	case !strings.Contains(string(existing), gitignoreEntry):
		content := string(existing)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content += "\n"
		}
		content += gitignoreEntry + "\n"
		if writeErr := os.WriteFile(gitignorePath, []byte(content), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", gitignorePath, writeErr)
		}
		created = append(created, gitignorePath)
	}

	// CHRONICLE.md
	chroniclePath := filepath.Join(dir, "CHRONICLE.md")
	if _, err := os.Stat(chroniclePath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(chroniclePath, []byte(implementationPlanTemplate), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", chroniclePath, writeErr)
		}
		created = append(created, chroniclePath)
	}

	return created, nil
}

const planPromptTemplate = `Read the specs in ` + "`specs/`" + ` and study the codebase.
Create or update ` + "`CHRONICLE.md`" + ` with:

- A summary of current state (what exists, test coverage)
- Remaining work organized by priority (highest-impact items first)
- Key learnings and architectural decisions

Do NOT write application code — this is a planning phase only.
`

const buildPromptTemplate = `Read the specs in ` + "`specs/`" + ` and the implementation plan.
Pick the highest-priority incomplete item from ` + "`CHRONICLE.md`" + `.

1. Study the codebase to understand what already exists.
2. Implement the feature fully — no placeholders, no stubs.
3. Run tests and ensure they pass.
4. Commit with a descriptive message.
5. Update ` + "`CHRONICLE.md`" + ` to reflect progress.
`

const implementationPlanTemplate = `> [Project]: spec-driven AI coding loop.
> Current state: **Initialization complete.** Specs pending implementation.

## Completed Work

| Phase | Features | Tags |
|-------|----------|------|

## Remaining Work

| Priority | Item | Location | Notes |
|----------|------|----------|-------|

## Key Learnings

-
`
