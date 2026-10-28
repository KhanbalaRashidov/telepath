> **Languages:** [English](../en/docker.md) · [Русский](../ru/docker.md) · [Türkçe](../tr/docker.md) · [Azərbaycanca](../az/docker.md)

# Docker

Telepath iki Dockerfile ile birlikte gelir.

## Bağımsız imaj (`Dockerfile`, varsayılan)

Çok aşamalı bir derleme: ilk aşama statik Go ikilisini derler (CGO kapalı), son aşama ise ajan CLI'leri ve git kurulu hafif bir Node tabanına telepath ikilisini ekler.

- **Neden Node?** Claude Code, Gemini CLI ve Codex CLI npm paketi olarak dağıtılır; bu nedenle çalışma zamanının bunları çalıştırabilmesi için Node'a ihtiyacı vardır.
- **Boyut.** Bu imaj Node ve ajan CLI'lerini içerdiği için daha büyüktür (~200MB+). "Kutudan çıktığı gibi çalışsın" diye ödenen bedel budur.
- Aider bir Python paketidir; ihtiyacınız varsa imaja kendiniz ekleyin.

```bash
docker build -t telepath:latest -f Dockerfile .
docker run --rm -it \
  -e TELEGRAM_BOT_TOKEN=xxxx \
  -e ALLOWED_CHAT_IDS=123456789 \
  -e PROJECTS_DIR=/work/project \
  -v /path/to/your/repo:/work/project \
  telepath:latest
```

## Hafif imaj (`Dockerfile.slim`, küçük ve hızlı)

Minimal bir Alpine tabanı üzerinde yalnızca telepath ikilisi ve git bulunur — **Node yok, ajan yok**. Kendi ajan CLI'lerinizi getirin: bunları bağlayıp `*_BIN` ortam değişkenlerini onlara yönlendirin; ya da bunları zaten içeren bir host'ta `BRANCH_MODE=off` ile kullanın.

```bash
docker build -t telepath:slim -f Dockerfile.slim .
```

## docker compose

`docker-compose.yml` bağımsız imajı çalıştırır, bir `.env` dosyasını okur, durdurulmadıkça yeniden başlatır ve projeler klasörünüzü bağlar:

```bash
cp .env.example .env     # fill in token, chat IDs
# set PROJECTS_DIR on the host; it is mounted to /work/project in the container
docker compose up -d
docker compose logs -f
```

Konteyner içinde `PROJECTS_DIR`, `/work/project` olarak ayarlanır; deponuz buraya bağlanır.

## Anahtarların ve token'ların nereye gittiği

- **Telegram bot token'ı** ve **izinli sohbet kimlikleri** Telepath yapılandırmasıdır — bunları env dosyası veya `-e` ile geçirin.
- **Ajan kimlik doğrulaması** (API anahtarları / oturumlar) Telepath'e değil, ajan CLI'lerine aittir. Bunları ajanların kendi ortam değişkenleriyle sağlayın veya yapılandırma dizinlerini bağlayın (örneğin `-v ~/.claude:/root/.claude`). Ayrıntılar için her ajanın dokümantasyonuna bakın.

---

<!-- prev-next -->
[← Önceki: Mimari](./architecture.md)  ·  [Sonraki: Ajan ekleme →](./adding-an-agent.md)
