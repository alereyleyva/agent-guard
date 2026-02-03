package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

func TestOpenRouterProvider_BuildUpstreamRequest(t *testing.T) {
	p := NewOpenRouter("", "test-key", "https://example.com", "AgentGuard")
	req := normalize.NormalizedRequest{
		Model:    "openai/gpt-4o",
		Messages: []normalize.Message{{Role: "user", Content: "hello"}},
		Stream:   true,
		Tools:    []normalize.Tool{{Type: "function", Function: normalize.ToolFunction{Name: "search_web"}}},
	}

	httpReq, err := p.BuildUpstreamRequest(req)
	if err != nil {
		t.Fatalf("BuildUpstreamRequest() error = %v", err)
	}

	if httpReq.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", httpReq.Method, http.MethodPost)
	}
	if httpReq.URL.String() != "https://openrouter.ai/api/v1/chat/completions" {
		t.Errorf("URL = %q, want %q", httpReq.URL.String(), "https://openrouter.ai/api/v1/chat/completions")
	}
	if httpReq.Header.Get("Authorization") != "Bearer test-key" {
		t.Errorf("Authorization header = %q", httpReq.Header.Get("Authorization"))
	}
	if httpReq.Header.Get("HTTP-Referer") != "https://example.com" {
		t.Errorf("HTTP-Referer header = %q", httpReq.Header.Get("HTTP-Referer"))
	}
	if httpReq.Header.Get("X-Title") != "AgentGuard" {
		t.Errorf("X-Title header = %q", httpReq.Header.Get("X-Title"))
	}
	if httpReq.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type header = %q", httpReq.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("reading body error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body error = %v", err)
	}
	if decoded["model"] != "openai/gpt-4o" {
		t.Errorf("model = %v, want openai/gpt-4o", decoded["model"])
	}
	if decoded["stream"] != true {
		t.Errorf("stream = %v, want true", decoded["stream"])
	}
}

func TestOpenRouterProvider_ParseUpstreamResponse(t *testing.T) {
	payload := `{"id":"chatcmpl-1","model":"openai/gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"hello","tool_calls":[{"id":"call-1","type":"function","function":{"name":"search_web","arguments":"{}"}}]},"finish_reason":"stop"}]}`
	resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString(payload))}

	normalized, err := NewOpenRouter("", "", "", "").ParseUpstreamResponse(resp)
	if err != nil {
		t.Fatalf("ParseUpstreamResponse() error = %v", err)
	}

	if normalized.ID != "chatcmpl-1" {
		t.Errorf("ID = %q, want %q", normalized.ID, "chatcmpl-1")
	}
	if normalized.Model != "openai/gpt-4o" {
		t.Errorf("Model = %q, want %q", normalized.Model, "openai/gpt-4o")
	}
	if normalized.Content != "hello" {
		t.Errorf("Content = %q, want %q", normalized.Content, "hello")
	}
	if len(normalized.ToolCalls) != 1 || normalized.ToolCalls[0].Function.Name != "search_web" {
		t.Errorf("ToolCalls = %#v, want one tool call named search_web", normalized.ToolCalls)
	}
}
