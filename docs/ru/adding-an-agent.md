> **Languages:** [English](../en/adding-an-agent.md) · [Русский](../ru/adding-an-agent.md) · [Türkçe](../tr/adding-an-agent.md) · [Azərbaycanca](../az/adding-an-agent.md)

# Добавление агента

Добавление нового агента — это новый пакет, реализующий один интерфейс, плюс одна строка
в реестре. Эта страница показывает контракт и минимальный адаптер.

## Интерфейс `Agent`

```go
type Agent interface {
    Name() string
    Available(ctx context.Context) bool
    Capabilities() Capabilities
    Run(ctx context.Context, t Task) (<-chan Event, error)
    Resume(ctx context.Context, sessionID, msg string) (<-chan Event, error)
}

type Capabilities struct {
    StructuredStream    bool // emits a machine-readable event stream
    Sessions            bool // reports a session id
    Resume              bool // can continue a previous session
    InteractiveApproval bool // supports live per-action approval (e.g. MCP)
    FileEditEvents      bool // emits per-file edit events during the run
}

type Task struct {
    Prompt, ProjectDir string
    PermissionMode     PermissionMode
    ApprovalHook       ApprovalFunc // non-nil only in interactive mode
    MaxTurns           int
    Model              string
}
```

## События, которые нужно отправлять

Сопоставьте вывод CLI с этими нормализованными событиями (единственным словарём, который
понимает UI):

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`,
`ToolResult`, `ApprovalRequest`, `Progress`, `Result{Summary, Duration,
CostUSD, SessionID}`, `Error`.

Развитый адаптер выдаёт полный набор. Простой адаптер выдаёт `AssistantText` и
позволяет диффу в конце запуска от git-слоя показать изменения файлов.

## Запасной вариант: текст + диф в конце запуска

Если у CLI нет структурированного потока, вам **не нужно** разбирать диффы. Общий
помощник `agent.StreamLines` запускает команду, сопоставляет каждую строку stdout
с событиями через вашу функцию и всегда завершается терминальным `Result` (или
`Error`). Установите `FileEditEvents: false`; изменения файлов всё равно появятся, потому что
git-слой в конце сравнивает ветку.

## Минимальный адаптер

```go
package myagent

import (
    "context"
    "strings"

    "github.com/telepath/telepath/internal/agent"
    "github.com/telepath/telepath/internal/permission"
)

const binaryName = "myagent"

type MyAgent struct{ bin string }

func New(bin string) *MyAgent { return &MyAgent{bin: bin} }

func (a *MyAgent) Name() string { return "myagent" }

func (a *MyAgent) Available(ctx context.Context) bool {
    return agent.IsInstalled(binaryName, a.bin)
}

func (a *MyAgent) Capabilities() agent.Capabilities {
    return agent.Capabilities{StructuredStream: false}
}

func (a *MyAgent) Run(ctx context.Context, t agent.Task) (<-chan agent.Event, error) {
    bin := agent.ResolveBinary(binaryName, a.bin)
    args := []string{"--message", t.Prompt}
    // Map the permission mode to real flags; keep the agent's own git disabled.
    _ = permission.AiderArgs // see internal/permission/policy.go for the pattern
    return agent.StreamLines(ctx, t.ProjectDir, bin, args, func(line string) []agent.Event {
        if s := strings.TrimSpace(line); s != "" {
            return []agent.Event{agent.AssistantText{Text: s}}
        }
        return nil
    }), nil
}

func (a *MyAgent) Resume(ctx context.Context, sessionID, msg string) (<-chan agent.Event, error) {
    return nil, agent.ErrResumeUnsupported
}
```

## Регистрация

Добавьте один case в `cmd/telepath/main.go`:

```go
case "myagent":
    reg.Register(myagent.New(cfg.MyAgentBin))
```

и (по желанию) поле `MYAGENT_BIN` в `internal/config`, если нужно переопределение
бинарного файла. Затем добавьте своего агента в `ENABLED_AGENTS` и опишите его в таблице
на странице [агенты](agents.md).

## Интерактивное подтверждение (необязательно)

Реализуйте интерактивное подтверждение, только если ваш CLI предоставляет хук разрешений (как
инструмент MCP prompt у Claude). Если его нет, оставьте
`InteractiveApproval: false`; Telepath автоматически переключится на
`allowlist`, когда пользователь выберет interactive.

## Сначала проверьте реальные флаги

Флаги CLI часто меняются. Перед написанием кода запустите `--help` инструмента и прочитайте
его документацию, и реализуйте под **реальный** текущий интерфейс — не доверяйте
угаданным флагам.
