package claude

import (
	"strings"
	"testing"
)

func TestParseStream(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		events []struct {
			typ      EventType
			toolName string
			text     string
			costUSD  float64
			errMsg   string
		}
	}{
		{
			name: "tool_use event",
			input: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"write_file","input":{"path":"main.go"}}]}}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventToolUse, toolName: "write_file"},
			},
		},
		{
			name: "text event",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventText, text: "Hello world"},
			},
		},
		{
			name: "result event",
			input: `{"type":"result","cost_usd":0.14,"duration_ms":4200}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventResult, costUSD: 0.14},
			},
		},
		{
			name: "error event",
			input: `{"type":"system","subtype":"error","error":"something broke"}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventError, errMsg: "something broke"},
			},
		},
		{
			name: "multiple content blocks",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"analyzing"},{"type":"tool_use","name":"bash","input":{"command":"go test"}}]}}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventText, text: "analyzing"},
				{typ: EventToolUse, toolName: "bash"},
			},
		},
		{
			name:  "empty text block ignored",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":""}]}}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{},
		},
		{
			name:  "invalid json ignored",
			input: `not valid json`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{},
		},
		{
			name:  "empty lines ignored",
			input: "\n\n",
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{},
		},
		{
			name:  "system non-error ignored",
			input: `{"type":"system","subtype":"init"}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{},
		},
		{
			name: "multi-line stream",
			input: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"read_file","input":{"path":"go.mod"}}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"write_file","input":{"path":"main.go"}}]}}
{"type":"result","cost_usd":0.05,"duration_ms":2000}`,
			events: []struct {
				typ      EventType
				toolName string
				text     string
				costUSD  float64
				errMsg   string
			}{
				{typ: EventToolUse, toolName: "read_file"},
				{typ: EventToolUse, toolName: "write_file"},
				{typ: EventResult, costUSD: 0.05},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := ParseStream(strings.NewReader(tt.input))

			var got []Event
			for ev := range ch {
				got = append(got, ev)
			}

			if len(got) != len(tt.events) {
				t.Fatalf("got %d events, want %d", len(got), len(tt.events))
			}

			for i, want := range tt.events {
				ev := got[i]
				if ev.Type != want.typ {
					t.Errorf("event[%d].Type = %q, want %q", i, ev.Type, want.typ)
				}
				if want.toolName != "" && ev.ToolName != want.toolName {
					t.Errorf("event[%d].ToolName = %q, want %q", i, ev.ToolName, want.toolName)
				}
				if want.text != "" && ev.Text != want.text {
					t.Errorf("event[%d].Text = %q, want %q", i, ev.Text, want.text)
				}
				if want.costUSD != 0 && ev.CostUSD != want.costUSD {
					t.Errorf("event[%d].CostUSD = %f, want %f", i, ev.CostUSD, want.costUSD)
				}
				if want.errMsg != "" && ev.Error != want.errMsg {
					t.Errorf("event[%d].Error = %q, want %q", i, ev.Error, want.errMsg)
				}
			}
		})
	}
}
