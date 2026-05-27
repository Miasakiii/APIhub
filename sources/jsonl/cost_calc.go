package jsonl

// CostCalc computes costs for JSONL usage records using a price index.
type CostCalc struct {
	// Lookup returns a ModelPrice for a given model_id.
	// Typically provided from ccswitch.PriceIndex.
	Lookup func(modelID string) (input, output, cacheRead, cacheCreate float64)
}

// CalcCost returns the cost in USD for a single UsageRecord.
func (cc *CostCalc) CalcCost(r UsageRecord) float64 {
	if cc.Lookup == nil {
		return 0
	}
	in, out, cr, cc_ := cc.Lookup(r.Model)
	million := 1_000_000.0
	return (float64(r.InputTokens)*in +
		float64(r.OutputTokens)*out +
		float64(r.CacheRead)*cr +
		float64(r.CacheCreate)*cc_) / million
}

// BatchCosts calculates costs for a batch of records and returns the total.
func (cc *CostCalc) BatchCosts(records []UsageRecord) (total float64, perRecord []float64) {
	perRecord = make([]float64, len(records))
	for i, r := range records {
		c := cc.CalcCost(r)
		perRecord[i] = c
		total += c
	}
	return total, perRecord
}
