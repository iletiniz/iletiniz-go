package iletiniz

import (
	"context"
	"net/url"
)

const maxBulkItems = 200

// MessagesService `/v1/messages` endpoint ailesini sarar.
type MessagesService struct {
	engine *httpEngine
}

// Send tek bir SMS mesajı gönderir.
//
// `Body` veya `Template` alanlarından **tam olarak biri** verilmelidir.
// `Variables` yalnızca `Template` ile birlikte kullanılabilir.
func (s *MessagesService) Send(
	ctx context.Context,
	params SendMessageParams,
	opts ...RequestOption,
) (*SendMessageResponse, error) {
	if err := validateSendParams(params); err != nil {
		return nil, err
	}
	out := &SendMessageResponse{}
	if err := s.engine.request(ctx, "POST", "/v1/messages", nil, params, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// SendBulk tek istekte birden fazla mesaj gönderir (en fazla 200 öğe).
//
// - Üst seviye `Template` verildiyse her item'da `Body` olmamalı,
//   yalnızca `Variables` opsiyoneldir.
// - Üst seviye `Template` yoksa her item'da `Body` zorunludur,
//   `Variables` kullanılamaz.
func (s *MessagesService) SendBulk(
	ctx context.Context,
	params SendBulkParams,
	opts ...RequestOption,
) (*SendBulkResponse, error) {
	if err := validateBulkParams(params); err != nil {
		return nil, err
	}
	out := &SendBulkResponse{}
	if err := s.engine.request(ctx, "POST", "/v1/messages/bulk", nil, params, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Retrieve daha önce gönderilmiş bir mesajın güncel durumunu döner.
func (s *MessagesService) Retrieve(
	ctx context.Context,
	jobID string,
	opts ...RequestOption,
) (*MessageStatusResponse, error) {
	if jobID == "" {
		return nil, newInvalidRequest("jobID boş olamaz")
	}
	out := &MessageStatusResponse{}
	path := "/v1/messages/" + url.PathEscape(jobID)
	if err := s.engine.request(ctx, "GET", path, nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// Status `Retrieve` için alias.
func (s *MessagesService) Status(
	ctx context.Context,
	jobID string,
	opts ...RequestOption,
) (*MessageStatusResponse, error) {
	return s.Retrieve(ctx, jobID, opts...)
}

func validateSendParams(p SendMessageParams) error {
	if l := len(p.To); l < 7 || l > 32 {
		return newInvalidRequest("'to' alanı 7-32 karakter arasında olmalıdır")
	}
	hasBody := p.Body != ""
	hasTemplate := p.Template != ""
	if hasBody == hasTemplate {
		return newInvalidRequest("'body' veya 'template' alanlarından tam olarak biri zorunludur")
	}
	if len(p.Variables) > 0 && !hasTemplate {
		return newInvalidRequest("'variables' yalnızca 'template' ile birlikte kullanılabilir")
	}
	if hasBody {
		if l := len(p.Body); l < 1 || l > 1600 {
			return newInvalidRequest("'body' 1-1600 karakter arasında olmalıdır")
		}
	}
	return nil
}

func validateBulkParams(p SendBulkParams) error {
	if len(p.Items) == 0 {
		return newInvalidRequest("'items' en az bir öğe içermelidir")
	}
	if len(p.Items) > maxBulkItems {
		return newInvalidRequest("'items' en fazla %d öğe içerebilir", maxBulkItems)
	}
	usingTemplate := p.Template != ""
	for i, item := range p.Items {
		if l := len(item.To); l < 7 || l > 32 {
			return newInvalidRequest("items[%d].to 7-32 karakter arasında olmalıdır", i)
		}
		if usingTemplate {
			if item.Body != "" {
				return newInvalidRequest("üst seviye 'template' verildi: items[%d].body kullanılamaz", i)
			}
		} else {
			if item.Body == "" {
				return newInvalidRequest("'template' yok: items[%d].body zorunludur", i)
			}
			if len(item.Variables) > 0 {
				return newInvalidRequest("'template' yok: items[%d].variables kullanılamaz", i)
			}
		}
	}
	return nil
}
