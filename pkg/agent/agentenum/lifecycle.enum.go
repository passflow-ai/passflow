package agentenum

// LifecycleState represents the approval workflow state of an agent.
type LifecycleState string

const (
	LifecycleDraft         LifecycleState = "draft"
	LifecyclePendingReview LifecycleState = "pending_review"
	LifecycleApproved      LifecycleState = "approved"
	LifecyclePublished     LifecycleState = "published"
	LifecycleArchived      LifecycleState = "archived"
)

// CanTransitionTo checks if this lifecycle state can transition to the target state.
func (s LifecycleState) CanTransitionTo(target LifecycleState) bool {
	transitions := map[LifecycleState][]LifecycleState{
		LifecycleDraft:         {LifecyclePendingReview},
		LifecyclePendingReview: {LifecycleApproved, LifecycleDraft},
		LifecycleApproved:      {LifecyclePublished, LifecycleArchived},
		LifecyclePublished:     {LifecycleArchived, LifecycleDraft},
		LifecycleArchived:      {LifecycleDraft},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// String returns the string representation of the lifecycle state.
func (s LifecycleState) String() string {
	return string(s)
}

// IsValid checks if the lifecycle state is a recognized value.
func (s LifecycleState) IsValid() bool {
	switch s {
	case LifecycleDraft, LifecyclePendingReview, LifecycleApproved, LifecyclePublished, LifecycleArchived:
		return true
	}
	return false
}
