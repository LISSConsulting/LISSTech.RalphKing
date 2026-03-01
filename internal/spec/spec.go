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
	// Legacy statuses (flat-file specs, CHRONICLE.md-based).
	StatusDone       Status = "done"
	StatusInProgress Status = "in_progress"
	StatusNotStarted Status = "not_started"

	// Directory-based statuses (artifact-presence detection).
	StatusSpecified Status = "specified" // spec.md exists
	StatusPlanned   Status = "planned"   // plan.md exists
	StatusTasked    Status = "tasked"    // tasks.md exists
)

// Symbol returns the display indicator for this status.
func (s Status) Symbol() string {
	switch s {
	case StatusDone:
		return "✅"
	case StatusInProgress:
		return "🔄"
	case StatusSpecified:
		return "📋"
	case StatusPlanned:
		return "📐"
	case StatusTasked:
		return "✅"
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
	case StatusSpecified:
		return "specified"
	case StatusPlanned:
		return "planned"
	case StatusTasked:
		return "tasked"
	default:
		return "not started"
	}
}

// SpecFile represents a discovered spec with its status.
type SpecFile struct {
	Name   string // feature name (e.g. "ralph-core" or "004-speckit-alignment")
	Path   string // relative path from project root (e.g. "specs/ralph-core.md" or "specs/004-speckit-alignment/spec.md")
	Dir    string // relative path to feature directory (e.g. "specs/004-speckit-alignment"); empty for flat files
	IsDir  bool   // true if this is a directory-based feature
	Status Status
}

// List discovers spec features in the specs/ directory.
//
// Directory entries (specs/NNN-name/) are treated as single features with
// artifact-presence-based status (spec.md→specified, plan.md→planned, tasks.md→tasked).
//
// Flat .md files (specs/name.md) use the legacy CHRONICLE.md-based status detection.
//
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

	planContent, _ := os.ReadFile(filepath.Join(dir, "CHRONICLE.md"))
	plan := string(planContent)

	var specs []SpecFile
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			// Directory-based feature: emit one SpecFile per directory.
			featureDir := filepath.Join("specs", entry.Name())
			absDir := filepath.Join(specsDir, entry.Name())
			specs = append(specs, SpecFile{
				Name:   entry.Name(),
				Dir:    featureDir,
				Path:   filepath.Join(featureDir, "spec.md"),
				IsDir:  true,
				Status: detectDirStatus(absDir),
			})
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

// detectDirStatus determines a directory-based spec's status by checking which
// artifact files are present. Priority: tasks.md > plan.md > spec.md > not_started.
func detectDirStatus(absDir string) Status {
	switch {
	case fileExists(filepath.Join(absDir, "tasks.md")):
		return StatusTasked
	case fileExists(filepath.Join(absDir, "plan.md")):
		return StatusPlanned
	case fileExists(filepath.Join(absDir, "spec.md")):
		return StatusSpecified
	default:
		return StatusNotStarted
	}
}

// fileExists reports whether a regular file exists at path.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
