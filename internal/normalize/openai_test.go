package normalize

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeOpenAIRequest_Valid(t *testing.T) {
	payload := `{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}],"stream":true,"tools":[{"type":"function","function":{"name":"search_web"}}]}`

	req, err := DecodeOpenAIRequest(bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("DecodeOpenAIRequest() error = %v", err)
	}

	if req.Model != "gpt-4o" {
		t.Errorf("Model = %q, want %q", req.Model, "gpt-4o")
	}
	if !req.Stream {
		t.Errorf("Stream = %v, want true", req.Stream)
	}
	if len(req.Messages) != 1 || req.Messages[0].Content != "hello" {
		t.Errorf("Messages = %#v, want one message with content", req.Messages)
	}
	if len(req.Tools) != 1 || req.Tools[0].Function.Name != "search_web" {
		t.Errorf("Tools = %#v, want one tool named search_web", req.Tools)
	}
}

func TestDecodeOpenAIRequest_UnknownField(t *testing.T) {
	payload := `{"model":"gpt-4o","messages":[],"unknown":true}`

	_, err := DecodeOpenAIRequest(bytes.NewBufferString(payload))
	if err == nil {
		t.Fatal("DecodeOpenAIRequest() expected error for unknown field")
	}
}

func TestDecodeOpenAIRequest_TrailingData(t *testing.T) {
	payload := `{"model":"gpt-4o","messages":[]} trailing`

	_, err := DecodeOpenAIRequest(bytes.NewBufferString(payload))
	if err == nil {
		t.Fatal("DecodeOpenAIRequest() expected error for trailing data")
	}
	if !strings.Contains(err.Error(), "trailing") {
		t.Errorf("DecodeOpenAIRequest() error = %v, want trailing data error", err)
	}
}
