# Telepath Makefile. CGO is disabled everywhere for clean cross-compilation.
BINARY      := telepath
PKG         := ./cmd/telepath
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X main.version=$(VERSION)
GO          := go
export CGO_ENABLED=0

.PHONY: all build run test vet lint fmt tidy clean cross docker docker-slim

all: build

build: ## Build the binary for the host platform
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

run: ## Build and run (reads .env)
	$(GO) run $(PKG)

test: ## Run all tests
	$(GO) test ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint (must be installed)
	golangci-lint run

fmt: ## Format the code
	gofmt -s -w .

tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist bin

# Cross-compile the common targets into ./bin.
cross: ## Cross-compile windows/linux/darwin x amd64/arm64
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-amd64.exe $(PKG)
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 $(PKG)
	GOOS=linux   GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 $(PKG)
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-amd64 $(PKG)
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-arm64 $(PKG)

docker: ## Build the self-contained image (bundles Node + agent CLIs)
	docker build -t telepath:latest -f Dockerfile .

docker-slim: ## Build the slim image (bring your own agent CLIs)
	docker build -t telepath:slim -f Dockerfile.slim .
