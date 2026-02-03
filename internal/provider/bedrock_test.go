package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

func TestBedrockProvider_BuildUpstreamRequest(t *testing.T) {
	p, err := NewBedrock("us-east-1", "https://bedrock-runtime.us-east-1.amazonaws.com", "test", "secret", "")
	if err != nil {
		t.Fatalf("NewBedrock() error = %v", err)
	}

	req := normalize.NormalizedRequest{
		Model: "anthropic.claude-3-5-sonnet-20240620-v1:0",
		Messages: []normalize.Message{
			{Role: "system", Content: "system"},
			{Role: "user", Content: "hello"},
			{Role: "assistant", ToolCalls: []normalize.ToolCall{
				{ID: "call-1", Type: "function", Function: normalize.FunctionCall{Name: "search_web", Arguments: `{"q":"hi"}`}},
			}},
			{Role: "tool", ToolCallID: "call-1", Content: "result"},
		},
		Tools: []normalize.Tool{{Type: "function", Function: normalize.ToolFunction{Name: "search_web"}}},
	}

	httpReq, err := p.BuildUpstreamRequest(req)
	if err != nil {
		t.Fatalf("BuildUpstreamRequest() error = %v", err)
	}

	if httpReq.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", httpReq.Method, http.MethodPost)
	}
	if !strings.Contains(httpReq.URL.String(), "/model/anthropic.claude-3-5-sonnet-20240620-v1:0/converse") {
		t.Errorf("URL = %q, expected converse path", httpReq.URL.String())
	}
	if httpReq.Header.Get("Authorization") == "" {
		t.Error("Authorization header should be set")
	}
	if httpReq.Header.Get("X-Amz-Date") == "" {
		t.Error("X-Amz-Date header should be set")
	}

	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("reading body error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body error = %v", err)
	}
	if decoded["messages"] == nil {
		t.Errorf("messages should be present")
	}
	if decoded["toolConfig"] == nil {
		t.Errorf("toolConfig should be present")
	}
}

func TestBedrockProvider_BuildUpstreamRequest_Stream(t *testing.T) {
	p, err := NewBedrock("us-east-1", "https://bedrock-runtime.us-east-1.amazonaws.com", "test", "secret", "")
	if err != nil {
		t.Fatalf("NewBedrock() error = %v", err)
	}

	req := normalize.NormalizedRequest{Model: "anthropic.claude-3-5-sonnet-20240620-v1:0", Stream: true}
	httpReq, err := p.BuildUpstreamRequest(req)
	if err != nil {
		t.Fatalf("BuildUpstreamRequest() error = %v", err)
	}
	if !strings.Contains(httpReq.URL.String(), "/converse-stream") {
		t.Errorf("URL = %q, expected converse-stream path", httpReq.URL.String())
	}
}

func TestBedrockProvider_ParseUpstreamResponse(t *testing.T) {
	payload := `{"output":{"message":{"role":"assistant","content":[{"text":"hello"},{"toolUse":{"toolUseId":"call-1","name":"search_web","input":{"q":"hi"}}}]}}}`
	resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString(payload))}

	p, err := NewBedrock("us-east-1", "https://bedrock-runtime.us-east-1.amazonaws.com", "test", "secret", "")
	if err != nil {
		t.Fatalf("NewBedrock() error = %v", err)
	}

	normalized, err := p.ParseUpstreamResponse(resp)
	if err != nil {
		t.Fatalf("ParseUpstreamResponse() error = %v", err)
	}

	if normalized.Content != "hello" {
		t.Errorf("Content = %q, want %q", normalized.Content, "hello")
	}
	if len(normalized.ToolCalls) != 1 || normalized.ToolCalls[0].Function.Name != "search_web" {
		t.Errorf("ToolCalls = %#v, want one tool call named search_web", normalized.ToolCalls)
	}
}
