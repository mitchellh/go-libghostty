package libghostty

// Key event representation and manipulation.
// Wraps the C APIs from key/event.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// KeyEvent is an opaque handle representing a keyboard input event
// containing information about the physical key pressed, modifiers,
// and generated text. It is mutable and reusable, but not safe for
// concurrent use.
//
// C: GhosttyKeyEvent
type KeyEvent struct {
	ptr C.GhosttyKeyEvent
}

// KeyAction represents the type of keyboard input event (press, release,
// or repeat).
//
// C: GhosttyKeyAction
type KeyAction int

const (
	// KeyActionRelease indicates a key was released.
	KeyActionRelease KeyAction = C.GHOSTTY_KEY_ACTION_RELEASE

	// KeyActionPress indicates a key was pressed.
	KeyActionPress KeyAction = C.GHOSTTY_KEY_ACTION_PRESS

	// KeyActionRepeat indicates a key is being held down (repeat).
	KeyActionRepeat KeyAction = C.GHOSTTY_KEY_ACTION_REPEAT
)

// Mods is a bitmask representing keyboard modifier keys.
// Use the Mods* constants to test and set individual modifiers.
//
// C: GhosttyMods
type Mods uint16

const (
	// ModShift indicates the Shift key is pressed.
	ModShift Mods = C.GHOSTTY_MODS_SHIFT

	// ModCtrl indicates the Control key is pressed.
	ModCtrl Mods = C.GHOSTTY_MODS_CTRL

	// ModAlt indicates the Alt/Option key is pressed.
	ModAlt Mods = C.GHOSTTY_MODS_ALT

	// ModSuper indicates the Super/Command/Windows key is pressed.
	ModSuper Mods = C.GHOSTTY_MODS_SUPER

	// ModCapsLock indicates Caps Lock is active.
	ModCapsLock Mods = C.GHOSTTY_MODS_CAPS_LOCK

	// ModNumLock indicates Num Lock is active.
	ModNumLock Mods = C.GHOSTTY_MODS_NUM_LOCK

	// ModShiftSide indicates right Shift is pressed (0 = left, 1 = right).
	// Only meaningful when ModShift is set.
	ModShiftSide Mods = C.GHOSTTY_MODS_SHIFT_SIDE

	// ModCtrlSide indicates right Ctrl is pressed (0 = left, 1 = right).
	// Only meaningful when ModCtrl is set.
	ModCtrlSide Mods = C.GHOSTTY_MODS_CTRL_SIDE

	// ModAltSide indicates right Alt is pressed (0 = left, 1 = right).
	// Only meaningful when ModAlt is set.
	ModAltSide Mods = C.GHOSTTY_MODS_ALT_SIDE

	// ModSuperSide indicates right Super is pressed (0 = left, 1 = right).
	// Only meaningful when ModSuper is set.
	ModSuperSide Mods = C.GHOSTTY_MODS_SUPER_SIDE
)

// Key represents a physical key code. These are layout-independent and
// based on the W3C UI Events KeyboardEvent code standard.
//
// C: GhosttyKey
type Key int

