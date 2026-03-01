package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ActiveSpec holds the resolved context for a speckit operation.
type ActiveSpec struct {
	Name     string // feature directory name (e.g. "004-speckit-alignment")
	Dir      string // absolute path to the feature directory
	Branch   string // git branch used for resolution (empty if --spec flag used)
	Explicit bool   // true if resolved via --spec flag rather than branch name
}

// Resolve determines the active spec directory from a --spec flag or git branch name.
//
// Resolution order:
//  1. specFlag (--spec flag) — must match an existing specs/<name>/ directory
//  2. branch — derived from the git branch name by stripping numeric prefixes
//     like "004-" when needed; must match an existing specs/<name>/ directory
//
// Returns an error if neither source resolves to an existing directory.
func Resolve(dir, specFlag, branch string) (ActiveSpec, error) {
	if specFlag != "" {
		absDir := filepath.Join(dir, "specs", specFlag)
		if err := checkDir(absDir); err != nil {
			return ActiveSpec{}, fmt.Errorf("spec %q: %w", specFlag, err)
		}
		return ActiveSpec{
			Name:     specFlag,
			Dir:      absDir,
			Explicit: true,
		}, nil
	}

	if branch == "" {
		return ActiveSpec{}, fmt.Errorf("no active spec: not on a feature branch (detached HEAD?); use --spec <name>")
	}

	// Reject main/master branches — they are not feature branches.
	switch branch {
	case "main", "master":
		return ActiveSpec{}, fmt.Errorf("no active spec: branch %q is not a feature branch; use --spec <name>", branch)
	}

	// Try the branch name directly as the spec directory name.
	candidates := branchCandidates(branch)
	for _, name := range candidates {
		absDir := filepath.Join(dir, "specs", name)
		if err := checkDir(absDir); err == nil {
			return ActiveSpec{
				Name:   name,
				Dir:    absDir,
				Branch: branch,
			}, nil
		}
	}

	return ActiveSpec{}, fmt.Errorf("no spec directory found for branch %q; use --spec <name>", branch)
}

// branchCandidates returns candidate spec directory names derived from the
// branch name. The branch name itself is always the first candidate.
func branchCandidates(branch string) []string {
	candidates := []string{branch}
	// Also try stripping a leading NNN- numeric prefix so branch "004-speckit-alignment"
	// can match spec dir "speckit-alignment" if the full name isn't found.
	parts := strings.SplitN(branch, "-", 2)
	if len(parts) == 2 && isNumeric(parts[0]) && parts[1] != "" {
		candidates = append(candidates, parts[1])
	}
	return candidates
}

// isNumeric reports whether s contains only ASCII digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// checkDir returns nil if absDir is an existing directory, or a descriptive error.
func checkDir(absDir string) error {
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist")
		}
		return fmt.Errorf("stat: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}
	return nil
}
