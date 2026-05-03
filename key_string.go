package libghostty

import "fmt"

// Human-friendly string conversion for Key values. The names use
// snake_case and match the canonical names used by upstream
// libghostty's Zig source (e.g. "key_a", "arrow_down", "fn"). This
// makes them suitable for serialization in JSON, configuration files,
// or any other text-based format.

// keyNames is the canonical list mapping each Key constant to its
// human-friendly snake_case string name. It is used as the source of
// truth for both String and FromString and is built into reverse
// lookup maps in init.
var keyNames = []struct {
	key  Key
	name string
}{
	// Writing System Keys (W3C § 3.1.1)
	{KeyUnidentified, "unidentified"},
	{KeyBackquote, "backquote"},
	{KeyBackslash, "backslash"},
	{KeyBracketLeft, "bracket_left"},
	{KeyBracketRight, "bracket_right"},
	{KeyComma, "comma"},
	{KeyDigit0, "digit_0"},
	{KeyDigit1, "digit_1"},
	{KeyDigit2, "digit_2"},
	{KeyDigit3, "digit_3"},
	{KeyDigit4, "digit_4"},
	{KeyDigit5, "digit_5"},
	{KeyDigit6, "digit_6"},
	{KeyDigit7, "digit_7"},
	{KeyDigit8, "digit_8"},
	{KeyDigit9, "digit_9"},
	{KeyEqual, "equal"},
	{KeyIntlBackslash, "intl_backslash"},
	{KeyIntlRo, "intl_ro"},
	{KeyIntlYen, "intl_yen"},
	{KeyA, "key_a"},
	{KeyB, "key_b"},
	{KeyC, "key_c"},
	{KeyD, "key_d"},
	{KeyE, "key_e"},
	{KeyF, "key_f"},
	{KeyG, "key_g"},
	{KeyH, "key_h"},
	{KeyI, "key_i"},
	{KeyJ, "key_j"},
	{KeyK, "key_k"},
	{KeyL, "key_l"},
	{KeyM, "key_m"},
	{KeyN, "key_n"},
	{KeyO, "key_o"},
	{KeyP, "key_p"},
	{KeyQ, "key_q"},
	{KeyR, "key_r"},
	{KeyS, "key_s"},
	{KeyT, "key_t"},
	{KeyU, "key_u"},
	{KeyV, "key_v"},
	{KeyW, "key_w"},
	{KeyX, "key_x"},
	{KeyY, "key_y"},
	{KeyZ, "key_z"},
	{KeyMinus, "minus"},
	{KeyPeriod, "period"},
	{KeyQuote, "quote"},
	{KeySemicolon, "semicolon"},
	{KeySlash, "slash"},

	// Functional Keys (W3C § 3.1.2)
	{KeyAltLeft, "alt_left"},
	{KeyAltRight, "alt_right"},
	{KeyBackspace, "backspace"},
	{KeyCapsLock, "caps_lock"},
	{KeyContextMenu, "context_menu"},
	{KeyControlLeft, "control_left"},
	{KeyControlRight, "control_right"},
	{KeyEnter, "enter"},
	{KeyMetaLeft, "meta_left"},
	{KeyMetaRight, "meta_right"},
	{KeyShiftLeft, "shift_left"},
	{KeyShiftRight, "shift_right"},
	{KeySpace, "space"},
	{KeyTab, "tab"},
	{KeyConvert, "convert"},
	{KeyKanaMode, "kana_mode"},
	{KeyNonConvert, "non_convert"},

	// Control Pad Section (W3C § 3.2)
	{KeyDelete, "delete"},
	{KeyEnd, "end"},
	{KeyHelp, "help"},
	{KeyHome, "home"},
	{KeyInsert, "insert"},
	{KeyPageDown, "page_down"},
	{KeyPageUp, "page_up"},

	// Arrow Pad Section (W3C § 3.3)
	{KeyArrowDown, "arrow_down"},
	{KeyArrowLeft, "arrow_left"},
	{KeyArrowRight, "arrow_right"},
	{KeyArrowUp, "arrow_up"},

	// Numpad Section (W3C § 3.4)
	{KeyNumLock, "num_lock"},
	{KeyNumpad0, "numpad_0"},
	{KeyNumpad1, "numpad_1"},
	{KeyNumpad2, "numpad_2"},
	{KeyNumpad3, "numpad_3"},
	{KeyNumpad4, "numpad_4"},
	{KeyNumpad5, "numpad_5"},
	{KeyNumpad6, "numpad_6"},
	{KeyNumpad7, "numpad_7"},
	{KeyNumpad8, "numpad_8"},
	{KeyNumpad9, "numpad_9"},
	{KeyNumpadAdd, "numpad_add"},
	{KeyNumpadBackspace, "numpad_backspace"},
	{KeyNumpadClear, "numpad_clear"},
	{KeyNumpadClearEntry, "numpad_clear_entry"},
	{KeyNumpadComma, "numpad_comma"},
	{KeyNumpadDecimal, "numpad_decimal"},
	{KeyNumpadDivide, "numpad_divide"},
	{KeyNumpadEnter, "numpad_enter"},
	{KeyNumpadEqual, "numpad_equal"},
	{KeyNumpadMemoryAdd, "numpad_memory_add"},
	{KeyNumpadMemoryClear, "numpad_memory_clear"},
	{KeyNumpadMemoryRecall, "numpad_memory_recall"},
	{KeyNumpadMemoryStore, "numpad_memory_store"},
	{KeyNumpadMemorySub, "numpad_memory_subtract"},
	{KeyNumpadMultiply, "numpad_multiply"},
	{KeyNumpadParenLeft, "numpad_paren_left"},
	{KeyNumpadParenRight, "numpad_paren_right"},
	{KeyNumpadSubtract, "numpad_subtract"},
	{KeyNumpadSeparator, "numpad_separator"},
	{KeyNumpadUp, "numpad_up"},
	{KeyNumpadDown, "numpad_down"},
	{KeyNumpadRight, "numpad_right"},
	{KeyNumpadLeft, "numpad_left"},
	{KeyNumpadBegin, "numpad_begin"},
	{KeyNumpadHome, "numpad_home"},
	{KeyNumpadEnd, "numpad_end"},
	{KeyNumpadInsert, "numpad_insert"},
	{KeyNumpadDelete, "numpad_delete"},
	{KeyNumpadPageUp, "numpad_page_up"},
	{KeyNumpadPageDown, "numpad_page_down"},

	// Function Section (W3C § 3.5)
	{KeyEscape, "escape"},
	{KeyF1, "f1"},
	{KeyF2, "f2"},
	{KeyF3, "f3"},
	{KeyF4, "f4"},
	{KeyF5, "f5"},
	{KeyF6, "f6"},
	{KeyF7, "f7"},
	{KeyF8, "f8"},
	{KeyF9, "f9"},
	{KeyF10, "f10"},
	{KeyF11, "f11"},
	{KeyF12, "f12"},
	{KeyF13, "f13"},
	{KeyF14, "f14"},
	{KeyF15, "f15"},
	{KeyF16, "f16"},
	{KeyF17, "f17"},
	{KeyF18, "f18"},
	{KeyF19, "f19"},
	{KeyF20, "f20"},
	{KeyF21, "f21"},
	{KeyF22, "f22"},
	{KeyF23, "f23"},
	{KeyF24, "f24"},
	{KeyF25, "f25"},
	{KeyFn, "fn"},
	{KeyFnLock, "fn_lock"},
	{KeyPrintScreen, "print_screen"},
	{KeyScrollLock, "scroll_lock"},
	{KeyPause, "pause"},

	// Media Keys (W3C § 3.6)
	{KeyBrowserBack, "browser_back"},
	{KeyBrowserFavorites, "browser_favorites"},
	{KeyBrowserForward, "browser_forward"},
	{KeyBrowserHome, "browser_home"},
	{KeyBrowserRefresh, "browser_refresh"},
	{KeyBrowserSearch, "browser_search"},
	{KeyBrowserStop, "browser_stop"},
	{KeyEject, "eject"},
	{KeyLaunchApp1, "launch_app_1"},
	{KeyLaunchApp2, "launch_app_2"},
	{KeyLaunchMail, "launch_mail"},
	{KeyMediaPlayPause, "media_play_pause"},
	{KeyMediaSelect, "media_select"},
	{KeyMediaStop, "media_stop"},
	{KeyMediaTrackNext, "media_track_next"},
	{KeyMediaTrackPrevious, "media_track_previous"},
	{KeyPower, "power"},
	{KeySleep, "sleep"},
	{KeyAudioVolumeDown, "audio_volume_down"},
	{KeyAudioVolumeMute, "audio_volume_mute"},
	{KeyAudioVolumeUp, "audio_volume_up"},
	{KeyWakeUp, "wake_up"},

	// Legacy, Non-standard, and Special Keys (W3C § 3.7)
	{KeyCopy, "copy"},
	{KeyCut, "cut"},
	{KeyPaste, "paste"},
}

