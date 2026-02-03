package provider

import (
	"fmt"

	"github.com/alereyleyva/agent-guard/internal/config"
)

type Factory func(cfg config.ProviderConfig) (Provider, error)

var providerRegistry = map[string]Factory{}

func RegisterFactory(providerType string, factory Factory) {
	if providerType == "" {
		panic("provider type is required")
	}
	if factory == nil {
		panic(fmt.Sprintf("provider factory is nil for %s", providerType))
	}
	if _, exists := providerRegistry[providerType]; exists {
		panic(fmt.Sprintf("provider factory already registered for %s", providerType))
	}
	providerRegistry[providerType] = factory
}

func NewFromConfig(cfg config.ProviderConfig) (Provider, error) {
	factory, ok := providerRegistry[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
	return factory(cfg)
}