// Writing System Keys (W3C § 3.1.1)
const (
	KeyUnidentified  Key = C.GHOSTTY_KEY_UNIDENTIFIED
	KeyBackquote     Key = C.GHOSTTY_KEY_BACKQUOTE
	KeyBackslash     Key = C.GHOSTTY_KEY_BACKSLASH
	KeyBracketLeft   Key = C.GHOSTTY_KEY_BRACKET_LEFT
	KeyBracketRight  Key = C.GHOSTTY_KEY_BRACKET_RIGHT
	KeyComma         Key = C.GHOSTTY_KEY_COMMA
	KeyDigit0        Key = C.GHOSTTY_KEY_DIGIT_0
	KeyDigit1        Key = C.GHOSTTY_KEY_DIGIT_1
	KeyDigit2        Key = C.GHOSTTY_KEY_DIGIT_2
	KeyDigit3        Key = C.GHOSTTY_KEY_DIGIT_3
	KeyDigit4        Key = C.GHOSTTY_KEY_DIGIT_4
	KeyDigit5        Key = C.GHOSTTY_KEY_DIGIT_5
	KeyDigit6        Key = C.GHOSTTY_KEY_DIGIT_6
	KeyDigit7        Key = C.GHOSTTY_KEY_DIGIT_7
	KeyDigit8        Key = C.GHOSTTY_KEY_DIGIT_8
	KeyDigit9        Key = C.GHOSTTY_KEY_DIGIT_9
	KeyEqual         Key = C.GHOSTTY_KEY_EQUAL
	KeyIntlBackslash Key = C.GHOSTTY_KEY_INTL_BACKSLASH
	KeyIntlRo        Key = C.GHOSTTY_KEY_INTL_RO
	KeyIntlYen       Key = C.GHOSTTY_KEY_INTL_YEN
	KeyA             Key = C.GHOSTTY_KEY_A
	KeyB             Key = C.GHOSTTY_KEY_B
	KeyC             Key = C.GHOSTTY_KEY_C
	KeyD             Key = C.GHOSTTY_KEY_D
	KeyE             Key = C.GHOSTTY_KEY_E
	KeyF             Key = C.GHOSTTY_KEY_F
	KeyG             Key = C.GHOSTTY_KEY_G
	KeyH             Key = C.GHOSTTY_KEY_H
	KeyI             Key = C.GHOSTTY_KEY_I
	KeyJ             Key = C.GHOSTTY_KEY_J
	KeyK             Key = C.GHOSTTY_KEY_K
	KeyL             Key = C.GHOSTTY_KEY_L
	KeyM             Key = C.GHOSTTY_KEY_M
	KeyN             Key = C.GHOSTTY_KEY_N
	KeyO             Key = C.GHOSTTY_KEY_O
	KeyP             Key = C.GHOSTTY_KEY_P
	KeyQ             Key = C.GHOSTTY_KEY_Q
	KeyR             Key = C.GHOSTTY_KEY_R
	KeyS             Key = C.GHOSTTY_KEY_S
	KeyT             Key = C.GHOSTTY_KEY_T
	KeyU             Key = C.GHOSTTY_KEY_U
	KeyV             Key = C.GHOSTTY_KEY_V
	KeyW             Key = C.GHOSTTY_KEY_W
	KeyX             Key = C.GHOSTTY_KEY_X
	KeyY             Key = C.GHOSTTY_KEY_Y
	KeyZ             Key = C.GHOSTTY_KEY_Z
	KeyMinus         Key = C.GHOSTTY_KEY_MINUS
	KeyPeriod        Key = C.GHOSTTY_KEY_PERIOD
	KeyQuote         Key = C.GHOSTTY_KEY_QUOTE
	KeySemicolon     Key = C.GHOSTTY_KEY_SEMICOLON
	KeySlash         Key = C.GHOSTTY_KEY_SLASH
)

// Functional Keys (W3C § 3.1.2)
const (
	KeyAltLeft      Key = C.GHOSTTY_KEY_ALT_LEFT
	KeyAltRight     Key = C.GHOSTTY_KEY_ALT_RIGHT
	KeyBackspace    Key = C.GHOSTTY_KEY_BACKSPACE
	KeyCapsLock     Key = C.GHOSTTY_KEY_CAPS_LOCK
	KeyContextMenu  Key = C.GHOSTTY_KEY_CONTEXT_MENU
	KeyControlLeft  Key = C.GHOSTTY_KEY_CONTROL_LEFT
	KeyControlRight Key = C.GHOSTTY_KEY_CONTROL_RIGHT
	KeyEnter        Key = C.GHOSTTY_KEY_ENTER
	KeyMetaLeft     Key = C.GHOSTTY_KEY_META_LEFT
	KeyMetaRight    Key = C.GHOSTTY_KEY_META_RIGHT
	KeyShiftLeft    Key = C.GHOSTTY_KEY_SHIFT_LEFT
	KeyShiftRight   Key = C.GHOSTTY_KEY_SHIFT_RIGHT
	KeySpace        Key = C.GHOSTTY_KEY_SPACE
	KeyTab          Key = C.GHOSTTY_KEY_TAB
	KeyConvert      Key = C.GHOSTTY_KEY_CONVERT
	KeyKanaMode     Key = C.GHOSTTY_KEY_KANA_MODE
	KeyNonConvert   Key = C.GHOSTTY_KEY_NON_CONVERT
)

// Control Pad Section (W3C § 3.2)
const (
	KeyDelete   Key = C.GHOSTTY_KEY_DELETE
	KeyEnd      Key = C.GHOSTTY_KEY_END
	KeyHelp     Key = C.GHOSTTY_KEY_HELP
	KeyHome     Key = C.GHOSTTY_KEY_HOME
	KeyInsert   Key = C.GHOSTTY_KEY_INSERT
	KeyPageDown Key = C.GHOSTTY_KEY_PAGE_DOWN
	KeyPageUp   Key = C.GHOSTTY_KEY_PAGE_UP
)

