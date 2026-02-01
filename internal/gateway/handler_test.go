package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alereyleyva/agent-guard/internal/audit"
	"github.com/alereyleyva/agent-guard/internal/config"
	"github.com/alereyleyva/agent-guard/internal/policy"
	"github.com/alereyleyva/agent-guard/internal/provider"
)

type noopLogger struct{}

func (l noopLogger) Emit(event audit.Event) {}

func TestHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(NewFlow(provider.NewOpenAI("https://api.openai.com", ""), policy.NewEngine(config.PolicyConfig{}), noopLogger{}))

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	handler := NewHandler(NewFlow(provider.NewOpenAI("https://api.openai.com", ""), policy.NewEngine(config.PolicyConfig{}), noopLogger{}))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_PolicyDenied(t *testing.T) {
	pol := policy.NewEngine(config.PolicyConfig{Models: config.ModelPolicy{Allow: []string{"gpt-4o"}}})
	handler := NewHandler(NewFlow(provider.NewOpenAI("https://api.openai.com", ""), pol, noopLogger{}))

	payload := `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var decoded map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if decoded["error"]["type"] != "policy_error" {
		t.Errorf("error.type = %q", decoded["error"]["type"])
	}
}

func TestHandler_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", "true")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-1","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer upstream.Close()

	pol := policy.NewEngine(config.PolicyConfig{Models: config.ModelPolicy{Allow: []string{"gpt-4o"}}})
	handler := NewHandler(NewFlow(provider.NewOpenAI(upstream.URL, ""), pol, noopLogger{}))

	payload := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("X-Upstream") != "true" {
		t.Errorf("X-Upstream header = %q", w.Header().Get("X-Upstream"))
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"id":"chatcmpl-1"`)) {
		t.Errorf("response body = %s", w.Body.String())
	}
}
