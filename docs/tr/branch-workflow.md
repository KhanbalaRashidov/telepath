> **Languages:** [English](../en/branch-workflow.md) · [Русский](../ru/branch-workflow.md) · [Türkçe](../tr/branch-workflow.md) · [Azərbaycanca](../az/branch-workflow.md)

# Branch iş akışı

Telepath tüm git işlemlerinin sahibidir. Ajanların kendi git ve otomatik commit özellikleri, sürüm kontrolünün tutarlı ve güvenli kalması için devre dışıdır. Mod, `BRANCH_MODE` ile ayarlanır.

## `per-task` (varsayılan)

Her görev kendi izole branch'inde çalışır:

1. **Ön kontrol.** Çalışma ağacı temiz olmalıdır. Kirliyse Telepath açık bir mesajla reddeder — `BRANCH_STASH_DIRTY=true` ise bu durumda hariç; o zaman önce değişiklikleri stash'ler. Temel branch'i ve temel commit SHA'sını kaydeder.
2. **Branch oluştur.** `<BRANCH_PREFIX><slug>-<yymmdd-hhmm>` adında bir branch oluşturur ve üzerine geçer, ör. `telepath/add-health-endpoint-260115-1430`. Slug, görev metninden türetilir (küçük harfe çevrilir, alfanümerik olmayan karakterler `-` ile birleştirilir ve kısaltılır).
3. **Ajanı çalıştırın.** Ajan yalnızca dosyaları düzenler; onun git'i kapalıdır.
4. **Branch'te commit.** Başarıda, `BRANCH_COMMIT=true` ise (varsayılan), Telepath her şeyi stage'ler ve **görev branch'inde** şu biçimde bir mesajla commit'ler: kısa bir görev özeti ve ardından şu satır:

   ```
   <summary>

   via Telepath (<agent>)
   ```

   Değiştirilecek bir şey yoksa bunu belirtir ve commit yapmaz.
5. **Main'e asla dokunmaz.** Asla push, merge, rebase veya zorla herhangi bir işlem yapmaz ve `main`/`master` üzerine commit yapmaz.
6. **Raporlama.** Branch adını, commit SHA'sını, `git diff <base>..HEAD --stat` çıktısını ve tam diff'i bir `.diff` belgesi olarak, **Keep branch** / **Delete branch & restore base** düğmeleriyle birlikte gönderir.

Branch'i siz incelersiniz ve **kendiniz merge/push edersiniz**.

### `BRANCH_COMMIT=false` ödünleşimi

`BRANCH_COMMIT=false` ise ajanın değişiklikleri görev branch'inde **commit edilmemiş** halde bırakılır ve Telepath bunu size bildirir. Tüm sohbetler tek bir çalışma ağacını paylaştığından, commit edilmemiş değişiklikler daha zayıf bir izolasyon sağlar: sonraki bir görevin ön kontrolü kirli bir ağaç görür ve reddeder (veya stash'ler). Özel bir nedeniniz yoksa varsayılan `true` değerini tercih edin.

## `off`

Telepath mevcut branch üzerinde çalışır, branch oluşturmaz ve commit yapmaz. Yine de canlı akışı ve **çalışma sonu çalışma ağacı diff'ini** alırsınız. Bu, branch'leri kendiniz yönettiğinizde kullanılan eski/minimal moddur.

## Sert korumalar

Her git komutu tek bir koruma noktasından (`internal/git/guard.go`) geçer ve şunları **reddeder**:

- `push`, `merge`, `rebase`, `pull`, `fetch`, `cherry-pick`, `revert`
- herhangi bir `--force` / `-f` / `--force-with-lease`
- `reset --hard`
- `main` / `master` üzerine commit yapma, oluşturma veya silme

**Garanti: `main` ve uzak depo asla değiştirilmez. Bir görevin yapabileceği en kötü şey, incelenebilir yerel bir branch bırakmaktır.**

## Temizlik

- **Keep branch** branch'i incelemeniz için bırakır.
- **Delete branch & restore base** görev branch'indeki değişiklikleri atar, temel branch'i yeniden checkout eder ve görev branch'ini siler. (Bir görev çalışırken reddedilir — önce `/cancel` kullanın.)
- `/branches` mevcut `telepath/*` branch'lerini istediğiniz zaman listeler.

---

<!-- prev-next -->
[← Önceki: İzinler](./permissions.md)  ·  [Sonraki: Mimari →](./architecture.md)
