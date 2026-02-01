package audit

import "time"

const (
	EventTypeLLMRequest     = "llm_request"
	EventTypeLLMResponse    = "llm_response"
	EventTypeToolProposal   = "tool_proposal"
	EventTypePolicyDecision = "policy_decision"
)

type Event struct {
	TraceID   string `json:"trace_id"`
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
	Decision  string `json:"decision,omitempty"`
	RuleID    string `json:"rule_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

func NewEvent(traceID, eventType string) Event {
	return Event{
		TraceID:   traceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
	}
}

func (e Event) WithProvider(provider string) Event {
	e.Provider = provider
	return e
}

func (e Event) WithModel(model string) Event {
	e.Model = model
	return e
}

func (e Event) WithDecision(action, ruleID, reason string) Event {
	e.Decision = action
	e.RuleID = ruleID
	e.Reason = reason
	return e
}

func (e Event) WithToolName(toolName string) Event {
	e.ToolName = toolName
	return e
}

func (e Event) WithHash(hash string) Event {
	e.Hash = hash
	return e
}

func (e Event) WithStream(stream bool) Event {
	e.Stream = stream
	return e
}
