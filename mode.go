package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

// Mode is a packed 16-bit terminal mode identifier. It encodes a mode
// value (bits 0–14) and an ANSI flag (bit 15). DEC private modes have
// the ANSI bit clear; standard ANSI modes have it set.
// C: GhosttyMode
type Mode uint16

// Value returns the numeric mode value (0–32767).
func (m Mode) Value() uint16 {
	return uint16(m) & 0x7FFF
}

// ANSI reports whether this is a standard ANSI mode. If false, it is
// a DEC private mode (?-prefixed).
func (m Mode) ANSI() bool {
	return (m >> 15) != 0
}

// ANSI modes.
var (
	// ModeKAM is keyboard action mode (disable keyboard).
	ModeKAM = Mode(C.GHOSTTY_MODE_KAM)

	// ModeInsert is insert mode.
	ModeInsert = Mode(C.GHOSTTY_MODE_INSERT)

	// ModeSRM is send/receive mode.
	ModeSRM = Mode(C.GHOSTTY_MODE_SRM)

	// ModeLinefeed is linefeed/new line mode.
	ModeLinefeed = Mode(C.GHOSTTY_MODE_LINEFEED)
)

// DEC private modes.
var (
	// ModeDECCKM is cursor keys mode.
	ModeDECCKM = Mode(C.GHOSTTY_MODE_DECCKM)

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

	// ModeX10Mouse is X10 mouse reporting mode.
	ModeX10Mouse = Mode(C.GHOSTTY_MODE_X10_MOUSE)

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

	// ModeLeftRightMargin is left/right margin mode.
	ModeLeftRightMargin = Mode(C.GHOSTTY_MODE_LEFT_RIGHT_MARGIN)

	// ModeNormalMouse is normal mouse tracking mode.
	ModeNormalMouse = Mode(C.GHOSTTY_MODE_NORMAL_MOUSE)

	// ModeButtonMouse is button-event mouse tracking mode.
	ModeButtonMouse = Mode(C.GHOSTTY_MODE_BUTTON_MOUSE)

	// ModeAnyMouse is any-event mouse tracking mode.
	ModeAnyMouse = Mode(C.GHOSTTY_MODE_ANY_MOUSE)

	// ModeFocusEvent enables focus in/out events.
	ModeFocusEvent = Mode(C.GHOSTTY_MODE_FOCUS_EVENT)

	// ModeUTF8Mouse is UTF-8 mouse format mode.
	ModeUTF8Mouse = Mode(C.GHOSTTY_MODE_UTF8_MOUSE)

	// ModeSGRMouse is SGR mouse format mode.
	ModeSGRMouse = Mode(C.GHOSTTY_MODE_SGR_MOUSE)

	// ModeAltScroll is alternate scroll mode.
	ModeAltScroll = Mode(C.GHOSTTY_MODE_ALT_SCROLL)

	// ModeURxvtMouse is URxvt mouse format mode.
	ModeURxvtMouse = Mode(C.GHOSTTY_MODE_URXVT_MOUSE)

	// ModeSGRPixelsMouse is SGR-Pixels mouse format mode.
	ModeSGRPixelsMouse = Mode(C.GHOSTTY_MODE_SGR_PIXELS_MOUSE)

	// ModeNumlockKeypad ignores keypad with NumLock.
	ModeNumlockKeypad = Mode(C.GHOSTTY_MODE_NUMLOCK_KEYPAD)

	// ModeAltEscPrefix makes Alt key send ESC prefix.
	ModeAltEscPrefix = Mode(C.GHOSTTY_MODE_ALT_ESC_PREFIX)

	// ModeAltSendsEsc makes Alt send escape.
	ModeAltSendsEsc = Mode(C.GHOSTTY_MODE_ALT_SENDS_ESC)

	// ModeReverseWrapExt is extended reverse wrap mode.
	ModeReverseWrapExt = Mode(C.GHOSTTY_MODE_REVERSE_WRAP_EXT)

	// ModeAltScreen is alternate screen mode (mode 1047).
	ModeAltScreen = Mode(C.GHOSTTY_MODE_ALT_SCREEN)

	// ModeSaveCursor saves cursor position (DECSC, mode 1048).
	ModeSaveCursor = Mode(C.GHOSTTY_MODE_SAVE_CURSOR)

	// ModeAltScreenSave is alt screen + save cursor + clear (mode 1049).
	ModeAltScreenSave = Mode(C.GHOSTTY_MODE_ALT_SCREEN_SAVE)

	// ModeBracketedPaste is bracketed paste mode.
	ModeBracketedPaste = Mode(C.GHOSTTY_MODE_BRACKETED_PASTE)

	// ModeSyncOutput is synchronized output mode.
	ModeSyncOutput = Mode(C.GHOSTTY_MODE_SYNC_OUTPUT)

	// ModeGraphemeCluster is grapheme cluster mode.
	ModeGraphemeCluster = Mode(C.GHOSTTY_MODE_GRAPHEME_CLUSTER)

	// ModeColorSchemeReport enables color scheme reporting.
	ModeColorSchemeReport = Mode(C.GHOSTTY_MODE_COLOR_SCHEME_REPORT)

	// ModeInBandResize enables in-band size reports.
	ModeInBandResize = Mode(C.GHOSTTY_MODE_IN_BAND_RESIZE)
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
