package libghostty

// Render-state row iterator wrapping the
// GhosttyRenderStateRowIterator C APIs.

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyRenderStateRowSelection
// (sized struct).
static inline GhosttyRenderStateRowSelection init_render_state_row_selection() {
	GhosttyRenderStateRowSelection sel = GHOSTTY_INIT_SIZED(GhosttyRenderStateRowSelection);
	return sel;
}

// render_row_text_result describes the outcome of render_row_text_append. It
// is returned by value so Go never has to pass a pointer for C to fill in.
typedef struct {
	// result is GHOSTTY_SUCCESS when the rest of the row was written.
	// GHOSTTY_OUT_OF_SPACE means dst filled up first, and the caller should
	// grow dst and call again starting at next_x. Any other value is an
	// error from libghostty.
	GhosttyResult result;

	// len is the number of bytes written to dst.
	size_t len;

	// next_x is the column to start from on the next call, set only with
	// GHOSTTY_OUT_OF_SPACE. Every cell before it has been handled.
	uint16_t next_x;

	// needed is the minimum number of extra bytes dst must have for the next
	// call to make progress, set only with GHOSTTY_OUT_OF_SPACE.
	size_t needed;
} render_row_text_result;

// render_row_text_append writes the text of the current row to dst as UTF-8,
// starting at column start_x. It loops over the cells in C so that Go makes
// one call per row instead of two per cell. Each cell's text comes from the
// GRAPHEMES_UTF8 getter, so encoding matches the per-cell API exactly.
//
// The text rules are documented on the Go method, AppendText.
//
// Empty cells are held back as a pending run of spaces and written only when
// a cell with text follows them. This is what drops empty cells at the end
// of the row. If dst fills up while spaces are pending, next_x points at the
// first pending space rather than at the cell with text, so the next call
// writes those spaces again along with the text that follows them.
//
// The cells handle is reset to the current row. Its position afterwards is
// unspecified.
static inline render_row_text_result render_row_text_append(
	GhosttyRenderStateRowIterator rows,
	GhosttyRenderStateRowCells cells,
	uint16_t start_x,
	uint8_t* dst,
	size_t cap
) {
	render_row_text_result out = {
		.result = GHOSTTY_SUCCESS,
		.next_x = start_x,
	};

	// Load the current row into the cells handle. libghostty updates the
	// object the handle points to, so the local copy of the handle is enough.
	GhosttyResult result = ghostty_render_state_row_get(
		rows,
		GHOSTTY_RENDER_STATE_ROW_DATA_CELLS,
		&cells
	);
	if (result != GHOSTTY_SUCCESS) {
		out.result = result;
		return out;
	}

	ghostty_render_state_row_cells_select(cells, start_x);

	// blanks counts the pending empty cells, and blank_x is the column of
	// the first one.
	size_t blanks = 0;
	uint16_t blank_x = start_x;
	uint16_t x = start_x;
	do {
		// Leave room for the pending spaces and ask for the cell's text
		// right after them. If even the spaces don't fit, pass a NULL
		// buffer. An empty cell still succeeds with len=0, and a cell with
		// text fails with GHOSTTY_OUT_OF_SPACE and reports the size it needs.
		size_t room = cap - out.len;
		GhosttyBuffer buf = {
			.ptr = room > blanks ? dst + out.len + blanks : NULL,
			.cap = room > blanks ? room - blanks : 0,
			.len = 0,
		};
		result = ghostty_render_state_row_cells_get(
			cells,
			GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_UTF8,
			&buf
		);
		if (result == GHOSTTY_OUT_OF_SPACE) {
			out.result = result;
			out.next_x = blanks > 0 ? blank_x : x;
			out.needed = blanks + buf.len;
			return out;
		}
		if (result != GHOSTTY_SUCCESS) {
			out.result = result;
			return out;
		}

		// An empty cell adds to the pending run of spaces, unless it is a
		// spacer cell. A spacer cell is either the second half of a wide
		// character or the leftover cell where a wide character didn't fit
		// at the end of a row, so it produces no text.
		if (buf.len == 0) {
			GhosttyCell raw;
			GhosttyCellWide wide;
			ghostty_render_state_row_cells_get(
				cells,
				GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW,
				&raw
			);
			ghostty_cell_get(raw, GHOSTTY_CELL_DATA_WIDE, &wide);
			if (wide == GHOSTTY_CELL_WIDE_SPACER_TAIL ||
				wide == GHOSTTY_CELL_WIDE_SPACER_HEAD) {
				continue;
			}

			if (blanks == 0) blank_x = x;
			blanks++;
			continue;
		}

		// A cell with text was written after the reserved room, so fill
		// that room with the pending spaces.
		for (size_t i = 0; i < blanks; i++) dst[out.len + i] = ' ';
		out.len += blanks + buf.len;
		blanks = 0;
	} while (x++, ghostty_render_state_row_cells_next(cells));

	return out;
}
*/
import "C"

