package libghostty

// Render-state row cell iterator wrapping the
// GhosttyRenderStateRowCells C APIs.

/*
#include <stdlib.h>
#include <ghostty/vt.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// RenderStateRowCellsData identifies a data field for render state cell
// queries.
// C: GhosttyRenderStateRowCellsData
type RenderStateRowCellsData int

const (
	// RenderStateRowCellsDataInvalid is an invalid / sentinel value.
	RenderStateRowCellsDataInvalid RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_INVALID

	// RenderStateRowCellsDataRaw is the raw cell value (GhosttyCell).
	RenderStateRowCellsDataRaw RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW

	// RenderStateRowCellsDataStyle is the style for the current cell
	// (GhosttyStyle).
	RenderStateRowCellsDataStyle RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE

	// RenderStateRowCellsDataGraphemesLen is the total number of grapheme
	// codepoints including the base codepoint (uint32_t).
	RenderStateRowCellsDataGraphemesLen RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_LEN

	// RenderStateRowCellsDataGraphemesBuf writes grapheme codepoints into
	// a caller-provided buffer (uint32_t*).
	RenderStateRowCellsDataGraphemesBuf RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF

	// RenderStateRowCellsDataBgColor is the resolved background color of
	// the cell (GhosttyColorRgb).
	RenderStateRowCellsDataBgColor RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR

	// RenderStateRowCellsDataFgColor is the resolved foreground color of
	// the cell (GhosttyColorRgb).
	RenderStateRowCellsDataFgColor RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR
)

// RenderStateRowCells iterates over cells in a render-state row.
// Create one with NewRenderStateRowCells, populate it via
// RenderStateRowIterator.Cells, then advance with Next (or jump
// with Select) and read data with getter methods.
//
// A single instance can be reused across rows to avoid repeated
// allocation. Cell data is only valid until the next call to
// RenderState.Update.
//
// C: GhosttyRenderStateRowCells
type RenderStateRowCells struct {
	ptr C.GhosttyRenderStateRowCells
}

// NewRenderStateRowCells creates a new row cells instance. The
// instance is empty until populated via RenderStateRowIterator.Cells.
func NewRenderStateRowCells() (*RenderStateRowCells, error) {
	var ptr C.GhosttyRenderStateRowCells
	if err := resultError(C.ghostty_render_state_row_cells_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &RenderStateRowCells{ptr: ptr}, nil
}

// Close frees the underlying row cells handle. After this call,
// the instance must not be used.
func (rc *RenderStateRowCells) Close() {
	C.ghostty_render_state_row_cells_free(rc.ptr)
}

// Next advances the iterator to the next cell. Returns true if the
// iterator moved successfully and cell data is available. Returns
// false when there are no more cells.
func (rc *RenderStateRowCells) Next() bool {
	return bool(C.ghostty_render_state_row_cells_next(rc.ptr))
}

// Select positions the iterator at the given column index (0-based)
// so that subsequent reads return data for that cell.
func (rc *RenderStateRowCells) Select(x uint16) error {
	return resultError(C.ghostty_render_state_row_cells_select(rc.ptr, C.uint16_t(x)))
}

// GetMulti queries multiple render-state cell data fields in a single
// cgo call. This is a low-level function; prefer the typed getters
// (Raw, Style, Graphemes, BgColor, FgColor) for normal use. GetMulti
// is useful when you need many fields at once and want to avoid
// per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateRowCellsData enum).
//
// Example:
//
//	var raw C.GhosttyCell
//	var graphemesLen C.uint32_t
//	err := rc.GetMulti(
//		[]RenderStateRowCellsData{RenderStateRowCellsDataRaw, RenderStateRowCellsDataGraphemesLen},
//		[]unsafe.Pointer{unsafe.Pointer(&raw), unsafe.Pointer(&graphemesLen)},
//	)
//
// C: ghostty_render_state_row_cells_get_multi
func (rc *RenderStateRowCells) GetMulti(keys []RenderStateRowCellsData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Allocate the void** array in C memory to satisfy cgo pointer-passing rules.
	cVals := cValuesArray(values)
	defer C.free(unsafe.Pointer(cVals))
	return resultError(C.ghostty_render_state_row_cells_get_multi(
		rc.ptr,
		C.size_t(len(keys)),
		(*C.GhosttyRenderStateRowCellsData)(unsafe.Pointer(&keys[0])),
		cVals,
		nil,
	))
}

// Raw returns the raw Cell value for the current iterator position.
// The returned Cell can be used with the same getter methods as cells
// obtained from GridRef.
func (rc *RenderStateRowCells) Raw() (*Cell, error) {
	var v C.GhosttyCell
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW, unsafe.Pointer(&v))); err != nil {
		return nil, err
	}
	return &Cell{c: v}, nil
}

// Style returns the style for the current cell.
func (rc *RenderStateRowCells) Style() (*Style, error) {
	cs := initCStyle()
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE, unsafe.Pointer(&cs))); err != nil {
		return nil, err
	}
	return &Style{c: cs}, nil
}

// Graphemes returns the full grapheme cluster codepoints for the
// current cell. The base codepoint is first, followed by any extra
// codepoints. Returns nil if the cell has no text.
func (rc *RenderStateRowCells) Graphemes() ([]uint32, error) {
	// Get the number of codepoints.
	var n C.uint32_t
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_LEN, unsafe.Pointer(&n))); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	// Read codepoints into a buffer.
	buf := make([]uint32, uint32(n))
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF, unsafe.Pointer(&buf[0]))); err != nil {
		return nil, err
	}
	return buf, nil
}

// BgColor returns the resolved background color for the current cell.
// Returns nil (without error) when the cell has no background color,
// in which case the caller should use the terminal default background.
func (rc *RenderStateRowCells) BgColor() (*ColorRGB, error) {
	var v C.GhosttyColorRgb
	err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR, unsafe.Pointer(&v)))
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

// FgColor returns the resolved foreground color for the current cell.
// Returns nil (without error) when the cell has no explicit foreground
// color, in which case the caller should use the terminal default
// foreground. Bold color handling is not applied.
func (rc *RenderStateRowCells) FgColor() (*ColorRGB, error) {
	var v C.GhosttyColorRgb
	err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR, unsafe.Pointer(&v)))
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
