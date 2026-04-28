package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// CellData identifies a data field for cell queries.
// C: GhosttyCellData
type CellData int

const (
	// CellDataInvalid is an invalid data type.
	CellDataInvalid CellData = C.GHOSTTY_CELL_DATA_INVALID

	// CellDataCodepoint is the codepoint of the cell (uint32_t).
	CellDataCodepoint CellData = C.GHOSTTY_CELL_DATA_CODEPOINT

	// CellDataContentTag is the content tag describing what kind of
	// content is in the cell (GhosttyCellContentTag).
	CellDataContentTag CellData = C.GHOSTTY_CELL_DATA_CONTENT_TAG

	// CellDataWide is the wide property of the cell (GhosttyCellWide).
	CellDataWide CellData = C.GHOSTTY_CELL_DATA_WIDE

	// CellDataHasText indicates whether the cell has text to render (bool).
	CellDataHasText CellData = C.GHOSTTY_CELL_DATA_HAS_TEXT

	// CellDataHasStyling indicates whether the cell has non-default
	// styling (bool).
	CellDataHasStyling CellData = C.GHOSTTY_CELL_DATA_HAS_STYLING

	// CellDataStyleID is the style ID for the cell (uint16_t).
	CellDataStyleID CellData = C.GHOSTTY_CELL_DATA_STYLE_ID

	// CellDataHasHyperlink indicates whether the cell has a hyperlink
	// (bool).
	CellDataHasHyperlink CellData = C.GHOSTTY_CELL_DATA_HAS_HYPERLINK

	// CellDataProtected indicates whether the cell is protected (bool).
	CellDataProtected CellData = C.GHOSTTY_CELL_DATA_PROTECTED

	// CellDataSemanticContent is the semantic content type of the cell
	// (GhosttyCellSemanticContent).
	CellDataSemanticContent CellData = C.GHOSTTY_CELL_DATA_SEMANTIC_CONTENT

	// CellDataColorPalette is the palette index for the cell's background
	// color (GhosttyColorPaletteIndex).
	CellDataColorPalette CellData = C.GHOSTTY_CELL_DATA_COLOR_PALETTE

	// CellDataColorRGBValue is the RGB value for the cell's background
	// color (GhosttyColorRgb).
	CellDataColorRGBValue CellData = C.GHOSTTY_CELL_DATA_COLOR_RGB
)

// RowData identifies a data field for row queries.
// C: GhosttyRowData
type RowData int

const (
	// RowDataInvalid is an invalid data type.
	RowDataInvalid RowData = C.GHOSTTY_ROW_DATA_INVALID

	// RowDataWrap indicates whether the row is soft-wrapped (bool).
	RowDataWrap RowData = C.GHOSTTY_ROW_DATA_WRAP

	// RowDataWrapContinuation indicates whether the row is a continuation
	// of a soft-wrapped row (bool).
	RowDataWrapContinuation RowData = C.GHOSTTY_ROW_DATA_WRAP_CONTINUATION

	// RowDataGrapheme indicates whether any cells in the row have grapheme
	// clusters (bool).
	RowDataGrapheme RowData = C.GHOSTTY_ROW_DATA_GRAPHEME

	// RowDataStyled indicates whether any cells in the row have styling
	// (bool).
	RowDataStyled RowData = C.GHOSTTY_ROW_DATA_STYLED

	// RowDataHyperlink indicates whether any cells in the row have
	// hyperlinks (bool).
	RowDataHyperlink RowData = C.GHOSTTY_ROW_DATA_HYPERLINK

	// RowDataSemanticPrompt is the semantic prompt state of the row
	// (GhosttyRowSemanticPrompt).
	RowDataSemanticPrompt RowData = C.GHOSTTY_ROW_DATA_SEMANTIC_PROMPT

	// RowDataKittyVirtualPlaceholder indicates whether the row contains
	// a Kitty virtual placeholder (bool).
	RowDataKittyVirtualPlaceholder RowData = C.GHOSTTY_ROW_DATA_KITTY_VIRTUAL_PLACEHOLDER

	// RowDataDirty indicates whether the row is dirty and requires a
	// redraw (bool).
	RowDataDirty RowData = C.GHOSTTY_ROW_DATA_DIRTY
)

// Cell is a wrapper around an opaque terminal grid cell value.
// Use getter methods to extract data from it. A Cell is a copied value
// snapshot, not a borrowed handle, so it may be retained after the
// [GridRef] or render-state iterator that produced it becomes invalid.
// C: GhosttyCell
type Cell struct {
	c C.GhosttyCell
}

// Row is a wrapper around an opaque terminal grid row value.
// Use getter methods to extract data from it. A Row is a copied value
// snapshot, not a borrowed handle, so it may be retained after the
// [GridRef] or render-state iterator that produced it becomes invalid.
// C: GhosttyRow
type Row struct {
	c C.GhosttyRow
}

