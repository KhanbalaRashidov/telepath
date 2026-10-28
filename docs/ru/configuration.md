> **Languages:** [English](../en/configuration.md) · [Русский](../ru/configuration.md) · [Türkçe](../tr/configuration.md) · [Azərbaycanca](../az/configuration.md)

# Конфигурация

Telepath читает конфигурацию из четырёх источников. Приоритет, от высшего к низшему:

**Флаги CLI > переменные окружения > файл `.env` > встроенные значения по умолчанию**

Каждой переменной окружения соответствует флаг в формате `--kebab-case`. Файл `.env`
заполняет только те значения, которые ещё не заданы в реальном окружении; запустите
`telepath --help`, чтобы увидеть флаги.

## Справочник

Обязательные переменные отмечены ✅. Значения в скобках — значения по умолчанию.

| Переменная окружения | Флаг CLI | По умолчанию | Обязательна | Описание |
|----------------------|----------|--------------|:-----------:|----------|
| `TELEGRAM_BOT_TOKEN` | `--telegram-bot-token` | — | ✅ | Токен бота от @BotFather. |
| `ALLOWED_CHAT_IDS` | `--allowed-chat-ids` | — | ✅ | Chat ID через запятую, которым разрешено использовать бота. Telepath не запустится без хотя бы одного. |
| `PROJECTS_DIR` | `--projects-dir` | `.` | | Папка, в которой работает агент. Должна быть git-репозиторием для режима веток `per-task`. |
| `DEFAULT_AGENT` | `--default-agent` | `claude` | | Агент по умолчанию для чата. Должен быть указан в `ENABLED_AGENTS`. |
| `ENABLED_AGENTS` | `--enabled-agents` | `claude,gemini,codex,aider` | | Включённые агенты через запятую. |
| `PERMISSION_MODE` | `--permission-mode` | `auto` | | `auto` \| `allowlist` \| `interactive`. |
| `APPROVAL_TIMEOUT_SECONDS` | `--approval-timeout-seconds` | `120` | | Сколько ждёт интерактивное подтверждение, прежде чем будет **отклонено**. |
| `BRANCH_MODE` | `--branch-mode` | `per-task` | | `off` \| `per-task`. |
| `BRANCH_PREFIX` | `--branch-prefix` | `telepath/` | | Префикс имён веток задач. |
| `BRANCH_COMMIT` | `--branch-commit` | `true` | | Коммитить изменения агента в ветку задачи. |
| `BRANCH_STASH_DIRTY` | `--branch-stash-dirty` | `false` | | Отложить (stash) незакоммиченное рабочее дерево вместо отказа от запуска задачи. |
| `MAX_TURNS` | `--max-turns` | `0` | | Максимум ходов агента (`0` = значение CLI по умолчанию). |
| `MODEL` | `--model` | *(пусто)* | | Переопределение модели, передаваемое агенту (`""` = модель CLI по умолчанию). |
| `CLAUDE_BIN` | `--claude-bin` | *(пусто)* | | Путь к бинарному файлу `claude`, если его нет в PATH. |
| `GEMINI_BIN` | `--gemini-bin` | *(пусто)* | | Путь к бинарному файлу `gemini`, если его нет в PATH. |
| `CODEX_BIN` | `--codex-bin` | *(пусто)* | | Путь к бинарному файлу `codex`, если его нет в PATH. |
| `AIDER_BIN` | `--aider-bin` | *(пусто)* | | Путь к бинарному файлу `aider`, если его нет в PATH. |
| `GIT_BIN` | `--git-bin` | *(пусто)* | | Путь к бинарному файлу `git`, если его нет в PATH. |
| `LOG_LEVEL` | `--log-level` | `info` | | `debug` \| `info` \| `warn` \| `error`. |

Есть также два мета-флага, которые выводят информацию и завершают работу: `--version` и `--help`
(а также `--env-file`, чтобы указать `.env` не из `./.env`).

## Примечания

- **Булевы значения** принимают `1/true/yes/on` и `0/false/no/off`.
- **`ALLOWED_CHAT_IDS`** — это жёсткая граница безопасности. Каждое обновление — сообщения
  и нажатия кнопок — проверяется по этому списку; чаты вне списка игнорируются.
- **`DEFAULT_AGENT` должен быть в `ENABLED_AGENTS`**, иначе запуск завершится с понятным
  сообщением. Аналогично `PERMISSION_MODE` и `BRANCH_MODE` должны иметь допустимые значения.
- **`PROJECTS_DIR`** должен существовать и быть директорией. Все чаты используют одно
  общее рабочее дерево (см. [Рабочий процесс с ветками](branch-workflow.md)).
- Настройки для отдельного чата (`/use`, `/perm`, `/branch`) переопределяют эти глобальные
  значения по умолчанию только для этого чата и хранятся в памяти до перезапуска процесса.

## Пример `.env`

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789,987654321
PROJECTS_DIR=/srv/projects/myrepo
DEFAULT_AGENT=claude
ENABLED_AGENTS=claude,gemini,codex,aider
PERMISSION_MODE=interactive
APPROVAL_TIMEOUT_SECONDS=120
BRANCH_MODE=per-task
BRANCH_PREFIX=telepath/
BRANCH_COMMIT=true
BRANCH_STASH_DIRTY=false
LOG_LEVEL=info
```
