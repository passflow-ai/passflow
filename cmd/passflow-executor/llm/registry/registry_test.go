package registry

import (
	"testing"
)

func TestRegistry_Get(t *testing.T) {
	r := New()

	t.Run("existing model", func(t *testing.T) {
		model, ok := r.Get("gpt-4o")
		if !ok {
			t.Fatal("expected model to exist")
		}
		if model.Provider != "openai" {
			t.Errorf("provider = %q, want %q", model.Provider, "openai")
		}
	})

	t.Run("non-existing model", func(t *testing.T) {
		_, ok := r.Get("nonexistent")
		if ok {
			t.Error("expected model to not exist")
		}
	})

	t.Run("alias lookup", func(t *testing.T) {
		model, ok := r.Get("gpt4o")
		if !ok {
			t.Fatal("expected alias to resolve")
		}
		if model.ID != "gpt-4o" {
			t.Errorf("ID = %q, want %q", model.ID, "gpt-4o")
		}
	})
}

func TestRegistry_ListByProvider(t *testing.T) {
	r := New()

	models := r.ListByProvider("openai")
	if len(models) == 0 {
		t.Fatal("expected at least one openai model")
	}
	for _, m := range models {
		if m.Provider != "openai" {
			t.Errorf("unexpected provider %q", m.Provider)
		}
	}
}

func TestRegistry_Register(t *testing.T) {
	r := New()

	custom := Model{
		ID:       "custom-model",
		Provider: "custom",
		Capabilities: Capabilities{
			ToolCalling:   true,
			ContextWindow: 32000,
		},
	}

	r.Register(custom)

	got, ok := r.Get("custom-model")
	if !ok {
		t.Fatal("expected custom model to exist")
	}
	if got.Provider != "custom" {
		t.Errorf("provider = %q, want %q", got.Provider, "custom")
	}
}

func TestRegistry_All(t *testing.T) {
	r := New()

	models := r.All()
	if len(models) == 0 {
		t.Fatal("expected at least one model")
	}

	// Verify we have models from multiple providers
	providers := make(map[string]bool)
	for _, m := range models {
		providers[m.Provider] = true
	}

	expectedProviders := []string{"openai", "anthropic", "gemini", "ollama"}
	for _, p := range expectedProviders {
		if !providers[p] {
			t.Errorf("expected provider %q to be present", p)
		}
	}
}

func TestRegistry_RegisterWithAliases(t *testing.T) {
	r := New()

	custom := Model{
		ID:       "my-custom-llm",
		Provider: "custom-provider",
		Aliases:  []string{"my-llm", "custom-llm"},
		Capabilities: Capabilities{
			ToolCalling:   true,
			ContextWindow: 16000,
		},
	}

	r.Register(custom)

	// Should be found by ID
	if _, ok := r.Get("my-custom-llm"); !ok {
		t.Error("expected model to be found by ID")
	}

	// Should be found by first alias
	if m, ok := r.Get("my-llm"); !ok {
		t.Error("expected model to be found by alias 'my-llm'")
	} else if m.ID != "my-custom-llm" {
		t.Errorf("ID = %q, want %q", m.ID, "my-custom-llm")
	}

	// Should be found by second alias
	if m, ok := r.Get("custom-llm"); !ok {
		t.Error("expected model to be found by alias 'custom-llm'")
	} else if m.ID != "my-custom-llm" {
		t.Errorf("ID = %q, want %q", m.ID, "my-custom-llm")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := New()

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				r.Get("gpt-4o")
				r.ListByProvider("openai")
				r.All()
			}
			done <- true
		}()
	}

	// Test concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				r.Register(Model{
					ID:       "concurrent-model",
					Provider: "test",
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
