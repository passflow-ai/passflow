package palres

// ApplyChange represents a single change applied during PAL application.
type ApplyChange struct {
	Type      string      `json:"type"`                  // created, updated, deleted
	Resource  string      `json:"resource"`              // agent, trigger, guardrail, etc.
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	Details   interface{} `json:"details,omitempty"`
}

// ApplyPALResponse represents the response for applying PAL content.
type ApplyPALResponse struct {
	Success   bool          `json:"success"`
	AgentID   string        `json:"agentId,omitempty"`
	Changes   []ApplyChange `json:"changes,omitempty"`
	DryRun    bool          `json:"dryRun"`
	Message   string        `json:"message,omitempty"`
	Timestamp string        `json:"timestamp"`
}

// NewApplyPALResponse creates a new ApplyPALResponse.
func NewApplyPALResponse(dryRun bool) *ApplyPALResponse {
	return &ApplyPALResponse{
		DryRun:  dryRun,
		Changes: make([]ApplyChange, 0),
	}
}

// AddChange adds a change to the apply response.
func (r *ApplyPALResponse) AddChange(changeType, resource, id, name string, details interface{}) {
	r.Changes = append(r.Changes, ApplyChange{
		Type:     changeType,
		Resource: resource,
		ID:       id,
		Name:     name,
		Details:  details,
	})
}

// SetAgentID sets the agent ID in the response.
func (r *ApplyPALResponse) SetAgentID(agentID string) {
	r.AgentID = agentID
}

// SetSuccess sets the success status.
func (r *ApplyPALResponse) SetSuccess(success bool) {
	r.Success = success
}

// SetMessage sets the message.
func (r *ApplyPALResponse) SetMessage(message string) {
	r.Message = message
}

// GetChangeCount returns the total number of changes.
func (r *ApplyPALResponse) GetChangeCount() int {
	return len(r.Changes)
}
