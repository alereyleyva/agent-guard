package provider

import (
	"testing"

	"github.com/alereyleyva/agent-guard/internal/config"
)

func TestNewFromConfig_OpenAI(t *testing.T) {
	cfg := config.ProviderConfig{
		Type:    "openai_compatible",
		BaseURL: "https://api.openai.com",
		APIKey:  "test-key",
	}

	prov, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if prov.Name() != "openai" {
		t.Errorf("Provider.Name() = %q, want %q", prov.Name(), "openai")
	}
}

func TestNewFromConfig_OpenRouter(t *testing.T) {
	cfg := config.ProviderConfig{
		Type:    "openrouter",
		BaseURL: "https://default.example.com",
		APIKey:  "fallback-key",
		OpenRouter: config.OpenRouterConfig{
			BaseURL: "https://openrouter.ai/api/v1",
			APIKey:  "openrouter-key",
			Referer: "https://example.com",
			Title:   "AgentGuard",
		},
	}

	prov, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	openRouter, ok := prov.(*OpenRouterProvider)
	if !ok {
		t.Fatalf("Provider type = %T, want *OpenRouterProvider", prov)
	}
	if openRouter.apiKey != "openrouter-key" {
		t.Errorf("OpenRouter apiKey = %q, want %q", openRouter.apiKey, "openrouter-key")
	}
	if openRouter.baseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("OpenRouter baseURL = %q, want %q", openRouter.baseURL, "https://openrouter.ai/api/v1")
	}
}

func TestNewFromConfig_OpenRouterMissingAPIKey(t *testing.T) {
	cfg := config.ProviderConfig{
		Type: "openrouter",
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("NewFromConfig() should return error when api key is missing")
	}
}

func TestNewFromConfig_UnsupportedProvider(t *testing.T) {
	cfg := config.ProviderConfig{Type: "unknown"}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("NewFromConfig() should return error for unsupported provider")
	}
}
