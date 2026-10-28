> **Languages:** [English](../en/index.md) · [Русский](../ru/index.md) · [Türkçe](../tr/index.md) · [Azərbaycanca](../az/index.md)

# Telepath — Dokümantasyon

Telepath, Telegram'dan gönderdiğiniz kodlama görevlerini kendi makinenizde çalışan bir kodlama ajanına ileten, kendi sunucunuzda barındırdığınız bir köprüdür. Ajanın ilerlemesini gerçek zamanlı olarak geri aktarır: ne yaptığını, hangi dosyaları düzenlediğini, hangi komutları çalıştırdığını, onay isteklerini ve tam git diff'ini içeren nihai özeti. **Ajan-bağımsızdır** (Claude Code, Gemini CLI, OpenAI Codex CLI ve Aider tek bir arayüz arkasında) ve her göreve kendi **git branch'ini** verir; isteğe bağlı olarak Telegram üzerinden canlı **onay** de alabilirsiniz.

## İçindekiler

| Sayfa | İçerik |
|-------|--------|
| [Başlarken](getting-started.md) | Ön koşullar, işletim sistemine göre kurulum, bot oluşturma, ilk çalıştırma, ilk görev |
| [Yapılandırma](configuration.md) | Her ortam değişkeni ve CLI bayrağı; varsayılanlar ve öncelik sırası |
| [Transport'lar](transports.md) | Telegram, Slack veya ikisi birden; nasıl seçilir; genel IP gerektirmeme özelliği |
| [Telegram komutları](telegram-commands.md) | Her komut ve satır içi düğme (Telegram ve Slack), örneklerle |
| [Ajanlar](agents.md) | Çoklu ajan modeli, yetenek matrisi ve tam CLI çağrıları |
| [İzinler](permissions.md) | `auto` / `allowlist` / `interactive` ve canlı onayın çalışma şekli |
| [Branch iş akışı](branch-workflow.md) | Göreve özel branch'ler, commit'ler, korumalar ve temizlik |
| [Mimari](architecture.md) | Pipeline, olay modeli, paket haritası, MCP sunucusu |
| [Docker](docker.md) | Bağımsız ve hafif imajlar, compose, mount'lar |
| [Ajan ekleme](adding-an-agent.md) | `Agent` arayüzünü uygulayın ve yeni bir adaptörü kaydedin |
| [Güvenlik](security.md) | Allowlist, ajan kimlik doğrulaması, `auto` uyarısı, git garantisi |
| [Sorun giderme](troubleshooting.md) | Yaygın hatalar ve çözümleri |
| [SSS](faq.md) | Sık sorulara kısa yanıtlar |

## Akılda tutmanız gereken tek şey

**Telepath asla push yapmaz, asla merge yapmaz ve `main`/`master` dalına ya da uzak depoya asla dokunmaz.** Bir görevin yapabileceği en kötü şey, sizin merge edip kendiniz push edebileceğiniz, incelenebilir yerel bir branch bırakmaktır.

---

<!-- prev-next -->
[Sonraki: Başlangıç →](./getting-started.md)
