package libghostty

// Key encoder — encodes key events into terminal escape sequences.
// Wraps the C APIs from key/encoder.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// KeyEncoder encodes key events into terminal escape sequences,
// supporting both legacy encoding and the Kitty Keyboard Protocol.
// It maintains mutable encoding options and is not safe for concurrent
// use.
//
// Basic usage:
//  1. Create an encoder with NewKeyEncoder.
//  2. Configure options with SetOpt* methods or SetOptFromTerminal.
//  3. Create key events, encode them with Encode, and free them.
//  4. Free the encoder with Close when done.
//
// C: GhosttyKeyEncoder
type KeyEncoder struct {
	ptr C.GhosttyKeyEncoder
}

// KittyKeyFlags is a bitmask of Kitty keyboard protocol flags.
// C: GhosttyKittyKeyFlags
type KittyKeyFlags uint8

const (
	// KittyKeyDisabled disables the Kitty keyboard protocol (all flags off).
	KittyKeyDisabled KittyKeyFlags = C.GHOSTTY_KITTY_KEY_DISABLED

	// KittyKeyDisambiguate enables disambiguating escape codes.
	KittyKeyDisambiguate KittyKeyFlags = C.GHOSTTY_KITTY_KEY_DISAMBIGUATE

	// KittyKeyReportEvents enables reporting key press and release events.
	KittyKeyReportEvents KittyKeyFlags = C.GHOSTTY_KITTY_KEY_REPORT_EVENTS

	// KittyKeyReportAlternates enables reporting alternate key codes.
	KittyKeyReportAlternates KittyKeyFlags = C.GHOSTTY_KITTY_KEY_REPORT_ALTERNATES

	// KittyKeyReportAll reports all key events including those normally
	// handled by the terminal.
	KittyKeyReportAll KittyKeyFlags = C.GHOSTTY_KITTY_KEY_REPORT_ALL

	// KittyKeyReportAssociated reports associated text with key events.
	KittyKeyReportAssociated KittyKeyFlags = C.GHOSTTY_KITTY_KEY_REPORT_ASSOCIATED

	// KittyKeyAll enables all Kitty keyboard protocol flags.
	KittyKeyAll KittyKeyFlags = C.GHOSTTY_KITTY_KEY_ALL
)

// OptionAsAlt determines whether the macOS "option" key is treated as
// "alt".
//
// C: GhosttyOptionAsAlt
type OptionAsAlt int

const (
	// OptionAsAltFalse means the option key is not treated as alt.
	OptionAsAltFalse OptionAsAlt = C.GHOSTTY_OPTION_AS_ALT_FALSE

	// OptionAsAltTrue means the option key is treated as alt.
	OptionAsAltTrue OptionAsAlt = C.GHOSTTY_OPTION_AS_ALT_TRUE

	// OptionAsAltLeft means only the left option key is treated as alt.
	OptionAsAltLeft OptionAsAlt = C.GHOSTTY_OPTION_AS_ALT_LEFT

	// OptionAsAltRight means only the right option key is treated as alt.
	OptionAsAltRight OptionAsAlt = C.GHOSTTY_OPTION_AS_ALT_RIGHT
)

// KeyEncoderOption identifies an encoder configuration option for use
// with SetOpt.
//
// C: GhosttyKeyEncoderOption
type KeyEncoderOption int

