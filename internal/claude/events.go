// Package claude provides the Claude CLI adapter and stream-JSON event parser.
package claude

import "time"

// EventType identifies the kind of stream-JSON event.
type EventType string

const (
	EventToolUse EventType = "tool_use"
	EventText    EventType = "text"
	EventResult  EventType = "result"
	EventError   EventType = "error"
)

// Event is a parsed stream-JSON event from Claude CLI output.
type Event struct {
	Type      EventType
	Timestamp time.Time

	// ToolUse fields
	ToolName  string
	ToolInput map[string]any

	// Text fields
	Text string

	// Result fields
	CostUSD  float64
	Duration float64 // seconds

	// Error fields
	Error string
}

// ToolUseEvent creates a tool_use event.
func ToolUseEvent(name string, input map[string]any) Event {
	return Event{
		Type:      EventToolUse,
		Timestamp: time.Now(),
		ToolName:  name,
		ToolInput: input,
	}
}

// TextEvent creates a text event.
func TextEvent(text string) Event {
	return Event{
		Type:      EventText,
		Timestamp: time.Now(),
		Text:      text,
	}
}

// ResultEvent creates a result event with cost and duration.
func ResultEvent(costUSD, duration float64) Event {
	return Event{
		Type:      EventResult,
		Timestamp: time.Now(),
		CostUSD:   costUSD,
		Duration:  duration,
	}
}

// ErrorEvent creates an error event.
func ErrorEvent(msg string) Event {
	return Event{
		Type:      EventError,
		Timestamp: time.Now(),
		Error:     msg,
	}
}
