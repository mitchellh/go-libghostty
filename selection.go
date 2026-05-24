package libghostty

// Selection helpers wrapping the C API from selection.h.

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyTerminalSelectWordOptions
// (sized struct).
static inline GhosttyTerminalSelectWordOptions init_terminal_select_word_options() {
	GhosttyTerminalSelectWordOptions opts = GHOSTTY_INIT_SIZED(GhosttyTerminalSelectWordOptions);
	return opts;
}

// Helper to create a properly initialized
// GhosttyTerminalSelectWordBetweenOptions (sized struct).
static inline GhosttyTerminalSelectWordBetweenOptions init_terminal_select_word_between_options() {
	GhosttyTerminalSelectWordBetweenOptions opts = GHOSTTY_INIT_SIZED(GhosttyTerminalSelectWordBetweenOptions);
	return opts;
}

// Helper to create a properly initialized GhosttyTerminalSelectLineOptions
// (sized struct).
static inline GhosttyTerminalSelectLineOptions init_terminal_select_line_options() {
	GhosttyTerminalSelectLineOptions opts = GHOSTTY_INIT_SIZED(GhosttyTerminalSelectLineOptions);
	return opts;
}

// Helper to create a properly initialized
// GhosttyTerminalSelectionFormatOptions (sized struct).
static inline GhosttyTerminalSelectionFormatOptions init_terminal_selection_format_options() {
	GhosttyTerminalSelectionFormatOptions opts = GHOSTTY_INIT_SIZED(GhosttyTerminalSelectionFormatOptions);
	return opts;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// SelectionOrder describes the ordering of selection endpoints in
// terminal coordinates.
// C: GhosttySelectionOrder
type SelectionOrder int

const (
	// SelectionOrderForward means Start is before End in top-left to
	// bottom-right order.
	SelectionOrderForward SelectionOrder = C.GHOSTTY_SELECTION_ORDER_FORWARD

	// SelectionOrderReverse means End is before Start in top-left to
	// bottom-right order.
	SelectionOrderReverse SelectionOrder = C.GHOSTTY_SELECTION_ORDER_REVERSE

	// SelectionOrderMirroredForward means a rectangular selection from
	// top-right to bottom-left.
	SelectionOrderMirroredForward SelectionOrder = C.GHOSTTY_SELECTION_ORDER_MIRRORED_FORWARD

	// SelectionOrderMirroredReverse means a rectangular selection from
	// bottom-left to top-right.
	SelectionOrderMirroredReverse SelectionOrder = C.GHOSTTY_SELECTION_ORDER_MIRRORED_REVERSE
)

// SelectionAdjust describes an operation used to adjust a selection's
// logical end endpoint.
// C: GhosttySelectionAdjust
type SelectionAdjust int

const (
	// SelectionAdjustLeft moves left to the previous non-empty cell,
	// wrapping upward.
	SelectionAdjustLeft SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_LEFT

	// SelectionAdjustRight moves right to the next non-empty cell,
	// wrapping downward.
	SelectionAdjustRight SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_RIGHT

	// SelectionAdjustUp moves up one row at the current column, or to the
	// beginning of the line if already at the top.
	SelectionAdjustUp SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_UP

	// SelectionAdjustDown moves down to the next non-blank row at the
	// current column, or to the end of the line if none exists.
	SelectionAdjustDown SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_DOWN

	// SelectionAdjustHome moves to the top-left cell of the screen.
	SelectionAdjustHome SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_HOME

	// SelectionAdjustEnd moves to the right edge of the last non-blank row
	// on the screen.
	SelectionAdjustEnd SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_END

	// SelectionAdjustPageUp moves up by one terminal page height, or to
	// home if that would move past the top.
	SelectionAdjustPageUp SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_PAGE_UP

	// SelectionAdjustPageDown moves down by one terminal page height, or
	// to end if that would move past the bottom.
	SelectionAdjustPageDown SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_PAGE_DOWN

	// SelectionAdjustBeginningOfLine moves to the left edge of the current
	// line.
	SelectionAdjustBeginningOfLine SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_BEGINNING_OF_LINE

	// SelectionAdjustEndOfLine moves to the right edge of the current line.
	SelectionAdjustEndOfLine SelectionAdjust = C.GHOSTTY_SELECTION_ADJUST_END_OF_LINE
)

// SelectWordOptions configures [Terminal.SelectWord].
// C: GhosttyTerminalSelectWordOptions
type SelectWordOptions struct {
	// Ref is the grid reference under which to derive the word selection.
	Ref *GridRef

	// BoundaryCodepoints optionally replaces Ghostty's default word-boundary
	// codepoints. Leave nil to use Ghostty's defaults.
	BoundaryCodepoints []uint32
}

// SelectWordBetweenOptions configures [Terminal.SelectWordBetween].
// C: GhosttyTerminalSelectWordBetweenOptions
type SelectWordBetweenOptions struct {
	// Start is the starting grid reference for the inclusive search range.
	Start *GridRef

	// End is the ending grid reference for the inclusive search range.
	End *GridRef

	// BoundaryCodepoints optionally replaces Ghostty's default word-boundary
	// codepoints. Leave nil to use Ghostty's defaults.
	BoundaryCodepoints []uint32
}

// SelectLineOptions configures [Terminal.SelectLine].
// C: GhosttyTerminalSelectLineOptions
type SelectLineOptions struct {
	// Ref is the grid reference under which to derive the line selection.
	Ref *GridRef

	// Whitespace optionally replaces Ghostty's default line-trim whitespace
	// codepoints. Leave nil to use Ghostty's defaults.
	Whitespace []uint32

	// SemanticPromptBoundary bounds line selection at semantic prompt state
	// changes when true.
	SemanticPromptBoundary bool
}

// SelectOutputOptions configures [Terminal.SelectOutput].
type SelectOutputOptions struct {
	// Ref is a grid reference within the command output to select.
	Ref *GridRef
}

// SelectionFormatOption configures one-shot terminal selection formatting.
type SelectionFormatOption func(*selectionFormatOpts)

// selectionFormatOpts wraps the C options struct so functional options can
// mutate only the fields the caller explicitly wants to set.
type selectionFormatOpts struct {
	c         C.GhosttyTerminalSelectionFormatOptions
	selection *Selection
}

// WithSelectionFormat sets the output format (plain, VT, or HTML).
// Defaults to FormatterFormatPlain if not specified.
func WithSelectionFormat(f FormatterFormat) SelectionFormatOption {
	return func(o *selectionFormatOpts) {
		o.c.emit = C.GhosttyFormatterFormat(f)
	}
}

// WithSelectionUnwrap enables unwrapping of soft-wrapped lines.
func WithSelectionUnwrap(unwrap bool) SelectionFormatOption {
	return func(o *selectionFormatOpts) {
		o.c.unwrap = C.bool(unwrap)
	}
}

// WithSelectionTrim enables trimming trailing whitespace on non-blank
// lines.
func WithSelectionTrim(trim bool) SelectionFormatOption {
	return func(o *selectionFormatOpts) {
		o.c.trim = C.bool(trim)
	}
}

// WithSelection formats the provided selection snapshot instead of the
// terminal's active selection. The selection must come from the same terminal
// and remain valid for the duration of the format call.
func WithSelection(sel *Selection) SelectionFormatOption {
	return func(o *selectionFormatOpts) {
		o.selection = sel
	}
}

// toC converts a Go Selection into a C GhosttySelection sized struct.
func (s Selection) toC() C.GhosttySelection {
	cs := initCSelection()
	cs.start = s.Start.ref
	cs.end = s.End.ref
	cs.rectangle = C.bool(s.Rectangle)
	return cs
}

// noValueNilSelection turns C APIs that report GHOSTTY_NO_VALUE into a nil
// selection result, matching the existing Terminal.Selection behavior.
func noValueNilSelection(err error) (*Selection, error) {
	if err == nil {
		return nil, nil
	}
	var ge *Error
	if errors.As(err, &ge) && ge.Result == ResultNoValue {
		return nil, nil
	}
	return nil, err
}

// cUint32Slice copies codepoints into C-owned memory for the duration of a
// call. This avoids storing Go pointers inside C option structs, which cgo's
// pointer checks correctly reject.
func cUint32Slice(codepoints []uint32) (*C.uint32_t, uintptr) {
	if len(codepoints) == 0 {
		return nil, 0
	}
	size := uintptr(len(codepoints)) * C.sizeof_uint32_t
	ptr := Alloc(size)
	if ptr == nil {
		return nil, size
	}
	copy(unsafe.Slice((*uint32)(ptr), len(codepoints)), codepoints)
	return (*C.uint32_t)(ptr), size
}

// prepare materializes pointer-valued selection format options in C-owned
// memory and returns a cleanup function. The C API reads the options only for
// the duration of a single call, so temporary C memory is enough.
func (o *selectionFormatOpts) prepare() (func(), error) {
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

// SelectWord derives a word selection snapshot from a terminal grid reference.
// The returned selection is not installed as the terminal's active selection.
// It returns nil when the reference has no selectable word content.
//
// Options must contain a Ref. BoundaryCodepoints optionally replaces
// Ghostty's default word-boundary codepoints.
func (t *Terminal) SelectWord(options SelectWordOptions) (*Selection, error) {
	opts := C.init_terminal_select_word_options()
	opts.ref = options.Ref.ref
	if len(options.BoundaryCodepoints) > 0 {
		ptr, size := cUint32Slice(options.BoundaryCodepoints)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(unsafe.Pointer(ptr), size)
		opts.boundary_codepoints = ptr
		opts.boundary_codepoints_len = C.size_t(len(options.BoundaryCodepoints))
	}

	cs := initCSelection()
	if err := resultError(C.ghostty_terminal_select_word(t.ptr, &opts, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// SelectWordBetween derives the nearest word selection snapshot between two
// terminal grid references. The search starts at start and moves toward end,
// inclusive. It returns nil when no selectable word content exists between
// the references.
//
// Options must contain Start and End references. BoundaryCodepoints
// optionally replaces Ghostty's default word-boundary codepoints.
func (t *Terminal) SelectWordBetween(options SelectWordBetweenOptions) (*Selection, error) {
	opts := C.init_terminal_select_word_between_options()
	opts.start = options.Start.ref
	opts.end = options.End.ref
	if len(options.BoundaryCodepoints) > 0 {
		ptr, size := cUint32Slice(options.BoundaryCodepoints)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(unsafe.Pointer(ptr), size)
		opts.boundary_codepoints = ptr
		opts.boundary_codepoints_len = C.size_t(len(options.BoundaryCodepoints))
	}

	cs := initCSelection()
	if err := resultError(C.ghostty_terminal_select_word_between(t.ptr, &opts, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// SelectLine derives a line selection snapshot from a terminal grid reference.
// The returned selection is not installed as the terminal's active selection.
// It returns nil when the reference has no selectable line content.
//
// Options must contain a Ref. Whitespace optionally replaces Ghostty's default
// line-trim whitespace codepoints. SemanticPromptBoundary controls whether
// semantic prompt state changes bound the line selection.
func (t *Terminal) SelectLine(options SelectLineOptions) (*Selection, error) {
	opts := C.init_terminal_select_line_options()
	opts.ref = options.Ref.ref
	opts.semantic_prompt_boundary = C.bool(options.SemanticPromptBoundary)
	if len(options.Whitespace) > 0 {
		ptr, size := cUint32Slice(options.Whitespace)
		if ptr == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		defer Free(unsafe.Pointer(ptr), size)
		opts.whitespace = ptr
		opts.whitespace_len = C.size_t(len(options.Whitespace))
	}

	cs := initCSelection()
	if err := resultError(C.ghostty_terminal_select_line(t.ptr, &opts, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// SelectAll derives a selection snapshot covering all selectable terminal
// content. The returned selection is not installed as the terminal's active
// selection. It returns nil when no selectable content exists.
func (t *Terminal) SelectAll() (*Selection, error) {
	cs := initCSelection()
	if err := resultError(C.ghostty_terminal_select_all(t.ptr, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// SelectOutput derives a command-output selection snapshot from a terminal
// grid reference. The returned selection is not installed as the terminal's
// active selection. It returns nil when the reference is not selectable
// command output.
func (t *Terminal) SelectOutput(options SelectOutputOptions) (*Selection, error) {
	cs := initCSelection()
	if err := resultError(C.ghostty_terminal_select_output(t.ptr, options.Ref.ref, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// SelectionFormat formats either the terminal's active selection or the
// selection supplied with WithSelection into an allocated byte slice. It
// returns nil when formatting the active selection and no selection is active.
func (t *Terminal) SelectionFormat(opts ...SelectionFormatOption) ([]byte, error) {
	fo := selectionFormatOpts{c: C.init_terminal_selection_format_options()}
	for _, opt := range opts {
		opt(&fo)
	}
	cleanup, err := fo.prepare()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var outPtr *C.uint8_t
	var outLen C.size_t
	err = resultError(C.ghostty_terminal_selection_format_alloc(t.ptr, nil, fo.c, &outPtr, &outLen))
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	defer C.ghostty_free(nil, outPtr, outLen)

	if outLen == 0 {
		return nil, nil
	}
	return C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen)), nil
}

// SelectionFormatString is a convenience wrapper around SelectionFormat that
// returns the formatted selection as a string.
func (t *Terminal) SelectionFormatString(opts ...SelectionFormatOption) (string, error) {
	b, err := t.SelectionFormat(opts...)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SelectionFormatBuf formats either the terminal's active selection or the
// selection supplied with WithSelection into buf. It returns the number of
// bytes written. If buf is nil or too small, it returns ResultOutOfSpace and
// the required size.
func (t *Terminal) SelectionFormatBuf(buf []byte, opts ...SelectionFormatOption) (int, error) {
	fo := selectionFormatOpts{c: C.init_terminal_selection_format_options()}
	for _, opt := range opts {
		opt(&fo)
	}
	cleanup, err := fo.prepare()
	if err != nil {
		return 0, err
	}
	defer cleanup()

	var ptr *C.uint8_t
	if len(buf) > 0 {
		ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
	}

	var written C.size_t
	err = resultError(C.ghostty_terminal_selection_format_buf(
		t.ptr,
		fo.c,
		ptr,
		C.size_t(len(buf)),
		&written,
	))
	if err != nil {
		return int(written), err
	}
	return int(written), nil
}

// SelectionAdjust mutates a selection snapshot using terminal selection
// semantics. It adjusts the logical end endpoint regardless of whether the
// selection is visually forward or reversed.
func (t *Terminal) SelectionAdjust(sel *Selection, adjustment SelectionAdjust) error {
	cs := sel.toC()
	if err := resultError(C.ghostty_terminal_selection_adjust(
		t.ptr,
		&cs,
		C.GhosttySelectionAdjust(adjustment),
	)); err != nil {
		return err
	}

	*sel = selectionFromC(cs)
	return nil
}

// SelectionOrder returns the current endpoint ordering of a selection
// snapshot.
func (t *Terminal) SelectionOrder(sel *Selection) (SelectionOrder, error) {
	cs := sel.toC()
	var order C.GhosttySelectionOrder
	if err := resultError(C.ghostty_terminal_selection_order(t.ptr, &cs, &order)); err != nil {
		return 0, err
	}
	return SelectionOrder(order), nil
}

// SelectionOrdered returns a fresh selection snapshot with endpoints ordered
// as requested. Mirrored desired orders are accepted by Ghostty but normalize
// the same as forward ordering.
func (t *Terminal) SelectionOrdered(sel *Selection, desired SelectionOrder) (*Selection, error) {
	cs := sel.toC()
	out := initCSelection()
	if err := resultError(C.ghostty_terminal_selection_ordered(
		t.ptr,
		&cs,
		C.GhosttySelectionOrder(desired),
		&out,
	)); err != nil {
		return nil, err
	}

	ordered := selectionFromC(out)
	return &ordered, nil
}

// SelectionContains reports whether point is inside a selection snapshot,
// using the terminal's linear and rectangular selection semantics.
func (t *Terminal) SelectionContains(sel *Selection, point Point) (bool, error) {
	cs := sel.toC()
	var contains C.bool
	if err := resultError(C.ghostty_terminal_selection_contains(
		t.ptr,
		&cs,
		point.toC(),
		&contains,
	)); err != nil {
		return false, err
	}
	return bool(contains), nil
}

// SelectionEqual reports whether two selection snapshots are equal according
// to the terminal's internal selection semantics.
func (t *Terminal) SelectionEqual(a *Selection, b *Selection) (bool, error) {
	ca := a.toC()
	cb := b.toC()
	var equal C.bool
	if err := resultError(C.ghostty_terminal_selection_equal(t.ptr, &ca, &cb, &equal)); err != nil {
		return false, err
	}
	return bool(equal), nil
}
