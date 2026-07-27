package libghostty

// Unicode display-width utilities from unicode.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// UnicodeCodepointWidth returns the terminal display width of cp in cells.
// It uses the same width table as libghostty's terminal text layout.
func UnicodeCodepointWidth(cp uint32) uint8 {
	return uint8(C.ghostty_unicode_codepoint_width(C.uint32_t(cp)))
}

// UnicodeGraphemeWidth measures the first grapheme cluster in codepoints.
// It returns the number of codepoints consumed and that cluster's display
// width in cells. Empty input returns zero for both values.
func UnicodeGraphemeWidth(codepoints []uint32) (consumed int, width uint8) {
	if len(codepoints) == 0 {
		return 0, 0
	}

	// The C API is read-only for this call, so the Go-owned slice remains
	// valid without a copy.
	var out C.uint8_t
	n := C.ghostty_unicode_grapheme_width(
		(*C.uint32_t)(unsafe.Pointer(&codepoints[0])),
		C.size_t(len(codepoints)),
		&out,
	)
	return int(n), uint8(out)
}
