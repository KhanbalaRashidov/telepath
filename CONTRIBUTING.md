# Contributing to Telepath

Thanks for your interest! Telepath is a small, dependency-lean Go project and we
want to keep it that way.

## Development

```bash
make build   # build ./telepath
make test    # go test ./...
make vet     # go vet ./...
make lint    # golangci-lint run (install golangci-lint first)
make fmt     # gofmt -s -w .
```

Before opening a PR, make sure `go build ./...`, `go vet ./...`, and
`go test ./...` all pass, and the code is `gofmt`-clean. CI runs all of these.

## Ground rules

- **Pure Go, CGO off.** Dependencies must be pure-Go so the binary
  cross-compiles cleanly. Keep the dependency list lean; justify any new one.
- **Keep the layers separate.** The `telegram`, `git`, and `session` layers must
  stay agent-independent — they consume only the normalized `agent.Event` stream
  and the `agent.Agent` interface. Never leak agent-specific types upward.
- **Telepath owns git.** All git goes through `internal/git` and its guard.
  Never add an operation that pushes, merges, rebases, force-anything, or commits
  onto `main`/`master`. If you touch git, add/extend a guard test.

## Adding a new agent

1. Create `internal/agent/<name>/<name>.go` implementing `agent.Agent`.
2. **Verify the CLI's real flags first** (`<cli> --help` / its docs) — these
   tools change often. Implement to the real interface.
3. Set `Capabilities()` honestly. If the agent has no structured stream or
   approval hook, emit `AssistantText` and rely on the end-of-run git diff, and
   leave `InteractiveApproval=false` (interactive will fall back to allowlist).
4. Map the permission mode to concrete flags in `internal/permission/policy.go`.
5. Register it with one line in `cmd/telepath/main.go` and add a row to the
   README "Supported agents" table.
6. Add a parser unit test if the adapter parses structured output.

## Adding a documentation language

The docs live in `docs/<locale>/`, one folder per language, each with the same
set of pages. English (`docs/en/`) is the canonical source. Each language's home
page is `index.md` (not `README.md`) — mkdocs-static-i18n builds the default
language to the site root, so a `README.md` home would collide with the
top-level `docs/README.md` hub. To add a language:

1. Copy `docs/en/` to `docs/<locale>/` (e.g. `docs/de/`); keep the home as
   `index.md`.
2. Translate the **prose** in every page. Do **not** translate code blocks,
   commands, env-var names, CLI flags, agent invocations, file/path names, JSON
   keys, tool names (e.g. `mcp__telepath__approve`), or Telegram command names
   (`/help`, `/perm`, …). Keep tables, mermaid diagrams, and internal links
   parallel to `en/`.
3. Keep the one-line **language switcher** at the top of each page; the switcher
   text is identical in every language (relative links like `../en/agents.md`),
   so just add your locale to it across all pages (including `en/`).
4. Add the locale to `mkdocs.yml` under `plugins → i18n → languages` with its
   `name` and a `nav_translations` block for the menu labels.
5. Add a row to the language table in `docs/README.md` and the switcher line in
   the root `README.md`.

Keep every language factually identical to the code — no locale may claim a
feature that doesn't exist.

## Commit / PR

- Small, focused PRs. Describe the change and link the issue.
- Update `README.md`, `ARCHITECTURE.md`, and `.env.example` when behavior or
  configuration changes.
