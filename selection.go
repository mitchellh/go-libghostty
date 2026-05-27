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

// SelectionGesture is an opaque handle to mutable state for interpreting
// pointer gestures as terminal selection snapshots. It is not safe for
// concurrent use and should be serialized with the terminal it is used with.
// C: GhosttySelectionGesture
type SelectionGesture struct {
	ptr C.GhosttySelectionGesture
}

// SelectionGestureEvent is an opaque, reusable input event for a selection
// gesture. Its event type is fixed at creation time, while its option values
// may be changed between calls.
// C: GhosttySelectionGestureEvent
type SelectionGestureEvent struct {
	ptr C.GhosttySelectionGestureEvent
}

// SelectionGestureBehavior is the selection behavior chosen for a gesture's
// click sequence.
// C: GhosttySelectionGestureBehavior
type SelectionGestureBehavior int

const (
	// SelectionGestureBehaviorCell performs cell-granular drag selection.
	SelectionGestureBehaviorCell SelectionGestureBehavior = C.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_CELL

	// SelectionGestureBehaviorWord performs word selection on press and
	// word-granular drag selection.
	SelectionGestureBehaviorWord SelectionGestureBehavior = C.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_WORD

	// SelectionGestureBehaviorLine performs line selection on press and
	// line-granular drag selection.
	SelectionGestureBehaviorLine SelectionGestureBehavior = C.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_LINE

	// SelectionGestureBehaviorOutput performs semantic command-output selection
	// on press and drag.
	SelectionGestureBehaviorOutput SelectionGestureBehavior = C.GHOSTTY_SELECTION_GESTURE_BEHAVIOR_OUTPUT
)

// SelectionGestureBehaviors configures the behavior selected for single-,
// double-, and triple-click gesture sequences.
// C: GhosttySelectionGestureBehaviors
type SelectionGestureBehaviors struct {
	// SingleClick is the behavior for single-click selection gestures.
	SingleClick SelectionGestureBehavior

	// DoubleClick is the behavior for double-click selection gestures.
	DoubleClick SelectionGestureBehavior

	// TripleClick is the behavior for triple-click selection gestures.
	TripleClick SelectionGestureBehavior
}

// toC converts Go selection gesture behaviors to C.
func (b SelectionGestureBehaviors) toC() C.GhosttySelectionGestureBehaviors {
	return C.GhosttySelectionGestureBehaviors{
		single_click: C.GhosttySelectionGestureBehavior(b.SingleClick),
		double_click: C.GhosttySelectionGestureBehavior(b.DoubleClick),
		triple_click: C.GhosttySelectionGestureBehavior(b.TripleClick),
	}
}

// SelectionGestureGeometry is the display geometry used to interpret drag and
// autoscroll selection gesture events.
// C: GhosttySelectionGestureGeometry
type SelectionGestureGeometry struct {
	// Columns is the number of columns in the rendered terminal grid. It must
	// be non-zero.
	Columns uint32

	// CellWidth is the width of one terminal cell in surface pixels. It must be
	// non-zero.
	CellWidth uint32

	// PaddingLeft is the left padding before the terminal grid begins in surface
	// pixels.
	PaddingLeft uint32

	// ScreenHeight is the height of the rendered terminal surface in surface
	// pixels. It must be non-zero.
	ScreenHeight uint32
}

// toC converts Go selection gesture geometry to C.
func (g SelectionGestureGeometry) toC() C.GhosttySelectionGestureGeometry {
	return C.GhosttySelectionGestureGeometry{
		columns:       C.uint32_t(g.Columns),
		cell_width:    C.uint32_t(g.CellWidth),
		padding_left:  C.uint32_t(g.PaddingLeft),
		screen_height: C.uint32_t(g.ScreenHeight),
	}
}

// SelectionGestureAutoscroll is the current autoscroll direction for an active
// selection drag gesture.
// C: GhosttySelectionGestureAutoscroll
type SelectionGestureAutoscroll int

