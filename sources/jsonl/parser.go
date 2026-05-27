package jsonl

import (
	"bufio"
	"encoding/json"
	"io"
)

// UsageRecord is the parsed token usage from a single assistant message in JSONL.
// Maps to the message.usage structure (verified by spike round 4).
type UsageRecord struct {
	Model           string `json:"model"`
	InputTokens     int64  `json:"input_tokens"`
	OutputTokens    int64  `json:"output_tokens"`
	CacheRead       int64  `json:"cache_read_input_tokens"`
	CacheCreate     int64  `json:"cache_creation_input_tokens"`
	Timestamp       string `json:"timestamp"` // ISO 8601, from top-level
	ByteOffset      int64  `json:"-"`         // position in file for incremental sync
}

// ParseFile reads all assistant messages from a JSONL file and returns usage records.
// It skips non-assistant lines silently.
func ParseFile(r io.Reader) ([]UsageRecord, error) {
	records, _, err := ParseFileAt(r, 0)
	return records, err
}

// ParseFileAt reads assistant messages from a JSONL file starting at the given byte offset.
// Returns the parsed records and the final byte offset (for persisting to sync_state).
func ParseFileAt(r io.Reader, startOffset int64) ([]UsageRecord, int64, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // handle long lines

	var records []UsageRecord
	var offset int64 = startOffset
	for scanner.Scan() {
		line := scanner.Bytes()
		rec, err := parseLine(line, offset)
		if err != nil {
			offset += int64(len(line)) + 1 // +1 for newline
			continue // skip malformed lines
		}
		if rec != nil {
			records = append(records, *rec)
		}
		offset += int64(len(line)) + 1
	}
	return records, offset, scanner.Err()
}

// parseLine tries to extract usage from a single JSONL line.
// Returns nil for non-assistant lines (message, result, etc.).
func parseLine(line []byte, offset int64) (*UsageRecord, error) {
	// Quick pre-check before full JSON decode
	if !containsSubstring(line, `"type"`) || !containsSubstring(line, `"assistant"`) {
		return nil, nil
	}

	var obj map[string]any
	if err := json.Unmarshal(line, &obj); err != nil {
		return nil, err
	}

	// Check type == "assistant"
	if t, _ := obj["type"].(string); t != "assistant" {
		return nil, nil
	}

	// Extract message.usage
	msg, ok := obj["message"].(map[string]any)
	if !ok {
		return nil, nil
	}
	usage, ok := msg["usage"].(map[string]any)
	if !ok {
		return nil, nil
	}

	ts, _ := obj["timestamp"].(string)
	model, _ := msg["model"].(string)

	return &UsageRecord{
		Model:       model,
		InputTokens: toInt64(usage, "input_tokens"),
		OutputTokens: toInt64(usage, "output_tokens"),
		CacheRead:   toInt64(usage, "cache_read_input_tokens"),
		CacheCreate: toInt64(usage, "cache_creation_input_tokens"),
		Timestamp:   ts,
		ByteOffset:  offset,
	}, nil
}

func containsSubstring(haystack []byte, needle string) bool {
	n := []byte(needle)
	for i := 0; i <= len(haystack)-len(n); i++ {
		if string(haystack[i:i+len(n)]) == needle {
			return true
		}
	}
	return false
}

func toInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}
