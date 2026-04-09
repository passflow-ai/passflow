package mcp

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListToolsResponse is returned by POST /tools/list.
type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// BuilderCatalogTool is a builder-oriented view of a tool with coarse
// governance metadata suitable for architecture design.
type BuilderCatalogTool struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	ServerName          string   `json:"server_name"`
	Category            string   `json:"category"`
	UsageMode           string   `json:"usage_mode"`
	RequiresApproval    bool     `json:"requires_approval"`
	AllowedEnvironments []string `json:"allowed_environments,omitempty"`
}

// BuilderCatalogResponse is returned by GET /catalog/tools.
type BuilderCatalogResponse struct {
	Tools []BuilderCatalogTool `json:"tools"`
}

// CallRequest is sent to POST /tools/call.
type CallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Content represents a content block in MCP response.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CallResponse is returned by POST /tools/call.
type CallResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// ServerInfo contains MCP server metadata.
type ServerInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Endpoint    string   `json:"endpoint"`
	Tools       []string `json:"tools"`
	Healthy     bool     `json:"healthy"`
	// ToolDetails holds optional per-tool metadata (description, input schema)
	// keyed by tool name (without the server prefix). When present, ListTools
	// uses this to expose accurate descriptions to LLM clients.
	ToolDetails map[string]Tool `json:"tool_details,omitempty"`
}

// ToolDetail provides governance metadata for a tool in the builder catalog.
type ToolDetail struct {
	UsageMode        string `json:"usage_mode"`         // read | write
	RequiresApproval bool   `json:"requires_approval"`
}

// ObsidianReadNoteParams represents read_note tool parameters.
type ObsidianReadNoteParams struct {
	Path   string `json:"path"`
	Source string `json:"source"` // workspace | org_shared
}

// ObsidianWriteNoteParams represents write_note tool parameters.
type ObsidianWriteNoteParams struct {
	Path            string `json:"path"`
	Content         string `json:"content"`
	CreateIfMissing bool   `json:"createIfMissing"`
}

// ObsidianSearchParams represents search_notes tool parameters.
type ObsidianSearchParams struct {
	Query            string `json:"query"`
	SearchIn         string `json:"searchIn"` // content, title, tags, all
	IncludeOrgShared bool   `json:"includeOrgShared"`
	Limit            int    `json:"limit"`
}

// ObsidianListParams represents list_notes tool parameters.
type ObsidianListParams struct {
	Folder           string `json:"folder,omitempty"`
	IncludeOrgShared bool   `json:"includeOrgShared"`
	Recursive        bool   `json:"recursive"`
}

// ObsidianDeleteParams represents delete_note tool parameters.
type ObsidianDeleteParams struct {
	Path string `json:"path"`
}
