> **Languages:** [English](../en/architecture.md) · [Русский](../ru/architecture.md) · [Türkçe](../tr/architecture.md) · [Azərbaycanca](../az/architecture.md)

# Arxitektura

Bu qısa istiqamətləndirmədir. Tam dizayn kök [ARCHITECTURE.md](../../ARCHITECTURE.md) faylındadır; bu səhifə onu ümumiləşdirir və geri istinad edir.

## Pipeline

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

## Transport abstraksiyası

Mesajlaşma platforması `internal/transport` interfeysinin arxasında yerləşir. Daxil olan mesajlarda
transport platforma hadisələrini `InboundMessage{Transport, ConversationID, ThreadID, UserID, Text}`
formatına çevirir və engine-ə ötürür. Çıxış istiqamətində engine transportdan asılı olmayan mesaj
`Doc`-larından qurulmuş hər tapşırıq üzrə `Session`-u (status / notice / file-diff / approval /
result) idarə edir; hər transport bunları öz markup-ında render edir (Telegram HTML + inline
klaviaturalar, Slack mrkdwn + Block Kit) və fayl yükləmə mexanizminə özü sahiblik edir.

Engine, git, agent, permission və session qatları **transport və agentdən müstəqildir**: yalnız
`Event` axınını, `Agent` interfeysini və `Transport`/`Session` interfeyslərini istehlak edirlər.
Çoxlu transport dəstəyi üçün yeganə vəziyyət dəyişikliyi sessiya açarıdır: sadə chat id əvəzinə
`transport:conversation`. Bax: [Transportlar](transports.md).

## Normallaşdırılmış hadisə modeli

Adapterlər satıcı CLI-nı bu hadisələrə çevirir:

`AssistantText`, `ToolUse`, `FileEdit{Path, UnifiedDiff}`, `Command`, `ToolResult`, `ApprovalRequest`, `Progress`, `BranchInfo`, `Committed`, `Result{Summary, Duration, CostUSD, SessionID}`, `Error`.

Zəngin agentlər (Claude) tam dəsti yayır. Zəif agentlər `AssistantText`-ə və git qatının iş sonu diff-inə enir, beləliklə `FileEdit` hadisələri olmadan da fayl dəyişiklikləri görünür.

## Paket xəritəsi

| Paket | Məsuliyyət |
|-------|-----------|
| `cmd/telepath` | Giriş nöqtəsi: konfiqurasiya, registry bağlantıları, zərif dayandırma. |
| `internal/config` | Konfiqurasiyanı yükləmək/yoxlamaq (bayraqlar > mühit > .env > standart). |
| `internal/agent` | `Agent` interfeysi, `Event` tipləri, registry, exec köməkçiləri, proses ağacının öldürülməsi. |
| `internal/agent/{claude,gemini,codex,aider}` | Hər agent üçün ayrıca adapter paketi. |
| `internal/permission` | `policy` (rejim → bayraqlar), `broker` (gözləmə/həll/timeout), `mcpserver` (təsdiq aləti). |
| `internal/git` | `guard` (təhlükəsizlik nöqtəsi), `git` (oxuma əməliyyatları), `branch` (hər tapşırıq üzrə iş axını). |
| `internal/transport` | Platforma abstraksiyası: `Transport`/`Session`/`Router` interfeysləri, mesaj `Doc` modeli, ümumi status render-i və throttle. |
| `internal/engine` | Transportdan asılı olmayan əmr göndərmə + tapşırıq pipeline-ı (`Router`-i implement edir). |
| `internal/telegram` | Telegram transportu: long-polling, HTML render, inline klaviaturalar. |
| `internal/slack` | Slack transportu: Socket Mode, mrkdwn render, Block Kit, xarici fayl yükləmə. |
| `internal/session` | `Store` interfeysi arxasında hər söhbət üzrə vəziyyət (v1-də yaddaşda). |
| `internal/logx` | Strukturlaşdırılmış loglama (slog). |

## MCP icazə serveri

İnteraktiv rejim üçün Claude ilə Telepath loopback-də (təsadüfi port) minimal **stateless Streamable-HTTP MCP server** işlədir. O, bir alət təqdim edir, `mcp__telepath__approve`, hər sorğunu broker-ə ötürür və istifadəçinin qərarını qaytarır. Yalnız POST sorğu/cavab işləyir (server tərəfindən başladılan SSE axını yoxdur); icazə-prompt axını üçün bu kifayətdir. Hər interaktiv icra üçün təzə server başlayır, belə ki, prosesllərarası IPC yoxdur və autentifikasiya səthi yoxdur. Bax: [İcazələr](permissions.md).

## Qlobal tapşırıq ardıcıllığı

Bütün chat-lar bir `PROJECTS_DIR` working tree-ni paylaşır, git branch checkout-u isə repo üzrə qlobaldır. Buna görə Telepath bütün chat-lar üzrə **eyni anda bir tapşırıq** işlədir (qlobal bir slot və hər chat üzrə bir işləyən tapşırıq). İş sahəsi məşğul olduqda gələn tapşırıq qısa "yenidən cəhd edin" cavabı alır.

## Platformalar arası

Binary-lər `exec.LookPath` ilə tapılır (Windows-da `.exe`/`PATHEXT`-i idarə edir). Agentlər öz proses qrupunda başladılır və `/cancel`-da bütöv ağac kimi öldürülür (Windows-da `taskkill /F /T`; Unix-də proses qrupu üzrə `SIGKILL`).

---

<!-- prev-next -->
[← Əvvəlki: Branch iş axını](./branch-workflow.md)  ·  [Növbəti: Docker →](./docker.md)
