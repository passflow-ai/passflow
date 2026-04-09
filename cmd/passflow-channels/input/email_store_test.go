package input

import (
	"testing"
)

// TestProcessedSet_AddAndContains verifies basic add/contains semantics.
func TestProcessedSet_AddAndContains(t *testing.T) {
	ps := newProcessedSet(100)
	ps.add("msg-001")
	if !ps.contains("msg-001") {
		t.Error("expected processedSet to contain 'msg-001' after add")
	}
}

// TestProcessedSet_ContainsFalseForUnknown verifies that an unknown ID is not
// reported as processed.
func TestProcessedSet_ContainsFalseForUnknown(t *testing.T) {
	ps := newProcessedSet(100)
	if ps.contains("msg-unknown") {
		t.Error("expected processedSet NOT to contain 'msg-unknown'")
	}
}

// TestProcessedSet_Idempotent verifies that adding the same ID twice does not
// cause errors or duplicate tracking.
func TestProcessedSet_Idempotent(t *testing.T) {
	ps := newProcessedSet(100)
	ps.add("msg-dup")
	ps.add("msg-dup")
	if !ps.contains("msg-dup") {
		t.Error("expected idempotent add to leave entry present")
	}
}

// TestProcessedSet_CapacityEviction verifies that when the set exceeds its
// capacity, old entries are evicted so the set does not grow unboundedly.
func TestProcessedSet_CapacityEviction(t *testing.T) {
	const cap = 5
	ps := newProcessedSet(cap)
	for i := 0; i < cap+3; i++ {
		ps.add(string(rune('A' + i))) // A, B, C, D, E, F, G, H
	}
	// The set must not exceed cap entries.
	if ps.size() > cap {
		t.Errorf("processedSet grew to %d, expected <= %d", ps.size(), cap)
	}
}

// TestProcessedSet_EmptyIDSkipped verifies that an empty message ID is not
// added (it would be useless as a dedup key).
func TestProcessedSet_EmptyIDSkipped(t *testing.T) {
	ps := newProcessedSet(10)
	ps.add("")
	if ps.contains("") {
		t.Error("expected empty ID to be skipped")
	}
}
