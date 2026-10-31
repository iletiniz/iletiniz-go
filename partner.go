package iletiniz

import (
	"context"
	"strconv"
	"strings"
)

// PartnerProfile `GET /v1/partner` yanıtı.
type PartnerProfile struct {
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	Pricing     struct {
		DiscountPct *float64 `json:"discount_pct"`
		Package     *string  `json:"package"`
	} `json:"pricing"`
	CreditBalance int64 `json:"credit_balance"`
	ChildCount    int   `json:"child_count"`
}

// ChildAccount partner alt hesabı.
type ChildAccount struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Status        string  `json:"status"`
	CreditBalance int64   `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
	SuspendedAt   *string `json:"suspended_at"`
}

// ChildAccountList `GET /v1/partner/accounts` yanıtı.
type ChildAccountList struct {
	Data   []ChildAccount `json:"data"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// CreateChildAccountParams `POST /v1/partner/accounts` gövdesi.
type CreateChildAccountParams struct {
	Name string `json:"name"`
	// ExternalRef kendi sistemindeki müşteri kimliği (harf, rakam, _ ve -; en fazla 32).
	ExternalRef string `json:"externalRef,omitempty"`
}

// CreateChildAPIKeyParams `POST /v1/partner/accounts/{id}/api-keys` gövdesi.
type CreateChildAPIKeyParams struct {
	Name string `json:"name"`
	// Scopes varsayılan: messages:send, messages:read.
	Scopes []string `json:"scopes,omitempty"`
	// Test true ise iltz_test_ (sandbox) anahtarı üretilir.
	Test       bool     `json:"test,omitempty"`
	AllowedIPs []string `json:"allowedIps,omitempty"`
}

// ChildAPIKey üretilen alt hesap anahtarı. Key yalnız bu yanıtta döner.
type ChildAPIKey struct {
	ID          int64    `json:"id"`
	WorkspaceID int64    `json:"workspace_id"`
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	KeyPrefix   string   `json:"key_prefix"`
	Scopes      []string `json:"scopes"`
	Sandbox     bool     `json:"sandbox"`
}

// ChildAPIKeyInfo alt hesabın bir API anahtarı (liste/iptal yanıtı). Düz metin anahtar
// veya hash asla dönmez.
type ChildAPIKeyInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// KeyPrefix anahtarın ilk 12 karakteri, örn. "iltz_live_Ab".
	KeyPrefix string   `json:"key_prefix"`
	Scopes    []string `json:"scopes"`
	Sandbox   bool     `json:"sandbox"`
	// Source anahtarın oluşturulduğu kaynak, örn. "partner" | "dashboard".
	Source     string  `json:"source"`
	Status     string  `json:"status"` // "active" | "revoked"
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at"`
	RevokedAt  *string `json:"revoked_at"`
}

// ChildAPIKeyList `GET /v1/partner/accounts/{id}/api-keys` yanıtı.
type ChildAPIKeyList struct {
	Data []ChildAPIKeyInfo `json:"data"`
}

// QuotaTransferResponse kota aktarım sonucu.
type QuotaTransferResponse struct {
	ChildWorkspaceID int64 `json:"child_workspace_id"`
	Amount           int64 `json:"amount"`
	PartnerBalance   int64 `json:"partner_balance"`
	ChildBalance     int64 `json:"child_balance"`
}

// WhatsappSignupLink alt hesabın WhatsApp numarasını bağlaması için tek kullanımlık bağlantı.
type WhatsappSignupLink struct {
	ChildWorkspaceID int64  `json:"child_workspace_id"`
	URL              string `json:"url"`
	ExpiresAt        string `json:"expires_at"`
}

// PartnerService `/v1/partner` endpoint ailesini sarar. Yalnız partner hesabının
// `partner:manage` yetkili CANLI (iltz_live_) anahtarıyla çalışır.
type PartnerService struct {
	engine *httpEngine
}

