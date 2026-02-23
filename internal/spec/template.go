package spec

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed spec-template.md
var defaultTemplate string

// New creates a new spec file at specs/<name>.md in the given project directory.
// It returns the absolute path to the created file.
// Returns an error if the file already exists.
func New(dir, name string) (string, error) {
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return "", fmt.Errorf("create specs directory: %w", err)
	}

	filename := name + ".md"
	path := filepath.Join(specsDir, filename)

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("spec already exists: %s", path)
	}

	content := renderTemplate(defaultTemplate, name)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write spec file: %w", err)
	}

	return path, nil
}

// renderTemplate replaces placeholders in the spec template with actual values.
func renderTemplate(tmpl, name string) string {
	title := toTitle(name)
	date := time.Now().Format("2006-01-02")

	result := strings.ReplaceAll(tmpl, "[FEATURE NAME]", title)
	result = strings.ReplaceAll(result, "[DATE]", date)
	result = strings.ReplaceAll(result, "[###-feature-name]", name)

	return result
}

// toTitle converts a kebab-case name to Title Case.
// e.g. "my-feature" → "My Feature"
func toTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
