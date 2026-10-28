package claude

import (
	"strings"
	"testing"

	"github.com/telepath/telepath/internal/agent"
)

func TestParserSessionAndEvents(t *testing.T) {
	p := NewParser()

	// system/init: captures session id, emits nothing.
	if evs := p.Parse(`{"type":"system","subtype":"init","session_id":"sess-abc","tools":["Read"]}`); len(evs) != 0 {
		t.Fatalf("system init should emit no events, got %d", len(evs))
	}
	if p.SessionID() != "sess-abc" {
		t.Fatalf("session id = %q, want sess-abc", p.SessionID())
	}

	// assistant: text + Write tool_use -> AssistantText + FileEdit.
	line := `{"type":"assistant","session_id":"sess-abc","message":{"content":[` +
		`{"type":"text","text":"Creating file"},` +
		`{"type":"tool_use","id":"t1","name":"Write","input":{"file_path":"/p/foo.go","content":"package main\n"}}]}}`
	evs := p.Parse(line)
	if len(evs) != 2 {
		t.Fatalf("assistant line: got %d events, want 2", len(evs))
	}
	if at, ok := evs[0].(agent.AssistantText); !ok || at.Text != "Creating file" {
		t.Fatalf("event 0 = %#v, want AssistantText{Creating file}", evs[0])
	}
	fe, ok := evs[1].(agent.FileEdit)
	if !ok || fe.Path != "/p/foo.go" {
		t.Fatalf("event 1 = %#v, want FileEdit for /p/foo.go", evs[1])
	}
	if !strings.Contains(fe.UnifiedDiff, "+package main") {
		t.Fatalf("diff missing added line:\n%s", fe.UnifiedDiff)
	}

	// Bash tool_use -> Command.
	evs = p.Parse(`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t2","name":"Bash","input":{"command":"go build ./..."}}]}}`)
	if len(evs) != 1 {
		t.Fatalf("bash line: got %d events, want 1", len(evs))
	}
	if c, ok := evs[0].(agent.Command); !ok || c.Cmd != "go build ./..." {
		t.Fatalf("event = %#v, want Command{go build ./...}", evs[0])
	}

	// tool_result -> ToolResult.
	evs = p.Parse(`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t2","content":"ok","is_error":false}]}}`)
	if len(evs) != 1 {
		t.Fatalf("tool_result line: got %d events, want 1", len(evs))
	}
	if tr, ok := evs[0].(agent.ToolResult); !ok || !tr.OK || tr.Summary != "ok" {
		t.Fatalf("event = %#v, want ToolResult{OK,ok}", evs[0])
	}

	// result -> Result with cost/duration/session.
	evs = p.Parse(`{"type":"result","subtype":"success","is_error":false,"result":"Done","total_cost_usd":0.0123,"duration_ms":4200,"session_id":"sess-abc"}`)
	if len(evs) != 1 {
		t.Fatalf("result line: got %d events, want 1", len(evs))
	}
	r, ok := evs[0].(agent.Result)
	if !ok {
		t.Fatalf("event = %#v, want Result", evs[0])
	}
	if r.Summary != "Done" {
		t.Fatalf("summary = %q, want Done", r.Summary)
	}
	if r.CostUSD == nil || *r.CostUSD != 0.0123 {
		t.Fatalf("cost = %v, want 0.0123", r.CostUSD)
	}
	if r.Duration.Milliseconds() != 4200 {
		t.Fatalf("duration = %v, want 4200ms", r.Duration)
	}
	if r.SessionID != "sess-abc" {
		t.Fatalf("result session id = %q, want sess-abc", r.SessionID)
	}
}

func TestParserEditDiff(t *testing.T) {
	p := NewParser()
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"a.txt","old_string":"foo","new_string":"bar"}}]}}`
	evs := p.Parse(line)
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	fe := evs[0].(agent.FileEdit)
	if !strings.Contains(fe.UnifiedDiff, "-foo") || !strings.Contains(fe.UnifiedDiff, "+bar") {
		t.Fatalf("edit diff wrong:\n%s", fe.UnifiedDiff)
	}
}

func TestParserIgnoresJunk(t *testing.T) {
	p := NewParser()
	if evs := p.Parse("not json"); evs != nil {
		t.Fatalf("junk line should yield nil, got %#v", evs)
	}
	if evs := p.Parse(""); evs != nil {
		t.Fatalf("empty line should yield nil")
	}
}
