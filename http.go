package iletiniz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpEngine struct {
	baseURL        string
	apiKey         string
	timeout        time.Duration
	maxRetries     int
	defaultHeaders map[string]string
	doer           httpDoer
}

type attemptOutcome struct {
	// resp non-nil ise istek tamamlandı (success ya da API hatası).
	status    int
	body      []byte
	requestID string
	retryAfter string

	// terminalErr non-nil ise sonlandırıcı bir hata var (TimeoutError/ConnectionError).
	terminalErr error
	// transientErr non-nil ise retry adayı bir ağ hatası var.
	transientErr error
}

func (e *httpEngine) request(
	ctx context.Context,
	method string,
	path string,
	query map[string]string,
	body any,
	out any,
	opts ...RequestOption,
) error {
	cfg := requestConfig{timeout: e.timeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	urlStr, err := e.buildURL(path, query)
	if err != nil {
		return err
	}

	var rawBody []byte
	if body != nil {
		rawBody, err = json.Marshal(body)
		if err != nil {
			return &ConnectionError{Message: "istek gövdesi JSON olarak kodlanamadı", Cause: err}
		}
	}

	attempt := 0
	for {
		outcome := e.doAttempt(ctx, method, urlStr, rawBody, cfg)

		if outcome.terminalErr != nil {
			return outcome.terminalErr
		}

		if outcome.transientErr != nil {
			if e.shouldRetry(0, attempt) {
				attempt++
				if !sleep(ctx, e.backoff(attempt, "")) {
					return &ConnectionError{Message: "context iptal edildi", Cause: ctx.Err()}
				}
				continue
			}
			return &ConnectionError{Message: "bağlantı hatası", Cause: outcome.transientErr}
		}

		status := outcome.status
		if status >= 200 && status < 300 {
			if status == http.StatusNoContent || len(outcome.body) == 0 || out == nil {
				return nil
			}
			if jsonErr := json.Unmarshal(outcome.body, out); jsonErr != nil {
				return &ConnectionError{Message: "sunucudan geçersiz JSON döndü", Cause: jsonErr}
			}
			return nil
		}

		if e.shouldRetry(status, attempt) {
			attempt++
			if !sleep(ctx, e.backoff(attempt, outcome.retryAfter)) {
				return &ConnectionError{Message: "context iptal edildi", Cause: ctx.Err()}
			}
			continue
		}

		return buildAPIError(status, outcome.body, outcome.requestID)
	}
}

func (e *httpEngine) doAttempt(
	parent context.Context,
	method, urlStr string,
	rawBody []byte,
	cfg requestConfig,
) attemptOutcome {
	reqCtx := parent
	var cancel context.CancelFunc
	if cfg.timeout > 0 {
		reqCtx, cancel = context.WithTimeout(parent, cfg.timeout)
		defer cancel()
	}

	var bodyReader io.Reader
	if rawBody != nil {
		bodyReader = bytes.NewReader(rawBody)
	}

	req, reqErr := http.NewRequestWithContext(reqCtx, method, urlStr, bodyReader)
	if reqErr != nil {
		return attemptOutcome{
			terminalErr: &ConnectionError{Message: "istek oluşturulamadı", Cause: reqErr},
		}
	}

	for k, v := range e.defaultHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Accept", "application/json")
	if rawBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	resp, doErr := e.doer.Do(req)
	if doErr != nil {
		if isContextTimeout(reqCtx, doErr) && parent.Err() == nil {
			return attemptOutcome{
				terminalErr: &TimeoutError{Message: "istek timeout süresinde tamamlanamadı"},
			}
		}
		if parent.Err() != nil {
			return attemptOutcome{
				terminalErr: &ConnectionError{Message: "context iptal edildi", Cause: parent.Err()},
			}
		}
		return attemptOutcome{transientErr: doErr}
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		if parent.Err() != nil {
			return attemptOutcome{
				terminalErr: &ConnectionError{Message: "context iptal edildi", Cause: parent.Err()},
			}
		}
		return attemptOutcome{transientErr: readErr}
	}

	return attemptOutcome{
		status:     resp.StatusCode,
		body:       raw,
		requestID:  resp.Header.Get("X-Request-Id"),
		retryAfter: resp.Header.Get("Retry-After"),
	}
}

func (e *httpEngine) buildURL(path string, query map[string]string) (string, error) {
	base := strings.TrimRight(e.baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	full := base + path
	if len(query) == 0 {
		return full, nil
	}
	u, err := url.Parse(full)
	if err != nil {
		return "", &ConnectionError{Message: "geçersiz URL", Cause: err}
	}
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (e *httpEngine) shouldRetry(status, attempt int) bool {
	if attempt >= e.maxRetries {
		return false
	}
	if status == 0 {
		return true
	}
	if status == http.StatusRequestTimeout || status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && status <= 599
}

func (e *httpEngine) backoff(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if sec, err := strconv.ParseFloat(retryAfter, 64); err == nil && sec > 0 {
			ms := sec * 1000
			if ms > 30000 {
				ms = 30000
			}
			return time.Duration(ms) * time.Millisecond
		}
	}
	base := (1 << attempt) * 250
	if base > 4000 {
		base = 4000
	}
	jitter := rand.Intn(101) //nolint:gosec // backoff jitter, kriptografik değil
	return time.Duration(base+jitter) * time.Millisecond
}

// sleep ctx iptal edilirse false döner.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func isContextTimeout(ctx context.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return true
	}
	return false
}
