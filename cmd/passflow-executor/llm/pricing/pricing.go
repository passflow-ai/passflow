package pricing

import (
	_ "embed"
	"encoding/json"
)

//go:embed models.json
var modelsJSON []byte

// ModelPrice holds the per-million-token pricing for a model in USD.
type ModelPrice struct {
	InputPerMillion  float64 `json:"inputPerMillion"`
	OutputPerMillion float64 `json:"outputPerMillion"`
}

// PricingTable maps model IDs to their pricing.
type PricingTable struct {
	Models map[string]ModelPrice `json:"models"`
}

var table PricingTable

func init() {
	if err := json.Unmarshal(modelsJSON, &table); err != nil {
		panic("pricing: failed to parse models.json: " + err.Error())
	}
}

// EstimateCost returns the estimated cost in USD and a source string.
// Source is "calculated" when the model has known pricing, "tokens_only" otherwise.
func EstimateCost(model string, tokensIn, tokensOut int64) (float64, string) {
	price, ok := table.Models[model]
	if !ok {
		return 0, "tokens_only"
	}
	cost := (float64(tokensIn) * price.InputPerMillion / 1_000_000) +
		(float64(tokensOut) * price.OutputPerMillion / 1_000_000)
	return cost, "calculated"
}

// HasPrice returns true if the model has known pricing data.
func HasPrice(model string) bool {
	_, ok := table.Models[model]
	return ok
}
