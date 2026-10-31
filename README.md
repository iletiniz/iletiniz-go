# İletiniz Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Iletiniz API için resmi Go SDK'si. Go 1.21+ üzerinde çalışır, hiçbir dış bağımlılığı yoktur (yalnızca standart kütüphane).

## Kurulum

```bash
go get github.com/iletiniz/iletiniz-go
```

Gereksinimler:

- Go `>= 1.21`

## Hızlı başlangıç

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/iletiniz/iletiniz-go"
)

func main() {
    client, err := iletiniz.NewClient(os.Getenv("ILETINIZ_API_KEY")) // 'iltz_live_…' veya 'iltz_test_…'
    if err != nil {
        log.Fatal(err)
    }

    res, err := client.Messages.Send(context.Background(), iletiniz.SendMessageParams{
        To:   "+905551234567",
        Body: "Merhaba!",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(res.JobID, res.Status)
}
```

`apiKey` boş geçildiğinde SDK `ILETINIZ_API_KEY` ortam değişkenini okur.

## Yapılandırma

```go
client, _ := iletiniz.NewClient(
    "iltz_live_…",
    iletiniz.WithBaseURL("https://api.iletiniz.com"),    // varsayılan
    iletiniz.WithTimeout(30*time.Second),                  // varsayılan
    iletiniz.WithMaxRetries(2),                            // 408/429/5xx ve ağ hatalarında
    iletiniz.WithDefaultHeaders(map[string]string{"X-Source": "crm"}),
    iletiniz.WithHTTPClient(http.DefaultClient),           // özel *http.Client
)
```

## Endpoint'ler

SDK, public API yüzeyini kapsar:

| Metot                                                | HTTP                              |
| ---------------------------------------------------- | --------------------------------- |
| `client.Health.Check(ctx)`                           | `GET /v1/health`                  |
| `client.Messages.Send(ctx, params)`                  | `POST /v1/messages`               |
| `client.Messages.SendBulk(ctx, params)`              | `POST /v1/messages/bulk`          |
| `client.Messages.Retrieve(ctx, jobID)`               | `GET /v1/messages/{jobID}`        |
| `client.Messages.Status(ctx, jobID)` (alias)         | `GET /v1/messages/{jobID}`        |
| `client.Scheduled.List(ctx, params)`                 | `GET /v1/scheduled`               |
| `client.Scheduled.Retrieve(ctx, id)`                 | `GET /v1/scheduled/{id}`          |
| `client.Scheduled.Cancel(ctx, id)`                   | `DELETE /v1/scheduled/{id}`       |
| `client.Reminders.Create(ctx, params)`               | `POST /v1/reminders`              |
| `client.Reminders.List(ctx, params)`                 | `GET /v1/reminders`               |
| `client.Reminders.Retrieve(ctx, id)`                 | `GET /v1/reminders/{id}`          |
| `client.Reminders.Update(ctx, id, params)`           | `PATCH /v1/reminders/{id}`        |
| `client.Reminders.Delete(ctx, id)`                   | `DELETE /v1/reminders/{id}`       |
| `client.Partner.Profile(ctx)`                        | `GET /v1/partner`                 |
| `client.Partner.CreateAccount(ctx, params)`          | `POST /v1/partner/accounts`       |
| `client.Partner.ListAccounts(ctx, params)`           | `GET /v1/partner/accounts`        |
| `client.Partner.RetrieveAccount(ctx, id)`            | `GET /v1/partner/accounts/{id}`   |
| `client.Partner.SuspendAccount` / `UnsuspendAccount` | `POST /v1/partner/accounts/{id}/suspend` / `unsuspend` |
| `client.Partner.CreateAPIKey(ctx, id, params)`       | `POST /v1/partner/accounts/{id}/api-keys` |
| `client.Partner.ListAPIKeys(ctx, id)`                | `GET /v1/partner/accounts/{id}/api-keys` |
| `client.Partner.RevokeAPIKey(ctx, id, keyID)`        | `DELETE /v1/partner/accounts/{id}/api-keys/{keyId}` |
| `client.Partner.TransferQuota(ctx, id, amount)`      | `POST /v1/partner/accounts/{id}/quota-transfers` |
| `client.Partner.CreateWhatsappSignupLink(ctx, id)`   | `POST /v1/partner/accounts/{id}/whatsapp-signup` |

### Tek mesaj göndermek

```go
res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:       "+905551234567",
    Body:     "Sipariş kodunuz: 4821",
    Sender:   "MAGAZA",   // opsiyonel
    Provider: "netgsm",   // opsiyonel
})
```

### Telegram üzerinden göndermek

`Provider: "telegram"` seçildiğinde `To` alanı SMS yerine Telegram alıcı tanımlayıcısı bekler:
numerik `chat_id` (örn `8409353994`, gruplar için `-1001234567890`) veya `@kullaniciadi`. `Sender` Telegram için kullanılmaz — bot kimliği bağlantıdaki token'a gömülüdür.

```go
res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:       "8409353994",
    Body:     "Merhaba!",
    Provider: "telegram",
})
```

### Sağlayıcılar-arası fallback

Birincil sağlayıcı mesajı **reddederse** (hard-fail: sağlayıcı hata döner veya bağlantı kurulamaz), aynı mesaj (aynı alıcı, aynı içerik, aynı SMS kanalı) sıradaki yedek sağlayıcıyla otomatik yeniden denenir. İlk **başarıda** durur. `Fallback` en fazla 3 sağlayıcı kodundan oluşan sıralı bir slice'tır; hepsi müşterinin bağlı `kind: sms` sağlayıcıları olmalı ve ne birincil ile ne de birbirleriyle aynı olabilir.

```go
res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:       "+905551234567",
    Body:     "Sipariş kodunuz: 4821",
    Provider: "netgsm",                                 // birincil
    Fallback: []string{"verimor", "iletimerkezi"},      // sıralı yedekler (max 3)
})

