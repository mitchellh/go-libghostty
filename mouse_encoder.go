package libghostty

// Mouse encoder — encodes mouse events into terminal escape sequences.
// Wraps the C APIs from mouse/encoder.h.

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyMouseEncoderSize (sized struct).
static inline GhosttyMouseEncoderSize init_mouse_encoder_size() {
	GhosttyMouseEncoderSize s = {0};
	s.size = sizeof(GhosttyMouseEncoderSize);
	return s;
}
*/
import "C"

import "unsafe"

// MouseEncoder encodes mouse events into terminal escape sequences,
// supporting X10, UTF-8, SGR, URxvt, and SGR-Pixels mouse protocols.
// It maintains mutable encoder state and is not safe for concurrent
// use.
//
// Basic usage:
//  1. Create an encoder with NewMouseEncoder.
//  2. Configure options with SetOpt* methods or SetOptFromTerminal.
//  3. Create mouse events, encode them with Encode, and free them.
//  4. Free the encoder with Close when done.
//
// C: GhosttyMouseEncoder
type MouseEncoder struct {
	ptr C.GhosttyMouseEncoder
}

// MouseTrackingMode selects which mouse events the terminal tracks.
//
// C: GhosttyMouseTrackingMode
type MouseTrackingMode int

const (
	// MouseTrackingNone disables mouse reporting.
	MouseTrackingNone MouseTrackingMode = C.GHOSTTY_MOUSE_TRACKING_NONE

	// MouseTrackingX10 enables X10 mouse mode.
	MouseTrackingX10 MouseTrackingMode = C.GHOSTTY_MOUSE_TRACKING_X10

	// MouseTrackingNormal enables normal mouse mode (button press/release only).
	MouseTrackingNormal MouseTrackingMode = C.GHOSTTY_MOUSE_TRACKING_NORMAL

	// MouseTrackingButton enables button-event tracking mode.
	MouseTrackingButton MouseTrackingMode = C.GHOSTTY_MOUSE_TRACKING_BUTTON

	// MouseTrackingAny enables any-event tracking mode.
	MouseTrackingAny MouseTrackingMode = C.GHOSTTY_MOUSE_TRACKING_ANY
)

// MouseFormat selects the wire format for mouse escape sequences.
//
// C: GhosttyMouseFormat
type MouseFormat int

const (
	MouseFormatX10       MouseFormat = C.GHOSTTY_MOUSE_FORMAT_X10
	MouseFormatUTF8      MouseFormat = C.GHOSTTY_MOUSE_FORMAT_UTF8
	MouseFormatSGR       MouseFormat = C.GHOSTTY_MOUSE_FORMAT_SGR
	MouseFormatURxvt     MouseFormat = C.GHOSTTY_MOUSE_FORMAT_URXVT
	MouseFormatSGRPixels MouseFormat = C.GHOSTTY_MOUSE_FORMAT_SGR_PIXELS
)

// MouseEncoderSize describes the rendered terminal geometry used to
// convert surface-space positions into encoded coordinates.
//
// C: GhosttyMouseEncoderSize
type MouseEncoderSize struct {
	// ScreenWidth is the full screen width in pixels.
	ScreenWidth uint32

	// ScreenHeight is the full screen height in pixels.
	ScreenHeight uint32

	// CellWidth is the cell width in pixels. Must be non-zero.
	CellWidth uint32

	// CellHeight is the cell height in pixels. Must be non-zero.
	CellHeight uint32

	// PaddingTop is the top padding in pixels.
	PaddingTop uint32

	// PaddingBottom is the bottom padding in pixels.
	PaddingBottom uint32

	// PaddingRight is the right padding in pixels.
	PaddingRight uint32

	// PaddingLeft is the left padding in pixels.
	PaddingLeft uint32
}

// MouseEncoderOption identifies a mouse encoder configuration option
// for use with SetOpt.
//
// C: GhosttyMouseEncoderOption
type MouseEncoderOption int

const (
	// MouseEncoderOptEvent sets the mouse tracking mode
	// (value: MouseTrackingMode).
	MouseEncoderOptEvent MouseEncoderOption = C.GHOSTTY_MOUSE_ENCODER_OPT_EVENT

	// MouseEncoderOptFormat sets the mouse output format
	// (value: MouseFormat).
	MouseEncoderOptFormat MouseEncoderOption = C.GHOSTTY_MOUSE_ENCODER_OPT_FORMAT

	// MouseEncoderOptSize sets the renderer size context
	// (value: MouseEncoderSize).
	MouseEncoderOptSize MouseEncoderOption = C.GHOSTTY_MOUSE_ENCODER_OPT_SIZE

	// MouseEncoderOptAnyButtonPressed sets whether any mouse button
	// is currently pressed (value: bool).
	MouseEncoderOptAnyButtonPressed MouseEncoderOption = C.GHOSTTY_MOUSE_ENCODER_OPT_ANY_BUTTON_PRESSED

	// MouseEncoderOptTrackLastCell enables motion deduplication by
	// last cell (value: bool).
	MouseEncoderOptTrackLastCell MouseEncoderOption = C.GHOSTTY_MOUSE_ENCODER_OPT_TRACK_LAST_CELL
)

