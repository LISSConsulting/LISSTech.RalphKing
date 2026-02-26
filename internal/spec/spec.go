// Package spec provides spec file discovery, status detection, and scaffolding.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Status represents the progress state of a spec file.
type Status string

const (
	StatusDone       Status = "done"
	StatusInProgress Status = "in_progress"
	StatusNotStarted Status = "not_started"
)

// Symbol returns the display indicator for this status.
func (s Status) Symbol() string {
	switch s {
	case StatusDone:
		return "✅"
	case StatusInProgress:
		return "🔄"
	default:
		return "⬜"
	}
}

// String returns a human-readable label.
func (s Status) String() string {
	switch s {
	case StatusDone:
		return "done"
	case StatusInProgress:
		return "in progress"
	default:
		return "not started"
	}
}

// SpecFile represents a discovered spec with its status.
type SpecFile struct {
	Name   string // filename without extension (e.g. "ralph-core")
	Path   string // relative path from project root (e.g. "specs/ralph-core.md")
	Status Status
}

// List discovers specs/*.md and specs/*/*.md files (one level of subdirectories)
// and detects their status by cross-referencing IMPLEMENTATION_PLAN.md.
// The dir argument is the project root directory.
func List(dir string) ([]SpecFile, error) {
	specsDir := filepath.Join(dir, "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	planContent, _ := os.ReadFile(filepath.Join(dir, "IMPLEMENTATION_PLAN.md"))
	plan := string(planContent)

	var specs []SpecFile
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			// Walk one level into subdirectory.
			subEntries, subErr := os.ReadDir(filepath.Join(specsDir, entry.Name()))
			if subErr != nil {
				continue
			}
			for _, sub := range subEntries {
				if sub.IsDir() || !strings.HasSuffix(sub.Name(), ".md") || strings.HasPrefix(sub.Name(), ".") {
					continue
				}
				name := strings.TrimSuffix(sub.Name(), ".md")
				specs = append(specs, SpecFile{
					Name:   name,
					Path:   filepath.Join("specs", entry.Name(), sub.Name()),
					Status: detectStatus(sub.Name(), plan),
				})
			}
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		specs = append(specs, SpecFile{
			Name:   name,
			Path:   filepath.Join("specs", entry.Name()),
			Status: detectStatus(entry.Name(), plan),
		})
	}

	return specs, nil
}

// detectStatus determines the spec's progress by looking at where its filename
// appears in the implementation plan content.
//
// Heuristic:
//   - If the filename appears in a "Remaining Work" section → in_progress
//   - If the filename appears only in a "Completed Work" section → done
//   - If the filename does not appear at all → not_started
func detectStatus(filename, plan string) Status {
	if plan == "" {
		return StatusNotStarted
	}

	completedIdx := strings.Index(plan, "## Completed Work")
	remainingIdx := strings.Index(plan, "## Remaining Work")

	inCompleted := false
	inRemaining := false

	if completedIdx >= 0 {
		completedSection := sectionAfter(plan, completedIdx, remainingIdx)
		inCompleted = strings.Contains(completedSection, filename)
	}

	if remainingIdx >= 0 {
		remainingSection := plan[remainingIdx:]
		inRemaining = strings.Contains(remainingSection, filename)
	}

	switch {
	case inRemaining:
		return StatusInProgress
	case inCompleted:
		return StatusDone
	default:
		// Fallback: if the filename appears anywhere in the plan text
		// (e.g., in an intro or summary outside recognized sections),
		// treat it as in-progress rather than incorrectly "not started".
		if strings.Contains(plan, filename) {
			return StatusInProgress
		}
		return StatusNotStarted
	}
}

// sectionAfter extracts text starting at fromIdx up to (but not including)
// the next section at nextIdx. If nextIdx <= fromIdx or is -1, returns
// text from fromIdx to end.
func sectionAfter(text string, fromIdx, nextIdx int) string {
	if nextIdx > fromIdx {
		return text[fromIdx:nextIdx]
	}
	return text[fromIdx:]
}
