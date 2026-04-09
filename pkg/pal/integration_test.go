package pal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	agentdomain "github.com/passflow-ai/passflow/pkg/agent/domain"
	"github.com/passflow-ai/passflow/pkg/pal/palctrl"
	"github.com/passflow-ai/passflow/pkg/pal/palrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// mockAgentRepository implements agentdomain.AgentRepository for testing
type mockAgentRepository struct {
	agents map[string]*agentdomain.Agent
}

func newMockAgentRepository() *mockAgentRepository {
	return &mockAgentRepository{
		agents: make(map[string]*agentdomain.Agent),
	}
}

func (m *mockAgentRepository) Create(_ context.Context, agent *agentdomain.Agent) (string, error) {
	id := primitive.NewObjectID().Hex()
	agent.ID = id
	m.agents[id] = agent
	return id, nil
}

func (m *mockAgentRepository) FindByID(_ context.Context, workspaceID, agentID string) (*agentdomain.Agent, error) {
	if agent, ok := m.agents[agentID]; ok {
		return agent, nil
	}
	return nil, nil
}

func (m *mockAgentRepository) FindByWorkspace(_ context.Context, workspaceID string, opts agentdomain.PaginationOptions) (*agentdomain.AgentPage, error) {
	var agents []*agentdomain.Agent
	for _, agent := range m.agents {
		if agent.WorkspaceID == workspaceID {
			agents = append(agents, agent)
		}
	}
	return &agentdomain.AgentPage{
		Agents:     agents,
		Total:      int64(len(agents)),
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: 1,
	}, nil
}

func (m *mockAgentRepository) Update(_ context.Context, agent *agentdomain.Agent) error {
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) Delete(_ context.Context, workspaceID, agentID string) error {
	delete(m.agents, agentID)
	return nil
}

func (m *mockAgentRepository) UpdateStatus(_ context.Context, workspaceID, agentID, status string) error {
	if agent, ok := m.agents[agentID]; ok {
		agent.Status = status
	}
	return nil
}

func (m *mockAgentRepository) CountByWorkspace(_ context.Context, workspaceID string) (int64, error) {
	var count int64
	for _, agent := range m.agents {
		if agent.WorkspaceID == workspaceID {
			count++
		}
	}
	return count, nil
}

func (m *mockAgentRepository) FindByStatus(_ context.Context, status string) ([]*agentdomain.Agent, error) {
	var agents []*agentdomain.Agent
	for _, agent := range m.agents {
		if agent.Status == status {
			agents = append(agents, agent)
		}
	}
	return agents, nil
}

