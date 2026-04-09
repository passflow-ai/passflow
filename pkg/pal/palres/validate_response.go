package palres

// ValidationIssue represents a validation error or warning.
type ValidationIssue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidatePALResponse represents the response for PAL validation.
type ValidatePALResponse struct {
	Valid    bool               `json:"valid"`
	Warnings []ValidationIssue  `json:"warnings,omitempty"`
	Errors   []ValidationIssue  `json:"errors,omitempty"`
}

// NewValidatePALResponse creates a new ValidatePALResponse.
func NewValidatePALResponse(valid bool) *ValidatePALResponse {
	return &ValidatePALResponse{
		Valid:    valid,
		Warnings: make([]ValidationIssue, 0),
		Errors:   make([]ValidationIssue, 0),
	}
}

// AddError adds an error to the validation response.
func (r *ValidatePALResponse) AddError(path, message, code string) {
	r.Errors = append(r.Errors, ValidationIssue{
		Path:    path,
		Message: message,
		Code:    code,
	})
	r.Valid = false
}

// AddWarning adds a warning to the validation response.
func (r *ValidatePALResponse) AddWarning(path, message, code string) {
	r.Warnings = append(r.Warnings, ValidationIssue{
		Path:    path,
		Message: message,
		Code:    code,
	})
}

// HasErrors returns whether the response contains any errors.
func (r *ValidatePALResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns whether the response contains any warnings.
func (r *ValidatePALResponse) HasWarnings() bool {
	return len(r.Warnings) > 0
}
