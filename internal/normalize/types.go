package normalize

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

type NormalizedRequest struct {
	Model    string            `json:"model"`
	Messages []Message         `json:"messages"`
	Stream   bool              `json:"stream"`
	Tools    []Tool            `json:"tools,omitempty"`
	Metadata map[string]string `json:"-"`
}

type NormalizedResponse struct {
	ID        string     `json:"id"`
	Model     string     `json:"model"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	RawBody   []byte     `json:"-"`
}

func (r *NormalizedResponse) ExtractToolNames() []string {
	names := make([]string, 0, len(r.ToolCalls))
	for _, tc := range r.ToolCalls {
		names = append(names, tc.Function.Name)
	}
	return names
}