// res.Provider  → mesajı KABUL eden sağlayıcı
// res.Attempts  → denenen her sağlayıcı + sonucu (opsiyonel)
```

> **Kota tek sayım:** Bir mantıksal mesaj, kaç sağlayıcı denenirse denensin **tek** kota harcar; hepsi başarısız olursa hiç kota harcanmaz.
>
> **Kapsam:** Yalnızca **reddte (hard-fail)** tetiklenir ve yalnızca **SMS→SMS**'tir (kanallar arası değil, örn. WhatsApp→SMS yok). "Teslim edilemedi / timeout" için otomatik fallback henüz yoktur (gelecek sürüm).

`SendBulk` yanıtında kabul eden sağlayıcı, öğe bazında `DeliveredVia` alanında döner.

### Template ile göndermek

```go
res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:       "+905551234567",
    Template: "order_shipped",
    Variables: map[string]any{
        "name":        "Ayşe",
        "tracking_no": "TR123",
    },
})
```

`Body` ve `Template` aynı anda kullanılamaz; tam olarak biri zorunludur. `Variables` yalnızca `Template` ile birlikte verilebilir.

### Toplu gönderim

Tek istekte en fazla 200 öğe gönderebilirsiniz.

```go
// Düz metin modu — her item'da Body zorunlu, Variables yok
res, err := client.Messages.SendBulk(ctx, iletiniz.SendBulkParams{
    Items: []iletiniz.BulkItemInput{
        {To: "+905551111111", Body: "Mesaj 1"},
        {To: "+905552222222", Body: "Mesaj 2"},
    },
})

