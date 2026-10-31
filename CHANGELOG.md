# Changelog / Değişiklik Günlüğü

Bu projedeki önemli değişiklikler bu dosyada belgelenir.
All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
Sürümleme / Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

## [Unreleased]

### Added
- `client.Partner.ListAPIKeys` ve `client.Partner.RevokeAPIKey`: alt hesabın API anahtarlarını listeleme ve iptal etme; `ChildAPIKeyInfo` / `ChildAPIKeyList` tipleri.

## [0.2.0] - 2026-10-31

### Eklenenler / Added
- Meta onaylı WhatsApp şablonu (`whatsappTemplate`), bulk alıcı başına `whatsappComponents`, WhatsApp için `sender` = `phone_number_id`.
- `scheduledAt` ile ileri tarihli gönderim.
- Zamanlanmış gönderimler (list/retrieve/cancel), hatırlatma kuralları (CRUD) ve Partner API (alt hesap, anahtar, kota aktarma, WhatsApp bağlantı linki).
- Yanıtlarda `whatsapp_template`, `sandbox`, `scheduled_id`, `delivered_via`.
- `client_test.go` (httptest).

### Düzeltmeler / Fixed
- API hata zarfı `{"error": {"code", "message"}}` artık çözülüyor (önceden yalnız düz `{"error": "kod"}` biçimi okunuyordu).

### Değişiklikler / Changed
- Partner kota aktarımı, alt hesap ve anahtar oluşturma istekleri otomatik yeniden denenmez.

## [0.1.0] - 2026-05-12

### Eklenenler / Added
- Açık kaynak proje hijyen dosyaları: LICENSE (MIT), CONTRIBUTING, CODE_OF_CONDUCT, SECURITY
- İki dilli dokümantasyon (Türkçe + English)
- İlk açık kaynak yayını / Initial open-source release
