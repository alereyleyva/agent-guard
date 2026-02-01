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
	Type    string `yaml:"type"`
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
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
	if c.Provider.BaseURL == "" {
		return fmt.Errorf("provider base_url is required")
	}
	return nil
}
