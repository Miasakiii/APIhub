package ws

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Broadcast with no clients should not panic
	msg := NewMessage(TypeUsageUpdate, &UsageUpdateData{
		RequestCount: 10,
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})
	hub.Broadcast(msg)

	// Give the hub time to process
	time.Sleep(10 * time.Millisecond)
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if count := hub.ClientCount(); count != 0 {
		t.Fatalf("expected 0 clients, got %d", count)
	}
}

func TestNewMessage(t *testing.T) {
	data := &AlertData{
		Level:   "warning",
		Title:   "Low Balance",
		Message: "Balance below $10",
	}
	msg := NewMessage(TypeAlertTrigger, data)

	if msg.Type != TypeAlertTrigger {
		t.Fatalf("expected type %s, got %s", TypeAlertTrigger, msg.Type)
	}
	if msg.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
	if msg.Data == nil {
		t.Fatal("expected non-nil data")
	}
}

func TestMessageTypes(t *testing.T) {
	// Verify message type constants are defined
	types := []string{
		TypePing, TypePong, TypeUsageUpdate, TypeAlertTrigger,
		TypeSyncProgress, TypeSyncComplete, TypeSyncError,
	}
	for _, tt := range types {
		if tt == "" {
			t.Fatal("empty message type constant")
		}
	}
}
