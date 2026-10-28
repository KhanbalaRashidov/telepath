> **Languages:** [English](../en/security.md) · [Русский](../ru/security.md) · [Türkçe](../tr/security.md) · [Azərbaycanca](../az/security.md)

# Təhlükəsizlik

Telepath maşınınızda real kodlaşdırma agentlərini işlədir, ona görə etibar modelini başa düşmək vacibdir.

## Chat-ID allowlist

Telepath yalnız `ALLOWED_CHAT_IDS`-də siyahıya salınmış chat-larda hərəkət edir. Allowlist **hər yeniləmədə — mesajlarda və düymə basışlarında eyni cür** — hər şeydən əvvəl tətbiq olunur. Telepath ən azı bir icazəli chat ID-si olmadan **başlamaqdan imtina edir** və allowlist-də olmayan chat-ları səssizcə nəzərə almır. Bot token-inizi gizli saxlayın: onu bilən hər kəs botunuza mesaj yaza bilər, lakin yalnız allowlist-dəki chat-lara əməl olunur.

## Agentlər öz etimadnamələrindən istifadə edir

Hər agent CLI-ı öz **öz** API açarı və ya login-i ilə autentifikasiya olunur; bu, agentin sənədlərinə uyğun olaraq host-da saxlanılır. Telepath bu etimadnamələri heç vaxt oxumur, saxlamır və ötürmür — o, yalnız CLI-ı işə salır. Telepath-ın özünün saxladığı yeganə sirr Telegram bot token-i və chat allowlist-idir.

## `auto` (icazələri ötürmə) ehtiyatlılığı

`auto` rejimində agent icazə sorğularını keçərək işləyir — soruşmadan fayl redaktə edə və shell əmrləri icra edə bilər. Bu rahat, lakin güclüdür. Üstünlük verin:

- Daimi nəzarət tələb etməyən təhlükəsizlik şəbəkəsi üçün `allowlist`, və ya
- hər riskli addımı görmək və təsdiq etmək istədiyinizdə `interactive` (Claude).

Bax: [İcazələr](permissions.md).

## Prompt-injection haqqında məlumatlılıq

Kodlaşdırma agenti repozitoriyanızın fayllarını və tapşırıq mətnini oxuyur. Hər ikisindəki zərərli məzmun (zəhərlənmiş fayl, hazırlanmış tapşırıq) agenti istənilməyən hərəkətlərə yönəltməyə cəhd edə bilər. Qarşısını almaq üçün:

- Tam etibar etmədiyiniz repozitoriyalarda `allowlist` və ya `interactive` istifadə edin.
- Birləşdirməzdən əvvəl tapşırıq branch-ını nəzərdən keçirin — iş axını bunun üzərində qurulub.
- Rədd siyahısının `allowlist` rejimində `curl`/`wget`/`WebFetch`, `rm`, `git push` və s.-ni blok etdiyini unutmayın.

## Git zəmanəti

Telepath bütün git-ə tək qorunan kod yolu ilə sahiblik edir; bu yol **heç vaxt** push, merge, rebase, məcburi əməliyyat və ya `main`/`master`-ə commit etmir. Remote və əsas branch-ınız heç vaxt dəyişdirilmir. **Hər tapşırığın ən pis halı nəzərdən keçirilə bilən yerli branch-dır**; onu saxlamaq və ya silmək sizin seçiminizdir. Bax: [Branch iş axını](branch-workflow.md).

## Şəbəkə vəziyyəti

Telepath Telegram ilə **long-polling** (yalnız çıxış istiqamətində) üzərindən danışır, buna görə NAT arxasında, ictimai IP, webhook və ya daxil olan port olmadan işləyir. İnteraktiv təsdiq MCP serveri isə tək icra müddətində təsadüfi portda yalnız **loopback** (`127.0.0.1`)-a bağlanır.
