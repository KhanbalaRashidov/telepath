> **Languages:** [English](../en/architecture.md) · [Русский](../ru/architecture.md) · [Türkçe](../tr/architecture.md) · [Azərbaycanca](../az/architecture.md)

# Архитектура

Это краткий обзор. Полная документация по проектированию находится в корневом файле
[ARCHITECTURE.md](../../ARCHITECTURE.md); эта страница её кратко описывает и ссылается на неё.

## Конвейер

```mermaid
flowchart LR
    TG["Telegram"] --> T["Transport"]
    SL["Slack"] --> T
    T -- "InboundMessage" --> E["Engine (command router + task pipeline)"]
    E --> S["Session store (per-conversation)"]
    E --> G["GitWorkspace (per-task branch)"]
    E --> A["Agent adapter<br/>claude · gemini · codex · aider"]
    A -- "normalized Event stream" --> E
    A -. "ApprovalRequest" .-> P["Permission broker<br/>+ in-process MCP server"]
    E -- "status · diffs · result" --> T
    P -- "Approve / Deny buttons" --> T
    T --> TG
    T --> SL
```

## Абстракция транспорта

Платформа обмена сообщениями скрыта за интерфейсом `internal/transport`. На вход транспорт
разбирает события платформы в `InboundMessage{Transport,
ConversationID, ThreadID, UserID, Text}` и передаёт его движку. На выход движок управляет
`Session` для каждой задачи (status / notice / file-diff / approval / result), построенным из
не зависящих от транспорта документов сообщений `Doc`; каждый транспорт отрисовывает их
в своей разметке (HTML + inline-клавиатуры для Telegram, mrkdwn + Block Kit для Slack) и сам
отвечает за механику загрузки файлов.

Слои движка, git, агента, разрешений и сессий **не зависят от транспорта и агента**: они
используют только поток `Event`, интерфейс `Agent` и интерфейсы `Transport`/`Session`.
Единственное изменение состояния для поддержки нескольких транспортов — ключ сессии:
`transport:conversation` вместо голого chat id. См. [Транспорты](transports.md).

## Модель нормализованных событий

Адаптеры переводят CLI вендора в эти события:

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`,
`ToolResult`, `ApprovalRequest`, `Progress`, `BranchInfo`, `Committed`,
`Result{Summary, Duration, CostUSD, SessionID}`, `Error`.

Развитые агенты (Claude) выдают полный набор. Более простые агенты деградируют до
`AssistantText` плюс диффа в конце запуска от git-слоя, поэтому изменения файлов
всё равно отображаются даже без событий `FileEdit`.

## Карта пакетов

| Пакет | Ответственность |
|-------|-----------------|
| `cmd/telepath` | Точка входа: конфигурация, связывание реестра, корректное завершение работы. |
| `internal/config` | Загрузка и проверка конфигурации (флаги > env > .env > значения по умолчанию). |
| `internal/agent` | Интерфейс `Agent`, типы `Event`, реестр, вспомогательные функции exec, завершение дерева процессов. |
| `internal/agent/{claude,gemini,codex,aider}` | Один пакет адаптера на каждого агента. |
| `internal/permission` | `policy` (режим → флаги), `broker` (await/resolve/timeout), `mcpserver` (инструмент подтверждения). |
| `internal/git` | `guard` (главная точка защиты безопасности), `git` (операции чтения), `branch` (рабочий процесс per-task). |
| `internal/transport` | Абстракция платформы: интерфейсы `Transport`/`Session`/`Router`, модель сообщений `Doc`, общий рендеринг статуса и throttling. |
| `internal/engine` | Не зависящая от транспорта диспетчеризация команд и конвейер задач (реализует `Router`). |
| `internal/telegram` | Транспорт Telegram: long-polling, HTML-рендеринг, inline-клавиатуры. |
| `internal/slack` | Транспорт Slack: Socket Mode, рендеринг mrkdwn, Block Kit, загрузка файлов. |
| `internal/session` | Состояние по диалогам за интерфейсом `Store` (в памяти в v1). |
| `internal/logx` | Структурированное логирование (slog). |

## MCP-сервер разрешений

Для интерактивного режима с Claude Telepath запускает минимальный **stateless
MCP-сервер Streamable-HTTP** на loopback (случайный порт). Он предоставляет один инструмент,
`mcp__telepath__approve`, который передаёт каждый запрос брокеру и возвращает
решение пользователя. Он обрабатывает только запросы/ответы POST (без серверного SSE-потока),
чего достаточно для потока запросов разрешений. Новый сервер запускается
для каждого интерактивного запуска, поэтому нет межпроцессного IPC и нет
поверхности аутентификации. См. [Разрешения](permissions.md).

## Глобальная сериализация задач

Все чаты используют один рабочий каталог `PROJECTS_DIR`, а переключение git-веток
глобально для репозитория. Поэтому Telepath выполняет **одну задачу за раз** во всех чатах,
защищённых одним глобальным слотом (плюс одна выполняющаяся задача на чат). Задача,
которая поступает, когда рабочее пространство занято, получает короткий ответ «попробуйте снова».

## Кроссплатформенность

Бинарные файлы обнаруживаются через `exec.LookPath` (обрабатывает `.exe`/`PATHEXT` в Windows).
Агенты запускаются в собственной группе процессов и при `/cancel` убиваются целиком вместе с деревом
процессов (`taskkill /F /T` в Windows; `SIGKILL` для группы процессов в Unix).

---

<!-- prev-next -->
[← Назад: Работа с ветками](./branch-workflow.md)  ·  [Вперёд: Docker →](./docker.md)
