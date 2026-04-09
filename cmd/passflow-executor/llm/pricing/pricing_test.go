package pricing

import "testing"

func TestEstimateCost_KnownModel(t *testing.T) {
	cost, source := EstimateCost("gpt-4o", 1_000_000, 1_000_000)
	if source != "calculated" {
		t.Errorf("expected source 'calculated', got %s", source)
	}
	expected := 2.50 + 10.00 // input + output per million
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestEstimateCost_UnknownModel(t *testing.T) {
	cost, source := EstimateCost("unknown-model", 1000, 1000)
	if source != "tokens_only" {
		t.Errorf("expected source 'tokens_only', got %s", source)
	}
	if cost != 0 {
		t.Errorf("expected cost 0, got %.2f", cost)
	}
}

func TestHasPrice(t *testing.T) {
	if !HasPrice("gpt-4o") {
		t.Error("gpt-4o should have a price")
	}
	if HasPrice("nonexistent") {
		t.Error("nonexistent should not have a price")
	}
}

func TestEstimateCost_ZeroTokens(t *testing.T) {
	cost, source := EstimateCost("gpt-4o", 0, 0)
	if source != "calculated" {
		t.Errorf("expected source 'calculated', got %s", source)
	}
	if cost != 0 {
		t.Errorf("expected cost 0, got %.2f", cost)
	}
}