const (
	// SelectionGestureAutoscrollNone means no selection autoscroll is requested.
	SelectionGestureAutoscrollNone SelectionGestureAutoscroll = C.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_NONE

	// SelectionGestureAutoscrollUp means selection dragging should autoscroll
	// the viewport upward.
	SelectionGestureAutoscrollUp SelectionGestureAutoscroll = C.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_UP

	// SelectionGestureAutoscrollDown means selection dragging should autoscroll
	// the viewport downward.
	SelectionGestureAutoscrollDown SelectionGestureAutoscroll = C.GHOSTTY_SELECTION_GESTURE_AUTOSCROLL_DOWN
)

// SelectionGestureData identifies data fields readable from a selection
// gesture.
// C: GhosttySelectionGestureData
type SelectionGestureData int

const (
	// SelectionGestureDataClickCount reads the current click count as *uint8.
	SelectionGestureDataClickCount SelectionGestureData = C.GHOSTTY_SELECTION_GESTURE_DATA_CLICK_COUNT

	// SelectionGestureDataDragged reads whether the current or last left-click
	// gesture dragged as *bool.
	SelectionGestureDataDragged SelectionGestureData = C.GHOSTTY_SELECTION_GESTURE_DATA_DRAGGED

	// SelectionGestureDataAutoscroll reads the current autoscroll request as a
	// *GhosttySelectionGestureAutoscroll-compatible value.
	SelectionGestureDataAutoscroll SelectionGestureData = C.GHOSTTY_SELECTION_GESTURE_DATA_AUTOSCROLL

	// SelectionGestureDataBehavior reads the current gesture behavior as a
	// *GhosttySelectionGestureBehavior-compatible value.
	SelectionGestureDataBehavior SelectionGestureData = C.GHOSTTY_SELECTION_GESTURE_DATA_BEHAVIOR

	// SelectionGestureDataAnchor reads the current left-click anchor as a
	// *GhosttyGridRef-compatible value. It returns ResultNoValue when there is
	// no valid active anchor.
	SelectionGestureDataAnchor SelectionGestureData = C.GHOSTTY_SELECTION_GESTURE_DATA_ANCHOR
)

// SelectionGestureEventType is the fixed type of a selection gesture event.
// C: GhosttySelectionGestureEventType
type SelectionGestureEventType int

const (
	// SelectionGestureEventTypePress is a press event.
	SelectionGestureEventTypePress SelectionGestureEventType = C.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_PRESS

	// SelectionGestureEventTypeRelease is a release event.
	SelectionGestureEventTypeRelease SelectionGestureEventType = C.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_RELEASE

	// SelectionGestureEventTypeDrag is a drag event.
	SelectionGestureEventTypeDrag SelectionGestureEventType = C.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_DRAG

	// SelectionGestureEventTypeAutoscrollTick is an autoscroll tick event.
	SelectionGestureEventTypeAutoscrollTick SelectionGestureEventType = C.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_AUTOSCROLL_TICK

	// SelectionGestureEventTypeDeepPress is a deep press event.
	SelectionGestureEventTypeDeepPress SelectionGestureEventType = C.GHOSTTY_SELECTION_GESTURE_EVENT_TYPE_DEEP_PRESS
)

// SelectionGestureEventOption identifies an option stored on a reusable
// selection gesture event.
// C: GhosttySelectionGestureEventOption
type SelectionGestureEventOption int

