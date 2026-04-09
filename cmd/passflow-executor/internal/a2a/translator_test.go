package a2a

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslator_TranslateA2AToEvent(t *testing.T) {
	translator := NewTranslator()

	tests := []struct {
		name        string
		msg         Message
		workspaceID string
		wantErr     bool
		validate    func(t *testing.T, event *Event)
	}{
		{
			name: "valid message with all fields",
			msg: Message{
				MessageID:  "msg-123",
				FromAgent:  "external-crm",
				ToAgent:    "passflow-agent-sales",
				Capability: "qualify_lead",
				Payload: map[string]interface{}{
					"lead_email": "prospect@example.com",
					"lead_name":  "Jane Doe",
				},
				Context: &MessageContext{
					ConversationID: "conv-456",
					CorrelationID:  "corr-789",
				},
				Timestamp:   time.Now().UTC(),
				CallbackURL: "https://external-crm.com/callback",
			},
			workspaceID: "ws-123",
			wantErr:     false,
			validate: func(t *testing.T, event *Event) {
				assert.Equal(t, EventTriggerAgent, event.Type)
				assert.Equal(t, "external-crm", event.SourceAgent)
				assert.Equal(t, "passflow-agent-sales", event.TargetAgent)
				assert.Equal(t, "ws-123", event.WorkspaceID)
				assert.Equal(t, "msg-123", event.Payload["a2a_message_id"])
				assert.Equal(t, "qualify_lead", event.Payload["a2a_capability"])
				assert.Equal(t, "https://external-crm.com/callback", event.Payload["a2a_callback_url"])
				assert.Equal(t, "conv-456", event.Payload["a2a_conversation_id"])
				assert.Equal(t, "corr-789", event.Payload["a2a_correlation_id"])

				input, ok := event.Payload["input"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "prospect@example.com", input["lead_email"])
				assert.Equal(t, "Jane Doe", input["lead_name"])
			},
		},
		{
			name: "minimal valid message",
			msg: Message{
				MessageID: "msg-min",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Payload:   map[string]interface{}{},
				Timestamp: time.Now().UTC(),
			},
			workspaceID: "ws-456",
			wantErr:     false,
			validate: func(t *testing.T, event *Event) {
				assert.Equal(t, EventTriggerAgent, event.Type)
				assert.Equal(t, "agent-a", event.SourceAgent)
				assert.Equal(t, "agent-b", event.TargetAgent)
				assert.NotEmpty(t, event.ID)
			},
		},
		{
			name: "missing message_id",
			msg: Message{
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Timestamp: time.Now().UTC(),
			},
			workspaceID: "ws-123",
			wantErr:     true,
		},
		{
			name: "missing from_agent",
			msg: Message{
				MessageID: "msg-123",
				ToAgent:   "agent-b",
				Timestamp: time.Now().UTC(),
			},
			workspaceID: "ws-123",
			wantErr:     true,
		},
		{
			name: "missing to_agent",
			msg: Message{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				Timestamp: time.Now().UTC(),
			},
			workspaceID: "ws-123",
			wantErr:     true,
		},
		{
			name: "missing timestamp",
			msg: Message{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
			},
			workspaceID: "ws-123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := translator.TranslateA2AToEvent(tt.msg, tt.workspaceID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, event)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, event)

			if tt.validate != nil {
				tt.validate(t, event)
			}
		})
	}
}

