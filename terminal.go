package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// Terminal wraps a Ghostty VT terminal handle.
// C: GhosttyTerminal
type Terminal struct {
	ptr C.GhosttyTerminal

	// handle is a cgo.Handle pointing back to this Terminal. It is
	// stored as the C-side userdata (GHOSTTY_TERMINAL_OPT_USERDATA)
	// so that C effect trampolines can recover the *Terminal and
	// dispatch to the appropriate Go effect handler.
	handle cgo.Handle

	onWritePty         WritePtyFn
	onBell             BellFn
	onTitleChanged     TitleChangedFn
	onEnquiry          EnquiryFn
	onXtversion        XtversionFn
	onSize             SizeFn
	onColorScheme      ColorSchemeFn
	onDeviceAttributes DeviceAttributesFn

	// effectBuf holds C-allocated memory for the most recent response
	// returned by an effect trampoline (e.g. enquiry, xtversion).
	// libghostty copies the data immediately, so a single buffer
	// shared across effects is sufficient.
	effectBuf    unsafe.Pointer
	effectBufLen uintptr
}

// TerminalOption is a functional option for configuring a Terminal.
type TerminalOption func(*TerminalConfig)

// TerminalConfig holds the configuration for creating a Terminal.
// It can be passed directly to NewTerminal or built up using
// functional options like WithSize and WithMaxScrollback.
// C: GhosttyTerminalOptions
type TerminalConfig struct {
	// Cols is the terminal width in cells. Must be greater than zero.
	Cols uint16

	// Rows is the terminal height in cells. Must be greater than zero.
	Rows uint16

	// MaxScrollback is the maximum number of lines to keep in scrollback
	// history. Defaults to 0 (no scrollback).
	MaxScrollback uint

	// Effect handlers applied after terminal creation.
	onWritePty         WritePtyFn
	onBell             BellFn
	onTitleChanged     TitleChangedFn
	onEnquiry          EnquiryFn
	onXtversion        XtversionFn
	onSize             SizeFn
	onColorScheme      ColorSchemeFn
	onDeviceAttributes DeviceAttributesFn
}

// WritePtyFn is called when the terminal writes data back to the pty
// (e.g. query responses). The first parameter is the terminal that
// triggered the effect. The data is only valid for the call duration.
// C: GhosttyTerminalWritePtyFn
type WritePtyFn func(t *Terminal, data []byte)

// BellFn is called when the terminal receives a BEL character (0x07).
// The parameter is the terminal that triggered the effect.
// C: GhosttyTerminalBellFn
type BellFn func(t *Terminal)

// TitleChangedFn is called when the terminal title changes via OSC 0/2.
// The parameter is the terminal that triggered the effect.
// C: GhosttyTerminalTitleChangedFn
type TitleChangedFn func(t *Terminal)

// EnquiryFn is called when the terminal receives ENQ (0x05).
// The first parameter is the terminal that triggered the effect.
// Return the response bytes; nil or empty means no response.
// C: GhosttyTerminalEnquiryFn
type EnquiryFn func(t *Terminal) []byte

// XtversionFn is called for XTVERSION queries (CSI > q).
// The first parameter is the terminal that triggered the effect.
// Return the version string; empty uses the default "libghostty".
// C: GhosttyTerminalXtversionFn
type XtversionFn func(t *Terminal) string

// SizeFn is called for XTWINOPS size queries (CSI 14/16/18 t).
// The first parameter is the terminal that triggered the effect.
// Return the size and true, or zero value and false to ignore the query.
// C: GhosttyTerminalSizeFn
type SizeFn func(t *Terminal) (SizeReportSize, bool)

// ColorSchemeFn is called for color scheme queries (CSI ? 996 n).
// The first parameter is the terminal that triggered the effect.
// Return the scheme and true, or zero value and false to ignore the query.
// C: GhosttyTerminalColorSchemeFn
type ColorSchemeFn func(t *Terminal) (ColorScheme, bool)

// DeviceAttributesFn is called for device attributes queries
// (CSI c / CSI > c / CSI = c). The first parameter is the terminal
// that triggered the effect. Return the attributes and true,
// or zero value and false to ignore the query.
// C: GhosttyTerminalDeviceAttributesFn
type DeviceAttributesFn func(t *Terminal) (DeviceAttributes, bool)

// WithSize sets the terminal dimensions in cells.
// Both cols and rows must be greater than zero.
func WithSize(cols, rows uint16) TerminalOption {
	return func(c *TerminalConfig) {
		c.Cols = cols
		c.Rows = rows
	}
}

