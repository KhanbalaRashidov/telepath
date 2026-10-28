package telegram

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEscapeHTML(t *testing.T) {
	if got := escapeHTML("<a> & </a>"); got != "&lt;a&gt; &amp; &lt;/a&gt;" {
		t.Fatalf("escapeHTML = %q", got)
	}
}

func TestFileEditMessageInline(t *testing.T) {
	body, asDoc := fileEditMessage("foo.go", "--- a/foo.go\n+++ b/foo.go\n+hi\n")
	if asDoc {
		t.Fatal("small diff should be inline, not a document")
	}
	if !strings.Contains(body, "<pre>") || !strings.Contains(body, "foo.go") {
		t.Fatalf("inline body missing markup/path: %q", body)
	}
}

func TestFileEditMessageDocumentWhenHuge(t *testing.T) {
	huge := strings.Repeat("+a line of diff\n", 1000) // well over 4096 chars
	body, asDoc := fileEditMessage("big.txt", huge)
	if !asDoc {
		t.Fatal("huge diff should be sent as a document")
	}
	if strings.Contains(body, "<pre>") {
		t.Fatal("document fallback should not inline the diff body")
	}
}

func TestCommitMessage(t *testing.T) {
	msg := commitMessage("Fix the null check in parser\nsecond line", "claude")
	if !strings.HasPrefix(msg, "Fix the null check in parser") {
		t.Fatalf("summary wrong: %q", msg)
	}
	if !strings.Contains(msg, "via Telepath (claude)") {
		t.Fatalf("missing trailer: %q", msg)
	}
}

func TestThrottleCoalesces(t *testing.T) {
	var flushes int32
	var mu sync.Mutex
	var last string
	th := newThrottle(40*time.Millisecond, func(text string) {
		atomic.AddInt32(&flushes, 1)
		mu.Lock()
		last = text
		mu.Unlock()
	})

	// Fire many rapid updates within a single interval.
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

func TestSplitCommand(t *testing.T) {
	cases := []struct{ in, cmd, rest string }{
		{"/help", "help", ""},
		{"/use claude", "use", "claude"},
		{"/claude@MyBot fix it", "claude", "fix it"},
		{"/perm   interactive", "perm", "interactive"},
	}
	for _, c := range cases {
		cmd, rest := splitCommand(c.in)
		if cmd != c.cmd || rest != c.rest {
			t.Errorf("splitCommand(%q) = (%q,%q), want (%q,%q)", c.in, cmd, rest, c.cmd, c.rest)
		}
	}
}
