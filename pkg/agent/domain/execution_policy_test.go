package domain

import (
	"testing"

	workspaceDomain "github.com/passflow-ai/passflow/pkg/workspace/domain"
)

func TestExecutionPolicy_ResolveMode(t *testing.T) {
	policy := NewExecutionPolicy()

	tests := []struct {
		name          string
		workspace     *workspaceDomain.Workspace
		expectedMode  workspaceDomain.ExecutionMode
	}{
		{
			name: "workspace with explicit shared_pool mode",
			workspace: &workspaceDomain.Workspace{
				ExecutionMode: workspaceDomain.ExecutionModeSharedPool,
				Billing:       workspaceDomain.Billing{Plan: "enterprise"},
			},
			expectedMode: workspaceDomain.ExecutionModeSharedPool,
		},
		{
			name: "workspace with explicit isolated_pod mode",
			workspace: &workspaceDomain.Workspace{
				ExecutionMode: workspaceDomain.ExecutionModeIsolatedPod,
				Billing:       workspaceDomain.Billing{Plan: "free"},
			},
			expectedMode: workspaceDomain.ExecutionModeIsolatedPod,
		},
		{
			name: "enterprise plan defaults to isolated_pod",
			workspace: &workspaceDomain.Workspace{
				Billing: workspaceDomain.Billing{Plan: "enterprise"},
			},
			expectedMode: workspaceDomain.ExecutionModeIsolatedPod,
		},
		{
			name: "business plan defaults to isolated_pod",
			workspace: &workspaceDomain.Workspace{
				Billing: workspaceDomain.Billing{Plan: "business"},
			},
			expectedMode: workspaceDomain.ExecutionModeIsolatedPod,
		},
		{
			name: "pro plan defaults to shared_pool",
			workspace: &workspaceDomain.Workspace{
				Billing: workspaceDomain.Billing{Plan: "pro"},
			},
			expectedMode: workspaceDomain.ExecutionModeSharedPool,
		},
		{
			name: "free plan defaults to shared_pool",
			workspace: &workspaceDomain.Workspace{
				Billing: workspaceDomain.Billing{Plan: "free"},
			},
			expectedMode: workspaceDomain.ExecutionModeSharedPool,
		},
		{
			name: "empty plan defaults to shared_pool",
			workspace: &workspaceDomain.Workspace{
				Billing: workspaceDomain.Billing{Plan: ""},
			},
			expectedMode: workspaceDomain.ExecutionModeSharedPool,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := policy.ResolveMode(tt.workspace)
			if mode != tt.expectedMode {
				t.Errorf("ResolveMode() = %v, want %v", mode, tt.expectedMode)
			}
		})
	}
}

func TestExecutionPolicy_ResolveModeNilWorkspace(t *testing.T) {
	policy := NewExecutionPolicy()
	mode := policy.ResolveMode(nil)
	if mode != workspaceDomain.ExecutionModeSharedPool {
		t.Errorf("ResolveMode(nil) = %v, want %v", mode, workspaceDomain.ExecutionModeSharedPool)
	}
}

func TestExecutionPolicy_GetNamespace(t *testing.T) {
	policy := NewExecutionPolicy()

	tests := []struct {
		name              string
		workspaceID       string
		expectedNamespace string
	}{
		{
			name:              "standard workspace ID",
			workspaceID:       "abc123def456",
			expectedNamespace: "ws-abc123",
		},
		{
			name:              "short workspace ID",
			workspaceID:       "abc",
			expectedNamespace: "ws-abc",
		},
		{
			name:              "UUID workspace ID",
			workspaceID:       "550e8400-e29b-41d4-a716-446655440000",
			expectedNamespace: "ws-550e84",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := policy.GetNamespace(tt.workspaceID)
			if ns != tt.expectedNamespace {
				t.Errorf("GetNamespace(%q) = %q, want %q", tt.workspaceID, ns, tt.expectedNamespace)
			}
		})
	}
}
