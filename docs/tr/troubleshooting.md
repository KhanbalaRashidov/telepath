> **Languages:** [English](../en/troubleshooting.md) · [Русский](../ru/troubleshooting.md) · [Türkçe](../tr/troubleshooting.md) · [Azərbaycanca](../az/troubleshooting.md)

# Sorun giderme

Daha fazla ayrıntı için `LOG_LEVEL=debug` ayarlayın. Aşağıda yaygın hata modları yer alır.

## "Agent is not installed"

`/agents` her ajan için ✅/❌ gösterir. Sizinki ❌ ise:

- CLI'nin PATH üzerinde olduğunu doğrulayın (`claude --version`, `gemini --version`, `codex --version`, `aider --version`).
- Standart dışı bir yerdeyse, ilgili `*_BIN` değişkenini (`CLAUDE_BIN`, `GEMINI_BIN`, `CODEX_BIN`, `AIDER_BIN`) tam yoluna ayarlayın.
- Docker'da bağımsız imajı kullanın (CLI'leri zaten içerir) ya da CLI'leri hafif imaja bağlayın.

## Bot yanıt vermiyor

- **Token:** `TELEGRAM_BOT_TOKEN` değerini kontrol edin. Başlangıçta Telepath token'ı Telegram ile doğrular; yanlış bir token hemen başarısız olur.
- **Allowlist:** sohbet kimliğiniz `ALLOWED_CHAT_IDS` içinde olmalıdır. Listede olmayan sohbetler sessizce yok sayılır (günlüklerdeki uyarıya bakın).
- **Bağlantı:** Telepath giden long-polling kullanır; host'un `api.telegram.org` adresine erişebildiğinden emin olun.
- **Başka bir yerde çalışıyor:** Bir token için aynı anda yalnızca bir bot süreci long-polling yapmalıdır.

## "Working tree is dirty" reddi

`per-task` modunda, `PROJECTS_DIR` içinde commit edilmemiş değişiklikler varsa Telepath görevi başlatmayı reddeder. Ya değişiklikleri kendiniz commit/stash edin ya da Telepath'in bunları otomatik stash'lemesi için `BRANCH_STASH_DIRTY=true` ayarlayın.

## Onay zaman aşımına uğradı → reddedildi

`interactive` modunda, `APPROVAL_TIMEOUT_SECONDS` (varsayılan 120) içinde kimse Approve/Deny düğmesine dokunmazsa istek **reddedilir** ve ajana bu bildirilir. Daha fazla süreye ihtiyacınız varsa zaman aşımını artırın; istem almak istemiyorsanız `allowlist`/`auto` kullanın.

## Etkileşimli mod bana sormadı

Etkileşimli onay **yalnızca Claude ile** çalışır. gemini/codex/aider ile `interactive` `allowlist`'e geri döner ve yanıt, gerçekte kullanılan modu belirtir. Canlı onaylar için sohbeti `/use claude` ile Claude'a geçirin.

## Görevler kuyrukta bekliyor gibi

Tüm sohbetler tek bir `PROJECTS_DIR` çalışma ağacını paylaştığından görevler **sıraya alınır** — tüm sohbetler genelinde aynı anda bir tane, sohbet başına da bir tane. Çalışma alanı meşgulse "tekrar deneyin" yanıtı alırsınız. Neyin çalıştığını görmek için `/status`, durdurmak için `/cancel` kullanın.

## `/cancel` her şeyi sonlandırmadı (Windows)

Telepath, ajan süreç **ağacını** en iyi çabayla sonlandırır — Windows'ta `taskkill /F /T`, Unix'te süreç grubu `SIGKILL`. Derinlemesine iç içe geçmiş yardımcı süreçler ara sıra kalabilir; bir Node süreci hayatta kalırsa Görev Yöneticisi'ni kontrol edin.

## `/resume` devam ettirilecek bir şey olmadığını söylüyor

Son oturum kimliği ve proje dizini **bellekte** tutulur ve Telepath yeniden başlatıldığında kaybolur. Yeniden başlatmadan sonra devam ettirilecek bir şey yoktur. Ayrıca devam ettirme yalnızca **Claude** içindir; diğer ajanlar bunu desteklenmiyor olarak bildirir.

## Commit başarısız

Repo bir git kimliğine sahip değilse Telepath, görev branch'indeki commit'in başarılı olması için `Telepath <telepath@localhost>` yedek kimliğiyle commit yapar. Yine de commit başarısız olursa hata nihai raporda gösterilir — deponun git durumunu ve izinlerini kontrol edin.
