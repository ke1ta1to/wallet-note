package testutils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MakeReq fills x-amzn-request-context to mirror what API Gateway HTTP API +
// LWA produce. nil claims = unauthenticated.
func MakeReq(t *testing.T, method, path string, body io.Reader, claims map[string]any) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if claims == nil {
		return req
	}
	ctx := map[string]any{
		"authorizer": map[string]any{
			"jwt": map[string]any{"claims": claims},
		},
	}
	b, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	req.Header.Set("x-amzn-request-context", string(b))
	return req
}
