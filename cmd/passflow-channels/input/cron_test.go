package input

import (
	"sync"
	"testing"
	"time"

	"github.com/jaak-ai/passflow-channels/domain"
	cronlib "github.com/robfig/cron/v3"
)

// mockRuleStore is a thread-safe in-memory store for test rules.
type mockRuleStore struct {
	mu    sync.RWMutex
	rules []domain.TriggerRule
}

func (s *mockRuleStore) GetAllRules() []domain.TriggerRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.TriggerRule, len(s.rules))
	copy(out, s.rules)
	return out
}

func (s *mockRuleStore) setRules(rules []domain.TriggerRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = rules
}

// cronCapture records which rule IDs fired.
type cronCapture struct {
	mu  sync.Mutex
	ids []string
}

func (c *cronCapture) record(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ids = append(c.ids, id)
}

func (c *cronCapture) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.ids)
}

// TestCronRunner_DiffReload_AddsNewRule verifies that reloading with a new rule
// registers it without touching existing unmodified entries.
func TestCronRunner_DiffReload_AddsNewRule(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
		},
	}

	runner := NewCronRunner(nil, store)

	runner.reloadRules()
	entryCountAfterFirst := len(runner.entryMap)
	if entryCountAfterFirst != 1 {
		t.Fatalf("expected 1 entry after first load, got %d", entryCountAfterFirst)
	}
	firstEntryID := runner.entryMap["rule-1"]

	// Add a second rule — the existing rule must keep its entry ID (no teardown).
	store.setRules([]domain.TriggerRule{
		{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
			ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
		{ID: "rule-2", Name: "r2", WorkspaceID: "ws-1", Enabled: true,
			ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
	})
	runner.reloadRules()

	if len(runner.entryMap) != 2 {
		t.Fatalf("expected 2 entries after second load, got %d", len(runner.entryMap))
	}
	if runner.entryMap["rule-1"] != firstEntryID {
		t.Errorf("rule-1 entry ID changed after diff reload — existing rule was re-registered unnecessarily")
	}
	if _, ok := runner.entryMap["rule-2"]; !ok {
		t.Error("rule-2 was not added")
	}
}

// TestCronRunner_DiffReload_RemovesDeletedRule verifies that a removed rule
// is de-registered from the cron scheduler.
func TestCronRunner_DiffReload_RemovesDeletedRule(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
			{ID: "rule-2", Name: "r2", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
		},
	}

	runner := NewCronRunner(nil, store)
	runner.reloadRules()

	if len(runner.entryMap) != 2 {
		t.Fatalf("expected 2 entries after first load, got %d", len(runner.entryMap))
	}

	// Remove rule-2
	store.setRules([]domain.TriggerRule{
		{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
			ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
	})
	runner.reloadRules()

	if len(runner.entryMap) != 1 {
		t.Fatalf("expected 1 entry after removal, got %d", len(runner.entryMap))
	}
	if _, ok := runner.entryMap["rule-2"]; ok {
		t.Error("rule-2 still in entryMap after being removed from the store")
	}
}

// TestCronRunner_DiffReload_UpdatesChangedSchedule verifies that a rule with
// a changed schedule is re-registered with the new schedule.
func TestCronRunner_DiffReload_UpdatesChangedSchedule(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "0 * * * * *"}},
		},
	}

	runner := NewCronRunner(nil, store)
	runner.reloadRules()
	originalID := runner.entryMap["rule-1"]

	// Change the schedule
	store.setRules([]domain.TriggerRule{
		{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
			ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "30 * * * * *"}},
	})
	runner.reloadRules()

	newID := runner.entryMap["rule-1"]
	if newID == originalID {
		t.Error("entry ID unchanged after schedule change — cron entry was not re-registered")
	}
}

// TestCronRunner_DiffReload_DisabledRuleNotRegistered verifies that disabled
// rules are not registered even when present in the store.
func TestCronRunner_DiffReload_DisabledRuleNotRegistered(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: false,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
		},
	}

	runner := NewCronRunner(nil, store)
	runner.reloadRules()

	if len(runner.entryMap) != 0 {
		t.Errorf("expected 0 entries for disabled rule, got %d", len(runner.entryMap))
	}
}

// TestCronRunner_DiffReload_ConcurrentSafe verifies that concurrent calls to
// reloadRules do not cause data races (run with -race).
func TestCronRunner_DiffReload_ConcurrentSafe(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "rule-1", Name: "r1", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "* * * * * *"}},
		},
	}

	runner := NewCronRunner(nil, store)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.reloadRules()
		}()
	}
	wg.Wait()
}

// TestCronRunner_entryKey_DetectsChange verifies the entryKey helper reflects
// all fields that should trigger re-registration.
func TestCronRunner_entryKey_DetectsChange(t *testing.T) {
	r1 := domain.TriggerRule{
		ID: "x", WorkspaceID: "ws", Name: "n", Enabled: true,
		ChannelType: domain.ChannelCron,
		Condition:   domain.Condition{Value: "* * * * * *"},
	}
	r2 := r1
	r2.Condition.Value = "0 * * * * *"

	if cronEntryKey(r1) == cronEntryKey(r2) {
		t.Error("entryKey should differ when schedule changes")
	}

	r3 := r1
	r3.Enabled = false
	if cronEntryKey(r1) == cronEntryKey(r3) {
		t.Error("entryKey should differ when enabled state changes")
	}
}

// TestCronRunner_reloadRules_InvalidScheduleSkipped verifies that an invalid
// cron expression does not panic and the rule is skipped.
func TestCronRunner_reloadRules_InvalidScheduleSkipped(t *testing.T) {
	store := &mockRuleStore{
		rules: []domain.TriggerRule{
			{ID: "bad-rule", Name: "bad", WorkspaceID: "ws-1", Enabled: true,
				ChannelType: domain.ChannelCron, Condition: domain.Condition{Value: "NOT_A_CRON"}},
		},
	}

	runner := NewCronRunner(nil, store)
	// Must not panic.
	runner.reloadRules()

	if _, ok := runner.entryMap["bad-rule"]; ok {
		t.Error("invalid schedule rule should not be in entryMap")
	}
}

// Ensure cronlib.EntryID is a comparable type for test assertions.
var _ cronlib.EntryID = cronlib.EntryID(0)

// Ensure CronRunner exported for test (same package).
var _ = time.Second