// Template modu — items'ta Body olmamalı
res, err = client.Messages.SendBulk(ctx, iletiniz.SendBulkParams{
    Template: "low_stock_alert",
    Items: []iletiniz.BulkItemInput{
        {To: "+905551111111", Variables: map[string]any{"product": "Ürün A", "stock": 3}},
        {To: "+905552222222", Variables: map[string]any{"product": "Ürün B", "stock": 1}},
    },
})
```

### Onaylı WhatsApp şablonu göndermek

Meta'da onaylanmış (APPROVED) şablonu `WhatsappTemplate` ile gönderin. `Components`, Meta Cloud API'nin `template.components` biçimini izler. WhatsApp'ta `Sender`, gönderici numaranın Meta `phone_number_id` değeridir.

```go
res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:       "+905551234567",
    Provider: "whatsapp",
    Sender:   "123456789012345",
    WhatsappTemplate: &iletiniz.WhatsappTemplate{
        Name:     "randevu_hatirlatma",
        Language: "tr",
        Components: []map[string]any{
            {"type": "body", "parameters": []map[string]any{
                {"type": "text", "text": "Ayşe"},
                {"type": "text", "text": "5 Kasım 10:00"},
            }},
        },
    },
})
```

Toplu gönderimde alıcı başına değişkenler `BulkItemInput.WhatsappComponents` ile verilir.

### Zamanlanmış gönderim

`ScheduledAt` (saat dilimli RFC 3339, en az 30 sn, en fazla 90 gün sonrası) verilirse mesaj şimdi gönderilmez; yanıtta `Status == "scheduled"` ve `ScheduledID` döner.

```go
job, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
    To:          "+905551234567",
    Body:        "Yarınki randevunuzu hatırlatırız.",
    ScheduledAt: time.Date(2026, 11, 5, 9, 0, 0, 0, istanbul).Format(time.RFC3339),
})
list, err := client.Scheduled.List(ctx, iletiniz.ListParams{Status: "scheduled"})
_, err = client.Scheduled.Cancel(ctx, job.ScheduledID)
```

### Hatırlatma kuralları

```go
// Randevudan 24 saat ve 2 saat önce
_, err := client.Reminders.Create(ctx, iletiniz.CreateReminderParams{
    Type:           "one_time",
    Name:           "Randevu #4411",
    EventAt:        "2026-11-10T14:00:00+03:00",
    OffsetsMinutes: []int{1440, 120},
    Message:        iletiniz.ReminderMessage{To: "+905551234567", Body: "Randevunuz 10 Kasım 14:00."},
})

// Her pazartesi 09:30 (Europe/Istanbul)
monday := 1
rule, err := client.Reminders.Create(ctx, iletiniz.CreateReminderParams{
    Type: "recurring", Name: "Haftalık bülten",
    Frequency: "weekly", DayOfWeek: &monday, TimeOfDay: "09:30",
    Message: iletiniz.ReminderMessage{To: "+905551234567", Template: "haftalik_bulten"},
})
_, err = client.Reminders.Update(ctx, rule.ID, iletiniz.UpdateReminderParams{Status: "paused"})
_, err = client.Reminders.Delete(ctx, rule.ID)
```

### Partner API

Partner hesabının `partner:manage` yetkili canlı anahtarıyla:

```go
partner, _ := iletiniz.NewClient(os.Getenv("ILETINIZ_PARTNER_KEY"))

child, err := partner.Partner.CreateAccount(ctx, iletiniz.CreateChildAccountParams{Name: "Klinik Ankara", ExternalRef: "klinik-42"})
key, err := partner.Partner.CreateAPIKey(ctx, child.ID, iletiniz.CreateChildAPIKeyParams{Name: "prod"})
// key.Key yalnız bu yanıtta döner; güvenli şekilde saklayın.

// Alt hesabın anahtarlarını listeleyin (düz metin anahtar dönmez, yalnız KeyPrefix) ve gerekirse iptal edin:
keys, err := partner.Partner.ListAPIKeys(ctx, child.ID)
_, err = partner.Partner.RevokeAPIKey(ctx, child.ID, keys.Data[0].ID) // idempotent: iptal edilmişse aynı nesne döner

_, err = partner.Partner.TransferQuota(ctx, child.ID, 5000)
link, err := partner.Partner.CreateWhatsappSignupLink(ctx, child.ID)
// link.URL müşterinize iletilir (24 saat geçerli, tek kullanımlık).

