package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// Mode is a packed 16-bit terminal mode identifier. It encodes a mode
// value (bits 0–14) and an ANSI flag (bit 15). DEC private modes have
// the ANSI bit clear; standard ANSI modes have it set.
// C: GhosttyMode
type Mode uint16

// NewMode packs a numeric mode value and ANSI flag into a Mode. Values are
// limited to 15 bits, matching GhosttyMode.
func NewMode(value uint16, ansi bool) Mode {
	return Mode(C.ghostty_mode_new(C.uint16_t(value), C.bool(ansi)))
}

// Value returns the numeric mode value (0–32767).
func (m Mode) Value() uint16 {
	return uint16(C.ghostty_mode_value(C.GhosttyMode(m)))
}

// ANSI reports whether this is a standard ANSI mode. If false, it is
// a DEC private mode (?-prefixed).
func (m Mode) ANSI() bool {
	return bool(C.ghostty_mode_ansi(C.GhosttyMode(m)))
}

// ANSI modes.
var (
	// ModeDisableKeyboard is keyboard action mode (disable keyboard).
	ModeDisableKeyboard = Mode(C.GHOSTTY_MODE_DISABLE_KEYBOARD)

	// ModeInsert is insert mode.
	ModeInsert = Mode(C.GHOSTTY_MODE_INSERT)

	// ModeSendReceiveMode is send/receive mode.
	ModeSendReceiveMode = Mode(C.GHOSTTY_MODE_SEND_RECEIVE_MODE)

	// ModeLinefeed is linefeed/new line mode.
	ModeLinefeed = Mode(C.GHOSTTY_MODE_LINEFEED)
)

// DEC private modes.
var (
	// ModeCursorKeys is cursor keys mode.
	ModeCursorKeys = Mode(C.GHOSTTY_MODE_CURSOR_KEYS)

	// Mode132Column is 132/80 column mode.
	Mode132Column = Mode(C.GHOSTTY_MODE_132_COLUMN)

	// ModeSlowScroll is slow scroll mode.
	ModeSlowScroll = Mode(C.GHOSTTY_MODE_SLOW_SCROLL)

	// ModeReverseColors is reverse video mode.
	ModeReverseColors = Mode(C.GHOSTTY_MODE_REVERSE_COLORS)

	// ModeOrigin is origin mode.
	ModeOrigin = Mode(C.GHOSTTY_MODE_ORIGIN)

	// ModeWraparound is auto-wrap mode.
	ModeWraparound = Mode(C.GHOSTTY_MODE_WRAPAROUND)

	// ModeAutorepeat is auto-repeat keys mode.
	ModeAutorepeat = Mode(C.GHOSTTY_MODE_AUTOREPEAT)

	// ModeMouseEventX10 is X10 mouse reporting mode.
	ModeMouseEventX10 = Mode(C.GHOSTTY_MODE_MOUSE_EVENT_X10)

	// ModeCursorBlinking is cursor blink mode.
	ModeCursorBlinking = Mode(C.GHOSTTY_MODE_CURSOR_BLINKING)

	// ModeCursorVisible is cursor visible mode (DECTCEM).
	ModeCursorVisible = Mode(C.GHOSTTY_MODE_CURSOR_VISIBLE)

	// ModeEnableMode3 allows 132 column mode.
	ModeEnableMode3 = Mode(C.GHOSTTY_MODE_ENABLE_MODE_3)

	// ModeReverseWrap is reverse wrap mode.
	ModeReverseWrap = Mode(C.GHOSTTY_MODE_REVERSE_WRAP)

	// ModeAltScreenLegacy is alternate screen (legacy, mode 47).
	ModeAltScreenLegacy = Mode(C.GHOSTTY_MODE_ALT_SCREEN_LEGACY)

	// ModeKeypadKeys is application keypad mode.
	ModeKeypadKeys = Mode(C.GHOSTTY_MODE_KEYPAD_KEYS)

	// ModeBackarrowKeyMode is backarrow key mode (DECBKM).
	ModeBackarrowKeyMode = Mode(C.GHOSTTY_MODE_BACKARROW_KEY_MODE)

	// ModeEnableLeftAndRightMargin is left/right margin mode.
	ModeEnableLeftAndRightMargin = Mode(C.GHOSTTY_MODE_ENABLE_LEFT_AND_RIGHT_MARGIN)

	// ModeMouseEventNormal is normal mouse tracking mode.
	ModeMouseEventNormal = Mode(C.GHOSTTY_MODE_MOUSE_EVENT_NORMAL)

	// ModeMouseEventButton is button-event mouse tracking mode.
	ModeMouseEventButton = Mode(C.GHOSTTY_MODE_MOUSE_EVENT_BUTTON)

	// ModeMouseEventAny is any-event mouse tracking mode.
	ModeMouseEventAny = Mode(C.GHOSTTY_MODE_MOUSE_EVENT_ANY)

	// ModeFocusEvent enables focus in/out events.
	ModeFocusEvent = Mode(C.GHOSTTY_MODE_FOCUS_EVENT)

	// ModeMouseFormatUTF8 is UTF-8 mouse format mode.
	ModeMouseFormatUTF8 = Mode(C.GHOSTTY_MODE_MOUSE_FORMAT_UTF8)

	// ModeMouseFormatSGR is SGR mouse format mode.
	ModeMouseFormatSGR = Mode(C.GHOSTTY_MODE_MOUSE_FORMAT_SGR)

	// ModeMouseAlternateScroll is alternate scroll mode.
	ModeMouseAlternateScroll = Mode(C.GHOSTTY_MODE_MOUSE_ALTERNATE_SCROLL)

	// ModeMouseFormatURxvt is URxvt mouse format mode.
	ModeMouseFormatURxvt = Mode(C.GHOSTTY_MODE_MOUSE_FORMAT_URXVT)

	// ModeMouseFormatSGRPixels is SGR-Pixels mouse format mode.
	ModeMouseFormatSGRPixels = Mode(C.GHOSTTY_MODE_MOUSE_FORMAT_SGR_PIXELS)

	// ModeIgnoreKeypadWithNumlock ignores keypad with NumLock.
	ModeIgnoreKeypadWithNumlock = Mode(C.GHOSTTY_MODE_IGNORE_KEYPAD_WITH_NUMLOCK)

	// ModeAltEscPrefix makes Alt key send ESC prefix.
	ModeAltEscPrefix = Mode(C.GHOSTTY_MODE_ALT_ESC_PREFIX)

	// ModeAltSendsEscape makes Alt send escape.
	ModeAltSendsEscape = Mode(C.GHOSTTY_MODE_ALT_SENDS_ESCAPE)

	// ModeReverseWrapExtended is extended reverse wrap mode.
	ModeReverseWrapExtended = Mode(C.GHOSTTY_MODE_REVERSE_WRAP_EXTENDED)

	// ModeAltScreen is alternate screen mode (mode 1047).
	ModeAltScreen = Mode(C.GHOSTTY_MODE_ALT_SCREEN)

	// ModeSaveCursor saves cursor position (DECSC, mode 1048).
	ModeSaveCursor = Mode(C.GHOSTTY_MODE_SAVE_CURSOR)

	// ModeAltScreenSaveCursorClearEnter is alt screen + save cursor + clear (mode 1049).
	ModeAltScreenSaveCursorClearEnter = Mode(C.GHOSTTY_MODE_ALT_SCREEN_SAVE_CURSOR_CLEAR_ENTER)

	// ModeBracketedPaste is bracketed paste mode.
	ModeBracketedPaste = Mode(C.GHOSTTY_MODE_BRACKETED_PASTE)

	// ModeSynchronizedOutput is synchronized output mode.
	ModeSynchronizedOutput = Mode(C.GHOSTTY_MODE_SYNCHRONIZED_OUTPUT)

	// ModeGraphemeCluster is grapheme cluster mode.
	ModeGraphemeCluster = Mode(C.GHOSTTY_MODE_GRAPHEME_CLUSTER)

	// ModeReportColorScheme enables color scheme reporting.
	ModeReportColorScheme = Mode(C.GHOSTTY_MODE_REPORT_COLOR_SCHEME)

	// ModeReportVisibility enables terminal visibility reporting.
	ModeReportVisibility = Mode(C.GHOSTTY_MODE_REPORT_VISIBILITY)

	// ModeInBandSizeReports enables in-band size reports.
	ModeInBandSizeReports = Mode(C.GHOSTTY_MODE_IN_BAND_SIZE_REPORTS)

	// ModeKittyPasteEvents enables Kitty clipboard protocol paste events.
	ModeKittyPasteEvents = Mode(C.GHOSTTY_MODE_KITTY_PASTE_EVENTS)
)

