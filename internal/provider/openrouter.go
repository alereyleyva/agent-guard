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

type OpenRouterProvider struct {
	baseURL string
	apiKey  string
	referer string
	title   string
}

func NewOpenRouter(baseURL, apiKey, referer, title string) *OpenRouterProvider {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &OpenRouterProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		referer: referer,
		title:   title,
	}
}

func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

func (p *OpenRouterProvider) BuildUpstreamRequest(req normalize.NormalizedRequest) (*http.Request, error) {
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

	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}
	if p.referer != "" {
		httpReq.Header.Set("HTTP-Referer", p.referer)
	}
	if p.title != "" {
		httpReq.Header.Set("X-Title", p.title)
	}

	return httpReq, nil
}

func (p *OpenRouterProvider) ParseUpstreamResponse(resp *http.Response) (normalize.NormalizedResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return normalize.NormalizedResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	return parseOpenAIResponse(body)
}
