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

## Commit / PR

- Small, focused PRs. Describe the change and link the issue.
- Update `README.md`, `ARCHITECTURE.md`, and `.env.example` when behavior or
  configuration changes.
