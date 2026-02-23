package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		specFiles  []string
		planContent string
		wantCount  int
		wantSpecs  map[string]Status // name → expected status
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
				if err := os.WriteFile(filepath.Join(dir, "IMPLEMENTATION_PLAN.md"), []byte(tt.planContent), 0o644); err != nil {
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
