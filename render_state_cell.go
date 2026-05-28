package libghostty

// Render-state row cell iterator wrapping the
// GhosttyRenderStateRowCells C APIs.

/*
#include <ghostty/vt.h>

// render_cell_style_snapshot is a small binding-side adapter over the current
// libghostty render-cell APIs. The upstream API exposes each style/color field
// separately, which is flexible but expensive for renderers because each Go
// getter is a cgo transition and the allocation-returning Go getters materialize
// *Style / *ColorRGB values. This helper keeps the same libghostty behavior but
// batches the fields into one cgo call and a caller-owned result struct.
typedef struct {
	bool has_styling;
	bool has_foreground;
	GhosttyColorRgb foreground;
	bool has_background;
	GhosttyColorRgb background;
	bool bold;
	bool faint;
	bool italic;
	bool underline;
	bool strikethrough;
	bool inverse;
} render_cell_style_snapshot;

typedef struct {
	GhosttyResult result;
	render_cell_style_snapshot style;
} render_cell_style_snapshot_result;

static inline render_cell_style_snapshot_result render_cell_style_snapshot_get(
	GhosttyRenderStateRowCells cells
) {
	render_cell_style_snapshot_result out = {
		.result = GHOSTTY_SUCCESS,
		.style = {0},
	};

	GhosttyResult result = ghostty_render_state_row_cells_get(
		cells,
		GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING,
		&out.style.has_styling
	);
	if (result != GHOSTTY_SUCCESS) {
		out.result = result;
		return out;
	}

	if (out.style.has_styling) {
		GhosttyStyle style = GHOSTTY_INIT_SIZED(GhosttyStyle);
		result = ghostty_render_state_row_cells_get(
			cells,
			GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE,
			&style
		);
		if (result != GHOSTTY_SUCCESS) {
			out.result = result;
			return out;
		}

		out.style.bold = style.bold;
		out.style.faint = style.faint;
		out.style.italic = style.italic;
		out.style.underline = style.underline != GHOSTTY_SGR_UNDERLINE_NONE;
		out.style.strikethrough = style.strikethrough;
		out.style.inverse = style.inverse;
	}

	result = ghostty_render_state_row_cells_get(
		cells,
		GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_FG_COLOR,
		&out.style.foreground
	);
	if (result == GHOSTTY_SUCCESS) {
		out.style.has_foreground = true;
	} else if (result != GHOSTTY_INVALID_VALUE) {
		out.result = result;
		return out;
	}

	result = ghostty_render_state_row_cells_get(
		cells,
		GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_BG_COLOR,
		&out.style.background
	);
	if (result == GHOSTTY_SUCCESS) {
		out.style.has_background = true;
	} else if (result != GHOSTTY_INVALID_VALUE) {
		out.result = result;
		return out;
	}

	return out;
}

// render_cell_graphemes_utf8_result is returned by value so the Go binding
// doesn't need to pass Go out-parameters through cgo for the hot per-cell text
// path. The current upstream API exposes grapheme extraction as a length query
// plus a codepoint-buffer query; using that directly from Go forces the length
// and any small Go scratch array to escape to the heap. Keeping both the length
// and the common small codepoint scratch in C avoids those per-cell heap
// allocations while preserving the upstream API's behavior.
typedef struct {
	GhosttyResult result;
	size_t written;
	size_t needed;
} render_cell_graphemes_utf8_result;

static inline uint32_t render_cell_utf8_valid(uint32_t cp) {
	if (cp > 0x10FFFF || (cp >= 0xD800 && cp <= 0xDFFF)) {
		return 0xFFFD;
	}
	return cp;
}

static inline size_t render_cell_utf8_len(uint32_t cp) {
	cp = render_cell_utf8_valid(cp);
	if (cp < 0x80) return 1;
	if (cp < 0x800) return 2;
	if (cp < 0x10000) return 3;
	return 4;
}

static inline size_t render_cell_utf8_write(uint8_t* dst, uint32_t cp) {
	cp = render_cell_utf8_valid(cp);
	if (cp < 0x80) {
		dst[0] = (uint8_t)cp;
		return 1;
	}
	if (cp < 0x800) {
		dst[0] = (uint8_t)(0xC0 | (cp >> 6));
		dst[1] = (uint8_t)(0x80 | (cp & 0x3F));
		return 2;
	}
	if (cp < 0x10000) {
		dst[0] = (uint8_t)(0xE0 | (cp >> 12));
		dst[1] = (uint8_t)(0x80 | ((cp >> 6) & 0x3F));
		dst[2] = (uint8_t)(0x80 | (cp & 0x3F));
		return 3;
	}
	dst[0] = (uint8_t)(0xF0 | (cp >> 18));
	dst[1] = (uint8_t)(0x80 | ((cp >> 12) & 0x3F));
	dst[2] = (uint8_t)(0x80 | ((cp >> 6) & 0x3F));
	dst[3] = (uint8_t)(0x80 | (cp & 0x3F));
	return 4;
}

static inline render_cell_graphemes_utf8_result render_cell_graphemes_utf8_append(
	GhosttyRenderStateRowCells cells,
	uint8_t* dst,
	size_t dst_len
) {
	render_cell_graphemes_utf8_result out = {
		.result = GHOSTTY_SUCCESS,
		.written = 0,
		.needed = 0,
	};

	uint32_t n = 0;
	GhosttyResult result = ghostty_render_state_row_cells_get(
		cells,
		GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_LEN,
		&n
	);
	if (result != GHOSTTY_SUCCESS || n == 0) {
		out.result = result;
		return out;
	}

	uint32_t small[8];
	uint32_t* graphemes = small;
	size_t graphemes_size = sizeof(uint32_t) * (size_t)n;
	if (n > 8) {
		graphemes = (uint32_t*)ghostty_alloc(NULL, graphemes_size);
		if (graphemes == NULL) {
			out.result = GHOSTTY_OUT_OF_MEMORY;
			return out;
		}
	}

	result = ghostty_render_state_row_cells_get(
		cells,
		GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF,
		graphemes
	);
	if (result != GHOSTTY_SUCCESS) {
		if (graphemes != small) ghostty_free(NULL, (uint8_t*)graphemes, graphemes_size);
		out.result = result;
		return out;
	}

	for (uint32_t i = 0; i < n; i++) {
		out.needed += render_cell_utf8_len(graphemes[i]);
	}
	if (out.needed > dst_len) {
		if (graphemes != small) ghostty_free(NULL, (uint8_t*)graphemes, graphemes_size);
		out.result = GHOSTTY_OUT_OF_SPACE;
		return out;
	}

	for (uint32_t i = 0; i < n; i++) {
		out.written += render_cell_utf8_write(dst + out.written, graphemes[i]);
	}

	if (graphemes != small) ghostty_free(NULL, (uint8_t*)graphemes, graphemes_size);
	return out;
}
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

	// RenderStateRowCellsDataSelected indicates whether the cell is contained
	// within the current selection (bool).
	RenderStateRowCellsDataSelected RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_SELECTED

	// RenderStateRowCellsDataHasStyling indicates whether the cell has any
	// explicit styling (bool).
	RenderStateRowCellsDataHasStyling RenderStateRowCellsData = C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING
)

// RenderCellStyle is a reusable, resolved style snapshot for the current
// render-state cell. It corresponds to the data returned by
// ghostty_render_state_row_cells_get for style, foreground color, background
// color, and has-styling, flattened into one Go value for hot render paths.
type RenderCellStyle struct {
	// Foreground is the resolved foreground color when HasForeground is true.
	Foreground ColorRGB

	// Background is the resolved background color when HasBackground is true.
	Background ColorRGB

	// HasForeground reports whether Foreground contains a resolved foreground
	// color. When false, the renderer should use its default foreground.
	HasForeground bool

	// HasBackground reports whether Background contains a resolved background
	// color. When false, the renderer should use its default background.
	HasBackground bool

	// HasStyling reports whether the cell has any explicit non-default styling.
	HasStyling bool

	// Bold reports whether bold text is set.
	Bold bool

	// Faint reports whether faint/dim text is set.
	Faint bool

	// Italic reports whether italic text is set.
	Italic bool

	// Underline reports whether any underline style is set.
	Underline bool

	// Strikethrough reports whether strikethrough text is set.
	Strikethrough bool

	// Inverse reports whether inverse video is set.
	Inverse bool
}

// RenderStateRowCells iterates over cells in a render-state row.
// Create one with NewRenderStateRowCells, populate it via
// [RenderStateRowIterator.Cells], then advance with
// [RenderStateRowCells.Next] (or jump with [RenderStateRowCells.Select])
// and read data with getter methods.
//
// A single instance can be reused across rows to avoid repeated
// allocation. The cells view is only valid until the next call to
// [RenderState.Update]. Do not use it while [RenderState.Update] may
// run on the same render state. Extracted [Cell], [Style], color
// values, and grapheme slices are copied values and may be retained.
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
	cVals, cValsSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cVals), cValsSize)
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

// StyleInto fills dst with the reusable resolved style snapshot for the current
// cell. This is the allocation-reusing form of querying Style, FgColor,
// BgColor, and HasStyling separately: it performs one cgo transition, stores
// colors as values, and sets HasForeground/HasBackground when the corresponding
// color is present.
func (rc *RenderStateRowCells) StyleInto(dst *RenderCellStyle) error {
	if dst == nil {
		return errors.New("libghostty: nil RenderCellStyle")
	}

	result := C.render_cell_style_snapshot_get(rc.ptr)
	if err := resultError(result.result); err != nil {
		return err
	}

	snap := result.style
	dst.HasStyling = bool(snap.has_styling)
	dst.HasForeground = bool(snap.has_foreground)
	dst.HasBackground = bool(snap.has_background)
	dst.Foreground = ColorRGB{R: uint8(snap.foreground.r), G: uint8(snap.foreground.g), B: uint8(snap.foreground.b)}
	dst.Background = ColorRGB{R: uint8(snap.background.r), G: uint8(snap.background.g), B: uint8(snap.background.b)}
	dst.Bold = bool(snap.bold)
	dst.Faint = bool(snap.faint)
	dst.Italic = bool(snap.italic)
	dst.Underline = bool(snap.underline)
	dst.Strikethrough = bool(snap.strikethrough)
	dst.Inverse = bool(snap.inverse)

	return nil
}

// Graphemes returns the full grapheme cluster codepoints for the
// current cell. The base codepoint is first, followed by any extra
// codepoints. Returns nil if the cell has no text.
func (rc *RenderStateRowCells) Graphemes() ([]uint32, error) {
	graphemes, err := rc.GraphemesInto(nil)
	if err != nil {
		return nil, err
	}
	if len(graphemes) == 0 {
		return nil, nil
	}
	return graphemes, nil
}

// GraphemesInto appends the full grapheme cluster codepoints for the
// current cell to dst and returns the extended slice. The base codepoint
// is appended first, followed by any extra codepoints. If the cell has no
// text, dst is returned unchanged.
//
// This is the allocation-reusing form of [RenderStateRowCells.Graphemes].
// Callers that want a per-cell scratch buffer should pass scratch[:0].
func (rc *RenderStateRowCells) GraphemesInto(dst []uint32) ([]uint32, error) {
	// Get the number of codepoints.
	var n C.uint32_t
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_LEN, unsafe.Pointer(&n))); err != nil {
		return dst, err
	}
	if n == 0 {
		return dst, nil
	}

	// Read codepoints into the newly appended portion of the caller's buffer.
	oldLen := len(dst)
	newLen := oldLen + int(n)
	if newLen <= cap(dst) {
		dst = dst[:newLen]
	} else {
		grown := make([]uint32, newLen)
		copy(grown, dst)
		dst = grown
	}
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF, unsafe.Pointer(&dst[oldLen]))); err != nil {
		return dst[:oldLen], err
	}
	return dst, nil
}

// AppendGraphemes appends the current cell's grapheme cluster encoded as
// UTF-8 to dst and returns the extended byte slice. If the cell has no text,
// dst is returned unchanged.
//
// This is intended for renderers and text extractors that ultimately need
// bytes or strings and want to avoid allocating a temporary codepoint slice
// and per-cell string. The common case of a short grapheme cluster uses C
// stack storage for codepoints, avoiding cgo-induced escapes of Go scratch
// values; unusually long clusters allocate temporary C memory for that call.
func (rc *RenderStateRowCells) AppendGraphemes(dst []byte) ([]byte, error) {
	oldLen := len(dst)
	for {
		available := cap(dst) - oldLen
		var ptr *C.uint8_t
		if available > 0 {
			buf := dst[:cap(dst)]
			ptr = (*C.uint8_t)(unsafe.Pointer(&buf[oldLen]))
		}

		result := C.render_cell_graphemes_utf8_append(rc.ptr, ptr, C.size_t(available))
		switch result.result {
		case C.GHOSTTY_SUCCESS:
			return dst[:oldLen+int(result.written)], nil
		case C.GHOSTTY_OUT_OF_SPACE:
			needed := int(result.needed)
			if needed <= available {
				return dst, resultError(result.result)
			}
			required := oldLen + needed
			newCap := cap(dst) * 2
			if newCap < 64 {
				newCap = 64
			}
			if newCap < required {
				newCap = required
			}
			grown := make([]byte, oldLen, newCap)
			copy(grown, dst[:oldLen])
			dst = grown
		default:
			return dst, resultError(result.result)
		}
	}
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

// Selected reports whether the current cell is contained within the
// terminal selection snapshot that was current when the render state
// was updated. Rendering policy for selected cells is left to the
// caller.
func (rc *RenderStateRowCells) Selected() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_SELECTED, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// HasStyling reports whether the current cell has any explicit styling.
// This is equivalent to querying the raw Cell's HasStyling value, but
// avoids materializing the raw cell when renderers only need to know
// whether fetching the full style is necessary.
func (rc *RenderStateRowCells) HasStyling() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_render_state_row_cells_get(rc.ptr, C.GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_HAS_STYLING, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}