// Arrow Pad Section (W3C § 3.3)
const (
	KeyArrowDown  Key = C.GHOSTTY_KEY_ARROW_DOWN
	KeyArrowLeft  Key = C.GHOSTTY_KEY_ARROW_LEFT
	KeyArrowRight Key = C.GHOSTTY_KEY_ARROW_RIGHT
	KeyArrowUp    Key = C.GHOSTTY_KEY_ARROW_UP
)

// Numpad Section (W3C § 3.4)
const (
	KeyNumLock            Key = C.GHOSTTY_KEY_NUM_LOCK
	KeyNumpad0            Key = C.GHOSTTY_KEY_NUMPAD_0
	KeyNumpad1            Key = C.GHOSTTY_KEY_NUMPAD_1
	KeyNumpad2            Key = C.GHOSTTY_KEY_NUMPAD_2
	KeyNumpad3            Key = C.GHOSTTY_KEY_NUMPAD_3
	KeyNumpad4            Key = C.GHOSTTY_KEY_NUMPAD_4
	KeyNumpad5            Key = C.GHOSTTY_KEY_NUMPAD_5
	KeyNumpad6            Key = C.GHOSTTY_KEY_NUMPAD_6
	KeyNumpad7            Key = C.GHOSTTY_KEY_NUMPAD_7
	KeyNumpad8            Key = C.GHOSTTY_KEY_NUMPAD_8
	KeyNumpad9            Key = C.GHOSTTY_KEY_NUMPAD_9
	KeyNumpadAdd          Key = C.GHOSTTY_KEY_NUMPAD_ADD
	KeyNumpadBackspace    Key = C.GHOSTTY_KEY_NUMPAD_BACKSPACE
	KeyNumpadClear        Key = C.GHOSTTY_KEY_NUMPAD_CLEAR
	KeyNumpadClearEntry   Key = C.GHOSTTY_KEY_NUMPAD_CLEAR_ENTRY
	KeyNumpadComma        Key = C.GHOSTTY_KEY_NUMPAD_COMMA
	KeyNumpadDecimal      Key = C.GHOSTTY_KEY_NUMPAD_DECIMAL
	KeyNumpadDivide       Key = C.GHOSTTY_KEY_NUMPAD_DIVIDE
	KeyNumpadEnter        Key = C.GHOSTTY_KEY_NUMPAD_ENTER
	KeyNumpadEqual        Key = C.GHOSTTY_KEY_NUMPAD_EQUAL
	KeyNumpadMemoryAdd    Key = C.GHOSTTY_KEY_NUMPAD_MEMORY_ADD
	KeyNumpadMemoryClear  Key = C.GHOSTTY_KEY_NUMPAD_MEMORY_CLEAR
	KeyNumpadMemoryRecall Key = C.GHOSTTY_KEY_NUMPAD_MEMORY_RECALL
	KeyNumpadMemoryStore  Key = C.GHOSTTY_KEY_NUMPAD_MEMORY_STORE
	KeyNumpadMemorySub    Key = C.GHOSTTY_KEY_NUMPAD_MEMORY_SUBTRACT
	KeyNumpadMultiply     Key = C.GHOSTTY_KEY_NUMPAD_MULTIPLY
	KeyNumpadParenLeft    Key = C.GHOSTTY_KEY_NUMPAD_PAREN_LEFT
	KeyNumpadParenRight   Key = C.GHOSTTY_KEY_NUMPAD_PAREN_RIGHT
	KeyNumpadSubtract     Key = C.GHOSTTY_KEY_NUMPAD_SUBTRACT
	KeyNumpadSeparator    Key = C.GHOSTTY_KEY_NUMPAD_SEPARATOR
	KeyNumpadUp           Key = C.GHOSTTY_KEY_NUMPAD_UP
	KeyNumpadDown         Key = C.GHOSTTY_KEY_NUMPAD_DOWN
	KeyNumpadRight        Key = C.GHOSTTY_KEY_NUMPAD_RIGHT
	KeyNumpadLeft         Key = C.GHOSTTY_KEY_NUMPAD_LEFT
	KeyNumpadBegin        Key = C.GHOSTTY_KEY_NUMPAD_BEGIN
	KeyNumpadHome         Key = C.GHOSTTY_KEY_NUMPAD_HOME
	KeyNumpadEnd          Key = C.GHOSTTY_KEY_NUMPAD_END
	KeyNumpadInsert       Key = C.GHOSTTY_KEY_NUMPAD_INSERT
	KeyNumpadDelete       Key = C.GHOSTTY_KEY_NUMPAD_DELETE
	KeyNumpadPageUp       Key = C.GHOSTTY_KEY_NUMPAD_PAGE_UP
	KeyNumpadPageDown     Key = C.GHOSTTY_KEY_NUMPAD_PAGE_DOWN
)

