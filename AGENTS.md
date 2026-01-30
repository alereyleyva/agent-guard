# AgentGuard

## Overview

AgentGuard is a multi-provider AI gateway and deterministic security boundary for LLM-powered systems.

It provides:

- A unified API surface for model invocation
- Deterministic policy validation of all LLM traffic
- Structured audit and observability events
- Provider abstraction through adapters
- Enforcement-ready decision logic

AgentGuard sits between clients (agents/applications) and upstream model providers.

---

## Core Responsibilities

AgentGuard must:

1. Accept LLM requests through a unified API.
2. Normalize requests into a canonical internal representation.
3. Evaluate deterministic security policy.
4. Forward requests to a configured upstream provider.
5. Normalize upstream responses.
6. Detect tool proposals in responses.
7. Emit structured audit events.
8. Return responses unchanged to clients.

All policy decisions are computed for every request and tool proposal.

---

## Public API Surface

Expose:

`POST /v1/chat/completions`

This API remains stable regardless of upstream provider.

---

## Architecture

### 1. Provider Adapter Layer

AgentGuard supports multiple upstream providers via adapters.

Examples:
- OpenAI-compatible providers
- AWS Bedrock
- OpenRouter
- Gemini (via adapter)
- Anthropic (via adapter)

Each provider must implement:

```go
type Provider interface {
    Name() string
    BuildUpstreamRequest(NormalizedRequest) (*http.Request, error)
    ParseUpstreamResponse(*http.Response) (NormalizedResponse, error)
}
```

The gateway must not contain provider-specific logic.

### 2. Normalized Internal Model

All requests must be converted into:

```go
type NormalizedRequest struct {
    Model    string
    Messages []Message
    Stream   bool
    Metadata map[string]string
}
```

All responses must be converted into:

```go
type NormalizedResponse struct {
    Model     string
    Content   string
    ToolCalls []ToolCall
    RawBody   []byte
}
```

Policy evaluation must operate only on normalized structures.

### 3. Policy Engine

The policy engine must be:

- **Deterministic**
- **YAML-configurable**
- **Executed for every request and tool proposal**
- **Independent of provider implementation**

Decision model:

```go
type Decision struct {
    Action string // allow | deny
    RuleID string
    Reason string
}
```

Policy must support:

- Model allow/deny
- Tool proposal allow/deny
- Basic prompt validation (pattern-based rules)

Policy decisions are logged but do not modify traffic.

### 4. Observability & Audit

AgentGuard emits structured JSON events to stdout.

Each event must include:

- `trace_id`
- `timestamp`
- `provider`
- `model`
- `event_type`
- `decision` (if applicable)
- `rule_id` (if applicable)
- `payload hashes` (never raw content by default)

Event types:

- `llm_request`
- `llm_response`
- `tool_proposal`
- `policy_decision`

Raw prompts and responses must not be logged by default.

---

## Configuration

Example:

```yaml
listen: "127.0.0.1:8080"

provider:
  type: "openai_compatible"
  base_url: "[https://api.openai.com](https://api.openai.com)"
  api_key: "env:OPENAI_API_KEY"

policy:
  models:
    allow:
      - "gpt-4o"
      - "gpt-4.1"
  tools:
    deny:
      - "shell_exec"
```

Provider type determines which adapter is used.

---

## Design Constraints

- **Language:** Go
- **Single binary**
- **Clear package separation:** `gateway`, `provider`, `policy`, `audit`, `config`, `normalize`
- **Fail-safe defaults**
- **Strict JSON parsing**
- **No dashboards**
- **No cost analytics**
- **No retry/fallback orchestration**
- **No routing optimization features**

AgentGuard is a security boundary, not a routing optimizer.

---

## Invariants

- All requests must be normalized before policy evaluation.
- Policy evaluation must not depend on provider-specific schemas.
- Decisions must be deterministic and reproducible.
- No probabilistic scoring or AI-based enforcement logic.