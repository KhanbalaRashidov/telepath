> **Languages:** [English](../en/agents.md) · [Русский](../ru/agents.md) · [Türkçe](../tr/agents.md) · [Azərbaycanca](../az/agents.md)

# Ajanlar

Telepath, birden çok kodlama ajanı CLI'sini ortak tek bir arayüz arkasından yönetir. Her adaptör, satıcıya özgü CLI'yi aynı normalleştirilmiş olay akışına çevirir; böylece Telegram ve git katmanları ajana özgü ayrıntıları hiç görmez. Sohbet bazında ajan değiştirmek için `/use` komutunu, tek bir görevi belirli bir ajanla çalıştırmak için ise `/<agent> <task>` komutunu kullanın.

## Yetenek matrisi

| Ajan | Kurulum | Akış (FileEdit olayları) | Oturumlar / devam ettirme | Etkileşimli onay | Durum |
|------|---------|:------------------------:|:-------------------------:|:----------------:|-------|
| **claude** | `npm install -g @anthropic-ai/claude-code` | ✅ tam (stream-json, düzenleme başına diff'ler) | ✅ | ✅ (MCP prompt aracı) | Tam referans adaptörü |
| **gemini** | `npm install -g @google/gemini-cli` | — (metin; düzenlemeler çalışma sonu diff'i ile) | — | — → `allowlist`'e geri döner | Destekleniyor |
| **codex** | `npm install -g @openai/codex` | JSONL (en iyi çaba; düzenlemeler çalışma sonu diff'i ile) | — | — → `allowlist`'e geri döner | Destekleniyor |
| **aider** | `pipx install aider-chat` | — (metin; düzenlemeler çalışma sonu diff'i ile) | — | — → `allowlist`'e geri döner | Destekleniyor |

Telegram'daki `/agents` komutu bu matrisi, ana makinenizdeki canlı kurulum durumuyla birlikte yazdırır.

## Tam çağrılar

Bunlar her adaptörün oluşturduğu gerçek komutlardır. Köşeli parantez içindeki bayraklar yalnızca ilgili olduğunda eklenir (`MAX_TURNS>0` ise `--max-turns`, `MODEL` ayarlıysa `--model` ve izin moduna göre izin bayrakları).

### Claude Code (referans)

```bash
claude -p "<task>" \
  --output-format stream-json --verbose --include-partial-messages \
  [--max-turns <N>] [--model <model>] \
  # auto:        --permission-mode bypassPermissions
  # allowlist:   --permission-mode default --allowedTools "<list>" --disallowedTools "<list>"
  # interactive: --permission-mode default \
  #              --permission-prompt-tool mcp__telepath__approve --mcp-config <file>
```

Devam ettirme:

```bash
claude --resume <session_id> -p "<message>" \
  --output-format stream-json --verbose --include-partial-messages \
  --permission-mode bypassPermissions
```

Claude tam olay setini üretir: asistan metni, `tool_use` → `FileEdit` (Edit/Write/MultiEdit, birleşik diff ile), `Bash` → `Command`, `tool_result` ve özeti, süreyi, maliyeti ve oturum kimliğini taşıyan son bir `result` olayı.

### Gemini CLI

```bash
gemini -p "<task>" [-m <model>] \
  # auto:  --approval-mode=yolo
  # other: --approval-mode=default
```

Gemini çıktısı metin olarak okunur (JSON çıktısı sürümler arasında kararsız olmuştur). Satırlar `AssistantText` olur; dosya değişiklikleri git katmanının çalışma sonu diff'i ile gösterilir.

### OpenAI Codex CLI

```bash
codex exec --json \
  # auto:  --sandbox workspace-write --ask-for-approval never
  # other: --sandbox workspace-write --ask-for-approval on-request
  [--model <model>] "<task>"
```

> Eski `--full-auto` bayrağı kullanımdan kaldırılmıştır; Telepath bunun yerine `--sandbox
> workspace-write` kullanır. Codex, JSONL olay akışı üretir; bu akış en iyi çabayla
> `AssistantText`/`Command` olaylarına ayrıştırılır; dosya değişiklikleri de çalışma sonu diff'inde görünür.

### Aider

```bash
aider --message "<task>" --no-stream [--model <model>] \
  --yes --no-auto-commits --no-git
```

Aider'ın kendi git entegrasyonu **devre dışıdır** (`--no-auto-commits --no-git`), böylece sürüm kontrolü Telepath'te kalır ve ajan etkileşimsiz çalışır (`--message --yes`). Çıktı metindir; dosya değişiklikleri çalışma sonu diff'inde görünür.

## Dürüst sınırlar (lütfen okuyun)

- **Etkileşimli onay yalnızca Claude ile çalışır** (MCP izin-istemi kancası gerektirir). gemini/codex/aider için `interactive` seçildiğinde otomatik olarak `allowlist`'e geri dönülür ve yanıt, gerçekte hangi modun kullanıldığını bildirir.
- **gemini/codex/aider çalışma sırasında düzenleme başına `FileEdit` olayı üretmez** — dosya değişiklikleri bunun yerine git katmanının çalışma sonu diff'i ile gösterilir.
- **Devam ettirme yalnızca Claude içindir.** Diğer ajanlar `/resume` için "desteklenmiyor" yanıtı verir.
- **Codex JSONL ayrıştırması en iyi çabayla yapılır**, çünkü şema sürümler arasında değişir; tanınmayan satırlar yok sayılır (veya metin olarak gösterilir).
- Her ajan CLI'si, sizin kurup kimlik doğrulaması yaptığınız **dış bir bağımlılıktır**. Bir CLI eksikse `/agents` onu kurulu değil olarak gösterir ve o ajana yönelik görevler açık bir mesajla reddedilir.

Kendi adaptörünüzü yazmak için bkz. [Ajan ekleme](adding-an-agent.md).

---

<!-- prev-next -->
[← Önceki: Telegram komutları](./telegram-commands.md)  ·  [Sonraki: İzinler →](./permissions.md)
