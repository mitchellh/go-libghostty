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

// SetEffectTitleChanged registers (or clears) the title-changed effect
// on a live terminal. Pass nil to clear.
func (t *Terminal) SetEffectTitleChanged(fn TitleChangedFn) {
	t.onTitleChanged = fn
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

// SetPwd sets the terminal working directory manually. An empty string
// clears it.
func (t *Terminal) SetPwd(pwd string) error {
	s := C.GhosttyString{
		ptr: (*C.uint8_t)(unsafe.Pointer(unsafe.StringData(pwd))),
		len: C.size_t(len(pwd)),
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_PWD,
		unsafe.Pointer(&s),
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

// SetKittyImageMediumTempFile enables or disables Kitty image loading via
// the temporary file medium.
func (t *Terminal) SetKittyImageMediumTempFile(enabled bool) error {
	v := C.bool(enabled)
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_KITTY_IMAGE_MEDIUM_TEMP_FILE,
		unsafe.Pointer(&v),
	))
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

// SetTitle sets the terminal title manually. An empty string clears it.
func (t *Terminal) SetTitle(title string) error {
	s := C.GhosttyString{
		ptr: (*C.uint8_t)(unsafe.Pointer(unsafe.StringData(title))),
		len: C.size_t(len(title)),
	}
	return resultError(C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_TITLE,
		unsafe.Pointer(&s),
	))
}
