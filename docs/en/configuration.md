> **Languages:** [English](../en/configuration.md) · [Русский](../ru/configuration.md) · [Türkçe](../tr/configuration.md) · [Azərbaycanca](../az/configuration.md)

# Configuration

Telepath reads configuration from four sources. The precedence, highest first,
is:

**CLI flags > environment variables > `.env` file > built-in defaults**

Every environment variable has a `--kebab-case` flag equivalent. The `.env` file
only fills values that are not already set in the real environment; run
`telepath --help` to see the flags.

## Reference

In the **Required** column, `telegram` / `slack` means the variable is required
only when that transport is enabled; blank means optional.

| Env var | CLI flag | Default | Required | Description |
|---------|----------|---------|:--------:|-------------|
| `TRANSPORT` | `--transport` | `telegram` | | `telegram` \| `slack` \| `both`. Which messaging front-end(s) to run. See [Transports](transports.md). |
| `TELEGRAM_BOT_TOKEN` | `--telegram-bot-token` | — | telegram | Bot token from @BotFather. Required when `TRANSPORT` includes `telegram`. |
| `ALLOWED_CHAT_IDS` | `--allowed-chat-ids` | — | telegram | Comma-separated Telegram chat IDs allowed to use the bot. Required (≥1) when `telegram` is enabled. |
| `SLACK_BOT_TOKEN` | `--slack-bot-token` | — | slack | Slack bot token (`xoxb-…`). Required when `TRANSPORT` includes `slack`. |
| `SLACK_APP_TOKEN` | `--slack-app-token` | — | slack | Slack app-level token (`xapp-…`) for Socket Mode. Required when `slack` is enabled. |
| `ALLOWED_SLACK_USER_IDS` | `--allowed-slack-user-ids` | — | | Comma-separated Slack user IDs allowed to trigger tasks (optional). |
| `ALLOWED_SLACK_CHANNEL_IDS` | `--allowed-slack-channel-ids` | — | | Comma-separated Slack channel IDs allowed to trigger tasks (optional). |
| `PROJECTS_DIR` | `--projects-dir` | `.` | | Directory the agent works in. Must be a git repository for `per-task` branch mode. |
| `DEFAULT_AGENT` | `--default-agent` | `claude` | | Default agent for a chat. Must be listed in `ENABLED_AGENTS`. |
| `ENABLED_AGENTS` | `--enabled-agents` | `claude,gemini,codex,aider` | | Comma-separated enabled agents. |
| `PERMISSION_MODE` | `--permission-mode` | `auto` | | `auto` \| `allowlist` \| `interactive`. |
| `APPROVAL_TIMEOUT_SECONDS` | `--approval-timeout-seconds` | `120` | | How long an interactive approval waits before it is **denied**. |
| `BRANCH_MODE` | `--branch-mode` | `per-task` | | `off` \| `per-task`. |
| `BRANCH_PREFIX` | `--branch-prefix` | `telepath/` | | Prefix for task branch names. |
| `BRANCH_COMMIT` | `--branch-commit` | `true` | | Commit the agent's changes on the task branch. |
| `BRANCH_STASH_DIRTY` | `--branch-stash-dirty` | `false` | | Stash a dirty working tree instead of refusing to start a task. |
| `MAX_TURNS` | `--max-turns` | `0` | | Maximum agent turns (`0` = the CLI's own default). |
| `MODEL` | `--model` | *(empty)* | | Model override passed to the agent (`""` = the CLI's default). |
| `CLAUDE_BIN` | `--claude-bin` | *(empty)* | | Path to the `claude` binary if not on PATH. |
| `GEMINI_BIN` | `--gemini-bin` | *(empty)* | | Path to the `gemini` binary if not on PATH. |
| `CODEX_BIN` | `--codex-bin` | *(empty)* | | Path to the `codex` binary if not on PATH. |
| `AIDER_BIN` | `--aider-bin` | *(empty)* | | Path to the `aider` binary if not on PATH. |
| `GIT_BIN` | `--git-bin` | *(empty)* | | Path to the `git` binary if not on PATH. |
| `LOG_LEVEL` | `--log-level` | `info` | | `debug` \| `info` \| `warn` \| `error`. |

There are also two meta-flags that print and exit: `--version` and `--help`
(plus `--env-file` to point at a `.env` other than `./.env`).

## Notes

- **Boolean values** accept `1/true/yes/on` and `0/false/no/off`.
- **`ALLOWED_CHAT_IDS`** is a hard security boundary. Every update — messages and
  button taps — is checked against it; non-allowlisted chats are ignored.
- **`DEFAULT_AGENT` must be in `ENABLED_AGENTS`**, otherwise startup fails with a
  clear message. Likewise `PERMISSION_MODE` and `BRANCH_MODE` must be valid.
- **`PROJECTS_DIR`** must exist and be a directory. All chats share this one
  working tree (see [Branch workflow](branch-workflow.md)).
- Per-chat settings (`/use`, `/perm`, `/branch`) override these global defaults
  for that chat only, and are kept in memory until the process restarts.

## Example `.env`

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789,987654321
PROJECTS_DIR=/srv/projects/myrepo
DEFAULT_AGENT=claude
ENABLED_AGENTS=claude,gemini,codex,aider
PERMISSION_MODE=interactive
APPROVAL_TIMEOUT_SECONDS=120
BRANCH_MODE=per-task
BRANCH_PREFIX=telepath/
BRANCH_COMMIT=true
BRANCH_STASH_DIRTY=false
LOG_LEVEL=info
```

---

<!-- prev-next -->
[← Previous: Getting started](./getting-started.md)  ·  [Next: Transports →](./transports.md)
