# Telepath

**Send coding tasks to an AI coding agent on your own machine — from Telegram.
Watch it work in real time, approve its risky moves, and get every change as a
clean, reviewable git branch.**

<!-- Badges point at the GitHub repo KhanbalaRashidov/telepath (from `git remote`).
     If you fork/rename it, replace KhanbalaRashidov/telepath below.
     The CI and Release badges stay "no status"/"no releases" until the first
     push + workflow run and the first tagged release, respectively. -->
[![CI](https://img.shields.io/github/actions/workflow/status/KhanbalaRashidov/telepath/ci.yml?branch=main&label=CI)](https://github.com/KhanbalaRashidov/telepath/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/KhanbalaRashidov/telepath)](https://github.com/KhanbalaRashidov/telepath/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/KhanbalaRashidov/telepath)](go.mod)
[![License](https://img.shields.io/github/license/KhanbalaRashidov/telepath)](./LICENSE)

Telepath is a self-hosted bridge between **Telegram or Slack** and a coding agent
(Claude Code, Gemini CLI, OpenAI Codex CLI, or Aider) running on your dev box.
You text it a task; it runs the agent inside your project and streams back what
it's doing — assistant notes, file diffs, commands, approval requests, and a
final summary with the full git diff.

Three things set it apart from a plain bridge:

- 💬 **Telegram, Slack, or both.** Pick your front-end with `TRANSPORT`; both
  work behind NAT with no public IP (long-polling / Socket Mode).
- 🧩 **Agent-agnostic.** One common interface drives multiple agents; switch per
  conversation with `/use`. Adding an agent is a new package plus one line.
- 🌿 **A real git + permission workflow.** Each task runs on its own branch, and
  you can approve/deny the agent's sensitive actions live from chat.
  **Telepath never pushes, never merges, and never touches `main`/`master`.**

## Documentation

Full docs live in [`docs/`](docs/README.md), in four languages:

- 🇬🇧 [English](docs/en/index.md) · 🇷🇺 [Русский](docs/ru/index.md) · 🇹🇷 [Türkçe](docs/tr/index.md) · 🇦🇿 [Azərbaycanca](docs/az/index.md)

They cover [getting started](docs/en/getting-started.md),
[configuration](docs/en/configuration.md),
[Telegram commands](docs/en/telegram-commands.md),
[agents](docs/en/agents.md), [permissions](docs/en/permissions.md),
[branch workflow](docs/en/branch-workflow.md), [Docker](docs/en/docker.md),
[security](docs/en/security.md), and more. The quick start below stays here.

## What it is

```mermaid
flowchart LR
    TG["Telegram<br/>(you)"] -- "task / approvals" --> R["Telepath<br/>command router"]
    R --> S["Session<br/>(per-chat state)"]
    R --> G["GitWorkspace<br/>(per-task branch)"]
    G --> A["Agent adapter<br/>claude · gemini · codex · aider"]
    A -- "normalized<br/>Event stream" --> UI["Renderer<br/>(live status + diffs)"]
    A -. "ApprovalRequest" .-> P["Permission broker<br/>+ in-process MCP server"]
    P -- "Approve / Deny buttons" --> TG
    UI -- "status · diffs · result" --> TG
    G -- "commit on branch · diff" --> UI
```

The agent runs with its working directory set to a per-task git branch. Telepath
owns all git; the agent's own git/auto-commit is disabled. See
[ARCHITECTURE.md](ARCHITECTURE.md) for the full design.

## Supported agents

| Agent  | Install                                             | Streaming | Sessions | Interactive approval | Status |
|--------|-----------------------------------------------------|:---------:|:--------:|:--------------------:|--------|
| **Claude Code** | `npm i -g @anthropic-ai/claude-code`       | ✅ stream-json | ✅ | ✅ (MCP prompt tool) | Full reference adapter |
| **Gemini CLI**  | `npm i -g @google/gemini-cli`              | — (text)  | — | — → falls back to allowlist | Supported (text + git diff) |
| **Codex CLI**   | `npm i -g @openai/codex`                   | ✅ JSONL   | — | — → falls back to allowlist | Supported (best-effort parse) |
| **Aider**       | `pipx install aider-chat`                  | — (text)  | — | — → falls back to allowlist | Supported (git disabled, text + diff) |

Agents without an MCP permission hook can't do live approval; Telepath falls
back to `allowlist` for them and tells you which mode it used.

> CLI flags for these tools change often. Telepath was built against their
> current interfaces; if an agent changes, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Quick start

### 1. Create a bot and find your chat ID

- Create a bot with [@BotFather](https://t.me/BotFather) and copy the token.
- Message [@userinfobot](https://t.me/userinfobot) to get your numeric chat ID.

### 2. Install Telepath

**Prebuilt binary** — download for your OS from
[Releases](https://github.com/KhanbalaRashidov/telepath/releases), then:

```bash
# Linux / macOS
chmod +x telepath && ./telepath --version
```
```powershell
# Windows (PowerShell)
.\telepath.exe --version
```

**From source (`go install`):**
```bash
go install github.com/telepath/telepath/cmd/telepath@latest
```

**Docker (self-contained — bundles Node + agent CLIs):**
```bash
docker run --rm -it \
  -e TELEGRAM_BOT_TOKEN=xxxx \
  -e ALLOWED_CHAT_IDS=123456789 \
  -e PROJECTS_DIR=/work/project \
  -v /path/to/your/repo:/work/project \
  ghcr.io/khanbalarashidov/telepath:latest
```

**Docker slim (bring your own agent CLIs):**
```bash
docker build -t telepath:slim -f Dockerfile.slim .
# mount your repo and point *_BIN at your mounted CLIs
```

Or with compose:
```bash
cp .env.example .env    # fill in token, chat IDs, PROJECTS_DIR
docker compose up -d
```

### 3. Configure and run

```bash
cp .env.example .env
# edit .env: TELEGRAM_BOT_TOKEN, ALLOWED_CHAT_IDS, PROJECTS_DIR
./telepath
```

Then in Telegram: send `/help`, then just describe a task, e.g.
`fix the null check in parser.go and add a test`.

**Prefer Slack?** Set `TRANSPORT=slack` with a Slack app in Socket Mode:

```dotenv
TRANSPORT=slack
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token
ALLOWED_SLACK_USER_IDS=U0123456789
PROJECTS_DIR=/path/to/your/git/repo
```

Then `@Telepath <task>`, DM the bot, or use `/telepath help`. Use
`TRANSPORT=both` to run Telegram and Slack together. See the
[Transports](docs/en/transports.md) and [getting-started](docs/en/getting-started.md) docs.

## Configuration

Resolved in precedence order **CLI flags > environment > `.env` > defaults**.
Every variable has a `--kebab-case` flag equivalent (see `telepath --help`).

| Env var | Flag | Default | Description |
|---------|------|---------|-------------|
| `TRANSPORT` | `--transport` | `telegram` | `telegram` \| `slack` \| `both` |
| `TELEGRAM_BOT_TOKEN` | `--telegram-bot-token` | *(telegram)* | Bot token from @BotFather |
| `ALLOWED_CHAT_IDS` | `--allowed-chat-ids` | *(telegram)* | CSV of Telegram chat IDs allowed to use the bot |
| `SLACK_BOT_TOKEN` | `--slack-bot-token` | *(slack)* | Slack bot token `xoxb-` |
| `SLACK_APP_TOKEN` | `--slack-app-token` | *(slack)* | Slack app-level token `xapp-` (Socket Mode) |
| `ALLOWED_SLACK_USER_IDS` | `--allowed-slack-user-ids` | *(empty)* | CSV of allowed Slack user IDs (optional) |
| `ALLOWED_SLACK_CHANNEL_IDS` | `--allowed-slack-channel-ids` | *(empty)* | CSV of allowed Slack channel IDs (optional) |
| `PROJECTS_DIR` | `--projects-dir` | `.` | Directory the agent works in (a git repo for per-task branches) |
| `DEFAULT_AGENT` | `--default-agent` | `claude` | Default agent (must be enabled) |
| `ENABLED_AGENTS` | `--enabled-agents` | `claude,gemini,codex,aider` | CSV of enabled agents |
| `PERMISSION_MODE` | `--permission-mode` | `auto` | `auto` \| `allowlist` \| `interactive` |
| `APPROVAL_TIMEOUT_SECONDS` | `--approval-timeout-seconds` | `120` | Interactive approval timeout (then denies) |
| `BRANCH_MODE` | `--branch-mode` | `per-task` | `off` \| `per-task` |
| `BRANCH_PREFIX` | `--branch-prefix` | `telepath/` | Task branch name prefix |
| `BRANCH_COMMIT` | `--branch-commit` | `true` | Commit the agent's changes on the task branch |
| `BRANCH_STASH_DIRTY` | `--branch-stash-dirty` | `false` | Stash a dirty tree instead of refusing |
| `MAX_TURNS` | `--max-turns` | `0` | Max agent turns (0 = CLI default) |
| `MODEL` | `--model` | *(empty)* | Model override passed to the agent |
| `CLAUDE_BIN` / `GEMINI_BIN` / `CODEX_BIN` / `AIDER_BIN` | `--claude-bin` … | *(empty)* | Agent binary path overrides |
| `GIT_BIN` | `--git-bin` | *(empty)* | git binary override |
| `LOG_LEVEL` | `--log-level` | `info` | `debug` \| `info` \| `warn` \| `error` |

## Telegram commands

| Command | Description |
|---------|-------------|
| `/help` | Usage |
| `/agents` | List configured agents + availability (installed? streaming? approval?) |
| `/use <agent>` | Set this chat's default agent |
| `/perm <auto\|allowlist\|interactive>` | Set this chat's permission mode |
| `/branch <off\|per-task>` | Set this chat's branch mode |
| `/<agent> <task>` | Run a task with a specific agent (e.g. `/claude fix the null check`) |
| `<plain message>` | Run with the chat's default agent |
| `/resume <text>` | Continue the last session in this chat (if the agent supports it) |
| `/branches` | List `telepath/*` branches |
| `/status` | Current running task + mode summary |
| `/cancel` | Cancel the running task (kills the process; offers branch cleanup) |

## Permissions

Set globally (`PERMISSION_MODE`), per chat (`/perm`), or effectively per task.

- **auto** (default) — no prompts; the agent runs autonomously.
- **allowlist** — a safe tool set is pre-approved and everything else is denied,
  with no prompts. Headless agents deny disallowed tools silently and continue.
  See the defaults in `internal/permission/policy.go`.
- **interactive** — you approve/deny each sensitive action live.

**How interactive approval works (Claude Code):** Telepath starts a tiny
in-process MCP server on loopback and launches Claude with
`--permission-prompt-tool mcp__telepath__approve` plus an `--mcp-config` that
registers it. When the agent wants a gated tool, the MCP tool is called;
Telepath forwards it to Telegram as an **Approve / Deny** message and blocks on
your tap. The timeout and cancellation both **deny** — a silent user never
auto-approves. Agents without an MCP hook fall back to `allowlist`.

## Branch workflow

With `BRANCH_MODE=per-task` (default), each task:

1. Checks the working tree is clean (or stashes it if `BRANCH_STASH_DIRTY=true`)
   and records the base branch + commit.
2. Creates and checks out `telepath/<slug>-<yymmdd-hhmm>`.
3. Runs the agent (file edits only — the agent's git is disabled).
4. Commits the changes **on the task branch** (if `BRANCH_COMMIT=true`) with a
   message ending `via Telepath (<agent>)`.
5. Reports the branch, commit SHA, `git diff base..HEAD --stat`, and the full
   diff as a `.diff` document — with **Keep branch** / **Delete branch & restore
   base** buttons.

**Telepath never pushes, never merges, and never modifies `main`/`master` or the
remote.** The most a task can do is leave a reviewable local branch; you merge
and push it yourself. `BRANCH_MODE=off` runs on the current branch with no
branch/commit and just shows the end-of-run diff.

## Security notes

- **Chat allowlist.** Every update (messages *and* button taps) is checked
  against `ALLOWED_CHAT_IDS`. Telepath refuses to start without it and ignores
  every non-allowlisted chat.
- **Agent auth stays local.** API keys / logins belong to the agent CLIs on your
  machine; Telepath never transmits them. Only your bot token and chat IDs are
  Telepath's.
- **Headless skip-permissions caveat.** `auto` mode runs the agent with its
  permission prompts bypassed — it can edit files and run commands without
  asking. Use `allowlist` or `interactive` if you want gating. Even so, the git
  guard ensures the blast radius is a local branch, never your remote or `main`.
- **Interactive defaults to deny** on timeout and cancellation.

## Add your own agent

1. Create `internal/agent/<name>/<name>.go` implementing the `agent.Agent`
   interface.
2. Verify the CLI's **real** current flags, then map its output to the
   normalized `Event` stream (or emit `AssistantText` and rely on the end-of-run
   git diff).
3. Register it with one line in `cmd/telepath/main.go` and add a row to the
   table above.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the details.

## Roadmap

- Persistent session store (SQLite) behind the existing `session.Store` interface.
- Native interactive approval for more agents as their CLIs expose hooks.
- Per-project selection (work in a subfolder of `PROJECTS_DIR`).
- Optional allowlist rules editable from Telegram.

## License

[MIT](LICENSE).