import (
	"errors"
	"slices"
	"unsafe"
)

// RenderStateRowData identifies a data field for render state row queries.
// C: GhosttyRenderStateRowData
type RenderStateRowData int

const (
	// RenderStateRowDataInvalid is an invalid / sentinel value.
	RenderStateRowDataInvalid RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_INVALID

	// RenderStateRowDataDirty indicates whether the current row is dirty
	// (bool).
	RenderStateRowDataDirty RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_DIRTY

	// RenderStateRowDataRaw is the raw row value (GhosttyRow).
	RenderStateRowDataRaw RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_RAW

	// RenderStateRowDataCells populates a pre-allocated row cells instance
	// (GhosttyRenderStateRowCells).
	RenderStateRowDataCells RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS

	// RenderStateRowDataSelection is the row-local selected cell range
	// (GhosttyRenderStateRowSelection).
	RenderStateRowDataSelection RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_SELECTION

	// RenderStateRowDataCellsRaw is a borrowed view of all packed cell values
	// in the current row (GhosttyCellsView).
	RenderStateRowDataCellsRaw RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS_RAW

	// RenderStateRowDataViewportY is the signed viewport-relative row position (int32_t).
	RenderStateRowDataViewportY RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_VIEWPORT_Y

	// RenderStateRowDataID is the stable row identity (GhosttyRenderStateRowId).
	RenderStateRowDataID RenderStateRowData = C.GHOSTTY_RENDER_STATE_ROW_DATA_ID
)

// RenderStateRowID identifies a row across render state updates, even when
// scrolling moves the row to a different position. It supports equality
// comparisons and can be used as a map key. The zero value is not a valid ID.
//
// A renderer can use an ID to cache work for a row. Reuse a cached result only
// if the ID appears in the new update and [RenderStateRowIterator.Dirty] is
// false. Discard entries for IDs that no longer appear. IDs are never reused
// for a different row.
//
// An ID can disappear when its row leaves the captured area, is removed from
// scrollback, or is replaced by a terminal operation.
//
// C: GhosttyRenderStateRowId
type RenderStateRowID struct {
	bits [2]uint64
}

// RenderStateRowSelection is the row-local selected cell range.
// C: GhosttyRenderStateRowSelection
type RenderStateRowSelection struct {
	// StartX is the start column of the row-local selection range,
	// inclusive.
	StartX uint16

	// EndX is the end column of the row-local selection range, inclusive.
	EndX uint16
}

// RenderStateRowIterator iterates over rows in a render state.
// Create one with NewRenderStateRowIterator, populate it via
// [RenderState.RowIterator], then advance with [RenderStateRowIterator.Next]
// and read data with getter methods.
//
// The iterator's current position is only valid as long as the
// underlying render state is not updated. Do not use the iterator
// while [RenderState.Update] may run on the same render state.
// Extracted [Row] values are copied snapshots and may be retained after
// the iterator itself becomes invalid.
//
// C: GhosttyRenderStateRowIterator
type RenderStateRowIterator struct {
	ptr C.GhosttyRenderStateRowIterator
}

