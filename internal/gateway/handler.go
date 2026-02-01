package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

type Handler struct {
	flow *Flow
}

func NewHandler(flow *Flow) *Handler {
	return &Handler{flow: flow}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	req, err := normalize.DecodeOpenAIRequest(r.Body)
	if err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

	result, err := h.flow.Process(r.Context(), req)
	if err != nil {
		var flowErr *FlowError
		if errors.As(err, &flowErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(flowErr.StatusCode)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": flowErr.Message,
					"type":    flowErr.Type,
					"code":    flowErr.Code,
				},
			})
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	for key, values := range result.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(result.StatusCode)
	if result.StreamBody != nil {
		defer result.StreamBody.Close()
		_, _ = io.Copy(w, result.StreamBody)
		return
	}
	_, _ = w.Write(result.Body)
}

type FlowError struct {
	StatusCode int
	Message    string
	Type       string
	Code       string
}

func (e *FlowError) Error() string {
	return e.Message
}

func NewPolicyDeniedError(reason string) *FlowError {
	return &FlowError{
		StatusCode: http.StatusForbidden,
		Message:    reason,
		Type:       "policy_error",
		Code:       "policy_denied",
	}
}
