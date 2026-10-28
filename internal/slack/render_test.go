package slack

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/telepath/telepath/internal/transport"
)

func TestRenderMrkdwn(t *testing.T) {
	d := transport.NewDoc()
	d.Line(transport.Text("🤖 "), transport.Bold("Telepath"))
	d.Line(transport.Text("path "), transport.Code("a<b>.go"), transport.Text(" "), transport.Italic("note"))
	d.Pre("diff & <stuff>")
	got := renderMrkdwn(*d)
	want := "🤖 *Telepath*\npath `a&lt;b&gt;.go` _note_\n```\ndiff &amp; &lt;stuff&gt;\n```"
	if got != want {
		t.Fatalf("renderMrkdwn:\n got %q\nwant %q", got, want)
	}
}

func TestEscapeMrkdwn(t *testing.T) {
	if got := escapeMrkdwn("a & b < c > d"); got != "a &amp; b &lt; c &gt; d" {
		t.Fatalf("escapeMrkdwn = %q", got)
	}
}

func TestShouldInline(t *testing.T) {
	if !shouldInline(strings.Repeat("x", slackMaxInline)) {
		t.Fatal("a diff at the limit should be inline")
	}
	if shouldInline(strings.Repeat("x", slackMaxInline+1)) {
		t.Fatal("a diff over the limit should be uploaded, not inline")
	}
}

func TestStripMention(t *testing.T) {
	if got := stripMention("<@U123> claude fix the bug"); got != "claude fix the bug" {
		t.Fatalf("stripMention = %q", got)
	}
	if got := stripMention("no mention here"); got != "no mention here" {
		t.Fatalf("stripMention passthrough = %q", got)
	}
}

func TestFirstWord(t *testing.T) {
	w, r := firstWord("claude: fix the bug")
	if w != "claude" || r != "fix the bug" {
		t.Fatalf("firstWord colon = (%q,%q)", w, r)
	}
	w, r = firstWord("help")
	if w != "help" || r != "" {
		t.Fatalf("firstWord single = (%q,%q)", w, r)
	}
}

// fakeRouter lets us exercise normalize without a live Slack connection.
type fakeRouter struct{ known map[string]bool }

func (f fakeRouter) HandleMessage(context.Context, transport.InboundMessage) {}
func (f fakeRouter) HandleBranchAction(context.Context, transport.ConversationRef, string) transport.ActionResult {
	return transport.ActionResult{}
}
func (f fakeRouter) Known(word string) bool { return f.known[word] }

func TestNormalize(t *testing.T) {
	tr, err := New("xoxb-test", "xapp-test", nil, nil, fakeRouter{known: map[string]bool{"help": true, "perm": true, "claude": true}}, nil, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"":                 "/help",             // empty mention/slash → help
		"help":             "/help",             // bare command
		"perm interactive": "/perm interactive", // command with args
		"claude fix bug":   "/claude fix bug",   // agent prefix
		"claude: fix bug":  "/claude fix bug",   // agent prefix with colon
		"fix the bug now":  "fix the bug now",   // plain task (unknown first word)
	}
	for in, want := range cases {
		if got := tr.normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