// NewMouseEncoder creates a new mouse encoder with default options.
// The encoder must be freed with Close when no longer needed.
func NewMouseEncoder() (*MouseEncoder, error) {
	var ptr C.GhosttyMouseEncoder
	if err := resultError(C.ghostty_mouse_encoder_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &MouseEncoder{ptr: ptr}, nil
}

// Close frees the underlying mouse encoder handle. After this call,
// the encoder must not be used.
func (enc *MouseEncoder) Close() {
	C.ghostty_mouse_encoder_free(enc.ptr)
}

// SetOptTrackingMode sets the mouse tracking mode on the encoder.
func (enc *MouseEncoder) SetOptTrackingMode(mode MouseTrackingMode) {
	v := C.GhosttyMouseTrackingMode(mode)
	C.ghostty_mouse_encoder_setopt(enc.ptr, C.GHOSTTY_MOUSE_ENCODER_OPT_EVENT, unsafe.Pointer(&v))
}

// SetOptFormat sets the mouse output format on the encoder.
func (enc *MouseEncoder) SetOptFormat(format MouseFormat) {
	v := C.GhosttyMouseFormat(format)
	C.ghostty_mouse_encoder_setopt(enc.ptr, C.GHOSTTY_MOUSE_ENCODER_OPT_FORMAT, unsafe.Pointer(&v))
}

// SetOptSize sets the renderer size context on the encoder.
func (enc *MouseEncoder) SetOptSize(s MouseEncoderSize) {
	cs := C.init_mouse_encoder_size()
	cs.screen_width = C.uint32_t(s.ScreenWidth)
	cs.screen_height = C.uint32_t(s.ScreenHeight)
	cs.cell_width = C.uint32_t(s.CellWidth)
	cs.cell_height = C.uint32_t(s.CellHeight)
	cs.padding_top = C.uint32_t(s.PaddingTop)
	cs.padding_bottom = C.uint32_t(s.PaddingBottom)
	cs.padding_right = C.uint32_t(s.PaddingRight)
	cs.padding_left = C.uint32_t(s.PaddingLeft)
	C.ghostty_mouse_encoder_setopt(enc.ptr, C.GHOSTTY_MOUSE_ENCODER_OPT_SIZE, unsafe.Pointer(&cs))
}

// SetOptAnyButtonPressed sets whether any mouse button is currently
// pressed.
func (enc *MouseEncoder) SetOptAnyButtonPressed(pressed bool) {
	v := C.bool(pressed)
	C.ghostty_mouse_encoder_setopt(enc.ptr, C.GHOSTTY_MOUSE_ENCODER_OPT_ANY_BUTTON_PRESSED, unsafe.Pointer(&v))
}

// SetOptTrackLastCell enables or disables motion deduplication by
// last cell.
func (enc *MouseEncoder) SetOptTrackLastCell(track bool) {
	v := C.bool(track)
	C.ghostty_mouse_encoder_setopt(enc.ptr, C.GHOSTTY_MOUSE_ENCODER_OPT_TRACK_LAST_CELL, unsafe.Pointer(&v))
}

// SetOptFromTerminal reads the terminal's current mouse tracking mode
// and output format and applies them to the encoder. It does not
// modify size or any-button state. The caller must serialize access to
// both the encoder and the terminal during this call.
func (enc *MouseEncoder) SetOptFromTerminal(t *Terminal) {
	C.ghostty_mouse_encoder_setopt_from_terminal(enc.ptr, t.ptr)
}

// Reset clears internal encoder state such as motion deduplication
// (last tracked cell).
func (enc *MouseEncoder) Reset() {
	C.ghostty_mouse_encoder_reset(enc.ptr)
}

// Encode encodes a mouse event into a terminal escape sequence and
// returns the result as a byte slice. Not all mouse events produce
// output; in that case, a nil slice and nil error are returned.
func (enc *MouseEncoder) Encode(event *MouseEvent) ([]byte, error) {
	// Most mouse escape sequences fit in 128 bytes. Try with a stack
	// buffer first; fall back to a larger heap allocation if needed.
	var buf [128]byte
	var outLen C.size_t
	result := C.ghostty_mouse_encoder_encode(
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
		if err := resultError(C.ghostty_mouse_encoder_encode(
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
