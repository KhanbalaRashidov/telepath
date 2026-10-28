package slack

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	slackgo "github.com/slack-go/slack"

	"github.com/telepath/telepath/internal/agent"
	"github.com/telepath/telepath/internal/transport"
)

// slackMaxInline is the largest diff posted inline as a code block; bigger
// diffs are uploaded as a .diff file via the external-upload flow.
const slackMaxInline = 2800

// slSession is the per-task outbound session for Slack. It threads all task
// output under the triggering message (or, for slash commands with no source
// message, under its own first post) and edits a single live status message.
type slSession struct {
	t       *Transport
	channel string

	mu       sync.Mutex
	threadTS string // thread to post under ("" until adopted from the first post)
	statusTS string // the live status message's timestamp
	th       *transport.Throttle
}

func (s *slSession) threadOpts(base []slackgo.MsgOption) []slackgo.MsgOption {
	s.mu.Lock()
	tt := s.threadTS
	s.mu.Unlock()
	if tt != "" {
		base = append(base, slackgo.MsgOptionTS(tt))
	}
	return base
}

func (s *slSession) adopt(ts string) {
	s.mu.Lock()
	if s.threadTS == "" {
		s.threadTS = ts
	}
	s.mu.Unlock()
}

func (s *slSession) postText(ctx context.Context, text string) string {
	_, ts, err := s.t.api.PostMessageContext(ctx, s.channel, s.threadOpts([]slackgo.MsgOption{slackgo.MsgOptionText(text, false)})...)
	if err != nil {
		s.t.log.Warn("slack: post failed", "channel", s.channel, "err", err)
		return ""
	}
	s.adopt(ts)
	return ts
}

func (s *slSession) postBlocks(ctx context.Context, fallback string, blocks []slackgo.Block) string {
	opts := s.threadOpts([]slackgo.MsgOption{slackgo.MsgOptionBlocks(blocks...), slackgo.MsgOptionText(fallback, false)})
	_, ts, err := s.t.api.PostMessageContext(ctx, s.channel, opts...)
	if err != nil {
		s.t.log.Warn("slack: post (blocks) failed", "channel", s.channel, "err", err)
		return ""
	}
	s.adopt(ts)
	return ts
}

func (s *slSession) updateText(ctx context.Context, ts, text string) {
	_, _, _, err := s.t.api.UpdateMessageContext(ctx, s.channel, ts, slackgo.MsgOptionText(text, false))
	if err != nil {
		s.t.log.Debug("slack: update failed", "channel", s.channel, "ts", ts, "err", err)
	}
}

// replaceMessage replaces a message (dropping any buttons) with a single
// mrkdwn section. Used after an approval/branch decision.
func (t *Transport) replaceMessage(ctx context.Context, channel, ts, text string) {
	_, _, _, err := t.api.UpdateMessageContext(ctx, channel, ts,
		slackgo.MsgOptionBlocks(section(text)), slackgo.MsgOptionText(text, false))
	if err != nil {
		t.log.Debug("slack: replace failed", "channel", channel, "ts", ts, "err", err)
	}
}

func (s *slSession) Status(ctx context.Context, v transport.StatusView) {
	text := renderStatusMrkdwn(v)
	s.mu.Lock()
	created := s.statusTS == ""
	th := s.th
	s.mu.Unlock()

	if !created {
		if th != nil {
			th.Update(text)
		}
		return
	}

	ts := s.postText(ctx, text)
	s.mu.Lock()
	s.statusTS = ts
	s.th = transport.NewThrottle(1500*time.Millisecond, func(t string) {
		s.mu.Lock()
		st := s.statusTS
		s.mu.Unlock()
		if st != "" {
			s.updateText(ctx, st, t)
		}
	})
	s.mu.Unlock()
}

func (s *slSession) flushStatus() {
	s.mu.Lock()
	th := s.th
	s.mu.Unlock()
	if th != nil {
		th.Flush()
	}
}

