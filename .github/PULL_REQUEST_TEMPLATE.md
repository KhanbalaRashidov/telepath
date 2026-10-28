<!-- Thanks for contributing to Telepath! -->

## What does this change?

<!-- A short description. Link any related issue: Closes #123 -->

## Checklist

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `gofmt -s -w .` applied (and `golangci-lint run` if available)
- [ ] Docs updated (README / ARCHITECTURE / .env.example) if behavior changed
- [ ] If this adds an agent: new package under `internal/agent/<name>/`, one
      registry line in `cmd/telepath/main.go`, and a row in the README table