func (m *mockAgentRepository) ExistsByNameInWorkspace(_ context.Context, workspaceID, name string) (bool, error) {
	for _, agent := range m.agents {
		if agent.WorkspaceID == workspaceID && agent.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockAgentRepository) FindWithHeartbeatEnabled(_ context.Context) ([]*agentdomain.Agent, error) {
	var agents []*agentdomain.Agent
	for _, agent := range m.agents {
		if agent.Heartbeat.Enabled {
			agents = append(agents, agent)
		}
	}
	return agents, nil
}

func (m *mockAgentRepository) UpdateLastRun(_ context.Context, workspaceID, agentID string, t time.Time) error {
	if agent, ok := m.agents[agentID]; ok {
		agent.LastRunAt = &t
	}
	return nil
}

func (m *mockAgentRepository) DeleteByWorkspaceID(_ context.Context, workspaceID string) (int64, error) {
	var count int64
	for id, agent := range m.agents {
		if agent.WorkspaceID == workspaceID {
			delete(m.agents, id)
			count++
		}
	}
	return count, nil
}

// =============================================================================
// Test Setup
// =============================================================================

func setupTestApp(repo agentdomain.AgentRepository) *fiber.App {
	app := fiber.New()
	router := palrouter.NewPALRouter(repo)
	router.Register(app)
	return app
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestPAL_ValidateEndpoint_Success(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	validYAML := `
agent:
  name: test-agent
  description: A test agent
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
    - tool2
`

	reqBody := map[string]string{
		"content": validYAML,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/validate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, true, result["valid"])
}

func TestPAL_ValidateEndpoint_InvalidYAML(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	reqBody := map[string]string{
		"content": "invalid: yaml: content: [",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/validate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, false, result["valid"])
	assert.NotEmpty(t, result["errors"])
}

func TestPAL_ValidateEndpoint_MissingRequiredFields(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	// Missing model section
	reqBody := map[string]string{
		"content": `
agent:
  name: test-agent
react:
  tools:
    - tool1
`,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/validate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	// The result will be invalid due to missing model
	assert.Equal(t, false, result["valid"])
}

func TestPAL_ApplyEndpoint_DryRun(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	validYAML := `
agent:
  name: test-agent
  description: A test agent
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`

	reqBody := map[string]interface{}{
		"content":      validYAML,
		"workspace_id": "ws-123",
		"dry_run":      true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/apply", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, true, result["dryRun"])
}

func TestPAL_ApplyEndpoint_CreateAgent(t *testing.T) {
	repo := newMockAgentRepository()
	app := setupTestApp(repo)

	validYAML := `
agent:
  name: new-agent
  description: A new agent
model:
  provider: openai
  name: gpt-4
react:
  tools:
    - github-search
`

	reqBody := map[string]interface{}{
		"content":      validYAML,
		"workspace_id": "ws-456",
		"dry_run":      false,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/apply", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.NotEmpty(t, result["agentId"])

	// Verify agent was created in repo
	assert.Len(t, repo.agents, 1)
}

func TestPAL_ApplyEndpoint_InvalidContent(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	reqBody := map[string]interface{}{
		"content":      "not: valid: yaml: [",
		"workspace_id": "ws-789",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/apply", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPAL_ExportEndpoint_Success(t *testing.T) {
	repo := newMockAgentRepository()

	// Pre-populate an agent
	agent := &agentdomain.Agent{
		ID:          "agent-123",
		Name:        "exported-agent",
		WorkspaceID: "ws-export",
		Persona:     "You are a helpful assistant",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-opus",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true},
			{ID: "tool2", Name: "tool2", Enabled: true},
		},
	}
	repo.agents["agent-123"] = agent

	app := setupTestApp(repo)

	req := httptest.NewRequest(http.MethodGet, "/workspaces/ws-export/pal/export/agent-123?format=yaml", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "agent-123", result["agentId"])
	assert.Equal(t, "yaml", result["format"])
	assert.NotEmpty(t, result["content"])
}

func TestPAL_ExportEndpoint_NotFound(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	req := httptest.NewRequest(http.MethodGet, "/workspaces/ws-123/pal/export/nonexistent", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPAL_ExportEndpoint_JSONFormat(t *testing.T) {
	repo := newMockAgentRepository()

	agent := &agentdomain.Agent{
		ID:          "agent-json",
		Name:        "json-export-agent",
		WorkspaceID: "ws-json",
		Model: agentdomain.ModelConfig{
			Provider: "openai",
			ModelID:  "gpt-4",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "search", Name: "search", Enabled: true},
		},
	}
	repo.agents["agent-json"] = agent

	app := setupTestApp(repo)

	req := httptest.NewRequest(http.MethodGet, "/workspaces/ws-json/pal/export/agent-json?format=json", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "json", result["format"])
}

func TestPAL_DiffEndpoint_Success(t *testing.T) {
	repo := newMockAgentRepository()

	// Pre-populate an agent
	agent := &agentdomain.Agent{
		ID:          "agent-diff",
		Name:        "diff-agent",
		WorkspaceID: "ws-diff",
		Persona:     "Original persona",
		Model: agentdomain.ModelConfig{
			Provider: "anthropic",
			ModelID:  "claude-3-sonnet",
		},
		Tools: []agentdomain.AgentTool{
			{ID: "tool1", Name: "tool1", Enabled: true},
		},
	}
	repo.agents["agent-diff"] = agent

	app := setupTestApp(repo)

	// New content with changes
	newYAML := `
agent:
  name: diff-agent
  description: Updated description
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
    - tool2
`

	reqBody := map[string]interface{}{
		"agent_id": "agent-diff",
		"content":  newYAML,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/diff", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "agent-diff", result["agentId"])
	// hasChanges and summary should always be present
	_, hasChangesField := result["hasChanges"]
	assert.True(t, hasChangesField, "response should have hasChanges field")
	_, hasSummary := result["summary"]
	assert.True(t, hasSummary, "response should have summary field")
}

func TestPAL_DiffEndpoint_AgentNotFound(t *testing.T) {
	app := setupTestApp(newMockAgentRepository())

	reqBody := map[string]interface{}{
		"agent_id": "nonexistent",
		"content": `
agent:
  name: test
model:
  provider: anthropic
  name: claude-3-opus
react:
  tools:
    - tool1
`,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pal/diff", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// Controller Unit Tests
// =============================================================================

func TestValidateController_EmptyContent(t *testing.T) {
	app := fiber.New()
	ctrl := palctrl.NewValidateController()
	app.Post("/validate", ctrl.Handler())

	reqBody := map[string]string{
		"content": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestValidateController_MissingContent(t *testing.T) {
	app := fiber.New()
	ctrl := palctrl.NewValidateController()
	app.Post("/validate", ctrl.Handler())

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// =============================================================================
// Round-Trip Tests
// =============================================================================

func TestPAL_RoundTrip_ApplyThenExport(t *testing.T) {
	repo := newMockAgentRepository()
	app := setupTestApp(repo)

	// Step 1: Apply a PAL spec
	originalYAML := `
agent:
  name: round-trip-agent
  description: Testing round trip
model:
  provider: anthropic
  name: claude-3-opus
  config:
    temperature: 0.7
    max_tokens: 4096
react:
  tools:
    - github-search
    - slack-post
  max_iterations: 10
guardrails:
  allowed_tools:
    - github-search
    - slack-post
  timeout_seconds: 300
`

	applyReq := map[string]interface{}{
		"content":      originalYAML,
		"workspace_id": "ws-roundtrip",
		"dry_run":      false,
	}
	jsonBody, _ := json.Marshal(applyReq)

	req := httptest.NewRequest(http.MethodPost, "/pal/apply", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var applyResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&applyResult)
	require.NoError(t, err)

	agentID := applyResult["agentId"].(string)
	require.NotEmpty(t, agentID)

	// Step 2: Export the agent back to YAML
	exportReq := httptest.NewRequest(http.MethodGet, "/workspaces/ws-roundtrip/pal/export/"+agentID+"?format=yaml", nil)

	exportResp, err := app.Test(exportReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, exportResp.StatusCode)

	var exportResult map[string]interface{}
	err = json.NewDecoder(exportResp.Body).Decode(&exportResult)
	require.NoError(t, err)

	exportedContent := exportResult["content"].(string)
	assert.Contains(t, exportedContent, "round-trip-agent")
	assert.Contains(t, exportedContent, "anthropic")
	assert.Contains(t, exportedContent, "claude-3-opus")
}