func TestTranslator_TranslateEventToA2A(t *testing.T) {
	translator := NewTranslator()

	tests := []struct {
		name     string
		event    Event
		validate func(t *testing.T, resp *Response)
	}{
		{
			name: "successful event",
			event: Event{
				ID:          "event-123",
				Type:        EventAgentCompleted,
				SourceAgent: "passflow-agent-sales",
				TargetAgent: "external-crm",
				WorkspaceID: "ws-123",
				Payload: map[string]interface{}{
					"a2a_message_id":      "msg-123",
					"a2a_conversation_id": "conv-456",
					"a2a_correlation_id":  "corr-789",
					"result": map[string]interface{}{
						"fit_score":    85,
						"assigned_rep": "john@example.com",
					},
				},
				Timestamp: time.Now().UTC(),
			},
			validate: func(t *testing.T, resp *Response) {
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, "msg-123", resp.InReplyTo)
				assert.Equal(t, "passflow-agent-sales", resp.FromAgent)
				assert.Equal(t, "external-crm", resp.ToAgent)
				assert.NotEmpty(t, resp.MessageID)
				assert.NotNil(t, resp.Context)
				assert.Equal(t, "conv-456", resp.Context.ConversationID)
				assert.Equal(t, "corr-789", resp.Context.CorrelationID)
				assert.EqualValues(t, 85, resp.Payload["fit_score"])
				assert.Equal(t, "john@example.com", resp.Payload["assigned_rep"])
				assert.Nil(t, resp.Error)
			},
		},
		{
			name: "failed event",
			event: Event{
				ID:          "event-failed",
				Type:        EventAgentFailed,
				SourceAgent: "passflow-agent-sales",
				TargetAgent: "external-crm",
				WorkspaceID: "ws-123",
				Payload: map[string]interface{}{
					"a2a_message_id":      "msg-456",
					"a2a_conversation_id": "conv-789",
					"error":               "lead_email is required",
					"error_details": map[string]interface{}{
						"field": "lead_email",
					},
				},
				Timestamp: time.Now().UTC(),
			},
			validate: func(t *testing.T, resp *Response) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, "msg-456", resp.InReplyTo)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, "AGENT_EXECUTION_FAILED", resp.Error.Code)
				assert.Equal(t, "lead_email is required", resp.Error.Message)
				assert.Equal(t, "lead_email", resp.Error.Details["field"])
			},
		},
		{
			name: "event without A2A metadata",
			event: Event{
				ID:          "event-no-metadata",
				Type:        EventAgentCompleted,
				SourceAgent: "agent-a",
				TargetAgent: "agent-b",
				WorkspaceID: "ws-123",
				Payload: map[string]interface{}{
					"result": map[string]interface{}{
						"status": "ok",
					},
				},
				Timestamp: time.Now().UTC(),
			},
			validate: func(t *testing.T, resp *Response) {
				assert.Equal(t, "success", resp.Status)
				assert.Empty(t, resp.InReplyTo)
				assert.Nil(t, resp.Context)
				assert.Equal(t, "ok", resp.Payload["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := translator.TranslateEventToA2A(tt.event)

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestTranslator_BuildAgentCard(t *testing.T) {
	translator := NewTranslator()

	capabilities := []Capability{
		{
			CapabilityID: "qualify_lead",
			Description:  "Analyzes lead and determines fit score",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"lead_email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
				},
			},
			OutputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fit_score": map[string]interface{}{
						"type": "number",
					},
				},
			},
		},
	}

	authConfig := &AuthConfig{
		Type:          "bearer",
		TokenEndpoint: "http://localhost:3000/api/v1/auth/a2a-token",
	}

	card := translator.BuildAgentCard(
		"passflow-agent-sales",
		"Sales Lead Qualifier",
		"Qualifies incoming sales leads",
		"1.0.0",
		capabilities,
		"http://localhost:3000/api/v1/a2a/messages",
		authConfig,
	)

	assert.Equal(t, "passflow-agent-sales", card.AgentID)
	assert.Equal(t, "Sales Lead Qualifier", card.Name)
	assert.Equal(t, "Qualifies incoming sales leads", card.Description)
	assert.Equal(t, "1.0.0", card.Version)
	assert.Len(t, card.Capabilities, 1)
	assert.Equal(t, "qualify_lead", card.Capabilities[0].CapabilityID)
	assert.Equal(t, "http://localhost:3000/api/v1/a2a/messages", card.Endpoint)
	assert.Equal(t, "bearer", card.Auth.Type)
}

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		wantErr error
	}{
		{
			name: "valid message",
			msg: Message{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Timestamp: time.Now().UTC(),
			},
			wantErr: nil,
		},
		{
			name: "missing message_id",
			msg: Message{
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Timestamp: time.Now().UTC(),
			},
			wantErr: ErrEmptyMessageID,
		},
		{
			name: "missing from_agent",
			msg: Message{
				MessageID: "msg-123",
				ToAgent:   "agent-b",
				Timestamp: time.Now().UTC(),
			},
			wantErr: ErrEmptyFromAgent,
		},
		{
			name: "missing to_agent",
			msg: Message{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				Timestamp: time.Now().UTC(),
			},
			wantErr: ErrEmptyToAgent,
		},
		{
			name: "missing timestamp",
			msg: Message{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
			},
			wantErr: ErrEmptyTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResponse_Validate(t *testing.T) {
	tests := []struct {
		name    string
		resp    Response
		wantErr error
	}{
		{
			name: "valid success response",
			resp: Response{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Status:    "success",
			},
			wantErr: nil,
		},
		{
			name: "valid error response",
			resp: Response{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Status:    "error",
				Error: &ErrorInfo{
					Code:    "FAILED",
					Message: "Something went wrong",
				},
			},
			wantErr: nil,
		},
		{
			name: "missing message_id",
			resp: Response{
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Status:    "success",
			},
			wantErr: ErrEmptyMessageID,
		},
		{
			name: "missing status",
			resp: Response{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
			},
			wantErr: ErrEmptyStatus,
		},
		{
			name: "invalid status",
			resp: Response{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Status:    "unknown",
			},
			wantErr: ErrInvalidStatus,
		},
		{
			name: "error status without error field",
			resp: Response{
				MessageID: "msg-123",
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Status:    "error",
			},
			wantErr: ErrMissingError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgentCard_Validate(t *testing.T) {
	tests := []struct {
		name    string
		card    AgentCard
		wantErr error
	}{
		{
			name: "valid agent card",
			card: AgentCard{
				AgentID:  "agent-123",
				Name:     "Test Agent",
				Endpoint: "https://example.com/a2a",
			},
			wantErr: nil,
		},
		{
			name: "missing agent_id",
			card: AgentCard{
				Name:     "Test Agent",
				Endpoint: "https://example.com/a2a",
			},
			wantErr: ErrEmptyAgentID,
		},
		{
			name: "missing name",
			card: AgentCard{
				AgentID:  "agent-123",
				Endpoint: "https://example.com/a2a",
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "missing endpoint",
			card: AgentCard{
				AgentID: "agent-123",
				Name:    "Test Agent",
			},
			wantErr: ErrEmptyEndpoint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
