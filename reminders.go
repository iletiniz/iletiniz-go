package iletiniz

import "context"

// ReminderMessage hatırlatmada gönderilecek tekli mesaj (SendMessageParams alt kümesi).
// `Body`, `Template` veya `WhatsappTemplate` alanlarından tam olarak biri verilmelidir.
type ReminderMessage struct {
	To               string            `json:"to"`
	Body             string            `json:"body,omitempty"`
	Template         string            `json:"template,omitempty"`
	Variables        map[string]any    `json:"variables,omitempty"`
	WhatsappTemplate *WhatsappTemplate `json:"whatsappTemplate,omitempty"`
	Sender           string            `json:"sender,omitempty"`
	Provider         string            `json:"provider,omitempty"`
	Iys              *bool             `json:"iys,omitempty"`
}

// CreateReminderParams `POST /v1/reminders` gövdesi.
//
// Type = "one_time": EventAt + OffsetsMinutes zorunlu.
// Type = "recurring": Frequency ("daily"|"weekly"|"monthly") + TimeOfDay ("09:30") zorunlu;
// weekly için DayOfWeek (0=Pazar ... 6=Cumartesi), monthly için DayOfMonth (1-31) zorunlu.
// Tarihler saat dilimli RFC 3339 olmalıdır.
type CreateReminderParams struct {
	Type    string          `json:"type"`
	Name    string          `json:"name"`
	Message ReminderMessage `json:"message"`

	EventAt        string `json:"eventAt,omitempty"`
	OffsetsMinutes []int  `json:"offsetsMinutes,omitempty"`

	Frequency  string `json:"frequency,omitempty"`
	TimeOfDay  string `json:"timeOfDay,omitempty"`
	DayOfWeek  *int   `json:"dayOfWeek,omitempty"`
	DayOfMonth *int   `json:"dayOfMonth,omitempty"`
	// Timezone şu an yalnız "Europe/Istanbul" (varsayılan).
	Timezone string `json:"timezone,omitempty"`
	EndsAt   string `json:"endsAt,omitempty"`
}

// UpdateReminderParams `PATCH /v1/reminders/{id}` gövdesi. Nil/boş alanlar gönderilmez.
type UpdateReminderParams struct {
	Name    string
	Message *ReminderMessage
	// Status "active" veya "paused".
	Status         string
	EventAt        string
	OffsetsMinutes []int
	Frequency      string
	TimeOfDay      string
	DayOfWeek      *int
	DayOfMonth     *int
	EndsAt         string
	// ClearEndsAt true ise bitiş tarihi kaldırılır (`endsAt: null`).
	ClearEndsAt bool
}

func (p UpdateReminderParams) body() map[string]any {
	m := map[string]any{}
	if p.Name != "" {
		m["name"] = p.Name
	}
	if p.Message != nil {
		m["message"] = p.Message
	}
	if p.Status != "" {
		m["status"] = p.Status
	}
	if p.EventAt != "" {
		m["eventAt"] = p.EventAt
	}
	if p.OffsetsMinutes != nil {
		m["offsetsMinutes"] = p.OffsetsMinutes
	}
	if p.Frequency != "" {
		m["frequency"] = p.Frequency
	}
	if p.TimeOfDay != "" {
		m["timeOfDay"] = p.TimeOfDay
	}
	if p.DayOfWeek != nil {
		m["dayOfWeek"] = *p.DayOfWeek
	}
	if p.DayOfMonth != nil {
		m["dayOfMonth"] = *p.DayOfMonth
	}
	if p.ClearEndsAt {
		m["endsAt"] = nil
	} else if p.EndsAt != "" {
		m["endsAt"] = p.EndsAt
	}
	return m
}

// Reminder hatırlatma kuralı.
type Reminder struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	Message        map[string]any `json:"message"`
	EventAt        *string        `json:"event_at"`
	OffsetsMinutes []int          `json:"offsets_minutes"`
	Frequency      *string        `json:"frequency"`
	TimeOfDay      *string        `json:"time_of_day"`
	DayOfWeek      *int           `json:"day_of_week"`
	DayOfMonth     *int           `json:"day_of_month"`
	Timezone       string         `json:"timezone"`
	EndsAt         *string        `json:"ends_at"`
	NextRunAt      *string        `json:"next_run_at"`
	RunCount       int            `json:"run_count"`
	Sandbox        bool           `json:"sandbox,omitempty"`
	CreatedAt      string         `json:"created_at"`
}

// ReminderList `GET /v1/reminders` yanıtı.
type ReminderList struct {
	Data   []Reminder `json:"data"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// RemindersService `/v1/reminders` endpoint ailesini sarar.
type RemindersService struct {
	engine *httpEngine
}

// Create hatırlatma kuralı oluşturur.
func (s *RemindersService) Create(ctx context.Context, params CreateReminderParams, opts ...RequestOption) (*Reminder, error) {
	switch params.Type {
	case "one_time", "recurring":
	default:
		return nil, newInvalidRequest("'type' alanı 'one_time' veya 'recurring' olmalıdır")
	}
	if params.Type == "recurring" {
		if params.Frequency == "weekly" && params.DayOfWeek == nil {
			return nil, newInvalidRequest("haftalık tekrar için 'dayOfWeek' zorunlu")
		}
		if params.Frequency == "monthly" && params.DayOfMonth == nil {
			return nil, newInvalidRequest("aylık tekrar için 'dayOfMonth' zorunlu")
		}
	}
	out := &Reminder{}
	if err := s.engine.request(ctx, "POST", "/v1/reminders", nil, params, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// List hatırlatma kurallarını listeler.
func (s *RemindersService) List(ctx context.Context, params ListParams, opts ...RequestOption) (*ReminderList, error) {
	out := &ReminderList{}
	if err := s.engine.request(ctx, "GET", "/v1/reminders", params.query(), nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Retrieve tek bir kuralı döner.
func (s *RemindersService) Retrieve(ctx context.Context, id int64, opts ...RequestOption) (*Reminder, error) {
	path, err := idPath("/v1/reminders/", id)
	if err != nil {
		return nil, err
	}
	out := &Reminder{}
	if err := s.engine.request(ctx, "GET", path, nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Update kuralı günceller; bekleyen gönderimler yeni ayarlarla yeniden planlanır.
func (s *RemindersService) Update(ctx context.Context, id int64, params UpdateReminderParams, opts ...RequestOption) (*Reminder, error) {
	path, err := idPath("/v1/reminders/", id)
	if err != nil {
		return nil, err
	}
	body := params.body()
	if len(body) == 0 {
		return nil, newInvalidRequest("en az bir alan güncellenmeli")
	}
	out := &Reminder{}
	if err := s.engine.request(ctx, "PATCH", path, nil, body, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete kuralı iptal eder (`canceled`) ve bekleyen gönderimlerini siler.
func (s *RemindersService) Delete(ctx context.Context, id int64, opts ...RequestOption) (*Reminder, error) {
	path, err := idPath("/v1/reminders/", id)
	if err != nil {
		return nil, err
	}
	out := &Reminder{}
	if err := s.engine.request(ctx, "DELETE", path, nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
