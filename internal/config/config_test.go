package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
listen: "127.0.0.1:8080"
provider:
  type: "openai_compatible"
  base_url: "https://api.openai.com"
  api_key: "test-key"
policy:
  models:
    allow:
      - "gpt-4o"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Listen != "127.0.0.1:8080" {
		t.Errorf("Listen = %q, want %q", cfg.Listen, "127.0.0.1:8080")
	}
	if cfg.Provider.Type != "openai_compatible" {
		t.Errorf("Provider.Type = %q, want %q", cfg.Provider.Type, "openai_compatible")
	}
	if cfg.Provider.APIKey != "test-key" {
		t.Errorf("Provider.APIKey = %q, want %q", cfg.Provider.APIKey, "test-key")
	}
	if len(cfg.Policy.Models.Allow) != 1 || cfg.Policy.Models.Allow[0] != "gpt-4o" {
		t.Errorf("Policy.Models.Allow = %v, want [gpt-4o]", cfg.Policy.Models.Allow)
	}
}

func TestLoad_EnvVarInjection(t *testing.T) {
	const testKey = "test-api-key-12345"
	os.Setenv("TEST_API_KEY", testKey)
	defer os.Unsetenv("TEST_API_KEY")

	content := `
listen: "127.0.0.1:8080"
provider:
  type: "openai"
  base_url: "https://api.openai.com"
  api_key: "env:TEST_API_KEY"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Provider.APIKey != testKey {
		t.Errorf("Provider.APIKey = %q, want %q (from env)", cfg.Provider.APIKey, testKey)
	}
}

func TestLoad_MissingListenAddress(t *testing.T) {
	content := `
provider:
  type: "openai"
  base_url: "https://api.openai.com"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error for missing listen address")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestLoad_OpenRouterRequiresAPIKey(t *testing.T) {
	content := `
listen: "127.0.0.1:8080"
provider:
  type: "openrouter"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error when openrouter api key is missing")
	}
}

func TestLoad_BedrockRequiresRegion(t *testing.T) {
	content := `
listen: "127.0.0.1:8080"
provider:
  type: "bedrock"
  bedrock:
    access_key_id: "test"
    secret_access_key: "secret"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error when bedrock region is missing")
	}
}

func TestLoad_BedrockStaticCredsValidation(t *testing.T) {
	content := `
listen: "127.0.0.1:8080"
provider:
  type: "bedrock"
  bedrock:
    region: "us-east-1"
    access_key_id: "test"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error when bedrock secret_access_key is missing")
	}
}
