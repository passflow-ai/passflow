package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/registry"
)

func TestListToolsHandler(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	rec := httptest.NewRecorder()

	h.ListTools(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp mcp.ListToolsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(resp.Tools))
	}
	if resp.Tools[0].Name != "slack.send_message" {
		t.Errorf("expected tool name 'slack.send_message', got %q", resp.Tools[0].Name)
	}
}

func TestListServersHandler(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message"},
		Healthy:  true,
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	rec := httptest.NewRecorder()

	h.ListServers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var servers []mcp.ServerInfo
	if err := json.NewDecoder(rec.Body).Decode(&servers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}
}

func TestListBuilderCatalogHandler(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message", "list_channels"},
		Healthy:  true,
		ToolDetails: map[string]mcp.Tool{
			"send_message":  {Description: "Send a Slack message"},
			"list_channels": {Description: "List channels"},
		},
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/catalog/tools", nil)
	rec := httptest.NewRecorder()

	h.ListBuilderCatalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp mcp.BuilderCatalogResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(resp.Tools))
	}
	if resp.Tools[0].ID == "" {
		t.Fatal("expected tool id")
	}
}

func TestListBuilderCatalogHandler_FiltersByWorkspacePolicy(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message", "list_channels"},
		Healthy:  true,
	})
	reg.SetWorkspacePolicies(map[string]registry.WorkspaceToolPolicy{
		"ws-1": {Allowed: []string{"slack.list_channels"}},
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/catalog/tools?workspace_id=ws-1", nil)
	rec := httptest.NewRecorder()

	h.ListBuilderCatalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp mcp.BuilderCatalogResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(resp.Tools))
	}
	if resp.Tools[0].ID != "slack.list_channels" {
		t.Fatalf("unexpected tool id %q", resp.Tools[0].ID)
	}
}

func TestListBuilderCatalogHandler_AppliesGranularWorkspacePolicy(t *testing.T) {
	reg := registry.New()
	reg.Register(mcp.ServerInfo{
		Name:     "github",
		Endpoint: "http://mcp-github:8080",
		Tools:    []string{"create_issue", "list_issues"},
		Healthy:  true,
	})
	reg.SetWorkspacePolicies(map[string]registry.WorkspaceToolPolicy{
		"ws-1": {
			ForceApproval:    []string{"github.create_issue"},
			ProductionDenied: []string{"github.create_issue"},
			ReadOnly:         []string{"github.list_issues"},
		},
	})

	h := New(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/catalog/tools?workspace_id=ws-1", nil)
	rec := httptest.NewRecorder()

	h.ListBuilderCatalog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp mcp.BuilderCatalogResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(resp.Tools))
	}

	toolsByID := map[string]mcp.BuilderCatalogTool{}
	for _, tool := range resp.Tools {
		toolsByID[tool.ID] = tool
	}

	if !toolsByID["github.create_issue"].RequiresApproval {
		t.Fatal("expected github.create_issue to require approval")
	}
	if len(toolsByID["github.create_issue"].AllowedEnvironments) != 1 || toolsByID["github.create_issue"].AllowedEnvironments[0] != "staging" {
		t.Fatalf("unexpected allowed environments: %#v", toolsByID["github.create_issue"].AllowedEnvironments)
	}
	if toolsByID["github.list_issues"].UsageMode != "read" {
		t.Fatalf("expected github.list_issues to be read-only, got %q", toolsByID["github.list_issues"].UsageMode)
	}
}

func TestCallToolHandler_NotFound(t *testing.T) {
	reg := registry.New()
	h := New(reg, nil)

	body := mcp.CallRequest{
		Name:      "unknown.tool",
		Arguments: map[string]interface{}{},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	h.CallTool(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
