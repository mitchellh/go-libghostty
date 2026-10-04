package libghostty

// Render state data getters and setters wrapping
// ghostty_render_state_get() and ghostty_render_state_set().
// Functions are ordered alphabetically.

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyRenderStateCursor (sized struct).
static inline GhosttyRenderStateCursor init_render_state_cursor() {
	GhosttyRenderStateCursor c = GHOSTTY_INIT_SIZED(GhosttyRenderStateCursor);
	return c;
}

// Helper to create a properly initialized GhosttyRenderStateColors (sized struct).
static inline GhosttyRenderStateColors init_render_state_colors() {
	GhosttyRenderStateColors c = GHOSTTY_INIT_SIZED(GhosttyRenderStateColors);
	return c;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// RenderStateData identifies a data field for render state queries.
// C: GhosttyRenderStateData
type RenderStateData int

const (
	// RenderStateDataInvalid is an invalid / sentinel value.
	RenderStateDataInvalid RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_INVALID

	// RenderStateDataCols is the viewport width in cells (uint16_t).
	RenderStateDataCols RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLS

	// RenderStateDataRows is the viewport height in cells, excluding overscan (uint16_t).
	RenderStateDataRows RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_ROWS

	// RenderStateDataOverscan is the number of extra rows captured by the last update
	// (GhosttyRenderStateOverscan).
	RenderStateDataOverscan RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_OVERSCAN

	// RenderStateDataOverscanRequest is the request used for the next update
	// (GhosttyRenderStateOverscan).
	RenderStateDataOverscanRequest RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_OVERSCAN_REQUEST

	// RenderStateDataDirty is the current dirty state
	// (GhosttyRenderStateDirty).
	RenderStateDataDirty RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_DIRTY

	// RenderStateDataRowIterator populates a pre-allocated row iterator
	// (GhosttyRenderStateRowIterator).
	RenderStateDataRowIterator RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_ROW_ITERATOR

	// RenderStateDataColorBackground is the default/current background
	// color (GhosttyColorRgb).
	RenderStateDataColorBackground RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLOR_BACKGROUND

	// RenderStateDataColorForeground is the default/current foreground
	// color (GhosttyColorRgb).
	RenderStateDataColorForeground RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLOR_FOREGROUND

	// RenderStateDataColorCursor is the cursor color when explicitly set
	// (GhosttyColorRgb).
	RenderStateDataColorCursor RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR

	// RenderStateDataColorCursorHasValue indicates whether an explicit
	// cursor color is set (bool).
	RenderStateDataColorCursorHasValue RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR_HAS_VALUE

	// RenderStateDataColorPalette is the active 256-color palette
	// (GhosttyColorRgb[256]).
	RenderStateDataColorPalette RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLOR_PALETTE

	// RenderStateDataCursorVisualStyle is the visual style of the cursor
	// (GhosttyRenderStateCursorVisualStyle).
	RenderStateDataCursorVisualStyle RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISUAL_STYLE

	// RenderStateDataCursorVisible indicates whether the cursor is visible
	// based on terminal modes (bool).
	RenderStateDataCursorVisible RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISIBLE

	// RenderStateDataCursorBlinking indicates whether the cursor should
	// blink based on terminal modes (bool).
	RenderStateDataCursorBlinking RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_BLINKING

	// RenderStateDataCursorPasswordInput indicates whether the cursor is
	// at a password input field (bool).
	RenderStateDataCursorPasswordInput RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_PASSWORD_INPUT

	// RenderStateDataCursorViewportHasValue indicates whether the cursor
	// is visible within the viewport (bool).
	RenderStateDataCursorViewportHasValue RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_HAS_VALUE

	// RenderStateDataCursorViewportX is the cursor viewport x position
	// in cells (uint16_t).
	RenderStateDataCursorViewportX RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_X

	// RenderStateDataCursorViewportY is the cursor viewport y position
	// in cells (uint16_t).
	RenderStateDataCursorViewportY RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_Y

	// RenderStateDataCursorViewportWideTail indicates whether the cursor
	// is on the tail of a wide character (bool).
	RenderStateDataCursorViewportWideTail RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_WIDE_TAIL

	// RenderStateDataCursor is all cursor state in one sized struct
	// (GhosttyRenderStateCursor).
	RenderStateDataCursor RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_CURSOR

	// RenderStateDataColors is all render-state colors in one sized struct
	// (GhosttyRenderStateColors).
	RenderStateDataColors RenderStateData = C.GHOSTTY_RENDER_STATE_DATA_COLORS
)

// Cols returns the viewport width in cells.
func (rs *RenderState) Cols() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// ColorBackground returns the default/current background color.
func (rs *RenderState) ColorBackground() (ColorRGB, error) {
	var v C.GhosttyColorRgb
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLOR_BACKGROUND, unsafe.Pointer(&v))); err != nil {
		return ColorRGB{}, err
	}
	return ColorRGB{R: uint8(v.r), G: uint8(v.g), B: uint8(v.b)}, nil
}

