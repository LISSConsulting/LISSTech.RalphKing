package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScaffoldProject creates the full ralph project structure in the given
// directory. It creates ralph.toml, prompt files for plan and build modes,
// and the specs/ directory. Files that already exist are left untouched.
// Returns the list of created paths.
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

	// PROMPT_plan.md
	planPath := filepath.Join(dir, "PROMPT_plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(planPath, []byte(planPromptTemplate), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", planPath, writeErr)
		}
		created = append(created, planPath)
	}

	// PROMPT_build.md
	buildPath := filepath.Join(dir, "PROMPT_build.md")
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
	if os.IsNotExist(err) {
		if writeErr := os.WriteFile(gitignorePath, []byte(gitignoreEntry+"\n"), 0644); writeErr != nil {
			return created, fmt.Errorf("scaffold: write %s: %w", gitignorePath, writeErr)
		}
		created = append(created, gitignorePath)
	} else if err != nil {
		return created, fmt.Errorf("scaffold: read %s: %w", gitignorePath, err)
	} else if !strings.Contains(string(existing), gitignoreEntry) {
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

	return created, nil
}

const planPromptTemplate = `Read the specs in ` + "`specs/`" + ` and study the codebase.
Create or update ` + "`IMPLEMENTATION_PLAN.md`" + ` with:

- A summary of current state (what exists, test coverage)
- Remaining work organized by priority (highest-impact items first)
- Key learnings and architectural decisions

Do NOT write application code — this is a planning phase only.
`

const buildPromptTemplate = `Read the specs in ` + "`specs/`" + ` and the implementation plan.
Pick the highest-priority incomplete item from ` + "`IMPLEMENTATION_PLAN.md`" + `.

1. Study the codebase to understand what already exists.
2. Implement the feature fully — no placeholders, no stubs.
3. Run tests and ensure they pass.
4. Commit with a descriptive message.
5. Update ` + "`IMPLEMENTATION_PLAN.md`" + ` to reflect progress.
`
