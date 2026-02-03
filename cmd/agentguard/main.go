package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alereyleyva/agent-guard/internal/audit"
	"github.com/alereyleyva/agent-guard/internal/config"
	"github.com/alereyleyva/agent-guard/internal/gateway"
	"github.com/alereyleyva/agent-guard/internal/policy"
	"github.com/alereyleyva/agent-guard/internal/provider"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	if envPath := os.Getenv("AGENTGUARD_CONFIG"); envPath != "" {
		*configPath = envPath
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var prov provider.Provider
	switch cfg.Provider.Type {
	case "openai_compatible", "openai":
		prov = provider.NewOpenAI(cfg.Provider.BaseURL, cfg.Provider.APIKey)
	case "openrouter":
		baseURL := cfg.Provider.OpenRouter.BaseURL
		if baseURL == "" {
			baseURL = cfg.Provider.BaseURL
		}
		apiKey := cfg.Provider.OpenRouter.APIKey
		if apiKey == "" {
			apiKey = cfg.Provider.APIKey
		}
		prov = provider.NewOpenRouter(baseURL, apiKey, cfg.Provider.OpenRouter.Referer, cfg.Provider.OpenRouter.Title)
	case "bedrock":
		bedrockCfg := cfg.Provider.Bedrock
		var err error
		prov, err = provider.NewBedrock(bedrockCfg.Region, bedrockCfg.Endpoint, bedrockCfg.AccessKeyID, bedrockCfg.SecretAccessKey, bedrockCfg.SessionToken)
		if err != nil {
			log.Fatalf("failed to initialize bedrock provider: %v", err)
		}
	default:
		log.Fatalf("unsupported provider type: %s", cfg.Provider.Type)
	}

	policyEngine := policy.NewEngine(cfg.Policy)

	logger := audit.NewStdoutLogger()

	flow := gateway.NewFlow(prov, policyEngine, logger)
	handler := gateway.NewHandler(flow)

	mux := http.NewServeMux()
	mux.Handle("/v1/chat/completions", handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	fmt.Printf("AgentGuard starting on %s\n", cfg.Listen)
	fmt.Printf("Provider: %s (%s)\n", prov.Name(), cfg.Provider.BaseURL)
	if err := http.ListenAndServe(cfg.Listen, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
