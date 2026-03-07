package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// init runs as a fake `wt` subprocess when _FAKE_WT=1 is set.
// Placing the guard in init() (before flag.Parse) avoids flag-parse failures
// from unrecognised wt flags like --json.
func init() {
	if os.Getenv("_FAKE_WT") != "1" {
		return
	}

	args := os.Args[1:]

	// Provide different responses based on what was requested.
	stdout := os.Getenv("_FAKE_WT_STDOUT")
	exitCode := 0
	if s := os.Getenv("_FAKE_WT_EXIT"); s != "" {
		_, _ = fmt.Sscan(s, &exitCode)
	}

	// Handle --version specially: confirm we are worktrunk.
	if len(args) == 1 && args[0] == "--version" {
		if os.Getenv("_FAKE_WT_NOT_WORKTRUNK") == "1" {
			fmt.Println("Windows Terminal v1.20.0")
		} else {
			fmt.Println("worktrunk v0.3.1")
		}
		os.Exit(0)
	}

	if stdout != "" {
		fmt.Print(stdout)
	}

	// Emit stderr if requested.
	if s := os.Getenv("_FAKE_WT_STDERR"); s != "" {
		_, _ = fmt.Fprint(os.Stderr, s)
	}

	os.Exit(exitCode)
}

// fakeRunner returns a Runner whose executable points at the current test
// binary (which acts as a fake wt via the init() guard above).
func fakeRunner(t *testing.T, dir string) *Runner {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	r := NewRunner(dir)
	r.executable = exe
	return r
}

// withEnv sets environment variables for the duration of a subtest and
// arranges cleanup via t.Cleanup.
func withEnv(t *testing.T, pairs ...string) {
	t.Helper()
	for i := 0; i+1 < len(pairs); i += 2 {
		t.Setenv(pairs[i], pairs[i+1])
	}
}

// ─── Detect tests ────────────────────────────────────────────────────────────

func TestDetect_Found(t *testing.T) {
	dir := t.TempDir()

	var wtBin string
	if runtime.GOOS == "windows" {
		// On Windows, use a .bat file so we avoid binary-locking issues with
		// copied executables. A batch file that echoes "worktrunk" and exits 0
		// satisfies the Detect() version check.
		wtBin = filepath.Join(dir, "wt.bat")
		script := "@echo off\r\necho worktrunk v0.3.1-test\r\n"
		if err := os.WriteFile(wtBin, []byte(script), 0644); err != nil {
			t.Fatalf("write wt.bat: %v", err)
		}
	} else {
		exe, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable: %v", err)
		}
		wtBin = filepath.Join(dir, "wt")
		if err := os.Link(exe, wtBin); err != nil {
			data, readErr := os.ReadFile(exe)
			if readErr != nil {
				t.Fatalf("copy test binary: %v", readErr)
			}
			if writeErr := os.WriteFile(wtBin, data, 0755); writeErr != nil {
				t.Fatalf("copy test binary: %v", writeErr)
			}
		}
		_ = os.Chmod(wtBin, 0755)
		t.Setenv("_FAKE_WT", "1")
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	r := NewRunner(dir)
	if err := r.Detect(); err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}
	if r.executable == "" {
		t.Error("Detect() did not set executable")
	}
}

func TestDetect_NotFound(t *testing.T) {
	t.Setenv("PATH", "")
	r := NewRunner(t.TempDir())
	err := r.Detect()
	if err == nil {
		t.Fatal("expected error when wt not on PATH")
	}
	if !strings.Contains(err.Error(), "worktrunk not found") {
		t.Errorf("error should mention 'worktrunk not found', got: %v", err)
	}
}

func TestDetect_NotWorktrunk(t *testing.T) {
	dir := t.TempDir()

	var wtBin string
	if runtime.GOOS == "windows" {
		// Use a batch script that pretends to be Windows Terminal, not worktrunk.
		wtBin = filepath.Join(dir, "wt.bat")
		script := "@echo off\r\necho Windows Terminal v1.20.0\r\n"
		if err := os.WriteFile(wtBin, []byte(script), 0644); err != nil {
			t.Fatalf("write wt.bat: %v", err)
		}
	} else {
		exe, _ := os.Executable()
		wtBin = filepath.Join(dir, "wt")
		data, _ := os.ReadFile(exe)
		_ = os.WriteFile(wtBin, data, 0755)
		t.Setenv("_FAKE_WT", "1")
		t.Setenv("_FAKE_WT_NOT_WORKTRUNK", "1")
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	r := NewRunner(dir)
	err := r.Detect()
	if err == nil {
		t.Fatal("expected error when wt binary is not worktrunk")
	}
	if !strings.Contains(err.Error(), "worktrunk not found") {
		t.Errorf("error should mention 'worktrunk not found', got: %v", err)
	}
}

// ─── Switch tests ─────────────────────────────────────────────────────────────

func TestSwitch_CreateNew(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1",
		"_FAKE_WT_STDOUT", "✓ Created branch feat/new from main and worktree @ /tmp/worktrees/feat-new\n")

	r := fakeRunner(t, dir)
	path, err := r.Switch("feat/new", true)
	if err != nil {
		t.Fatalf("Switch create: %v", err)
	}
	if path != "/tmp/worktrees/feat-new" {
		t.Errorf("path: got %q, want %q", path, "/tmp/worktrees/feat-new")
	}
}

