> **Languages:** [English](../en/getting-started.md) · [Русский](../ru/getting-started.md) · [Türkçe](../tr/getting-started.md) · [Azərbaycanca](../az/getting-started.md)

# Başlarken

Bu kılavuz, sıfırdan başlayarak Telegram'da ilk görevinizi göndermenize kadar sizi yönlendirir.

## 1. Ön koşullar

- **Go 1.26+** (yalnızca kaynaktan derlemek için) — <https://go.dev/dl/>
- PATH'inizde **git**.
- Host makinede en az bir **ajan CLI'si** kurulu olmalı. Yalnızca etkinleştirmeyi planladıklarınızı kurmanız yeterlidir:

  | Ajan | Kurulum |
  |------|---------|
  | Claude Code | `npm install -g @anthropic-ai/claude-code` |
  | Gemini CLI | `npm install -g @google/gemini-cli` |
  | OpenAI Codex CLI | `npm install -g @openai/codex` |
  | Aider | `pipx install aider-chat` (veya `pip install aider-chat`) |

  Her ajan, kendi kimlik doğrulamasına (API anahtarı veya oturum) sahip bir **dış bağımlılıktır**; bunu kendi dokümanlarına göre kurarsınız. Telepath bu kimlik bilgilerine hiçbir zaman dokunmaz. Telepath'i başlattıktan sonra `/agents` komutu hangi CLI'lerin bulunduğunu gösterir.

## 2. Telepath'i kurun

### Hazır ikili dosya

Proje'nin Releases sayfasından işletim sisteminize ve mimarinize uygun arşivi indirin, çıkarın ve çalıştığını kontrol edin:

```bash
# Linux / macOS
chmod +x telepath
./telepath --version
```
```powershell
# Windows (PowerShell)
.\telepath.exe --version
```

### `go install` ile kaynaktan

```bash
go install github.com/telepath/telepath/cmd/telepath@latest
```

> Depo çatallandıysa veya yeniden adlandırıldıysa, `github.com/telepath/telepath` yerine gerçek modül yolunu kullanın (örneğin `github.com/KhanbalaRashidov/telepath`).

### Makefile ile kaynaktan

```bash
git clone https://github.com/KhanbalaRashidov/telepath
cd telepath
make build      # produces ./telepath (or telepath.exe on Windows)
```

## 3. Bir Telegram botu oluşturun

1. Telegram'da [@BotFather](https://t.me/BotFather) ile konuşmayı açın.
2. `/newbot` gönderin ve istemleri izleyin.
3. Size verilen **bot token'ını** kopyalayın.

## 4. Sohbet kimliğinizi bulun

[@userinfobot](https://t.me/userinfobot) adresine mesaj gönderin; size sayısal sohbet kimliğinizi yanıtlayacaktır. Telepath yalnızca açıkça allowlist'e eklediğiniz sohbetlere hizmet verir.

## 5. Minimal `.env`

Örnek dosyayı kopyalayın ve üç temel değeri doldurun:

```bash
cp .env.example .env
```

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789
PROJECTS_DIR=/path/to/your/git/repo
```

`PROJECTS_DIR`, ajanın çalıştığı klasördür. Varsayılan görev başına branch iş akışı için **bir git deposu olmalıdır**. Tüm seçenekler için bkz. [Yapılandırma](configuration.md).

## 6. İlk çalıştırma

```bash
./telepath
```

Başlarken Telepath yapılandırmayı doğrular (bot token'ı ve en az bir izinli sohbet kimliği olmadan başlamayı reddeder) ve ardından Telegram'ı long-polling ile dinlemeye başlar.

## 7. İlk göreviniz

Bot ile Telegram sohbetinizde:

1. Komutları görmek için `/help` gönderin.
2. Ajanınızın kurulu olduğunu doğrulamak için `/agents` gönderin.
3. Düz metin olarak bir görev tarif edin, örneğin:

   ```
   add a /health endpoint that returns 200 OK and a test for it
   ```

Telepath `telepath/add-a-health-endpoint-260115-1430` gibi bir branch oluşturur, ajanı orada çalıştırır, canlı bir durum mesajı ile dosya başına diff'leri gösterir, sonucu o branch'e commit eder ve son olarak `git diff --stat` çıktısını ile tam diff'i `.diff` belgesi olarak, **Keep branch** ve **Delete branch & restore base** düğmeleriyle birlikte gönderir.

## Slack kurulumu (isteğe bağlı)

Slack'i tercih ediyorsanız ya da ikisini birden istiyorsanız, Telepath **Socket Mode** üzerinden Slack ile de çalışır (herkese açık bir URL gerekmez). Karşılaştırma için bkz. [Transport'lar](transports.md).

1. <https://api.slack.com/apps> adresinde **bir Slack uygulaması oluşturun** → *Create New App* → *From scratch*. Bir ad ve çalışma alanı seçin.
2. **Socket Mode'u etkinleştirin**: *Settings → Socket Mode → Enable*. Bu, `connections:write` kapsamına sahip bir **uygulama düzeyinde token** (`xapp-…`) üretir; kopyalayın.
3. **Bot kapsamlarını ekleyin**: *Features → OAuth & Permissions → Scopes → Bot Token Scopes* altına şunları ekleyin: `app_mentions:read`, `chat:write`, `commands`, `files:write`, `im:history`, `im:read`, `channels:history`, `groups:history`.
4. **Slash komutunu oluşturun**: *Features → Slash Commands → Create New Command* → `/telepath` (açıklama serbest; Socket Mode altında Request URL yok sayılır, ancak Slack bir değer ister — herhangi bir yer tutucu olabilir).
5. **Olaylara abone olun**: *Features → Event Subscriptions → Enable*, ardından *Subscribe to bot events* altına `app_mention` ve `message.im` ekleyin.
6. **Çalışma alanına yükleyin**: *OAuth & Permissions → Install to Workspace*. **Bot User OAuth Token**'ı (`xoxb-…`) kopyalayın.
7. Botu kullanmak istediğiniz kanallara **davet edin** (`/invite @YourApp`) ya da doğrudan DM gönderin.
8. Allowlist için **kimlikleri bulun**: kullanıcı kimliği (profil → *Copy member ID*, `U…`) ve/veya kanal kimliği (kanal ayrıntıları, `C…`).

Minimal Slack `.env`:

```dotenv
TRANSPORT=slack
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token
ALLOWED_SLACK_USER_IDS=U0123456789
PROJECTS_DIR=/path/to/your/git/repo
```

`./telepath` komutunu çalıştırın, ardından Slack'te botu etiketleyin ya da slash komutunu kullanın:

```
@Telepath add a /health endpoint that returns 200 OK and a test for it
/telepath help
/telepath claude fix the null check in parser.go
```

**Her iki** transport'u aynı anda çalıştırmak için `TRANSPORT=both` ayarlayın ve hem Telegram hem de Slack kimlik bilgilerini sağlayın.

Sonraki adım: [Transport'lar](transports.md), [Telegram komutları](telegram-commands.md), [İzinler](permissions.md) ve [Branch iş akışı](branch-workflow.md) sayfalarını okuyun.

---

<!-- prev-next -->
[← Önceki: Ana sayfa](./index.md)  ·  [Sonraki: Yapılandırma →](./configuration.md)
