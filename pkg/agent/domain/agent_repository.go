package domain

import (
	"context"
	"time"
)

// AgentRepository defines the interface for agent persistence operations.
type AgentRepository interface {
	// Create creates a new agent and returns its ID.
	Create(ctx context.Context, agent *Agent) (string, error)

	// FindByID finds an agent by its ID and workspace ID.
	FindByID(ctx context.Context, workspaceID, agentID string) (*Agent, error)

	// FindByWorkspace finds all agents in a workspace with pagination.
	FindByWorkspace(ctx context.Context, workspaceID string, opts PaginationOptions) (*AgentPage, error)

	// Update updates an existing agent.
	Update(ctx context.Context, agent *Agent) error

	// Delete deletes an agent by its ID.
	Delete(ctx context.Context, workspaceID, agentID string) error

	// UpdateStatus updates only the status of an agent.
	UpdateStatus(ctx context.Context, workspaceID, agentID, status string) error

	// CountByWorkspace returns the total count of agents in a workspace.
	CountByWorkspace(ctx context.Context, workspaceID string) (int64, error)

	// FindByStatus finds all agents with a specific status.
	FindByStatus(ctx context.Context, status string) ([]*Agent, error)

	// ExistsByNameInWorkspace checks if an agent with the given name exists in the workspace.
	ExistsByNameInWorkspace(ctx context.Context, workspaceID, name string) (bool, error)

	// FindWithHeartbeatEnabled returns all agents with heartbeat enabled across all workspaces.
	FindWithHeartbeatEnabled(ctx context.Context) ([]*Agent, error)

	// UpdateLastRun updates the last run timestamp of an agent.
	UpdateLastRun(ctx context.Context, workspaceID, agentID string, t time.Time) error

	// DeleteByWorkspaceID deletes all agents in a workspace (for cascade delete).
	DeleteByWorkspaceID(ctx context.Context, workspaceID string) (int64, error)
}

// PaginationOptions represents pagination parameters.
type PaginationOptions struct {
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	SortBy  string `json:"sortBy"`
	SortDir string `json:"sortDir"`
	Search  string `json:"search"`
}

// AgentPage represents a paginated list of agents.
type AgentPage struct {
	Agents     []*Agent `json:"agents"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PerPage    int      `json:"perPage"`
	TotalPages int      `json:"totalPages"`
}

// NewPaginationOptions creates pagination options with defaults.
func NewPaginationOptions(page, perPage int, sortBy, sortDir, search string) PaginationOptions {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortDir == "" {
		sortDir = "desc"
	}

	return PaginationOptions{
		Page:    page,
		PerPage: perPage,
		SortBy:  sortBy,
		SortDir: sortDir,
		Search:  search,
	}
}

// Offset returns the offset for pagination.
func (p PaginationOptions) Offset() int64 {
	return int64((p.Page - 1) * p.PerPage)
}

// CalculateTotalPages calculates the total number of pages.
func CalculateTotalPages(total int64, perPage int) int {
	if total == 0 {
		return 0
	}
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}
