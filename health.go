package iletiniz

import "context"

// HealthService `/v1/health` endpoint'ini sarar.
type HealthService struct {
	engine *httpEngine
}

// Check API ve veritabanının erişilebilirliğini kontrol eder.
func (s *HealthService) Check(
	ctx context.Context,
	opts ...RequestOption,
) (*HealthResponse, error) {
	out := &HealthResponse{}
	if err := s.engine.request(ctx, "GET", "/v1/health", nil, nil, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
