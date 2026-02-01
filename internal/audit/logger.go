package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

type Logger interface {
	Emit(event Event)
}

type JSONLogger struct {
	writer  io.Writer
	encoder *json.Encoder
}

func NewJSONLogger(w io.Writer) *JSONLogger {
	return &JSONLogger{
		writer:  w,
		encoder: json.NewEncoder(w),
	}
}

func NewStdoutLogger() *JSONLogger {
	return NewJSONLogger(os.Stdout)
}

func (l *JSONLogger) Emit(event Event) {
	_ = l.encoder.Encode(event)
}

func HashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}
