> **Languages:** [English](../en/troubleshooting.md) · [Русский](../ru/troubleshooting.md) · [Türkçe](../tr/troubleshooting.md) · [Azərbaycanca](../az/troubleshooting.md)

# Troubleshooting

Set `LOG_LEVEL=debug` for more detail. Below are the common failure modes.

## "Agent is not installed"

`/agents` shows a ✅/❌ for each agent. If yours is ❌:

- Confirm the CLI is on PATH (`claude --version`, `gemini --version`,
  `codex --version`, `aider --version`).
- If it lives somewhere non-standard, set the matching `*_BIN` variable
  (`CLAUDE_BIN`, `GEMINI_BIN`, `CODEX_BIN`, `AIDER_BIN`) to its full path.
- In Docker, use the self-contained image (it bundles the CLIs) or mount them
  into the slim image.

## Bot doesn't respond

- **Token:** check `TELEGRAM_BOT_TOKEN`. On startup Telepath validates the token
  with Telegram; a wrong token fails fast.
- **Allowlist:** your chat ID must be in `ALLOWED_CHAT_IDS`. Non-allowlisted
  chats are ignored silently (see the warning in the logs).
- **Connectivity:** Telepath uses outbound long-polling; make sure the host can
  reach `api.telegram.org`.
- **Running elsewhere:** only one bot process should long-poll a given token at a
  time.

## "Working tree is dirty" refusal

In `per-task` mode Telepath refuses to start a task if `PROJECTS_DIR` has
uncommitted changes. Either commit/stash them yourself, or set
`BRANCH_STASH_DIRTY=true` to let Telepath stash them automatically.

## Approval timed out → denied

In `interactive` mode, if no one taps Approve/Deny within
`APPROVAL_TIMEOUT_SECONDS` (default 120), the request is **denied** and the agent
is told so. Increase the timeout if you need more time, or use `allowlist`/`auto`
if you don't want to be prompted.

## Interactive didn't prompt me

Interactive approval works **only with Claude**. With gemini/codex/aider,
`interactive` falls back to `allowlist`, and the reply states the mode actually
used. Switch the chat to Claude with `/use claude` for live approvals.

## Tasks seem to queue

All chats share one `PROJECTS_DIR` working tree, so tasks are **serialized** —
one at a time across all chats, one per chat. If the workspace is busy you get a
"try again" reply. Use `/status` to see what's running and `/cancel` to stop it.

## `/cancel` didn't kill everything (Windows)

Telepath kills the agent process **tree** best-effort — `taskkill /F /T` on
Windows, process-group `SIGKILL` on Unix. Deeply nested helper processes can
occasionally linger; check Task Manager if a Node process survives.

## `/resume` says there's nothing to resume

The last session id and its project directory are kept **in memory** and are lost
when Telepath restarts. After a restart there is nothing to resume. Resume is
also **Claude-only**; other agents report it as unsupported.

## Commit failed

If the repo has no git identity, Telepath commits with a fallback
`Telepath <telepath@localhost>` identity so the task-branch commit still
succeeds. If a commit still fails, the error is shown in the final report —
check the repo's git state and permissions.
