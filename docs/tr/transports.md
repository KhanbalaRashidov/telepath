> **Languages:** [English](../en/transports.md) · [Русский](../ru/transports.md) · [Türkçe](../tr/transports.md) · [Azərbaycanca](../az/transports.md)

# Transport'lar

Telepath sizinle **Telegram**, **Slack** veya **ikisi birden** üzerinden iletişim kurabilir. Transport, yalnızca mesajlaşma ön yüzüdür; çekirdek (ajanlar, izinler, git branch'leri, normalleştirilmiş olay akışı) hangisini kullanırsanız kullanın aynıdır. Transport'u `TRANSPORT` ayarıyla seçin (`telegram` | `slack` | `both`, varsayılan `telegram`).

## Genel IP gerekmez

Her iki transport da yalnızca **giden** bağlantı kurar; webhook, gelen port veya genel URL olmadan NAT arkasında çalışır:

- **Telegram** long-polling kullanır.
- **Slack**, **Socket Mode** kullanır (botun Slack'e açtığı bir WebSocket).

Bu, Telepath'in temel bir özelliğidir: onu bir dizüstü bilgisayarda veya ev sunucusunda çalıştırıp telefonunuzdan yönetebilirsiniz.

## Telegram ve Slack

| | Telegram | Slack |
|---|----------|-------|
| Bağlantı | long-polling | Socket Mode (WebSocket) |
| Token'lar | tek bot token'ı | bot token'ı `xoxb-` + uygulama düzeyinde token `xapp-` |
| Görevi tetikleme | herhangi bir mesaj veya `/<agent> …` | `@Telepath …`, DM veya `/telepath …` |
| Komutlar | `/help`, `/perm`, … | `/telepath help`, `/telepath perm …`, … |
| Canlı durum | tek bir düzenlenen mesaj | tek bir düzenlenen mesaj (thread içinde) |
| Onaylar | satır içi Approve/Deny düğmeleri | Block Kit Approve/Deny düğmeleri |
| Allowlist | sohbet kimlikleri | kullanıcı kimlikleri ve/veya kanal kimlikleri |

Komutlar iki platform arasında **1:1** eşlenir. Slack'te bunlar tek `/telepath` slash komutunun alt komutlarıdır (her ikisini de belgeleyen [Telegram komutları](telegram-commands.md) sayfasına bakın).

## İkisini birlikte çalıştırmak

`TRANSPORT=both` ayarlayın ve hem Telegram hem de Slack kimlik bilgilerini sağlayın. Telepath, her iki ön yüzü de aynı motora, ajanlara ve `PROJECTS_DIR` dizinine bağlı olarak başlatır. Tüm konuşmalar tek bir çalışma ağacını paylaştığı için görevler yine genel olarak sıraya alınır: tüm Telegram sohbetleri ve Slack kanalları genelinde aynı anda bir görev çalışır. Sohbet bazlı durum (seçilen ajan, izin modu, branch modu) `transport:conversation` anahtarıyla tutulur; böylece bir Telegram sohbeti ile bir Slack kanalı birbirine karışmaz.

## Hangisini seçmeli

- Yalnızca siz, telefonunuzdan → **Telegram**, tek bir token ile en hızlı kurulumdur.
- Ekibiniz zaten Slack'te çalışıyorsa ya da onayların bir kanalda görünmesini istiyorsanız → **Slack**.
- Geçiş yapıyorsanız ya da iki kitleye birden hizmet vermek istiyorsanız → **both**.

Her birinin adım adım kurulumu için bkz. [Başlarken](getting-started.md).

---

<!-- prev-next -->
[← Önceki: Yapılandırma](./configuration.md)  ·  [Sonraki: Telegram komutları →](./telegram-commands.md)
