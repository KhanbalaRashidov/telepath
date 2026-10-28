package engine

import (
	"strings"
	"testing"
)

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

func TestCommitMessage(t *testing.T) {
	msg := commitMessage("Fix the null check in parser\nsecond line", "claude")
	if !strings.HasPrefix(msg, "Fix the null check in parser") {
		t.Fatalf("summary wrong: %q", msg)
	}
	if !strings.Contains(msg, "via Telepath (claude)") {
		t.Fatalf("missing trailer: %q", msg)
	}
	if strings.Contains(msg, "second line") {
		t.Fatalf("only the first line should be the summary: %q", msg)
	}
}
