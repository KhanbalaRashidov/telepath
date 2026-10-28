# Telepath — self-contained image (DEFAULT).
# Bundles Node + the agent CLIs (claude/gemini/codex) + git, so it works out of
# the box. This image is larger (~250MB+) because of Node and the agents; for a
# small image where you provide the agent CLIs on the host, use Dockerfile.slim.

# --- stage 1: build the static Go binary -----------------------------------
FROM golang:1.26-alpine AS build
ENV CGO_ENABLED=0
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=docker
RUN go build -ldflags "-s -w -X main.version=${VERSION}" -o /out/telepath ./cmd/telepath

# --- stage 2: runtime with Node + agent CLIs --------------------------------
FROM node:22-bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install the npm-distributed agent CLIs. Pin or update as you prefer.
# (Aider is a Python package; install it separately if you need it.)
RUN npm install -g \
    @anthropic-ai/claude-code \
    @google/gemini-cli \
    @openai/codex \
    || echo "warning: one or more agent CLIs failed to install; they can still be mounted/installed at runtime"

COPY --from=build /out/telepath /usr/local/bin/telepath

# Agent auth (API keys / logins) and the Telegram token are provided via env.
WORKDIR /work
ENTRYPOINT ["telepath"]
