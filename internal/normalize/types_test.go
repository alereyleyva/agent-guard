package normalize

import "testing"

func TestNormalizedResponse_ExtractToolNames(t *testing.T) {
	resp := NormalizedResponse{
		ToolCalls: []ToolCall{
			{Function: FunctionCall{Name: "search_web"}},
			{Function: FunctionCall{Name: "get_weather"}},
		},
	}

	names := resp.ExtractToolNames()
	if len(names) != 2 {
		t.Fatalf("ExtractToolNames() len = %d, want 2", len(names))
	}
	if names[0] != "search_web" || names[1] != "get_weather" {
		t.Errorf("ExtractToolNames() = %v, want [search_web get_weather]", names)
	}
}
