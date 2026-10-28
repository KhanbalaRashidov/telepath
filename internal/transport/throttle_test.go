package transport

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestThrottleCoalesces(t *testing.T) {
	var flushes int32
	var mu sync.Mutex
	var last string
	th := NewThrottle(40*time.Millisecond, func(text string) {
		atomic.AddInt32(&flushes, 1)
		mu.Lock()
		last = text
		mu.Unlock()
	})

	for i := 0; i < 20; i++ {
		th.Update("status " + string(rune('a'+i%26)))
		time.Sleep(2 * time.Millisecond)
	}
	th.Update("final")
	time.Sleep(60 * time.Millisecond)
	th.Stop()

	n := atomic.LoadInt32(&flushes)
	if n == 0 {
		t.Fatal("expected at least one flush")
	}
	if n >= 20 {
		t.Fatalf("expected coalescing (far fewer than 20 flushes), got %d", n)
	}
	mu.Lock()
	defer mu.Unlock()
	if last != "final" {
		t.Fatalf("last flushed text = %q, want final", last)
	}
}

func TestRenderStatusText(t *testing.T) {
	v := StatusView{Agent: "claude", Branch: "telepath/fix", File: "parser.go", Starting: true}
	got := RenderStatusText(v)
	if !strings.HasPrefix(got, "⏳ ") {
		t.Fatalf("Starting status should begin with the hourglass: %q", got)
	}
	if !strings.Contains(got, "🤖 claude") || !strings.Contains(got, "parser.go") {
		t.Fatalf("status text missing parts: %q", got)
	}
	v.Starting = false
	v.State = "done"
	if d := RenderStatusText(v); !strings.Contains(d, "✅ done") {
		t.Fatalf("done status missing marker: %q", d)
	}
}
