package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen   string         `yaml:"listen"`
	Provider ProviderConfig `yaml:"provider"`
	Policy   PolicyConfig   `yaml:"policy"`
}

type ProviderConfig struct {
	Type       string           `yaml:"type"`
	BaseURL    string           `yaml:"base_url"`
	APIKey     string           `yaml:"api_key"`
	OpenRouter OpenRouterConfig `yaml:"openrouter"`
	Bedrock    BedrockConfig    `yaml:"bedrock"`
}

type OpenRouterConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Referer string `yaml:"referer"`
	Title   string `yaml:"title"`
}

type BedrockConfig struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	SessionToken    string `yaml:"session_token"`
	Endpoint        string `yaml:"endpoint"`
}

type PolicyConfig struct {
	Models ModelPolicy `yaml:"models"`
	Tools  ToolPolicy  `yaml:"tools"`
}

type ModelPolicy struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}

type ToolPolicy struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg.Provider.APIKey = resolveEnvVar(cfg.Provider.APIKey)
	cfg.Provider.BaseURL = resolveEnvVar(cfg.Provider.BaseURL)
	cfg.Provider.OpenRouter.APIKey = resolveEnvVar(cfg.Provider.OpenRouter.APIKey)
	cfg.Provider.OpenRouter.BaseURL = resolveEnvVar(cfg.Provider.OpenRouter.BaseURL)
	cfg.Provider.OpenRouter.Referer = resolveEnvVar(cfg.Provider.OpenRouter.Referer)
	cfg.Provider.OpenRouter.Title = resolveEnvVar(cfg.Provider.OpenRouter.Title)
	cfg.Provider.Bedrock.Region = resolveEnvVar(cfg.Provider.Bedrock.Region)
	cfg.Provider.Bedrock.AccessKeyID = resolveEnvVar(cfg.Provider.Bedrock.AccessKeyID)
	cfg.Provider.Bedrock.SecretAccessKey = resolveEnvVar(cfg.Provider.Bedrock.SecretAccessKey)
	cfg.Provider.Bedrock.SessionToken = resolveEnvVar(cfg.Provider.Bedrock.SessionToken)
	cfg.Provider.Bedrock.Endpoint = resolveEnvVar(cfg.Provider.Bedrock.Endpoint)

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func resolveEnvVar(value string) string {
	if strings.HasPrefix(value, "env:") {
		envName := strings.TrimPrefix(value, "env:")
		return os.Getenv(envName)
	}
	return value
}

func (c *Config) validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen address is required")
	}
	if c.Provider.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	switch c.Provider.Type {
	case "openai_compatible", "openai":
		if c.Provider.BaseURL == "" {
			return fmt.Errorf("provider base_url is required")
		}
	case "openrouter":
		if c.Provider.APIKey == "" && c.Provider.OpenRouter.APIKey == "" {
			return fmt.Errorf("openrouter api key is required")
		}
	case "bedrock":
		if c.Provider.Bedrock.Region == "" {
			return fmt.Errorf("bedrock region is required")
		}
		if c.Provider.Bedrock.AccessKeyID != "" && c.Provider.Bedrock.SecretAccessKey == "" {
			return fmt.Errorf("bedrock secret_access_key is required when access_key_id is set")
		}
		if c.Provider.Bedrock.SecretAccessKey != "" && c.Provider.Bedrock.AccessKeyID == "" {
			return fmt.Errorf("bedrock access_key_id is required when secret_access_key is set")
		}
	default:
		return fmt.Errorf("unsupported provider type: %s", c.Provider.Type)
	}
	return nil
}