// ColorCursor returns the cursor color when explicitly set by terminal
// state. Returns nil (without error) when no explicit cursor color is set.
func (rs *RenderState) ColorCursor() (*ColorRGB, error) {
	// Check whether a cursor color is set first.
	var has C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR_HAS_VALUE, unsafe.Pointer(&has))); err != nil {
		return nil, err
	}
	if !bool(has) {
		return nil, nil
	}

	var v C.GhosttyColorRgb
	err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLOR_CURSOR, unsafe.Pointer(&v)))
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultInvalidValue {
			return nil, nil
		}
		return nil, err
	}
	c := ColorRGB{R: uint8(v.r), G: uint8(v.g), B: uint8(v.b)}
	return &c, nil
}

// ColorForeground returns the default/current foreground color.
func (rs *RenderState) ColorForeground() (ColorRGB, error) {
	var v C.GhosttyColorRgb
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLOR_FOREGROUND, unsafe.Pointer(&v))); err != nil {
		return ColorRGB{}, err
	}
	return ColorRGB{R: uint8(v.r), G: uint8(v.g), B: uint8(v.b)}, nil
}

// ColorPalette returns the active 256-color palette.
func (rs *RenderState) ColorPalette() (*Palette, error) {
	var cp [PaletteSize]C.GhosttyColorRgb
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_COLOR_PALETTE, unsafe.Pointer(&cp[0]))); err != nil {
		return nil, err
	}
	var p Palette
	for i, c := range cp {
		p[i] = ColorRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}
	}
	return &p, nil
}

// Colors returns all color information from the render state in a
// single call using the sized-struct API.
func (rs *RenderState) Colors() (*RenderStateColors, error) {
	cc := C.init_render_state_colors()
	if err := resultError(C.ghostty_render_state_get(
		rs.ptr,
		C.GHOSTTY_RENDER_STATE_DATA_COLORS,
		unsafe.Pointer(&cc),
	)); err != nil {
		return nil, err
	}

	result := &RenderStateColors{
		Background:     ColorRGB{R: uint8(cc.background.r), G: uint8(cc.background.g), B: uint8(cc.background.b)},
		Foreground:     ColorRGB{R: uint8(cc.foreground.r), G: uint8(cc.foreground.g), B: uint8(cc.foreground.b)},
		Cursor:         ColorRGB{R: uint8(cc.cursor.r), G: uint8(cc.cursor.g), B: uint8(cc.cursor.b)},
		CursorHasValue: bool(cc.cursor_has_value),
	}

	for i, c := range cc.palette {
		result.Palette[i] = ColorRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}
	}
	return result, nil
}

// Cursor returns all cursor information from the render state in one call.
// This is the preferred cursor read for render loops because it avoids the
// separate cgo transition required by each scalar cursor getter.
func (rs *RenderState) Cursor() (*RenderStateCursor, error) {
	cc := C.init_render_state_cursor()
	if err := resultError(C.ghostty_render_state_get(
		rs.ptr,
		C.GHOSTTY_RENDER_STATE_DATA_CURSOR,
		unsafe.Pointer(&cc),
	)); err != nil {
		return nil, err
	}

	return &RenderStateCursor{
		ViewportHasValue: bool(cc.viewport_has_value),
		ViewportX:        uint16(cc.viewport_x),
		ViewportY:        uint16(cc.viewport_y),
		WideTail:         bool(cc.wide_tail),
		Visible:          bool(cc.visible),
		Blinking:         bool(cc.blinking),
		PasswordInput:    bool(cc.password_input),
		VisualStyle:      CursorVisualStyle(cc.visual_style),
	}, nil
}