// CellContentTag describes what kind of content a cell holds.
// C: GhosttyCellContentTag
type CellContentTag int

const (
	// CellContentCodepoint means a single codepoint (may be zero for empty).
	CellContentCodepoint CellContentTag = C.GHOSTTY_CELL_CONTENT_CODEPOINT

	// CellContentCodepointGrapheme means a codepoint that is part of a
	// multi-codepoint grapheme cluster.
	CellContentCodepointGrapheme CellContentTag = C.GHOSTTY_CELL_CONTENT_CODEPOINT_GRAPHEME

	// CellContentBgColorPalette means no text; background color from palette.
	CellContentBgColorPalette CellContentTag = C.GHOSTTY_CELL_CONTENT_BG_COLOR_PALETTE

	// CellContentBgColorRGB means no text; background color as RGB.
	CellContentBgColorRGB CellContentTag = C.GHOSTTY_CELL_CONTENT_BG_COLOR_RGB
)

// CellWide describes the width behavior of a cell.
// C: GhosttyCellWide
type CellWide int

const (
	// CellWideNarrow means not a wide character, cell width 1.
	CellWideNarrow CellWide = C.GHOSTTY_CELL_WIDE_NARROW

	// CellWideWide means wide character, cell width 2.
	CellWideWide CellWide = C.GHOSTTY_CELL_WIDE_WIDE

	// CellWideSpacerTail means spacer after wide character (do not render).
	CellWideSpacerTail CellWide = C.GHOSTTY_CELL_WIDE_SPACER_TAIL

	// CellWideSpacerHead means spacer at end of soft-wrapped line for a
	// wide character.
	CellWideSpacerHead CellWide = C.GHOSTTY_CELL_WIDE_SPACER_HEAD
)

// CellSemanticContent is the semantic content type of a cell,
// as set by OSC 133 sequences.
// C: GhosttyCellSemanticContent
type CellSemanticContent int

const (
	// CellSemanticOutput means regular output content (e.g. command output).
	CellSemanticOutput CellSemanticContent = C.GHOSTTY_CELL_SEMANTIC_OUTPUT

	// CellSemanticInput means content that is part of user input.
	CellSemanticInput CellSemanticContent = C.GHOSTTY_CELL_SEMANTIC_INPUT

	// CellSemanticPrompt means content that is part of a shell prompt.
	CellSemanticPrompt CellSemanticContent = C.GHOSTTY_CELL_SEMANTIC_PROMPT
)

// RowSemanticPrompt indicates whether any cells in a row are part of
// a shell prompt, as reported by OSC 133 sequences.
// C: GhosttyRowSemanticPrompt
type RowSemanticPrompt int

const (
	// RowSemanticNone means no prompt cells in this row.
	RowSemanticNone RowSemanticPrompt = C.GHOSTTY_ROW_SEMANTIC_NONE

	// RowSemanticPromptPrimary means prompt cells exist and this is a
	// primary prompt line.
	RowSemanticPromptPrimary RowSemanticPrompt = C.GHOSTTY_ROW_SEMANTIC_PROMPT

	// RowSemanticPromptContinuation means prompt cells exist and this is
	// a continuation line.
	RowSemanticPromptContinuation RowSemanticPrompt = C.GHOSTTY_ROW_SEMANTIC_PROMPT_CONTINUATION
)

// GetMulti queries multiple cell data fields in a single cgo call.
// This is a low-level function; prefer the typed getters (Codepoint,
// Wide, HasText, etc.) for normal use. GetMulti is useful when you
// need many fields at once and want to avoid per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/screen.h, GhosttyCellData enum).
//
// Example:
//
//	var cp C.uint32_t
//	var wide C.GhosttyCellWide
//	err := cell.GetMulti(
//		[]CellData{CellDataCodepoint, CellDataWide},
//		[]unsafe.Pointer{unsafe.Pointer(&cp), unsafe.Pointer(&wide)},
//	)
//
// C: ghostty_cell_get_multi
func (c *Cell) GetMulti(keys []CellData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Allocate the void** array in C memory to satisfy cgo pointer-passing rules.
	cVals, cValsSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cVals), cValsSize)
	return resultError(C.ghostty_cell_get_multi(
		c.c,
		C.size_t(len(keys)),
		(*C.GhosttyCellData)(unsafe.Pointer(&keys[0])),
		cVals,
		nil,
	))
}

// Codepoint returns the codepoint of the cell (0 if empty).
func (c *Cell) Codepoint() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_CODEPOINT, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// ContentTag returns the content tag describing what kind of content
// the cell holds.
func (c *Cell) ContentTag() (CellContentTag, error) {
	var v C.GhosttyCellContentTag
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_CONTENT_TAG, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return CellContentTag(v), nil
}

