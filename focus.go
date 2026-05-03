package libghostty

// Focus encoding — encode focus in/out events into terminal escape
// sequences (CSI I / CSI O) for focus reporting mode (mode 1004).
// Wraps the C API from focus.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// FocusEvent represents a focus gained or lost event for focus
// reporting mode (mode 1004).
//
// C: GhosttyFocusEvent
type FocusEvent int

const (
	// FocusGained indicates the terminal window gained focus.
	FocusGained FocusEvent = C.GHOSTTY_FOCUS_GAINED

	// FocusLost indicates the terminal window lost focus.
	FocusLost FocusEvent = C.GHOSTTY_FOCUS_LOST
)

// String returns a human-friendly name for the focus event:
// "gained" or "lost". Unknown values render as "unknown".
func (f FocusEvent) String() string {
	switch f {
	case FocusGained:
		return "gained"
	case FocusLost:
		return "lost"
	default:
		return "unknown"
	}
}

// FromString parses a focus event name ("gained" or "lost") and
// stores the corresponding FocusEvent value in the receiver.
// Returns an error if the name is not recognized.
func (f *FocusEvent) FromString(s string) error {
	switch s {
	case "gained":
		*f = FocusGained
	case "lost":
		*f = FocusLost
	default:
		return fmt.Errorf("libghostty: unknown focus event %q", s)
	}
	return nil
}

// NewFocusEventFromString returns the FocusEvent value for the
// given name ("gained" or "lost"). Returns an error if the name
// is not recognized.
func NewFocusEventFromString(s string) (FocusEvent, error) {
	var f FocusEvent
	if err := f.FromString(s); err != nil {
		return 0, err
	}
	return f, nil
}

// FocusEncode encodes a focus event into a terminal escape sequence
// and returns the result as a byte slice.
func FocusEncode(event FocusEvent) ([]byte, error) {
	// Focus sequences are short (CSI I or CSI O = 3 bytes).
	var buf [16]byte
	var outLen C.size_t
	result := C.ghostty_focus_encode(
		C.GhosttyFocusEvent(event),
		(*C.char)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&outLen,
	)

	if result == C.GHOSTTY_SUCCESS {
		if outLen == 0 {
			return nil, nil
		}
		out := make([]byte, outLen)
		copy(out, buf[:outLen])
		return out, nil
	}

	if result == C.GHOSTTY_OUT_OF_SPACE {
		dynBuf := make([]byte, outLen)
		var written C.size_t
		if err := resultError(C.ghostty_focus_encode(
			C.GhosttyFocusEvent(event),
			(*C.char)(unsafe.Pointer(&dynBuf[0])),
			outLen,
			&written,
		)); err != nil {
			return nil, err
		}
		if written == 0 {
			return nil, nil
		}
		return dynBuf[:written], nil
	}

	return nil, &Error{Result: Result(result)}
}
