package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONLogger_Emit(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf)

	logger.Emit(NewEvent("trace-1", EventTypeLLMRequest).WithProvider("openai"))

	if buf.Len() == 0 {
		t.Fatal("expected JSON output")
	}

	var decoded Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json unmarshal error = %v", err)
	}
	if decoded.TraceID != "trace-1" || decoded.Provider != "openai" {
		t.Errorf("decoded event = %#v", decoded)
	}
}

func TestHashContent(t *testing.T) {
	hash := HashContent([]byte("hello"))
	if !strings.HasPrefix(hash, "sha256:") {
		t.Errorf("HashContent() = %q, want sha256 prefix", hash)
	}

	hash2 := HashContent([]byte("hello"))
	if hash != hash2 {
		t.Errorf("HashContent() = %q, want deterministic", hash)
	}
}
