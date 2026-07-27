package libghostty

// Terminal option setters wrapping ghostty_terminal_set().
// Functions are ordered alphabetically.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// SetEffectWritePty registers (or clears) the write-pty effect on a
// live terminal. Pass nil to clear.
func (t *Terminal) SetEffectWritePty(fn WritePtyFn) {
	t.onWritePty = fn
	t.syncEffects()
}

// SetEffectBell registers (or clears) the bell effect on a live terminal.
// Pass nil to clear.
func (t *Terminal) SetEffectBell(fn BellFn) {
	t.onBell = fn
	t.syncEffects()
}

// SetEffectClipboardWrite registers (or clears) the clipboard-write effect
// on a live terminal. Pass nil to clear.
func (t *Terminal) SetEffectClipboardWrite(fn ClipboardWriteFn) {
	t.onClipboardWrite = fn
	t.syncEffects()
}

// SetEffectTitleChanged registers (or clears) the title-changed effect
// on a live terminal. Pass nil to clear.
func (t *Terminal) SetEffectTitleChanged(fn TitleChangedFn) {
	t.onTitleChanged = fn
	t.syncEffects()
}

// SetEffectPwdChanged registers (or clears) the pwd-changed effect on a live
// terminal. Pass nil to clear.
func (t *Terminal) SetEffectPwdChanged(fn PwdChangedFn) {
	t.onPwdChanged = fn
	t.syncEffects()
}

// SetEffectEnquiry registers (or clears) the enquiry effect on a live
// terminal. Pass nil to clear.
func (t *Terminal) SetEffectEnquiry(fn EnquiryFn) {
	t.onEnquiry = fn
	t.syncEffects()
}

// SetEffectXtversion registers (or clears) the xtversion effect on a
// live terminal. Pass nil to clear.
func (t *Terminal) SetEffectXtversion(fn XtversionFn) {
	t.onXtversion = fn
	t.syncEffects()
}

// SetEffectSize registers (or clears) the size-report effect on a
// live terminal. Pass nil to clear.
func (t *Terminal) SetEffectSize(fn SizeFn) {
	t.onSize = fn
	t.syncEffects()
}

// SetEffectColorScheme registers (or clears) the color-scheme effect
// on a live terminal. Pass nil to clear.
func (t *Terminal) SetEffectColorScheme(fn ColorSchemeFn) {
	t.onColorScheme = fn
	t.syncEffects()
}

// SetEffectDeviceAttributes registers (or clears) the device-attributes
// effect on a live terminal. Pass nil to clear.
func (t *Terminal) SetEffectDeviceAttributes(fn DeviceAttributesFn) {
	t.onDeviceAttributes = fn
	t.syncEffects()
}

// SetColorBackground sets the default background color. Pass nil to
// clear (unset).
func (t *Terminal) SetColorBackground(c *ColorRGB) error {
	var val unsafe.Pointer
	if c != nil {
		cc := C.GhosttyColorRgb{r: C.uint8_t(c.R), g: C.uint8_t(c.G), b: C.uint8_t(c.B)}
		val = unsafe.Pointer(&cc)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_COLOR_BACKGROUND,
		val,
	))
}

// SetColorCursor sets the default cursor color. Pass nil to clear (unset).
func (t *Terminal) SetColorCursor(c *ColorRGB) error {
	var val unsafe.Pointer
	if c != nil {
		cc := C.GhosttyColorRgb{r: C.uint8_t(c.R), g: C.uint8_t(c.G), b: C.uint8_t(c.B)}
		val = unsafe.Pointer(&cc)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_COLOR_CURSOR,
		val,
	))
}

// SetColorForeground sets the default foreground color. Pass nil to
// clear (unset).
func (t *Terminal) SetColorForeground(c *ColorRGB) error {
	var val unsafe.Pointer
	if c != nil {
		cc := C.GhosttyColorRgb{r: C.uint8_t(c.R), g: C.uint8_t(c.G), b: C.uint8_t(c.B)}
		val = unsafe.Pointer(&cc)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_COLOR_FOREGROUND,
		val,
	))
}

// SetColorPalette sets the default 256-color palette. Pass nil to reset
// to the built-in default palette.
func (t *Terminal) SetColorPalette(palette *Palette) error {
	var val unsafe.Pointer
	if palette != nil {
		// Convert Go palette to C palette.
		var cp [PaletteSize]C.GhosttyColorRgb
		for i, c := range palette {
			cp[i] = C.GhosttyColorRgb{r: C.uint8_t(c.R), g: C.uint8_t(c.G), b: C.uint8_t(c.B)}
		}
		val = unsafe.Pointer(&cp[0])
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_COLOR_PALETTE,
		val,
	))
}

// SetDefaultCursorBlink sets whether DECSCUSR reset selects a blinking
// cursor. Passing nil restores the built-in default of not blinking.
func (t *Terminal) SetDefaultCursorBlink(blink *bool) error {
	var val unsafe.Pointer
	if blink != nil {
		v := C.bool(*blink)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_DEFAULT_CURSOR_BLINK,
		val,
	))
}

// SetDefaultCursorStyle sets the cursor style selected by DECSCUSR reset.
// Passing nil restores the built-in block cursor.
func (t *Terminal) SetDefaultCursorStyle(style *TerminalCursorStyle) error {
	var val unsafe.Pointer
	if style != nil {
		v := C.GhosttyTerminalCursorStyle(*style)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_DEFAULT_CURSOR_STYLE,
		val,
	))
}