// GetMulti queries multiple render state data fields in a single cgo
// call. This is a low-level function. Prefer the typed getters (Cols,
// Rows, CursorVisible, etc.) for normal use. GetMulti is useful when
// you need many fields at once and want to avoid per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateData enum).
// Use a Go type with the same size as the C type, such as uint32
// for uint32_t, bool for bool, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths
// or if a key cannot be read. In the second case, values for the keys
// before it may already have been written.
//
// Example:
//
//	var cols, rows uint16
//	err := rs.GetMulti(
//		[]RenderStateData{RenderStateDataCols, RenderStateDataRows},
//		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
//	)
//
// C: ghostty_render_state_get_multi
func (rs *RenderState) GetMulti(keys []RenderStateData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[C.GhosttyRenderStateData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	return resultError(C.ghostty_render_state_get_multi(
		rs.ptr,
		C.size_t(len(keys)),
		cKeys,
		args.values,
		nil,
	))
}

// CursorBlinking reports whether the cursor should blink based on
// terminal modes.
func (rs *RenderState) CursorBlinking() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_BLINKING, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorPasswordInput reports whether the cursor is at a password
// input field.
func (rs *RenderState) CursorPasswordInput() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_PASSWORD_INPUT, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorVisible reports whether the cursor is visible based on
// terminal modes.
func (rs *RenderState) CursorVisible() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISIBLE, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorVisualStyle returns the visual style of the cursor.
func (rs *RenderState) CursorVisualStyle() (CursorVisualStyle, error) {
	var v C.GhosttyRenderStateCursorVisualStyle
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VISUAL_STYLE, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return CursorVisualStyle(v), nil
}

// CursorViewportHasValue reports whether the cursor is visible within
// the viewport. If false, the cursor viewport position values are
// undefined.
func (rs *RenderState) CursorViewportHasValue() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_HAS_VALUE, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorViewportWideTail reports whether the cursor is on the tail
// of a wide character. Only valid when CursorViewportHasValue
// returns true.
func (rs *RenderState) CursorViewportWideTail() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_WIDE_TAIL, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// CursorViewportX returns the cursor viewport x position in cells.
// Only valid when CursorViewportHasValue returns true.
func (rs *RenderState) CursorViewportX() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_X, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// CursorViewportY returns the cursor viewport y position in cells.
// Only valid when CursorViewportHasValue returns true.
func (rs *RenderState) CursorViewportY() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_CURSOR_VIEWPORT_Y, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// Dirty returns the current dirty state.
func (rs *RenderState) Dirty() (RenderStateDirty, error) {
	var v C.GhosttyRenderStateDirty
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_DIRTY, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return RenderStateDirty(v), nil
}

// RowIterator initializes ri to iterate over the rows from the last update.
// Advance it with [RenderStateRowIterator.Next] before reading row data.
// Rows are ordered from top to bottom and include any overscan rows. Use
// [RenderStateRowIterator.ViewportY] to position each row in the viewport.
//
// The iterator can be reused by calling RowIterator again. Its data is valid
// only until the next call to [RenderState.Update].
func (rs *RenderState) RowIterator(ri *RenderStateRowIterator) error {
	return resultError(C.ghostty_render_state_get(
		rs.ptr,
		C.GHOSTTY_RENDER_STATE_DATA_ROW_ITERATOR,
		unsafe.Pointer(&ri.ptr),
	))
}

// Rows returns the viewport height in cells, excluding overscan.
func (rs *RenderState) Rows() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_render_state_get(rs.ptr, C.GHOSTTY_RENDER_STATE_DATA_ROWS, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// SetDirty sets the dirty state.
func (rs *RenderState) SetDirty(dirty RenderStateDirty) error {
	v := C.GhosttyRenderStateDirty(dirty)
	return resultError(C.ghostty_render_state_set(rs.ptr, C.GHOSTTY_RENDER_STATE_OPTION_DIRTY, unsafe.Pointer(&v)))
}

// SetOverscan sets the number of extra rows to capture above and below the
// viewport on subsequent updates. The request remains in effect until changed.
// A zero [RenderStateOverscan] captures only the visible viewport.
//
// Changing the request leaves the current row data valid. The next update
// applies the request and marks the entire render state as dirty. Use
// [RenderState.Overscan] after that update to find how many rows were available.
//
// C: GHOSTTY_RENDER_STATE_OPTION_OVERSCAN
func (rs *RenderState) SetOverscan(request RenderStateOverscan) error {
	v := C.GhosttyRenderStateOverscan{above: C.uint16_t(request.Above), below: C.uint16_t(request.Below)}
	return resultError(C.ghostty_render_state_set(rs.ptr, C.GHOSTTY_RENDER_STATE_OPTION_OVERSCAN, unsafe.Pointer(&v)))
}

// Overscan returns the number of extra rows captured by the last update.
// Each count is at most the corresponding count set by [RenderState.SetOverscan].
// Fewer rows are captured when the viewport is near the start or end of the
// terminal's contents. Below is zero when the viewport is scrolled to the bottom.
//
// [RenderState.RowIterator] includes these extra rows. Cursor coordinates still
// refer to the visible viewport and are available only when the cursor is
// inside it.
func (rs *RenderState) Overscan() (RenderStateOverscan, error) {
	return rs.getOverscan(C.GHOSTTY_RENDER_STATE_DATA_OVERSCAN)
}

// OverscanRequest returns the counts most recently set by [RenderState.SetOverscan].
// Both counts are zero if SetOverscan has not been called. The next update uses
// this request. Use [RenderState.Overscan] for the counts from the last update.
func (rs *RenderState) OverscanRequest() (RenderStateOverscan, error) {
	return rs.getOverscan(C.GHOSTTY_RENDER_STATE_DATA_OVERSCAN_REQUEST)
}

// getOverscan reads either the requested or captured overscan counts.
func (rs *RenderState) getOverscan(data C.GhosttyRenderStateData) (RenderStateOverscan, error) {
	var v C.GhosttyRenderStateOverscan
	if err := resultError(C.ghostty_render_state_get(rs.ptr, data, unsafe.Pointer(&v))); err != nil {
		return RenderStateOverscan{}, err
	}
	return RenderStateOverscan{Above: uint16(v.above), Below: uint16(v.below)}, nil
}
