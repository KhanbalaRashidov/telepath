> **Languages:** [English](../en/troubleshooting.md) · [Русский](../ru/troubleshooting.md) · [Türkçe](../tr/troubleshooting.md) · [Azərbaycanca](../az/troubleshooting.md)

# Problemlərin həlli

Ətraflı məlumat üçün `LOG_LEVEL=debug` təyin edin. Aşağıda ümumi xəta halları var.

## "Agent is not installed"

`/agents` hər agent üçün ✅/❌ göstərir. Sizin agent ❌-dırsa:

- CLI-ın PATH-də olduğunu təsdiqləyin (`claude --version`, `gemini --version`, `codex --version`, `aider --version`).
- Standart olmayan yerdədirsə, uyğun `*_BIN` dəyişənini (`CLAUDE_BIN`, `GEMINI_BIN`, `CODEX_BIN`, `AIDER_BIN`) tam yola təyin edin.
- Docker-də müstəqil image-dən istifadə edin (CLI-ları özündə saxlayır) və ya CLI-ları yüngül image-ə mount edin.

## Bot cavab vermir

- **Token:** `TELEGRAM_BOT_TOKEN`-i yoxlayın. Başlanğıcda Telepath token-i Telegram ilə yoxlayır; yanlış token dərhal uğursuz olur.
- **Allowlist:** chat ID-niz `ALLOWED_CHAT_IDS`-də olmalıdır. Allowlist-də olmayan chat-lar səssizcə nəzərə alınır (loglardakı xəbərdarlığa baxın).
- **Bağlantı:** Telepath çıxan long-polling istifadə edir; host-un `api.telegram.org`-a çatdığından əmin olun.
- **Başqa yerdə işləmə:** verilmiş token-i eyni anda yalnız bir bot prosesi long-poll etməlidir.

## "Working tree is dirty" imtinası

`per-task` rejimində, `PROJECTS_DIR`-də commit edilməmiş dəyişikliklər varsa, Telepath tapşırığa başlamaqdan imtina edir. Ya onları özünüz commit/stash edin, ya da `BRANCH_STASH_DIRTY=true` təyin edərək Telepath-ın onları avtomatik stash etməsinə icazə verin.

## Təsdiq vaxtı bitdi → rədd

`interactive` rejimində, `APPROVAL_TIMEOUT_SECONDS` (standart 120) ərzində heç kim Approve/Deny basmasa, sorğu **rədd edilir** və agentə bu bildirilir. Daha çox vaxt lazımdırsa timeout-u artırın, və ya soruşulmaq istəmirsinizsə `allowlist`/`auto` istifadə edin.

## Slack: bot səssizdir

Əgər `TRANSPORT=slack` (və ya `both`) təyin olunubsa, lakin heç nə baş vermirsə:

- **Socket Mode söndürülüb / yanlış token.** Socket Mode aktiv olmalıdır, `SLACK_APP_TOKEN`
  **tətbiq səviyyəli** token olmalıdır (`xapp-…`, scope `connections:write`), `SLACK_BOT_TOKEN` isə
  **bot** token-i (`xoxb-…`) olmalıdır. Onları qarışdırmaq ən çox rast gəlinən səhvdir.
- **Scope-lar çatışmır.** Sizə `app_mentions:read`, `chat:write`, `commands`, `files:write`,
  `im:history`, `im:read`, `channels:history`, `groups:history` lazımdır. Scope-ları əlavə etdikdən
  sonra tətbiqi **yenidən quraşdırın**.
- **Bot dəvət olunmayıb.** Kanalda onu `/invite @YourApp` ilə dəvət edin; və ya sadəcə ona DM yazın.
- **Hadisə abunələri.** `app_mention` və `message.im` *Event Subscriptions → bot events* altında
  abunə olunmalıdır.
- **Allowlist uyğunsuzluğu.** `ALLOWED_SLACK_USER_IDS` / `ALLOWED_SLACK_CHANNEL_IDS` təyin olunubsa,
  istifadəçiniz və ya kanalınız siyahıda olmalıdır; əks halda sorğu nəzərə alınmır (xəbərdarlıq
  loglanır). Bax: [Təhlükəsizlik](security.md).
- **Diff-lər yüklənmir.** `.diff` yükləmə axını üçün `files:write` tələb olunur.

## Interactive məndən soruşmadı

İnteraktiv təsdiq **yalnız Claude ilə** işləyir. gemini/codex/aider ilə `interactive` `allowlist`-ə keçir və cavab faktiki istifadə olunan rejimi bildirir. Canlı təsdiqlər üçün chat-ı `/use claude` ilə Claude-a keçirin.

## Tapşırıqlar növbəyə düşür

Bütün chat-lar bir `PROJECTS_DIR` working tree-ni paylaşır, ona görə tapşırıqlar **ardıcıl** icra olunur — bütün chat-lar üzrə bir-bir, hər chat üçün bir. İş sahəsi məşğul olduqda "yenidən cəhd edin" cavabı alırsınız. Nəyin işlədiyini görmək üçün `/status`, dayandırmaq üçün `/cancel` istifadə edin.

## `/cancel` hər şeyi öldürmədi (Windows)

Telepath agent proses **ağacını** ən yaxşı cəhdlə öldürür — Windows-da `taskkill /F /T`, Unix-də proses qrupu üzrə `SIGKILL`. Dərin iç-içə köməkçi proseslər nadir hallarda qala bilər; Node prosesi sağ qalırsa Task Manager-i yoxlayın.

## `/resume` davam etdiriləcək heç nə olmadığını deyir

Son sessiya id-si və onun layihə qovluğu **yaddaşda** saxlanılır və Telepath yenidən başladıqda itir. Yenidən başladıqdan sonra davam etdiriləcək heç nə qalmır. Resume həm də **yalnız Claude** üçündür; digər agentlər bunu dəstəklənmir kimi bildirir.

## Commit uğursuz oldu

Əgər repozitoriyada git şəxsiyyəti yoxdursa, Telepath ehtiyat `Telepath <telepath@localhost>` şəxsiyyəti ilə commit edir ki, tapşırıq branch-ının commit-i yenə uğurlu olsun. Commit yenə də uğursuz olarsa, xəta yekun hesabatda göstərilir — repozitoriyanın git vəziyyətini və icazələrini yoxlayın.

---

<!-- prev-next -->
[← Əvvəlki: Təhlükəsizlik](./security.md)  ·  [Növbəti: Tez-tez verilən suallar →](./faq.md)
