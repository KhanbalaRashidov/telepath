// Package slack implements the Slack transport using Socket Mode, which — like
// Telegram long-polling — needs no public URL or webhook and works behind NAT.
// It turns Slack events (slash command, app mention, direct message) into the
// engine's normalized inbound messages and renders outbound Docs/status/diffs/
// approvals as Slack mrkdwn with Block Kit buttons, threading task output under
// the triggering message.
//
// Required Slack scopes (configure on the app; verify against current Slack
// docs): app_mentions:read, chat:write, commands, files:write, im:history,
// im:read, channels:history, groups:history. Socket Mode also needs an
// app-level token (xapp-) with connections:write. The bot token is xoxb-.
package slack

import (
	"context"
	"log/slog"
	"strings"

	slackgo "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/telepath/telepath/internal/permission"
	"github.com/telepath/telepath/internal/transport"
)

const transportName = "slack"

// Transport is the Slack implementation of transport.Transport.
type Transport struct {
	api    *slackgo.Client
	sm     *socketmode.Client
	log    *slog.Logger
	router transport.Router
	broker *permission.Broker

	botUserID       string
	allowedUsers    map[string]bool
	allowedChannels map[string]bool
}

// New builds the Slack transport. botToken is the xoxb- bot token; appToken is
// the xapp- app-level token for Socket Mode.
func New(botToken, appToken string, allowedUsers, allowedChannels []string, router transport.Router, broker *permission.Broker, log *slog.Logger) (*Transport, error) {
	api := slackgo.New(botToken, slackgo.OptionAppLevelToken(appToken))
	t := &Transport{
		api:             api,
		sm:              socketmode.New(api),
		log:             log,
		router:          router,
		broker:          broker,
		allowedUsers:    toSet(allowedUsers),
		allowedChannels: toSet(allowedChannels),
	}
	if len(t.allowedUsers) == 0 && len(t.allowedChannels) == 0 {
		log.Warn("slack: no ALLOWED_SLACK_USER_IDS or ALLOWED_SLACK_CHANNEL_IDS set — every user in every channel the bot can see may trigger tasks")
	}
	return t, nil
}

func (t *Transport) Name() string { return transportName }

// Start resolves the bot's own user id, then runs the Socket Mode loop until
// ctx is cancelled.
func (t *Transport) Start(ctx context.Context) error {
	if auth, err := t.api.AuthTestContext(ctx); err == nil {
		t.botUserID = auth.UserID
	} else {
		t.log.Warn("slack: AuthTest failed; own-message filtering is best-effort", "err", err)
	}
	go t.loop(ctx)
	return t.sm.RunContext(ctx)
}

func (t *Transport) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-t.sm.Events:
			t.dispatch(ctx, evt)
		}
	}
}

func (t *Transport) dispatch(ctx context.Context, evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slackgo.SlashCommand)
		if !ok {
			return
		}
		t.sm.Ack(*evt.Request)
		t.onSlash(ctx, cmd)
	case socketmode.EventTypeInteractive:
		cb, ok := evt.Data.(slackgo.InteractionCallback)
		if !ok {
			return
		}
		t.sm.Ack(*evt.Request)
		t.onInteractive(ctx, cb)
	case socketmode.EventTypeEventsAPI:
		ev, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}
		t.sm.Ack(*evt.Request)
		t.onEventsAPI(ctx, ev)
	default:
		// connecting / connected / hello / disconnect / errors: nothing to do.
	}
}

// onSlash handles "/telepath <subcommand> …". The subcommand mirrors the
// Telegram commands 1:1; anything that is not a known command/agent is treated
// as a default-agent task.
func (t *Transport) onSlash(ctx context.Context, cmd slackgo.SlashCommand) {
	if !t.allowed(cmd.UserID, cmd.ChannelID) {
		t.log.Warn("slack: rejected slash command from non-allowlisted source", "user", cmd.UserID, "channel", cmd.ChannelID)
		return
	}
	conv := transport.ConversationRef{Transport: transportName, ConversationID: cmd.ChannelID}
	t.router.HandleMessage(ctx, transport.InboundMessage{
		Conversation: conv,
		UserID:       cmd.UserID,
		Text:         t.normalize(strings.TrimSpace(cmd.Text)),
	})
}

