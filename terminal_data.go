package libghostty

// Terminal data getters wrapping ghostty_terminal_get().
// Functions are ordered alphabetically.

/*
#include <stdlib.h>
#include <ghostty/vt.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// TerminalData identifies a data field for terminal queries.
// C: GhosttyTerminalData
type TerminalData int

const (
	// TerminalDataInvalid is an invalid / sentinel value.
	TerminalDataInvalid TerminalData = C.GHOSTTY_TERMINAL_DATA_INVALID

	// TerminalDataCols is the terminal width in cells (uint16_t).
	TerminalDataCols TerminalData = C.GHOSTTY_TERMINAL_DATA_COLS

	// TerminalDataRows is the terminal height in cells (uint16_t).
	TerminalDataRows TerminalData = C.GHOSTTY_TERMINAL_DATA_ROWS

	// TerminalDataCursorX is the cursor column position, 0-indexed (uint16_t).
	TerminalDataCursorX TerminalData = C.GHOSTTY_TERMINAL_DATA_CURSOR_X

	// TerminalDataCursorY is the cursor row position within the active area,
	// 0-indexed (uint16_t).
	TerminalDataCursorY TerminalData = C.GHOSTTY_TERMINAL_DATA_CURSOR_Y

	// TerminalDataCursorPendingWrap indicates whether the cursor has a
	// pending wrap (bool).
	TerminalDataCursorPendingWrap TerminalData = C.GHOSTTY_TERMINAL_DATA_CURSOR_PENDING_WRAP

	// TerminalDataActiveScreen is the currently active screen
	// (GhosttyTerminalScreen).
	TerminalDataActiveScreen TerminalData = C.GHOSTTY_TERMINAL_DATA_ACTIVE_SCREEN

	// TerminalDataCursorVisible indicates whether the cursor is visible,
	// DEC mode 25 (bool).
	TerminalDataCursorVisible TerminalData = C.GHOSTTY_TERMINAL_DATA_CURSOR_VISIBLE

	// TerminalDataKittyKeyboardFlags is the current Kitty keyboard protocol
	// flags (uint8_t).
	TerminalDataKittyKeyboardFlags TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_KEYBOARD_FLAGS

	// TerminalDataScrollbar is the scrollbar state for the terminal viewport
	// (GhosttyTerminalScrollbar).
	TerminalDataScrollbar TerminalData = C.GHOSTTY_TERMINAL_DATA_SCROLLBAR

	// TerminalDataCursorStyle is the current SGR style of the cursor
	// (GhosttyStyle).
	TerminalDataCursorStyle TerminalData = C.GHOSTTY_TERMINAL_DATA_CURSOR_STYLE

	// TerminalDataMouseTracking indicates whether any mouse tracking mode
	// is active (bool).
	TerminalDataMouseTracking TerminalData = C.GHOSTTY_TERMINAL_DATA_MOUSE_TRACKING

	// TerminalDataTitle is the terminal title as set by escape sequences
	// (GhosttyString).
	TerminalDataTitle TerminalData = C.GHOSTTY_TERMINAL_DATA_TITLE

	// TerminalDataPwd is the terminal's current working directory as set
	// by escape sequences (GhosttyString).
	TerminalDataPwd TerminalData = C.GHOSTTY_TERMINAL_DATA_PWD

	// TerminalDataTotalRows is the total number of rows in the active screen
	// including scrollback (size_t).
	TerminalDataTotalRows TerminalData = C.GHOSTTY_TERMINAL_DATA_TOTAL_ROWS

	// TerminalDataScrollbackRows is the number of scrollback rows (size_t).
	TerminalDataScrollbackRows TerminalData = C.GHOSTTY_TERMINAL_DATA_SCROLLBACK_ROWS

	// TerminalDataWidthPx is the total terminal width in pixels (uint32_t).
	TerminalDataWidthPx TerminalData = C.GHOSTTY_TERMINAL_DATA_WIDTH_PX

	// TerminalDataHeightPx is the total terminal height in pixels (uint32_t).
	TerminalDataHeightPx TerminalData = C.GHOSTTY_TERMINAL_DATA_HEIGHT_PX

	// TerminalDataColorForeground is the effective foreground color
	// (GhosttyColorRgb).
	TerminalDataColorForeground TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND

	// TerminalDataColorBackground is the effective background color
	// (GhosttyColorRgb).
	TerminalDataColorBackground TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND

	// TerminalDataColorCursor is the effective cursor color
	// (GhosttyColorRgb).
	TerminalDataColorCursor TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR

	// TerminalDataColorPalette is the current 256-color palette
	// (GhosttyColorRgb[256]).
	TerminalDataColorPalette TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE

	// TerminalDataColorForegroundDefault is the default foreground color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorForegroundDefault TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND_DEFAULT

	// TerminalDataColorBackgroundDefault is the default background color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorBackgroundDefault TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND_DEFAULT

	// TerminalDataColorCursorDefault is the default cursor color,
	// ignoring OSC overrides (GhosttyColorRgb).
	TerminalDataColorCursorDefault TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR_DEFAULT

	// TerminalDataColorPaletteDefault is the default 256-color palette,
	// ignoring OSC overrides (GhosttyColorRgb[256]).
	TerminalDataColorPaletteDefault TerminalData = C.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE_DEFAULT

	// TerminalDataKittyImageStorageLimit is the Kitty image storage limit
	// in bytes for the active screen (uint64_t).
	TerminalDataKittyImageStorageLimit TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_STORAGE_LIMIT

	// TerminalDataKittyImageMediumFile indicates whether the file medium
	// is enabled for Kitty image loading (bool).
	TerminalDataKittyImageMediumFile TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_FILE

	// TerminalDataKittyImageMediumTempFile indicates whether the temporary
	// file medium is enabled for Kitty image loading (bool).
	TerminalDataKittyImageMediumTempFile TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_TEMP_FILE

	// TerminalDataKittyImageMediumSharedMem indicates whether the shared
	// memory medium is enabled for Kitty image loading (bool).
	TerminalDataKittyImageMediumSharedMem TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_IMAGE_MEDIUM_SHARED_MEM

	// TerminalDataKittyGraphics is the Kitty graphics image storage for
	// the active screen (GhosttyKittyGraphics).
	TerminalDataKittyGraphics TerminalData = C.GHOSTTY_TERMINAL_DATA_KITTY_GRAPHICS
)

// ActiveScreen returns which screen buffer is currently active.
func (t *Terminal) ActiveScreen() (TerminalScreen, error) {
	var v C.GhosttyTerminalScreen
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_ACTIVE_SCREEN, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return TerminalScreen(v), nil
}

// Cols returns the terminal width in cells.
func (t *Terminal) Cols() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_COLS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// ColorBackground returns the effective background color (OSC override
// or default). Returns nil if no background color is set.
func (t *Terminal) ColorBackground() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND)
}

// ColorBackgroundDefault returns the default background color, ignoring
// any OSC override. Returns nil if no default is set.
func (t *Terminal) ColorBackgroundDefault() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_BACKGROUND_DEFAULT)
}

// ColorCursor returns the effective cursor color (OSC override or
// default). Returns nil if no cursor color is set.
func (t *Terminal) ColorCursor() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR)
}

// ColorCursorDefault returns the default cursor color, ignoring any
// OSC override. Returns nil if no default is set.
func (t *Terminal) ColorCursorDefault() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_CURSOR_DEFAULT)
}

// ColorForeground returns the effective foreground color (OSC override
// or default). Returns nil if no foreground color is set.
func (t *Terminal) ColorForeground() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND)
}

// ColorForegroundDefault returns the default foreground color, ignoring
// any OSC override. Returns nil if no default is set.
func (t *Terminal) ColorForegroundDefault() (*ColorRGB, error) {
	return t.getColorRGB(C.GHOSTTY_TERMINAL_DATA_COLOR_FOREGROUND_DEFAULT)
}

// ColorPalette returns the current 256-color palette (with any OSC
// overrides applied).
func (t *Terminal) ColorPalette() (*Palette, error) {
	return t.getPalette(C.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE)
}

// ColorPaletteDefault returns the default 256-color palette, ignoring
// any OSC overrides.
func (t *Terminal) ColorPaletteDefault() (*Palette, error) {
	return t.getPalette(C.GHOSTTY_TERMINAL_DATA_COLOR_PALETTE_DEFAULT)
}

// GetMulti queries multiple terminal data fields in a single cgo call.
// This is a low-level function; prefer the typed getters (Cols, Rows,
// CursorX, etc.) for normal use. GetMulti is useful when you need many
// fields at once and want to avoid per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/terminal.h, GhosttyTerminalData enum).
//
// Example:
//
//	var cols, rows C.uint16_t
//	err := t.GetMulti(
//		[]TerminalData{TerminalDataCols, TerminalDataRows},
//		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
//	)
//
// C: ghostty_terminal_get_multi
func (t *Terminal) GetMulti(keys []TerminalData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Allocate the void** array in C memory to satisfy cgo pointer-passing rules.
	cVals := cValuesArray(values)
	defer C.free(unsafe.Pointer(cVals))
	return resultError(C.ghostty_terminal_get_multi(
		t.ptr,
		C.size_t(len(keys)),
		(*C.GhosttyTerminalData)(unsafe.Pointer(&keys[0])),
		cVals,
		nil,
	))
}

// CursorPendingWrap reports whether the cursor has a pending wrap
// (the next printed character will soft-wrap to the next line).
func (t *Terminal) CursorPendingWrap() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_CURSOR_PENDING_WRAP, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorStyle returns the current SGR style of the cursor. This is
// the style that will be applied to newly printed characters.
func (t *Terminal) CursorStyle() (*Style, error) {
	cs := initCStyle()
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_CURSOR_STYLE, unsafe.Pointer(&cs))); err != nil {
		return nil, err
	}
	return &Style{c: cs}, nil
}

// CursorVisible reports whether the cursor is visible (DEC mode 25).
func (t *Terminal) CursorVisible() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_CURSOR_VISIBLE, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorX returns the cursor column position (0-indexed).
func (t *Terminal) CursorX() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_CURSOR_X, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// CursorY returns the cursor row position within the active area
// (0-indexed).
func (t *Terminal) CursorY() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_CURSOR_Y, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// HeightPx returns the total terminal height in pixels
// (rows * cell_height_px as set by Resize).
func (t *Terminal) HeightPx() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_HEIGHT_PX, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// KittyKeyboardFlags returns the current Kitty keyboard protocol flags.
func (t *Terminal) KittyKeyboardFlags() (KittyKeyFlags, error) {
	var v C.uint8_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_KITTY_KEYBOARD_FLAGS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return KittyKeyFlags(v), nil
}

// MouseTracking reports whether any mouse tracking mode is active.
func (t *Terminal) MouseTracking() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_MOUSE_TRACKING, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Pwd returns the terminal's current working directory as set by
// escape sequences (e.g. OSC 7). Returns an empty string if unset.
// The returned string is copied; it remains valid after subsequent
// calls to VTWrite or Reset.
func (t *Terminal) Pwd() (string, error) {
	var s C.GhosttyString
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_PWD, unsafe.Pointer(&s))); err != nil {
		return "", err
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(s.ptr)), C.int(s.len)), nil
}

// Rows returns the terminal height in cells.
func (t *Terminal) Rows() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_ROWS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// Scrollbar returns the scrollbar state for the terminal viewport.
// This may be expensive to calculate depending on the viewport position;
// call only as needed.
func (t *Terminal) Scrollbar() (Scrollbar, error) {
	var v C.GhosttyTerminalScrollbar
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_SCROLLBAR, unsafe.Pointer(&v))); err != nil {
		return Scrollbar{}, err
	}
	return Scrollbar{
		Total:  uint64(v.total),
		Offset: uint64(v.offset),
		Len:    uint64(v.len),
	}, nil
}

// ScrollbackRows returns the number of scrollback rows (total rows
// minus viewport rows).
func (t *Terminal) ScrollbackRows() (uint, error) {
	var v C.size_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_SCROLLBACK_ROWS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint(v), nil
}

// Title returns the terminal title as set by escape sequences
// (e.g. OSC 0/2). Returns an empty string if unset. The returned
// string is copied; it remains valid after subsequent calls to
// VTWrite or Reset.
func (t *Terminal) Title() (string, error) {
	var s C.GhosttyString
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_TITLE, unsafe.Pointer(&s))); err != nil {
		return "", err
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(s.ptr)), C.int(s.len)), nil
}

// TotalRows returns the total number of rows in the active screen
// including scrollback.
func (t *Terminal) TotalRows() (uint, error) {
	var v C.size_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_TOTAL_ROWS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint(v), nil
}

// WidthPx returns the total terminal width in pixels
// (cols * cell_width_px as set by Resize).
func (t *Terminal) WidthPx() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_terminal_get(t.ptr, C.GHOSTTY_TERMINAL_DATA_WIDTH_PX, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// getColorRGB is a helper that reads a single ColorRGB value from the
// terminal. Returns nil (without error) when the result is NO_VALUE.
func (t *Terminal) getColorRGB(data C.GhosttyTerminalData) (*ColorRGB, error) {
	var c C.GhosttyColorRgb
	err := resultError(C.ghostty_terminal_get(t.ptr, data, unsafe.Pointer(&c)))
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	return &ColorRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}, nil
}

// getPalette is a helper that reads a full 256-color palette from the
// terminal.
func (t *Terminal) getPalette(data C.GhosttyTerminalData) (*Palette, error) {
	var cp [PaletteSize]C.GhosttyColorRgb
	if err := resultError(C.ghostty_terminal_get(t.ptr, data, unsafe.Pointer(&cp[0]))); err != nil {
		return nil, err
	}
	var p Palette
	for i, c := range cp {
		p[i] = ColorRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}
	}
	return &p, nil
}