const (
	// SelectionGestureEventOptionRef stores a *GridRef pointer under the
	// pointer. It is required for press and drag events and optional for release.
	SelectionGestureEventOptionRef SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REF

	// SelectionGestureEventOptionPosition stores a *SurfacePosition pointer
	// position.
	SelectionGestureEventOptionPosition SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_POSITION

	// SelectionGestureEventOptionRepeatDistance stores a *float64 maximum
	// repeat-click distance in pixels.
	SelectionGestureEventOptionRepeatDistance SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_DISTANCE

	// SelectionGestureEventOptionTimeNS stores a *uint64 monotonic event time in
	// nanoseconds.
	SelectionGestureEventOptionTimeNS SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_TIME_NS

	// SelectionGestureEventOptionRepeatIntervalNS stores a *uint64 maximum
	// interval between repeat clicks in nanoseconds.
	SelectionGestureEventOptionRepeatIntervalNS SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_INTERVAL_NS

	// SelectionGestureEventOptionWordBoundaryCodepoints stores borrowed word
	// boundary codepoints that libghostty copies into the event.
	SelectionGestureEventOptionWordBoundaryCodepoints SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_WORD_BOUNDARY_CODEPOINTS

	// SelectionGestureEventOptionBehaviors stores a *SelectionGestureBehaviors
	// behavior table.
	SelectionGestureEventOptionBehaviors SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_BEHAVIORS

	// SelectionGestureEventOptionRectangle stores a *bool indicating whether a
	// drag or autoscroll tick should produce a rectangular selection.
	SelectionGestureEventOptionRectangle SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_RECTANGLE

	// SelectionGestureEventOptionGeometry stores a *SelectionGestureGeometry
	// drag display geometry.
	SelectionGestureEventOptionGeometry SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_GEOMETRY

	// SelectionGestureEventOptionViewport stores a *PointCoordinate viewport
	// coordinate for an autoscroll tick.
	SelectionGestureEventOptionViewport SelectionGestureEventOption = C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_VIEWPORT
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

// NewSelectionGesture creates a selection gesture state machine. The returned
// gesture must be freed with [SelectionGesture.Close] when no longer needed.
func NewSelectionGesture() (*SelectionGesture, error) {
	var ptr C.GhosttySelectionGesture
	if err := resultError(C.ghostty_selection_gesture_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &SelectionGesture{ptr: ptr}, nil
}

// Close frees the selection gesture. If the terminal most recently used with
// the gesture is still alive, pass it so libghostty can release tracked
// terminal references. Pass nil only after the terminal has already been freed.
func (g *SelectionGesture) Close(t *Terminal) {
	var terminal C.GhosttyTerminal
	if t != nil {
		terminal = t.ptr
	}
	C.ghostty_selection_gesture_free(g.ptr, terminal)
}

// Reset cancels active selection gesture state and releases tracked terminal
// references without freeing the gesture.
func (g *SelectionGesture) Reset(t *Terminal) {
	var terminal C.GhosttyTerminal
	if t != nil {
		terminal = t.ptr
	}
	C.ghostty_selection_gesture_reset(g.ptr, terminal)
}

// Event applies a selection gesture event to the terminal and returns the
// resulting selection snapshot. It returns nil when the event updates gesture
// state but does not currently produce a selection.
func (g *SelectionGesture) Event(t *Terminal, event *SelectionGestureEvent) (*Selection, error) {
	cs := initCSelection()
	if err := resultError(C.ghostty_selection_gesture_event(g.ptr, t.ptr, event.ptr, &cs)); err != nil {
		return noValueNilSelection(err)
	}

	sel := selectionFromC(cs)
	return &sel, nil
}

// Get reads a raw data field from the selection gesture into value. The value
// pointer type must match the documented type for the data key; prefer the
// typed helpers such as [SelectionGesture.ClickCount] and
// [SelectionGesture.Anchor] when possible.
// C: ghostty_selection_gesture_get
func (g *SelectionGesture) Get(t *Terminal, data SelectionGestureData, value unsafe.Pointer) error {
	return resultError(C.ghostty_selection_gesture_get(
		g.ptr,
		t.ptr,
		C.GhosttySelectionGestureData(data),
		value,
	))
}

// GetMulti reads multiple raw data fields from the selection gesture in one C
// call. Each element in values must point to storage with the type documented
// by the corresponding key. On partial failure, written is the failing index.
// C: ghostty_selection_gesture_get_multi
func (g *SelectionGesture) GetMulti(t *Terminal, keys []SelectionGestureData, values []unsafe.Pointer) (written int, err error) {
	if len(keys) != len(values) {
		return 0, errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	cKeys := make([]C.GhosttySelectionGestureData, len(keys))
	for i, key := range keys {
		cKeys[i] = C.GhosttySelectionGestureData(key)
	}
	cVals, cValsSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cVals), cValsSize)
	var cWritten C.size_t
	err = resultError(C.ghostty_selection_gesture_get_multi(
		g.ptr,
		t.ptr,
		C.size_t(len(keys)),
		(*C.GhosttySelectionGestureData)(unsafe.Pointer(&cKeys[0])),
		cVals,
		&cWritten,
	))
	return int(cWritten), err
}

// ClickCount returns the current gesture click count. Zero means inactive.
func (g *SelectionGesture) ClickCount(t *Terminal) (uint8, error) {
	var v C.uint8_t
	if err := g.Get(t, SelectionGestureDataClickCount, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return uint8(v), nil
}

// Dragged reports whether the current or last left-click gesture dragged.
func (g *SelectionGesture) Dragged(t *Terminal) (bool, error) {
	var v C.bool
	if err := g.Get(t, SelectionGestureDataDragged, unsafe.Pointer(&v)); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Autoscroll returns the current autoscroll request for the gesture.
func (g *SelectionGesture) Autoscroll(t *Terminal) (SelectionGestureAutoscroll, error) {
	var v C.GhosttySelectionGestureAutoscroll
	if err := g.Get(t, SelectionGestureDataAutoscroll, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return SelectionGestureAutoscroll(v), nil
}

// Behavior returns the current gesture behavior.
func (g *SelectionGesture) Behavior(t *Terminal) (SelectionGestureBehavior, error) {
	var v C.GhosttySelectionGestureBehavior
	if err := g.Get(t, SelectionGestureDataBehavior, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return SelectionGestureBehavior(v), nil
}

// Anchor returns the current left-click anchor as an untracked grid reference
// snapshot. It returns nil when there is no valid active anchor.
func (g *SelectionGesture) Anchor(t *Terminal) (*GridRef, error) {
	ref := initCGridRef()
	if err := g.Get(t, SelectionGestureDataAnchor, unsafe.Pointer(&ref)); err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}
	return &GridRef{ref: ref}, nil
}

// NewSelectionGestureEvent creates a reusable gesture event of a fixed type.
// The returned event must be freed with [SelectionGestureEvent.Close].
func NewSelectionGestureEvent(eventType SelectionGestureEventType) (*SelectionGestureEvent, error) {
	var ptr C.GhosttySelectionGestureEvent
	if err := resultError(C.ghostty_selection_gesture_event_new(
		nil,
		&ptr,
		C.GhosttySelectionGestureEventType(eventType),
	)); err != nil {
		return nil, err
	}
	return &SelectionGestureEvent{ptr: ptr}, nil
}

// Close frees the selection gesture event. After this call, the event must not
// be used.
func (e *SelectionGestureEvent) Close() {
	C.ghostty_selection_gesture_event_free(e.ptr)
}

// ClearOption clears a selection gesture event option.
func (e *SelectionGestureEvent) ClearOption(option SelectionGestureEventOption) error {
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GhosttySelectionGestureEventOption(option),
		nil,
	))
}

// SetRef sets the grid reference under the pointer for a press, drag, or
// release event.
func (e *SelectionGestureEvent) SetRef(ref *GridRef) error {
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REF,
		unsafe.Pointer(&ref.ref),
	))
}