// WithMaxScrollback sets the maximum number of lines to keep in
// scrollback history. Defaults to 0 (no scrollback).
func WithMaxScrollback(lines uint) TerminalOption {
	return func(c *TerminalConfig) {
		c.MaxScrollback = lines
	}
}

// WithWritePty registers an effect handler invoked when the terminal
// writes data back to the pty (e.g. query responses). The data slice
// is only valid for the duration of the call.
func WithWritePty(fn WritePtyFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onWritePty = fn
	}
}

// WithBell registers an effect handler invoked when the terminal
// receives a BEL character (0x07).
func WithBell(fn BellFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onBell = fn
	}
}

// WithTitleChanged registers an effect handler invoked when the
// terminal title changes via OSC 0 or OSC 2.
func WithTitleChanged(fn TitleChangedFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onTitleChanged = fn
	}
}

// WithEnquiry registers an effect handler invoked when the terminal
// receives an ENQ character (0x05). Return the response bytes; nil
// or empty means no response.
func WithEnquiry(fn EnquiryFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onEnquiry = fn
	}
}

// WithXtversion registers an effect handler invoked for XTVERSION
// queries (CSI > q). Return the version string; empty uses the
// default "libghostty" version.
func WithXtversion(fn XtversionFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onXtversion = fn
	}
}

// WithSizeReport registers an effect handler invoked for XTWINOPS
// size queries (CSI 14/16/18 t). Return the size and true, or
// zero value and false to silently ignore the query.
func WithSizeReport(fn SizeFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onSize = fn
	}
}

// WithColorScheme registers an effect handler invoked for color
// scheme queries (CSI ? 996 n). Return the scheme and true, or
// zero value and false to silently ignore the query.
func WithColorScheme(fn ColorSchemeFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onColorScheme = fn
	}
}

// WithDeviceAttributes registers an effect handler invoked for
// device attributes queries (CSI c / CSI > c / CSI = c). Return
// the attributes and true, or zero value and false to silently
// ignore the query.
func WithDeviceAttributes(fn DeviceAttributesFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onDeviceAttributes = fn
	}
}

// NewTerminal creates a new terminal with the given options.
// WithSize is required; cols and rows must both be greater than zero.
func NewTerminal(opts ...TerminalOption) (*Terminal, error) {
	// Apply defaults and user options.
	cfg := TerminalConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	options := C.GhosttyTerminalOptions{
		cols:           C.uint16_t(cfg.Cols),
		rows:           C.uint16_t(cfg.Rows),
		max_scrollback: C.size_t(cfg.MaxScrollback),
	}

	var cterm C.GhosttyTerminal
	if err := resultError(C.ghostty_terminal_new(nil, &cterm, options)); err != nil {
		return nil, err
	}

	t := &Terminal{
		ptr:                cterm,
		onWritePty:         cfg.onWritePty,
		onBell:             cfg.onBell,
		onTitleChanged:     cfg.onTitleChanged,
		onEnquiry:          cfg.onEnquiry,
		onXtversion:        cfg.onXtversion,
		onSize:             cfg.onSize,
		onColorScheme:      cfg.onColorScheme,
		onDeviceAttributes: cfg.onDeviceAttributes,
	}

	// Always set userdata to our handle so trampolines can find us.
	t.handle = cgo.NewHandle(t)
	C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_USERDATA,
		handleToPointer(t.handle),
	)

	// Register any effects that were provided via options.
	t.syncEffects()

	return t, nil
}

// Close frees the underlying terminal handle and releases the cgo.Handle.
// After this call, the terminal must not be used.
func (t *Terminal) Close() {
	t.handle.Delete()
	C.ghostty_terminal_free(t.ptr)
	if t.effectBuf != nil {
		Free(t.effectBuf, t.effectBufLen)
	}
}

// Reset performs a full terminal reset (RIS).
// All state is reset to initial configuration (modes, scrollback,
// scrolling region, screen contents). Dimensions are preserved.
func (t *Terminal) Reset() {
	C.ghostty_terminal_reset(t.ptr)
}

// Resize changes the terminal dimensions.
// Both cols and rows must be greater than zero. cellWidthPx and
// cellHeightPx specify the pixel dimensions of a single cell, used
// for image protocols and size reports.
func (t *Terminal) Resize(cols, rows uint16, cellWidthPx, cellHeightPx uint32) error {
	return resultError(C.ghostty_terminal_resize(
		t.ptr,
		C.uint16_t(cols),
		C.uint16_t(rows),
		C.uint32_t(cellWidthPx),
		C.uint32_t(cellHeightPx),
	))
}

