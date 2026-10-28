> **Languages:** [English](../en/faq.md) · [Русский](../ru/faq.md) · [Türkçe](../tr/faq.md) · [Azərbaycanca](../az/faq.md)

# SSS

**Hangi işletim sistemleri destekleniyor?**
Windows, Linux ve macOS'ta yerel olarak (tek kod tabanı, saf Go, CGO kapalı) ve bir Docker konteyneri olarak.

**Telepath kodumu commit veya push ediyor mu?**
`BRANCH_COMMIT=true` olduğunda bir **görev branch'inde** commit eder, ancak `main`/`master` veya uzak depoya **asla push, merge veya dokunmaz**. İncelersiniz ve merge/push işlemini kendiniz yaparsınız. `BRANCH_MODE=off` ile hiç commit yapmaz.

**Hangi ajanlar canlı onayı destekliyor?**
Yalnızca **Claude Code** (MCP izin-istemi aracı üzerinden). gemini, codex ve aider için `interactive` otomatik olarak `allowlist`'e geri döner.

**Aynı anda birden fazla görev çalıştırabilir mi?**
Hayır. Tüm sohbetler tek bir `PROJECTS_DIR` çalışma ağacını paylaştığı için görevler **sıraya alınır** — genel olarak aynı anda bir tane, sohbet başına da bir tane. Bkz. [Sorun giderme](troubleshooting.md).

**NAT arkasında / genel IP olmadan çalışır mı?**
Evet. Telegram **long-polling** (yalnızca giden bağlantı) kullanır — webhook, gelen port veya genel IP gerekmez.

**Ajan API anahtarlarım nereye gider?**
Ajan CLI'lerinde, host makinenizde kalırlar; her ajanın kendi dokümanlarına göre yapılandırılır. Telepath yalnızca Telegram bot token'ını ve sohbet allowlist'ini tutar. Bkz. [Güvenlik](security.md).

**Sohbet bazında ajan veya mod değiştirebilir miyim?**
Evet — `/use <agent>`, `/perm <mode>`, `/branch <mode>` o sohbete uygulanır ve süreç yeniden başlayana kadar geçerli kalır.

**Başka bir ajanı nasıl eklerim?**
Yeni bir pakette `Agent` arayüzünü uygulayın ve tek bir satırla kaydedin. Bkz. [Ajan ekleme](adding-an-agent.md).

**Bir dokümantasyon dili nasıl eklerim?**
`docs/en/` klasörünü `docs/<locale>/` olarak kopyalayın, metni çevirin ve dili `mkdocs.yml` ile geçiş bağlantılarına kaydedin. Ayrıntılar için bkz. [CONTRIBUTING.md](../../CONTRIBUTING.md#adding-a-documentation-language).

**Telepath'i nasıl güncellerim?**
Releases sayfasından en son ikili dosyayı indirin ya da kaynaktan yeniden derleyin (`go install …@latest` veya `make build`). Yapılandırma ve `.env` değişmez.

**Maliyetler nedir?**
Telepath ücretsizdir (MIT). Yalnızca kendi ajan hesaplarınız üzerinden kullanılan ajan/model kullanımının bedelini ödersiniz.
