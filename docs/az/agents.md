> **Languages:** [English](../en/agents.md) · [Русский](../ru/agents.md) · [Türkçe](../tr/agents.md) · [Azərbaycanca](../az/agents.md)

# Agentlər

Telepath bir neçə kodlaşdırma-agent CLI-nı ümumi interfeys arxasında idarə edir. Hər adapter satıcı CLI-nı eyni normallaşdırılmış hadisə axınına çevirir, beləliklə Telegram və git qatları agentə xas detalları heç vaxt görmür. Agenti chat üzrə `/use` ilə dəyişin və ya `/<agent> <task>` ilə konkret agentlə bir tapşırıq icra edin.

## İmkanlar cədvəli

| Agent | Quraşdırma | Streaming (FileEdit hadisələri) | Sessiyalar / resume | İnteraktiv təsdiq | Status |
|-------|------------|:-------------------------------:|:-------------------:|:-----------------:|--------|
| **claude** | `npm install -g @anthropic-ai/claude-code` | ✅ tam (stream-json, hər redaktə üzrə diff-lər) | ✅ | ✅ (MCP prompt aləti) | Tam istinad adapteri |
| **gemini** | `npm install -g @google/gemini-cli` | — (mətn; redaktələr iş sonu diff-i ilə) | — | — → `allowlist`-ə geri düşür | Dəstəklənir |
| **codex** | `npm install -g @openai/codex` | JSONL (ən yaxşı cəhdlə; redaktələr iş sonu diff-i ilə) | — | — → `allowlist`-ə geri düşür | Dəstəklənir |
| **aider** | `pipx install aider-chat` | — (mətn; redaktələr iş sonu diff-i ilə) | — | — → `allowlist`-ə geri düşür | Dəstəklənir |

Telegram-da `/agents` bu cədvəli host üçün canlı quraşdırma statusu ilə çap edir.

## Dəqiq çağırışlar

Bunlar hər adapterin qurduğu real əmrlərdir. Kvadrat mötərizədəki bayraqlar yalnız uyğun olduqda əlavə olunur (`MAX_TURNS>0` olduqda `--max-turns`, `MODEL` təyin olunduqda `--model`, və hər rejim üzrə icazə bayraqları).

### Claude Code (istinad)

```bash
claude -p "<task>" \
  --output-format stream-json --verbose --include-partial-messages \
  [--max-turns <N>] [--model <model>] \
  # auto:        --permission-mode bypassPermissions
  # allowlist:   --permission-mode default --allowedTools "<list>" --disallowedTools "<list>"
  # interactive: --permission-mode default \
  #              --permission-prompt-tool mcp__telepath__approve --mcp-config <file>
```

Resume:

```bash
claude --resume <session_id> -p "<message>" \
  --output-format stream-json --verbose --include-partial-messages \
  --permission-mode bypassPermissions
```

Claude tam hadisə dəstini yayır: assistant mətni, `tool_use` → `FileEdit` (Edit/Write/MultiEdit, unified diff ilə), `Bash` → `Command`, `tool_result`, və xülasə, müddət, dəyər və sessiya id-sini daşıyan son `result`.

### Gemini CLI

```bash
gemini -p "<task>" [-m <model>] \
  # auto:  --approval-mode=yolo
  # other: --approval-mode=default
```

Gemini çıxışı mətn kimi oxunur (onun JSON çıxışı buraxılışlar arasında qeyri-sabit olub). Sətirlər `AssistantText` olur; fayl dəyişiklikləri git qatının iş sonu diff-i ilə göstərilir.

### OpenAI Codex CLI

```bash
codex exec --json \
  # auto:  --sandbox workspace-write --ask-for-approval never
  # other: --sandbox workspace-write --ask-for-approval on-request
  [--model <model>] "<task>"
```

> Köhnə `--full-auto` bayrağı artıq istifadədən çıxarılıb; Telepath `--sandbox
> workspace-write` istifadə edir. Codex JSONL hadisə axını yayır, bu da ən yaxşı
> cəhdlə `AssistantText`/`Command`-a çevrilir; fayl dəyişiklikləri iş sonu
> diff-ində də görünür.

### Aider

```bash
aider --message "<task>" --no-stream [--model <model>] \
  --yes --no-auto-commits --no-git
```

Aider-in öz git inteqrasiyası **söndürülüb** (`--no-auto-commits --no-git`), beləliklə versiya nəzarətinə Telepath sahiblik edir və qeyri-interaktiv işləyir (`--message --yes`). Çıxış mətndir; fayl dəyişiklikləri iş sonu diff-ində görünür.

## Dürüst məhdudiyyətlər (mütləq oxuyun)

- **İnteraktiv təsdiq yalnız Claude ilə işləyir** (MCP icazə-prompt hook-u tələb edir). gemini/codex/aider üçün `interactive` seçildikdə avtomatik `allowlist`-ə keçilir və cavab hansı rejimin faktiki istifadə olunduğunu sizə bildirir.
- **gemini/codex/aider icra zamanı hər redaktə üzrə `FileEdit` hadisələri yaymır** — onların fayl dəyişiklikləri əvəzinə git qatının iş sonu diff-i ilə göstərilir.
- **Resume yalnız Claude üçündür.** Digər agentlər `/resume` üçün "dəstəklənmir" cavabı verir.
- **Codex JSONL təhlili ən yaxşı cəhddir**, çünki sxem versiyalar arasında dəyişir; tanınmayan sətirlər nəzərə alınmır (və ya mətn kimi göstərilir).
- Hər agent CLI-ı sizin özünüz quraşdırdığınız və autentifikasiya etdiyiniz **xarici asılılıqdır**. CLI yoxdursa, `/agents` onu quraşdırılmamış kimi göstərir və həmin agent üçün tapşırıqlar aydın mesajla rədd edilir.

Öz adapterinizi yazmaq üçün bax: [Agent əlavə etmək](adding-an-agent.md).
