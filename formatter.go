package libghostty

/*
#include <ghostty/vt.h>
#include <string.h>

// Helper to create a properly initialized GhosttyFormatterTerminalOptions (sized struct).
static inline GhosttyFormatterTerminalOptions init_formatter_terminal_options() {
	GhosttyFormatterTerminalOptions opts = GHOSTTY_INIT_SIZED(GhosttyFormatterTerminalOptions);
	opts.extra.size = sizeof(GhosttyFormatterTerminalExtra);
	opts.extra.screen.size = sizeof(GhosttyFormatterScreenExtra);
	return opts;
}

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
	GhosttyWriter downstream
) {
	uint8_t buffer[16 * 1024];
	ghostty_go_formatter_writer context = {
		.downstream = downstream,
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

// formatterOpts wraps the C options struct so that functional options
// can mutate it directly. Only fields explicitly set by an option are
// modified; everything else retains the GHOSTTY_INIT_SIZED defaults.
type formatterOpts struct {
	c         C.GhosttyFormatterTerminalOptions
	selection *Selection
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

// WithFormatterSelection restricts formatter output to sel. Passing nil
// formats the entire active screen.
//
// The selection must come from the same terminal passed to [NewFormatter] and
// must still be valid when NewFormatter is called. NewFormatter copies the
// selection immediately, so the Selection value itself does not need to
// outlive that call. The copied selection still contains borrowed grid
// references into the terminal; as with every [Selection], later terminal
// mutations may invalidate those references.
func WithFormatterSelection(sel *Selection) FormatterOption {
	return func(o *formatterOpts) {
		o.selection = sel
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

// prepare materializes the selection option in C-owned memory for the
// formatter constructor. libghostty copies the selection into the formatter
// during ghostty_formatter_terminal_new, so this allocation only needs to
// remain alive for that call. Using C-owned memory also avoids placing a Go
// pointer inside the C options struct passed through cgo.
func (o *formatterOpts) prepare() (func(), error) {
	if o.selection == nil {
		return func() {}, nil
	}

	size := uintptr(C.sizeof_GhosttySelection)
	ptr := Alloc(size)
	if ptr == nil {
		return nil, &Error{Result: ResultOutOfMemory}
	}
	csel := (*C.GhosttySelection)(ptr)
	*csel = o.selection.toC()
	o.c.selection = csel

	return func() {
		Free(ptr, size)
	}, nil
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
func NewFormatter(t *Terminal, opts ...FormatterOption) (*Formatter, error) {
	// Start with GHOSTTY_INIT_SIZED defaults; options only touch
	// fields the caller explicitly sets.
	fo := formatterOpts{c: C.init_formatter_terminal_options()}
	for _, opt := range opts {
		opt(&fo)
	}
	cleanup, err := fo.prepare()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var ptr C.GhosttyFormatter
	if err := resultError(C.ghostty_formatter_terminal_new(nil, &ptr, t.ptr, fo.c)); err != nil {
		return nil, err
	}

	return &Formatter{ptr: ptr}, nil
}

// Close frees the formatter handle. After this call, the formatter
// must not be used.
func (f *Formatter) Close() {
	f.writer.close()
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

// FormatBuf formats into buf and returns the number of bytes written. If buf
// is too small, the returned count is the required size and the error has
// result [ResultOutOfSpace]. A nil buffer can be used to query the required
// size.
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
	writer, err := bridge.reset(w)
	if err != nil {
		return 0, err
	}

	result := C.ghostty_go_formatter_format(f.ptr, writer)
	written := bridge.written
	callbackErr := bridge.err
	bridge.finish()
	return written, resultErrorWithCallback(result, callbackErr)
}
