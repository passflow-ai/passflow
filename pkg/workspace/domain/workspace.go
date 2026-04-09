package domain

// ExecutionMode represents the agent execution mode for a workspace.
type ExecutionMode string

const (
	// ExecutionModeSharedPool uses a shared agent pool.
	ExecutionModeSharedPool ExecutionMode = "shared_pool"
	// ExecutionModeIsolatedPod uses isolated Kubernetes pods.
	ExecutionModeIsolatedPod ExecutionMode = "isolated_pod"
)

// Billing contains workspace billing information.
type Billing struct {
	Plan string `bson:"plan" json:"plan"`
}

// Workspace represents a workspace entity.
type Workspace struct {
	ID            string        `bson:"_id" json:"id"`
	Name          string        `bson:"name" json:"name"`
	ExecutionMode ExecutionMode `bson:"executionMode,omitempty" json:"executionMode,omitempty"`
	Billing       Billing   `bson:"billing" json:"billing"`
}
