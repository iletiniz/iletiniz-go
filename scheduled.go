package iletiniz

import (
	"context"
	"strconv"
)

// ScheduledStatus zamanlanmış gönderim durumu.
type ScheduledStatus string

const (
	ScheduledStatusScheduled  ScheduledStatus = "scheduled"
	ScheduledStatusProcessing ScheduledStatus = "processing"
	ScheduledStatusSent       ScheduledStatus = "sent"
	ScheduledStatusFailed     ScheduledStatus = "failed"
	ScheduledStatusCanceled   ScheduledStatus = "canceled"
)

// ScheduledMessage `ScheduledAt` ile oluşturulan ileri tarihli gönderim kaydı.
type ScheduledMessage struct {
	ID          int64           `json:"id"`
	Kind        string          `json:"kind"`
	Status      ScheduledStatus `json:"status"`
	ScheduledAt string          `json:"scheduled_at"`
	ReminderID  *int64          `json:"reminder_id"`
	Recipients  int             `json:"recipients"`
	Sandbox     bool            `json:"sandbox,omitempty"`
	Result      map[string]any  `json:"result"`
	ErrorCode   *string         `json:"error_code"`
	CreatedAt   string          `json:"created_at"`
}

// ListParams ortak sayfalama parametreleri. Sıfır değerler gönderilmez.
type ListParams struct {
	// Status isteğe bağlı durum filtresi.
	Status string
	// Limit 1-100 (varsayılan 50).
	Limit int
	// Offset varsayılan 0.
	Offset int
}

func (p ListParams) query() map[string]string {
	q := map[string]string{}
	if p.Status != "" {
		q["status"] = p.Status
	}
	if p.Limit > 0 {
		q["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Offset > 0 {
		q["offset"] = strconv.Itoa(p.Offset)
	}
	return q
}

// ScheduledList `GET /v1/scheduled` yanıtı.
type ScheduledList struct {
	Data   []ScheduledMessage `json:"data"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// ScheduledService `/v1/scheduled` endpoint ailesini sarar.
type ScheduledService struct {
	engine *httpEngine
}

func idPath(prefix string, id int64) (string, error) {
	if id <= 0 {
		return "", newInvalidRequest("id pozitif tam sayı olmalıdır")
	}
	return prefix + strconv.FormatInt(id, 10), nil
}

// List zamanlanmış gönderimleri listeler.
func (s *ScheduledService) List(ctx context.Context, params ListParams, opts ...RequestOption) (*ScheduledList, error) {
	out := &ScheduledList{}
	if err := s.engine.request(ctx, "GET", "/v1/scheduled", params.query(), nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Retrieve tek bir zamanlanmış gönderimi döner.
func (s *ScheduledService) Retrieve(ctx context.Context, id int64, opts ...RequestOption) (*ScheduledMessage, error) {
	path, err := idPath("/v1/scheduled/", id)
	if err != nil {
		return nil, err
	}
	out := &ScheduledMessage{}
	if err := s.engine.request(ctx, "GET", path, nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Cancel bekleyen (`scheduled`) gönderimi iptal eder; aksi halde 409 `not_cancelable`.
func (s *ScheduledService) Cancel(ctx context.Context, id int64, opts ...RequestOption) (*ScheduledMessage, error) {
	path, err := idPath("/v1/scheduled/", id)
	if err != nil {
		return nil, err
	}
	out := &ScheduledMessage{}
	if err := s.engine.request(ctx, "DELETE", path, nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
