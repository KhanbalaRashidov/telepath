package telegram

import "github.com/go-telegram/bot/models"

// Callback data prefixes. Telegram limits callback_data to 64 bytes, so these
// stay short and reference server-side state (the broker / session) by id.
const (
	cbApprovePrefix = "ap:" // ap:<id>:a (allow) | ap:<id>:d (deny)
	cbBranchKeep    = "br:keep"
	cbBranchDelete  = "br:del"
)

// approvalKeyboard builds the Approve/Deny inline keyboard for an approval id.
func approvalKeyboard(id string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Approve", CallbackData: cbApprovePrefix + id + ":a"},
				{Text: "⛔ Deny", CallbackData: cbApprovePrefix + id + ":d"},
			},
		},
	}
}

// branchKeyboard builds the post-task Keep/Delete inline keyboard.
func branchKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🌿 Keep branch", CallbackData: cbBranchKeep},
				{Text: "🗑 Delete branch & restore base", CallbackData: cbBranchDelete},
			},
		},
	}
}
