package libghostty

// Color scheme report encoding from color_scheme.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// ColorSchemeReportEncode encodes a color scheme report escape sequence.
func ColorSchemeReportEncode(scheme ColorScheme) ([]byte, error) {
	var buf [32]byte
	var written C.size_t
	result := C.ghostty_color_scheme_report_encode(
		C.GhosttyColorScheme(scheme),
		(*C.char)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&written,
	)
	if result == C.GHOSTTY_SUCCESS {
		return append([]byte(nil), buf[:int(written)]...), nil
	}
	if result != C.GHOSTTY_OUT_OF_SPACE {
		return nil, &Error{Result: Result(result)}
	}

	out := make([]byte, int(written))
	var outWritten C.size_t
	if err := resultError(C.ghostty_color_scheme_report_encode(
		C.GhosttyColorScheme(scheme),
		(*C.char)(unsafe.Pointer(&out[0])),
		C.size_t(len(out)),
		&outWritten,
	)); err != nil {
		return nil, err
	}
	return out[:int(outWritten)], nil
}
