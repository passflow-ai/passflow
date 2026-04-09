package input

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/config"
	"github.com/jaak-ai/passflow-channels/domain"
)

// emailDispatcher is the narrow interface EmailPoller needs from the Dispatcher.
// Defined here so tests can inject a stub without importing trigger (which needs Redis).
type emailDispatcher interface {
	Dispatch(ctx context.Context, event domain.Event)
}

// processedSetCapacity is the maximum number of message IDs held in memory as
// a local deduplication guard. 1000 entries covers typical email volumes while
// keeping the memory footprint small.
const processedSetCapacity = 1000

// EmailPoller polls an IMAP mailbox for new messages and dispatches events.
type EmailPoller struct {
	cfg          *config.Config
	dispatcher   emailDispatcher
	processedIDs *processedSet
}

// NewEmailPoller creates a new EmailPoller.
// It logs an error at construction time when EMAIL_WORKSPACE_ID is empty, so
// the misconfiguration is visible in startup logs.
func NewEmailPoller(cfg *config.Config, dispatcher emailDispatcher) *EmailPoller {
	if cfg.EmailWorkspaceID == "" {
		log.Println("[email] ERROR: EMAIL_WORKSPACE_ID is not set — email events will be dropped to prevent cross-tenant fan-out")
	}
	return &EmailPoller{
		cfg:          cfg,
		dispatcher:   dispatcher,
		processedIDs: newProcessedSet(processedSetCapacity),
	}
}

// WorkspaceID returns the workspace ID this poller is associated with.
func (p *EmailPoller) WorkspaceID() string {
	return p.cfg.EmailWorkspaceID
}

// Start begins polling in a background goroutine.
func (p *EmailPoller) Start(ctx context.Context) {
	if p.cfg.IMAPHost == "" {
		log.Println("[email] IMAP not configured — skipping email polling")
		return
	}

	interval := time.Duration(p.cfg.IMAPPollSec) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	log.Printf("[email] Starting IMAP poller (host: %s, mailbox: %s, interval: %s)",
		p.cfg.IMAPHost, p.cfg.IMAPMailbox, interval)

	go func() {
		p.poll(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[email] IMAP poller stopped")
				return
			case <-ticker.C:
				p.poll(ctx)
			}
		}
	}()
}

// poll connects to IMAP, fetches unseen messages, and dispatches events.
func (p *EmailPoller) poll(ctx context.Context) {
	addr := fmt.Sprintf("%s:%s", p.cfg.IMAPHost, p.cfg.IMAPPort)

	var c *imapclient.Client
	var err error

	if p.cfg.IMAPPort == "993" {
		c, err = imapclient.DialTLS(addr, nil)
	} else {
		c, err = imapclient.DialStartTLS(addr, nil)
	}
	if err != nil {
		log.Printf("[email] failed to connect to %s: %v", addr, err)
		return
	}
	defer c.Close()

	if err := c.Login(p.cfg.IMAPUser, p.cfg.IMAPPassword).Wait(); err != nil {
		log.Printf("[email] login failed: %v", err)
		return
	}

	if _, err := c.Select(p.cfg.IMAPMailbox, nil).Wait(); err != nil {
		log.Printf("[email] failed to select mailbox %s: %v", p.cfg.IMAPMailbox, err)
		return
	}

	// Search for unseen messages
	searchData, err := c.Search(&imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}, nil).Wait()
	if err != nil {
		log.Printf("[email] search failed: %v", err)
		return
	}

	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		return
	}

	log.Printf("[email] found %d unseen message(s)", len(seqNums))

	seqSet := imap.SeqSetNum(seqNums...)
	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierText},
		},
	}

	messages, err := c.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		log.Printf("[email] fetch failed: %v", err)
		return
	}

	for _, msg := range messages {
		p.processMessage(ctx, c, msg)
	}
}

// processMessage extracts fields and dispatches an event.
// If EMAIL_WORKSPACE_ID is not configured, the event is dropped to prevent
// dispatching against all workspaces (cross-tenant fan-out).
func (p *EmailPoller) processMessage(ctx context.Context, c *imapclient.Client, msg *imapclient.FetchMessageBuffer) {
	workspaceID := p.cfg.EmailWorkspaceID
	if workspaceID == "" {
		log.Printf("[email] skipping message dispatch: EMAIL_WORKSPACE_ID is not configured")
		return
	}

	fields := make(map[string]string)

	if msg.Envelope != nil {
		env := msg.Envelope
		fields["subject"] = env.Subject

		if len(env.From) > 0 {
			from := env.From[0]
			addr := from.Mailbox + "@" + from.Host
			fields["from"] = addr
			if from.Name != "" {
				fields["from"] = fmt.Sprintf("%s <%s>", from.Name, addr)
			}
			fields["from_email"] = addr
		}

		if !env.Date.IsZero() {
			fields["date"] = env.Date.Format(time.RFC3339)
		}
		fields["message_id"] = env.MessageID
	}

	// Extract text body
	for _, section := range msg.BodySection {
		if len(section.Bytes) > 0 {
			fields["body"] = string(section.Bytes)
			break
		}
	}

	messageID := fields["message_id"]

	// Guard against duplicate dispatch when a previous Store call failed.
	// The processedSet is an in-process deduplication layer; the primary
	// deduplication mechanism remains the \Seen flag on the IMAP server.
	if messageID != "" && p.processedIDs.contains(messageID) {
		log.Printf("[email] skipping already-processed message %q", messageID)
		return
	}

	event := domain.Event{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Channel:     domain.ChannelEmail,
		Fields:      fields,
		Raw: map[string]interface{}{
			"subject": fields["subject"],
			"from":    fields["from"],
		},
		ReceivedAt: time.Now(),
	}

	log.Printf("[email] dispatching event for message from %s: %q", fields["from_email"], fields["subject"])
	p.dispatcher.Dispatch(ctx, event)

	// Mark as seen on the IMAP server. Check the return value so that failures
	// are visible in logs rather than silently causing duplicate dispatches on
	// the next poll cycle.
	seqSet := imap.SeqSetNum(msg.SeqNum)
	if _, err := c.Store(seqSet, &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}, nil).Collect(); err != nil {
		log.Printf("[email] WARNING: failed to mark message %q as seen — adding to local dedup set to prevent duplicate dispatch: %v",
			messageID, err)
		// Fall through to record it in the local set regardless of server failure.
	}

	// Record in the local processed set so that if the server-side \Seen flag
	// was not applied (e.g. network error), we still suppress duplicates within
	// this process's lifetime.
	p.processedIDs.add(messageID)
}
