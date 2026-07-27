package libghostty

import "testing"

func TestUnicodeWidths(t *testing.T) {
	if got := UnicodeCodepointWidth('a'); got != 1 {
		t.Fatalf("expected narrow ASCII width 1, got %d", got)
	}
	if got := UnicodeCodepointWidth(0x0301); got != 0 {
		t.Fatalf("expected combining mark width 0, got %d", got)
	}
	if got := UnicodeCodepointWidth(0x1f600); got != 2 {
		t.Fatalf("expected emoji width 2, got %d", got)
	}

	// Woman technologist: woman + ZWJ + laptop.
	consumed, width := UnicodeGraphemeWidth([]uint32{0x1f469, 0x200d, 0x1f4bb})
	if consumed != 3 || width != 2 {
		t.Fatalf("expected one 3-codepoint wide cluster, got consumed=%d width=%d", consumed, width)
	}

	if consumed, width := UnicodeGraphemeWidth(nil); consumed != 0 || width != 0 {
		t.Fatalf("expected empty input result 0,0, got %d,%d", consumed, width)
	}
}
