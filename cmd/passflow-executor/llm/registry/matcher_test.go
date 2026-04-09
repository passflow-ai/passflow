package registry

import (
	"testing"
)

func TestMatcher_FindFallbacks(t *testing.T) {
	r := New()
	m := NewMatcher(r)

	t.Run("finds compatible models", func(t *testing.T) {
		required := Capabilities{
			ToolCalling: true,
			Vision:      true,
		}
		fallbacks := m.FindFallbacks(required, []string{"gpt-4o"})

		if len(fallbacks) == 0 {
			t.Fatal("expected at least one fallback")
		}

		for _, fb := range fallbacks {
			if fb.ID == "gpt-4o" {
				t.Error("should not include excluded model")
			}
			if !fb.Capabilities.ToolCalling || !fb.Capabilities.Vision {
				t.Errorf("fallback %s does not satisfy requirements", fb.ID)
			}
		}
	})

	t.Run("excludes models without required capabilities", func(t *testing.T) {
		required := Capabilities{
			Vision: true,
		}
		fallbacks := m.FindFallbacks(required, nil)

		for _, fb := range fallbacks {
			if !fb.Capabilities.Vision {
				t.Errorf("fallback %s should have vision", fb.ID)
			}
		}
	})

	t.Run("respects max fallbacks limit", func(t *testing.T) {
		required := Capabilities{ToolCalling: true}
		fallbacks := m.FindFallbacks(required, nil)

		if len(fallbacks) > 3 {
			t.Errorf("got %d fallbacks, max should be 3", len(fallbacks))
		}
	})
}

func TestMatcher_MatchScore(t *testing.T) {
	m := NewMatcher(New())

	model := Model{
		ID:       "gpt-4o",
		Provider: "openai",
		Capabilities: Capabilities{
			ToolCalling:   true,
			Vision:        true,
			ContextWindow: 128000,
			Streaming:     true,
			JSONMode:      true,
			FunctionStyle: "openai",
		},
	}

	t.Run("same function style scores higher", func(t *testing.T) {
		required := Capabilities{FunctionStyle: "openai"}
		score := m.MatchScore(model, required)
		if score <= 0 {
			t.Error("expected positive score for same function style")
		}
	})

	t.Run("larger context window scores higher", func(t *testing.T) {
		required := Capabilities{}
		score := m.MatchScore(model, required)
		// gpt-4o has 128000 context, should get context bonus
		if score < 3 {
			t.Errorf("expected context window bonus, got score %d", score)
		}
	})
}
