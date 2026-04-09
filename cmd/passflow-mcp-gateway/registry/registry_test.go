package registry

import (
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-mcp-gateway/mcp"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := New()

	server := mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message", "list_channels"},
		Healthy:  true,
	}

	r.Register(server)

	got, ok := r.Get("slack")
	if !ok {
		t.Fatal("expected to find server 'slack'")
	}
	if got.Endpoint != server.Endpoint {
		t.Errorf("expected endpoint %q, got %q", server.Endpoint, got.Endpoint)
	}
}

func TestRegistryFindByTool(t *testing.T) {
	r := New()

	r.Register(mcp.ServerInfo{
		Name:     "slack",
		Endpoint: "http://mcp-slack:8080",
		Tools:    []string{"send_message", "list_channels"},
		Healthy:  true,
	})
	r.Register(mcp.ServerInfo{
		Name:     "github",
		Endpoint: "http://mcp-github:8080",
		Tools:    []string{"create_issue", "list_issues"},
		Healthy:  true,
	})

	server, ok := r.FindByTool("slack.send_message")
	if !ok {
		t.Fatal("expected to find server for slack.send_message")
	}
	if server.Name != "slack" {
		t.Errorf("expected server 'slack', got %q", server.Name)
	}

	server, ok = r.FindByTool("github.create_issue")
	if !ok {
		t.Fatal("expected to find server for github.create_issue")
	}
	if server.Name != "github" {
		t.Errorf("expected server 'github', got %q", server.Name)
	}

	_, ok = r.FindByTool("unknown.tool")
	if ok {
		t.Error("expected not to find server for unknown.tool")
	}
}

func TestRegistryListAll(t *testing.T) {
	r := New()

	r.Register(mcp.ServerInfo{Name: "slack", Endpoint: "http://mcp-slack:8080", Tools: []string{"send_message"}, Healthy: true})
	r.Register(mcp.ServerInfo{Name: "github", Endpoint: "http://mcp-github:8080", Tools: []string{"create_issue"}, Healthy: true})

	servers := r.List()
	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}
