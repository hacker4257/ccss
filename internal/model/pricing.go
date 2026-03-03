package model

import "strings"

// ModelPricing holds per-million-token pricing in USD.
type ModelPricing struct {
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheReadPerMillion  float64 // ~0.1x input
	CacheWritePerMillion float64 // ~1.25x input
}

// PricingTable maps model ID prefixes to their pricing.
var PricingTable = map[string]ModelPricing{
	"claude-opus-4-6":   {15.00, 75.00, 1.50, 18.75},
	"claude-opus-4-5":   {15.00, 75.00, 1.50, 18.75},
	"claude-sonnet-4-5": {3.00, 15.00, 0.30, 3.75},
	"claude-sonnet-4":   {3.00, 15.00, 0.30, 3.75},
	"claude-haiku-4-5":  {0.80, 4.00, 0.08, 1.00},
	"claude-haiku-3-5":  {0.80, 4.00, 0.08, 1.00},
}

// LookupPricing finds pricing for a model ID by prefix matching.
func LookupPricing(modelID string) ModelPricing {
	for prefix, pricing := range PricingTable {
		if strings.HasPrefix(modelID, prefix) {
			return pricing
		}
	}
	// Unknown model: return sonnet pricing as safe default
	return PricingTable["claude-sonnet-4-5"]
}

// CalculateCost computes the USD cost for token counts using given pricing.
func CalculateCost(pricing ModelPricing, inputTokens, outputTokens, cacheReadTokens, cacheWriteTokens int) float64 {
	cost := float64(inputTokens) / 1_000_000 * pricing.InputPerMillion
	cost += float64(outputTokens) / 1_000_000 * pricing.OutputPerMillion
	cost += float64(cacheReadTokens) / 1_000_000 * pricing.CacheReadPerMillion
	cost += float64(cacheWriteTokens) / 1_000_000 * pricing.CacheWritePerMillion
	return cost
}