func (s *slSession) stopStatus() {
	s.mu.Lock()
	th := s.th
	s.mu.Unlock()
	if th != nil {
		th.Stop()
	}
}

func (s *slSession) Notice(ctx context.Context, doc transport.Doc) {
	s.postText(ctx, renderMrkdwn(doc))
}

func (s *slSession) FileEdit(ctx context.Context, path, unifiedDiff string) {
	s.flushStatus()
	header := "✏️ *" + escapeMrkdwn(path) + "*"
	if shouldInline(unifiedDiff) {
		s.postText(ctx, header+"\n```\n"+escapeMrkdwn(unifiedDiff)+"\n```")
		return
	}
	s.upload(ctx, transport.SafeDiffName(path), unifiedDiff, header)
}

// shouldInline reports whether a diff is small enough to post inline as a code
// block; larger diffs are uploaded as a .diff file.
func shouldInline(diff string) bool { return len([]rune(diff)) <= slackMaxInline }

func (s *slSession) Approval(ctx context.Context, req agent.ApprovalRequest) agent.ApprovalDecision {
	id := newID()
	promptText := renderMrkdwn(transport.ApprovalPromptDoc(req))
	ts := s.postBlocks(ctx, promptText, approvalBlocks(id, promptText))
	dec := s.t.broker.Await(ctx, id)
	if ts != "" {
		s.t.replaceMessage(ctx, s.channel, ts, renderMrkdwn(transport.ApprovalResolvedDoc(req, dec.Allow, dec.Note)))
	}
	return agent.ApprovalDecision{Allow: dec.Allow, Note: dec.Note}
}

func (s *slSession) Result(ctx context.Context, r transport.ResultView) {
	s.stopStatus()
	s.mu.Lock()
	st := s.statusTS
	s.mu.Unlock()
	if st != "" {
		s.updateText(ctx, st, renderStatusMrkdwn(r.Status))
	}

	body := renderMrkdwn(r.Body)
	if r.Actions != nil && r.Prompt == nil {
		s.postBlocks(ctx, body, branchBlocks(body))
		if r.Diff != nil {
			s.upload(ctx, r.Diff.Name, r.Diff.Content, renderMrkdwn(r.Diff.Caption))
		}
		return
	}

	s.postText(ctx, body)
	if r.Diff != nil {
		s.upload(ctx, r.Diff.Name, r.Diff.Content, renderMrkdwn(r.Diff.Caption))
	}
	if r.Actions != nil && r.Prompt != nil {
		pt := renderMrkdwn(*r.Prompt)
		s.postBlocks(ctx, pt, branchBlocks(pt))
	}
}

// upload sends a diff as a file using Slack's current external-upload flow
// (GetUploadURLExternal → UploadToURL → CompleteUploadExternal), NOT the
// deprecated files.upload.
func (s *slSession) upload(ctx context.Context, name, content, comment string) {
	up, err := s.t.api.GetUploadURLExternalContext(ctx, slackgo.GetUploadURLExternalParameters{
		FileName: name,
		FileSize: len(content),
	})
	if err != nil {
		s.t.log.Warn("slack: get upload url failed", "err", err)
		s.postText(ctx, comment+"\n⚠️ could not upload diff: "+escapeMrkdwn(err.Error()))
		return
	}
	if err := s.t.api.UploadToURL(ctx, slackgo.UploadToURLParameters{
		UploadURL: up.UploadURL,
		Content:   content,
		Filename:  name,
	}); err != nil {
		s.t.log.Warn("slack: upload to url failed", "err", err)
		return
	}
	params := slackgo.CompleteUploadExternalParameters{
		Files:          []slackgo.FileSummary{{ID: up.FileID, Title: name}},
		Channel:        s.channel,
		InitialComment: comment,
	}
	s.mu.Lock()
	params.ThreadTimestamp = s.threadTS
	s.mu.Unlock()
	if _, err := s.t.api.CompleteUploadExternalContext(ctx, params); err != nil {
		s.t.log.Warn("slack: complete upload failed", "err", err)
	}
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
