> **Languages:** [English](../en/agents.md) · [Русский](../ru/agents.md) · [Türkçe](../tr/agents.md) · [Azərbaycanca](../az/agents.md)

# Agents

Telepath drives several coding-agent CLIs behind one common interface. Every
adapter translates a vendor CLI into the same normalized event stream, so the
Telegram and git layers never see agent-specific details. Switch agents per chat
with `/use`, or run one task with a specific agent via `/<agent> <task>`.

## Capability matrix

| Agent | Install | Streaming (FileEdit events) | Sessions / resume | Interactive approval | Status |
|-------|---------|:---------------------------:|:-----------------:|:--------------------:|--------|
| **claude** | `npm install -g @anthropic-ai/claude-code` | ✅ full (stream-json, per-edit diffs) | ✅ | ✅ (MCP prompt tool) | Full reference adapter |
| **gemini** | `npm install -g @google/gemini-cli` | — (text; edits via end-of-run diff) | — | — → falls back to `allowlist` | Supported |
| **codex** | `npm install -g @openai/codex` | JSONL (best-effort; edits via end-of-run diff) | — | — → falls back to `allowlist` | Supported |
| **aider** | `pipx install aider-chat` | — (text; edits via end-of-run diff) | — | — → falls back to `allowlist` | Supported |

`/agents` in Telegram prints this matrix with live install status for your host.

## Exact invocations

These are the real commands each adapter builds. Flags in brackets are added
only when relevant (`--max-turns` when `MAX_TURNS>0`, `--model` when `MODEL` is
set, permission flags per mode).

### Claude Code (reference)

```bash
claude -p "<task>" \
  --output-format stream-json --verbose --include-partial-messages \
  [--max-turns <N>] [--model <model>] \
  # auto:        --permission-mode bypassPermissions
  # allowlist:   --permission-mode default --allowedTools "<list>" --disallowedTools "<list>"
  # interactive: --permission-mode default \
  #              --permission-prompt-tool mcp__telepath__approve --mcp-config <file>
```

Resume:

```bash
claude --resume <session_id> -p "<message>" \
  --output-format stream-json --verbose --include-partial-messages \
  --permission-mode bypassPermissions
```

Claude emits the full event set: assistant text, `tool_use` → `FileEdit`
(Edit/Write/MultiEdit, with a unified diff), `Bash` → `Command`, `tool_result`,
and a final `result` carrying the summary, duration, cost, and session id.

### Gemini CLI

```bash
gemini -p "<task>" [-m <model>] \
  # auto:  --approval-mode=yolo
  # other: --approval-mode=default
```

Gemini output is read as text (its JSON output has been unstable across
releases). Lines become `AssistantText`; file changes are shown by the git
layer's end-of-run diff.

### OpenAI Codex CLI

```bash
codex exec --json \
  # auto:  --sandbox workspace-write --ask-for-approval never
  # other: --sandbox workspace-write --ask-for-approval on-request
  [--model <model>] "<task>"
```

> The older `--full-auto` flag is deprecated; Telepath uses `--sandbox
> workspace-write`. Codex emits a JSONL event stream that is parsed on a
> best-effort basis into `AssistantText`/`Command`; file changes also appear in
> the end-of-run diff.

### Aider

```bash
aider --message "<task>" --no-stream [--model <model>] \
  --yes --no-auto-commits --no-git
```

Aider's own git integration is **disabled** (`--no-auto-commits --no-git`) so
Telepath owns version control, and it runs non-interactively (`--message
--yes`). Output is text; file changes appear in the end-of-run diff.

## Honest limits (read this)

- **Interactive approval works only with Claude** (it needs an MCP
  permission-prompt hook). For gemini/codex/aider, choosing `interactive`
  automatically falls back to `allowlist`, and the reply tells you which mode was
  actually used.
- **gemini/codex/aider emit no per-edit `FileEdit` events** during a run — their
  file changes are surfaced by the git layer's end-of-run diff instead.
- **Resume is Claude-only.** Other agents return "not supported" for `/resume`.
- **Codex JSONL parsing is best-effort** because the schema varies across
  versions; unrecognized lines are ignored (or shown as text).
- Each agent CLI is an **external dependency** you install and authenticate
  yourself. If a CLI is missing, `/agents` shows it as not installed and tasks
  for it are refused with a clear message.

See [Adding an agent](adding-an-agent.md) to write your own adapter.
