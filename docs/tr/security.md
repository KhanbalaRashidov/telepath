> **Languages:** [English](../en/security.md) · [Русский](../ru/security.md) · [Türkçe](../tr/security.md) · [Azərbaycanca](../az/security.md)

# Güvenlik

Telepath gerçek kodlama ajanlarını makinenizde çalıştırır, bu yüzden güven modelini anlamanız önemlidir.

## Sohbet kimliği allowlist'i

Telepath yalnızca `ALLOWED_CHAT_IDS` içinde listelenen sohbetlerde işlem yapar. Allowlist, başka hiçbir şey yapılmadan önce **her güncellemede — mesajlarda ve düğme tıklamalarında aynı şekilde** uygulanır. Telepath, en az bir izinli sohbet kimliği olmadan **başlamayı reddeder** ve allowlist'te olmayan sohbetleri sessizce yok sayar. Bot token'ınızı gizli tutun; onu bilen herkes botunuza mesaj gönderebilir, ancak yalnızca allowlist'teki sohbetler üzerinde işlem yapılır.

## Slack allowlist'i

Slack'te erişim `ALLOWED_SLACK_USER_IDS` ve/veya `ALLOWED_SLACK_CHANNEL_IDS` ile kontrol edilir (Telegram sohbet allowlist'inden ayrıdır). Tetikleyen **kullanıcı** kullanıcı listesinde ya da **kanal** kanal listesinde varsa istek kabul edilir. Bu listeler isteğe bağlıdır; Telegram'dan farklı olarak, **ikisi de boşsa Telepath botun görebildiği her kanaldaki her kullanıcıya izin verir** ve başlangıçta bir uyarı kaydeder. Özel bir test çalışma alanının ötesinde kullanım için en az bir liste ayarlayın. `xoxb-`/`xapp-` token'larını gizli tutun: bunlar botun çalışma alanınıza erişimini sağlar.

## Ajanlar kendi kimlik bilgilerini kullanır

Her ajan CLI'si, o ajanın dokümanlarına göre host üzerinde saklanan **kendi** API anahtarı veya oturumuyla kimlik doğrulaması yapar. Telepath bu kimlik bilgilerini asla okumaz, saklamaz veya iletmez; yalnızca CLI'yi başlatır. Telepath'in kendi tuttuğu tek gizli bilgiler transport token'larıdır (Telegram bot token'ı ve/veya Slack `xoxb-`/`xapp-` token'ları) ile allowlist'lerdir.

## `auto` (izinleri atlama) uyarısı

`auto` modunda ajan, izin istemleri atlanmış olarak çalışır — sormadan dosya düzenleyebilir ve kabuk komutları çalıştırabilir. Bu kullanışlı ama güçlüdür. Tercih edin:

- Yine de sürekli ilgilenmeyi gerektirmeyen bir güvenlik ağı için `allowlist`, ya da
- Her riskli adımı görmek ve onaylamak istediğinizde `interactive` (Claude).

Bkz. [İzinler](permissions.md).

## İstem enjeksiyonu farkındalığı

Bir kodlama ajanı, deponuzun dosyalarını ve görev metnini okur. Bunlardan herhangi birindeki kötü amaçlı içerik (zehirlenmiş bir dosya, hazırlanmış bir görev) ajanı istenmeyen eylemlere yönlendirmeye çalışabilir. Önlemler:

- Tamamen güvenmediğiniz depolarda `allowlist` veya `interactive` kullanın.
- Birleştirmeden önce görev branch'ini inceleyin — iş akışı bunun üzerine kuruludur.
- Reddedilenler kümesinin, `allowlist` modunda `curl`/`wget`/`WebFetch`, `rm`, `git push` vb. komutları engellediğini unutmayın.

## Git garantisi

Telepath tüm git işlemlerinin sahibidir; bunu tek bir korumalı kod yolu üzerinden yapar ve **asla** push, merge, rebase veya zorla herhangi bir işlem yapmaz, `main`/`master` üzerine commit yapmaz. Uzak depo ve ana branch'iniz asla değiştirilmez. **Herhangi bir görevin en kötü durumu, sizin tutmayı veya silmeyi seçebileceğiniz, incelenebilir yerel bir branch'tir.** Bkz. [Branch iş akışı](branch-workflow.md).

## Ağ yapısı

Telepath yalnızca **giden** bağlantı kurar — Telegram ile **long-polling**, Slack ile **Socket Mode** (botun açtığı bir WebSocket) üzerinden — bu sayede genel IP, webhook veya gelen port olmadan NAT arkasında da çalışır. Etkileşimli onay MCP sunucusu, tek bir çalıştırmanın ömrü boyunca rastgele bir porta **yalnızca loopback'e** (`127.0.0.1`) bağlanır.

---

<!-- prev-next -->
[← Önceki: Ajan ekleme](./adding-an-agent.md)  ·  [Sonraki: Sorun giderme →](./troubleshooting.md)
