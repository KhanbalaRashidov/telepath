> **Languages:** [English](../en/telegram-commands.md) · [Русский](../ru/telegram-commands.md) · [Türkçe](../tr/telegram-commands.md) · [Azərbaycanca](../az/telegram-commands.md)

# Telegram əmrləri

Bütün əmrlər və düymələr yalnız `ALLOWED_CHAT_IDS`-də siyahıya salınmış chat-larda işləyir. `/use`, `/perm` və `/branch` ilə dəyişdirilən parametrlər **chat üzrə** tətbiq olunur.

## Əmrlər

| Əmr | Sintaksis | Nümunə | Təsir |
|-----|-----------|--------|-------|
| `/help` | `/help` | `/help` | Əmrlər siyahısını göstərir. |
| `/agents` | `/agents` | `/agents` | Konfiqurasiya olunmuş agentləri quraşdırma statusu və imkanları (streaming, resume, interaktiv təsdiq, fayl redaktə hadisələri) ilə siyahılayır. |
| `/use` | `/use <agent>` | `/use gemini` | Bu chat-ın standart agentini təyin edir. |
| `/perm` | `/perm <mode>` | `/perm interactive` | Bu chat-ın icazə rejimini təyin edir: `auto`, `allowlist` və ya `interactive`. |
| `/branch` | `/branch <mode>` | `/branch per-task` | Bu chat-ın branch rejimini təyin edir: `off` və ya `per-task`. |
| `/<agent>` | `/<agent> <task>` | `/claude fix the null check in parser.go` | Konkret agentlə tapşırıq icra edir; yalnız bu bir tapşırıq üçün chat standartını nəzərə almır. |
| *(düz mesaj)* | `<task text>` | `add retry logic to the HTTP client` | Chat-ın standart agenti ilə tapşırıq icra edir. |
| `/resume` | `/resume <text>` | `/resume now add a test for it` | Bu chat-da son sessiyanı davam etdirir (yalnız Claude). |
| `/branches` | `/branches` | `/branches` | `PROJECTS_DIR`-dəki mövcud `telepath/*` branch-larını siyahılayır. |
| `/status` | `/status` | `/status` | Cari agent/icazə/branch rejimlərini və tapşırığın işləyib-işləmədiyini göstərir. |
| `/cancel` | `/cancel` | `/cancel` | İşləyən tapşırığı ləğv edir. Telepath konteksti ləğv edir və agent proses ağacını öldürür. |

Qeydlər:

- Əmr botu adla ünvanlana bilər (məsələn `/claude@MyBot …`); `@name` şəkilçisi silinir.
- Hər chat-da eyni anda **yalnız bir tapşırıq** işləyir və bütün chat-lar eyni working tree-ni paylaşdığı üçün tapşırıqlar **qlobal ardıcıllıqla** icra olunur. İş sahəsi məşğul olduqda qısa "yenidən cəhd edin" cavabı alırsınız.
- `/resume` yalnız resume imkanını bildirən agentlər üçün işləyir — hal-hazırda **Claude**. `auto` icazə rejimində işləyir və layihə qovluğu yaddaşdakı xəritədən bərpa olunur; Telepath yenidən başladıqda bu xəritə itir.

## Slack əmrləri

Slack-də yuxarıdakı hər əmr tək `/telepath` slash əmrinin alt əmridir, tapşırıqlar isə botu
mention etməklə və ya ona DM yazmaqla işə salınır. Uyğunluq Telegram ilə **1:1**-dir:

| Telegram | Slack |
|----------|-------|
| `/help` | `/telepath help` |
| `/agents` | `/telepath agents` |
| `/use <agent>` | `/telepath use <agent>` |
| `/perm <mode>` | `/telepath perm <mode>` |
| `/branch <mode>` | `/telepath branch <mode>` |
| `/branches` | `/telepath branches` |
| `/status` | `/telepath status` |
| `/cancel` | `/telepath cancel` |
| `/resume <text>` | `/telepath resume <text>` |
| `/<agent> <task>` | `/telepath <agent> <task>` və ya `@Telepath <agent>: <task>` |
| *(düz mesaj)* | `@Telepath <task>` və ya birbaşa mesaj |

Qeydlər:

- `/telepath` alt əmri tanınmış əmr və ya agent adı deyilsə, standart agentin tapşırığı kimi qəbul olunur
  (məsələn `/telepath fix the flaky test`).
- Tapşırığın çıxışı işə salan mesajın altında **thread**-də göndərilir (slash əmrləri üçün botun
  ilk cavabının altında). Canlı status mesajı Telegram-da olduğu kimi yerində redaktə olunur.
- Botdan kimin istifadə edə biləcəyi `ALLOWED_SLACK_USER_IDS` / `ALLOWED_SLACK_CHANNEL_IDS` ilə
  idarə olunur (bax: [Təhlükəsizlik](security.md)).

## Inline düymələr

Aşağıdakı düymələr hər iki transportda işləyir (Telegram inline klaviaturaları, Slack Block Kit).

### Təsdiq (interactive rejim)

Tapşırıq Claude ilə `interactive` rejimində işlədikdə, hər nəzarət olunan hərəkət iki düymə olan mesaj kimi görünür:

- **✅ Approve** — hərəkətə icazə verin; agent davam edir.
- **⛔ Deny** — hərəkəti rədd edin; agentə rədd edildiyi bildirilir.

Əgər heç kim `APPROVAL_TIMEOUT_SECONDS` ərzində cavab verməsə, sorğu avtomatik **rədd edilir**. Bax: [İcazələr](permissions.md).

### Tapşırıqdan sonra (per-task branch rejimi)

Tapşırıq tapşırıq branch-ında bitdikdə (və ya uğursuz olduqda) Telepath bunları təklif edir:

- **🌿 Keep branch** — branch-ı nəzərdən keçirmək, merge etmək və özünüz push etmək üçün yerində saxlayın.
- **🗑 Delete branch & restore base** — tapşırıq branch-ını ləğv edin və baza branch-ını yenidən checkout edin.

## Canlı status mesajı

İcra zamanı Telepath **tək status mesajı** saxlayır və onu yerində redaktə edir, məsələn:

```
🤖 claude · 🌿 telepath/fix-null-… · ✏️ parser.go · ▶ go build ./... · ✅ committed a1b2c3d
```

Redaktələr throttle olunub, beləliklə mesaj hər token üçün deyil, təxminən saniyədə bir dəfə yenilənir. Hər fayl redaktəsi də öz mesajı kimi unified diff ilə göndərilir (Telegram üçün çox böyükdürsə, `.diff` sənədi kimi).

---

<!-- prev-next -->
[← Əvvəlki: Nəqliyyatlar](./transports.md)  ·  [Növbəti: Agentlər →](./agents.md)
