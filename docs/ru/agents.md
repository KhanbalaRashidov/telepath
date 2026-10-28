> **Languages:** [English](../en/agents.md) · [Русский](../ru/agents.md) · [Türkçe](../tr/agents.md) · [Azərbaycanca](../az/agents.md)

# Агенты

Telepath управляет несколькими CLI coding-агентов за одним общим интерфейсом. Каждый
адаптер переводит CLI конкретного вендора в один и тот же нормализованный поток событий,
поэтому слои Telegram и git никогда не видят деталей, специфичных для агента. Переключайте
агентов для каждого чата с помощью `/use` или запускайте одну задачу с конкретным агентом через
`/<agent> <task>`.

## Матрица возможностей

| Агент | Установка | Потоковая передача (события FileEdit) | Сессии / resume | Интерактивное подтверждение | Статус |
|-------|-----------|:------------------------------------:|:---------------:|:---------------------------:|--------|
| **claude** | `npm install -g @anthropic-ai/claude-code` | ✅ полная (stream-json, диффы по каждой правке) | ✅ | ✅ (инструмент MCP prompt) | Эталонный адаптер |
| **gemini** | `npm install -g @google/gemini-cli` | — (текст; правки через диff в конце запуска) | — | — → откат на `allowlist` | Поддерживается |
| **codex** | `npm install -g @openai/codex` | JSONL (по возможности; правки через диff в конце запуска) | — | — → откат на `allowlist` | Поддерживается |
| **aider** | `pipx install aider-chat` | — (текст; правки через диff в конце запуска) | — | — → откат на `allowlist` | Поддерживается |

`/agents` в Telegram выводит эту матрицу с актуальным статусом установки на вашем хосте.

## Точные вызовы

Это реальные команды, которые формирует каждый адаптер. Флаги в квадратных скобках
добавляются только при необходимости (`--max-turns` — когда `MAX_TURNS>0`, `--model` — когда
задан `MODEL`, флаги разрешений — в зависимости от режима).

### Claude Code (эталон)

```bash
claude -p "<task>" \
  --output-format stream-json --verbose --include-partial-messages \
  [--max-turns <N>] [--model <model>] \
  # auto:        --permission-mode bypassPermissions
  # allowlist:   --permission-mode default --allowedTools "<list>" --disallowedTools "<list>"
  # interactive: --permission-mode default \
  #              --permission-prompt-tool mcp__telepath__approve --mcp-config <file>
```

Resume:

```bash
claude --resume <session_id> -p "<message>" \
  --output-format stream-json --verbose --include-partial-messages \
  --permission-mode bypassPermissions
```

Claude выдаёт полный набор событий: текст ассистента, `tool_use` → `FileEdit`
(Edit/Write/MultiEdit, с unified diff), `Bash` → `Command`, `tool_result`,
и финальный `result` с итогом, длительностью, стоимостью и ID сессии.

### Gemini CLI

```bash
gemini -p "<task>" [-m <model>] \
  # auto:  --approval-mode=yolo
  # other: --approval-mode=default
```

Вывод Gemini читается как текст (его JSON-вывод был нестабильным между
релизами). Строки становятся `AssistantText`; изменения файлов отображаются диффом
в конце запуска, который формирует git-слой.

### OpenAI Codex CLI

```bash
codex exec --json \
  # auto:  --sandbox workspace-write --ask-for-approval never
  # other: --sandbox workspace-write --ask-for-approval on-request
  [--model <model>] "<task>"
```

> Старый флаг `--full-auto` устарел; Telepath использует `--sandbox
> workspace-write`. Codex выдаёт поток событий в формате JSONL, который разбирается
> по возможности в `AssistantText`/`Command`; изменения файлов также появляются
> в диффе в конце запуска.

### Aider

```bash
aider --message "<task>" --no-stream [--model <model>] \
  --yes --no-auto-commits --no-git
```

Собственная git-интеграция Aider **отключена** (`--no-auto-commits --no-git`), чтобы
управление версиями полностью принадлежало Telepath, и он работает без интерактива (`--message
--yes`). Вывод — текст; изменения файлов появляются в диффе в конце запуска.

## Честные ограничения (прочитайте)

- **Интерактивное подтверждение работает только с Claude** (нужен хук разрешений MCP).
  Для gemini/codex/aider выбор `interactive` автоматически переключается на `allowlist`,
  а ответ сообщает, какой режим был фактически использован.
- **gemini/codex/aider не выдают события `FileEdit` для каждой правки** во время запуска —
  их изменения файлов показываются диффом в конце запуска, который формирует git-слой.
- **Resume доступен только для Claude.** Остальные агенты возвращают «not supported» на `/resume`.
- **Разбор JSONL у Codex выполняется по возможности**, так как схема отличается в разных
  версиях; нераспознанные строки игнорируются (или показываются как текст).
- Каждый CLI агента — это **внешняя зависимость**, которую вы устанавливаете и аутентифицируете
  сами. Если CLI отсутствует, `/agents` покажет его как не установленный, а задачи
  для него будут отклоняться с понятным сообщением.

Чтобы написать собственный адаптер, см. [Добавление агента](adding-an-agent.md).

---

<!-- prev-next -->
[← Назад: Команды Telegram](./telegram-commands.md)  ·  [Вперёд: Разрешения →](./permissions.md)
