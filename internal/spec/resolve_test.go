package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		specDirs []string // directories to create under specs/
		specFlag string
		branch   string
		wantName string
		wantErr  string
		wantExpl bool
	}{
		{
			name:     "spec flag matches existing directory",
			specDirs: []string{"004-speckit-alignment"},
			specFlag: "004-speckit-alignment",
			wantName: "004-speckit-alignment",
			wantExpl: true,
		},
		{
			name:     "branch name matches directory exactly",
			specDirs: []string{"004-speckit-alignment"},
			branch:   "004-speckit-alignment",
			wantName: "004-speckit-alignment",
		},
		{
			name:     "branch with numeric prefix matches dir without prefix",
			specDirs: []string{"speckit-alignment"},
			branch:   "004-speckit-alignment",
			wantName: "speckit-alignment",
		},
		{
			name:     "spec flag overrides branch",
			specDirs: []string{"001-first", "002-second"},
			specFlag: "001-first",
			branch:   "002-second",
			wantName: "001-first",
			wantExpl: true,
		},
		{
			name:     "spec flag missing directory → error",
			specFlag: "nonexistent",
			wantErr:  "nonexistent",
		},
		{
			name:    "empty branch with no spec flag → error",
			branch:  "",
			wantErr: "detached HEAD",
		},
		{
			name:    "main branch → error suggesting --spec",
			branch:  "main",
			wantErr: "--spec",
		},
		{
			name:    "master branch → error suggesting --spec",
			branch:  "master",
			wantErr: "--spec",
		},
		{
			name:    "branch with no matching directory → error",
			branch:  "no-match-feature",
			wantErr: "--spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, d := range tt.specDirs {
				if err := os.MkdirAll(filepath.Join(dir, "specs", d), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			got, err := Resolve(dir, tt.specFlag, tt.branch)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Resolve() expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Resolve() error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Dir != filepath.Join(dir, "specs", tt.wantName) {
				t.Errorf("Dir = %q, want %q", got.Dir, filepath.Join(dir, "specs", tt.wantName))
			}
			if got.Explicit != tt.wantExpl {
				t.Errorf("Explicit = %v, want %v", got.Explicit, tt.wantExpl)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"0", true},
		{"123", true},
		{"12a", false},
		{"abc", false},
	}
	for _, tt := range tests {
		if got := isNumeric(tt.input); got != tt.want {
			t.Errorf("isNumeric(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestResolve_BranchSetWhenFromBranch(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "specs", "my-feature"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(dir, "", "my-feature")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if got.Branch != "my-feature" {
		t.Errorf("Branch = %q, want %q", got.Branch, "my-feature")
	}
	if got.Explicit {
		t.Error("Explicit should be false when resolved from branch")
	}
}
