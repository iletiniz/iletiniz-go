package iletiniz

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Sentinel hatalar `errors.Is` ile kategori kontrolü için kullanılır.
//
//	if errors.Is(err, iletiniz.ErrNotFound) { ... }
var (
	// ErrInvalidRequest istemci tarafı parametre validasyonu hatalarını işaretler.
	ErrInvalidRequest = errors.New("iletiniz: invalid request")
	// ErrAuthentication HTTP 401 — geçersiz/iptal edilmiş API anahtarı.
	ErrAuthentication = errors.New("iletiniz: authentication failed")
	// ErrPermission HTTP 403 — yetki yok.
	ErrPermission = errors.New("iletiniz: permission denied")
	// ErrValidation HTTP 400 / 422 — istek API tarafından doğrulanamadı.
	ErrValidation = errors.New("iletiniz: validation error")
	// ErrRateLimit HTTP 429 — istek hız limitini aştı.
	ErrRateLimit = errors.New("iletiniz: rate limit exceeded")
	// ErrNotFound HTTP 404.
	ErrNotFound = errors.New("iletiniz: not found")
	// ErrServer HTTP 5xx.
	ErrServer = errors.New("iletiniz: server error")
	// ErrTimeout istek timeout süresinde tamamlanamadı.
	ErrTimeout = errors.New("iletiniz: request timed out")
	// ErrConnection ağ kaynaklı bağlantı hatası.
	ErrConnection = errors.New("iletiniz: connection error")
)

// APIError API tarafından dönen HTTP hatalarını temsil eder.
type APIError struct {
	// Status HTTP status kodu.
	Status int
	// Code API tarafından dönen makine-okunur hata kodu (varsa).
	Code string
	// Message kullanıcıya gösterilebilecek mesaj.
	Message string
	// Body sunucudan dönen ham gövde (JSON ya da düz metin).
	Body json.RawMessage
	// RequestID sunucu tarafında üretilen istek kimliği (varsa).
	RequestID string
}

// Error error arayüzünü uygular.
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.RequestID != "" {
		return fmt.Sprintf("iletiniz: HTTP %d: %s (request_id=%s)", e.Status, e.Message, e.RequestID)
	}
	return fmt.Sprintf("iletiniz: HTTP %d: %s", e.Status, e.Message)
}

// Is `errors.Is` ile sentinel kontrol için.
func (e *APIError) Is(target error) bool {
	if e == nil {
		return false
	}
	switch target {
	case ErrAuthentication:
		return e.Status == 401
	case ErrPermission:
		return e.Status == 403
	case ErrNotFound:
		return e.Status == 404
	case ErrValidation:
		return e.Status == 400 || e.Status == 422
	case ErrRateLimit:
		return e.Status == 429
	case ErrServer:
		return e.Status >= 500 && e.Status <= 599
	}
	return false
}

// TimeoutError istek timeout süresinde tamamlanamadığında döner.
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	if e == nil || e.Message == "" {
		return "iletiniz: request timed out"
	}
	return "iletiniz: " + e.Message
}

// Is `errors.Is(err, ErrTimeout)` desteği.
func (*TimeoutError) Is(target error) bool { return target == ErrTimeout }

// ConnectionError ağ kaynaklı hatalarda döner.
type ConnectionError struct {
	Message string
	Cause   error
}

func (e *ConnectionError) Error() string {
	if e == nil {
		return "iletiniz: connection error"
	}
	if e.Cause != nil {
		return fmt.Sprintf("iletiniz: %s: %v", e.Message, e.Cause)
	}
	return "iletiniz: " + e.Message
}

// Unwrap zincir desteği.
func (e *ConnectionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is `errors.Is(err, ErrConnection)` desteği.
func (*ConnectionError) Is(target error) bool { return target == ErrConnection }

// invalidRequestError istemci tarafı validasyon hatalarını döner.
type invalidRequestError struct {
	msg string
}

func (e *invalidRequestError) Error() string {
	return "iletiniz: " + e.msg
}

func (*invalidRequestError) Is(target error) bool { return target == ErrInvalidRequest }

func newInvalidRequest(format string, args ...any) error {
	return &invalidRequestError{msg: fmt.Sprintf(format, args...)}
}

// buildAPIError HTTP status'a uygun mesaj/kodla bir *APIError üretir.
func buildAPIError(status int, raw []byte, requestID string) *APIError {
	apiErr := &APIError{
		Status:    status,
		Body:      append(json.RawMessage(nil), raw...),
		RequestID: requestID,
	}

	if len(raw) > 0 {
		var parsed struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(raw, &parsed); err == nil {
			apiErr.Code = parsed.Error
			apiErr.Message = parsed.Message
		}
	}

	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("HTTP %d", status)
	}

	return apiErr
}
