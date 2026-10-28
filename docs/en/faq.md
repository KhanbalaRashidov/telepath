> **Languages:** [English](../en/faq.md) · [Русский](../ru/faq.md) · [Türkçe](../tr/faq.md) · [Azərbaycanca](../az/faq.md)

# FAQ

**Which operating systems are supported?**
Windows, Linux, and macOS natively (one codebase, pure-Go, CGO off), and as a
Docker container.

**Does Telepath commit or push my code?**
It commits on a **task branch** (when `BRANCH_COMMIT=true`), but it **never
pushes, merges, or touches `main`/`master` or the remote**. You review and
merge/push yourself. With `BRANCH_MODE=off` it doesn't even commit.

**Which agents support live approval?**
Only **Claude Code** (via its MCP permission-prompt tool). For gemini, codex, and
aider, `interactive` automatically falls back to `allowlist`.

**Does it work with Slack, or only Telegram?**
Both. Set `TRANSPORT=slack` (or `both`) and provide the `xoxb-`/`xapp-` tokens.
Trigger tasks by mentioning the bot, DMing it, or using `/telepath …`. See
[Transports](transports.md).

**Can it run multiple tasks at once?**
No. All conversations (Telegram chats and Slack channels) share one
`PROJECTS_DIR` working tree, so tasks are **serialized** — one at a time overall,
one per conversation. See [Troubleshooting](troubleshooting.md).

**Does it work behind NAT / without a public IP?**
Yes. Telegram uses **long-polling** and Slack uses **Socket Mode** — both are
outbound only, with no webhook, inbound port, or public IP required.

**Where do my agent API keys go?**
They stay with the agent CLIs on your host, configured per each agent's own docs.
Telepath only holds the Telegram bot token and chat allowlist. See
[Security](security.md).

**Can I switch agents or modes per chat?**
Yes — `/use <agent>`, `/perm <mode>`, `/branch <mode>` apply to that chat and
persist until the process restarts.

**How do I add another agent?**
Implement the `Agent` interface in a new package and register it with one line.
See [Adding an agent](adding-an-agent.md).

**How do I add a documentation language?**
Copy `docs/en/` to `docs/<locale>/`, translate the prose, and register the locale
in `mkdocs.yml` and the switcher links. See
[CONTRIBUTING.md](../../CONTRIBUTING.md#adding-a-documentation-language).

**How do I update Telepath?**
Pull the latest binary from Releases, or rebuild from source
(`go install …@latest` or `make build`). Config and `.env` are unchanged.

**What are the costs?**
Telepath is free (MIT). You pay only for whatever the underlying agent/model
usage costs under your own agent accounts.

---

<!-- prev-next -->
[← Previous: Troubleshooting](./troubleshooting.md)
