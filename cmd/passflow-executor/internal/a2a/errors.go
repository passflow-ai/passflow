package a2a

import "errors"

// A2A validation errors
var (
	ErrEmptyMessageID   = errors.New("message_id cannot be empty")
	ErrEmptyFromAgent   = errors.New("from_agent cannot be empty")
	ErrEmptyToAgent     = errors.New("to_agent cannot be empty")
	ErrEmptyTimestamp   = errors.New("timestamp cannot be empty")
	ErrEmptyStatus      = errors.New("status cannot be empty")
	ErrInvalidStatus    = errors.New("status must be 'success' or 'error'")
	ErrMissingError     = errors.New("error field required when status is 'error'")
	ErrEmptyAgentID     = errors.New("agent_id cannot be empty")
	ErrEmptyName        = errors.New("name cannot be empty")
	ErrEmptyEndpoint    = errors.New("endpoint cannot be empty")
	ErrTranslationFailed = errors.New("failed to translate message")
)