func TestSwitch_ReuseExisting(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1",
		"_FAKE_WT_STDOUT", "Switched to existing worktree @ /tmp/worktrees/feat-old\n")

	r := fakeRunner(t, dir)
	path, err := r.Switch("feat/old", false)
	if err != nil {
		t.Fatalf("Switch reuse: %v", err)
	}
	if path != "/tmp/worktrees/feat-old" {
		t.Errorf("path: got %q, want %q", path, "/tmp/worktrees/feat-old")
	}
}

func TestSwitch_Error(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1",
		"_FAKE_WT_EXIT", "1",
		"_FAKE_WT_STDERR", "branch 'bad/branch' not found")

	r := fakeRunner(t, dir)
	_, err := r.Switch("bad/branch", false)
	if err == nil {
		t.Fatal("expected error from wt switch exit 1")
	}
	if !strings.Contains(err.Error(), "bad/branch") {
		t.Errorf("error should mention branch name, got: %v", err)
	}
}

// ─── List tests ───────────────────────────────────────────────────────────────

func TestList_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	jsonOut := `[{"branch":"main","path":"/repo","bare":false},{"branch":"feat/x","path":"/worktrees/feat-x","bare":false}]`
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", jsonOut)

	r := fakeRunner(t, dir)
	infos, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(infos))
	}
	if infos[0].Branch != "main" || infos[1].Branch != "feat/x" {
		t.Errorf("unexpected branches: %v", infos)
	}
}

func TestList_Empty(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", "[]")

	r := fakeRunner(t, dir)
	infos, err := r.List()
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 entries, got %d", len(infos))
	}
}

// ─── Merge tests ──────────────────────────────────────────────────────────────

func TestMerge_Success(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", "Merged feat/x into main\n")

	r := fakeRunner(t, dir)
	if err := r.Merge("feat/x", "main"); err != nil {
		t.Errorf("Merge: %v", err)
	}
}

func TestMerge_Conflict(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_EXIT", "1",
		"_FAKE_WT_STDOUT", "Conflict: cannot merge feat/conflict into main")

	r := fakeRunner(t, dir)
	err := r.Merge("feat/conflict", "main")
	if err == nil {
		t.Fatal("expected error from wt merge conflict")
	}
	if !strings.Contains(err.Error(), "feat/conflict") {
		t.Errorf("error should mention branch, got: %v", err)
	}
}

// ─── Remove tests ─────────────────────────────────────────────────────────────

func TestRemove_Success(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", "Removed worktree feat/done\n")

	r := fakeRunner(t, dir)
	if err := r.Remove("feat/done"); err != nil {
		t.Errorf("Remove: %v", err)
	}
}

func TestRemove_Error(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_EXIT", "1",
		"_FAKE_WT_STDOUT", "error: agent still running in worktree")

	r := fakeRunner(t, dir)
	err := r.Remove("feat/busy")
	if err == nil {
		t.Fatal("expected error from wt remove failure")
	}
}

// ─── parsePorcelain tests ─────────────────────────────────────────────────────

func TestParsePorcelain(t *testing.T) {
	input := `worktree /home/user/repo
HEAD abc123
branch refs/heads/main

worktree /home/user/.worktrees/feat-x
HEAD def456
branch refs/heads/feat/x

`
	infos := parsePorcelain(input)
	if len(infos) != 2 {
		t.Fatalf("expected 2, got %d", len(infos))
	}
	if infos[0].Path != "/home/user/repo" || infos[0].Branch != "main" {
		t.Errorf("first entry: %+v", infos[0])
	}
	if infos[1].Path != "/home/user/.worktrees/feat-x" || infos[1].Branch != "feat/x" {
		t.Errorf("second entry: %+v", infos[1])
	}
}

func TestParseWorktreePath(t *testing.T) {
	tests := []struct {
		output string
		want   string
	}{
		{"✓ Created branch foo from main and worktree @ /tmp/wt/foo\n", "/tmp/wt/foo"},
		{"Switched to existing worktree @ /home/user/.worktrees/bar\n", "/home/user/.worktrees/bar"},
		{"no path here\n", ""},
		{"@ \n", ""}, // empty after @
	}
	for _, tt := range tests {
		got := parseWorktreePath(tt.output)
		if got != tt.want {
			t.Errorf("parseWorktreePath(%q) = %q, want %q", tt.output, got, tt.want)
		}
	}
}

// ─── exe() fallback test ──────────────────────────────────────────────────────

func TestExeFallback(t *testing.T) {
	r := NewRunner(t.TempDir())
	// exe() without Detect() should return the first candidate binary name.
	got := r.exe()
	// We can also verify it's callable (lookup may fail — that's OK).
	if _, err := exec.LookPath(got); err != nil {
		// Binary not on PATH; just verify the name is one of our candidates.
		for _, name := range wtExecutables() {
			if got == name {
				return
			}
		}
		t.Errorf("exe() returned unexpected name %q", got)
	}
}
