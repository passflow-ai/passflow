package palres

// ExportPALResponse represents the response for exporting PAL content.
type ExportPALResponse struct {
	AgentID   string      `json:"agentId"`
	AgentName string      `json:"agentName"`
	Format    string      `json:"format"` // json, yaml
	Content   string      `json:"content"`
	Timestamp string      `json:"timestamp"`
}

// NewExportPALResponse creates a new ExportPALResponse.
func NewExportPALResponse(agentID, agentName, format, content, timestamp string) *ExportPALResponse {
	return &ExportPALResponse{
		AgentID:   agentID,
		AgentName: agentName,
		Format:    format,
		Content:   content,
		Timestamp: timestamp,
	}
}

// GetFormat returns the export format.
func (r *ExportPALResponse) GetFormat() string {
	if r.Format == "" {
		return "json"
	}
	return r.Format
}

// IsYAML returns whether the format is YAML.
func (r *ExportPALResponse) IsYAML() bool {
	return r.Format == "yaml"
}

// IsJSON returns whether the format is JSON.
func (r *ExportPALResponse) IsJSON() bool {
	return r.Format == "json"
}
