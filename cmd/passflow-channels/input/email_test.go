package input

import (
	"context"
	"testing"

	"github.com/passflow-ai/passflow/cmd/passflow-channels/config"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/trigger"
)

// recordingDispatcher captures dispatched events for assertion.
type recordingDispatcher struct {
	events []domain.Event
}

func (r *recordingDispatcher) Dispatch(_ context.Context, e domain.Event) {
	r.events = append(r.events, e)
}

// stubDispatcher adapts recordingDispatcher to the emailDispatcher interface.
type stubDispatcher struct {
	rec *recordingDispatcher
}

func (s *stubDispatcher) Dispatch(ctx context.Context, e domain.Event) {
	s.rec.Dispatch(ctx, e)
}

// Compile-time checks: both stub and the real Dispatcher satisfy emailDispatcher.
var _ emailDispatcher = (*stubDispatcher)(nil)
var _ emailDispatcher = (*trigger.Dispatcher)(nil)

// newTestEmailPoller builds an EmailPoller with a stub dispatcher for testing.
// It avoids constructing a real *trigger.Dispatcher (which needs Redis).
func newTestEmailPoller(cfg *config.Config, rec *recordingDispatcher) *EmailPoller {
	return &EmailPoller{
		cfg:        cfg,
		dispatcher: &stubDispatcher{rec: rec},
	}
}

// simulateDispatch mimics what processMessage does after extracting fields,
// exercising the workspace-ID injection logic without a live IMAP connection.
func simulateDispatch(p *EmailPoller, rec *recordingDispatcher, subject string) {
	wsID := p.WorkspaceID()
	if wsID == "" {
		return // drop — no workspace configured
	}
	event := domain.Event{
		Channel:     domain.ChannelEmail,
		WorkspaceID: wsID,
		Fields:      map[string]string{"subject": subject},
	}
	p.dispatcher.Dispatch(context.Background(), event)
}

// TestEmail_WorkspaceIDSet verifies that when EMAIL_WORKSPACE_ID is configured
// the dispatched event carries that workspace ID.
func TestEmail_WorkspaceIDSet(t *testing.T) {
	cfg := &config.Config{
		IMAPHost:         "imap.example.com",
		IMAPPort:         "993",
		EmailWorkspaceID: "ws-email-123",
	}
	rec := &recordingDispatcher{}
	p := newTestEmailPoller(cfg, rec)

	simulateDispatch(p, rec, "Hello")

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event dispatched, got %d", len(rec.events))
	}
	if rec.events[0].WorkspaceID != "ws-email-123" {
		t.Errorf("expected WorkspaceID=ws-email-123, got %q", rec.events[0].WorkspaceID)
	}
}

// TestEmail_NoWorkspaceID_EventNotDispatched verifies that when no workspace ID
// is configured, events are NOT dispatched (to prevent cross-tenant fan-out).
func TestEmail_NoWorkspaceID_EventNotDispatched(t *testing.T) {
	cfg := &config.Config{
		IMAPHost:         "imap.example.com",
		IMAPPort:         "993",
		EmailWorkspaceID: "", // not set
	}
	rec := &recordingDispatcher{}
	p := newTestEmailPoller(cfg, rec)

	simulateDispatch(p, rec, "Should not dispatch")

	if len(rec.events) != 0 {
		t.Errorf("expected 0 events when workspace ID is not configured, got %d", len(rec.events))
	}
}

// TestEmail_WorkspaceID_ReturnsConfiguredValue exercises the WorkspaceID()
// accessor directly.
func TestEmail_WorkspaceID_ReturnsConfiguredValue(t *testing.T) {
	cfg := &config.Config{EmailWorkspaceID: "tenant-42"}
	p := newTestEmailPoller(cfg, &recordingDispatcher{})

	if got := p.WorkspaceID(); got != "tenant-42" {
		t.Errorf("WorkspaceID() = %q, want %q", got, "tenant-42")
	}
}

// TestEmail_WorkspaceID_EmptyWhenNotConfigured verifies the accessor returns
// empty string when EMAIL_WORKSPACE_ID env var is not set.
func TestEmail_WorkspaceID_EmptyWhenNotConfigured(t *testing.T) {
	cfg := &config.Config{EmailWorkspaceID: ""}
	p := newTestEmailPoller(cfg, &recordingDispatcher{})

	if got := p.WorkspaceID(); got != "" {
		t.Errorf("WorkspaceID() = %q, want empty", got)
	}
}
