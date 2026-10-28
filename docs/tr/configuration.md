> **Languages:** [English](../en/configuration.md) · [Русский](../ru/configuration.md) · [Türkçe](../tr/configuration.md) · [Azərbaycanca](../az/configuration.md)

# Yapılandırma

Telepath yapılandırmayı dört kaynaktan okur. Öncelik sırası, en yüksekten başlayarak şöyledir:

**CLI bayrakları > ortam değişkenleri > `.env` dosyası > yerleşik varsayılanlar**

Her ortam değişkeninin bir `--kebab-case` bayrak karşılığı vardır. `.env` dosyası yalnızca gerçek ortamda henüz ayarlanmamış değerleri doldurur; bayrakların listesi için `telepath --help` komutunu çalıştırın.

## Referans

**Zorunlu** (Required) sütununda `telegram` / `slack`, değişkenin yalnızca ilgili transport etkinleştirildiğinde zorunlu olduğu anlamına gelir; boş bırakılan değerler isteğe bağlıdır.

| Ortam değişkeni | CLI bayrağı | Varsayılan | Zorunlu | Açıklama |
|-----------------|-------------|------------|:-------:|----------|
| `TRANSPORT` | `--transport` | `telegram` | | `telegram` \| `slack` \| `both`. Hangi mesajlaşma ön yüzlerinin çalıştırılacağı. Bkz. [Transport'lar](transports.md). |
| `TELEGRAM_BOT_TOKEN` | `--telegram-bot-token` | — | telegram | @BotFather'dan alınan bot token'ı. `TRANSPORT` `telegram` içerdiğinde zorunludur. |
| `ALLOWED_CHAT_IDS` | `--allowed-chat-ids` | — | telegram | Bota erişebilecek, virgülle ayrılmış Telegram sohbet kimlikleri. `telegram` etkinken zorunludur (en az 1). |
| `SLACK_BOT_TOKEN` | `--slack-bot-token` | — | slack | Slack bot token'ı (`xoxb-…`). `TRANSPORT` `slack` içerdiğinde zorunludur. |
| `SLACK_APP_TOKEN` | `--slack-app-token` | — | slack | Socket Mode için Slack uygulama düzeyinde token'ı (`xapp-…`). `slack` etkinken zorunludur. |
| `ALLOWED_SLACK_USER_IDS` | `--allowed-slack-user-ids` | — | | Görev tetikleyebilecek, virgülle ayrılmış Slack kullanıcı kimlikleri (isteğe bağlı). |
| `ALLOWED_SLACK_CHANNEL_IDS` | `--allowed-slack-channel-ids` | — | | Görev tetikleyebilecek, virgülle ayrılmış Slack kanal kimlikleri (isteğe bağlı). |
| `PROJECTS_DIR` | `--projects-dir` | `.` | | Ajanın çalıştığı klasör. `per-task` branch modu için bir git deposu olmalıdır. |
| `DEFAULT_AGENT` | `--default-agent` | `claude` | | Bir sohbet için varsayılan ajan. `ENABLED_AGENTS` içinde listelenmiş olmalıdır. |
| `ENABLED_AGENTS` | `--enabled-agents` | `claude,gemini,codex,aider` | | Virgülle ayrılmış, etkin ajanlar. |
| `PERMISSION_MODE` | `--permission-mode` | `auto` | | `auto` \| `allowlist` \| `interactive`. |
| `APPROVAL_TIMEOUT_SECONDS` | `--approval-timeout-seconds` | `120` | | Etkileşimli bir onayın **reddedilmeden** önce ne kadar bekleyeceği. |
| `BRANCH_MODE` | `--branch-mode` | `per-task` | | `off` \| `per-task`. |
| `BRANCH_PREFIX` | `--branch-prefix` | `telepath/` | | Görev branch adlarının öneki. |
| `BRANCH_COMMIT` | `--branch-commit` | `true` | | Ajanın değişikliklerini görev branch'inde commit et. |
| `BRANCH_STASH_DIRTY` | `--branch-stash-dirty` | `false` | | Görevi başlatmayı reddetmek yerine kirli çalışma ağacını stash'le. |
| `MAX_TURNS` | `--max-turns` | `0` | | Ajanın en fazla tur sayısı (`0` = CLI'nin kendi varsayılanı). |
| `MODEL` | `--model` | *(boş)* | | Ajana geçirilen model geçersiz kılma (`""` = CLI'nin varsayılanı). |
| `CLAUDE_BIN` | `--claude-bin` | *(boş)* | | PATH'te değilse `claude` ikilisinin yolu. |
| `GEMINI_BIN` | `--gemini-bin` | *(boş)* | | PATH'te değilse `gemini` ikilisinin yolu. |
| `CODEX_BIN` | `--codex-bin` | *(boş)* | | PATH'te değilse `codex` ikilisinin yolu. |
| `AIDER_BIN` | `--aider-bin` | *(boş)* | | PATH'te değilse `aider` ikilisinin yolu. |
| `GIT_BIN` | `--git-bin` | *(boş)* | | PATH'te değilse `git` ikilisinin yolu. |
| `LOG_LEVEL` | `--log-level` | `info` | | `debug` \| `info` \| `warn` \| `error`. |

Ayrıca çıktı verip sonlanan iki meta-bayrak vardır: `--version` ve `--help` (ayrıca `./.env` dışında bir `.env` dosyasını göstermek için `--env-file`).

## Notlar

- **Boolean değerler** `1/true/yes/on` ve `0/false/no/off` değerlerini kabul eder.
- **`ALLOWED_CHAT_IDS`** sert bir güvenlik sınırıdır. Her güncelleme — mesajlar ve düğme tıklamaları dahil — buna göre kontrol edilir; listede olmayan sohbetler yok sayılır.
- **`DEFAULT_AGENT` mutlaka `ENABLED_AGENTS` içinde olmalıdır**; aksi halde başlatma açık bir mesajla başarısız olur. Aynı şekilde `PERMISSION_MODE` ve `BRANCH_MODE` da geçerli olmalıdır.
- **`PROJECTS_DIR`** var olmalı ve bir klasör olmalıdır. Tüm sohbetler bu tek çalışma ağacını paylaşır (bkz. [Branch iş akışı](branch-workflow.md)).
- Sohbete özel ayarlar (`/use`, `/perm`, `/branch`) bu genel varsayılanların yerine yalnızca o sohbet için geçerlidir ve süreç yeniden başlayana kadar bellekte tutulur.

## Örnek `.env`

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

---

<!-- prev-next -->
[← Önceki: Başlangıç](./getting-started.md)  ·  [Sonraki: Taşıyıcılar →](./transports.md)
