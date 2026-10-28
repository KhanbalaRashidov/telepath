package telegram

// Telegram message-size limit and the inline-vs-document decision for file
// diffs. (The status-update throttling itself is shared and lives in
// internal/transport.)

// telegramMaxLen is Telegram's hard per-message length limit. Content that
// would exceed it (large diffs) is sent as a document instead.
const telegramMaxLen = 4096

// fileEditSplit builds the inline HTML body for a file diff, and reports
// whether it must be sent as a document because it exceeds the length limit.
func fileEditSplit(path, diff string) (body string, asDocument bool) {
	header := "✏️ <b>" + escapeHTML(path) + "</b>\n"
	body = header + "<pre>" + escapeHTML(diff) + "</pre>"
	if len([]rune(body)) > telegramMaxLen {
		return header, true
	}
	return body, false
}
