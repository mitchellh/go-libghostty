package libghostty

import (
	"encoding/json"
	"testing"
)

func TestTypeJSON(t *testing.T) {
	value := TypeJSON()
	if value == "" {
		t.Fatal("expected non-empty type metadata")
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		t.Fatalf("invalid type metadata JSON: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("expected type metadata entries")
	}
}
