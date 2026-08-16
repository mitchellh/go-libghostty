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
	if schema, ok := decoded["schema"].(float64); !ok || schema != 1 {
		t.Fatalf("expected schema version 1, got %#v", decoded["schema"])
	}
	if _, ok := decoded["abi"].(map[string]any); !ok {
		t.Fatalf("expected ABI metadata, got %#v", decoded["abi"])
	}
	types, ok := decoded["types"].(map[string]any)
	if !ok {
		t.Fatalf("expected type descriptors, got %#v", decoded["types"])
	}
	cell, ok := types["GhosttyCell"].(map[string]any)
	if !ok {
		t.Fatalf("expected GhosttyCell descriptor, got %#v", types["GhosttyCell"])
	}
	if cell["kind"] != "packed" {
		t.Fatalf("expected packed GhosttyCell descriptor, got %#v", cell["kind"])
	}
}
