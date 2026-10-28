# Telepath Documentation

Welcome! Telepath is a self-hosted bridge between Telegram and a coding agent
(Claude Code, Gemini CLI, OpenAI Codex CLI, or Aider) running on your own
machine. These docs are available in four languages — pick yours:

| Language | Home |
|----------|------|
| 🇬🇧 English | [en/index.md](en/index.md) |
| 🇷🇺 Русский | [ru/index.md](ru/index.md) |
| 🇹🇷 Türkçe | [tr/index.md](tr/index.md) |
| 🇦🇿 Azərbaycanca | [az/index.md](az/index.md) |

English is the canonical source; the other languages are kept in sync with it.

> **Maintainer note:** examples use the module path `github.com/telepath/telepath`
> from `go.mod`. If you forked/renamed the repository, replace it everywhere with
> your real path (for example `github.com/KhanbalaRashidov/telepath`).

## Adding a language

Anyone can add a new language — see the short guide in
[CONTRIBUTING.md](../CONTRIBUTING.md#adding-a-documentation-language). In short:
copy `docs/en/` to `docs/<locale>/`, translate the prose (never the code),
and register the locale in `mkdocs.yml` and the switcher links.
