package mcp

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListToolsResponse from gateway.
type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// CallRequest to gateway.
type CallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Content in response.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CallResponse from gateway.
type CallResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}