// VTWrite feeds raw VT-encoded bytes through the terminal's parser,
// updating terminal state. Malformed input is handled gracefully and
// will not cause an error.
func (t *Terminal) VTWrite(data []byte) {
	if len(data) == 0 {
		return
	}
	C.ghostty_terminal_vt_write(t.ptr, (*C.uint8_t)(&data[0]), C.size_t(len(data)))
}

// Write implements io.Writer by feeding data through the terminal's
// VT parser. It always consumes all bytes and never returns an error.
func (t *Terminal) Write(p []byte) (int, error) {
	t.VTWrite(p)
	return len(p), nil
}

// ModeGet returns the current value of a terminal mode.
func (t *Terminal) ModeGet(mode Mode) (bool, error) {
	var val C.bool
	if err := resultError(C.ghostty_terminal_mode_get(t.ptr, C.GhosttyMode(mode), &val)); err != nil {
		return false, err
	}
	return bool(val), nil
}

// ModeSet sets a terminal mode to the given value.
func (t *Terminal) ModeSet(mode Mode, value bool) error {
	return resultError(C.ghostty_terminal_mode_set(t.ptr, C.GhosttyMode(mode), C.bool(value)))
}

// ScrollViewportTag describes the scroll behavior.
// C: GhosttyTerminalScrollViewportTag
type ScrollViewportTag int

const (
	// ScrollViewportTop scrolls to the top of scrollback.
	ScrollViewportTop ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_TOP

	// ScrollViewportBottom scrolls to the bottom (active area).
	ScrollViewportBottom ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_BOTTOM

	// ScrollViewportDelta scrolls by a delta amount (up is negative).
	ScrollViewportDelta ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_DELTA
)

// ScrollViewport scrolls the terminal viewport to the top of scrollback.
func (t *Terminal) ScrollViewportTop() {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_TOP
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// ScrollViewportBottom scrolls the terminal viewport to the bottom
// (active area).
func (t *Terminal) ScrollViewportBottom() {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_BOTTOM
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// ScrollViewportDelta scrolls the terminal viewport by the given delta
// (negative for up, positive for down).
func (t *Terminal) ScrollViewportDelta(delta int) {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_DELTA
	// Set the delta in the value union. The delta field is at offset 0.
	*(*C.intptr_t)(unsafe.Pointer(&sv.value[0])) = C.intptr_t(delta)
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// TerminalScreen identifies which screen buffer is active.
// C: GhosttyTerminalScreen
type TerminalScreen int

const (
	// ScreenPrimary is the primary (normal) screen.
	ScreenPrimary TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_PRIMARY

	// ScreenAlternate is the alternate screen.
	ScreenAlternate TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_ALTERNATE
)

// Scrollbar holds the scrollbar state for the terminal viewport.
// C: GhosttyTerminalScrollbar
type Scrollbar struct {
	// Total is the total size of the scrollable area in rows.
	Total uint64

	// Offset is the offset into the total area that the viewport is at.
	Offset uint64

	// Len is the length of the visible area in rows.
	Len uint64
}

// KittyGraphics returns the Kitty graphics image storage for the
// terminal's active screen. The returned handle is borrowed from
// the terminal and remains valid until the next mutating call
// (e.g. VTWrite or Reset).
func (t *Terminal) KittyGraphics() (*KittyGraphics, error) {
	var ptr C.GhosttyKittyGraphics
	if err := resultError(C.ghostty_terminal_get(
		t.ptr,
		C.GHOSTTY_TERMINAL_DATA_KITTY_GRAPHICS,
		unsafe.Pointer(&ptr),
	)); err != nil {
		return nil, err
	}
	return &KittyGraphics{ptr: ptr}, nil
}

// GridRef resolves a point in the terminal grid to a grid reference.
// The returned GridRef is only valid until the next terminal update.
//
// Lookups using PointTagActive and PointTagViewport are fast.
// PointTagScreen and PointTagHistory may be expensive for large
// scrollback buffers.
func (t *Terminal) GridRef(point Point) (*GridRef, error) {
	ref := initCGridRef()
	if err := resultError(C.ghostty_terminal_grid_ref(
		t.ptr,
		point.toC(),
		&ref,
	)); err != nil {
		return nil, err
	}
	return &GridRef{ref: ref}, nil
}

// handleToPointer converts a cgo.Handle (uintptr) to unsafe.Pointer
// for passing as C userdata. The handle is an opaque integer, not a
// real Go pointer, so we suppress checkptr which would otherwise
// reject it under -race.
//
//go:nocheckptr
func handleToPointer(h cgo.Handle) unsafe.Pointer {
	return unsafe.Pointer(h)
}
