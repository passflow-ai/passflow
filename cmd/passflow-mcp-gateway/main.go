package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/handlers"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/proxy"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/registry"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	reg := registry.New()
	applyWorkspacePolicies(reg)
	prx := proxy.New(nil)
	h := handlers.New(reg, prx)

	registerFromEnv(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /servers", h.ListServers)
	mux.HandleFunc("GET /catalog/tools", h.ListBuilderCatalog)
	mux.HandleFunc("POST /tools/list", h.ListTools)
	mux.HandleFunc("POST /tools/call", h.CallTool)

	log.Printf("MCP Gateway starting on :%s", port)
	if err := http.ListenAndServe(":"+port, handlers.BearerAuth(mux)); err != nil {
		log.Fatal(err)
	}
}

func applyWorkspacePolicies(reg *registry.Registry) {
	raw := strings.TrimSpace(os.Getenv("MCP_WORKSPACE_TOOL_POLICY_JSON"))
	if raw == "" {
		return
	}
	policies := map[string]registry.WorkspaceToolPolicy{}
	if err := json.Unmarshal([]byte(raw), &policies); err != nil {
		log.Printf("failed to parse MCP_WORKSPACE_TOOL_POLICY_JSON: %v", err)
		return
	}
	reg.SetWorkspacePolicies(policies)
}

func registerFromEnv(reg *registry.Registry) {
	servers := map[string]string{
		"slack":      os.Getenv("MCP_SERVER_SLACK"),
		"github":     os.Getenv("MCP_SERVER_GITHUB"),
		"kubernetes": os.Getenv("MCP_SERVER_KUBERNETES"),
	}

	for name, config := range servers {
		if config == "" {
			continue
		}
		parts := strings.Split(config, ",")
		if len(parts) < 2 {
			continue
		}
		reg.Register(mcp.ServerInfo{
			Name:     name,
			Endpoint: parts[0],
			Tools:    parts[1:],
			Healthy:  true,
		})
		log.Printf("Registered MCP server: %s at %s with tools %v", name, parts[0], parts[1:])
	}

	// Register Obsidian vault server with explicit tool metadata
	registerObsidianServer(reg)
}

func registerObsidianServer(reg *registry.Registry) {
	endpoint := os.Getenv("MCP_SERVER_OBSIDIAN")
	if endpoint == "" {
		return
	}

	reg.Register(mcp.ServerInfo{
		Name:        "obsidian",
		DisplayName: "Obsidian Vault",
		Endpoint:    endpoint,
		Tools:       []string{"read_note", "write_note", "search_notes", "list_notes", "delete_note"},
		Healthy:     true,
		ToolDetails: map[string]mcp.Tool{
			"read_note": {
				Name:        "read_note",
				Description: "Read the content of a note from the vault by path",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":   map[string]interface{}{"type": "string", "description": "Path to the note relative to vault root"},
						"source": map[string]interface{}{"type": "string", "enum": []string{"workspace", "org_shared"}, "description": "Source vault scope"},
					},
					"required": []string{"path"},
				},
			},
			"write_note": {
				Name:        "write_note",
				Description: "Create or update a note in the vault",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":            map[string]interface{}{"type": "string", "description": "Path to the note relative to vault root"},
						"content":         map[string]interface{}{"type": "string", "description": "Markdown content of the note"},
						"createIfMissing": map[string]interface{}{"type": "boolean", "description": "Create the note if it does not exist"},
					},
					"required": []string{"path", "content"},
				},
			},
			"search_notes": {
				Name:        "search_notes",
				Description: "Search for notes in the vault by content, title, or tags",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query":            map[string]interface{}{"type": "string", "description": "Search query string"},
						"searchIn":         map[string]interface{}{"type": "string", "enum": []string{"content", "title", "tags", "all"}, "description": "Where to search"},
						"includeOrgShared": map[string]interface{}{"type": "boolean", "description": "Include org-shared vault in search"},
						"limit":            map[string]interface{}{"type": "integer", "description": "Maximum number of results"},
					},
					"required": []string{"query"},
				},
			},
			"list_notes": {
				Name:        "list_notes",
				Description: "List notes in the vault, optionally within a specific folder",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"folder":           map[string]interface{}{"type": "string", "description": "Folder path to list (empty for root)"},
						"includeOrgShared": map[string]interface{}{"type": "boolean", "description": "Include org-shared vault"},
						"recursive":        map[string]interface{}{"type": "boolean", "description": "List recursively into subfolders"},
					},
				},
			},
			"delete_note": {
				Name:        "delete_note",
				Description: "Delete a note from the vault (moves to trash with retention period)",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string", "description": "Path to the note to delete"},
					},
					"required": []string{"path"},
				},
			},
		},
	})
	log.Printf("Registered MCP server: obsidian at %s with tools %v",
		endpoint, []string{"read_note", "write_note", "search_notes", "list_notes", "delete_note"})
}
