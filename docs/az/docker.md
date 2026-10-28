> **Languages:** [English](../en/docker.md) · [Русский](../ru/docker.md) · [Türkçe](../tr/docker.md) · [Azərbaycanca](../az/docker.md)

# Docker

Telepath iki Dockerfile təqdim edir.

## Müstəqil image (`Dockerfile`, standart)

Çoxmərhələli build: birinci mərhələ statik Go binary-sini qurur (CGO söndürülüb), son mərhələ isə agent CLI-ları və git quraşdırılmış yüngül Node bazasıdır, üstəlik telepath binary-si ilə.

- **Niyə Node?** Claude Code, Gemini CLI və Codex CLI npm paketləri kimi paylanır, ona görə işləmə mühitinə onları işlətmək üçün Node lazımdır.
- **Ölçü.** Bu image daha böyükdür (~200MB+), çünki Node və agent CLI-larını özündə saxlayır. "Qutudan çıxan kimi işləsin" üçün ödənilən qiymətdir.
- Aider Python paketidir; lazım olsa onu image-ə özünüz əlavə edin.

```bash
docker build -t telepath:latest -f Dockerfile .
docker run --rm -it \
  -e TELEGRAM_BOT_TOKEN=xxxx \
  -e ALLOWED_CHAT_IDS=123456789 \
  -e PROJECTS_DIR=/work/project \
  -v /path/to/your/repo:/work/project \
  telepath:latest
```

## Yüngül image (`Dockerfile.slim`, kiçik və sürətli)

Minimal Alpine bazasında yalnız telepath binary-si və git — **Node yoxdur, agentlər yoxdur**. Öz agent CLI-larınızı gətirin: onları mount edin və `*_BIN` mühit dəyişənlərini onlara yönəldin, və ya agentləri olan host üçün `BRANCH_MODE=off` ilə bundan istifadə edin.

```bash
docker build -t telepath:slim -f Dockerfile.slim .
```

## docker compose

`docker-compose.yml` müstəqil image-i işlədir, `.env` faylını oxuyur, dayandırılmadıqca yenidən başlayır və layihələr qovluğunuzu mount edir:

```bash
cp .env.example .env     # fill in token, chat IDs
# set PROJECTS_DIR on the host; it is mounted to /work/project in the container
docker compose up -d
docker compose logs -f
```

Konteyner daxilində `PROJECTS_DIR` `/work/project`-ə təyin olunur — repozitoriyanız orada mount olunur.

## Açarlar və tokenlər haradadır

- **Telegram bot token-i** və **icazəli chat ID-ləri** Telepath konfiqurasiyasıdır — onları env faylı / `-e` ilə ötürün.
- **Agent autentifikasiyası** (API açarları / login-lər) Telepath-a yox, agent CLI-larına aiddir. Onları agentlərin öz mühit dəyişənləri ilə təqdim edin və ya konfiqurasiya qovluqlarını mount edin (məsələn `-v ~/.claude:/root/.claude`). Bax hər agentin sənədləri.
