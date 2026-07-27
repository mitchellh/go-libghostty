package libghostty

import "testing"

func TestDefaultStyle(t *testing.T) {
	style := DefaultStyle()
	if !style.IsDefault() {
		t.Fatal("expected default style")
	}
	if style.Underline() != UnderlineNone {
		t.Fatalf("expected no underline, got %d", style.Underline())
	}
}
