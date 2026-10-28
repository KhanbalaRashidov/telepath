> **Languages:** [English](../en/index.md) · [Русский](../ru/index.md) · [Türkçe](../tr/index.md) · [Azərbaycanca](../az/index.md)

# Telepath — Documentation

Telepath is a self-hosted bridge that lets you send coding tasks from Telegram
to a coding agent running on your own machine, and streams the agent's progress
back in real time: what it's doing, which files it edits, commands it runs,
approval requests, and a final summary with the full git diff. It is
**agent-agnostic** (Claude Code, Gemini CLI, OpenAI Codex CLI, Aider behind one
interface) and gives each task its own **git branch**, with optional live
**approvals** from Telegram.

## Contents

| Page | What's inside |
|------|---------------|
| [Getting started](getting-started.md) | Prerequisites, install per OS, create a bot, first run, first task |
| [Configuration](configuration.md) | Every environment variable and CLI flag, with defaults and precedence |
| [Transports](transports.md) | Telegram, Slack, or both; how to choose; the no-public-IP property |
| [Telegram commands](telegram-commands.md) | Every command and inline button (Telegram and Slack), with examples |
| [Agents](agents.md) | The multi-agent model, capability matrix, and exact CLI invocations |
| [Permissions](permissions.md) | `auto` / `allowlist` / `interactive` and how live approval works |
| [Branch workflow](branch-workflow.md) | Per-task branches, commits, guards, and cleanup |
| [Architecture](architecture.md) | Pipeline, event model, package map, MCP server |
| [Docker](docker.md) | Self-contained vs slim images, compose, mounts |
| [Adding an agent](adding-an-agent.md) | Implement the `Agent` interface and register a new adapter |
| [Security](security.md) | Allowlist, agent auth, the `auto` caveat, the git guarantee |
| [Troubleshooting](troubleshooting.md) | Common failures and their fixes |
| [FAQ](faq.md) | Short answers to common questions |

## The one thing to remember

**Telepath never pushes, never merges, and never touches `main`/`master` or the
remote.** The worst a task can do is leave a reviewable local branch that you
merge and push yourself.

---

<!-- prev-next -->
[Next: Getting started →](./getting-started.md)
