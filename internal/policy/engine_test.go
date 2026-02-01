package policy

import (
	"testing"

	"github.com/alereyleyva/agent-guard/internal/config"
)

func TestEvaluateModel_AllowList(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{
		Models: config.ModelPolicy{
			Allow: []string{"gpt-4o", "gpt-4o-mini"},
		},
	})

	tests := []struct {
		name      string
		model     string
		wantAllow bool
	}{
		{"allowed model", "gpt-4o", true},
		{"another allowed model", "gpt-4o-mini", true},
		{"not in allow list", "gpt-3.5-turbo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.EvaluateModel(tt.model)
			if decision.IsAllowed() != tt.wantAllow {
				t.Errorf("EvaluateModel(%q) = %v, want allowed=%v", tt.model, decision.Action, tt.wantAllow)
			}
		})
	}
}

func TestEvaluateModel_DenyList(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{
		Models: config.ModelPolicy{
			Allow: []string{"gpt-4o", "gpt-4o-mini"},
			Deny:  []string{"gpt-4o-mini"},
		},
	})

	decision := engine.EvaluateModel("gpt-4o-mini")
	if decision.IsAllowed() {
		t.Error("EvaluateModel should deny gpt-4o-mini (deny takes precedence)")
	}

	decision = engine.EvaluateModel("gpt-4o")
	if !decision.IsAllowed() {
		t.Error("EvaluateModel should allow gpt-4o")
	}
}

func TestEvaluateModel_NoRules(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{})

	decision := engine.EvaluateModel("any-model")
	if decision.IsAllowed() {
		t.Error("EvaluateModel should deny when no rules defined")
	}
}

func TestEvaluateTool_DenyList(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{
		Tools: config.ToolPolicy{
			Deny: []string{"shell_exec", "dangerous_command"},
		},
	})

	tests := []struct {
		name      string
		tool      string
		wantAllow bool
	}{
		{"denied tool", "shell_exec", false},
		{"another denied tool", "dangerous_command", false},
		{"not denied tool", "search_web", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.EvaluateTool(tt.tool)
			if decision.IsAllowed() != tt.wantAllow {
				t.Errorf("EvaluateTool(%q) = %v, want allowed=%v", tt.tool, decision.Action, tt.wantAllow)
			}
		})
	}
}

func TestEvaluateTool_AllowList(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{
		Tools: config.ToolPolicy{
			Allow: []string{"search_web", "get_weather"},
		},
	})

	tests := []struct {
		name      string
		tool      string
		wantAllow bool
	}{
		{"allowed tool", "search_web", true},
		{"another allowed tool", "get_weather", true},
		{"not in allow list", "shell_exec", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.EvaluateTool(tt.tool)
			if decision.IsAllowed() != tt.wantAllow {
				t.Errorf("EvaluateTool(%q) = %v, want allowed=%v", tt.tool, decision.Action, tt.wantAllow)
			}
		})
	}
}

func TestMatchesPattern_Wildcard(t *testing.T) {
	engine := NewEngine(config.PolicyConfig{
		Models: config.ModelPolicy{
			Allow: []string{"gpt-4*"},
		},
	})

	tests := []struct {
		name      string
		model     string
		wantAllow bool
	}{
		{"matches prefix", "gpt-4o", true},
		{"matches prefix with suffix", "gpt-4-turbo", true},
		{"does not match", "gpt-3.5-turbo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.EvaluateModel(tt.model)
			if decision.IsAllowed() != tt.wantAllow {
				t.Errorf("EvaluateModel(%q) with wildcard = %v, want allowed=%v", tt.model, decision.Action, tt.wantAllow)
			}
		})
	}
}