const (
	// KeyEncoderOptCursorKeyApplication sets DEC mode 1: cursor key
	// application mode (value: bool).
	KeyEncoderOptCursorKeyApplication KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_CURSOR_KEY_APPLICATION

	// KeyEncoderOptKeypadKeyApplication sets DEC mode 66: keypad key
	// application mode (value: bool).
	KeyEncoderOptKeypadKeyApplication KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_KEYPAD_KEY_APPLICATION

	// KeyEncoderOptIgnoreKeypadWithNumlock sets DEC mode 1035: ignore
	// keypad with numlock (value: bool).
	KeyEncoderOptIgnoreKeypadWithNumlock KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_IGNORE_KEYPAD_WITH_NUMLOCK

	// KeyEncoderOptAltEscPrefix sets DEC mode 1036: alt sends escape
	// prefix (value: bool).
	KeyEncoderOptAltEscPrefix KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_ALT_ESC_PREFIX

	// KeyEncoderOptModifyOtherKeysState2 sets xterm modifyOtherKeys
	// mode 2 (value: bool).
	KeyEncoderOptModifyOtherKeysState2 KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_MODIFY_OTHER_KEYS_STATE_2

	// KeyEncoderOptKittyFlags sets Kitty keyboard protocol flags
	// (value: KittyKeyFlags bitmask).
	KeyEncoderOptKittyFlags KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_KITTY_FLAGS

	// KeyEncoderOptMacOSOptionAsAlt sets the macOS option-as-alt
	// setting (value: OptionAsAlt).
	KeyEncoderOptMacOSOptionAsAlt KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_MACOS_OPTION_AS_ALT

	// KeyEncoderOptBackarrowKeyMode sets backarrow key mode (value: bool).
	// When false (default), backspace emits 0x7f; when true, 0x08.
	KeyEncoderOptBackarrowKeyMode KeyEncoderOption = C.GHOSTTY_KEY_ENCODER_OPT_BACKARROW_KEY_MODE
)

// NewKeyEncoder creates a new key encoder with default options.
// The encoder must be freed with Close when no longer needed.
func NewKeyEncoder() (*KeyEncoder, error) {
	var ptr C.GhosttyKeyEncoder
	if err := resultError(C.ghostty_key_encoder_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &KeyEncoder{ptr: ptr}, nil
}

// Close frees the underlying key encoder handle. After this call,
// the encoder must not be used.
func (enc *KeyEncoder) Close() {
	C.ghostty_key_encoder_free(enc.ptr)
}

// SetOptBool sets a boolean encoder option. Use this for options
// that accept a bool value (most encoder options).
func (enc *KeyEncoder) SetOptBool(opt KeyEncoderOption, val bool) {
	v := C.bool(val)
	C.ghostty_key_encoder_setopt(enc.ptr, C.GhosttyKeyEncoderOption(opt), unsafe.Pointer(&v))
}

// SetOptKittyFlags sets the Kitty keyboard protocol flags on the
// encoder.
func (enc *KeyEncoder) SetOptKittyFlags(flags KittyKeyFlags) {
	v := C.GhosttyKittyKeyFlags(flags)
	C.ghostty_key_encoder_setopt(enc.ptr, C.GHOSTTY_KEY_ENCODER_OPT_KITTY_FLAGS, unsafe.Pointer(&v))
}

// SetOptOptionAsAlt sets the macOS option-as-alt behavior on the
// encoder.
func (enc *KeyEncoder) SetOptOptionAsAlt(val OptionAsAlt) {
	v := C.GhosttyOptionAsAlt(val)
	C.ghostty_key_encoder_setopt(enc.ptr, C.GHOSTTY_KEY_ENCODER_OPT_MACOS_OPTION_AS_ALT, unsafe.Pointer(&v))
}

// SetOptFromTerminal reads the terminal's current modes and flags and
// applies them to the encoder's options. This sets cursor key
// application mode, keypad mode, alt escape prefix, modifyOtherKeys
// state, and Kitty keyboard protocol flags from the terminal state.
//
// Note that the macOS option-as-alt option cannot be determined from
// terminal state and is reset to OptionAsAltFalse by this call. Use
// SetOptOptionAsAlt afterward if needed. The caller must serialize
// access to both the encoder and the terminal during this call.
func (enc *KeyEncoder) SetOptFromTerminal(t *Terminal) {
	C.ghostty_key_encoder_setopt_from_terminal(enc.ptr, t.ptr)
}

// Encode encodes a key event into a terminal escape sequence and
// returns the result as a byte slice. Not all key events produce
// output (e.g. unmodified modifier keys); in that case, a nil slice
// and nil error are returned.
func (enc *KeyEncoder) Encode(event *KeyEvent) ([]byte, error) {
	// Most escape sequences fit in 128 bytes. Try with a stack buffer
	// first; fall back to a larger heap allocation if needed.
	var buf [128]byte
	var outLen C.size_t
	result := C.ghostty_key_encoder_encode(
		enc.ptr,
		event.ptr,
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
		// outLen contains the required buffer size.
		dynBuf := make([]byte, outLen)
		var written C.size_t
		if err := resultError(C.ghostty_key_encoder_encode(
			enc.ptr,
			event.ptr,
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
