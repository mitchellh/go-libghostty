package libghostty

// Paste utilities — validate and encode paste data for terminal input.
// Wraps the C APIs from paste.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// PasteIsSafe reports whether data is safe to paste into a terminal.
//
// Ghostty's safety check is intentionally conservative: data containing
// newlines or bracketed-paste end markers is considered unsafe because it can
// inject commands into interactive programs. Empty data is safe.
func PasteIsSafe(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	return bool(C.ghostty_paste_is_safe(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	))
}

// PasteEncode prepares data for writing to a terminal pty.
//
// The encoder applies Ghostty's terminal-input paste rules: unsafe control
// bytes are replaced with spaces, bracketed paste markers are added when
// bracketed is true, and newlines become carriage returns when bracketed is
// false. The input slice is copied before calling into Ghostty because the C
// encoder mutates its data buffer in place.
func PasteEncode(data []byte, bracketed bool) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// The current encoder only grows output by adding bracketed-paste markers,
	// but intentionally treat buffer sizing as an implementation detail of the C
	// API. Starting with modest slack keeps the common path allocation-light
	// while still honoring GHOSTTY_OUT_OF_SPACE below if the encoder changes.
	in := append([]byte(nil), data...)
	buf := make([]byte, len(data)+32)
	out, required, err := pasteEncodeInto(in, bracketed, buf)
	if err == nil {
		return out, nil
	}

	ge, ok := err.(*Error)
	if !ok || ge.Result != ResultOutOfSpace || required <= 0 {
		return nil, err
	}

	in = append([]byte(nil), data...)
	buf = make([]byte, required)
	out, _, err = pasteEncodeInto(in, bracketed, buf)
	return out, err
}

func pasteEncodeInto(
	data []byte,
	bracketed bool,
	buf []byte,
) ([]byte, int, error) {
	var written C.size_t
	result := C.ghostty_paste_encode(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		C.bool(bracketed),
		(*C.char)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&written,
	)
	if result == C.GHOSTTY_SUCCESS {
		return buf[:int(written)], int(written), nil
	}

	return nil, int(written), &Error{Result: Result(result)}
}
