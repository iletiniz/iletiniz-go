package iletiniz

// MessageStatus mesajın olası nihai durumları.
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusQueued    MessageStatus = "queued"
	StatusFailed    MessageStatus = "failed"
	StatusDelivered MessageStatus = "delivered"
	StatusExpired   MessageStatus = "expired"
	StatusRejected  MessageStatus = "rejected"
	StatusUnknown   MessageStatus = "unknown"
)

// SendStatus tek mesaj gönderim sonucu.
type SendStatus string

const (
	SendStatusSent   SendStatus = "sent"
	SendStatusQueued SendStatus = "queued"
	SendStatusFailed SendStatus = "failed"
)

// ErrorInfo API yanıtlarında dönen hata gövdesi.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendMessageParams tek mesaj gönderim girdisi.
//
// `Body` veya `Template` alanlarından **tam olarak biri** verilmelidir.
// `Variables` yalnızca `Template` ile birlikte kullanılabilir.
type SendMessageParams struct {
	// To alıcı telefon numarası (E.164 önerilir).
	To string `json:"to"`
	// Body düz metin gövde. `Template` ile birlikte kullanılamaz.
	Body string `json:"body,omitempty"`
	// Template template anahtarı. `Body` ile birlikte kullanılamaz.
	Template string `json:"template,omitempty"`
	// Variables yalnızca `Template` ile birlikte kullanılabilir.
	Variables map[string]any `json:"variables,omitempty"`
	// Sender gönderici adı / başlık.
	Sender string `json:"sender,omitempty"`
	// Provider belirli bir provider seçmek için kod.
	Provider string `json:"provider,omitempty"`
	// Iys İYS (İleti Yönetim Sistemi) izni. true → ticari (sağlayıcının
	// İYS filtresi devreye girer). false → bilgilendirme (İYS sorgusu yok).
	// Pointer olduğu için nil bırakılırsa istek body'sine eklenmez.
	// Yalnızca SMS sağlayıcılarında işlenir; WhatsApp/Telegram için yok sayılır.
	Iys *bool `json:"iys,omitempty"`
}

// SendMessageResponse `POST /v1/messages` yanıtı.
type SendMessageResponse struct {
	JobID       string     `json:"job_id"`
	Status      SendStatus `json:"status"`
	To          string     `json:"to"`
	Provider    string     `json:"provider"`
	Template    string     `json:"template,omitempty"`
	TemplateKey string     `json:"template_key,omitempty"`
	Error       *ErrorInfo `json:"error,omitempty"`
	CreatedAt   string     `json:"created_at"`
}

// MessageStatusResponse `GET /v1/messages/{job_id}` yanıtı.
type MessageStatusResponse struct {
	JobID       string        `json:"job_id"`
	Status      MessageStatus `json:"status"`
	To          string        `json:"to"`
	Provider    string        `json:"provider"`
	Error       *ErrorInfo    `json:"error,omitempty"`
	CreatedAt   string        `json:"created_at"`
	SentAt      *string       `json:"sent_at"`
	DeliveredAt *string       `json:"delivered_at"`
}

// BulkItemInput toplu gönderimde tek bir mesaj öğesi.
type BulkItemInput struct {
	To        string         `json:"to"`
	Body      string         `json:"body,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

// SendBulkParams toplu gönderim girdisi.
//
// `Template` verildiyse her item'da `Body` olmamalı (yalnızca `Variables` opsiyonel).
// `Template` yoksa her item'da `Body` zorunludur, `Variables` kullanılamaz.
type SendBulkParams struct {
	Provider string          `json:"provider,omitempty"`
	Sender   string          `json:"sender,omitempty"`
	Template string          `json:"template,omitempty"`
	// Iys bkz. SendMessageParams.Iys. Tüm batch için tek değer.
	Iys   *bool           `json:"iys,omitempty"`
	Items []BulkItemInput `json:"items"`
}

// SendBulkItemResult toplu gönderimde tek bir öğe için sonuç.
type SendBulkItemResult struct {
	To     string     `json:"to"`
	Status string     `json:"status"`
	JobID  string     `json:"job_id,omitempty"`
	Error  *ErrorInfo `json:"error,omitempty"`
}

// SendBulkResponse `POST /v1/messages/bulk` yanıtı.
type SendBulkResponse struct {
	Total       int                  `json:"total"`
	Sent        int                  `json:"sent"`
	Failed      int                  `json:"failed"`
	Provider    string               `json:"provider"`
	Template    string               `json:"template,omitempty"`
	TemplateKey string               `json:"template_key,omitempty"`
	CreatedAt   string               `json:"created_at"`
	Results     []SendBulkItemResult `json:"results"`
}

// HealthResponse `GET /v1/health` yanıtı.
type HealthResponse struct {
	OK bool   `json:"ok"`
	DB string `json:"db"`
}
