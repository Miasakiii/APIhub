package ws

import "time"

// Message types for WebSocket communication.
const (
	// Client -> Server
	TypePing = "ping"

	// Server -> Client
	TypePong          = "pong"
	TypeUsageUpdate   = "usage.update"
	TypeAlertTrigger  = "alert.triggered"
	TypeSyncProgress  = "sync.progress"
	TypeSyncComplete  = "sync.complete"
	TypeSyncError     = "sync.error"
)

// Message is the envelope for all WebSocket messages.
type Message struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// UsageUpdateData is the payload for usage.update messages.
type UsageUpdateData struct {
	RequestCount int     `json:"request_count"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

// AlertData is the payload for alert.triggered messages.
type AlertData struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// SyncProgressData is the payload for sync.progress messages.
type SyncProgressData struct {
	ProviderID    string  `json:"provider_id"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	ProcessedKeys int     `json:"processed_keys"`
	TotalKeys     int     `json:"total_keys"`
}

// SyncCompleteData is the payload for sync.complete messages.
type SyncCompleteData struct {
	ProviderID string `json:"provider_id"`
}

// SyncErrorData is the payload for sync.error messages.
type SyncErrorData struct {
	ProviderID string `json:"provider_id"`
	Error      string `json:"error"`
}

// NewMessage creates a new Message with the current timestamp.
func NewMessage(msgType string, data interface{}) *Message {
	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}
}
