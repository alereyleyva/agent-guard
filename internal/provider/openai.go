package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alereyleyva/agent-guard/internal/normalize"
)

type OpenAIProvider struct {
	baseURL string
	apiKey  string
}

func NewOpenAI(baseURL, apiKey string) *OpenAIProvider {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &OpenAIProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

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

func (p *OpenAIProvider) BuildUpstreamRequest(req normalize.NormalizedRequest) (*http.Request, error) {
	openAIReq := openAIRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   req.Stream,
		Tools:    req.Tools,
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	return httpReq, nil
}

func (p *OpenAIProvider) ParseUpstreamResponse(resp *http.Response) (normalize.NormalizedResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return normalize.NormalizedResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	normalized := normalize.NormalizedResponse{
		RawBody: body,
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return normalize.NormalizedResponse{RawBody: body}, fmt.Errorf("parsing response: %w", err)
	}

	normalized.ID = openAIResp.ID
	normalized.Model = openAIResp.Model

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
