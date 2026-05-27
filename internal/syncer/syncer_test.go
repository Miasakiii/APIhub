package syncer

import (
	"context"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	names := registry.Names()
	if len(names) != 0 {
		t.Fatalf("expected 0 syncers, got %d", len(names))
	}

	// Create a mock syncer
	mockSyncer := &mockSyncer{name: "test"}
	registry.Register(mockSyncer)

	// Test get
	s, ok := registry.Get("test")
	if !ok {
		t.Fatal("expected to find syncer 'test'")
	}
	if s.Name() != "test" {
		t.Fatalf("expected name 'test', got '%s'", s.Name())
	}

	// Test names
	names = registry.Names()
	if len(names) != 1 || names[0] != "test" {
		t.Fatalf("expected ['test'], got %v", names)
	}

	// Test non-existent
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find syncer 'nonexistent'")
	}
}

type mockSyncer struct {
	name string
}

func (m *mockSyncer) Name() string { return m.name }
func (m *mockSyncer) FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]Record, error) {
	return nil, nil
}
func (m *mockSyncer) FetchBalance(ctx context.Context, apiKey string, baseURL string) (*BalanceInfo, error) {
	return nil, nil
}
func (m *mockSyncer) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	return nil
}
func (m *mockSyncer) SupportsUsage() bool  { return true }
func (m *mockSyncer) SupportsBalance() bool { return true }