// Function Section (W3C § 3.5)
const (
	KeyEscape      Key = C.GHOSTTY_KEY_ESCAPE
	KeyF1          Key = C.GHOSTTY_KEY_F1
	KeyF2          Key = C.GHOSTTY_KEY_F2
	KeyF3          Key = C.GHOSTTY_KEY_F3
	KeyF4          Key = C.GHOSTTY_KEY_F4
	KeyF5          Key = C.GHOSTTY_KEY_F5
	KeyF6          Key = C.GHOSTTY_KEY_F6
	KeyF7          Key = C.GHOSTTY_KEY_F7
	KeyF8          Key = C.GHOSTTY_KEY_F8
	KeyF9          Key = C.GHOSTTY_KEY_F9
	KeyF10         Key = C.GHOSTTY_KEY_F10
	KeyF11         Key = C.GHOSTTY_KEY_F11
	KeyF12         Key = C.GHOSTTY_KEY_F12
	KeyF13         Key = C.GHOSTTY_KEY_F13
	KeyF14         Key = C.GHOSTTY_KEY_F14
	KeyF15         Key = C.GHOSTTY_KEY_F15
	KeyF16         Key = C.GHOSTTY_KEY_F16
	KeyF17         Key = C.GHOSTTY_KEY_F17
	KeyF18         Key = C.GHOSTTY_KEY_F18
	KeyF19         Key = C.GHOSTTY_KEY_F19
	KeyF20         Key = C.GHOSTTY_KEY_F20
	KeyF21         Key = C.GHOSTTY_KEY_F21
	KeyF22         Key = C.GHOSTTY_KEY_F22
	KeyF23         Key = C.GHOSTTY_KEY_F23
	KeyF24         Key = C.GHOSTTY_KEY_F24
	KeyF25         Key = C.GHOSTTY_KEY_F25
	KeyFn          Key = C.GHOSTTY_KEY_FN
	KeyFnLock      Key = C.GHOSTTY_KEY_FN_LOCK
	KeyPrintScreen Key = C.GHOSTTY_KEY_PRINT_SCREEN
	KeyScrollLock  Key = C.GHOSTTY_KEY_SCROLL_LOCK
	KeyPause       Key = C.GHOSTTY_KEY_PAUSE
)

// Media Keys (W3C § 3.6)
const (
	KeyBrowserBack        Key = C.GHOSTTY_KEY_BROWSER_BACK
	KeyBrowserFavorites   Key = C.GHOSTTY_KEY_BROWSER_FAVORITES
	KeyBrowserForward     Key = C.GHOSTTY_KEY_BROWSER_FORWARD
	KeyBrowserHome        Key = C.GHOSTTY_KEY_BROWSER_HOME
	KeyBrowserRefresh     Key = C.GHOSTTY_KEY_BROWSER_REFRESH
	KeyBrowserSearch      Key = C.GHOSTTY_KEY_BROWSER_SEARCH
	KeyBrowserStop        Key = C.GHOSTTY_KEY_BROWSER_STOP
	KeyEject              Key = C.GHOSTTY_KEY_EJECT
	KeyLaunchApp1         Key = C.GHOSTTY_KEY_LAUNCH_APP_1
	KeyLaunchApp2         Key = C.GHOSTTY_KEY_LAUNCH_APP_2
	KeyLaunchMail         Key = C.GHOSTTY_KEY_LAUNCH_MAIL
	KeyMediaPlayPause     Key = C.GHOSTTY_KEY_MEDIA_PLAY_PAUSE
	KeyMediaSelect        Key = C.GHOSTTY_KEY_MEDIA_SELECT
	KeyMediaStop          Key = C.GHOSTTY_KEY_MEDIA_STOP
	KeyMediaTrackNext     Key = C.GHOSTTY_KEY_MEDIA_TRACK_NEXT
	KeyMediaTrackPrevious Key = C.GHOSTTY_KEY_MEDIA_TRACK_PREVIOUS
	KeyPower              Key = C.GHOSTTY_KEY_POWER
	KeySleep              Key = C.GHOSTTY_KEY_SLEEP
	KeyAudioVolumeDown    Key = C.GHOSTTY_KEY_AUDIO_VOLUME_DOWN
	KeyAudioVolumeMute    Key = C.GHOSTTY_KEY_AUDIO_VOLUME_MUTE
	KeyAudioVolumeUp      Key = C.GHOSTTY_KEY_AUDIO_VOLUME_UP
	KeyWakeUp             Key = C.GHOSTTY_KEY_WAKE_UP
)

