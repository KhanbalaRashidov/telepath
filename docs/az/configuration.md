> **Languages:** [English](../en/configuration.md) · [Русский](../ru/configuration.md) · [Türkçe](../tr/configuration.md) · [Azərbaycanca](../az/configuration.md)

# Konfiqurasiya

Telepath konfiqurasiyanı dörd mənbədən oxuyur. Prioritet, ən yüksəkdən aşağıya doğru, belədir:

**CLI bayraqları > mühit dəyişənləri > `.env` faylı > daxili standart dəyərlər**

Hər mühit dəyişəninin `--kebab-case` bayraq ekvivalenti var. `.env` faylı yalnız real mühitdə artıq təyin olunmamış dəyərləri doldurur; bayraqları görmək üçün `telepath --help` işlədin.

## İstinad

Məcburi dəyişənlər ✅ ilə qeyd olunub. Mötərizədəki dəyərlər standart dəyərlərdir.

| Mühit dəyişəni | CLI bayrağı | Standart | Məcburi | Təsvir |
|----------------|-------------|----------|:-------:|--------|
| `TELEGRAM_BOT_TOKEN` | `--telegram-bot-token` | — | ✅ | @BotFather-dən bot token-i. |
| `ALLOWED_CHAT_IDS` | `--allowed-chat-ids` | — | ✅ | Botdan istifadə etməyə icazəli vergüllə ayrılmış chat ID-ləri. Telepath ən azı bir ID olmadan başlamaqdan imtina edir. |
| `PROJECTS_DIR` | `--projects-dir` | `.` | | Agentin işlədiyi qovluq. `per-task` branch rejimi üçün git repozitori olmalıdır. |
| `DEFAULT_AGENT` | `--default-agent` | `claude` | | Chat üçün standart agent. `ENABLED_AGENTS` daxilində olmalıdır. |
| `ENABLED_AGENTS` | `--enabled-agents` | `claude,gemini,codex,aider` | | Vergüllə ayrılmış aktiv agentlər. |
| `PERMISSION_MODE` | `--permission-mode` | `auto` | | `auto` \| `allowlist` \| `interactive`. |
| `APPROVAL_TIMEOUT_SECONDS` | `--approval-timeout-seconds` | `120` | | İnteraktiv təsdiqin nə qədər gözləyəcəyi, bundan sonra **rədd edilir**. |
| `BRANCH_MODE` | `--branch-mode` | `per-task` | | `off` \| `per-task`. |
| `BRANCH_PREFIX` | `--branch-prefix` | `telepath/` | | Tapşırıq branch adlarının prefiksi. |
| `BRANCH_COMMIT` | `--branch-commit` | `true` | | Agentin dəyişikliklərini tapşırıq branch-ında commit etmək. |
| `BRANCH_STASH_DIRTY` | `--branch-stash-dirty` | `false` | | Tapşırığa başlamaqdan imtina etmək əvəzinə çirkli working tree-ni stash etmək. |
| `MAX_TURNS` | `--max-turns` | `0` | | Agent növbələrinin maksimum sayı (`0` = CLI-nin öz standart dəyəri). |
| `MODEL` | `--model` | *(boş)* | | Agentə ötürülən model override-u (`""` = CLI-nin standart modeli). |
| `CLAUDE_BIN` | `--claude-bin` | *(boş)* | | PATH-də deyilsə `claude` binary-sinin yolu. |
| `GEMINI_BIN` | `--gemini-bin` | *(boş)* | | PATH-də deyilsə `gemini` binary-sinin yolu. |
| `CODEX_BIN` | `--codex-bin` | *(boş)* | | PATH-də deyilsə `codex` binary-sinin yolu. |
| `AIDER_BIN` | `--aider-bin` | *(boş)* | | PATH-də deyilsə `aider` binary-sinin yolu. |
| `GIT_BIN` | `--git-bin` | *(boş)* | | PATH-də deyilsə `git` binary-sinin yolu. |
| `LOG_LEVEL` | `--log-level` | `info` | | `debug` \| `info` \| `warn` \| `error`. |

Həmçinin çap edib çıxan iki meta-bayraq var: `--version` və `--help` (həmçinin `./.env`-dən başqa `.env` faylına işarə etmək üçün `--env-file`).

## Qeydlər

- **Boolean dəyərlər** `1/true/yes/on` və `0/false/no/off` qəbul edir.
- **`ALLOWED_CHAT_IDS`** sərt təhlükəsizlik sərhədidir. Hər yeniləmə — mesajlar və düymə basışları — bununla yoxlanılır; allowlist-də olmayan chat-lar nəzərə alınmır.
- **`DEFAULT_AGENT` `ENABLED_AGENTS` daxilində olmalıdır**, əks halda başlanğıc aydın mesajla uğursuz olur. Eləcə də `PERMISSION_MODE` və `BRANCH_MODE` düzgün olmalıdır.
- **`PROJECTS_DIR`** mövcud olmalı və qovluq olmalıdır. Bütün chat-lar bu bir working tree-ni paylaşır (bax: [Branch iş axını](branch-workflow.md)).
- Hər chat üçün parametrlər (`/use`, `/perm`, `/branch`) yalnız həmin chat üçün bu qlobal standart dəyərləri əvəz edir və proses yenidən başlayana qədər yaddaşda saxlanılır.

## Nümunə `.env`

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
