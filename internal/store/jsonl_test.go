package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
	"github.com/LISSConsulting/LISSTech.RalphKing/internal/store"
)

// Compile-time check: *JSONL implements Store.
var _ store.Store = (*store.JSONL)(nil)

func TestNewJSONL_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatalf("NewJSONL: %v", err)
	}
	defer s.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file in dir, got %d", len(entries))
	}
	if ext := filepath.Ext(entries[0].Name()); ext != ".jsonl" {
		t.Errorf("expected .jsonl extension, got %q", ext)
	}
}

func TestNewJSONL_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "logs")
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatalf("NewJSONL on non-existent dir: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir to exist after NewJSONL: %v", err)
	}
}

func TestAppendAndIterationLog(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now()
	entries := []loop.LogEntry{
		{Kind: loop.LogIterStart, Timestamp: now, Iteration: 1, Mode: "build", Branch: "main"},
		{Kind: loop.LogToolUse, Timestamp: now, Iteration: 1, ToolName: "Read", ToolInput: "main.go"},
		{Kind: loop.LogIterComplete, Timestamp: now, Iteration: 1, CostUSD: 0.05, Duration: 1.2, Subtype: "success"},
	}
	for _, e := range entries {
		if err := s.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := s.IterationLog(1)
	if err != nil {
		t.Fatalf("IterationLog(1): %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Kind != loop.LogIterStart {
		t.Errorf("got[0].Kind: expected LogIterStart, got %v", got[0].Kind)
	}
	if got[1].Kind != loop.LogToolUse {
		t.Errorf("got[1].Kind: expected LogToolUse, got %v", got[1].Kind)
	}
	if got[1].ToolName != "Read" {
		t.Errorf("got[1].ToolName: expected %q, got %q", "Read", got[1].ToolName)
	}
	if got[2].Kind != loop.LogIterComplete {
		t.Errorf("got[2].Kind: expected LogIterComplete, got %v", got[2].Kind)
	}
	if got[2].CostUSD != 0.05 {
		t.Errorf("got[2].CostUSD: expected 0.05, got %v", got[2].CostUSD)
	}
}

func TestIterations(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now()
	for i := 1; i <= 3; i++ {
		_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: i, Mode: "build", Branch: "main", Timestamp: now})
		_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: i, CostUSD: float64(i) * 0.01, Duration: float64(i), Subtype: "success", Timestamp: now})
	}

	iters, err := s.Iterations()
	if err != nil {
		t.Fatal(err)
	}
	if len(iters) != 3 {
		t.Fatalf("expected 3 iterations, got %d", len(iters))
	}
	for i, it := range iters {
		wantNum := i + 1
		if it.Number != wantNum {
			t.Errorf("iters[%d].Number: expected %d, got %d", i, wantNum, it.Number)
		}
		if it.Mode != "build" {
			t.Errorf("iters[%d].Mode: expected build, got %q", i, it.Mode)
		}
		wantCost := float64(wantNum) * 0.01
		if it.CostUSD != wantCost {
			t.Errorf("iters[%d].CostUSD: expected %.2f, got %.2f", i, wantCost, it.CostUSD)
		}
		if it.Subtype != "success" {
			t.Errorf("iters[%d].Subtype: expected success, got %q", i, it.Subtype)
		}
	}
}

func TestIterationsReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now()
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: 1, Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: 1, CostUSD: 0.05, Timestamp: now})

	first, _ := s.Iterations()
	first[0].CostUSD = 999.0 // mutate the returned copy

	second, _ := s.Iterations()
	if second[0].CostUSD == 999.0 {
		t.Error("Iterations should return a copy; mutation affected internal state")
	}
}

func TestSessionSummary(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now()
	_ = s.Append(loop.LogEntry{Kind: loop.LogInfo, Branch: "feat/test", Commit: "abc123", Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: 1, Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: 1, CostUSD: 1.0, Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: 2, Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: 2, CostUSD: 2.0, Timestamp: now})

	sum, err := s.SessionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Iterations != 2 {
		t.Errorf("Iterations: expected 2, got %d", sum.Iterations)
	}
	if sum.TotalCost != 3.0 {
		t.Errorf("TotalCost: expected 3.0, got %v", sum.TotalCost)
	}
	if sum.Branch != "feat/test" {
		t.Errorf("Branch: expected %q, got %q", "feat/test", sum.Branch)
	}
	if sum.LastCommit != "abc123" {
		t.Errorf("LastCommit: expected %q, got %q", "abc123", sum.LastCommit)
	}
	if sum.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if sum.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}
}

func TestSessionSummary_Empty(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	sum, err := s.SessionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Iterations != 0 {
		t.Errorf("expected 0 iterations, got %d", sum.Iterations)
	}
	if sum.TotalCost != 0 {
		t.Errorf("expected 0 total cost, got %.2f", sum.TotalCost)
	}
	if sum.SessionID == "" {
		t.Error("SessionID should not be empty even with no entries")
	}
}

func TestIterationLog_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.IterationLog(42)
	if err == nil {
		t.Fatal("expected error for nonexistent iteration")
	}
}

func TestIterationLog_InProgress(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Write LogIterStart but not LogIterComplete
	now := time.Now()
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: 1, Timestamp: now})
	_ = s.Append(loop.LogEntry{Kind: loop.LogToolUse, Iteration: 1, ToolName: "Read", Timestamp: now})

	// Iteration 1 is in progress — not in the completed index
	_, err = s.IterationLog(1)
	if err == nil {
		t.Fatal("expected error for in-progress (incomplete) iteration")
	}
}

