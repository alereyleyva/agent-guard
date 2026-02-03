package provider

type bedrockConverseRequest struct {
	Messages   []bedrockMessage      `json:"messages,omitempty"`
	System     []bedrockContentBlock `json:"system,omitempty"`
	ToolConfig *bedrockToolConfig    `json:"toolConfig,omitempty"`
}

type bedrockMessage struct {
	Role    string                `json:"role"`
	Content []bedrockContentBlock `json:"content"`
}

type bedrockContentBlock struct {
	Text       *string            `json:"text,omitempty"`
	ToolUse    *bedrockToolUse    `json:"toolUse,omitempty"`
	ToolResult *bedrockToolResult `json:"toolResult,omitempty"`
}

type bedrockToolUse struct {
	ToolUseID string      `json:"toolUseId"`
	Name      string      `json:"name"`
	Input     interface{} `json:"input,omitempty"`
}

type bedrockToolResult struct {
	ToolUseID string                `json:"toolUseId"`
	Content   []bedrockContentBlock `json:"content"`
	Status    string                `json:"status,omitempty"`
}

type bedrockToolConfig struct {
	Tools []bedrockTool `json:"tools"`
}

type bedrockTool struct {
	ToolSpec bedrockToolSpec `json:"toolSpec"`
}

type bedrockToolSpec struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema bedrockToolInputSchema `json:"inputSchema"`
}

type bedrockToolInputSchema struct {
	JSON interface{} `json:"json"`
}

type bedrockConverseResponse struct {
	Output struct {
		Message bedrockMessage `json:"message"`
	} `json:"output"`
}
