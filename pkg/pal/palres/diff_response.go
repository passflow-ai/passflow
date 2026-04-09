package palres

// DiffChange represents a single difference between current and proposed PAL content.
type DiffChange struct {
	Path       string      `json:"path"`
	Field      string      `json:"field"`
	ChangeType string      `json:"type"` // added, removed, modified
	Current    interface{} `json:"current,omitempty"`
	Proposed   interface{} `json:"proposed,omitempty"`
}

// DiffPALResponse represents the response for diffing PAL content.
type DiffPALResponse struct {
	AgentID    string       `json:"agentId"`
	HasChanges bool         `json:"hasChanges"`
	Changes    []DiffChange `json:"changes,omitempty"`
	Summary    DiffSummary  `json:"summary"`
}

// DiffSummary represents a summary of the differences.
type DiffSummary struct {
	Added    int `json:"added"`
	Modified int `json:"modified"`
	Removed  int `json:"removed"`
	Total    int `json:"total"`
}

// NewDiffPALResponse creates a new DiffPALResponse.
func NewDiffPALResponse(agentID string) *DiffPALResponse {
	return &DiffPALResponse{
		AgentID:    agentID,
		HasChanges: false,
		Changes:    make([]DiffChange, 0),
		Summary: DiffSummary{
			Added:    0,
			Modified: 0,
			Removed:  0,
			Total:    0,
		},
	}
}

// AddChange adds a change to the diff response.
func (r *DiffPALResponse) AddChange(path, field, changeType string, current, proposed interface{}) {
	r.Changes = append(r.Changes, DiffChange{
		Path:       path,
		Field:      field,
		ChangeType: changeType,
		Current:    current,
		Proposed:   proposed,
	})
	r.HasChanges = true

	// Update summary
	switch changeType {
	case "added":
		r.Summary.Added++
	case "removed":
		r.Summary.Removed++
	case "modified":
		r.Summary.Modified++
	}
	r.Summary.Total++
}

// GetChangeCount returns the total number of changes.
func (r *DiffPALResponse) GetChangeCount() int {
	return len(r.Changes)
}

// GetAddedCount returns the number of added fields.
func (r *DiffPALResponse) GetAddedCount() int {
	return r.Summary.Added
}

// GetModifiedCount returns the number of modified fields.
func (r *DiffPALResponse) GetModifiedCount() int {
	return r.Summary.Modified
}

// GetRemovedCount returns the number of removed fields.
func (r *DiffPALResponse) GetRemovedCount() int {
	return r.Summary.Removed
}
