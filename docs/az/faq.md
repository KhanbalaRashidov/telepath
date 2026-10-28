> **Languages:** [English](../en/faq.md) · [Русский](../ru/faq.md) · [Türkçe](../tr/faq.md) · [Azərbaycanca](../az/faq.md)

# Tez-tez verilən suallar

**Hansı əməliyyat sistemləri dəstəklənir?**
Windows, Linux və macOS yerli olaraq (bir kod bazası, təmiz Go, CGO söndürülüb) və Docker konteyneri kimi.

**Telepath kodumu commit və ya push edir?**
O, **tapşırıq branch-ında** commit edir (`BRANCH_COMMIT=true` olduqda), lakin **heç vaxt push və ya merge etmir, `main`/`master`-ə və ya remote-a toxunmur**. Nəzərdən keçirib merge/push-u özünüz edirsiniz. `BRANCH_MODE=off` ilə heç commit də etmir.

**Hansı agentlər canlı təsdiqi dəstəkləyir?**
Yalnız **Claude Code** (onun MCP icazə-prompt aləti vasitəsilə). gemini, codex və aider üçün `interactive` avtomatik `allowlist`-ə keçir.

**Eyni anda bir neçə tapşırıq işlədə bilərmi?**
Yox. Bütün chat-lar bir `PROJECTS_DIR` working tree-ni paylaşır, ona görə tapşırıqlar **ardıcıl** icra olunur — ümumilikdə bir-bir, hər chat üçün bir. Bax: [Problemlərin həlli](troubleshooting.md).

**NAT arxasında / ictimai IP olmadan işləyirmi?**
Bəli. Telegram **long-polling** (yalnız çıxış istiqamətində) istifadə edir — webhook, daxil olan port və ya ictimai IP tələb olunmur.

**Agent API açarlarım haraya gedir?**
Onlar host-unuzda agent CLI-ları ilə qalır və hər agentin öz sənədlərinə görə konfiqurasiya olunur. Telepath yalnız Telegram bot token-i və chat allowlist-ini saxlayır. Bax: [Təhlükəsizlik](security.md).

**Hər chat üçün agenti və ya rejimi dəyişə bilərəmmi?**
Bəli — `/use <agent>`, `/perm <mode>`, `/branch <mode>` həmin chat-a tətbiq olunur və proses yenidən başlayana qədər qalır.

**Başqa agent necə əlavə edirəm?**
`Agent` interfeysini yeni paketdə implement edin və bir sətirlə qeydiyyatdan keçirin. Bax: [Agent əlavə etmək](adding-an-agent.md).

**Sənədləşmə dilini necə əlavə edirəm?**
`docs/en/`-i `docs/<locale>/`-ə kopyalayın, mətni tərcümə edin və dili `mkdocs.yml`-də və switcher linklərində qeydiyyatdan keçirin. Bax: [CONTRIBUTING.md](../../CONTRIBUTING.md#adding-a-documentation-language).

**Telepath-ı necə yeniləyirəm?**
Releases-dən son binary-ni yükləyin, və ya mənbədən yenidən build edin (`go install …@latest` və ya `make build`). Konfiqurasiya və `.env` dəyişmir.

**Xərclər nə qədərdir?**
Telepath pulsuzdur (MIT lisenziyası). Yalnız öz agent hesablarınız altında əsas agent/model istifadəsi üçün ödəyirsiniz.
