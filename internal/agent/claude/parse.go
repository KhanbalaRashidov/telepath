package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/telepath/telepath/internal/agent"
)

// Parser converts Claude Code `--output-format stream-json` lines into the
// normalized agent.Event stream. It is stateful (captures the session id) and
// is unit-tested in isolation.
type Parser struct {
	sessionID string
}

// NewParser returns a ready parser.
func NewParser() *Parser { return &Parser{} }

// SessionID returns the session id captured so far ("" until the init event).
func (p *Parser) SessionID() string { return p.sessionID }

// streamLine is the envelope shared by every stream-json record.
type streamLine struct {
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype"`
	SessionID string          `json:"session_id"`
	Message   json.RawMessage `json:"message"`
	Event     json.RawMessage `json:"event"`
	// result record fields
	Result     string   `json:"result"`
	IsError    bool     `json:"is_error"`
	TotalCost  *float64 `json:"total_cost_usd"`
	DurationMS int64    `json:"duration_ms"`
	NumTurns   int      `json:"num_turns"`
}

type apiMessage struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

// Parse maps one line to zero or more events. Unparseable lines yield nothing.
func (p *Parser) Parse(line string) []agent.Event {
	line = strings.TrimSpace(line)
	if line == "" || line[0] != '{' {
		return nil
	}
	var sl streamLine
	if err := json.Unmarshal([]byte(line), &sl); err != nil {
		return nil
	}
	if sl.SessionID != "" {
		p.sessionID = sl.SessionID
	}

	switch sl.Type {
	case "system":
		// init carries session_id (already captured); nothing to surface.
		return nil
	case "assistant":
		return p.parseAssistant(sl.Message)
	case "user":
		return p.parseUser(sl.Message)
	case "stream_event":
		return p.parsePartial(sl.Event)
	case "result":
		return []agent.Event{p.parseResult(sl)}
	default:
		return nil
	}
}

func (p *Parser) parseAssistant(raw json.RawMessage) []agent.Event {
	var msg apiMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil
	}
	var events []agent.Event
	for _, b := range msg.Content {
		switch b.Type {
		case "text":
			if t := strings.TrimSpace(b.Text); t != "" {
				events = append(events, agent.AssistantText{Text: t})
			}
		case "tool_use":
			events = append(events, toolUseEvent(b.Name, b.Input)...)
		}
	}
	return events
}

func (p *Parser) parseUser(raw json.RawMessage) []agent.Event {
	var msg apiMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil
	}
	var events []agent.Event
	for _, b := range msg.Content {
		if b.Type != "tool_result" {
			continue
		}
		summary := flattenToolResult(b.Content)
		events = append(events, agent.ToolResult{OK: !b.IsError, Summary: truncate(summary, 300)})
	}
	return events
}

func (p *Parser) parsePartial(raw json.RawMessage) []agent.Event {
	var ev struct {
		Type  string `json:"type"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(raw, &ev); err != nil {
		return nil
	}
	if ev.Type == "content_block_delta" && ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
		return []agent.Event{agent.Progress{PartialText: ev.Delta.Text}}
	}
	return nil
}

func (p *Parser) parseResult(sl streamLine) agent.Event {
	if sl.IsError {
		msg := sl.Result
		if msg == "" {
			msg = "agent reported an error"
		}
		return agent.Error{Err: fmt.Errorf("%s", msg)}
	}
	return agent.Result{
		Summary:   strings.TrimSpace(sl.Result),
		Duration:  durationFromMS(sl.DurationMS),
		CostUSD:   sl.TotalCost,
		SessionID: p.sessionID,
	}
}

// toolUseEvent maps a tool_use block to the right normalized event(s).
func toolUseEvent(name string, input json.RawMessage) []agent.Event {
	switch name {
	case "Edit", "Write", "MultiEdit":
		path, diff := buildFileEdit(name, input)
		if path == "" {
			return []agent.Event{agent.ToolUse{Name: name, InputSummary: summarizeInput(input)}}
		}
		return []agent.Event{agent.FileEdit{Path: path, UnifiedDiff: diff}}
	case "Bash":
		var in struct {
			Command string `json:"command"`
		}
		_ = json.Unmarshal(input, &in)
		return []agent.Event{agent.Command{Cmd: in.Command}}
	default:
		return []agent.Event{agent.ToolUse{Name: name, InputSummary: summarizeInput(input)}}
	}
}

// buildFileEdit extracts the path and builds a best-effort unified diff from the
// edit payload (we do not have the full file, so line numbers are nominal).
func buildFileEdit(name string, input json.RawMessage) (path, diff string) {
	switch name {
	case "Write":
		var in struct {
			FilePath string `json:"file_path"`
			Content  string `json:"content"`
		}
		_ = json.Unmarshal(input, &in)
		return in.FilePath, diffBlock(in.FilePath, "", in.Content, true)
	case "Edit":
		var in struct {
			FilePath  string `json:"file_path"`
			OldString string `json:"old_string"`
			NewString string `json:"new_string"`
		}
		_ = json.Unmarshal(input, &in)
		return in.FilePath, diffBlock(in.FilePath, in.OldString, in.NewString, false)
	case "MultiEdit":
		var in struct {
			FilePath string `json:"file_path"`
			Edits    []struct {
				OldString string `json:"old_string"`
				NewString string `json:"new_string"`
			} `json:"edits"`
		}
		_ = json.Unmarshal(input, &in)
		var b strings.Builder
		fmt.Fprintf(&b, "--- a/%s\n+++ b/%s\n", in.FilePath, in.FilePath)
		for i, e := range in.Edits {
			fmt.Fprintf(&b, "@@ edit %d @@\n", i+1)
			b.WriteString(diffBody(e.OldString, e.NewString))
		}
		return in.FilePath, b.String()
	}
	return "", ""
}

// diffBlock builds a full unified-diff-style block with headers.
func diffBlock(path, oldStr, newStr string, isNew bool) string {
	var b strings.Builder
	if isNew {
		fmt.Fprintf(&b, "--- /dev/null\n+++ b/%s\n@@ new file @@\n", path)
	} else {
		fmt.Fprintf(&b, "--- a/%s\n+++ b/%s\n@@ edit @@\n", path, path)
	}
	b.WriteString(diffBody(oldStr, newStr))
	return b.String()
}

// diffBody renders removed lines with '-' and added lines with '+'.
func diffBody(oldStr, newStr string) string {
	var b strings.Builder
	if oldStr != "" {
		for _, l := range strings.Split(strings.TrimRight(oldStr, "\n"), "\n") {
			b.WriteString("-" + l + "\n")
		}
	}
	if newStr != "" {
		for _, l := range strings.Split(strings.TrimRight(newStr, "\n"), "\n") {
			b.WriteString("+" + l + "\n")
		}
	}
	return b.String()
}

// summarizeInput produces a short human summary from a tool input object.
func summarizeInput(raw json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return ""
	}
	for _, key := range []string{"file_path", "path", "pattern", "query", "url", "command", "prompt"} {
		if v, ok := m[key]; ok {
			return fmt.Sprintf("%s=%v", key, truncate(fmt.Sprint(v), 120))
		}
	}
	return truncate(fmt.Sprint(m), 120)
}

// flattenToolResult turns tool_result content (string OR array of blocks) into
// a plain string.
func flattenToolResult(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) == nil {
		var parts []string
		for _, b := range blocks {
			if b.Text != "" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return string(raw)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
