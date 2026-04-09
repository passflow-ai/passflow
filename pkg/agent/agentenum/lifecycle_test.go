package agentenum

import "testing"

func TestLifecycleState_ValidTransitions(t *testing.T) {
	tests := []struct {
		from    LifecycleState
		to      LifecycleState
		allowed bool
	}{
		{LifecycleDraft, LifecyclePendingReview, true},
		{LifecycleDraft, LifecyclePublished, false},
		{LifecyclePendingReview, LifecycleApproved, true},
		{LifecyclePendingReview, LifecycleDraft, true},
		{LifecycleApproved, LifecyclePublished, true},
		{LifecycleApproved, LifecycleArchived, true},
		{LifecycleApproved, LifecycleDraft, false},
		{LifecyclePublished, LifecycleArchived, true},
		{LifecyclePublished, LifecycleDraft, true},
		{LifecycleArchived, LifecycleDraft, true},
		{LifecycleArchived, LifecyclePublished, false},
	}

	for _, tt := range tests {
		result := tt.from.CanTransitionTo(tt.to)
		if result != tt.allowed {
			t.Errorf("%s -> %s: expected %v, got %v", tt.from, tt.to, tt.allowed, result)
		}
	}
}

func TestLifecycleState_IsValid(t *testing.T) {
	valid := []LifecycleState{LifecycleDraft, LifecyclePendingReview, LifecycleApproved, LifecyclePublished, LifecycleArchived}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %s to be valid", s)
		}
	}

	invalid := LifecycleState("unknown")
	if invalid.IsValid() {
		t.Error("expected unknown to be invalid")
	}
}

func TestLifecycleState_String(t *testing.T) {
	if LifecycleDraft.String() != "draft" {
		t.Errorf("expected 'draft', got '%s'", LifecycleDraft.String())
	}
}