// ClearRef clears the optional grid reference under the pointer.
func (e *SelectionGestureEvent) ClearRef() error {
	return e.ClearOption(SelectionGestureEventOptionRef)
}

// SetPosition sets the surface-space pointer position for a press, drag, or
// autoscroll tick event.
func (e *SelectionGestureEvent) SetPosition(position SurfacePosition) error {
	cp := position.toC()
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_POSITION,
		unsafe.Pointer(&cp),
	))
}

// ClearPosition clears the surface-space pointer position.
func (e *SelectionGestureEvent) ClearPosition() error {
	return e.ClearOption(SelectionGestureEventOptionPosition)
}

// SetRepeatDistance sets the maximum repeat-click distance in pixels.
func (e *SelectionGestureEvent) SetRepeatDistance(distance float64) error {
	v := C.double(distance)
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_DISTANCE,
		unsafe.Pointer(&v),
	))
}

// ClearRepeatDistance clears the maximum repeat-click distance.
func (e *SelectionGestureEvent) ClearRepeatDistance() error {
	return e.ClearOption(SelectionGestureEventOptionRepeatDistance)
}

// SetTimeNS sets the optional monotonic event time in nanoseconds.
func (e *SelectionGestureEvent) SetTimeNS(ns uint64) error {
	v := C.uint64_t(ns)
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_TIME_NS,
		unsafe.Pointer(&v),
	))
}

