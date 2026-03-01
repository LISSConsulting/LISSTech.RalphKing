package spec

import (
	"os"
	"path/filepath"
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
					if s.IsDir {
						t.Errorf("spec %q should not be IsDir (flat file)", s.Name)
					}
					if s.Path != filepath.Join("specs", s.Name+".md") {
						t.Errorf("spec %q path = %q, want %q", s.Name, s.Path, filepath.Join("specs", s.Name+".md"))
					}
				}
			}
		})
	}
}

func TestList_DirSpecKitLayout(t *testing.T) {
	tests := []struct {
		name       string
		dirName    string
		artifacts  []string // files to create inside the dir
		wantStatus Status
	}{
		{
			name:       "only spec.md → specified",
			dirName:    "001-alpha",
			artifacts:  []string{"spec.md"},
			wantStatus: StatusSpecified,
		},
		{
			name:       "spec.md + plan.md → planned",
			dirName:    "002-beta",
			artifacts:  []string{"spec.md", "plan.md"},
			wantStatus: StatusPlanned,
		},
		{
			name:       "spec.md + plan.md + tasks.md → tasked",
			dirName:    "003-gamma",
			artifacts:  []string{"spec.md", "plan.md", "tasks.md"},
			wantStatus: StatusTasked,
		},
		{
			name:       "empty directory → not_started",
			dirName:    "004-delta",
			artifacts:  []string{},
			wantStatus: StatusNotStarted,
		},
		{
			name:       "extra files + tasks.md → tasked",
			dirName:    "005-epsilon",
			artifacts:  []string{"spec.md", "plan.md", "tasks.md", "research.md", "data-model.md"},
			wantStatus: StatusTasked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			featureDir := filepath.Join(dir, "specs", tt.dirName)
			if err := os.MkdirAll(featureDir, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range tt.artifacts {
				if err := os.WriteFile(filepath.Join(featureDir, f), []byte("# "+f), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			specs, err := List(dir)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}
			if len(specs) != 1 {
				t.Fatalf("List() returned %d specs, want 1 for directory %q", len(specs), tt.dirName)
			}

			s := specs[0]
			if s.Name != tt.dirName {
				t.Errorf("Name = %q, want %q", s.Name, tt.dirName)
			}
			if !s.IsDir {
				t.Error("IsDir should be true for directory-based spec")
			}
			if s.Dir != filepath.Join("specs", tt.dirName) {
				t.Errorf("Dir = %q, want %q", s.Dir, filepath.Join("specs", tt.dirName))
			}
			if s.Path != filepath.Join("specs", tt.dirName, "spec.md") {
				t.Errorf("Path = %q, want %q", s.Path, filepath.Join("specs", tt.dirName, "spec.md"))
			}
			if s.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", s.Status, tt.wantStatus)
			}
		})
	}
}

func TestList_MixedFlatAndDir(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")

	// Flat spec file
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "flat.md"), []byte("# Flat"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Directory-based feature
	featureDir := filepath.Join(specsDir, "001-feature")
	if err := os.MkdirAll(featureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "plan.md"), []byte("# Plan"), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := List(dir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("List() returned %d specs, want 2 (flat + directory)", len(specs))
	}

	byName := make(map[string]SpecFile, len(specs))
	for _, s := range specs {
		byName[s.Name] = s
	}

	flat, ok := byName["flat"]
	if !ok {
		t.Fatal("flat spec not found")
	}
	if flat.IsDir {
		t.Error("flat spec should not be IsDir")
	}
	if flat.Dir != "" {
		t.Errorf("flat spec Dir should be empty, got %q", flat.Dir)
	}

	feat, ok := byName["001-feature"]
	if !ok {
		t.Fatal("001-feature spec not found")
	}
	if !feat.IsDir {
		t.Error("directory spec should be IsDir")
	}
	if feat.Status != StatusPlanned {
		t.Errorf("directory spec status = %v, want %v", feat.Status, StatusPlanned)
	}
}

func TestList_DirHiddenIgnored(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")

	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Hidden directory — should be ignored
	hiddenDir := filepath.Join(specsDir, ".hidden-feature")
	if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "spec.md"), []byte("# Hidden"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Visible feature
	featureDir := filepath.Join(specsDir, "001-visible")
	if err := os.MkdirAll(featureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Visible"), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := List(dir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("List() returned %d specs, want 1 (hidden dir excluded)", len(specs))
	}
	if specs[0].Name != "001-visible" {
		t.Errorf("got spec %q, want 001-visible", specs[0].Name)
	}
}

func TestDetectDirStatus(t *testing.T) {
	tests := []struct {
		name      string
		artifacts []string
		want      Status
	}{
		{
			name:      "no files",
			artifacts: nil,
			want:      StatusNotStarted,
		},
		{
			name:      "spec.md only",
			artifacts: []string{"spec.md"},
			want:      StatusSpecified,
		},
		{
			name:      "plan.md only (unusual)",
			artifacts: []string{"plan.md"},
			want:      StatusPlanned,
		},
		{
			name:      "spec + plan",
			artifacts: []string{"spec.md", "plan.md"},
			want:      StatusPlanned,
		},
		{
			name:      "tasks.md present → tasked regardless of others",
			artifacts: []string{"spec.md", "plan.md", "tasks.md"},
			want:      StatusTasked,
		},
		{
			name:      "tasks.md only",
			artifacts: []string{"tasks.md"},
			want:      StatusTasked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.artifacts {
				if err := os.WriteFile(filepath.Join(dir, f), []byte("# "+f), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := detectDirStatus(dir)
			if got != tt.want {
				t.Errorf("detectDirStatus() = %v, want %v", got, tt.want)
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

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status Status
		symbol string
		str    string
	}{
		{StatusDone, "✅", "done"},
		{StatusInProgress, "🔄", "in progress"},
		{StatusNotStarted, "⬜", "not started"},
		{StatusSpecified, "📋", "specified"},
		{StatusPlanned, "📐", "planned"},
		{StatusTasked, "✅", "tasked"},
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
