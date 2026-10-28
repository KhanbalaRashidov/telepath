> **Languages:** [English](../en/transports.md) · [Русский](../ru/transports.md) · [Türkçe](../tr/transports.md) · [Azərbaycanca](../az/transports.md)

# Transports

Telepath can talk to you over **Telegram**, **Slack**, or **both** at once. A
transport is just the messaging front-end; the core (agents, permissions, git
branches, the normalized event stream) is identical regardless of which one you
use. Pick transports with the `TRANSPORT` setting (`telegram` | `slack` |
`both`, default `telegram`).

## No public IP required

Both transports connect **outbound only** and work behind NAT with no webhook,
no inbound port, and no public URL:

- **Telegram** uses long-polling.
- **Slack** uses **Socket Mode** (a WebSocket the bot opens to Slack).

This is a core property of Telepath: you can run it on a laptop or a home server
and drive it from your phone.

## Telegram vs Slack

| | Telegram | Slack |
|---|----------|-------|
| Connection | long-polling | Socket Mode (WebSocket) |
| Tokens | one bot token | bot token `xoxb-` + app-level token `xapp-` |
| Trigger a task | any message, or `/<agent> …` | `@Telepath …`, a DM, or `/telepath …` |
| Commands | `/help`, `/perm`, … | `/telepath help`, `/telepath perm …`, … |
| Live status | one edited message | one edited message (in a thread) |
| Approvals | inline Approve/Deny buttons | Block Kit Approve/Deny buttons |
| Allowlist | chat IDs | user IDs and/or channel IDs |

Commands map **1:1** between the two. On Slack they are subcommands of the
single `/telepath` slash command (see [Telegram commands](telegram-commands.md),
which documents both).

## Running both at once

Set `TRANSPORT=both` and provide both the Telegram and Slack credentials.
Telepath starts both front-ends against the same engine, agents, and
`PROJECTS_DIR`. Because all conversations share one working tree, tasks are still
serialized globally — one task at a time across every Telegram chat and Slack
channel — and per-conversation state (chosen agent, permission mode, branch
mode) is keyed by `transport:conversation`, so a Telegram chat and a Slack
channel never interfere with each other.

## Choosing one

- Just you, on your phone → **Telegram** is the quickest to set up (one token).
- Your team already lives in Slack, or you want approvals visible to a channel →
  **Slack**.
- Migrating or want both audiences → **both**.

See [Getting started](getting-started.md) for step-by-step setup of each.

---

<!-- prev-next -->
[← Previous: Configuration](./configuration.md)  ·  [Next: Telegram commands →](./telegram-commands.md)
