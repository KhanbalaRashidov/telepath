package telegram

import (
	"context"
	"sync"
	"time"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/transport"
)

// tgSession is the per-task outbound session for Telegram. It owns the single
// live status message (edited in place, throttled) and renders notices, file
// diffs, approvals, and the final report as Telegram HTML.
type tgSession struct {
	t      *Transport
	conv   transport.ConversationRef
	chatID int64

	mu          sync.Mutex
	th          *transport.Throttle
	statusMsgID int
}

// Status creates the live status message on first call, then edits it in place
// (throttled) on subsequent calls.
func (s *tgSession) Status(ctx context.Context, v transport.StatusView) {
	html := escapeHTML(transport.RenderStatusText(v))

	s.mu.Lock()
	if s.statusMsgID == 0 {
		id, _ := s.t.send(ctx, s.chatID, html, nil)
		s.statusMsgID = id
		s.th = transport.NewThrottle(1500*time.Millisecond, func(text string) {
			s.mu.Lock()
			mid := s.statusMsgID
			s.mu.Unlock()
			if mid != 0 {
				s.t.edit(ctx, s.chatID, mid, text, nil)
			}
		})
		s.mu.Unlock()
		return
	}
	th := s.th
	s.mu.Unlock()
	if th != nil {
		th.Update(html)
	}
}

func (s *tgSession) flushStatus() {
	s.mu.Lock()
	th := s.th
	s.mu.Unlock()
	if th != nil {
		th.Flush()
	}
}

func (s *tgSession) stopStatus() {
	s.mu.Lock()
	th := s.th
	s.mu.Unlock()
	if th != nil {
		th.Stop()
	}
}

// Notice posts a one-off message.
func (s *tgSession) Notice(ctx context.Context, doc transport.Doc) {
	s.t.sendText(ctx, s.chatID, renderHTML(doc))
}

// FileEdit posts a file's diff — inline for small diffs, as a .diff document
// when it would exceed Telegram's message-length limit.
func (s *tgSession) FileEdit(ctx context.Context, path, unifiedDiff string) {
	s.flushStatus()
	body, asDoc := fileEditSplit(path, unifiedDiff)
	if asDoc {
		s.t.sendDoc(ctx, s.chatID, transport.SafeDiffName(path), []byte(unifiedDiff), "✏️ <b>"+escapeHTML(path)+"</b>")
		return
	}
	s.t.sendText(ctx, s.chatID, body)
}

// Approval shows Approve/Deny buttons and blocks on the broker until the user
// taps or the broker times out.
func (s *tgSession) Approval(ctx context.Context, req agent.ApprovalRequest) agent.ApprovalDecision {
	id := newID()
	msgID, _ := s.t.send(ctx, s.chatID, renderHTML(transport.ApprovalPromptDoc(req)), approvalKeyboard(id))
	dec := s.t.broker.Await(ctx, id)
	if msgID != 0 {
		s.t.edit(ctx, s.chatID, msgID, renderHTML(transport.ApprovalResolvedDoc(req, dec.Allow, dec.Note)), nil)
	}
	return agent.ApprovalDecision{Allow: dec.Allow, Note: dec.Note}
}

// Result posts the final report, matching the original ordering: on a per-task
// success the report message carries the Keep/Delete buttons and the full diff
// follows; on failure the error message is followed by a separate branch prompt
// carrying the buttons.
func (s *tgSession) Result(ctx context.Context, r transport.ResultView) {
	s.stopStatus()
	if s.statusMsgID != 0 {
		s.t.edit(ctx, s.chatID, s.statusMsgID, escapeHTML(transport.RenderStatusText(r.Status)), nil)
	}

	bodyHTML := renderHTML(r.Body)
	if r.Actions != nil && r.Prompt == nil {
		s.t.send(ctx, s.chatID, bodyHTML, branchKeyboard())
		if r.Diff != nil {
			s.t.sendDoc(ctx, s.chatID, r.Diff.Name, []byte(r.Diff.Content), renderHTML(r.Diff.Caption))
		}
		return
	}

	s.t.sendText(ctx, s.chatID, bodyHTML)
	if r.Diff != nil {
		s.t.sendDoc(ctx, s.chatID, r.Diff.Name, []byte(r.Diff.Content), renderHTML(r.Diff.Caption))
	}
	if r.Actions != nil && r.Prompt != nil {
		s.t.send(ctx, s.chatID, renderHTML(*r.Prompt), branchKeyboard())
	}
}