// ClearTimeNS clears the optional monotonic event time.
func (e *SelectionGestureEvent) ClearTimeNS() error {
	return e.ClearOption(SelectionGestureEventOptionTimeNS)
}

// SetRepeatIntervalNS sets the maximum interval between repeat clicks in
// nanoseconds.
func (e *SelectionGestureEvent) SetRepeatIntervalNS(ns uint64) error {
	v := C.uint64_t(ns)
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_REPEAT_INTERVAL_NS,
		unsafe.Pointer(&v),
	))
}

// ClearRepeatIntervalNS clears the maximum interval between repeat clicks.
func (e *SelectionGestureEvent) ClearRepeatIntervalNS() error {
	return e.ClearOption(SelectionGestureEventOptionRepeatIntervalNS)
}

// SetWordBoundaryCodepoints sets custom word-boundary codepoints for events
// that derive word selections. Libghostty copies the values into the event.
func (e *SelectionGestureEvent) SetWordBoundaryCodepoints(codepoints []uint32) error {
	ptr, size := cUint32Slice(codepoints)
	if len(codepoints) > 0 && ptr == nil {
		return &Error{Result: ResultOutOfMemory}
	}
	if size > 0 {
		defer Free(unsafe.Pointer(ptr), size)
	}

	cp := C.GhosttyCodepoints{
		ptr: (*C.uint32_t)(ptr),
		len: C.size_t(len(codepoints)),
	}
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_WORD_BOUNDARY_CODEPOINTS,
		unsafe.Pointer(&cp),
	))
}

// ClearWordBoundaryCodepoints clears custom word-boundary codepoints so the
// event uses Ghostty's defaults again.
func (e *SelectionGestureEvent) ClearWordBoundaryCodepoints() error {
	return e.ClearOption(SelectionGestureEventOptionWordBoundaryCodepoints)
}

// SetBehaviors sets the behavior table for a press event.
func (e *SelectionGestureEvent) SetBehaviors(behaviors SelectionGestureBehaviors) error {
	cb := behaviors.toC()
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_BEHAVIORS,
		unsafe.Pointer(&cb),
	))
}

// ClearBehaviors clears custom press behaviors so Ghostty's default behavior
// table is used again.
func (e *SelectionGestureEvent) ClearBehaviors() error {
	return e.ClearOption(SelectionGestureEventOptionBehaviors)
}

// SetRectangle sets whether a drag or autoscroll tick should produce a
// rectangular selection.
func (e *SelectionGestureEvent) SetRectangle(rectangle bool) error {
	v := C.bool(rectangle)
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_RECTANGLE,
		unsafe.Pointer(&v),
	))
}

// ClearRectangle clears the rectangular-selection setting.
func (e *SelectionGestureEvent) ClearRectangle() error {
	return e.ClearOption(SelectionGestureEventOptionRectangle)
}

// SetGeometry sets the display geometry required by drag and autoscroll tick
// events.
func (e *SelectionGestureEvent) SetGeometry(geometry SelectionGestureGeometry) error {
	cg := geometry.toC()
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_GEOMETRY,
		unsafe.Pointer(&cg),
	))
}

// ClearGeometry clears the display geometry.
func (e *SelectionGestureEvent) ClearGeometry() error {
	return e.ClearOption(SelectionGestureEventOptionGeometry)
}

// SetViewport sets the viewport coordinate required by autoscroll tick events.
func (e *SelectionGestureEvent) SetViewport(viewport PointCoordinate) error {
	cp := viewport.toC()
	return resultError(C.ghostty_selection_gesture_event_set(
		e.ptr,
		C.GHOSTTY_SELECTION_GESTURE_EVENT_OPT_VIEWPORT,
		unsafe.Pointer(&cp),
	))
}

// ClearViewport clears the autoscroll tick viewport coordinate.
func (e *SelectionGestureEvent) ClearViewport() error {
	return e.ClearOption(SelectionGestureEventOptionViewport)
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
