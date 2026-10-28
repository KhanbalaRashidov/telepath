> **Languages:** [English](../en/transports.md) · [Русский](../ru/transports.md) · [Türkçe](../tr/transports.md) · [Azərbaycanca](../az/transports.md)

# Transportlar

Telepath sizinlə **Telegram**, **Slack** və ya **hər ikisi** eyni anda ilə danışa bilər. Transport
sadəcə mesajlaşma ön hissəsidir; nüvə (agentlər, icazələr, git branch-ları, normallaşdırılmış
hadisə axını) hansını istifadə etməyinizdən asılı olmayaraq eynidir. Transportları `TRANSPORT`
parametri ilə seçin (`telegram` | `slack` | `both`, standart `telegram`).

## İctimai IP tələb olunmur

Hər iki transport **yalnız xaricə** qoşulur və webhook, daxil olan port və ya ictimai URL
olmadan NAT arxasında işləyir:

- **Telegram** long-polling istifadə edir.
- **Slack** **Socket Mode** istifadə edir (botun Slack-ə açdığı WebSocket).

Bu, Telepath-ın əsas xüsusiyyətidir: onu noutbukda və ya ev serverində işlədə və telefonunuzdan
idarə edə bilərsiniz.

## Telegram və Slack

| | Telegram | Slack |
|---|----------|-------|
| Bağlantı | long-polling | Socket Mode (WebSocket) |
| Token-lər | bir bot token-i | bot token-i `xoxb-` + tətbiq səviyyəli token `xapp-` |
| Tapşırığı başlatmaq | istənilən mesaj, və ya `/<agent> …` | `@Telepath …`, DM, və ya `/telepath …` |
| Əmrlər | `/help`, `/perm`, … | `/telepath help`, `/telepath perm …`, … |
| Canlı status | bir redaktə edilmiş mesaj | bir redaktə edilmiş mesaj (thread-də) |
| Təsdiqlər | inline Approve/Deny düymələri | Block Kit Approve/Deny düymələri |
| Allowlist | chat ID-ləri | user ID-ləri və/və ya channel ID-ləri |

Əmrlər ikisi arasında **1:1** uyğunluqdadır. Slack-də onlar tək `/telepath` slash əmrinin
alt əmrləridir (hər ikisini sənədləşdirən [Telegram əmrləri](telegram-commands.md) səhifəsinə baxın).

## Eyni anda hər ikisi

`TRANSPORT=both` təyin edin və həm Telegram, həm də Slack məlumatlarını təqdim edin. Telepath
hər iki ön hissəni eyni mühərrik, agentlər və `PROJECTS_DIR` üzərində işlədir. Bütün söhbətlər
bir working tree paylaşdığı üçün tapşırıqlar yenə qlobal olaraq ardıcıl icra olunur — bütün
Telegram chat-ları və Slack kanalları arasında eyni anda bir tapşırıq — və hər söhbətin vəziyyəti
(seçilmiş agent, icazə rejimi, branch rejimi) `transport:conversation` ilə açarlanır, belə ki,
Telegram chat-ı və Slack kanalı bir-birinə heç vaxt müdaxilə etmir.

## Birini seçmək

- Yalnız siz, telefonunuzda → **Telegram** ən sürətli quraşdırılandır (bir token).
- Komandanız artıq Slack-dədir və ya təsdiqlərin kanalda görünməsini istəyirsiniz → **Slack**.
- Köçürürsünüzsə və ya hər iki auditoriyanı istəyirsinizsə → **both**.

Hər birinin addım-addım quraşdırılması üçün [Başlanğıc](getting-started.md) səhifəsinə baxın.

---

<!-- prev-next -->
[← Əvvəlki: Konfiqurasiya](./configuration.md)  ·  [Növbəti: Telegram əmrləri →](./telegram-commands.md)
