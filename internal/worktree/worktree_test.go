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

	// Handle --version specially: confirm we are the wt CLI.
	if len(args) == 1 && args[0] == "--version" {
		if os.Getenv("_FAKE_WT_VERSION_FAIL") == "1" {
			os.Exit(1)
		}
		if os.Getenv("_FAKE_WT_NOT_WORKTRUNK") == "1" {
			fmt.Println("Windows Terminal v1.20.0")
		} else {
			fmt.Println("wt v0.28.2")
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
		script := "@echo off\r\necho wt v0.28.2-test\r\n"
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
	if !strings.Contains(err.Error(), "wt not found") {
		t.Errorf("error should mention 'wt not found', got: %v", err)
	}
}

func TestDetect_NotWorktrunk(t *testing.T) {
	dir := t.TempDir()

	var wtBin string
	if runtime.GOOS == "windows" {
		// Batch scripts that pretend to be Windows Terminal, not the wt CLI.
		script := "@echo off\r\necho Windows Terminal v1.20.0\r\n"
		for _, name := range []string{"wt.bat", "git-wt.bat"} {
			p := filepath.Join(dir, name)
			if err := os.WriteFile(p, []byte(script), 0644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		_ = wtBin
	} else {
		exe, _ := os.Executable()
		wtBin = filepath.Join(dir, "wt")
		data, _ := os.ReadFile(exe)
		_ = os.WriteFile(wtBin, data, 0755)
		t.Setenv("_FAKE_WT", "1")
		t.Setenv("_FAKE_WT_NOT_WORKTRUNK", "1")
	}

	// Use only the temp dir so the real git-wt on the system PATH is not found.
	t.Setenv("PATH", dir)

	r := NewRunner(dir)
	err := r.Detect()
	if err == nil {
		t.Fatal("expected error when wt binary is Windows Terminal")
	}
	if !strings.Contains(err.Error(), "wt not found") {
		t.Errorf("error should mention 'wt not found', got: %v", err)
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

func TestSwitch_NonExitError(t *testing.T) {
	dir := t.TempDir()
	r := NewRunner(dir)
	// Point to a non-existent binary so cmd.Output() fails with a path error,
	// not *exec.ExitError — covers the fmt.Errorf("wt switch %s: %w") branch.
	r.executable = filepath.Join(dir, "totally-nonexistent-wt-binary")

	_, err := r.Switch("feat/x", false)
	if err == nil {
		t.Fatal("expected error from non-existent executable")
	}
	if !strings.Contains(err.Error(), "feat/x") {
		t.Errorf("error should mention branch name, got: %v", err)
	}
}

func TestSwitch_NoPath_FallbackFails(t *testing.T) {
	// Switch succeeds (exit 0) but output has no "@ " path marker.
	// List() falls back to listPorcelain() which fails in a non-git directory,
	// so Switch returns a "could not determine worktree path" error.
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", "switch succeeded but no path marker")

	r := fakeRunner(t, dir)
	_, err := r.Switch("feat/x", false)
	if err == nil {
		t.Fatal("expected error when path cannot be determined")
	}
	if !strings.Contains(err.Error(), "could not determine worktree path") {
		t.Errorf("error should mention path determination, got: %v", err)
	}
}

func TestSwitch_NoPath_FallbackFindsViaList(t *testing.T) {
	// Switch succeeds (exit 0) but output has no "@ " path marker.
	// listJSON fails (invalid JSON), so List() falls back to listPorcelain()
	// which succeeds in a real git repo and finds the branch by name.
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Determine the actual branch name created by initGitRepo.
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = dir
	branchOut, gitErr := branchCmd.Output()
	if gitErr != nil {
		t.Fatalf("git branch --show-current: %v", gitErr)
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" {
		t.Skip("could not determine git branch name")
	}

	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_STDOUT", "switch succeeded but no path marker")

	r := fakeRunner(t, dir)
	path, err := r.Switch(branch, false)
	if err != nil {
		t.Fatalf("Switch fallback via porcelain: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path from porcelain fallback")
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

// ─── listPorcelain tests ──────────────────────────────────────────────────────

// initGitRepo sets up a minimal git repo in dir so that git commands work.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup %v: %v\n%s", args, err, out)
		}
	}
}

func TestListPorcelain_Success(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	r := NewRunner(dir)
	infos, err := r.listPorcelain()
	if err != nil {
		t.Fatalf("listPorcelain: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("expected at least one worktree from porcelain output")
	}
}

func TestListPorcelain_Error(t *testing.T) {
	// Not a git repo — git worktree list should fail.
	dir := t.TempDir()
	r := NewRunner(dir)
	_, err := r.listPorcelain()
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestList_FallbackToPorcelain(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	// Make wt list --json fail (exit 1) so List() falls back to git porcelain.
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_EXIT", "1")

	r := fakeRunner(t, dir)
	infos, err := r.List()
	if err != nil {
		t.Fatalf("List fallback: %v", err)
	}
	if len(infos) == 0 {
		t.Error("expected at least one worktree via porcelain fallback")
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

// ─── Additional coverage tests ────────────────────────────────────────────────

// TestDetect_VersionCmdFails covers the path in Detect() where exec.LookPath
// succeeds but the --version command itself exits non-zero.
func TestDetect_VersionCmdFails(t *testing.T) {
	dir := t.TempDir()

	var wtBin string
	if runtime.GOOS == "windows" {
		// Batch scripts that exit 1 unconditionally (fail on --version).
		for _, name := range []string{"wt.bat", "git-wt.bat"} {
			p := filepath.Join(dir, name)
			script := "@echo off\r\nexit /b 1\r\n"
			if err := os.WriteFile(p, []byte(script), 0644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		_ = wtBin
	} else {
		exe, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable: %v", err)
		}
		wtBin = filepath.Join(dir, "wt")
		data, readErr := os.ReadFile(exe)
		if readErr != nil {
			t.Fatalf("read test binary: %v", readErr)
		}
		if writeErr := os.WriteFile(wtBin, data, 0755); writeErr != nil {
			t.Fatalf("write fake wt: %v", writeErr)
		}
		t.Setenv("_FAKE_WT", "1")
		t.Setenv("_FAKE_WT_VERSION_FAIL", "1")
	}

	// Use only the temp dir so the real git-wt on the system PATH is not found.
	t.Setenv("PATH", dir)

	r := NewRunner(dir)
	err := r.Detect()
	if err == nil {
		t.Fatal("expected error when --version fails")
	}
	if !strings.Contains(err.Error(), "wt not found") {
		t.Errorf("error should mention 'wt not found', got: %v", err)
	}
}

// TestParsePorcelain_BareWorktree covers the "bare" line case in parsePorcelain.
func TestParsePorcelain_BareWorktree(t *testing.T) {
	input := `worktree /home/user/repo
HEAD abc123
bare

worktree /home/user/.worktrees/feat-x
HEAD def456
branch refs/heads/feat/x

`
	infos := parsePorcelain(input)
	if len(infos) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(infos))
	}
	if !infos[0].Bare {
		t.Errorf("first entry should be bare, got: %+v", infos[0])
	}
	if infos[1].Bare {
		t.Errorf("second entry should not be bare, got: %+v", infos[1])
	}
}

// TestRemove_ErrorNoOutput covers the msg=="" branch in Remove when the
// command fails but produces no output.
func TestRemove_ErrorNoOutput(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_EXIT", "1")
	// No _FAKE_WT_STDOUT set → empty combined output → msg == "".

	r := fakeRunner(t, dir)
	err := r.Remove("feat/silent-fail")
	if err == nil {
		t.Fatal("expected error from wt remove exit 1 with no output")
	}
	if !strings.Contains(err.Error(), "feat/silent-fail") {
		t.Errorf("error should mention branch name, got: %v", err)
	}
}

// TestMerge_ErrorNoOutput covers the msg=="" branch in Merge when the
// command fails but produces no output.
func TestMerge_ErrorNoOutput(t *testing.T) {
	dir := t.TempDir()
	withEnv(t, "_FAKE_WT", "1", "_FAKE_WT_EXIT", "1")
	// No _FAKE_WT_STDOUT set → empty combined output → msg == "".

	r := fakeRunner(t, dir)
	err := r.Merge("feat/silent-fail", "main")
	if err == nil {
		t.Fatal("expected error from wt merge exit 1 with no output")
	}
	if !strings.Contains(err.Error(), "feat/silent-fail") {
		t.Errorf("error should mention branch name, got: %v", err)
	}
}
