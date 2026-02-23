package claude

import (
	"bufio"
	"encoding/json"
	"io"
)

// ParseStream reads stream-JSON lines from r and sends parsed Events on the
// returned channel. The channel is closed when r reaches EOF or an error.
// This parses Claude CLI output from --output-format=stream-json --verbose.
func ParseStream(r io.Reader) <-chan Event {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		// Allow up to 1MB lines (Claude can produce large tool outputs)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			events := parseLine(line)
			for _, ev := range events {
				ch <- ev
			}
		}
	}()
	return ch
}

// streamMessage is the top-level JSON object in Claude's stream-json output.
type streamMessage struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype"`
	Message *messageContent `json:"message"`
	// Result fields (type=result)
	CostUSD  float64 `json:"cost_usd"`
	Duration float64 `json:"duration_ms"`
	// Error fields (type=system, subtype=error)
	Error string `json:"error"`
}

type messageContent struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type  string         `json:"type"`
	Text  string         `json:"text"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// parseLine parses a single line of stream-JSON output into zero or more Events.
func parseLine(line []byte) []Event {
	var msg streamMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil
	}

	switch msg.Type {
	case "assistant":
		return parseAssistantMessage(msg)
	case "result":
		return []Event{ResultEvent(msg.CostUSD, msg.Duration/1000)}
	case "system":
		if msg.Subtype == "error" {
			return []Event{ErrorEvent(msg.Error)}
		}
	}
	return nil
}

// parseAssistantMessage extracts tool_use and text events from an assistant message.
func parseAssistantMessage(msg streamMessage) []Event {
	if msg.Message == nil {
		return nil
	}

	var events []Event
	for _, block := range msg.Message.Content {
		switch block.Type {
		case "tool_use":
			events = append(events, ToolUseEvent(block.Name, block.Input))
		case "text":
			text := block.Text
			if text != "" {
				events = append(events, TextEvent(text))
			}
		}
	}
	return events
}