// Profile partner profilini döner.
func (s *PartnerService) Profile(ctx context.Context, opts ...RequestOption) (*PartnerProfile, error) {
	out := &PartnerProfile{}
	if err := s.engine.request(ctx, "GET", "/v1/partner", nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAccount alt hesap oluşturur. Yeniden denenmez.
func (s *PartnerService) CreateAccount(ctx context.Context, params CreateChildAccountParams, opts ...RequestOption) (*ChildAccount, error) {
	if len([]rune(strings.TrimSpace(params.Name))) < 2 {
		return nil, newInvalidRequest("'name' en az 2 karakter olmalıdır")
	}
	out := &ChildAccount{}
	opts = append(opts, withNoRetry())
	if err := s.engine.request(ctx, "POST", "/v1/partner/accounts", nil, params, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAccounts alt hesapları listeler. Status: "active" veya "suspended".
func (s *PartnerService) ListAccounts(ctx context.Context, params ListParams, opts ...RequestOption) (*ChildAccountList, error) {
	out := &ChildAccountList{}
	if err := s.engine.request(ctx, "GET", "/v1/partner/accounts", params.query(), nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *PartnerService) accountCall(ctx context.Context, method string, id int64, suffix string, body any, out any, opts []RequestOption) error {
	path, err := idPath("/v1/partner/accounts/", id)
	if err != nil {
		return err
	}
	return s.engine.request(ctx, method, path+suffix, nil, body, out, opts...)
}

// RetrieveAccount tek bir alt hesabı döner.
func (s *PartnerService) RetrieveAccount(ctx context.Context, id int64, opts ...RequestOption) (*ChildAccount, error) {
	out := &ChildAccount{}
	if err := s.accountCall(ctx, "GET", id, "", nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SuspendAccount alt hesabı askıya alır.
func (s *PartnerService) SuspendAccount(ctx context.Context, id int64, opts ...RequestOption) (*ChildAccount, error) {
	out := &ChildAccount{}
	if err := s.accountCall(ctx, "POST", id, "/suspend", nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// UnsuspendAccount askıdaki alt hesabı yeniden etkinleştirir.
func (s *PartnerService) UnsuspendAccount(ctx context.Context, id int64, opts ...RequestOption) (*ChildAccount, error) {
	out := &ChildAccount{}
	if err := s.accountCall(ctx, "POST", id, "/unsuspend", nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAPIKey alt hesap için API anahtarı üretir. Yeniden denenmez.
func (s *PartnerService) CreateAPIKey(ctx context.Context, id int64, params CreateChildAPIKeyParams, opts ...RequestOption) (*ChildAPIKey, error) {
	if params.Name == "" {
		return nil, newInvalidRequest("'name' zorunludur")
	}
	out := &ChildAPIKey{}
	if err := s.accountCall(ctx, "POST", id, "/api-keys", params, out, append(opts, withNoRetry())); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAPIKeys alt hesabın API anahtarlarını listeler (en yeni önce). Düz metin anahtar/hash dönmez.
func (s *PartnerService) ListAPIKeys(ctx context.Context, id int64, opts ...RequestOption) (*ChildAPIKeyList, error) {
	out := &ChildAPIKeyList{}
	if err := s.accountCall(ctx, "GET", id, "/api-keys", nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// RevokeAPIKey alt hesabın API anahtarını iptal eder. İdempotenttir: iptal edilmiş anahtar
// için aynı nesne döner.
func (s *PartnerService) RevokeAPIKey(ctx context.Context, id int64, keyID int64, opts ...RequestOption) (*ChildAPIKeyInfo, error) {
	if keyID <= 0 {
		return nil, newInvalidRequest("keyID pozitif tam sayı olmalıdır")
	}
	out := &ChildAPIKeyInfo{}
	if err := s.accountCall(ctx, "DELETE", id, "/api-keys/"+strconv.FormatInt(keyID, 10), nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// TransferQuota partner kotasından alt hesaba mesaj kotası aktarır. Yeniden denenmez.
func (s *PartnerService) TransferQuota(ctx context.Context, id int64, amount int64, opts ...RequestOption) (*QuotaTransferResponse, error) {
	if amount < 1 || amount > 10_000_000 {
		return nil, newInvalidRequest("'amount' 1 ile 10.000.000 arasında olmalıdır")
	}
	out := &QuotaTransferResponse{}
	body := map[string]int64{"amount": amount}
	if err := s.accountCall(ctx, "POST", id, "/quota-transfers", body, out, append(opts, withNoRetry())); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateWhatsappSignupLink alt hesap için tek kullanımlık (24 saat) WhatsApp bağlantı linki üretir.
func (s *PartnerService) CreateWhatsappSignupLink(ctx context.Context, id int64, opts ...RequestOption) (*WhatsappSignupLink, error) {
	out := &WhatsappSignupLink{}
	if err := s.accountCall(ctx, "POST", id, "/whatsapp-signup", nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
