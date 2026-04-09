package job

import "fmt"

// RedactCredentials returns a deep copy of the Spec with all credential values
// replaced by "[REDACTED]". The receiver is never modified. Use this copy
// whenever the spec needs to appear in logs or error messages.
func (s *Spec) RedactCredentials() Spec {
	copy := *s

	if len(s.Tools) == 0 {
		return copy
	}

	tools := make([]ToolConfig, len(s.Tools))
	for i, t := range s.Tools {
		tools[i] = t
		if t.IntegrationConfig == nil {
			continue
		}

		// Deep-copy the IntegrationToolConfig so the original is untouched.
		icCopy := *t.IntegrationConfig
		if len(t.IntegrationConfig.Credentials) > 0 {
			redactedCreds := make(map[string]string, len(t.IntegrationConfig.Credentials))
			for k := range t.IntegrationConfig.Credentials {
				redactedCreds[k] = "[REDACTED]"
			}
			icCopy.Credentials = redactedCreds
		}
		tools[i].IntegrationConfig = &icCopy
	}

	copy.Tools = tools
	return copy
}

// SensitiveFields returns a slice of dot-notation field paths that contain
// sensitive data (credentials) in this spec. Callers can use the list to
// document which fields were scrubbed or to drive structured redaction logic.
func (s *Spec) SensitiveFields() []string {
	var paths []string
	for i, t := range s.Tools {
		if t.IntegrationConfig == nil {
			continue
		}
		for k := range t.IntegrationConfig.Credentials {
			paths = append(paths, fmt.Sprintf("tools[%d].integration_config.credentials.%s", i, k))
		}
	}
	return paths
}