// keyToName maps a Key value to its canonical string name. Built in
// init from keyNames.
var keyToName map[Key]string

// nameToKey maps a string name to its Key value. Built in init from
// keyNames.
var nameToKey map[string]Key

func init() {
	keyToName = make(map[Key]string, len(keyNames))
	nameToKey = make(map[string]Key, len(keyNames))
	for _, e := range keyNames {
		keyToName[e.key] = e.name
		nameToKey[e.name] = e.key
	}
}

// String returns the canonical snake_case name of the Key (e.g.
// "key_a", "arrow_down", "fn"). Returns "unidentified" for unknown
// values, mirroring the behavior of KeyUnidentified.
func (k Key) String() string {
	if name, ok := keyToName[k]; ok {
		return name
	}
	return keyToName[KeyUnidentified]
}

// FromString parses a canonical snake_case key name and stores the
// corresponding Key value in the receiver. Returns an error if the
// name is not recognized.
func (k *Key) FromString(s string) error {
	if v, ok := nameToKey[s]; ok {
		*k = v
		return nil
	}
	return fmt.Errorf("libghostty: unknown key name %q", s)
}

// NewKeyFromString returns the Key value for the given canonical
// snake_case key name (e.g. "key_a", "arrow_down"). Returns an
// error if the name is not recognized.
func NewKeyFromString(s string) (Key, error) {
	var k Key
	if err := k.FromString(s); err != nil {
		return 0, err
	}
	return k, nil
}
