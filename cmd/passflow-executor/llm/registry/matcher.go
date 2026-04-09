package registry

import (
	"sort"
)

const maxFallbacks = 3

// Matcher finds compatible fallback models.
type Matcher struct {
	registry *Registry
}

// NewMatcher creates a new capability matcher.
func NewMatcher(r *Registry) *Matcher {
	return &Matcher{registry: r}
}

// FindFallbacks returns up to 3 models that satisfy the required capabilities,
// excluding the specified model IDs, sorted by match score.
func (m *Matcher) FindFallbacks(required Capabilities, exclude []string) []Model {
	excludeSet := make(map[string]bool)
	for _, id := range exclude {
		excludeSet[id] = true
	}

	var candidates []Model
	for _, model := range m.registry.All() {
		if excludeSet[model.ID] {
			continue
		}
		if !model.Capabilities.Satisfies(required) {
			continue
		}
		candidates = append(candidates, model)
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		scoreI := m.MatchScore(candidates[i], required)
		scoreJ := m.MatchScore(candidates[j], required)
		return scoreI > scoreJ
	})

	if len(candidates) > maxFallbacks {
		candidates = candidates[:maxFallbacks]
	}
	return candidates
}

// MatchScore calculates how well a model matches the requirements.
// Higher score = better match.
func (m *Matcher) MatchScore(model Model, required Capabilities) int {
	score := 0

	// Same function style = +10 (less conversion needed)
	if required.FunctionStyle != "" && model.Capabilities.FunctionStyle == required.FunctionStyle {
		score += 10
	}

	// Context window bonus (log scale)
	if model.Capabilities.ContextWindow >= 200000 {
		score += 5
	} else if model.Capabilities.ContextWindow >= 128000 {
		score += 3
	} else if model.Capabilities.ContextWindow >= 32000 {
		score += 1
	}

	// Extra capabilities bonus
	if model.Capabilities.Vision {
		score += 2
	}
	if model.Capabilities.JSONMode {
		score += 1
	}
	if model.Capabilities.Streaming {
		score += 1
	}

	return score
}
