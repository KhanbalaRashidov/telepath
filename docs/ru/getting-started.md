> **Languages:** [English](../en/getting-started.md) · [Русский](../ru/getting-started.md) · [Türkçe](../tr/getting-started.md) · [Azərbaycanca](../az/getting-started.md)

# Начало работы

Это пошаговое руководство проведёт вас от нуля до первой задачи в Telegram.

## 1. Требования

- **Go 1.26+** (только для сборки из исходников) — <https://go.dev/dl/>
- **git** в PATH.
- Хотя бы один **CLI агента**, установленный на хосте. Нужны только те, которые вы
  планируете включить:

  | Агент | Установка |
  |-------|-----------|
  | Claude Code | `npm install -g @anthropic-ai/claude-code` |
  | Gemini CLI | `npm install -g @google/gemini-cli` |
  | OpenAI Codex CLI | `npm install -g @openai/codex` |
  | Aider | `pipx install aider-chat` (или `pip install aider-chat`) |

  Каждый агент — это **внешняя зависимость** со своей аутентификацией (API-ключ
  или логин), которую вы настраиваете по его собственной документации. Telepath
  никогда не работает с этими учётными данными. После запуска Telepath команда `/agents`
  покажет, какие CLI он нашёл.

## 2. Установка Telepath

### Готовый бинарный файл

Скачайте архив для вашей ОС/архитектуры со страницы Releases проекта, распакуйте
его и проверьте, что он запускается:

```bash
# Linux / macOS
chmod +x telepath
./telepath --version
```
```powershell
# Windows (PowerShell)
.\telepath.exe --version
```

### Из исходников через `go install`

```bash
go install github.com/telepath/telepath/cmd/telepath@latest
```

> Если репозиторий был форкнут или переименован, используйте его реальный путь модуля
> вместо `github.com/telepath/telepath` (например, `github.com/KhanbalaRashidov/telepath`).

### Из исходников через Makefile

```bash
git clone https://github.com/KhanbalaRashidov/telepath
cd telepath
make build      # produces ./telepath (or telepath.exe on Windows)
```

## 3. Создание Telegram-бота

1. Откройте [@BotFather](https://t.me/BotFather) в Telegram.
2. Отправьте `/newbot` и следуйте инструкциям.
3. Скопируйте **токен бота**, который он вам выдаст.

## 4. Узнайте свой chat ID

Напишите [@userinfobot](https://t.me/userinfobot); он ответит вашим числовым
chat ID. Telepath обслуживает только те чаты, которые вы явно добавили в allowlist.

## 5. Минимальный `.env`

Скопируйте пример и заполните три обязательных параметра:

```bash
cp .env.example .env
```

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789
PROJECTS_DIR=/path/to/your/git/repo
```

`PROJECTS_DIR` — это папка, в которой работает агент. Для рабочего процесса по умолчанию
(ветка на задачу) она **должна быть git-репозиторием**. Все опции см. в разделе
[Конфигурация](configuration.md).

## 6. Первый запуск

```bash
./telepath
```

При старте Telepath проверяет конфигурацию (и отказывается запускаться без токена
бота и хотя бы одного разрешённого chat ID), после чего начинает long-polling Telegram.

## 7. Ваша первая задача

В вашем чате с ботом в Telegram:

1. Отправьте `/help`, чтобы увидеть команды.
2. Отправьте `/agents`, чтобы убедиться, что ваш агент установлен.
3. Просто опишите задачу обычным текстом, например:

   ```
   add a /health endpoint that returns 200 OK and a test for it
   ```

Telepath создаёт ветку вида `telepath/add-a-health-endpoint-260115-1430`, запускает
там агента, транслирует живое статус-сообщение и диффы по файлам, коммитит результат в эту
ветку и в завершение отправляет `git diff --stat` с полным диффом в виде документа `.diff`
и кнопки **Keep branch / Delete branch & restore base**.

## Настройка Slack (необязательно)

Предпочитаете Slack или хотите оба транспорта? Telepath также работает через Slack по
**Socket Mode** (публичный URL не нужен). Сравнение см. в разделе [Транспорты](transports.md).

1. **Создайте Slack-приложение** на <https://api.slack.com/apps> → *Create New App* →
   *From scratch*. Укажите название и рабочее пространство.
2. **Включите Socket Mode**: *Settings → Socket Mode → Enable*. Это создаст **токен уровня
   приложения** (`xapp-…`) с областью `connections:write` — скопируйте его.
3. **Добавьте scope-ы бота**: *Features → OAuth & Permissions → Scopes → Bot Token
   Scopes*, добавьте: `app_mentions:read`, `chat:write`, `commands`, `files:write`,
   `im:history`, `im:read`, `channels:history`, `groups:history`.
4. **Создайте slash-команду**: *Features → Slash Commands → Create New
   Command* → `/telepath` (описание любое; Request URL в Socket Mode игнорируется, но
   Slack требует значение — подойдёт любой placeholder).
5. **Подпишитесь на события**: *Features → Event Subscriptions → Enable*, затем в
   *Subscribe to bot events* добавьте `app_mention` и `message.im`.
6. **Установите в рабочее пространство**: *OAuth & Permissions → Install to Workspace*.
   Скопируйте **Bot User OAuth Token** (`xoxb-…`).
7. **Пригласите бота** в любой канал, где хотите его использовать (`/invite @YourApp`),
   или просто напишите ему в личные сообщения.
8. **Найдите ID** для allowlist: ID пользователя (профиль → *Copy member ID*,
   `U…`) и/или ID канала (детали канала, `C…`).

Минимальный `.env` для Slack:

```dotenv
TRANSPORT=slack
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token
ALLOWED_SLACK_USER_IDS=U0123456789
PROJECTS_DIR=/path/to/your/git/repo
```

Запустите `./telepath`, затем в Slack упомяните бота или используйте slash-команду:

```
@Telepath add a /health endpoint that returns 200 OK and a test for it
/telepath help
/telepath claude fix the null check in parser.go
```

Чтобы запустить **оба** транспорта одновременно, установите `TRANSPORT=both` и укажите
учётные данные и для Telegram, и для Slack.

Далее прочитайте [Транспорты](transports.md),
[Команды Telegram](telegram-commands.md), [Разрешения](permissions.md) и
[Рабочий процесс с ветками](branch-workflow.md).

---

<!-- prev-next -->
[← Назад: Главная](./index.md)  ·  [Вперёд: Конфигурация →](./configuration.md)
