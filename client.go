// Package iletiniz, Iletiniz API için resmi Go SDK'sini sağlar.
//
// Hızlı başlangıç:
//
//	client, err := iletiniz.NewClient(os.Getenv("ILETINIZ_API_KEY"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	res, err := client.Messages.Send(ctx, iletiniz.SendMessageParams{
//	    To:   "+905551234567",
//	    Body: "Merhaba!",
//	})
package iletiniz

import (
	"net/http"
	"os"
	"regexp"
	"time"
)

// Version SDK sürümü.
const Version = "0.1.0"

const (
	defaultBaseURL    = "https://api.iletiniz.com"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
)

var apiKeyRE = regexp.MustCompile(`^iltz_(?:live|test)_[A-Za-z0-9_-]+$`)

// Client Iletiniz API'sine erişim sağlayan ana istemci.
type Client struct {
	// Messages mesaj gönderim ve durum sorgulama servisi.
	Messages *MessagesService
	// Health sağlık kontrolü servisi.
	Health *HealthService

	engine *httpEngine
}

// NewClient verilen API anahtarı ile bir `*Client` oluşturur.
//
// `apiKey` boş geçilirse `ILETINIZ_API_KEY` ortam değişkeni okunur.
// Anahtar `iltz_live_…` veya `iltz_test_…` formatında olmalıdır.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ILETINIZ_API_KEY")
	}
	if apiKey == "" {
		return nil, newInvalidRequest(
			"API anahtarı gerekli. NewClient(apiKey) veya ILETINIZ_API_KEY ortam değişkeni kullanın.",
		)
	}
	if !apiKeyRE.MatchString(apiKey) {
		return nil, newInvalidRequest(
			"geçersiz API anahtar formatı. Beklenen: 'iltz_live_…' veya 'iltz_test_…'.",
		)
	}

	cfg := clientConfig{
		baseURL:    defaultBaseURL,
		timeout:    defaultTimeout,
		maxRetries: defaultMaxRetries,
	}
	if envBase := os.Getenv("ILETINIZ_BASE_URL"); envBase != "" {
		cfg.baseURL = envBase
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{}
	}

	headers := map[string]string{
		"User-Agent": "iletiniz-go/" + Version,
	}
	for k, v := range cfg.defaultHeaders {
		headers[k] = v
	}

	engine := &httpEngine{
		baseURL:        cfg.baseURL,
		apiKey:         apiKey,
		timeout:        cfg.timeout,
		maxRetries:     cfg.maxRetries,
		defaultHeaders: headers,
		doer:           cfg.httpClient,
	}

	c := &Client{engine: engine}
	c.Messages = &MessagesService{engine: engine}
	c.Health = &HealthService{engine: engine}
	return c, nil
}