// Alt hesap adına gönderim alt hesabın anahtarıyla yapılır:
childClient, _ := iletiniz.NewClient(key.Key)
```

`TransferQuota`, `CreateAccount` ve `CreateAPIKey` hata durumunda otomatik yeniden denenmez.

### Mesaj durumunu sorgulamak

```go
info, err := client.Messages.Retrieve(ctx, jobID)
// info.Status: iletiniz.StatusSent | StatusQueued | StatusFailed | StatusDelivered | …
```

### Sağlık kontrolü

```go
health, err := client.Health.Check(ctx)
// health.OK == true, health.DB == "up"
```

## Hata yönetimi

Tüm hatalar `errors.Is` ile sentinel'lara karşı kontrol edilebilir. API hataları `*iletiniz.APIError` ile `errors.As` üzerinden detaylandırılabilir.

```go
import "errors"

_, err := client.Messages.Send(ctx, params)
switch {
case errors.Is(err, iletiniz.ErrAuthentication):
    // 401
case errors.Is(err, iletiniz.ErrValidation):
    // 400 / 422
    var apiErr *iletiniz.APIError
    if errors.As(err, &apiErr) {
        log.Println(apiErr.Code, apiErr.Body)
    }
case errors.Is(err, iletiniz.ErrRateLimit):
    // 429
case errors.Is(err, iletiniz.ErrNotFound):
    // 404
case errors.Is(err, iletiniz.ErrServer):
    // 5xx
case errors.Is(err, iletiniz.ErrTimeout):
    // istek timeout'a takıldı
case errors.Is(err, iletiniz.ErrConnection):
    // ağ hatası
case errors.Is(err, iletiniz.ErrInvalidRequest):
    // istemci tarafı parametre validasyonu
}
```

## Yeniden deneme stratejisi

SDK, aşağıdaki durumlarda otomatik olarak `WithMaxRetries(n)` defa yeniden dener (varsayılan: 2):

- Ağ kaynaklı bağlantı hataları
- HTTP 408, 429, 500–599

`Retry-After` başlığı varsa beklenir; aksi halde exponential backoff (jitter ile) uygulanır. Yeniden denemeyi kapatmak için `WithMaxRetries(0)` verin.

## Timeout

Client default'unu ezen istek bazlı timeout:

```go
res, err := client.Messages.Send(
    ctx,
    iletiniz.SendMessageParams{To: "+905551234567", Body: "merhaba"},
    iletiniz.WithRequestTimeout(10*time.Second),
)
```

## Test

SDK, `*http.Client` enjeksiyonu üzerinden HTTP katmanını dışarı açar. `httptest.NewServer` ile gerçek ağ trafiği oluşturmadan SDK'yı test edebilirsiniz:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte(`{"ok":true,"db":"up"}`))
}))
defer srv.Close()

client, _ := iletiniz.NewClient(
    "iltz_test_xxx",
    iletiniz.WithBaseURL(srv.URL),
    iletiniz.WithHTTPClient(srv.Client()),
)
```

## Katkıda Bulunma / Contributing

Katkı sağlamak ister misiniz? Lütfen [CONTRIBUTING.md](./CONTRIBUTING.md) dosyasını inceleyin. English: [CONTRIBUTING.en.md](./CONTRIBUTING.en.md).

## Davranış Kuralları / Code of Conduct

Bu proje [Contributor Covenant](./CODE_OF_CONDUCT.md) davranış kurallarına bağlıdır. English: [CODE_OF_CONDUCT.en.md](./CODE_OF_CONDUCT.en.md).

## Güvenlik / Security

Güvenlik açığı bildirmek için lütfen [SECURITY.md](./SECURITY.md) dosyasındaki adımları izleyin — **public issue açmayın**. English: [SECURITY.en.md](./SECURITY.en.md).

## Lisans / License

MIT — bkz. / see [LICENSE](./LICENSE).
