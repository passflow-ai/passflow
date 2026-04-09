package domain

import (
	"fmt"

	workspaceDomain "github.com/passflow-ai/passflow/pkg/workspace/domain"
)

// ExecutionPolicy determines the execution mode for an agent based on workspace configuration.
type ExecutionPolicy struct{}

// NewExecutionPolicy creates a new ExecutionPolicy.
func NewExecutionPolicy() *ExecutionPolicy {
	return &ExecutionPolicy{}
}

// ResolveMode determines the execution mode for an agent in the given workspace.
// Priority:
// 1. Workspace's explicit ExecutionMode override
// 2. Billing plan defaults (enterprise/business → isolated_pod, others → shared_pool)
func (p *ExecutionPolicy) ResolveMode(workspace *workspaceDomain.Workspace) workspaceDomain.ExecutionMode {
	if workspace == nil {
		return workspaceDomain.ExecutionModeSharedPool
	}

	// 1. Workspace override takes priority
	if workspace.ExecutionMode != "" {
		return workspace.ExecutionMode
	}

	// 2. Determine by billing plan
	switch workspace.Billing.Plan {
	case "enterprise", "business":
		return workspaceDomain.ExecutionModeIsolatedPod
	default:
		return workspaceDomain.ExecutionModeSharedPool
	}
}

// GetNamespace returns the Kubernetes namespace for a workspace.
// Format: ws-{first 6 chars of workspaceID}
func (p *ExecutionPolicy) GetNamespace(workspaceID string) string {
	prefix := workspaceID
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	return fmt.Sprintf("ws-%s", prefix)
}

// ShouldUseIsolatedPod returns true if the workspace should use isolated pod execution.
func (p *ExecutionPolicy) ShouldUseIsolatedPod(workspace *workspaceDomain.Workspace) bool {
	return p.ResolveMode(workspace) == workspaceDomain.ExecutionModeIsolatedPod
}
