package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/registry"
)

// PublicServerInfo is the external representation of a registered MCP server
// returned by GET /servers. It deliberately omits the internal Endpoint URL
// to prevent infrastructure details from leaking to callers.
type PublicServerInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ToolCount   int    `json:"tool_count"`
}

// toPublicServerInfo converts an internal ServerInfo to its public form.
func toPublicServerInfo(s mcp.ServerInfo) PublicServerInfo {
	return PublicServerInfo{
		Name:      s.Name,
		ToolCount: len(s.Tools),
	}
}

// Proxy handles forwarding requests to MCP servers.
type Proxy interface {
	Forward(server mcp.ServerInfo, req mcp.CallRequest) (mcp.CallResponse, error)
}

// Handlers contains HTTP handlers for the gateway.
type Handlers struct {
	registry *registry.Registry
	proxy    Proxy
}

// New creates a new Handlers instance.
func New(reg *registry.Registry, proxy Proxy) *Handlers {
	return &Handlers{
		registry: reg,
		proxy:    proxy,
	}
}

// ListTools handles POST /tools/list.
func (h *Handlers) ListTools(w http.ResponseWriter, r *http.Request) {
	tools := h.registry.ListTools()

	resp := mcp.ListToolsResponse{Tools: tools}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListBuilderCatalog handles GET /catalog/tools.
func (h *Handlers) ListBuilderCatalog(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	tools := h.registry.ListBuilderCatalogToolsForWorkspace(workspaceID)
	resp := mcp.BuilderCatalogResponse{Tools: tools}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListServers handles GET /servers.
// The response uses PublicServerInfo to avoid exposing internal Endpoint URLs.
func (h *Handlers) ListServers(w http.ResponseWriter, r *http.Request) {
	servers := h.registry.List()

	public := make([]PublicServerInfo, len(servers))
	for i, s := range servers {
		public[i] = toPublicServerInfo(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(public)
}

// maxRequestBodyBytes is the maximum permitted size of a JSON request body for
// CallTool. Requests exceeding this size are rejected with 400 to prevent
// memory exhaustion from oversized payloads.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// CallTool handles POST /tools/call.
func (h *Handlers) CallTool(w http.ResponseWriter, r *http.Request) {
	// Enforce an upper bound on request body size before decoding to prevent
	// unbounded memory consumption from maliciously large payloads.
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req mcp.CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	server, ok := h.registry.FindByTool(req.Name)
	if !ok {
		http.Error(w, `{"error":"tool not found"}`, http.StatusNotFound)
		return
	}

	if h.proxy == nil {
		http.Error(w, `{"error":"proxy not configured"}`, http.StatusInternalServerError)
		return
	}

	resp, err := h.proxy.Forward(server, req)
	if err != nil {
		// Log the full error server-side (may contain internal hostnames/details).
		log.Printf("[mcp-gateway] upstream error for tool %q: %v", req.Name, err)
		// Return a generic message to the caller — never expose raw Go errors
		// which may leak internal infrastructure details such as hostnames.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(mcp.CallResponse{
			Content: []mcp.Content{{Type: "text", Text: "upstream MCP server error"}},
			IsError: true,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Health handles GET /health.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
