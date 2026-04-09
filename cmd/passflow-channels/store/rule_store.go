package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
)

const redisSyncInterval = 30 * time.Second

// RuleStore holds trigger rules with an in-memory cache backed by Redis.
type RuleStore struct {
	mu    sync.RWMutex
	rules map[string]domain.TriggerRule // keyed by rule ID
	rdb   *redis.Client
}

// New creates a new RuleStore and starts background sync.
func New(ctx context.Context, rdb *redis.Client) *RuleStore {
	s := &RuleStore{
		rules: make(map[string]domain.TriggerRule),
		rdb:   rdb,
	}
	// Initial load
	s.syncFromRedis(ctx)

	// Background sync
	go func() {
		ticker := time.NewTicker(redisSyncInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.syncFromRedis(ctx)
			}
		}
	}()

	return s
}

// GetAllRules returns all rules.
func (s *RuleStore) GetAllRules() []domain.TriggerRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]domain.TriggerRule, 0, len(s.rules))
	for _, r := range s.rules {
		rules = append(rules, r)
	}
	return rules
}

// GetRulesForWorkspace returns rules for a specific workspace.
func (s *RuleStore) GetRulesForWorkspace(workspaceID string) []domain.TriggerRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var rules []domain.TriggerRule
	for _, r := range s.rules {
		if r.WorkspaceID == workspaceID {
			rules = append(rules, r)
		}
	}
	return rules
}

// GetByID returns a rule by its ID. Returns nil if not found.
func (s *RuleStore) GetByID(ruleID string) *domain.TriggerRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if rule, ok := s.rules[ruleID]; ok {
		return &rule
	}
	return nil
}

// Upsert adds or updates a rule in both the cache and Redis.
func (s *RuleStore) Upsert(ctx context.Context, rule domain.TriggerRule) error {
	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("store: failed to marshal rule: %w", err)
	}

	if err := s.rdb.HSet(ctx, "passflow:trigger_rules", rule.ID, string(data)).Err(); err != nil {
		return fmt.Errorf("store: failed to persist rule: %w", err)
	}

	s.mu.Lock()
	s.rules[rule.ID] = rule
	s.mu.Unlock()

	return nil
}

// Delete removes a rule.
func (s *RuleStore) Delete(ctx context.Context, ruleID string) error {
	if err := s.rdb.HDel(ctx, "passflow:trigger_rules", ruleID).Err(); err != nil {
		return fmt.Errorf("store: failed to delete rule: %w", err)
	}

	s.mu.Lock()
	delete(s.rules, ruleID)
	s.mu.Unlock()

	return nil
}

// syncFromRedis loads all rules from Redis into the in-memory cache.
func (s *RuleStore) syncFromRedis(ctx context.Context) {
	all, err := s.rdb.HGetAll(ctx, "passflow:trigger_rules").Result()
	if err != nil {
		log.Printf("[store] failed to sync rules from Redis: %v", err)
		return
	}

	rules := make(map[string]domain.TriggerRule, len(all))
	for id, data := range all {
		var rule domain.TriggerRule
		if err := json.Unmarshal([]byte(data), &rule); err != nil {
			log.Printf("[store] failed to unmarshal rule %s: %v", id, err)
			continue
		}
		rules[id] = rule
	}

	s.mu.Lock()
	s.rules = rules
	s.mu.Unlock()
}