// Legacy, Non-standard, and Special Keys (W3C § 3.7)
const (
	KeyCopy  Key = C.GHOSTTY_KEY_COPY
	KeyCut   Key = C.GHOSTTY_KEY_CUT
	KeyPaste Key = C.GHOSTTY_KEY_PASTE
)

// NewKeyEvent creates a new key event with default values. The event
// must be freed with Close when no longer needed.
func NewKeyEvent() (*KeyEvent, error) {
	var ptr C.GhosttyKeyEvent
	if err := resultError(C.ghostty_key_event_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &KeyEvent{ptr: ptr}, nil
}

// Close frees the underlying key event handle. After this call, the
// key event must not be used.
func (e *KeyEvent) Close() {
	C.ghostty_key_event_free(e.ptr)
}

// SetAction sets the key action (press, release, repeat).
func (e *KeyEvent) SetAction(action KeyAction) {
	C.ghostty_key_event_set_action(e.ptr, C.GhosttyKeyAction(action))
}

// Action returns the key action (press, release, repeat).
func (e *KeyEvent) Action() KeyAction {
	return KeyAction(C.ghostty_key_event_get_action(e.ptr))
}

// SetKey sets the physical key code.
func (e *KeyEvent) SetKey(key Key) {
	C.ghostty_key_event_set_key(e.ptr, C.GhosttyKey(key))
}

// Key returns the physical key code.
func (e *KeyEvent) Key() Key {
	return Key(C.ghostty_key_event_get_key(e.ptr))
}

// SetMods sets the modifier keys bitmask.
func (e *KeyEvent) SetMods(mods Mods) {
	C.ghostty_key_event_set_mods(e.ptr, C.GhosttyMods(mods))
}

// Mods returns the modifier keys bitmask.
func (e *KeyEvent) Mods() Mods {
	return Mods(C.ghostty_key_event_get_mods(e.ptr))
}

// SetConsumedMods sets the consumed modifiers bitmask.
func (e *KeyEvent) SetConsumedMods(mods Mods) {
	C.ghostty_key_event_set_consumed_mods(e.ptr, C.GhosttyMods(mods))
}

// ConsumedMods returns the consumed modifiers bitmask.
func (e *KeyEvent) ConsumedMods() Mods {
	return Mods(C.ghostty_key_event_get_consumed_mods(e.ptr))
}

// SetComposing sets whether the key event is part of a composition sequence.
func (e *KeyEvent) SetComposing(composing bool) {
	C.ghostty_key_event_set_composing(e.ptr, C.bool(composing))
}

// Composing reports whether the key event is part of a composition sequence.
func (e *KeyEvent) Composing() bool {
	return bool(C.ghostty_key_event_get_composing(e.ptr))
}

// SetUTF8 sets the UTF-8 text generated by the key for the current
// keyboard layout. Must contain the unmodified character before any
// Ctrl/Meta transformations. Do not pass C0 control characters or
// platform function key codes; pass an empty string instead and let
// the encoder use the logical key.
//
// The string is not copied by the key event. The caller must ensure
// the string remains valid for the lifetime needed by the event.
func (e *KeyEvent) SetUTF8(s string) {
	if len(s) == 0 {
		C.ghostty_key_event_set_utf8(e.ptr, nil, 0)
		return
	}
	C.ghostty_key_event_set_utf8(e.ptr, (*C.char)(unsafe.Pointer(unsafe.StringData(s))), C.size_t(len(s)))
}

// UTF8 returns the UTF-8 text generated by the key event.
func (e *KeyEvent) UTF8() string {
	var n C.size_t
	p := C.ghostty_key_event_get_utf8(e.ptr, &n)
	if p == nil || n == 0 {
		return ""
	}
	return C.GoStringN(p, C.int(n))
}

// SetUnshiftedCodepoint sets the unshifted Unicode codepoint.
func (e *KeyEvent) SetUnshiftedCodepoint(cp rune) {
	C.ghostty_key_event_set_unshifted_codepoint(e.ptr, C.uint32_t(cp))
}

// UnshiftedCodepoint returns the unshifted Unicode codepoint.
func (e *KeyEvent) UnshiftedCodepoint() rune {
	return rune(C.ghostty_key_event_get_unshifted_codepoint(e.ptr))
}
