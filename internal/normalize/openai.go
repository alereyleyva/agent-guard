package normalize

import (
	"encoding/json"
	"errors"
	"io"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
	Tools    []Tool    `json:"tools,omitempty"`
}

func DecodeOpenAIRequest(r io.Reader) (NormalizedRequest, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var req OpenAIRequest
	if err := dec.Decode(&req); err != nil {
		return NormalizedRequest{}, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return NormalizedRequest{}, errors.New("invalid trailing data")
	}
	return NormalizedRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   req.Stream,
		Tools:    req.Tools,
	}, nil
}
