package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alereyleyva/agent-guard/internal/audit"
	"github.com/alereyleyva/agent-guard/internal/config"
	"github.com/alereyleyva/agent-guard/internal/normalize"
	"github.com/alereyleyva/agent-guard/internal/policy"
	"github.com/alereyleyva/agent-guard/internal/provider"
)

type captureLogger struct {
	events []audit.Event
}

func (l *captureLogger) Emit(event audit.Event) {
	l.events = append(l.events, event)
}

func TestFlowProcess_NonStreaming_EmitsToolEvents(t *testing.T) {
	response := map[string]interface{}{
		"id":    "chatcmpl-1",
		"model": "gpt-4o",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "hello",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "call-1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "search_web",
								"arguments": "{}",
							},
						},
					},
				},
				"finish_reason": "stop",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := &captureLogger{}
	pol := policy.NewEngine(config.PolicyConfig{
		Models: config.ModelPolicy{Allow: []string{"gpt-4o"}},
		Tools:  config.ToolPolicy{Allow: []string{"search_web"}},
	})
	flow := NewFlow(provider.NewOpenAI(server.URL, ""), pol, logger)

	req := normalize.NormalizedRequest{
		Model:    "gpt-4o",
		Messages: []normalize.Message{{Role: "user", Content: "hi"}},
	}

	result, err := flow.Process(context.Background(), req)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if len(result.Body) == 0 {
		t.Error("expected response body")
	}

	if len(logger.events) != 5 {
		t.Fatalf("events len = %d, want 5", len(logger.events))
	}
	if logger.events[0].EventType != audit.EventTypeLLMRequest {
		t.Errorf("event[0] type = %q", logger.events[0].EventType)
	}
	if logger.events[1].EventType != audit.EventTypePolicyDecision {
		t.Errorf("event[1] type = %q", logger.events[1].EventType)
	}
	if logger.events[2].EventType != audit.EventTypeLLMResponse {
		t.Errorf("event[2] type = %q", logger.events[2].EventType)
	}
	if logger.events[3].EventType != audit.EventTypeToolProposal || logger.events[3].ToolName != "search_web" {
		t.Errorf("event[3] = %#v", logger.events[3])
	}
	if logger.events[4].EventType != audit.EventTypePolicyDecision || logger.events[4].ToolName != "search_web" {
		t.Errorf("event[4] = %#v", logger.events[4])
	}
}

func TestFlowProcess_ModelDenied(t *testing.T) {
	logger := &captureLogger{}
	pol := policy.NewEngine(config.PolicyConfig{Models: config.ModelPolicy{Allow: []string{"gpt-4o"}}})
	flow := NewFlow(provider.NewOpenAI("https://api.openai.com", ""), pol, logger)

	_, err := flow.Process(context.Background(), normalize.NormalizedRequest{Model: "gpt-3.5-turbo"})
	if err == nil {
		t.Fatal("Process() expected policy error")
	}
	flowErr, ok := err.(*FlowError)
	if !ok {
		t.Fatalf("Process() error type = %T, want *FlowError", err)
	}
	if flowErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", flowErr.StatusCode, http.StatusForbidden)
	}
}

func TestFlowProcess_Streaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.Copy(w, bytes.NewBufferString("data: hello\n\n"))
	}))
	defer server.Close()

	logger := &captureLogger{}
	pol := policy.NewEngine(config.PolicyConfig{Models: config.ModelPolicy{Allow: []string{"gpt-4o"}}})
	flow := NewFlow(provider.NewOpenAI(server.URL, ""), pol, logger)

	req := normalize.NormalizedRequest{Model: "gpt-4o", Stream: true}
	result, err := flow.Process(context.Background(), req)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if result.StreamBody == nil {
		t.Fatal("expected StreamBody")
	}
	_ = result.StreamBody.Close()

	if len(logger.events) < 3 {
		t.Fatalf("events len = %d, want at least 3", len(logger.events))
	}
	if logger.events[2].EventType != audit.EventTypeLLMResponse || !logger.events[2].Stream {
		t.Errorf("event[2] = %#v", logger.events[2])
	}
}
