package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alereyleyva/agent-guard/internal/config"
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

func init() {
	RegisterFactory("openai_compatible", openAIFactory)
	RegisterFactory("openai", openAIFactory)
}

func openAIFactory(cfg config.ProviderConfig) (Provider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("provider base_url is required")
	}
	return NewOpenAI(cfg.BaseURL, cfg.APIKey), nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
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

	return parseOpenAIResponse(body)
}
