package iletiniz

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type recorded struct {
	Method string
	Path   string
	Query  string
	Body   map[string]any
}

func testServer(t *testing.T, status int, resp string) (*Client, *[]recorded) {
	t.Helper()
	var mu sync.Mutex
	calls := []recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
		mu.Lock()
		calls = append(calls, recorded{r.Method, r.URL.Path, r.URL.RawQuery, body})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient("iltz_test_abc", WithBaseURL(srv.URL), WithMaxRetries(2))
	if err != nil {
		t.Fatal(err)
	}
	return c, &calls
}

func TestSendWhatsappTemplateScheduled(t *testing.T) {
	c, calls := testServer(t, 201, `{"scheduled_id":7,"status":"scheduled","scheduled_at":"2026-11-05T06:00:00.000Z"}`)
	res, err := c.Messages.Send(context.Background(), SendMessageParams{
		To:       "+905551234567",
		Provider: "whatsapp",
		Sender:   "123456789012345",
		WhatsappTemplate: &WhatsappTemplate{
			Name: "randevu_hatirlatma", Language: "tr",
			Components: []map[string]any{{"type": "body", "parameters": []map[string]any{{"type": "text", "text": "Ayşe"}}}},
		},
		ScheduledAt: "2026-11-05T09:00:00+03:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ScheduledID != 7 || res.Status != "scheduled" {
		t.Fatalf("beklenmeyen yanıt: %+v", res)
	}
	b := (*calls)[0].Body
	if b["scheduledAt"] != "2026-11-05T09:00:00+03:00" || b["sender"] != "123456789012345" {
		t.Fatalf("gövde hatalı: %v", b)
	}
	if b["whatsappTemplate"].(map[string]any)["name"] != "randevu_hatirlatma" {
		t.Fatalf("whatsappTemplate yok: %v", b)
	}
}

func TestBodyAndWhatsappTemplateRejected(t *testing.T) {
	c, _ := testServer(t, 200, `{}`)
	_, err := c.Messages.Send(context.Background(), SendMessageParams{
		To: "+905551234567", Body: "x", WhatsappTemplate: &WhatsappTemplate{Name: "a", Language: "tr"},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("ErrInvalidRequest bekleniyordu: %v", err)
	}
}

func TestScheduledAndRemindersPaths(t *testing.T) {
	c, calls := testServer(t, 200, `{"data":[],"limit":10,"offset":0}`)
	ctx := context.Background()
	if _, err := c.Scheduled.List(ctx, ListParams{Status: "scheduled", Limit: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Scheduled.Cancel(ctx, 5); err != nil {
		t.Fatal(err)
	}
	dow := 1
	if _, err := c.Reminders.Create(ctx, CreateReminderParams{
		Type: "recurring", Name: "Haftalık", Frequency: "weekly", TimeOfDay: "09:30", DayOfWeek: &dow,
		Message: ReminderMessage{To: "+905551234567", Body: "Merhaba"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Reminders.Update(ctx, 3, UpdateReminderParams{ClearEndsAt: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Reminders.Delete(ctx, 3); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /v1/scheduled", "DELETE /v1/scheduled/5", "POST /v1/reminders",
		"PATCH /v1/reminders/3", "DELETE /v1/reminders/3",
	}
	for i, w := range want {
		got := (*calls)[i].Method + " " + (*calls)[i].Path
		if got != w {
			t.Fatalf("çağrı %d: %s != %s", i, got, w)
		}
	}
	if (*calls)[0].Query != "limit=10&status=scheduled" {
		t.Fatalf("query: %s", (*calls)[0].Query)
	}
	if v, ok := (*calls)[3].Body["endsAt"]; !ok || v != nil {
		t.Fatalf("endsAt null bekleniyordu: %v", (*calls)[3].Body)
	}
	if _, err := c.Reminders.Create(ctx, CreateReminderParams{Type: "recurring", Name: "x", Frequency: "weekly", TimeOfDay: "09:30"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("dayOfWeek eksik reddedilmeli: %v", err)
	}
}

func TestQuotaTransferNotRetriedAndNestedError(t *testing.T) {
	c, calls := testServer(t, 503, `{"error":{"code":"x","message":"y"}}`)
	_, err := c.Partner.TransferQuota(context.Background(), 12, 500)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("APIError bekleniyordu: %v", err)
	}
	if apiErr.Code != "x" || apiErr.Message != "y" || !errors.Is(err, ErrServer) {
		t.Fatalf("hata çözümü yanlış: %+v", apiErr)
	}
	if len(*calls) != 1 {
		t.Fatalf("yeniden denenmemeli, çağrı sayısı: %d", len(*calls))
	}
	if (*calls)[0].Path != "/v1/partner/accounts/12/quota-transfers" || (*calls)[0].Body["amount"] != float64(500) {
		t.Fatalf("istek hatalı: %+v", (*calls)[0])
	}
}

func TestPartnerAPIKeyListAndRevoke(t *testing.T) {
	key := `{"id":7,"name":"prod","key_prefix":"iltz_live_Ab","scopes":["messages:send"],"sandbox":false,` +
		`"source":"partner","status":"active","created_at":"2026-10-31T10:00:00.000Z","last_used_at":null,"revoked_at":null}`
	c, calls := testServer(t, 200, `{"data":[`+key+`]}`)
	ctx := context.Background()
	list, err := c.Partner.ListAPIKeys(ctx, 12)
	if err != nil || len(list.Data) != 1 || list.Data[0].KeyPrefix != "iltz_live_Ab" || list.Data[0].LastUsedAt != nil {
		t.Fatalf("liste çözülemedi: %+v %v", list, err)
	}
	if got := (*calls)[0].Method + " " + (*calls)[0].Path; got != "GET /v1/partner/accounts/12/api-keys" {
		t.Fatalf("yol: %s", got)
	}

	c, calls = testServer(t, 200, strings.Replace(key, `"active"`, `"revoked"`, 1))
	revoked, err := c.Partner.RevokeAPIKey(ctx, 12, 7)
	if err != nil || revoked.Status != "revoked" {
		t.Fatalf("iptal yanıtı: %+v %v", revoked, err)
	}
	if got := (*calls)[0].Method + " " + (*calls)[0].Path; got != "DELETE /v1/partner/accounts/12/api-keys/7" {
		t.Fatalf("yol: %s", got)
	}
	if _, err := c.Partner.RevokeAPIKey(ctx, 12, 0); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("keyID 0 reddedilmeli: %v", err)
	}
}

func TestFlatValidationError(t *testing.T) {
	c, _ := testServer(t, 400, `{"error":"validation_failed","message":"Geçersiz istek gövdesi"}`)
	_, err := c.Scheduled.Retrieve(context.Background(), 1)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "validation_failed" || apiErr.Message != "Geçersiz istek gövdesi" {
		t.Fatalf("düz zarf çözülemedi: %v", err)
	}
}

func TestPartnerPaths(t *testing.T) {
	c, calls := testServer(t, 200, `{}`)
	ctx := context.Background()
	_, _ = c.Partner.CreateAccount(ctx, CreateChildAccountParams{Name: "Klinik A", ExternalRef: "k-1"})
	_, _ = c.Partner.CreateAPIKey(ctx, 12, CreateChildAPIKeyParams{Name: "prod", Test: true})
	_, _ = c.Partner.CreateWhatsappSignupLink(ctx, 12)
	_, _ = c.Partner.SuspendAccount(ctx, 12)
	_, _ = c.Partner.Profile(ctx)
	want := []string{
		"POST /v1/partner/accounts", "POST /v1/partner/accounts/12/api-keys",
		"POST /v1/partner/accounts/12/whatsapp-signup", "POST /v1/partner/accounts/12/suspend", "GET /v1/partner",
	}
	for i, w := range want {
		if got := (*calls)[i].Method + " " + (*calls)[i].Path; got != w {
			t.Fatalf("çağrı %d: %s != %s", i, got, w)
		}
	}
	if (*calls)[0].Body["externalRef"] != "k-1" || (*calls)[1].Body["test"] != true {
		t.Fatalf("gövde hatalı: %+v", *calls)
	}
}
