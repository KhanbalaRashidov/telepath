package transport

// A tiny rich-text model so the transport-agnostic engine can build messages
// once and each transport render them in its own markup (Telegram HTML, Slack
// mrkdwn). It intentionally supports only what Telepath needs: plain / bold /
// inline-code spans, preformatted blocks, and blank lines.

// Span is a run of text with at most one emphasis.
type Span struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
}

// Text makes a plain span.
func Text(s string) Span { return Span{Text: s} }

// Bold makes a bold span.
func Bold(s string) Span { return Span{Text: s, Bold: true} }

// Italic makes an italic span.
func Italic(s string) Span { return Span{Text: s, Italic: true} }

// Code makes an inline-code span.
func Code(s string) Span { return Span{Text: s, Code: true} }

// Block is one line of inline spans, OR a preformatted code block (Pre), OR a
// blank separator line (Blank).
type Block struct {
	Spans []Span
	Pre   string
	Blank bool
}

// Doc is an ordered list of blocks rendered as consecutive lines.
type Doc struct {
	Blocks []Block
}

// NewDoc returns an empty document.
func NewDoc() *Doc { return &Doc{} }

// Line appends a line of spans.
func (d *Doc) Line(spans ...Span) *Doc {
	d.Blocks = append(d.Blocks, Block{Spans: spans})
	return d
}

// Pre appends a preformatted code block.
func (d *Doc) Pre(s string) *Doc {
	d.Blocks = append(d.Blocks, Block{Pre: s})
	return d
}

// Blank appends a blank separator line.
func (d *Doc) Blank() *Doc {
	d.Blocks = append(d.Blocks, Block{Blank: true})
	return d
}

// Empty reports whether the doc has no blocks.
func (d *Doc) Empty() bool { return len(d.Blocks) == 0 }