// Wide returns the wide property of the cell.
func (c *Cell) Wide() (CellWide, error) {
	var v C.GhosttyCellWide
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_WIDE, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return CellWide(v), nil
}

// HasText reports whether the cell has text to render.
func (c *Cell) HasText() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_HAS_TEXT, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// HasStyling reports whether the cell has non-default styling.
func (c *Cell) HasStyling() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_HAS_STYLING, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// StyleID returns the style ID for the cell.
func (c *Cell) StyleID() (uint16, error) {
	var v C.uint16_t
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_STYLE_ID, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// HasHyperlink reports whether the cell has a hyperlink.
func (c *Cell) HasHyperlink() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_HAS_HYPERLINK, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Protected reports whether the cell is protected.
func (c *Cell) Protected() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_PROTECTED, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Semantic returns the semantic content type of the cell.
func (c *Cell) Semantic() (CellSemanticContent, error) {
	var v C.GhosttyCellSemanticContent
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_SEMANTIC_CONTENT, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return CellSemanticContent(v), nil
}

// ColorPalette returns the palette index for the cell's background
// color. Only valid when the cell's content tag is CellContentBgColorPalette.
func (c *Cell) ColorPalette() (uint8, error) {
	var v C.GhosttyColorPaletteIndex
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_COLOR_PALETTE, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return uint8(v), nil
}

// ColorRGB returns the RGB color for the cell's background color.
// Only valid when the cell's content tag is CellContentBgColorRGB.
func (c *Cell) ColorRGB() (ColorRGB, error) {
	var v C.GhosttyColorRgb
	if err := resultError(C.ghostty_cell_get(c.c, C.GHOSTTY_CELL_DATA_COLOR_RGB, unsafe.Pointer(&v))); err != nil {
		return ColorRGB{}, err
	}
	return ColorRGB{R: uint8(v.r), G: uint8(v.g), B: uint8(v.b)}, nil
}

// GetMulti queries multiple row data fields in a single cgo call.
// This is a low-level function; prefer the typed getters (Wrap,
// Grapheme, Styled, Semantic, etc.) for normal use. GetMulti is
// useful when you need many fields at once and want to avoid
// per-field cgo overhead.
//
// Each element in keys specifies a data kind, and the corresponding
// element in values must be an unsafe.Pointer to a variable whose type
// matches the "Output type" documented for that key in the upstream C
// header (ghostty/vt/screen.h, GhosttyRowData enum).
//
// Example:
//
//	var wrap, styled C.bool
//	err := row.GetMulti(
//		[]RowData{RowDataWrap, RowDataStyled},
//		[]unsafe.Pointer{unsafe.Pointer(&wrap), unsafe.Pointer(&styled)},
//	)
//
// C: ghostty_row_get_multi
func (r *Row) GetMulti(keys []RowData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Allocate the void** array in C memory to satisfy cgo pointer-passing rules.
	cVals, cValsSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cVals), cValsSize)
	return resultError(C.ghostty_row_get_multi(
		r.c,
		C.size_t(len(keys)),
		(*C.GhosttyRowData)(unsafe.Pointer(&keys[0])),
		cVals,
		nil,
	))
}

// Wrap reports whether the row is soft-wrapped.
func (r *Row) Wrap() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_WRAP, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// WrapContinuation reports whether the row is a continuation of
// a soft-wrapped row.
func (r *Row) WrapContinuation() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_WRAP_CONTINUATION, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Grapheme reports whether any cells in the row have grapheme clusters.
func (r *Row) Grapheme() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_GRAPHEME, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Styled reports whether any cells in the row have styling
// (may have false positives).
func (r *Row) Styled() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_STYLED, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Hyperlink reports whether any cells in the row have hyperlinks
// (may have false positives).
func (r *Row) Hyperlink() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_HYPERLINK, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Semantic returns the semantic prompt state of the row.
func (r *Row) Semantic() (RowSemanticPrompt, error) {
	var v C.GhosttyRowSemanticPrompt
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_SEMANTIC_PROMPT, unsafe.Pointer(&v))); err != nil {
		return 0, err
	}
	return RowSemanticPrompt(v), nil
}

// KittyVirtualPlaceholder reports whether the row contains a
// Kitty virtual placeholder.
func (r *Row) KittyVirtualPlaceholder() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_KITTY_VIRTUAL_PLACEHOLDER, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}

// Dirty reports whether the row is dirty and requires a redraw.
func (r *Row) Dirty() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_row_get(r.c, C.GHOSTTY_ROW_DATA_DIRTY, unsafe.Pointer(&v))); err != nil {
		return false, err
	}
	return bool(v), nil
}
