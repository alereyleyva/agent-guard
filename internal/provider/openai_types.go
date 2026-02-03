package provider

import (
	"encoding/json"
	"fmt"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

type openAIRequest struct {
	Model    string              `json:"model"`
	Messages []normalize.Message `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
	Tools    []normalize.Tool    `json:"tools,omitempty"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string               `json:"role"`
			Content   string               `json:"content"`
			ToolCalls []normalize.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func parseOpenAIResponse(body []byte) (normalize.NormalizedResponse, error) {
	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return normalize.NormalizedResponse{RawBody: body}, fmt.Errorf("parsing response: %w", err)
	}

	normalized := normalize.NormalizedResponse{
		RawBody: body,
		ID:      openAIResp.ID,
		Model:   openAIResp.Model,
	}

	for _, choice := range openAIResp.Choices {
		if normalized.Content == "" {
			normalized.Content = choice.Message.Content
		}
		if len(choice.Message.ToolCalls) > 0 {
			normalized.ToolCalls = append(normalized.ToolCalls, choice.Message.ToolCalls...)
		}
	}

	return normalized, nil
}