// NewRenderStateRowIterator creates a new row iterator instance.
// The iterator is empty until populated via RenderState.RowIterator.
func NewRenderStateRowIterator() (*RenderStateRowIterator, error) {
	var ptr C.GhosttyRenderStateRowIterator
	if err := resultError(C.ghostty_render_state_row_iterator_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &RenderStateRowIterator{ptr: ptr}, nil
}

// Close frees the underlying row iterator handle. After this call,
// the iterator must not be used, except that calling Close again is a
// safe no-op.
func (ri *RenderStateRowIterator) Close() {
	if ri == nil || ri.ptr == nil {
		return
	}
	C.ghostty_render_state_row_iterator_free(ri.ptr)
	ri.ptr = nil
}

// Next advances to the next row and reports whether row data is available.
// It returns false when no rows remain. Rows are visited from top to bottom,
// including any overscan rows. Use [RenderStateRowIterator.ViewportY] to
// position the current row relative to the viewport.
func (ri *RenderStateRowIterator) Next() bool {
	return bool(C.ghostty_render_state_row_iterator_next(ri.ptr))
}

// NextDirty advances to the next row that requires a redraw. It returns the
// row's index and true, or zero and false when no such rows remain.
//
// The index counts from the first captured row, including overscan. Without
// overscan, it equals the row's position in the viewport. With overscan, use
// [RenderStateRowIterator.ViewportY] to position the row instead.
//
// When the entire render state is dirty, NextDirty visits every remaining row.
// Otherwise it skips clean rows. It does not clear any dirty flags.
//
// C: ghostty_render_state_row_iterator_next_dirty
func (ri *RenderStateRowIterator) NextDirty() (index uint16, ok bool) {
	var y C.uint16_t
	if !bool(C.ghostty_render_state_row_iterator_next_dirty(ri.ptr, &y)) {
		return 0, false
	}
	return uint16(y), true
}

// GetMulti queries multiple render-state row data fields in a single
// cgo call. This is a low-level function; prefer the typed getters
// (Dirty, Raw, Cells) for normal use. GetMulti is useful when you
// need many fields at once and want to avoid per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/render.h, GhosttyRenderStateRowData enum).
//
// Example:
//
//	var dirty C.bool
//	var raw C.GhosttyRow
//	err := ri.GetMulti(
//		[]RenderStateRowData{RenderStateRowDataDirty, RenderStateRowDataRaw},
//		[]unsafe.Pointer{unsafe.Pointer(&dirty), unsafe.Pointer(&raw)},
//	)
//
// C: ghostty_render_state_row_get_multi
func (ri *RenderStateRowIterator) GetMulti(keys []RenderStateRowData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Allocate the void** array in C memory to satisfy cgo pointer-passing rules.
	cVals, cValsSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cVals), cValsSize)
	return resultError(C.ghostty_render_state_row_get_multi(
		ri.ptr,
		C.size_t(len(keys)),
		(*C.GhosttyRenderStateRowData)(unsafe.Pointer(&keys[0])),
		cVals,
		nil,
	))
}

