package input

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/trigger"
	cronlib "github.com/robfig/cron/v3"
)

// CronRunner manages cron-based triggers.
// Each TriggerRule with ChannelType==cron and a "schedule" field in its
// Condition.Value (standard cron expression) gets a scheduled job.
type CronRunner struct {
	cron       *cronlib.Cron
	dispatcher *trigger.Dispatcher
	store      TriggerRuleStore

	mu          sync.Mutex
	entryMap    map[string]cronlib.EntryID // ruleID → cron entry
	entryKeyMap map[string]string          // ruleID → entryKey (for diff detection)
}

// TriggerRuleStore is the subset of the rule store the CronRunner needs.
type TriggerRuleStore interface {
	GetAllRules() []domain.TriggerRule
}

// NewCronRunner creates a new CronRunner.
func NewCronRunner(dispatcher *trigger.Dispatcher, store TriggerRuleStore) *CronRunner {
	return &CronRunner{
		cron:        cronlib.New(cronlib.WithSeconds()),
		dispatcher:  dispatcher,
		store:       store,
		entryMap:    make(map[string]cronlib.EntryID),
		entryKeyMap: make(map[string]string),
	}
}

// Start loads cron rules and starts the scheduler.
func (r *CronRunner) Start(ctx context.Context) {
	r.reloadRules()
	r.cron.Start()
	log.Printf("[cron] scheduler started with %d entries", len(r.entryMap))

	// Reload rules periodically to pick up new/changed cron rules
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				r.cron.Stop()
				return
			case <-ticker.C:
				r.reloadRules()
			}
		}
	}()
}

// cronEntryKey returns a string that uniquely represents the observable state of
// a cron rule. When any key field changes, the entry must be re-registered.
func cronEntryKey(rule domain.TriggerRule) string {
	return fmt.Sprintf("%s|%s|%v|%s", rule.ID, rule.Condition.Value, rule.Enabled, rule.WorkspaceID)
}

// reloadRules performs a diff-based reload of cron jobs from the rule store.
// Only added/changed/removed rules are touched; unchanged entries are left in
// place so no jobs are missed or fired twice during the reload window.
func (r *CronRunner) reloadRules() {
	rules := r.store.GetAllRules()

	// Build a lookup of desired active rules keyed by rule ID.
	desired := make(map[string]domain.TriggerRule)
	for _, rule := range rules {
		if rule.ChannelType != domain.ChannelCron || !rule.Enabled {
			continue
		}
		if rule.Condition.Value == "" {
			continue
		}
		desired[rule.ID] = rule
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove entries that are no longer desired or whose key changed.
	for id, entryID := range r.entryMap {
		rule, stillWanted := desired[id]
		if !stillWanted || cronEntryKey(rule) != r.entryKeyMap[id] {
			r.cron.Remove(entryID)
			delete(r.entryMap, id)
			delete(r.entryKeyMap, id)
		}
	}

	// Add entries for rules not yet registered.
	for id, rule := range desired {
		if _, exists := r.entryMap[id]; exists {
			// Already registered and key matches (checked above) — skip.
			continue
		}

		capturedRule := rule
		schedule := rule.Condition.Value
		entryID, err := r.cron.AddFunc(schedule, func() {
			event := domain.Event{
				ID:          uuid.New().String(),
				WorkspaceID: capturedRule.WorkspaceID,
				Channel:     domain.ChannelCron,
				Fields: map[string]string{
					"rule_id":   capturedRule.ID,
					"rule_name": capturedRule.Name,
					"schedule":  schedule,
				},
				Raw:        map[string]interface{}{"rule_id": capturedRule.ID},
				ReceivedAt: time.Now(),
			}
			r.dispatcher.Dispatch(context.Background(), event)
		})

		if err != nil {
			log.Printf("[cron] invalid schedule %q for rule %s: %v", schedule, rule.ID, err)
			continue
		}

		r.entryMap[id] = entryID
		r.entryKeyMap[id] = cronEntryKey(rule)
		log.Printf("[cron] registered rule %q with schedule %q", rule.Name, schedule)
	}
}
