package telegram

import (
	"strings"
	"testing"

	"github.com/telepath/telepath/internal/transport"
)

func TestEscapeHTML(t *testing.T) {
	if got := escapeHTML("<a> & </a>"); got != "&lt;a&gt; &amp; &lt;/a&gt;" {
		t.Fatalf("escapeHTML = %q", got)
	}
}

func TestRenderHTML(t *testing.T) {
	d := transport.NewDoc()
	d.Line(transport.Text("🤖 "), transport.Bold("Telepath"))
	d.Blank()
	d.Line(transport.Text("path "), transport.Code("a<b>.go"))
	d.Pre("diff & <stuff>")
	got := renderHTML(*d)
	want := "🤖 <b>Telepath</b>\n\npath <code>a&lt;b&gt;.go</code>\n<pre>diff &amp; &lt;stuff&gt;</pre>"
	if got != want {
		t.Fatalf("renderHTML:\n got %q\nwant %q", got, want)
	}
}

func TestFileEditSplitInline(t *testing.T) {
	body, asDoc := fileEditSplit("foo.go", "--- a/foo.go\n+++ b/foo.go\n+hi\n")
	if asDoc {
		t.Fatal("small diff should be inline, not a document")
	}
	if !strings.Contains(body, "<pre>") || !strings.Contains(body, "foo.go") {
		t.Fatalf("inline body missing markup/path: %q", body)
	}
}

func TestFileEditSplitDocumentWhenHuge(t *testing.T) {
	huge := strings.Repeat("+a line of diff\n", 1000) // well over 4096 chars
	body, asDoc := fileEditSplit("big.txt", huge)
	if !asDoc {
		t.Fatal("huge diff should be sent as a document")
	}
	if strings.Contains(body, "<pre>") {
		t.Fatal("document fallback should not inline the diff body")
	}
}
