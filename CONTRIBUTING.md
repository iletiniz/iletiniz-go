# İletiniz Go SDK'ya Katkı

Katkıda bulunmak istediğiniz için teşekkür ederiz! Her türlü katkıyı memnuniyetle karşılarız.

## Geliştirme Ortamı

```bash
git clone https://github.com/iletiniz/iletiniz-go.git
cd iletiniz-go
go mod download
```

## Kod Stili

```bash
go vet ./...
gofmt -w .
```

## Test Çalıştırma

```bash
go test ./...
```

## Commit Mesajı Kuralları

[Conventional Commits](https://www.conventionalcommits.org/) standardını kullanıyoruz:

- `feat:` yeni özellik
- `fix:` hata düzeltmesi
- `docs:` yalnızca belge değişikliği
- `chore:` yapılandırma, bağımlılık güncellemesi vb.
- `refactor:` davranış değiştirmeyen kod yeniden düzenlemesi
- `test:` test ekleme veya düzeltme
- `build:` derleme sistemi veya dış bağımlılıklar

## Pull Request Süreci

1. Bu repoyu fork edin.
2. `git checkout -b feat/ozellik-adi` ile yeni bir dal oluşturun.
3. Değişikliklerinizi commit edin.
4. Dalınızı kendi fork'unuza push edin.
5. GitHub üzerinden bir Pull Request açın.

## Issue Bildirme

Lütfen aşağıdaki bilgileri ekleyin:

- Net ve açıklayıcı bir başlık
- Sorunu yeniden üretme adımları
- Beklenen ve gerçekleşen davranış
- SDK sürümü ve Go sürümü

## İletişim

support@iletiniz.com
