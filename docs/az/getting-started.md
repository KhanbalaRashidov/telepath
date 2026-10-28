> **Languages:** [English](../en/getting-started.md) · [Русский](../ru/getting-started.md) · [Türkçe](../tr/getting-started.md) · [Azərbaycanca](../az/getting-started.md)

# Başlanğıc

Bu təlimat sizi heç nədən Telegram-da ilk tapşırığınıza qədər aparacaq.

## 1. Tələblər

- **Go 1.26+** (yalnız mənbə kodundan build etmək üçün) — <https://go.dev/dl/>
- **git** PATH-də olmalıdır.
- Host-da ən azı bir **agent CLI** quraşdırılmalıdır. Yalnız aktivləşdirmək istədiyiniz agentlər lazımdır:

  | Agent | Quraşdırma |
  |-------|------------|
  | Claude Code | `npm install -g @anthropic-ai/claude-code` |
  | Gemini CLI | `npm install -g @google/gemini-cli` |
  | OpenAI Codex CLI | `npm install -g @openai/codex` |
  | Aider | `pipx install aider-chat` (və ya `pip install aider-chat`) |

  Hər agent öz autentifikasiyası (API açarı və ya login) ilə **xarici asılılıqdır** və onu öz sənədlərinə uyğun quraşdırırsınız. Telepath bu etimadnamələrə heç vaxt toxunmur. Telepath-ı işə saldıqdan sonra `/agents` hansı CLI-ların tapıldığını göstərir.

## 2. Telepath-ı quraşdırın

### Hazır binary

Layihənin Releases səhifəsindən öz OS/arch-ınız üçün arxivi yükləyin, açın və işlədiyini yoxlayın:

```bash
# Linux / macOS
chmod +x telepath
./telepath --version
```
```powershell
# Windows (PowerShell)
.\telepath.exe --version
```

### `go install` ilə mənbə kodundan

```bash
go install github.com/telepath/telepath/cmd/telepath@latest
```

> Əgər repozitori fork edilib/adı dəyişdirilibsə, `github.com/telepath/telepath` əvəzinə onun real modul yolundan istifadə edin (məsələn `github.com/KhanbalaRashidov/telepath`).

### Makefile ilə mənbə kodundan

```bash
git clone https://github.com/telepath/telepath
cd telepath
make build      # produces ./telepath (or telepath.exe on Windows)
```

## 3. Telegram botu yaradın

1. Telegram-da [@BotFather](https://t.me/BotFather)-i açın.
2. `/newbot` göndərin və təlimatları izləyin.
3. Verdiyi **bot token**-i kopyalayın.

## 4. Chat ID-nizi tapın

[@userinfobot](https://t.me/userinfobot)-a mesaj yazın; o, sizə rəqəmsal chat ID-nizi cavab verəcək. Telepath yalnız açıq şəkildə allowlist-ə əlavə etdiyiniz chat-lara xidmət göstərir.

## 5. Minimal `.env`

Nümunəni kopyalayın və üç əsas dəyəri doldurun:

```bash
cp .env.example .env
```

```dotenv
TELEGRAM_BOT_TOKEN=123456:ABC-your-token
ALLOWED_CHAT_IDS=123456789
PROJECTS_DIR=/path/to/your/git/repo
```

`PROJECTS_DIR` agentin işlədiyi qovluqdur. Standart hər tapşırıq üzrə branch iş axını üçün o, **git repozitori olmalıdır**. Bütün seçimlər üçün [Konfiqurasiya](configuration.md) səhifəsinə baxın.

## 6. İlk işə salma

```bash
./telepath
```

Başladıqda Telepath konfiqurasiyanı yoxlayır (bot token-i və ən azı bir icazəli chat ID-si olmadan başlamaqdan imtina edir), sonra Telegram-a long-polling ilə qoşulmağa başlayır.

## 7. İlk tapşırığınız

Telegram-da botla söhbətinizdə:

1. Əmrləri görmək üçün `/help` göndərin.
2. Agentinizin quraşdırıldığını təsdiqləmək üçün `/agents` göndərin.
3. Sadəcə düz mətnlə tapşırıq təsvir edin, məsələn:

   ```
   add a /health endpoint that returns 200 OK and a test for it
   ```

Telepath `telepath/add-a-health-endpoint-260115-1430` kimi bir branch yaradır, agenti orada işlədir, canlı status mesajı və hər fayl üzrə diff-lər göndərir, nəticəni həmin branch-da commit edir və nəhayət `git diff --stat`-ı tam diff ilə birlikdə `.diff` sənədi kimi, həmçinin **Keep branch / Delete branch & restore base** düymələri ilə göndərir.

Növbəti: [Telegram əmrləri](telegram-commands.md), [İcazələr](permissions.md) və [Branch iş axını](branch-workflow.md) səhifələrini oxuyun.
