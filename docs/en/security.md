> **Languages:** [English](../en/security.md) · [Русский](../ru/security.md) · [Türkçe](../tr/security.md) · [Azərbaycanca](../az/security.md)

# Security

Telepath runs real coding agents on your machine, so understand the trust model.

## Chat-ID allowlist

Telepath acts only on chats listed in `ALLOWED_CHAT_IDS`. The allowlist is
enforced on **every update — messages and button taps alike** — before anything
else happens. Telepath **refuses to start** without at least one allowed chat ID,
and silently ignores any non-allowlisted chat. Keep your bot token secret; anyone
with it can message your bot, but only allowlisted chats are ever acted on.

## Slack allowlist

On Slack, access is controlled by `ALLOWED_SLACK_USER_IDS` and/or
`ALLOWED_SLACK_CHANNEL_IDS` (separate from the Telegram chat allowlist). A
request is accepted if the triggering **user** is in the user list **or** the
**channel** is in the channel list. These lists are optional, and — unlike
Telegram — if **both are empty Telepath allows every user in every channel the
bot can see** and logs a warning at startup. For anything beyond a private test
workspace, set at least one list. Keep the `xoxb-`/`xapp-` tokens secret: they
grant the bot's access to your workspace.

## Agents use their own credentials

Each agent CLI authenticates with its **own** API key or login, stored on the
host per that agent's documentation. Telepath never reads, stores, or transmits
those credentials — it only spawns the CLI. The only secrets Telepath itself
holds are its transport tokens (the Telegram bot token and/or the Slack
`xoxb-`/`xapp-` tokens) and the allowlists.

## The `auto` (skip-permissions) caveat

In `auto` mode the agent runs with its permission prompts bypassed — it can edit
files and run shell commands without asking. That is convenient but powerful.
Prefer:

- `allowlist` for a safety net that still needs no babysitting, or
- `interactive` (Claude) when you want to see and approve every risky step.

See [Permissions](permissions.md).

## Prompt-injection awareness

A coding agent reads your repository's files and the task text. Malicious content
in either (a poisoned file, a crafted task) can try to steer the agent into
unwanted actions. Mitigations:

- Use `allowlist` or `interactive` on repositories you don't fully trust.
- Review the task branch before merging — which the workflow is designed around.
- Remember the denied set blocks `curl`/`wget`/`WebFetch`, `rm`, `git push`, etc.
  in `allowlist` mode.

## The git guarantee

Telepath owns all git through a single guarded code path that **never** pushes,
merges, rebases, force-anything, or commits onto `main`/`master`. The remote and
your main branch are never modified. **The worst case of any task is a reviewable
local branch** that you choose to keep or delete. See
[Branch workflow](branch-workflow.md).

## Network posture

Telepath connects **outbound only** — Telegram over **long-polling** and Slack
over **Socket Mode** (a WebSocket the bot opens) — so it works behind NAT with
no public IP, webhook, or inbound port. The interactive-approval MCP server
binds to **loopback only** (`127.0.0.1`) on a random port for the lifetime of a
single run.

---

<!-- prev-next -->
[← Previous: Adding an agent](./adding-an-agent.md)  ·  [Next: Troubleshooting →](./troubleshooting.md)
