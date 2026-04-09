package registry

import (
	"strings"
	"sync"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
)

// Registry manages MCP server registrations.
type Registry struct {
	mu              sync.RWMutex
	servers         map[string]mcp.ServerInfo
	circuitBreakers map[string]*CircuitBreaker
	workspacePolicy map[string]WorkspaceToolPolicy
}

type WorkspaceToolPolicy struct {
	Allowed          []string `json:"allowed"`
	Denied           []string `json:"denied"`
	ForceApproval    []string `json:"force_approval"`
	ReadOnly         []string `json:"read_only"`
	ProductionDenied []string `json:"production_denied"`
}

// New creates a new Registry.
func New() *Registry {
	return &Registry{
		servers:         make(map[string]mcp.ServerInfo),
		circuitBreakers: make(map[string]*CircuitBreaker),
		workspacePolicy: make(map[string]WorkspaceToolPolicy),
	}
}

func (r *Registry) SetWorkspacePolicies(policies map[string]WorkspaceToolPolicy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workspacePolicy = policies
}

// Register adds or updates an MCP server.
func (r *Registry) Register(server mcp.ServerInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers[server.Name] = server
}

// Unregister removes an MCP server.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.servers, name)
}

// Get retrieves a server by name.
func (r *Registry) Get(name string) (mcp.ServerInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	server, ok := r.servers[name]
	return server, ok
}

// FindByTool finds the server that provides a tool.
// Tool names are prefixed with server name: "slack.send_message".
func (r *Registry) FindByTool(toolName string) (mcp.ServerInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parts := strings.SplitN(toolName, ".", 2)
	if len(parts) < 2 {
		return mcp.ServerInfo{}, false
	}

	serverName := parts[0]
	localTool := parts[1]

	server, ok := r.servers[serverName]
	if !ok {
		return mcp.ServerInfo{}, false
	}

	for _, t := range server.Tools {
		if t == localTool {
			return server, true
		}
	}

	return mcp.ServerInfo{}, false
}

// List returns all registered servers.
func (r *Registry) List() []mcp.ServerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]mcp.ServerInfo, 0, len(r.servers))
	for _, s := range r.servers {
		result = append(result, s)
	}
	return result
}

// ListTools returns all tools from all healthy servers.
// If a server was registered with ToolDetails, the description and input
// schema for each tool are included so that LLM clients receive accurate
// metadata for function calling. Tools without detail entries are returned
// with an empty description.
func (r *Registry) ListTools() []mcp.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []mcp.Tool
	for _, server := range r.servers {
		if !server.Healthy {
			continue
		}
		for _, toolName := range server.Tools {
			t := mcp.Tool{
				Name: server.Name + "." + toolName,
			}
			if detail, ok := server.ToolDetails[toolName]; ok {
				t.Description = detail.Description
				t.InputSchema = detail.InputSchema
			}
			tools = append(tools, t)
		}
	}
	return tools
}

// ListBuilderCatalogTools returns Builder-oriented metadata for all healthy tools.
func (r *Registry) ListBuilderCatalogTools() []mcp.BuilderCatalogTool {
	return r.listBuilderCatalogTools("")
}

func (r *Registry) ListBuilderCatalogToolsForWorkspace(workspaceID string) []mcp.BuilderCatalogTool {
	return r.listBuilderCatalogTools(workspaceID)
}

func (r *Registry) listBuilderCatalogTools(workspaceID string) []mcp.BuilderCatalogTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]mcp.BuilderCatalogTool, 0)
	for _, server := range r.servers {
		if !server.Healthy {
			continue
		}
		for _, toolName := range server.Tools {
			qualifiedName := server.Name + "." + toolName
			if !isToolAllowedForWorkspace(r.workspacePolicy, workspaceID, qualifiedName, server.Name) {
				continue
			}
			usageMode, requiresApproval := inferUsageMode(toolName)
			allowedEnvironments := []string{"staging", "production"}
			if policy, ok := r.workspacePolicy[workspaceID]; ok {
				if matchesAnyPolicyPattern(policy.ReadOnly, qualifiedName, server.Name) {
					usageMode = "read"
					requiresApproval = false
				}
				if matchesAnyPolicyPattern(policy.ForceApproval, qualifiedName, server.Name) {
					requiresApproval = true
				}
				if matchesAnyPolicyPattern(policy.ProductionDenied, qualifiedName, server.Name) {
					allowedEnvironments = []string{"staging"}
				}
			}
			tool := mcp.BuilderCatalogTool{
				ID:                  qualifiedName,
				Name:                qualifiedName,
				ServerName:          server.Name,
				Category:            server.Name,
				UsageMode:           usageMode,
				RequiresApproval:    requiresApproval,
				AllowedEnvironments: allowedEnvironments,
			}
			if detail, ok := server.ToolDetails[toolName]; ok {
				tool.Description = detail.Description
			}
			tools = append(tools, tool)
		}
	}
	return tools
}

func matchesAnyPolicyPattern(patterns []string, qualifiedName, serverName string) bool {
	for _, pattern := range patterns {
		if matchesPolicyPattern(pattern, qualifiedName, serverName) {
			return true
		}
	}
	return false
}

func isToolAllowedForWorkspace(
	policies map[string]WorkspaceToolPolicy,
	workspaceID string,
	qualifiedName string,
	serverName string,
) bool {
	if workspaceID == "" {
		return true
	}
	policy, ok := policies[workspaceID]
	if !ok {
		return true
	}
	for _, denied := range policy.Denied {
		if matchesPolicyPattern(denied, qualifiedName, serverName) {
			return false
		}
	}
	if len(policy.Allowed) == 0 {
		return true
	}
	for _, allowed := range policy.Allowed {
		if matchesPolicyPattern(allowed, qualifiedName, serverName) {
			return true
		}
	}
	return false
}

func matchesPolicyPattern(pattern, qualifiedName, serverName string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if pattern == qualifiedName || pattern == serverName {
		return true
	}
	if strings.HasSuffix(pattern, ".*") {
		return strings.HasPrefix(qualifiedName, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

func inferUsageMode(name string) (string, bool) {
	lower := strings.ToLower(name)
	writePrefixes := []string{
		"create", "update", "delete", "send", "post", "write", "publish",
		"trigger", "deploy", "restart", "approve", "archive",
	}
	for _, prefix := range writePrefixes {
		if strings.Contains(lower, prefix) {
			return "write", true
		}
	}
	return "read", false
}
