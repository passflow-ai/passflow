package api

import "fmt"

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors,omitempty"`
	Warnings []ValidationIssue `json:"warnings,omitempty"`
}

type ValidationIssue struct {
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ApplyResult struct {
	AgentID   string   `json:"agentId"`
	Action    string   `json:"action"` // "created" or "updated"
	Changes   []Change `json:"changes,omitempty"`
	DryRun    bool     `json:"dryRun"`
	AgentName string   `json:"agentName,omitempty"`
	Success   bool     `json:"success"`
	Timestamp string   `json:"timestamp,omitempty"`
}

type Change struct {
	Field    string `json:"field"`
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty"`
}

type ExportResult struct {
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName"`
	Content   string `json:"content"`
	Format    string `json:"format"`
	Timestamp string `json:"timestamp,omitempty"`
}

type DiffResult struct {
	AgentID    string      `json:"agentId"`
	HasChanges bool        `json:"hasChanges"`
	Changes    []DiffChange `json:"changes,omitempty"`
	Summary    DiffSummary `json:"summary,omitempty"`
}

type DiffChange struct {
	Path     string `json:"path"`
	Type     string `json:"type"` // "added", "removed", "modified"
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty"`
}

type DiffSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}

// ValidatePAL validates a PAL spec without applying it.
// POST /api/v1/pal/validate
func (c *Client) ValidatePAL(content string, mode string) (*ValidationResult, error) {
	req := map[string]string{
		"content": content,
	}

	var resp APIResponse[ValidationResult]
	if err := c.post("/api/v1/pal/validate", req, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// ApplyPAL applies a PAL spec to create or update an agent.
// POST /api/v1/pal/apply
func (c *Client) ApplyPAL(content, workspaceID string, dryRun bool) (*ApplyResult, error) {
	req := map[string]interface{}{
		"content":      content,
		"workspace_id": workspaceID,
		"dry_run":      dryRun,
	}

	var resp APIResponse[ApplyResult]
	if err := c.post("/api/v1/pal/apply", req, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// ExportPAL exports an agent configuration as PAL (YAML or JSON).
// GET /api/v1/workspaces/:workspaceId/pal/export/:agentId
func (c *Client) ExportPAL(workspaceID, agentID, format string) (*ExportResult, error) {
	path := fmt.Sprintf("/api/v1/workspaces/%s/pal/export/%s?format=%s", workspaceID, agentID, format)

	var resp APIResponse[ExportResult]
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// DiffPAL compares a PAL spec with the current agent configuration.
// POST /api/v1/pal/diff
func (c *Client) DiffPAL(workspaceID, agentID, content string) (*DiffResult, error) {
	req := map[string]string{
		"workspace_id": workspaceID,
		"agent_id":     agentID,
		"content":      content,
	}

	var resp APIResponse[DiffResult]
	if err := c.post("/api/v1/pal/diff", req, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
