> **Languages:** [English](../en/branch-workflow.md) · [Русский](../ru/branch-workflow.md) · [Türkçe](../tr/branch-workflow.md) · [Azərbaycanca](../az/branch-workflow.md)

# Branch iş axını

Telepath bütün git əməliyyatlarına sahiblik edir. Agentlərin öz git və avtomatik commit funksiyaları söndürülüb ki, versiya nəzarəti ardıcıl və təhlükəsiz olsun. Rejim `BRANCH_MODE` ilə təyin olunur.

## `per-task` (standart)

Hər tapşırıq öz təcrid olunmuş branch-ında işləyir:

1. **Preflight.** Working tree təmiz olmalıdır. Çirkli olarsa, Telepath aydın mesajla imtina edir — `BRANCH_STASH_DIRTY=true` olmadıqda; bu halda dəyişiklikləri əvvəl stash edir. Baza branch-ını və baza commit SHA-nı qeyd edir.
2. **Branch yaratmaq.** `<BRANCH_PREFIX><slug>-<yymmdd-hhmm>` adında branch yaradır və ona keçir, məsələn `telepath/add-health-endpoint-260115-1430`. Slug tapşırıq mətnindən alınır (kiçik hərflə, alfanumerik olmayan simvollar `-`-yə yığılır, kəsilir).
3. **Agenti işlədin.** Agent yalnız fayl redaktə edir; onun git-i söndürülüb.
4. **Branch-da commit.** Uğurlu olduqda, `BRANCH_COMMIT=true` (standart) isə Telepath hər şeyi stage edir və **tapşırıq branch-ında** qısa tapşırıq xülasəsi və aşağıdakı trailer ilə commit edir:

   ```
   <summary>

   via Telepath (<agent>)
   ```

   Dəyişiklik yoxdursa, bunu bildirir və commit etmir.
5. **main-ə heç vaxt toxunmur.** Heç vaxt push, merge, rebase, məcburi bir əməliyyat və ya `main`/`master`-ə commit etmir.
6. **Hesabat.** Branch adını, commit SHA-nı, `git diff <base>..HEAD --stat`-ı və tam diff-i `.diff` sənədi kimi, **Keep branch** / **Delete branch & restore base** düymələri ilə göndərir.

Branch-ı nəzərdən keçirir və **özünüz merge/push edirsiniz**.

### `BRANCH_COMMIT=false` kompromisi

Əgər `BRANCH_COMMIT=false` olsa, agentin dəyişiklikləri tapşırıq branch-ında **commit edilmədən** qalır. Telepath bunu sizə bildirir. Bütün chat-lar eyni working tree-ni paylaşdığı üçün commit edilməmiş dəyişikliklər zəif təcriddir: sonrakı tapşırığın preflight-ı çirkli tree görəcək və imtina edəcək (və ya stash edəcək). Xüsusi səbəbiniz olmadıqca standart `true`-ni üstün tutun.

## `off`

Telepath cari branch-da işləyir, branch yaratmır və commit etmir. Yenə də canlı axını və **iş sonu working-tree diff-ini** alırsınız. Bu, branch-ları özünüz idarə etdiyiniz hallar üçün köhnə / minimal rejimdir.

## Sərt qoruyucular

Hər git əmri tək qoruyucudan (`internal/git/guard.go`) keçir və bunları **rədd edir**:

- `push`, `merge`, `rebase`, `pull`, `fetch`, `cherry-pick`, `revert`
- hər hansı `--force` / `-f` / `--force-with-lease`
- `reset --hard`
- `main` / `master`-ə commit etmək, yaratmaq və ya silmək

**Zəmanət: `main` və remote heç vaxt dəyişdirilmir. Bir tapşırığın edə biləcəyi ən pis şey yoxlanıla bilən yerli branch qoymaqdır.**

## Təmizlənmə

- **Keep branch** onu nəzərdən keçirmək üçün saxlayır.
- **Delete branch & restore base** tapşırıq branch-ının dəyişikliklərini ləğv edir, baza branch-ını yenidən checkout edir və tapşırıq branch-ını silir. (Tapşırıq işləyərkən rədd edilir — əvvəlcə `/cancel`.)
- `/branches` istənilən vaxt mövcud `telepath/*` branch-larını siyahılayır.
