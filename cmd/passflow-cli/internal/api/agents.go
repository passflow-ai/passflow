package api

import "fmt"

type Agent struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Model       string   `json:"model,omitempty"`
	Tools       []string `json:"tools,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

func (c *Client) ListAgents(workspaceID string) ([]Agent, error) {
	path := fmt.Sprintf("/api/v1/workspaces/%s/agents", workspaceID)

	var resp PaginatedResponse[Agent]
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) GetAgent(workspaceID, agentID string) (*Agent, error) {
	path := fmt.Sprintf("/api/v1/workspaces/%s/agents/%s", workspaceID, agentID)

	var resp APIResponse[Agent]
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