// Dirty reports whether the current row is dirty and requires a
// redraw.
func (ri *RenderStateRowIterator) Dirty() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_row_get(ri.ptr, C.GHOSTTY_RENDER_STATE_ROW_DATA_DIRTY, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// ViewportY returns the current row's position in rows relative to the top of
// the viewport. The first visible row is zero. Overscan rows above the viewport
// have negative positions. Those below it start at the height returned by
// [RenderState.Rows].
func (ri *RenderStateRowIterator) ViewportY() (int32, error) {
	var v C.int32_t
	if err := resultError(C.ghostty_render_state_row_get(ri.ptr, C.GHOSTTY_RENDER_STATE_ROW_DATA_VIEWPORT_Y, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return int32(v), nil
}

// ID returns the current row's identity. The value can be kept after the
// iterator advances or the render state is updated. See [RenderStateRowID]
// for how to use IDs when caching rendered rows.
func (ri *RenderStateRowIterator) ID() (RenderStateRowID, error) {
	var v C.GhosttyRenderStateRowId
	if err := resultError(C.ghostty_render_state_row_get(ri.ptr, C.GHOSTTY_RENDER_STATE_ROW_DATA_ID, unsafe.Pointer(&v))); err != nil {
		return RenderStateRowID{}, err
	}
	return RenderStateRowID{bits: [2]uint64{uint64(v.bits[0]), uint64(v.bits[1])}}, nil
}

// SetDirty sets the dirty state for the current row.
func (ri *RenderStateRowIterator) SetDirty(dirty bool) error {
	v := C.bool(dirty)
	return resultError(C.ghostty_render_state_row_set(ri.ptr, C.GHOSTTY_RENDER_STATE_ROW_OPTION_DIRTY, unsafe.Pointer(&v)))
}

// Raw returns the raw Row value for the current iterator position.
// The returned Row can be used with the same getter methods as rows
// obtained from GridRef.
func (ri *RenderStateRowIterator) Raw() (*Row, error) {
	var v C.GhosttyRow
	if err := resultError(C.ghostty_render_state_row_get(ri.ptr, C.GHOSTTY_RENDER_STATE_ROW_DATA_RAW, unsafe.Pointer(&v))); err != nil {
		return nil, err
	}
	return &Row{c: v}, nil
}

// Selection returns the row-local selected cell range for the current row. It
// returns nil when the current row does not intersect the terminal's active
// selection at the time the render state was updated.
func (ri *RenderStateRowIterator) Selection() (*RenderStateRowSelection, error) {
	v := C.init_render_state_row_selection()
	err := resultError(C.ghostty_render_state_row_get(
		ri.ptr,
		C.GHOSTTY_RENDER_STATE_ROW_DATA_SELECTION,
		unsafe.Pointer(&v),
	))
	if err != nil {
		var ge *Error
		if errors.As(err, &ge) && ge.Result == ResultNoValue {
			return nil, nil
		}
		return nil, err
	}

	return &RenderStateRowSelection{
		StartX: uint16(v.start_x),
		EndX:   uint16(v.end_x),
	}, nil
}

// Cells populates a pre-allocated row cells instance with cell data
// for the current row. The cells instance can then be advanced with
// Next or positioned with Select.
//
// The cells instance can be reused across rows. Cell data is only
// valid until the next call to RenderState.Update.
func (ri *RenderStateRowIterator) Cells(rc *RenderStateRowCells) error {
	return resultError(C.ghostty_render_state_row_get(
		ri.ptr,
		C.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS,
		unsafe.Pointer(&rc.ptr),
	))
}

// AppendText appends the text of the current row to dst as UTF-8 and returns
// the extended slice. Any existing content in dst is kept.
//
// The text follows what the row shows on screen:
//
//   - Empty cells between characters become spaces.
//   - Empty cells at the end of the row are left out. Spaces the program
//     actually printed are text and are kept, even at the end of the row.
//   - A wide character, such as a CJK character or most emoji, appears once
//     even though it covers two cells.
//   - A character made of several code points, such as an emoji with a skin
//     tone modifier, is kept whole.
//
// No newline is added. When the terminal wraps a long line onto several
// rows, each row is returned separately. To join them, check [Row.Wrap] on
// the value returned by [RenderStateRowIterator.Raw].
//
// rc is working storage. AppendText loads the current row into it and leaves
// it at an unspecified cell. As with [RenderStateRowIterator.Cells], one rc
// can be reused for every row.
//
// AppendText is much faster than calling [RenderStateRowCells.AppendGraphemes]
// on each cell, because it reads the whole row in one call into libghostty.
// To avoid allocations, reuse one buffer across rows by passing buf[:0].
func (ri *RenderStateRowIterator) AppendText(dst []byte, rc *RenderStateRowCells) ([]byte, error) {
	var x C.uint16_t
	for {
		oldLen := len(dst)
		available := cap(dst) - oldLen
		var ptr *C.uint8_t
		if available > 0 {
			buf := dst[:cap(dst)]
			ptr = (*C.uint8_t)(unsafe.Pointer(&buf[oldLen]))
		}

		result := C.render_row_text_append(ri.ptr, rc.ptr, x, ptr, C.size_t(available))
		dst = dst[:oldLen+int(result.len)]
		switch result.result {
		case C.GHOSTTY_SUCCESS:
			return dst, nil

		case C.GHOSTTY_OUT_OF_SPACE:
			// Grow dst so the next cell fits and continue from where the C
			// loop stopped.
			x = result.next_x
			dst = slices.Grow(dst, int(result.needed))

		default:
			return dst, resultError(result.result)
		}
	}
}

// CellsRaw returns a borrowed, contiguous view of the packed cell values in
// the current row. It avoids one cgo call per cell and is intended for render
// paths that decode [Cell.PackedValue] using the layout from [TypeJSON].
//
// The view is invalidated by the next [RenderState.Update]. Call
// [CellsView.Clone] when the values must outlive the current render state
// snapshot.
func (ri *RenderStateRowIterator) CellsRaw() (CellsView, error) {
	var view C.GhosttyCellsView
	if err := resultError(C.ghostty_render_state_row_get(
		ri.ptr,
		C.GHOSTTY_RENDER_STATE_ROW_DATA_CELLS_RAW,
		unsafe.Pointer(&view),
	)); err != nil {
		return CellsView{}, err
	}
	return cellsViewFromC(view), nil
}
