# Telepath — Architecture

Telepath is a self-hosted bridge that lets you send coding tasks from Telegram
to a coding agent running on your own machine, and streams the agent's progress
back in real time. It adds two things on top of a plain bridge:

1. **Agent-agnostic** multi-agent support behind one interface (Claude Code is
   the reference; Gemini CLI, Codex CLI, and Aider are also supported).
2. A professional **git + permission** workflow: each task runs on its own
   isolated branch, and sensitive actions can be approved live from Telegram.

## Pipeline

```
Telegram update
   │  (chat allowlist middleware)
   ▼
command router (internal/telegram)
   │
   ▼
Session (internal/session)  ── per-chat agent / permission / branch mode
   │
   ▼
GitWorkspace.Prepare  (internal/git)  ── clean preflight, record base, create task branch
   │
   ▼
Agent.Run  (internal/agent/<name>)  ── spawn the vendor CLI, cwd = branch worktree
   │
   ▼
normalized Event stream  (internal/agent)  ── the ONLY thing the UI consumes
   │        │
   │        └── ApprovalRequest → Permission broker → Telegram buttons → decision
   ▼
Telegram renderer  (internal/telegram)  ── one live status message + per-file diffs
   │
   ▼
GitWorkspace.CommitAll  (internal/git)  ── commit on the task branch only
   │
   ▼
final report  ── branch, commit SHA, diff --stat, full .diff document
```

The Telegram layer and the git/session layers are **agent-independent**: they
consume only the normalized `Event` stream and the `Agent` interface. No
agent-specific type leaks upward.

## The Agent contract (`internal/agent`)

```go
type Agent interface {
    Name() string
    Available(ctx) bool
    Capabilities() Capabilities
    Run(ctx, Task) (<-chan Event, error)
    Resume(ctx, sessionID, msg string) (<-chan Event, error)
}

type Capabilities struct {
    StructuredStream, Sessions, Resume, InteractiveApproval, FileEditEvents bool
}

type Task struct {
    Prompt, ProjectDir string
    PermissionMode     PermissionMode
    ApprovalHook       ApprovalFunc  // non-nil only in interactive mode
    MaxTurns           int
    Model              string
}
```

### Normalized event model

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`,
`ToolResult`, `ApprovalRequest`, `Progress`, `BranchInfo`, `Committed`,
`Result{Summary, Duration, CostUSD, SessionID}`, `Error`.

Rich agents (Claude Code) emit the full set. Poorer agents degrade to
`AssistantText` plus the end-of-run git diff produced by the git layer. Every
downstream consumer handles both paths via capability checks, so a text-only
agent still shows which files changed (from the diff) even without `FileEdit`
events.

### Driver/registry pattern

Each agent is one package under `internal/agent/<name>/` implementing `Agent`.
A `Registry` holds the enabled agents and the global default; a chat overrides
its agent with `/use`. **Adding an agent = a new package + one registry line**
in `cmd/telepath/main.go`.

Adapters translate a vendor CLI into events:

| Agent  | Invocation (verified against current CLIs)                                      | Stream | Resume | Interactive |
|--------|---------------------------------------------------------------------------------|:------:|:------:|:-----------:|
| claude | `claude -p … --output-format stream-json --verbose --include-partial-messages`  |   ✅   |   ✅   |     ✅      |
| gemini | `gemini -p … --approval-mode=yolo\|default`                                      |   —    |   —    |     —       |
| codex  | `codex exec --json --sandbox workspace-write …`                                 |   JSONL|   —    |     —       |
| aider  | `aider --message … --yes --no-auto-commits --no-git`                            |   —    |   —    |     —       |

Gemini/Codex/Aider have no MCP permission hook, so interactive mode falls back
to `allowlist` and the reply states which mode was actually used.

## Permission subsystem (`internal/permission`)

`PermissionMode` is chosen globally, per-chat (`/perm`), or per-task:

- **auto** — no prompts; the agent runs autonomously. Claude:
  `--permission-mode bypassPermissions`; Gemini: `--approval-mode=yolo`; Codex:
  `--ask-for-approval never`; Aider: `--yes`. (default)
- **allowlist** — pre-approve a safe tool set, deny the rest, no prompts. Claude:
  `--allowedTools`/`--disallowedTools` with the defaults in `policy.go`. Headless
  denies silently and continues.
- **interactive** — the user approves/denies each gated action live.

### Interactive approval (Claude Code)

```
Claude wants a gated tool
   │  calls MCP tool  mcp__telepath__approve
   ▼
Telepath in-process MCP server (Streamable HTTP, loopback, random port)   [mcpserver.go]
   │  invokes Task.ApprovalHook(req)
   ▼
Broker.Await(id)  [broker.go]  ── registers pending, blocks
   │
   ├── Telegram shows Approve/Deny buttons
   │        └── tap → callback → Broker.Resolve(id, allow, note)
   └── timeout (default 120s) → DENY   /   task cancelled → DENY
   ▼
decision → MCP result JSON  ({"behavior":"allow","updatedInput":…} | {"behavior":"deny","message":…})
```

A fresh MCP server is started per interactive run and bound to `127.0.0.1:0`,
so there is **no cross-process IPC and no auth surface**. Claude is launched
with `--permission-prompt-tool mcp__telepath__approve` and an `--mcp-config`
file registering the HTTP server. A silent user never auto-approves: the broker
defaults to **deny** on timeout and on cancellation.

## Git workspace (`internal/git`) — Telepath owns git

Agents' own git/auto-commit is disabled. The **guard** (`guard.go`) is the
single choke point for every git invocation and refuses:

- `push`, `merge`, `rebase`, `pull`, `fetch`, `cherry-pick`, `revert`
- any `--force`/`-f`/`--force-with-lease`
- `reset --hard`
- committing onto, creating, or deleting `main`/`master`

**main and the remote are never modified. The most a task can do is leave a
reviewable local branch.**

`BranchMode`:

- **off** — run on the current branch, no branch, no commit; stream + end-of-run
  diff (legacy behavior).
- **per-task** (default) — per task:
  1. Preflight: refuse a dirty tree (or stash if `BRANCH_STASH_DIRTY=true`);
     record base branch + base SHA.
  2. Create & checkout `telepath/<slug>-<yymmdd-hhmm>`.
  3. Run the agent (file edits only).
  4. On success, if `BRANCH_COMMIT=true` (default) stage and commit **on the task
     branch** with message `<summary>` + trailer `via Telepath (<agent>)`.
  5. Never push/merge/touch main. Leave the branch for review.
  6. Report branch, commit SHA(s), `git diff base..HEAD --stat`, and the full
     diff as a `.diff` document, with Keep / Delete-&-restore buttons.

Because all chats share one `PROJECTS_DIR` working tree, tasks are serialized by
a global one-slot semaphore (and one running task per chat).

## Session store (`internal/session`)

`chat_id → {agent, permissionMode, branchMode, lastSessionID, running,
taskBranch, baseBranch, cancelFunc}`. In-memory in v1 behind a `Store` interface
so a persistent store can be dropped in later.

## Config (`internal/config`)

Resolved with precedence **flags > env > .env > defaults**, validated on startup
with actionable messages (e.g. refuses to start without `ALLOWED_CHAT_IDS`). See
`.env.example` and the README configuration table.

## Cross-platform notes

- Binary detection uses `exec.LookPath` (handles `.exe`/`PATHEXT` on Windows).
- `path/filepath` is used for paths; agents are spawned in their own process
  group and killed as a tree on `/cancel` (`taskkill /F /T` on Windows, process
  group `SIGKILL` on Unix).
