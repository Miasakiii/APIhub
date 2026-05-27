package ccswitch

import (
	"math"
	"strconv"
)

// ModelPrice holds the pricing info for a single model, mirrored from cc-switch.
type ModelPrice struct {
	ModelID     string
	DisplayName string
	InputCost   float64 // per million tokens (TEXT in cc-switch, parsed here)
	OutputCost  float64
	CacheRead   float64 // cache_read cost per million (may be NULL)
	CacheCreate float64 // cache_creation cost per million (may be NULL)
}

// FetchModelPrices returns all model prices from cc-switch's model_pricing table.
func (r *Reader) FetchModelPrices() ([]ModelPrice, error) {
	rows, err := r.db.Query(`
		SELECT model_id, display_name,
		       input_cost_per_million,
		       output_cost_per_million,
		       cache_read_cost_per_million,
		       cache_creation_cost_per_million
		FROM model_pricing
		ORDER BY model_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []ModelPrice
	for rows.Next() {
		var p ModelPrice
		var inputStr, outputStr, cacheReadStr, cacheCreateStr sqlNullStr
		if err := rows.Scan(
			&p.ModelID, &p.DisplayName,
			&inputStr, &outputStr, &cacheReadStr, &cacheCreateStr,
		); err != nil {
			return nil, err
		}
		p.InputCost = parseFloat(inputStr.s)
		p.OutputCost = parseFloat(outputStr.s)
		p.CacheRead = parseFloat(cacheReadStr.s)
		p.CacheCreate = parseFloat(cacheCreateStr.s)
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// CalcCost computes the cost for a given token usage using the matched model price.
func CalcCost(p ModelPrice, input, output, cacheRead, cacheCreate int64) float64 {
	million := float64(1_000_000)
	cost := float64(input)*p.InputCost/million +
		float64(output)*p.OutputCost/million +
		float64(cacheRead)*p.CacheRead/million +
		float64(cacheCreate)*p.CacheCreate/million
	return math.Round(cost*10000) / 10000
}

// PriceIndex builds a lookup map from model_id to ModelPrice.
type PriceIndex map[string]ModelPrice

// NewPriceIndex builds an index from a price slice.
func NewPriceIndex(prices []ModelPrice) PriceIndex {
	idx := make(PriceIndex, len(prices))
	for _, p := range prices {
		idx[p.ModelID] = p
	}
	return idx
}

// Lookup finds the price for a model_id. Returns zero-cost ModelPrice if not found.
func (idx PriceIndex) Lookup(modelID string) ModelPrice {
	if p, ok := idx[modelID]; ok {
		return p
	}
	// Fallback: try prefix match (e.g. "claude-sonnet-4-6-20250514" → "claude-sonnet-4-6")
	for mid, p := range idx {
		if len(modelID) > len(mid) && modelID[:len(mid)] == mid {
			return p
		}
	}
	return ModelPrice{}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
