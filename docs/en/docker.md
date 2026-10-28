> **Languages:** [English](../en/docker.md) · [Русский](../ru/docker.md) · [Türkçe](../tr/docker.md) · [Azərbaycanca](../az/docker.md)

# Docker

Telepath ships two Dockerfiles.

## Self-contained image (`Dockerfile`, default)

A multi-stage build: stage one builds the static Go binary (CGO off), and the
final stage is a slim Node base with the agent CLIs and git installed, plus the
telepath binary.

- **Why Node?** Claude Code, Gemini CLI, and Codex CLI ship as npm packages, so
  the runtime needs Node to run them.
- **Size.** This image is larger (~200MB+) because it bundles Node and the agent
  CLIs. That's the price of "works out of the box."
- Aider is a Python package; add it to the image yourself if you need it.

```bash
docker build -t telepath:latest -f Dockerfile .
docker run --rm -it \
  -e TELEGRAM_BOT_TOKEN=xxxx \
  -e ALLOWED_CHAT_IDS=123456789 \
  -e PROJECTS_DIR=/work/project \
  -v /path/to/your/repo:/work/project \
  telepath:latest
```

## Slim image (`Dockerfile.slim`, small & fast)

Only the telepath binary plus git on a minimal Alpine base — **no Node, no
agents**. Bring your own agent CLIs: mount them in and point the `*_BIN` env vars
at them, or use this for `BRANCH_MODE=off` against a host that has them.

```bash
docker build -t telepath:slim -f Dockerfile.slim .
```

## docker compose

`docker-compose.yml` runs the self-contained image, reads a `.env` file,
restarts unless stopped, and mounts your projects folder:

```bash
cp .env.example .env     # fill in token, chat IDs
# set PROJECTS_DIR on the host; it is mounted to /work/project in the container
docker compose up -d
docker compose logs -f
```

Inside the container, `PROJECTS_DIR` is set to `/work/project`, which is where
your repository is mounted.

## Where keys and tokens go

- The **Telegram bot token** and **allowed chat IDs** are Telepath config —
  pass them via the env file / `-e`.
- **Agent authentication** (API keys / logins) belongs to the agent CLIs, not to
  Telepath. Provide them via the agents' own environment variables or mount their
  config directories (for example `-v ~/.claude:/root/.claude`). See each
  agent's documentation.

---

<!-- prev-next -->
[← Previous: Architecture](./architecture.md)  ·  [Next: Adding an agent →](./adding-an-agent.md)