func TestIterationLog_EntriesBetweenIterations(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Write entries interleaved with info messages outside iterations
	now := time.Now()
	all := []loop.LogEntry{
		{Kind: loop.LogInfo, Message: "session started", Timestamp: now},
		{Kind: loop.LogIterStart, Iteration: 1, Mode: "build", Timestamp: now},
		{Kind: loop.LogToolUse, Iteration: 1, ToolName: "Read", Timestamp: now},
		{Kind: loop.LogGitPull, Iteration: 1, Message: "pulled", Timestamp: now},
		{Kind: loop.LogIterComplete, Iteration: 1, CostUSD: 0.10, Subtype: "success", Timestamp: now},
		{Kind: loop.LogInfo, Message: "between iterations", Timestamp: now},
		{Kind: loop.LogIterStart, Iteration: 2, Mode: "build", Timestamp: now},
		{Kind: loop.LogToolUse, Iteration: 2, ToolName: "Write", Timestamp: now},
		{Kind: loop.LogIterComplete, Iteration: 2, CostUSD: 0.20, Subtype: "success", Timestamp: now},
	}
	for _, e := range all {
		if err := s.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	log1, err := s.IterationLog(1)
	if err != nil {
		t.Fatalf("IterationLog(1): %v", err)
	}
	// Should contain LogIterStart + LogToolUse + LogGitPull + LogIterComplete
	if len(log1) != 4 {
		t.Errorf("iter 1: expected 4 entries, got %d", len(log1))
	}

	log2, err := s.IterationLog(2)
	if err != nil {
		t.Fatalf("IterationLog(2): %v", err)
	}
	// Should contain LogIterStart + LogToolUse + LogIterComplete
	if len(log2) != 3 {
		t.Errorf("iter 2: expected 3 entries, got %d", len(log2))
	}

	// The LogInfo entries between/outside iterations must not appear in either log
	for i, e := range log1 {
		if e.Kind == loop.LogInfo {
			t.Errorf("iter 1 log1[%d] is LogInfo (should not be in iteration range)", i)
		}
	}
	for i, e := range log2 {
		if e.Kind == loop.LogInfo {
			t.Errorf("iter 2 log2[%d] is LogInfo (should not be in iteration range)", i)
		}
	}
}

func TestIterations_Empty(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	iters, err := s.Iterations()
	if err != nil {
		t.Fatal(err)
	}
	if len(iters) != 0 {
		t.Fatalf("expected 0 iterations, got %d", len(iters))
	}
}

func TestNewJSONL_DirIsFile(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "notadir")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	// "file" exists as a regular file; MkdirAll should fail.
	_, err := store.NewJSONL(file)
	if err == nil {
		t.Fatal("expected error when dir argument is an existing file")
	}
}

func TestAppend_AfterClose(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	err = s.Append(loop.LogEntry{Kind: loop.LogInfo, Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error when appending to a closed store")
	}
}

func TestIterationLog_CompleteWithoutStart(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Write LogIterComplete without a preceding LogIterStart.
	// The index should ignore this entry (no pending iteration).
	now := time.Now()
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: 5, CostUSD: 0.01, Timestamp: now})

	_, err = s.IterationLog(5)
	if err == nil {
		t.Fatal("expected error: iteration 5 was never started")
	}

	iters, _ := s.Iterations()
	if len(iters) != 0 {
		t.Errorf("expected 0 completed iterations, got %d", len(iters))
	}
}

func TestAppend_RoundTripsAllFields(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().Truncate(time.Millisecond) // JSON uses millisecond precision
	orig := loop.LogEntry{
		Kind:      loop.LogToolUse,
		Timestamp: now,
		Message:   "tool message",
		ToolName:  "Bash",
		ToolInput: "echo hello",
		CostUSD:   0.001,
		Duration:  0.5,
		TotalCost: 1.23,
		Subtype:   "success",
		Iteration: 3,
		MaxIter:   10,
		Branch:    "feature",
		Commit:    "deadbeef",
		Mode:      "build",
	}
	// Write without an iteration wrapper to test raw field round-trip
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterStart, Iteration: 3, Timestamp: now})
	if err := s.Append(orig); err != nil {
		t.Fatalf("Append: %v", err)
	}
	_ = s.Append(loop.LogEntry{Kind: loop.LogIterComplete, Iteration: 3, Timestamp: now})

	got, err := s.IterationLog(3)
	if err != nil {
		t.Fatal(err)
	}
	// got[0] = LogIterStart, got[1] = our entry, got[2] = LogIterComplete
	if len(got) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(got))
	}
	e := got[1]
	if e.Kind != orig.Kind {
		t.Errorf("Kind: want %v, got %v", orig.Kind, e.Kind)
	}
	if !e.Timestamp.Equal(orig.Timestamp) {
		t.Errorf("Timestamp: want %v, got %v", orig.Timestamp, e.Timestamp)
	}
	if e.ToolName != orig.ToolName {
		t.Errorf("ToolName: want %q, got %q", orig.ToolName, e.ToolName)
	}
	if e.ToolInput != orig.ToolInput {
		t.Errorf("ToolInput: want %q, got %q", orig.ToolInput, e.ToolInput)
	}
	if e.Branch != orig.Branch {
		t.Errorf("Branch: want %q, got %q", orig.Branch, e.Branch)
	}
	if e.Commit != orig.Commit {
		t.Errorf("Commit: want %q, got %q", orig.Commit, e.Commit)
	}
	if e.Mode != orig.Mode {
		t.Errorf("Mode: want %q, got %q", orig.Mode, e.Mode)
	}
}
