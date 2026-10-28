> **Languages:** [English](../en/telegram-commands.md) · [Русский](../ru/telegram-commands.md) · [Türkçe](../tr/telegram-commands.md) · [Azərbaycanca](../az/telegram-commands.md)

# Telegram commands

All commands and buttons work only in chats listed in `ALLOWED_CHAT_IDS`.
Settings changed with `/use`, `/perm`, and `/branch` apply **per chat**.

## Commands

| Command | Syntax | Example | Effect |
|---------|--------|---------|--------|
| `/help` | `/help` | `/help` | Show the command list. |
| `/agents` | `/agents` | `/agents` | List configured agents with their install status and capabilities (streaming, resume, interactive approval, file-edit events). |
| `/use` | `/use <agent>` | `/use gemini` | Set this chat's default agent. |
| `/perm` | `/perm <mode>` | `/perm interactive` | Set this chat's permission mode: `auto`, `allowlist`, or `interactive`. |
| `/branch` | `/branch <mode>` | `/branch per-task` | Set this chat's branch mode: `off` or `per-task`. |
| `/<agent>` | `/<agent> <task>` | `/claude fix the null check in parser.go` | Run a task with a specific agent, ignoring the chat default for this one task. |
| *(plain message)* | `<task text>` | `add retry logic to the HTTP client` | Run a task with the chat's default agent. |
| `/resume` | `/resume <text>` | `/resume now add a test for it` | Continue the last session in this chat (Claude only). |
| `/branches` | `/branches` | `/branches` | List existing `telepath/*` branches in `PROJECTS_DIR`. |
| `/status` | `/status` | `/status` | Show the current agent/permission/branch modes and whether a task is running. |
| `/cancel` | `/cancel` | `/cancel` | Cancel the running task. Telepath cancels the context and kills the agent process tree. |

Notes:

- A command may be addressed to the bot by name (e.g. `/claude@MyBot …`); the
  `@name` suffix is stripped.
- Only **one task per chat** runs at a time, and because all chats share one
  working tree, tasks are **serialized globally**. If the workspace is busy you
  get a short "try again" reply.
- `/resume` works only for agents that report the resume capability — currently
  **Claude**. It runs in `auto` permission mode, and the project directory is
  restored from an in-memory map that is lost if Telepath restarts.

## Slack commands

On Slack, every command above is a subcommand of the single `/telepath` slash
command, and tasks are triggered by mentioning the bot or DMing it. The mapping
is 1:1 with Telegram:

| Telegram | Slack |
|----------|-------|
| `/help` | `/telepath help` |
| `/agents` | `/telepath agents` |
| `/use <agent>` | `/telepath use <agent>` |
| `/perm <mode>` | `/telepath perm <mode>` |
| `/branch <mode>` | `/telepath branch <mode>` |
| `/branches` | `/telepath branches` |
| `/status` | `/telepath status` |
| `/cancel` | `/telepath cancel` |
| `/resume <text>` | `/telepath resume <text>` |
| `/<agent> <task>` | `/telepath <agent> <task>` or `@Telepath <agent>: <task>` |
| *(plain message)* | `@Telepath <task>` or a direct message |

Notes:

- A `/telepath` subcommand that is not a known command or agent name is treated
  as a default-agent task (e.g. `/telepath fix the flaky test`).
- Task output is posted in a **thread** under the triggering message (or under
  the bot's first reply for slash commands). The live status message is edited
  in place, exactly like Telegram.
- Who may use the bot is controlled by `ALLOWED_SLACK_USER_IDS` /
  `ALLOWED_SLACK_CHANNEL_IDS` (see [Security](security.md)).

## Inline buttons

The buttons below work on both transports (Telegram inline keyboards, Slack
Block Kit).

### Approval (interactive mode)

When a task runs in `interactive` mode with Claude, each gated action appears as
a message with two buttons:

- **✅ Approve** — allow the action; the agent proceeds.
- **⛔ Deny** — refuse the action; the agent is told it was denied.

If no one responds within `APPROVAL_TIMEOUT_SECONDS`, the request is **denied**
automatically. See [Permissions](permissions.md).

### After a task (per-task branch mode)

When a task finishes (or fails) on a task branch, Telepath offers:

- **🌿 Keep branch** — leave the branch in place for you to review, merge, and
  push yourself.
- **🗑 Delete branch & restore base** — discard the task branch and check the
  base branch back out.

## The live status message

During a run, Telepath maintains a **single status message** that it edits in
place, for example:

```
🤖 claude · 🌿 telepath/fix-null-… · ✏️ parser.go · ▶ go build ./... · ✅ committed a1b2c3d
```

Edits are throttled so the message updates about once per second rather than on
every token. Each file edit is also posted as its own message with a unified
diff (or, if too large for Telegram, as a `.diff` document).

---

<!-- prev-next -->
[← Previous: Transports](./transports.md)  ·  [Next: Agents →](./agents.md)