func (t *Transport) onInteractive(ctx context.Context, cb slackgo.InteractionCallback) {
	if len(cb.ActionCallback.BlockActions) == 0 {
		return
	}
	action := cb.ActionCallback.BlockActions[0].ActionID
	channel := cb.Channel.ID
	msgTS := cb.Message.Timestamp

	switch {
	case strings.HasPrefix(action, "ap:"):
		parts := strings.Split(strings.TrimPrefix(action, "ap:"), ":")
		if len(parts) != 2 {
			return
		}
		id, allow := parts[0], parts[1] == "a"
		note := "approved by user"
		if !allow {
			note = "denied by user"
		}
		// The blocked Approval call edits the prompt once it returns.
		t.broker.Resolve(id, allow, note)
	case action == "br:keep" || action == "br:del":
		act := "keep"
		if action == "br:del" {
			act = "delete"
		}
		conv := transport.ConversationRef{Transport: transportName, ConversationID: channel}
		res := t.router.HandleBranchAction(ctx, conv, act)
		text := res.Toast
		if res.Replace != nil {
			text = renderMrkdwn(*res.Replace)
		}
		if msgTS != "" && text != "" {
			t.replaceMessage(ctx, channel, msgTS, text)
		}
	}
}

func (t *Transport) onEventsAPI(ctx context.Context, ev slackevents.EventsAPIEvent) {
	switch inner := ev.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		if inner.User == "" || inner.User == t.botUserID {
			return
		}
		if !t.allowed(inner.User, inner.Channel) {
			t.log.Warn("slack: rejected mention from non-allowlisted source", "user", inner.User, "channel", inner.Channel)
			return
		}
		t.routeEvent(ctx, inner.Channel, threadRoot(inner.ThreadTimeStamp, inner.TimeStamp), inner.User, stripMention(inner.Text))
	case *slackevents.MessageEvent:
		// Only plain user direct messages trigger tasks; ignore bot messages,
		// edits/joins (subtypes), and the bot's own posts.
		if inner.ChannelType != "im" || inner.BotID != "" || inner.SubType != "" || inner.User == "" || inner.User == t.botUserID {
			return
		}
		if !t.allowed(inner.User, inner.Channel) {
			t.log.Warn("slack: rejected DM from non-allowlisted source", "user", inner.User, "channel", inner.Channel)
			return
		}
		t.routeEvent(ctx, inner.Channel, threadRoot(inner.ThreadTimeStamp, inner.TimeStamp), inner.User, strings.TrimSpace(inner.Text))
	}
}

func (t *Transport) routeEvent(ctx context.Context, channel, thread, user, text string) {
	conv := transport.ConversationRef{Transport: transportName, ConversationID: channel, ThreadID: thread}
	t.router.HandleMessage(ctx, transport.InboundMessage{
		Conversation: conv,
		UserID:       user,
		Text:         t.normalize(text),
	})
}

// normalize turns Slack free text into the engine's "/command" convention: a
// leading known command or agent name becomes "/word …"; empty text becomes
// "/help"; anything else is passed through as a default-agent task prompt.
func (t *Transport) normalize(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "/help"
	}
	word, rest := firstWord(text)
	if t.router.Known(word) {
		if rest != "" {
			return "/" + word + " " + rest
		}
		return "/" + word
	}
	return text
}

// firstWord returns the first whitespace-delimited token (with any trailing
// colon stripped, to support the "@Telepath claude: <task>" style) and the rest.
func firstWord(text string) (word, rest string) {
	if i := strings.IndexAny(text, " \t\n"); i >= 0 {
		word, rest = text[:i], strings.TrimSpace(text[i+1:])
	} else {
		word = text
	}
	word = strings.TrimSuffix(strings.ToLower(word), ":")
	return word, rest
}

// stripMention removes a leading "<@U…>" mention token from app_mention text.
func stripMention(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "<@") {
		if i := strings.IndexByte(text, '>'); i >= 0 {
			return strings.TrimSpace(text[i+1:])
		}
	}
	return text
}

func threadRoot(threadTS, ts string) string {
	if threadTS != "" {
		return threadTS
	}
	return ts
}

func (t *Transport) allowed(user, channel string) bool {
	if len(t.allowedUsers) == 0 && len(t.allowedChannels) == 0 {
		return true
	}
	return t.allowedUsers[user] || t.allowedChannels[channel]
}

// Reply renders a Doc and posts it to the conversation (threaded if a thread is
// known).
func (t *Transport) Reply(ctx context.Context, conv transport.ConversationRef, doc transport.Doc) {
	opts := []slackgo.MsgOption{slackgo.MsgOptionText(renderMrkdwn(doc), false)}
	if conv.ThreadID != "" {
		opts = append(opts, slackgo.MsgOptionTS(conv.ThreadID))
	}
	if _, _, err := t.api.PostMessageContext(ctx, conv.ConversationID, opts...); err != nil {
		t.log.Warn("slack: reply failed", "channel", conv.ConversationID, "err", err)
	}
}

// NewSession creates a per-task outbound session bound to a conversation.
func (t *Transport) NewSession(conv transport.ConversationRef) transport.Session {
	return &slSession{t: t, channel: conv.ConversationID, threadTS: conv.ThreadID}
}

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			m[s] = true
		}
	}
	return m
}
