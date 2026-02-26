package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	tests := []struct {
		name        string
		specFiles   []string
		planContent string
		wantCount   int
		wantSpecs   map[string]Status // name → expected status
	}{
		{
			name:      "no specs directory",
			specFiles: nil,
			wantCount: 0,
		},
		{
			name:      "empty specs directory",
			specFiles: []string{},
			wantCount: 0,
		},
		{
			name:      "specs with no implementation plan",
			specFiles: []string{"alpha.md", "beta.md"},
			wantCount: 2,
			wantSpecs: map[string]Status{
				"alpha": StatusNotStarted,
				"beta":  StatusNotStarted,
			},
		},
		{
			name:      "spec referenced in completed work only",
			specFiles: []string{"core.md"},
			planContent: `# Plan
## Completed Work
| Feature | Spec |
|---------|------|
| Config | core.md |
`,
			wantCount: 1,
			wantSpecs: map[string]Status{
				"core": StatusDone,
			},
		},
		{
			name:      "spec referenced in remaining work",
			specFiles: []string{"core.md"},
			planContent: `# Plan
## Completed Work
| Feature | Spec |
|---------|------|
| Config | core.md |

## Remaining Work
### P4
- Spec commands — core.md
`,
			wantCount: 1,
			wantSpecs: map[string]Status{
				"core": StatusInProgress,
			},
		},
		{
			name:      "mixed statuses",
			specFiles: []string{"core.md", "regent.md", "future.md"},
			planContent: `# Plan
## Completed Work
| Feature | Spec |
|---------|------|
| Config | core.md |
| Loop | core.md |
| Status | regent.md |

## Remaining Work
### P4
- Spec commands — core.md
`,
			wantCount: 3,
			wantSpecs: map[string]Status{
				"core":   StatusInProgress,
				"regent": StatusDone,
				"future": StatusNotStarted,
			},
		},
		{
			name:      "spec referenced outside sections (intro/summary)",
			specFiles: []string{"core.md", "regent.md", "future.md"},
			planContent: `# Plan
> Both specs (core.md, regent.md) fully implemented.

## Completed Work
| Feature | Notes |
|---------|-------|
| Config | done |
`,
			wantCount: 3,
			wantSpecs: map[string]Status{
				"core":   StatusInProgress,
				"regent": StatusInProgress,
				"future": StatusNotStarted,
			},
		},
		{
			name:      "non-md files are ignored",
			specFiles: []string{"readme.txt", "notes.md", ".hidden.md"},
			wantCount: 1,
			wantSpecs: map[string]Status{
				"notes": StatusNotStarted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.specFiles != nil {
				specsDir := filepath.Join(dir, "specs")
				if err := os.MkdirAll(specsDir, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, f := range tt.specFiles {
					if err := os.WriteFile(filepath.Join(specsDir, f), []byte("# Spec"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
			}

			if tt.planContent != "" {
				if err := os.WriteFile(filepath.Join(dir, "CHRONICLE.md"), []byte(tt.planContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			specs, err := List(dir)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}

			if len(specs) != tt.wantCount {
				t.Fatalf("List() returned %d specs, want %d", len(specs), tt.wantCount)
			}

			for _, s := range specs {
				if wantStatus, ok := tt.wantSpecs[s.Name]; ok {
					if s.Status != wantStatus {
						t.Errorf("spec %q status = %v, want %v", s.Name, s.Status, wantStatus)
					}
					if s.Path != filepath.Join("specs", s.Name+".md") {
						t.Errorf("spec %q path = %q, want %q", s.Name, s.Path, filepath.Join("specs", s.Name+".md"))
					}
				}
			}
		})
	}
}

func TestList_SubdirLayout(t *testing.T) {
	dir := t.TempDir()

	// Create specs/001-the-genesis/{ralph-core.md,the-regent.md}
	subDir := filepath.Join(dir, "specs", "001-the-genesis")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"ralph-core.md", "the-regent.md"} {
		if err := os.WriteFile(filepath.Join(subDir, f), []byte("# Spec"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Write a plan that marks ralph-core as completed and the-regent as remaining.
	plan := `# Plan
## Completed Work
| Feature | Notes |
|---------|-------|
| Config | ralph-core.md |

## Remaining Work
- Hang detection — the-regent.md
`
	if err := os.WriteFile(filepath.Join(dir, "CHRONICLE.md"), []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := List(dir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("List() returned %d specs, want 2", len(specs))
	}

	byName := make(map[string]SpecFile, len(specs))
	for _, s := range specs {
		byName[s.Name] = s
	}

	wantPath := map[string]string{
		"ralph-core": filepath.Join("specs", "001-the-genesis", "ralph-core.md"),
		"the-regent": filepath.Join("specs", "001-the-genesis", "the-regent.md"),
	}
	wantStatus := map[string]Status{
		"ralph-core": StatusDone,
		"the-regent": StatusInProgress,
	}

	for name, want := range wantStatus {
		s, ok := byName[name]
		if !ok {
			t.Errorf("spec %q not found in results", name)
			continue
		}
		if s.Status != want {
			t.Errorf("spec %q status = %v, want %v", name, s.Status, want)
		}
		if s.Path != wantPath[name] {
			t.Errorf("spec %q path = %q, want %q", name, s.Path, wantPath[name])
		}
	}
}

func TestList_SubdirHiddenAndNested(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")

	// Flat spec
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "flat.md"), []byte("# Flat"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Subdirectory spec
	sub := filepath.Join(specsDir, "v2")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "deep.md"), []byte("# Deep"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Nested two levels deep — should be ignored.
	nested := filepath.Join(sub, "further")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "ignored.md"), []byte("# Ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Hidden file inside subdirectory — should be ignored.
	if err := os.WriteFile(filepath.Join(sub, ".hidden.md"), []byte("# Hidden"), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := List(dir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("List() returned %d specs, want 2 (flat + one subdir)", len(specs))
	}

	byName := make(map[string]SpecFile, len(specs))
	for _, s := range specs {
		byName[s.Name] = s
	}

	if _, ok := byName["flat"]; !ok {
		t.Error("flat spec not found")
	}
	if s, ok := byName["deep"]; !ok {
		t.Error("deep spec not found")
	} else if s.Path != filepath.Join("specs", "v2", "deep.md") {
		t.Errorf("deep spec path = %q, want %q", s.Path, filepath.Join("specs", "v2", "deep.md"))
	}
	if _, ok := byName["ignored"]; ok {
		t.Error("two-levels-deep spec should not be found")
	}
	if _, ok := byName[".hidden"]; ok {
		t.Error("hidden spec should not be found")
	}
}

func TestDetectStatus(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		plan     string
		want     Status
	}{
		{
			name:     "empty plan",
			filename: "core.md",
			plan:     "",
			want:     StatusNotStarted,
		},
		{
			name:     "not mentioned at all",
			filename: "other.md",
			plan:     "## Completed Work\nsome stuff\n## Remaining Work\nmore stuff",
			want:     StatusNotStarted,
		},
		{
			name:     "only in completed",
			filename: "core.md",
			plan:     "## Completed Work\n| Config | core.md |\n",
			want:     StatusDone,
		},
		{
			name:     "in both completed and remaining",
			filename: "core.md",
			plan:     "## Completed Work\n| Config | core.md |\n## Remaining Work\n- TUI — core.md",
			want:     StatusInProgress,
		},
		{
			name:     "only in remaining",
			filename: "future.md",
			plan:     "## Completed Work\n| Config | core.md |\n## Remaining Work\n- Stuff — future.md",
			want:     StatusInProgress,
		},
		{
			name:     "mentioned outside sections (fallback)",
			filename: "core.md",
			plan:     "> Both specs (core.md) fully implemented.\n\n## Completed Work\n| Config | done |",
			want:     StatusInProgress,
		},
		{
			name:     "mentioned outside sections with no sections at all",
			filename: "core.md",
			plan:     "# Plan\nWe reference core.md here but no section headers.",
			want:     StatusInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectStatus(tt.filename, tt.plan)
			if got != tt.want {
				t.Errorf("detectStatus(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("creates spec file", func(t *testing.T) {
		dir := t.TempDir()

		path, err := New(dir, "my-feature")
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}

		wantPath := filepath.Join(dir, "specs", "my-feature.md")
		if path != wantPath {
			t.Errorf("New() path = %q, want %q", path, wantPath)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read created file: %v", err)
		}
		content := string(data)

		if !strings.Contains(content, "My Feature") {
			t.Error("template should contain title-cased name")
		}
		if !strings.Contains(content, "my-feature") {
			t.Error("template should contain kebab-case branch name")
		}
		if !strings.Contains(content, "Feature Specification") {
			t.Error("template should contain spec header")
		}
	})

	t.Run("refuses to overwrite existing spec", func(t *testing.T) {
		dir := t.TempDir()

		if _, err := New(dir, "existing"); err != nil {
			t.Fatalf("first New() error: %v", err)
		}

		_, err := New(dir, "existing")
		if err == nil {
			t.Fatal("second New() should return error for existing spec")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error should mention 'already exists', got: %v", err)
		}
	})

	t.Run("creates specs directory if missing", func(t *testing.T) {
		dir := t.TempDir()

		path, err := New(dir, "new-thing")
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}

		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("created file should exist: %v", statErr)
		}
	})
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-feature", "My Feature"},
		{"single", "Single"},
		{"multi-word-name", "Multi Word Name"},
		{"ALLCAPS", "ALLCAPS"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toTitle(tt.input)
			if got != tt.want {
				t.Errorf("toTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status Status
		symbol string
		str    string
	}{
		{StatusDone, "✅", "done"},
		{StatusInProgress, "🔄", "in progress"},
		{StatusNotStarted, "⬜", "not started"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Symbol(); got != tt.symbol {
				t.Errorf("Symbol() = %q, want %q", got, tt.symbol)
			}
			if got := tt.status.String(); got != tt.str {
				t.Errorf("String() = %q, want %q", got, tt.str)
			}
		})
	}
}
