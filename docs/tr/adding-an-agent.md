> **Languages:** [English](../en/adding-an-agent.md) · [Русский](../ru/adding-an-agent.md) · [Türkçe](../tr/adding-an-agent.md) · [Azərbaycanca](../az/adding-an-agent.md)

# Ajan ekleme

Yeni bir ajan eklemek, tek bir arayüzü uygulayan yeni bir paket ve tek bir kayıt defteri satırı demektir. Bu sayfa sözleşmeyi ve minimal bir adaptörü gösterir.

## `Agent` arayüzü

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

## Üretilecek olaylar

CLI'nin çıktısını şu normalleştirilmiş olaylara eşleyin (arayüzün anladığı tek kelime dağarcığı):

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`, `ToolResult`, `ApprovalRequest`, `Progress`, `Result{Summary, Duration, CostUSD, SessionID}`, `Error`.

Zengin bir adaptör tam seti üretir. Basit bir adaptör `AssistantText` üretir ve dosya değişikliklerini git katmanının çalışma sonu diff'ine bırakır.

## Metin + çalışma sonu diff yedeği

CLI'nin yapılandırılmış bir akışı yoksa diff'leri ayrıştırmanız **gerekmez**. Ortak `agent.StreamLines` yardımcısı komutu çalıştırır, her stdout satırını sizin fonksiyonunuz aracılığıyla olaylara eşler ve her zaman bir son `Result` (veya `Error`) ile biter. `FileEditEvents: false` olarak ayarlayın; dosya değişiklikleri yine görünür, çünkü git katmanı en sonda branch'i diff'ler.

## Minimal adaptör

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

## Kaydedin

`cmd/telepath/main.go` dosyasına bir `case` ekleyin:

```go
case "myagent":
    reg.Register(myagent.New(cfg.MyAgentBin))
```

ve (isteğe bağlı olarak) bir ikili dosya geçersiz kılma istiyorsanız `internal/config` içine bir `MYAGENT_BIN` alanı ekleyin. Ardından ajanınızı `ENABLED_AGENTS` içine ekleyin ve [ajanlar](agents.md) tablosunda belgeleyin.

## Etkileşimli onay (isteğe bağlı)

Etkileşimli onayı yalnızca CLI'niz bir izin kancası sunuyorsa (Claude'un MCP izin-istemi aracı gibi) uygulayın. Sunmuyorsa `InteractiveApproval: false` bırakın; kullanıcı etkileşimli modu seçtiğinde Telepath otomatik olarak `allowlist`'e geri döner.

## Bayrakları önce doğrulayın

CLI bayrakları sık sık değişir. Kodlamadan önce aracın `--help` çıktısını çalıştırın ve dokümanlarını okuyun; tahmin edilen bayraklara değil, **gerçek** mevcut arayüze göre uygulayın.
