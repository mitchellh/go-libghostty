package libghostty

/*
#include <ghostty/vt.h>
#include "go_io.h"
#include <string.h>

// Buffer small formatter writes before forwarding them to Go. Styled output
// can contain hundreds of small writes, and crossing into Go for each one is
// expensive. The buffer is flushed when full and when formatting completes;
// large writes bypass it. libghostty formats synchronously and is fast, so
// buffering is not expected to add noticeable latency.
typedef struct {
	GhosttyWriter downstream;
	uint8_t* buffer;
	size_t len;
	size_t capacity;
} ghostty_go_formatter_writer;

static bool ghostty_go_formatter_writer_flush(
	ghostty_go_formatter_writer* writer
) {
	if (writer->len == 0) return true;
	if (!writer->downstream.write(
		writer->downstream.userdata,
		writer->buffer,
		writer->len)) {
		return false;
	}
	writer->len = 0;
	return true;
}

static bool ghostty_go_formatter_writer_write(
	void* userdata,
	const uint8_t* data,
	size_t len
) {
	ghostty_go_formatter_writer* writer = userdata;
	while (len > 0) {
		if (writer->len == 0 && len >= writer->capacity) {
			return writer->downstream.write(
				writer->downstream.userdata,
				data,
				len);
		}

		size_t available = writer->capacity - writer->len;
		size_t count = len < available ? len : available;
		memcpy(writer->buffer + writer->len, data, count);
		writer->len += count;
		data += count;
		len -= count;

		if (writer->len == writer->capacity &&
			!ghostty_go_formatter_writer_flush(writer)) {
			return false;
		}
	}
	return true;
}

static inline GhosttyResult ghostty_go_formatter_format(
	GhosttyFormatter formatter,
	uintptr_t userdata
) {
	uint8_t buffer[16 * 1024];
	ghostty_go_formatter_writer context = {
		.downstream = ghostty_go_writer(userdata),
		.buffer = buffer,
		.len = 0,
		.capacity = sizeof(buffer),
	};
	GhosttyWriter writer = {
		.write = ghostty_go_formatter_writer_write,
		.userdata = &context,
	};

	GhosttyResult result = ghostty_formatter_format(formatter, writer);
	if (result != GHOSTTY_SUCCESS) return result;
	if (!ghostty_go_formatter_writer_flush(&context)) {
		return GHOSTTY_IO_ERROR;
	}
	return GHOSTTY_SUCCESS;
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

// FormatterExtras selects terminal state to add to VT output. Combine values
// with |.
//
// VT output is meant to be written to another terminal to recreate the screen.
// Without extras, it reproduces the text and how each cell is styled. Each
// extra adds sequences that restore one more piece of terminal state, such as
// the cursor position or the colors in the palette. Plain text ignores extras.
//
// [FormatterFormatHTML] uses only [FormatterExtraPalette], which adds the
// <style> block defining the --vt-palette-N CSS variables. Palette colors in
// HTML output refer to those variables, so without this extra the document has
// to define them itself.
//
// A new formatter uses [FormatterExtrasDefault]. Use [FormatterExtrasAll] to
// recreate the terminal as closely as possible.
//
// C: GhosttyFormatterExtras
type FormatterExtras uint32

const (
	// FormatterExtraPalette restores the color palette with OSC 4.
	FormatterExtraPalette FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_PALETTE

	// FormatterExtraModes restores terminal modes that differ from their
	// defaults with CSI h and l.
	FormatterExtraModes FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_MODES

	// FormatterExtraScrollingRegion restores the scrolling region with DECSTBM
	// and DECSLRM.
	FormatterExtraScrollingRegion FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_SCROLLING_REGION

	// FormatterExtraTabstops restores tab stops by clearing them all and
	// setting each one again.
	FormatterExtraTabstops FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_TABSTOPS

	// FormatterExtraPwd restores the working directory with OSC 7.
	FormatterExtraPwd FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_PWD

	// FormatterExtraKeyboard restores keyboard modes such as modifyOtherKeys.
	FormatterExtraKeyboard FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_KEYBOARD

	// FormatterExtraCursor restores the cursor position with CUP.
	FormatterExtraCursor FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_CURSOR

	// FormatterExtraStyle restores the style the cursor writes new text with,
	// using SGR.
	FormatterExtraStyle FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_STYLE

	// FormatterExtraHyperlink restores the hyperlink the cursor writes new text
	// with, using OSC 8.
	FormatterExtraHyperlink FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_HYPERLINK

	// FormatterExtraProtection restores character protection with DECSCA.
	FormatterExtraProtection FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_PROTECTION

	// FormatterExtraKittyKeyboard restores the Kitty keyboard protocol flags.
	FormatterExtraKittyKeyboard FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_KITTY_KEYBOARD

	// FormatterExtraCharsets restores the character sets, such as the line
	// drawing set.
	FormatterExtraCharsets FormatterExtras = C.GHOSTTY_FORMATTER_EXTRA_CHARSETS

	// FormatterExtrasNone selects no extras. VT output has only the text and
	// its styles.
	FormatterExtrasNone FormatterExtras = C.GHOSTTY_FORMATTER_EXTRAS_NONE

	// FormatterExtrasAll selects every extra.
	FormatterExtrasAll FormatterExtras = C.GHOSTTY_FORMATTER_EXTRAS_ALL

	// FormatterExtrasDefault is what a new formatter starts with: the palette,
	// the cursor style and the cursor hyperlink. These are the extras that
	// affect how output looks.
	FormatterExtrasDefault FormatterExtras = C.GHOSTTY_FORMATTER_EXTRAS_DEFAULT
)

// FormatterOption is a functional option for configuring a Formatter. Each one
// applies a setting to the new formatter, so anything an option does can also
// be done later with the corresponding Formatter setter.
type FormatterOption func(*Formatter) error

// WithFormatterFormat sets the output format (plain, VT, or HTML).
// Defaults to FormatterFormatPlain if not specified. See [Formatter.SetEmit].
func WithFormatterFormat(format FormatterFormat) FormatterOption {
	return func(f *Formatter) error {
		return f.SetEmit(format)
	}
}

// WithFormatterUnwrap enables unwrapping of soft-wrapped lines. Defaults to
// false. See [Formatter.SetUnwrap].
func WithFormatterUnwrap(unwrap bool) FormatterOption {
	return func(f *Formatter) error {
		return f.SetUnwrap(unwrap)
	}
}

// WithFormatterTrim controls whether trailing whitespace is removed from
// lines that contain text. Defaults to true. See [Formatter.SetTrim].
func WithFormatterTrim(trim bool) FormatterOption {
	return func(f *Formatter) error {
		return f.SetTrim(trim)
	}
}

// WithFormatterSelection restricts formatter output to sel. Passing nil
// formats the entire active screen, which is the default. See
// [Formatter.SetSelection].
func WithFormatterSelection(sel *Selection) FormatterOption {
	return func(f *Formatter) error {
		return f.SetSelection(sel)
	}
}

// WithFormatterExtras sets the terminal state added to VT output. Defaults to
// [FormatterExtrasDefault]. Set this to [FormatterExtrasNone] for just the text
// and its styles, since the default adds the whole palette to VT and HTML
// output. See [Formatter.SetExtras].
func WithFormatterExtras(extras FormatterExtras) FormatterOption {
	return func(f *Formatter) error {
		return f.SetExtras(extras)
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

	// writer is reused across WriteTo calls to avoid allocating cgo handles.
	writer ghosttyWriterBridge
}

// NewFormatter creates a formatter for the given terminal's active screen.
// The terminal must outlive the formatter. The formatter captures a
// borrowed reference to the terminal and reads its current state on
// each [Formatter.Format] call, so formatter calls must be serialized
// with other access to the terminal.
//
// Without options, the formatter emits plain text with trailing whitespace
// trimmed. Settings can be changed at any time with the Formatter setters and
// stay in effect across format calls.
func NewFormatter(t *Terminal, opts ...FormatterOption) (*Formatter, error) {
	var ptr C.GhosttyFormatter
	if err := resultError(C.ghostty_formatter_terminal_new(nil, &ptr, t.ptr)); err != nil {
		return nil, err
	}

	f := &Formatter{ptr: ptr}
	for _, opt := range opts {
		if err := opt(f); err != nil {
			f.Close()
			return nil, err
		}
	}

	return f, nil
}

// set writes one formatter option. A nil value restores the option's default.
func (f *Formatter) set(option C.GhosttyFormatterOption, value unsafe.Pointer) error {
	return resultError(C.ghostty_formatter_set(f.ptr, option, value))
}

// get reads one formatter option.
func (f *Formatter) get(option C.GhosttyFormatterOption, out unsafe.Pointer) error {
	return resultError(C.ghostty_formatter_get(f.ptr, option, out))
}

// SetEmit sets the output format. The change applies to the next format call.
func (f *Formatter) SetEmit(format FormatterFormat) error {
	v := C.GhosttyFormatterFormat(format)
	return f.set(C.GHOSTTY_FORMATTER_OPT_EMIT, unsafe.Pointer(&v))
}

// Emit returns the output format.
func (f *Formatter) Emit() (FormatterFormat, error) {
	var v C.GhosttyFormatterFormat
	if err := f.get(C.GHOSTTY_FORMATTER_OPT_EMIT, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return FormatterFormat(v), nil
}

// SetUnwrap controls whether lines that wrapped because they were too long for
// the screen are joined. When false, each row of the screen is its own line in
// the output. When true, a line the terminal wrapped onto several rows comes
// out as one line, the way the program originally wrote it.
func (f *Formatter) SetUnwrap(unwrap bool) error {
	v := C.bool(unwrap)
	return f.set(C.GHOSTTY_FORMATTER_OPT_UNWRAP, unsafe.Pointer(&v))
}

// Unwrap reports whether soft-wrapped lines are joined.
func (f *Formatter) Unwrap() (bool, error) {
	var v C.bool
	if err := f.get(C.GHOSTTY_FORMATTER_OPT_UNWRAP, unsafe.Pointer(&v)); err != nil {
		return false, err
	}
	return bool(v), nil
}

// SetTrim controls whether trailing whitespace is removed from lines that
// contain text.
//
// Blank lines at the end of the output are always removed. The one exception
// is [FormatterFormatVT] output with trim set to false, which keeps those lines
// so that the output reproduces the screen layout.
func (f *Formatter) SetTrim(trim bool) error {
	v := C.bool(trim)
	return f.set(C.GHOSTTY_FORMATTER_OPT_TRIM, unsafe.Pointer(&v))
}

// Trim reports whether trailing whitespace is removed.
func (f *Formatter) Trim() (bool, error) {
	var v C.bool
	if err := f.get(C.GHOSTTY_FORMATTER_OPT_TRIM, unsafe.Pointer(&v)); err != nil {
		return false, err
	}
	return bool(v), nil
}

// SetSelection restricts formatter output to sel. Passing nil clears it and
// formats the entire active screen again.
//
// To format the terminal's active selection, read it with [Terminal.Selection]
// and pass it here. Check for nil first: Terminal.Selection returns nil when
// nothing is selected, which here would format the whole screen. For the same
// text Ghostty copies to the clipboard, use plain output with both unwrap and
// trim enabled.
//
// The selection must come from the same terminal passed to [NewFormatter]. The
// formatter copies it, so the Selection value itself does not need to outlive
// this call. The copy still contains borrowed grid references into the
// terminal, and like any [Selection] it is only valid until the terminal next
// changes. After that, set the selection again before formatting, or clear it.
func (f *Formatter) SetSelection(sel *Selection) error {
	if sel == nil {
		return f.set(C.GHOSTTY_FORMATTER_OPT_SELECTION, nil)
	}
	cs := sel.toC()
	return f.set(C.GHOSTTY_FORMATTER_OPT_SELECTION, unsafe.Pointer(&cs))
}

// Selection returns the selection the formatter is restricted to, or nil when
// it formats the entire active screen. The result is subject to the same
// validity rules as the value passed to [Formatter.SetSelection].
func (f *Formatter) Selection() (*Selection, error) {
	var cs C.GhosttySelection
	if err := f.get(C.GHOSTTY_FORMATTER_OPT_SELECTION, unsafe.Pointer(&cs)); err != nil {
		return noValueNilSelection(err)
	}
	sel := selectionFromC(cs)
	return &sel, nil
}

// SetExtras sets the terminal state added to VT output. To change one extra and
// keep the rest, read the current value with [Formatter.Extras] first.
func (f *Formatter) SetExtras(extras FormatterExtras) error {
	v := C.GhosttyFormatterExtras(extras)
	return f.set(C.GHOSTTY_FORMATTER_OPT_EXTRAS, unsafe.Pointer(&v))
}

// Extras returns the terminal state added to VT output.
func (f *Formatter) Extras() (FormatterExtras, error) {
	var v C.GhosttyFormatterExtras
	if err := f.get(C.GHOSTTY_FORMATTER_OPT_EXTRAS, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return FormatterExtras(v), nil
}

// Close frees the formatter handle. After this call, the formatter
// must not be used, except that calling Close again is a safe no-op.
func (f *Formatter) Close() {
	if f == nil || f.ptr == nil {
		return
	}
	f.writer.close()
	C.ghostty_formatter_free(f.ptr)
	f.ptr = nil
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

// FormatBuf formats into buf and returns the number of bytes written. If buf
// is too small, the returned count is the required size and the error has
// result [ResultOutOfSpace]. A nil buffer can be used to query the required
// size. Output that is empty needs no space, so a query for it returns 0 and
// no error.
func (f *Formatter) FormatBuf(buf []byte) (int, error) {
	var ptr *C.uint8_t
	if len(buf) > 0 {
		ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
	}

	var written C.size_t
	result := C.ghostty_formatter_format_buf(
		f.ptr,
		ptr,
		C.size_t(len(buf)),
		&written,
	)
	return int(written), resultError(result)
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

// WriteTo implements io.WriterTo by formatting the current terminal state and
// writing it to w. It does not allocate a buffer for the complete result. The
// returned count includes bytes accepted before an error.
//
// WriteTo buffers small formatter writes to reduce calls into Go. The buffer
// is flushed when full and when formatting completes. Since libghostty formats
// quickly, buffering is not expected to add noticeable latency. WriteTo may
// block if w blocks. The writer must not call methods on f or its terminal.
// C: ghostty_formatter_format
func (f *Formatter) WriteTo(w io.Writer) (int64, error) {
	bridge := &f.writer
	if err := bridge.reset(w); err != nil {
		return 0, err
	}

	result := C.ghostty_go_formatter_format(f.ptr, C.uintptr_t(bridge.handle))
	written := bridge.written
	callbackErr := bridge.err
	bridge.writer = nil
	return written, resultErrorWithCallback(result, callbackErr)
}
