> **Languages:** [English](../en/adding-an-agent.md) · [Русский](../ru/adding-an-agent.md) · [Türkçe](../tr/adding-an-agent.md) · [Azərbaycanca](../az/adding-an-agent.md)

# Agent əlavə etmək

Yeni agent əlavə etmək bir interfeysi implement edən yeni paket və bir registry sətiri deməkdir. Bu səhifə müqaviləni və minimal adapteri göstərir.

## `Agent` interfeysi

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

## Yayılacaq hadisələr

CLI-ın çıxışını bu normallaşdırılmış hadisələrə xəritələyin (UI-ın anladığı yeganə lüğət):

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`, `ToolResult`, `ApprovalRequest`, `Progress`, `Result{Summary, Duration, CostUSD, SessionID}`, `Error`.

Zəngin adapter tam dəsti yayır. Sadə adapter `AssistantText` yayır və fayl dəyişikliklərinin git qatının iş sonu diff-i ilə göstərilməsinə imkan verir.

## Mətn + iş sonu diff-i geri düşməsi

Əgər CLI-ın strukturlaşdırılmış axını yoxdursa, diff-ləri təhlil etməyiniz **lazım deyil**. Paylaşılan `agent.StreamLines` köməkçisi əmri işə salır, hər stdout sətirini sizin funksiyanızla hadisələrə xəritələyir və həmişə terminal `Result` (və ya `Error`) ilə bitir. `FileEditEvents: false` təyin edin; fayl dəyişiklikləri yenə görünəcək, çünki git qatı iş sonunda branch-ı diff edir.

## Minimal adapter

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

## Qeydiyyatdan keçirin

`cmd/telepath/main.go`-da bir case əlavə edin:

```go
case "myagent":
    reg.Register(myagent.New(cfg.MyAgentBin))
```

və (istəyə görə) binary override üçün `internal/config`-də `MYAGENT_BIN` sahəsi. Sonra agentinizi `ENABLED_AGENTS`-ə əlavə edin və [agentlər](agents.md) cədvəlində sənədləşdirin.

## İnteraktiv təsdiq (istəyə görə)

İnteraktiv təsdiqi yalnız CLI-ınız icazə hook-unu təqdim edirsə (Claude-un MCP permission-prompt aləti kimi) implement edin. Etmirsə, `InteractiveApproval: false` buraxın; istifadəçi `interactive` seçəndə Telepath avtomatik `allowlist`-ə keçəcək.

## Əvvəlcə real bayraqları yoxlayın

CLI bayraqları tez-tez dəyişir. Kodlaşdırmadan əvvəl alətin `--help`-ini işlədin və sənədlərini oxuyun, və **real hazırkı** interfeysə uyğun implement edin — təxmin edilmiş bayraqlara inanmayın.

---

<!-- prev-next -->
[← Əvvəlki: Docker](./docker.md)  ·  [Növbəti: Təhlükəsizlik →](./security.md)