// ModeReportState represents DECRPM report state values (Ps2 parameter).
// C: GhosttyModeReportState
type ModeReportState int

const (
	// ModeReportNotRecognized means the mode is not recognized.
	ModeReportNotRecognized ModeReportState = C.GHOSTTY_MODE_REPORT_NOT_RECOGNIZED

	// ModeReportSet means the mode is set (enabled).
	ModeReportSet ModeReportState = C.GHOSTTY_MODE_REPORT_SET

	// ModeReportReset means the mode is reset (disabled).
	ModeReportReset ModeReportState = C.GHOSTTY_MODE_REPORT_RESET

	// ModeReportPermanentlySet means the mode is permanently set.
	ModeReportPermanentlySet ModeReportState = C.GHOSTTY_MODE_REPORT_PERMANENTLY_SET

	// ModeReportPermanentlyReset means the mode is permanently reset.
	ModeReportPermanentlyReset ModeReportState = C.GHOSTTY_MODE_REPORT_PERMANENTLY_RESET
)

// ModeReportEncode encodes a DECRPM response for mode and state.
func ModeReportEncode(mode Mode, state ModeReportState) ([]byte, error) {
	var buf [32]byte
	var written C.size_t
	result := C.ghostty_mode_report_encode(
		C.GhosttyMode(mode),
		C.GhosttyModeReportState(state),
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
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
	if err := resultError(C.ghostty_mode_report_encode(
		C.GhosttyMode(mode),
		C.GhosttyModeReportState(state),
		(*C.uint8_t)(unsafe.Pointer(&out[0])),
		C.size_t(len(out)),
		&outWritten,
	)); err != nil {
		return nil, err
	}
	return out[:int(outWritten)], nil
}
