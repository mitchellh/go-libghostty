package libghostty

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyFormatterTerminalOptions (sized struct).
static inline GhosttyFormatterTerminalOptions init_formatter_terminal_options() {
	GhosttyFormatterTerminalOptions opts = GHOSTTY_INIT_SIZED(GhosttyFormatterTerminalOptions);
	opts.extra.size = sizeof(GhosttyFormatterTerminalExtra);
	opts.extra.screen.size = sizeof(GhosttyFormatterScreenExtra);
	return opts;
}
*/
import "C"

import (
	"io"
	"unsafe"
)

// FormatterFormat selects the output format for a Formatter.
// C: GhosttyFormatterFormat
type FormatterFormat int

const (
	// FormatterFormatPlain emits plain text (no escape sequences).
	FormatterFormatPlain FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_PLAIN

	// FormatterFormatVT emits VT sequences preserving colors, styles, URLs, etc.
	FormatterFormatVT FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_VT

	// FormatterFormatHTML emits HTML with inline styles.
	FormatterFormatHTML FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_HTML
)

// formatterOpts wraps the C options struct so that functional options
// can mutate it directly. Only fields explicitly set by an option are
// modified; everything else retains the GHOSTTY_INIT_SIZED defaults.
type formatterOpts struct {
	c C.GhosttyFormatterTerminalOptions
}

// FormatterOption is a functional option for configuring a Formatter.
type FormatterOption func(*formatterOpts)

// WithFormatterFormat sets the output format (plain, VT, or HTML).
// Defaults to FormatterFormatPlain if not specified.
func WithFormatterFormat(f FormatterFormat) FormatterOption {
	return func(o *formatterOpts) {
		o.c.emit = C.GhosttyFormatterFormat(f)
	}
}

// WithFormatterUnwrap enables unwrapping of soft-wrapped lines.
func WithFormatterUnwrap(unwrap bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.unwrap = C.bool(unwrap)
	}
}

// WithFormatterTrim enables trimming of trailing whitespace on
// non-blank lines.
func WithFormatterTrim(trim bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.trim = C.bool(trim)
	}
}

// WithFormatterExtraPalette emits the palette using OSC 4 sequences.
func WithFormatterExtraPalette(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.palette = C.bool(v)
	}
}

// WithFormatterExtraModes emits terminal modes that differ from their
// defaults using CSI h/l.
func WithFormatterExtraModes(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.modes = C.bool(v)
	}
}

// WithFormatterExtraScrollingRegion emits scrolling region state using
// DECSTBM and DECSLRM sequences.
func WithFormatterExtraScrollingRegion(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.scrolling_region = C.bool(v)
	}
}

// WithFormatterExtraTabstops emits tabstop positions by clearing all
// tabs and setting each one.
func WithFormatterExtraTabstops(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.tabstops = C.bool(v)
	}
}

// WithFormatterExtraPwd emits the present working directory using OSC 7.
func WithFormatterExtraPwd(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.pwd = C.bool(v)
	}
}

// WithFormatterExtraKeyboard emits keyboard modes such as
// ModifyOtherKeys.
func WithFormatterExtraKeyboard(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.keyboard = C.bool(v)
	}
}

// WithFormatterExtraCursor emits cursor position using CUP (CSI H).
func WithFormatterExtraCursor(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.cursor = C.bool(v)
	}
}

// WithFormatterExtraStyle emits current SGR style state based on the
// cursor's active style_id.
func WithFormatterExtraStyle(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.style = C.bool(v)
	}
}

// WithFormatterExtraHyperlink emits current hyperlink state using
// OSC 8 sequences.
func WithFormatterExtraHyperlink(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.hyperlink = C.bool(v)
	}
}

// WithFormatterExtraProtection emits character protection mode using
// DECSCA.
func WithFormatterExtraProtection(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.protection = C.bool(v)
	}
}

// WithFormatterExtraKittyKeyboard emits Kitty keyboard protocol state
// using CSI > u and CSI = sequences.
func WithFormatterExtraKittyKeyboard(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.kitty_keyboard = C.bool(v)
	}
}

// WithFormatterExtraCharsets emits character set designations and
// invocations.
func WithFormatterExtraCharsets(v bool) FormatterOption {
	return func(o *formatterOpts) {
		o.c.extra.screen.charsets = C.bool(v)
	}
}

// Formatter wraps a Ghostty formatter handle that can produce
// plain text, VT sequences, or HTML from a terminal's current state.
// The formatter stores a borrowed reference to a terminal, so the
// terminal must outlive the formatter and formatter calls must be
// serialized with all other access to that terminal.
//
// Formatter implements io.WriterTo so formatted output can be written
// directly to any io.Writer.
// C: GhosttyFormatter
type Formatter struct {
	ptr C.GhosttyFormatter
}

// NewFormatter creates a formatter for the given terminal's active screen.
// The terminal must outlive the formatter. The formatter captures a
// borrowed reference to the terminal and reads its current state on
// each [Formatter.Format] call, so formatter calls must be serialized
// with other access to the terminal.
func NewFormatter(t *Terminal, opts ...FormatterOption) (*Formatter, error) {
	// Start with GHOSTTY_INIT_SIZED defaults; options only touch
	// fields the caller explicitly sets.
	fo := formatterOpts{c: C.init_formatter_terminal_options()}
	for _, opt := range opts {
		opt(&fo)
	}

	var ptr C.GhosttyFormatter
	if err := resultError(C.ghostty_formatter_terminal_new(nil, &ptr, t.ptr, fo.c)); err != nil {
		return nil, err
	}

	return &Formatter{ptr: ptr}, nil
}

// Close frees the formatter handle. After this call, the formatter
// must not be used.
func (f *Formatter) Close() {
	C.ghostty_formatter_free(f.ptr)
}

// Format runs the formatter and returns the output as a byte slice.
// Each call reflects the terminal's current state at the time of the
// call. Serialize Format with all other access to the underlying
// terminal. The returned buffer is allocated by libghostty and copied
// into Go memory.
func (f *Formatter) Format() ([]byte, error) {
	var outPtr *C.uint8_t
	var outLen C.size_t
	if err := resultError(C.ghostty_formatter_format_alloc(f.ptr, nil, &outPtr, &outLen)); err != nil {
		return nil, err
	}
	defer C.ghostty_free(nil, outPtr, outLen)

	return C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen)), nil
}

// FormatString runs the formatter and returns the output as a string.
// This is a convenience wrapper around Format.
func (f *Formatter) FormatString() (string, error) {
	b, err := f.Format()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteTo implements io.WriterTo. It formats the current terminal
// state and writes the entire output to w.
func (f *Formatter) WriteTo(w io.Writer) (int64, error) {
	b, err := f.Format()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}
