package iletiniz

import (
	"net/http"
	"time"
)

// ClientOption `NewClient`'a aktarılan yapılandırma opsiyonları.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL        string
	timeout        time.Duration
	maxRetries     int
	defaultHeaders map[string]string
	httpClient     *http.Client
}

// WithBaseURL API base URL'sini değiştirir. Varsayılan: https://api.iletiniz.com
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		if url != "" {
			c.baseURL = url
		}
	}
}

// WithTimeout istekler için varsayılan timeout süresini ayarlar.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithMaxRetries 408/429/5xx ve ağ hatalarında yeniden deneme sayısını ayarlar.
func WithMaxRetries(n int) ClientOption {
	return func(c *clientConfig) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

// WithDefaultHeaders her isteğe eklenecek varsayılan başlıkları ayarlar.
func WithDefaultHeaders(headers map[string]string) ClientOption {
	return func(c *clientConfig) {
		if c.defaultHeaders == nil {
			c.defaultHeaders = make(map[string]string, len(headers))
		}
		for k, v := range headers {
			c.defaultHeaders[k] = v
		}
	}
}

// WithHTTPClient SDK'nın kullanacağı `*http.Client`'i değiştirir (test, proxy, vb.).
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *clientConfig) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// RequestOption tekil istek bazlı opsiyon.
type RequestOption func(*requestConfig)

type requestConfig struct {
	timeout time.Duration
	headers map[string]string
}

// WithRequestTimeout bu istek için client default'unu ezen timeout.
func WithRequestTimeout(d time.Duration) RequestOption {
	return func(r *requestConfig) {
		if d > 0 {
			r.timeout = d
		}
	}
}

// WithRequestHeaders bu isteğe eklenecek ekstra HTTP başlıkları.
func WithRequestHeaders(headers map[string]string) RequestOption {
	return func(r *requestConfig) {
		if r.headers == nil {
			r.headers = make(map[string]string, len(headers))
		}
		for k, v := range headers {
			r.headers[k] = v
		}
	}
}
