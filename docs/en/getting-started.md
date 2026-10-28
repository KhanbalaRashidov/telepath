> **Languages:** [English](../en/getting-started.md) · [Русский](../ru/getting-started.md) · [Türkçe](../tr/getting-started.md) · [Azərbaycanca](../az/getting-started.md)

# Getting started

This walkthrough takes you from nothing to your first task in Telegram.

## 1. Prerequisites

- **Go 1.26+** (only to build from source) — <https://go.dev/dl/>
- **git** on your PATH.
- At least one **agent CLI** installed on the host. You only need the ones you
  plan to enable:

  | Agent | Install |
  |-------|---------|
  | Claude Code | `npm install -g @anthropic-ai/claude-code` |
  | Gemini CLI | `npm install -g @google/gemini-cli` |
  | OpenAI Codex CLI | `npm install -g @openai/codex` |
  | Aider | `pipx install aider-chat` (or `pip install aider-chat`) |

  Each agent is an **external dependency** with its own authentication (API key
  or login) that you set up per its own docs. Telepath never handles those
  credentials. After starting Telepath, `/agents` shows which CLIs it found.

## 2. Install Telepath

### Prebuilt binary

Download the archive for your OS/arch from the project's Releases page, extract
it, and check it runs:

```bash
# Linux / macOS
chmod +x telepath
./telepath --version
```
```powershell
# Windows (PowerShell)
.\telepath.exe --version
```

### From source with `go install`

```bash
go install github.com/telepath/telepath/cmd/telepath@latest
```

> If the repository was forked/renamed, use its real module path instead of
> `github.com/telepath/telepath` (for example `github.com/KhanbalaRashidov/telepath`).

### From source with the Makefile

```bash
git clone https://github.com/telepath/telepath
cd telepath
make build      # produces ./telepath (or telepath.exe on Windows)
```

## 3. Create a Telegram bot

1. Open [@BotFather](https://t.me/BotFather) in Telegram.
2. Send `/newbot` and follow the prompts.
3. Copy the **bot token** it gives you.

## 4. Find your chat ID

Message [@userinfobot](https://t.me/userinfobot); it replies with your numeric
chat ID. Telepath only serves chats you explicitly allowlist.

## 5. Minimal `.env`

Copy the example and fill in the three essentials:

```bash
cp .env.example .env
```

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789
PROJECTS_DIR=/path/to/your/git/repo
```

`PROJECTS_DIR` is the folder the agent works in. For the default per-task branch
workflow it **must be a git repository**. See [Configuration](configuration.md)
for all options.

## 6. First run

```bash
./telepath
```

On start, Telepath validates the config (and refuses to start without a bot
token and at least one allowed chat ID), then begins long-polling Telegram.

## 7. Your first task

In your Telegram chat with the bot:

1. Send `/help` to see the commands.
2. Send `/agents` to confirm your agent is installed.
3. Just describe a task in plain text, for example:

   ```
   add a /health endpoint that returns 200 OK and a test for it
   ```

Telepath creates a branch like `telepath/add-a-health-endpoint-260115-1430`,
runs the agent there, streams a live status message plus per-file diffs, commits
the result on that branch, and finally posts the `git diff --stat` with the full
diff as a `.diff` document and **Keep branch / Delete branch & restore base**
buttons.

Next: read [Telegram commands](telegram-commands.md),
[Permissions](permissions.md), and the [Branch workflow](branch-workflow.md).
