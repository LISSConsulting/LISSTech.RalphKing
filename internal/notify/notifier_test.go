package notify

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// captureServer starts an httptest.Server that records incoming requests.
// It returns the server and a function to collect all captured requests.
func captureServer(t *testing.T) (*httptest.Server, func() []capturedReq) {
	t.Helper()
	var mu sync.Mutex
	var reqs []capturedReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		reqs = append(reqs, capturedReq{
			method:      r.Method,
			body:        string(body),
			contentType: r.Header.Get("Content-Type"),
			title:       r.Header.Get("X-Title"),
		})
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, func() []capturedReq {
		mu.Lock()
		defer mu.Unlock()
		out := make([]capturedReq, len(reqs))
		copy(out, reqs)
		return out
	}
}

type capturedReq struct {
	method      string
	body        string
	contentType string
	title       string
}

// waitForRequests polls until count requests are captured or the deadline is reached.
func waitForRequests(t *testing.T, collect func() []capturedReq, count int) []capturedReq {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := collect(); len(got) >= count {
			return got
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d request(s)", count)
	return nil
}

func TestHook_OnComplete(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "myapp", true, false, false)
	n.Hook(loop.LogEntry{Kind: loop.LogIterComplete, Message: "Iteration 1 complete"})

	reqs := waitForRequests(t, collect, 1)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	r := reqs[0]
	if r.method != http.MethodPost {
		t.Errorf("method = %q, want POST", r.method)
	}
	if r.body != "Iteration 1 complete" {
		t.Errorf("body = %q, want %q", r.body, "Iteration 1 complete")
	}
	if r.contentType != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", r.contentType)
	}
	if r.title != "myapp" {
		t.Errorf("X-Title = %q, want myapp", r.title)
	}
}

func TestHook_OnComplete_Disabled(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", false, false, false)
	n.Hook(loop.LogEntry{Kind: loop.LogIterComplete, Message: "Iteration 1 complete"})

	// Give the goroutine time to fire (it shouldn't, but we need to be sure).
	time.Sleep(50 * time.Millisecond)
	if got := collect(); len(got) != 0 {
		t.Errorf("expected no requests, got %d", len(got))
	}
}

func TestHook_OnError(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "proj", false, true, false)
	n.Hook(loop.LogEntry{Kind: loop.LogError, Message: "Error: something failed"})

	reqs := waitForRequests(t, collect, 1)
	if reqs[0].body != "Error: something failed" {
		t.Errorf("body = %q, want %q", reqs[0].body, "Error: something failed")
	}
}

func TestHook_OnError_Disabled(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", false, false, false)
	n.Hook(loop.LogEntry{Kind: loop.LogError, Message: "oops"})

	time.Sleep(50 * time.Millisecond)
	if got := collect(); len(got) != 0 {
		t.Errorf("expected no requests, got %d", len(got))
	}
}

func TestHook_OnStop_LogDone(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", false, false, true)
	n.Hook(loop.LogEntry{Kind: loop.LogDone, Message: "Loop complete"})

	reqs := waitForRequests(t, collect, 1)
	if reqs[0].body != "Loop complete" {
		t.Errorf("body = %q, want %q", reqs[0].body, "Loop complete")
	}
}

func TestHook_OnStop_LogStopped(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", false, false, true)
	n.Hook(loop.LogEntry{Kind: loop.LogStopped, Message: "Stop requested"})

	reqs := waitForRequests(t, collect, 1)
	if reqs[0].body != "Stop requested" {
		t.Errorf("body = %q, want %q", reqs[0].body, "Stop requested")
	}
}

func TestHook_OnStop_Disabled(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", false, false, false)
	n.Hook(loop.LogEntry{Kind: loop.LogDone, Message: "done"})
	n.Hook(loop.LogEntry{Kind: loop.LogStopped, Message: "stopped"})

	time.Sleep(50 * time.Millisecond)
	if got := collect(); len(got) != 0 {
		t.Errorf("expected no requests, got %d", len(got))
	}
}

func TestHook_IgnoresOtherKinds(t *testing.T) {
	srv, collect := captureServer(t)

	n := New(srv.URL, "", true, true, true)
	// These kinds should never trigger a notification.
	for _, kind := range []loop.LogKind{loop.LogInfo, loop.LogIterStart, loop.LogToolUse, loop.LogText, loop.LogGitPull, loop.LogGitPush, loop.LogRegent} {
		n.Hook(loop.LogEntry{Kind: kind, Message: "noise"})
	}

	time.Sleep(50 * time.Millisecond)
	if got := collect(); len(got) != 0 {
		t.Errorf("expected no requests for non-notification kinds, got %d", len(got))
	}
}

func TestHook_FallbackTitle(t *testing.T) {
	srv, collect := captureServer(t)

	// Empty project name → fallback title "RalphKing"
	n := New(srv.URL, "", true, false, false)
	n.Hook(loop.LogEntry{Kind: loop.LogIterComplete, Message: "done"})

	reqs := waitForRequests(t, collect, 1)
	if reqs[0].title != "RalphKing" {
		t.Errorf("X-Title = %q, want RalphKing", reqs[0].title)
	}
}

func TestHook_PostFailureSilent(t *testing.T) {
	// Point at a server that is already closed → connection refused.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately

	n := New(srv.URL, "", true, true, true)
	// None of these should panic or block.
	n.Hook(loop.LogEntry{Kind: loop.LogIterComplete, Message: "done"})
	n.Hook(loop.LogEntry{Kind: loop.LogError, Message: "err"})
	n.Hook(loop.LogEntry{Kind: loop.LogDone, Message: "done"})

	// Allow goroutines to finish.
	time.Sleep(100 * time.Millisecond)
}
