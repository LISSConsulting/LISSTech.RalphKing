// Package notify sends fire-and-forget HTTP notifications for loop events.
// The primary use case is ntfy.sh, but any HTTP webhook works.
package notify

import (
	"net/http"
	"strings"
	"time"

	"github.com/LISSConsulting/LISSTech.RalphKing/internal/loop"
)

// Notifier posts plain-text HTTP notifications for selected loop events.
type Notifier struct {
	url        string
	title      string
	onComplete bool
	onError    bool
	onStop     bool
	client     *http.Client
}

// New creates a Notifier. projectName is used as the X-Title header; if empty,
// "RalphKing" is used instead.
func New(notifURL, projectName string, onComplete, onError, onStop bool) *Notifier {
	title := "RalphKing"
	if projectName != "" {
		title = projectName
	}
	return &Notifier{
		url:        notifURL,
		title:      title,
		onComplete: onComplete,
		onError:    onError,
		onStop:     onStop,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Hook is a loop.Loop.NotificationHook-compatible function. It fires
// asynchronous POSTs for events that match the configured notification flags.
func (n *Notifier) Hook(entry loop.LogEntry) {
	switch entry.Kind {
	case loop.LogIterComplete:
		if n.onComplete {
			go n.post(entry.Message)
		}
	case loop.LogError:
		if n.onError {
			go n.post(entry.Message)
		}
	case loop.LogDone, loop.LogStopped:
		if n.onStop {
			go n.post(entry.Message)
		}
	}
}

// post sends a plain-text POST to the configured URL. Errors are silently
// discarded so notification failures never interrupt the loop.
func (n *Notifier) post(message string) {
	req, err := http.NewRequest(http.MethodPost, n.url, strings.NewReader(message))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Title", n.title)
	resp, err := n.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
