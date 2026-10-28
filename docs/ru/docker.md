> **Languages:** [English](../en/docker.md) · [Русский](../ru/docker.md) · [Türkçe](../tr/docker.md) · [Azərbaycanca](../az/docker.md)

# Docker

Telepath поставляется с двумя Dockerfile.

## Самодостаточный образ (`Dockerfile`, по умолчанию)

Многоэтапная сборка: на первом этапе собирается статический бинарный файл Go (CGO отключён),
а финальный этап — это облегчённая базовая Node-среда с установленными CLI агентов и git,
плюс бинарный файл telepath.

- **Зачем Node?** Claude Code, Gemini CLI и Codex CLI распространяются как npm-пакеты,
  поэтому для их запуска нужен Node.
- **Размер.** Этот образ больше (~200MB+), потому что включает Node и CLI агентов.
  Это цена за то, что «работает из коробки».
- Aider — это Python-пакет; если он нужен, добавьте его в образ самостоятельно.

```bash
docker build -t telepath:latest -f Dockerfile .
docker run --rm -it \
  -e TELEGRAM_BOT_TOKEN=xxxx \
  -e ALLOWED_CHAT_IDS=123456789 \
  -e PROJECTS_DIR=/work/project \
  -v /path/to/your/repo:/work/project \
  telepath:latest
```

## Облегчённый образ (`Dockerfile.slim`, маленький и быстрый)

Только бинарный файл telepath и git на минимальной базе Alpine — **без Node, без
агентов**. Принесите собственные CLI агентов: смонтируйте их и укажите на них переменные окружения
`*_BIN`, либо используйте этот образ для `BRANCH_MODE=off` с хостом, на котором они уже есть.

```bash
docker build -t telepath:slim -f Dockerfile.slim .
```

## docker compose

`docker-compose.yml` запускает самодостаточный образ, читает файл `.env`, перезапускается,
пока не будет остановлен, и монтирует вашу папку проектов:

```bash
cp .env.example .env     # fill in token, chat IDs
# set PROJECTS_DIR on the host; it is mounted to /work/project in the container
docker compose up -d
docker compose logs -f
```

Внутри контейнера `PROJECTS_DIR` установлен в `/work/project` — туда монтируется
ваш репозиторий.

## Где хранятся ключи и токены

- **Токен Telegram-бота** и **разрешённые chat ID** — это конфигурация Telepath;
  передавайте их через env-файл / `-e`.
- **Аутентификация агентов** (API-ключи / логины) принадлежит CLI агентов, а не Telepath.
  Передавайте их через собственные переменные окружения агентов или монтируйте их каталоги
  конфигурации (например, `-v ~/.claude:/root/.claude`). См. документацию
  каждого агента.
