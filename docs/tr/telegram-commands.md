> **Languages:** [English](../en/telegram-commands.md) · [Русский](../ru/telegram-commands.md) · [Türkçe](../tr/telegram-commands.md) · [Azərbaycanca](../az/telegram-commands.md)

# Telegram komutları

Tüm komutlar ve düğmeler yalnızca `ALLOWED_CHAT_IDS` içinde listelenen sohbetlerde çalışır. `/use`, `/perm` ve `/branch` ile yapılan ayarlar **sohbet bazında** uygulanır.

## Komutlar

| Komut | Sözdizimi | Örnek | Etkisi |
|-------|-----------|-------|--------|
| `/help` | `/help` | `/help` | Komut listesini gösterir. |
| `/agents` | `/agents` | `/agents` | Yapılandırılmış ajanları, kurulum durumları ve yetenekleriyle (akış, devam ettirme, etkileşimli onay, dosya düzenleme olayları) listeler. |
| `/use` | `/use <agent>` | `/use gemini` | Bu sohbetin varsayılan ajanını ayarlar. |
| `/perm` | `/perm <mode>` | `/perm interactive` | Bu sohbetin izin modunu ayarlar: `auto`, `allowlist` veya `interactive`. |
| `/branch` | `/branch <mode>` | `/branch per-task` | Bu sohbetin branch modunu ayarlar: `off` veya `per-task`. |
| `/<agent>` | `/<agent> <task>` | `/claude fix the null check in parser.go` | Belirli bir ajanla görev çalıştırır; yalnızca bu tek görev için sohbetin varsayılanını yok sayar. |
| *(düz mesaj)* | `<task text>` | `add retry logic to the HTTP client` | Görevi sohbetin varsayılan ajanıyla çalıştırır. |
| `/resume` | `/resume <text>` | `/resume now add a test for it` | Bu sohbetteki son oturumu sürdürür (yalnızca Claude). |
| `/branches` | `/branches` | `/branches` | `PROJECTS_DIR` içindeki mevcut `telepath/*` branch'lerini listeler. |
| `/status` | `/status` | `/status` | Geçerli ajan, izin ve branch modlarını ve bir görevin çalışıp çalışmadığını gösterir. |
| `/cancel` | `/cancel` | `/cancel` | Çalışan görevi iptal eder. Telepath bağlamı iptal eder ve ajan süreç ağacını sonlandırır. |

Notlar:

- Bir komut botun adıyla birlikte gönderilebilir (ör. `/claude@MyBot …`); `@name` eki çıkarılır.
- Sohbet başına yalnızca **bir görev** çalışır. Tüm sohbetler tek bir çalışma ağacını paylaştığı için görevler **genel olarak sıraya alınır**. Çalışma alanı meşgulse kısa bir "tekrar deneyin" yanıtı alırsınız.
- `/resume` yalnızca devam ettirme yeteneğini bildiren ajanlarla çalışır — şu anda **Claude**. `auto` izin modunda çalışır ve proje dizini, Telepath yeniden başlarsa kaybolan bellek içi bir haritadan geri yüklenir.

## Satır içi düğmeler

### Onay (etkileşimli mod)

Bir görev Claude ile `interactive` modunda çalıştığında, kontrol edilen her eylem iki düğmeli bir mesaj olarak görünür:

- **✅ Approve** — eyleme izin verir; ajan devam eder.
- **⛔ Deny** — eylemi reddeder; ajana reddedildiği bildirilir.

`APPROVAL_TIMEOUT_SECONDS` süresi içinde kimse yanıt vermezse istek otomatik olarak **reddedilir**. Bkz. [İzinler](permissions.md).

### Görev sonrası (görev başına branch modu)

Bir görev bir task branch'inde tamamlandığında (veya başarısız olduğunda) Telepath şunları sunar:

- **🌿 Keep branch** — branch'i, sizin incelemeniz, merge etmeniz ve kendiniz push etmeniz için olduğu gibi bırakır.
- **🗑 Delete branch & restore base** — görev branch'indeki değişiklikleri atar, temel branch'i yeniden checkout eder.

## Canlı durum mesajı

Bir çalıştırma sırasında Telepath, tek bir **durum mesajını** yerinde düzenleyerek günceller, örneğin:

```
🤖 claude · 🌿 telepath/fix-null-… · ✏️ parser.go · ▶ go build ./... · ✅ committed a1b2c3d
```

Düzenlemeler her token'da değil, mesajın yaklaşık saniyede bir güncellenmesi için sınırlandırılır. Her dosya düzenlemesi, birleşik bir diff içeren ayrı bir mesaj olarak da gönderilir (Telegram için çok büyükse `.diff` belgesi olarak).
