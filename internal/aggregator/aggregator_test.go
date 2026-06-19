package aggregator

import (
	"apihub/internal/model"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == id2 {
		t.Fatal("generateID should produce unique IDs")
	}
	if len(id1) != 32 { // 16 bytes hex-encoded
		t.Fatalf("expected 32-char hex ID, got %d chars: %s", len(id1), id1)
	}
}

func TestNormalizeModel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"claude-opus-4-8", "claude-opus-4-8"},
		{"mimo-v2.5-pro[1M]", "mimo-v2.5-pro"},
		{"claude-fable-5-thinking-high", "claude-fable-5"},
		{"claude-fable-5-thinking-xhigh", "claude-fable-5"},
		{"mimo-v2.5-pro-ultraspeed", "mimo-v2.5-pro"},
		{"gpt-4o[128k]", "gpt-4o"},
	}
	for _, tt := range tests {
		got := normalizeModel(tt.input)
		if got != tt.want {
			t.Errorf("normalizeModel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPricingLookup(t *testing.T) {
	// Set up a mock pricing cache
	globalPricing = &pricingCache{
		loaded: true,
		prices: map[string]model.ModelPricing{
			"claude-sonnet-4-6": {ModelID: "claude-sonnet-4-6", InputCostPerM: 3.0, OutputCostPerM: 15.0},
			"mimo-v2.5-pro":     {ModelID: "mimo-v2.5-pro", InputCostPerM: 0.5, OutputCostPerM: 2.0},
			"gpt-4o":            {ModelID: "gpt-4o", InputCostPerM: 2.5, OutputCostPerM: 10.0},
		},
	}

	tests := []struct {
		model   string
		wantOK  bool
		wantIn  float64
		wantOut float64
	}{
		{"claude-sonnet-4-6", true, 3.0, 15.0},
		{"claude-sonnet-4-6-thinking-high", true, 3.0, 15.0}, // prefix match
		{"mimo-v2.5-pro[1M]", true, 0.5, 2.0},               // bracket strip
		{"gpt-4o", true, 2.5, 10.0},
		{"unknown-model", false, 0, 0},
	}
	for _, tt := range tests {
		p, ok := globalPricing.Lookup(tt.model)
		if ok != tt.wantOK {
			t.Errorf("Lookup(%q) ok = %v, want %v", tt.model, ok, tt.wantOK)
			continue
		}
		if ok {
			if p.InputCostPerM != tt.wantIn || p.OutputCostPerM != tt.wantOut {
				t.Errorf("Lookup(%q) costs = (%v, %v), want (%v, %v)",
					tt.model, p.InputCostPerM, p.OutputCostPerM, tt.wantIn, tt.wantOut)
			}
		}
	}
}

func TestComputeCost(t *testing.T) {
	// Set up mock pricing
	globalPricing = &pricingCache{
		loaded: true,
		prices: map[string]model.ModelPricing{
			"gpt-4o": {
				ModelID:       "gpt-4o",
				InputCostPerM: 2.5,
				OutputCostPerM: 10.0,
				CacheReadCostPerM: 1.25,
			},
		},
	}

	// 1M input tokens at $2.5/M = $2.5, 500K output at $10/M = $5.0 → total $7.5
	cost := computeCost("gpt-4o", 1_000_000, 500_000, 0, 0)
	if cost < 7.49 || cost > 7.51 {
		t.Errorf("computeCost = %v, want ~7.5", cost)
	}

	// Unknown model → 0
	cost = computeCost("nonexistent", 1_000_000, 1_000_000, 0, 0)
	if cost != 0 {
		t.Errorf("computeCost for unknown model = %v, want 0", cost)
	}
}

func TestTruncateHour(t *testing.T) {
	input := time.Date(2026, 6, 19, 14, 37, 22, 999, time.UTC)
	got := truncateHour(input)
	want := time.Date(2026, 6, 19, 14, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("truncateHour(%v) = %v, want %v", input, got, want)
	}

	// Already at hour boundary
	input2 := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	got2 := truncateHour(input2)
	if !got2.Equal(input2) {
		t.Errorf("truncateHour(%v) = %v, want %v", input2, got2, input2)
	}
}

func TestSessionWindow(t *testing.T) {
	base := time.Date(2026, 6, 19, 14, 0, 0, 0, time.UTC)

	// Within window (10 min gap)
	r1 := model.UsageRecord{Timestamp: base}
	r2 := model.UsageRecord{Timestamp: base.Add(10 * time.Minute)}
	gap := r2.Timestamp.Sub(r1.Timestamp)
	if gap > sessionWindow {
		t.Errorf("10min gap should be within session window")
	}

	// Outside window (45 min gap)
	r3 := model.UsageRecord{Timestamp: base.Add(45 * time.Minute)}
	gap2 := r3.Timestamp.Sub(r1.Timestamp)
	if gap2 <= sessionWindow {
		t.Errorf("45min gap should be outside session window")
	}

	// Exactly at boundary (30 min)
	r4 := model.UsageRecord{Timestamp: base.Add(30 * time.Minute)}
	gap3 := r4.Timestamp.Sub(r1.Timestamp)
	if gap3 > sessionWindow {
		t.Errorf("30min gap should be within session window (inclusive)")
	}
}
