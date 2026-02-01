package policy

import (
	"fmt"

	"github.com/alereyleyva/agent-guard/internal/config"
)

type Engine struct {
	modelPolicy config.ModelPolicy
	toolPolicy  config.ToolPolicy
}

func NewEngine(cfg config.PolicyConfig) *Engine {
	return &Engine{
		modelPolicy: cfg.Models,
		toolPolicy:  cfg.Tools,
	}
}

func (e *Engine) EvaluateModel(model string) Decision {
	for _, denied := range e.modelPolicy.Deny {
		if matchesPattern(model, denied) {
			return NewDenyDecision(
				"MODEL_DENY",
				fmt.Sprintf("model %q is explicitly denied", model),
			)
		}
	}

	for _, allowed := range e.modelPolicy.Allow {
		if matchesPattern(model, allowed) {
			return NewAllowDecision(
				"MODEL_ALLOW",
				fmt.Sprintf("model %q is explicitly allowed", model),
			)
		}
	}

	if len(e.modelPolicy.Allow) > 0 {
		return NewDenyDecision(
			"MODEL_DEFAULT_DENY",
			fmt.Sprintf("model %q is not in the allow list", model),
		)
	}

	return NewDenyDecision(
		"MODEL_DEFAULT_DENY",
		"no model policy rules defined",
	)
}

func (e *Engine) EvaluateTool(toolName string) Decision {
	for _, denied := range e.toolPolicy.Deny {
		if matchesPattern(toolName, denied) {
			return NewDenyDecision(
				"TOOL_DENY",
				fmt.Sprintf("tool %q is explicitly denied", toolName),
			)
		}
	}

	for _, allowed := range e.toolPolicy.Allow {
		if matchesPattern(toolName, allowed) {
			return NewAllowDecision(
				"TOOL_ALLOW",
				fmt.Sprintf("tool %q is explicitly allowed", toolName),
			)
		}
	}

	if len(e.toolPolicy.Allow) > 0 {
		return NewDenyDecision(
			"TOOL_DEFAULT_DENY",
			fmt.Sprintf("tool %q is not in the allow list", toolName),
		)
	}

	return NewDenyDecision(
		"TOOL_DEFAULT_DENY",
		"no tool policy rules defined",
	)
}

func matchesPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}
	return value == pattern
}