// SetGlyphProtocol enables or disables Glyph Protocol APC handling.
func (t *Terminal) SetGlyphProtocol(enabled bool) error {
	v := C.bool(enabled)
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_GLYPH_PROTOCOL,
		unsafe.Pointer(&v),
	))
}

// SetAPCMaxBytes sets the maximum bytes the APC handler will buffer for
// all protocols. Passing nil removes all overrides and reverts to the
// built-in defaults.
func (t *Terminal) SetAPCMaxBytes(limit *uint) error {
	var val unsafe.Pointer
	if limit != nil {
		v := C.size_t(*limit)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_APC_MAX_BYTES,
		val,
	))
}

// SetAPCMaxBytesKitty sets the maximum bytes the APC handler will buffer
// for Kitty graphics protocol data. Passing nil removes the override and
// reverts to the built-in default.
func (t *Terminal) SetAPCMaxBytesKitty(limit *uint) error {
	var val unsafe.Pointer
	if limit != nil {
		v := C.size_t(*limit)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_APC_MAX_BYTES_KITTY,
		val,
	))
}

// SetPwd sets the terminal working directory manually. An empty string
// clears it.
func (t *Terminal) SetPwd(pwd string) error {
	return t.setStringOption(C.GHOSTTY_TERMINAL_OPT_PWD, &pwd)
}

// SetSelection sets the active screen selection. Passing nil clears the
// active screen selection.
//
// The terminal copies the selection immediately and converts it to
// terminal-owned tracked state, so the Selection value and its untracked
// grid references do not need to outlive this call.
func (t *Terminal) SetSelection(sel *Selection) error {
	var val unsafe.Pointer
	if sel != nil {
		cs := initCSelection()
		cs.start = sel.Start.ref
		cs.end = sel.End.ref
		cs.rectangle = C.bool(sel.Rectangle)
		val = unsafe.Pointer(&cs)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_SELECTION,
		val,
	))
}

// SetKittyImageStorageLimit sets the Kitty image storage limit in bytes.
// Applied to all initialized screens (primary and alternate). A value of
// zero disables the Kitty graphics protocol entirely, deleting all stored
// images and placements. Pass nil to disable (equivalent to zero).
func (t *Terminal) SetKittyImageStorageLimit(limit *uint64) error {
	var val unsafe.Pointer
	if limit != nil {
		v := C.uint64_t(*limit)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_KITTY_IMAGE_STORAGE_LIMIT,
		val,
	))
}

// SetKittyImageMediumFile enables or disables Kitty image loading via the
// file medium.
func (t *Terminal) SetKittyImageMediumFile(enabled bool) error {
	v := C.bool(enabled)
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_KITTY_IMAGE_MEDIUM_FILE,
		unsafe.Pointer(&v),
	))
}

// SetKittyImageMediumTempFile enables Kitty image loading via the temporary
// file medium and restricts it to directory. Passing nil disables the medium.
func (t *Terminal) SetKittyImageMediumTempFile(directory *string) error {
	return t.setStringOption(
		C.GHOSTTY_TERMINAL_OPT_KITTY_IMAGE_MEDIUM_TEMP_FILE,
		directory,
	)
}

// SetKittyImageMediumSharedMem enables or disables Kitty image loading via
// the shared memory medium.
func (t *Terminal) SetKittyImageMediumSharedMem(enabled bool) error {
	v := C.bool(enabled)
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_KITTY_IMAGE_MEDIUM_SHARED_MEM,
		unsafe.Pointer(&v),
	))
}

// SetScrollbackMaxBytes sets the approximate maximum scrollback allocation
// in bytes. Passing nil removes the byte limit. Lowering the limit may
// immediately remove eligible historical pages.
func (t *Terminal) SetScrollbackMaxBytes(limit *uint) error {
	var val unsafe.Pointer
	if limit != nil {
		v := C.size_t(*limit)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_SCROLLBACK_MAX_BYTES,
		val,
	))
}

// SetScrollbackMaxLines sets the approximate maximum number of physical
// lines retained in scrollback. Passing nil removes the line limit. Lowering
// the limit may immediately remove eligible historical pages.
func (t *Terminal) SetScrollbackMaxLines(limit *uint) error {
	var val unsafe.Pointer
	if limit != nil {
		v := C.size_t(*limit)
		val = unsafe.Pointer(&v)
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_SCROLLBACK_MAX_LINES,
		val,
	))
}

// SetTitle sets the terminal title manually. An empty string clears it.
func (t *Terminal) SetTitle(title string) error {
	return t.setStringOption(C.GHOSTTY_TERMINAL_OPT_TITLE, &title)
}

// setStringOption stages a Go string through C-owned memory before passing a
// GhosttyString descriptor to cgo. ghostty_terminal_set copies string options
// synchronously, so the temporary allocation can be released on return.
func (t *Terminal) setStringOption(
	option C.GhosttyTerminalOption,
	value *string,
) error {
	if value == nil {
		return resultError(C.ghostty_terminal_set(t.ptr, option, nil))
	}

	length := uintptr(len(*value))
	var ptr unsafe.Pointer
	if length > 0 {
		ptr = Alloc(length)
		if ptr == nil {
			return &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(ptr), int(length)), *value)
		defer Free(ptr, length)
	}

	s := C.GhosttyString{
		ptr: (*C.uint8_t)(ptr),
		len: C.size_t(length),
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		option,
		unsafe.Pointer(&s),
	))
}
