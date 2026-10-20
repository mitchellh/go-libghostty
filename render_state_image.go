package libghostty

// Render-state image iterator wrapping the
// GhosttyRenderStateImageIterator C APIs.

/*
#include <ghostty/vt.h>

// render_image_info holds every value of one image placement. It is
// filled by render_image_info_get and returned by value, so Go never has
// to pass C pointers to Go memory and the whole read is one cgo call.
typedef struct {
	// result is GHOSTTY_SUCCESS when every field below was read.
	// Otherwise it is the error from the first value that failed.
	GhosttyResult result;

	GhosttyImageProtocol protocol;
	uint32_t image_id;
	uint64_t image_generation;
	GhosttyRenderStateImageLayer layer;
	int32_t z;
	int32_t viewport_col;
	int32_t viewport_row;
	uint32_t grid_cols;
	uint32_t grid_rows;
	uint32_t cell_offset_x;
	uint32_t cell_offset_y;
	uint32_t pixel_width;
	uint32_t pixel_height;
	uint32_t source_x;
	uint32_t source_y;
	uint32_t source_width;
	uint32_t source_height;

	// The Kitty fields are only read when protocol is
	// GHOSTTY_IMAGE_PROTOCOL_KITTY, and are zero otherwise.
	uint32_t kitty_placement_id;
	GhosttyRenderStateKittyPlacementKind kitty_kind;
} render_image_info;

// render_image_info_get reads every value of the iterator's current
// placement. The common values are read with one get_multi call. The
// protocol-specific values are read with a second call only for that
// protocol, because reading them for another protocol fails with
// GHOSTTY_NO_VALUE and get_multi stops at the first error.
static inline render_image_info render_image_info_get(
	GhosttyRenderStateImageIterator it
) {
	render_image_info out = {0};

	GhosttyRenderStateImageData keys[] = {
		GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_Z,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH,
		GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT,
	};
	void* values[] = {
		&out.protocol,
		&out.image_id,
		&out.image_generation,
		&out.layer,
		&out.z,
		&out.viewport_col,
		&out.viewport_row,
		&out.grid_cols,
		&out.grid_rows,
		&out.cell_offset_x,
		&out.cell_offset_y,
		&out.pixel_width,
		&out.pixel_height,
		&out.source_x,
		&out.source_y,
		&out.source_width,
		&out.source_height,
	};
	out.result = ghostty_render_state_image_get_multi(
		it,
		sizeof(keys) / sizeof(keys[0]),
		keys,
		values,
		NULL
	);
	if (out.result != GHOSTTY_SUCCESS) return out;

	if (out.protocol == GHOSTTY_IMAGE_PROTOCOL_KITTY) {
		GhosttyRenderStateImageData kitty_keys[] = {
			GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID,
			GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND,
		};
		void* kitty_values[] = {
			&out.kitty_placement_id,
			&out.kitty_kind,
		};
		out.result = ghostty_render_state_image_get_multi(
			it,
			sizeof(kitty_keys) / sizeof(kitty_keys[0]),
			kitty_keys,
			kitty_values,
			NULL
		);
	}

	return out;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// ImageProtocol identifies the graphics protocol a program used to send an
// image. It tells a renderer where to read the image's pixels.
//
// Later versions may add more protocols. A renderer should skip images
// with a protocol it doesn't support, so that it keeps working when that
// happens.
//
// C: GhosttyImageProtocol
type ImageProtocol int

const (
	// ImageProtocolKitty is the Kitty graphics protocol. Read the image's
	// pixels with [KittyGraphics.Image].
	ImageProtocolKitty ImageProtocol = C.GHOSTTY_IMAGE_PROTOCOL_KITTY
)

// RenderStateImageLayer says whether an image is drawn behind or in front
// of the terminal's text.
//
// Draw a frame from back to front in this order:
//
//  1. Images in [RenderStateImageLayerBelowBackground].
//  2. Cell background colors.
//  3. Images in [RenderStateImageLayerBelowText].
//  4. Text.
//  5. Images in [RenderStateImageLayerAboveText].
//
// Within one layer, draw images with a lower z-index first.
//
// C: GhosttyRenderStateImageLayer
type RenderStateImageLayer int

const (
	// RenderStateImageLayerBelowBackground draws the image behind the cell
	// background colors.
	RenderStateImageLayerBelowBackground RenderStateImageLayer = C.GHOSTTY_RENDER_STATE_IMAGE_LAYER_BELOW_BACKGROUND

	// RenderStateImageLayerBelowText draws the image in front of the cell
	// background colors and behind the text.
	RenderStateImageLayerBelowText RenderStateImageLayer = C.GHOSTTY_RENDER_STATE_IMAGE_LAYER_BELOW_TEXT

	// RenderStateImageLayerAboveText draws the image in front of the text.
	RenderStateImageLayerAboveText RenderStateImageLayer = C.GHOSTTY_RENDER_STATE_IMAGE_LAYER_ABOVE_TEXT
)

// RenderStateKittyPlacementKind describes how a Kitty graphics image in a
// render state got its position. A renderer can draw every kind the same
// way. The kind is mostly useful for debugging.
//
// C: GhosttyRenderStateKittyPlacementKind
type RenderStateKittyPlacementKind int

const (
	// RenderStateKittyPlacementKindVirtual is one row of a virtual
	// placement. A program shows a virtual placement by printing
	// placeholder characters where the image should appear. Each run of
	// neighboring placeholder cells in a row becomes one image in the
	// render state, so this kind is always one row tall.
	RenderStateKittyPlacementKindVirtual RenderStateKittyPlacementKind = C.GHOSTTY_RENDER_STATE_KITTY_PLACEMENT_KIND_VIRTUAL

	// RenderStateKittyPlacementKindRelative is a placement that a program
	// positioned relative to a virtual placement. Its position is measured
	// from the top-left corner of the virtual placement's placeholder
	// cells. It is left out when none of those cells are in the viewport.
	RenderStateKittyPlacementKindRelative RenderStateKittyPlacementKind = C.GHOSTTY_RENDER_STATE_KITTY_PLACEMENT_KIND_RELATIVE
)

// RenderStateImageData identifies one value of an image in a render state.
// Most callers should use [RenderStateImageIterator.Info] or the typed
// getters instead. These keys are for [RenderStateImageIterator.GetMulti].
//
// Each key's comment ends with the C type of its value in parentheses.
// Pixel values are based on the cell size set with [Terminal.Resize].
//
// Keys with a protocol in their name, such as
// [RenderStateImageDataKittyKind], only apply to images from that
// protocol. Reading one for any other image returns an error matching
// [ErrNoValue].
//
// C: GhosttyRenderStateImageData
type RenderStateImageData int

const (
	// RenderStateImageDataInvalid is not a valid key. Reading it returns
	// an error.
	RenderStateImageDataInvalid RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_INVALID

	// RenderStateImageDataProtocol is the protocol the image came from
	// (GhosttyImageProtocol).
	RenderStateImageDataProtocol RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL

	// RenderStateImageDataImageID is the ID of the image (uint32_t).
	RenderStateImageDataImageID RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID

	// RenderStateImageDataImageGeneration is the generation of the image's
	// pixels (uint64_t).
	RenderStateImageDataImageGeneration RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION

	// RenderStateImageDataLayer is the layer to draw the image in
	// (GhosttyRenderStateImageLayer).
	RenderStateImageDataLayer RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER

	// RenderStateImageDataZ is the z-index of the image within its layer
	// (int32_t).
	RenderStateImageDataZ RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_Z

	// RenderStateImageDataViewportCol is the column of the image's
	// top-left cell (int32_t).
	RenderStateImageDataViewportCol RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL

	// RenderStateImageDataViewportRow is the row of the image's top-left
	// cell (int32_t).
	RenderStateImageDataViewportRow RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW

	// RenderStateImageDataGridCols is the number of columns the image
	// covers (uint32_t).
	RenderStateImageDataGridCols RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS

	// RenderStateImageDataGridRows is the number of rows the image covers
	// (uint32_t).
	RenderStateImageDataGridRows RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS

	// RenderStateImageDataCellOffsetX is the distance in pixels from the
	// left edge of the top-left cell to the image (uint32_t).
	RenderStateImageDataCellOffsetX RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X

	// RenderStateImageDataCellOffsetY is the distance in pixels from the
	// top edge of the top-left cell to the image (uint32_t).
	RenderStateImageDataCellOffsetY RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y

	// RenderStateImageDataPixelWidth is the width to draw the image at, in
	// pixels (uint32_t).
	RenderStateImageDataPixelWidth RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH

	// RenderStateImageDataPixelHeight is the height to draw the image at,
	// in pixels (uint32_t).
	RenderStateImageDataPixelHeight RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT

	// RenderStateImageDataSourceX is the left edge of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceX RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X

	// RenderStateImageDataSourceY is the top edge of the part of the image
	// to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceY RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y

	// RenderStateImageDataSourceWidth is the width of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceWidth RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH

	// RenderStateImageDataSourceHeight is the height of the part of the
	// image to draw, in image pixels (uint32_t).
	RenderStateImageDataSourceHeight RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT

	// RenderStateImageDataKittyPlacementID is the Kitty placement ID, or 0
	// if there isn't one (uint32_t). Kitty images only.
	RenderStateImageDataKittyPlacementID RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID

	// RenderStateImageDataKittyKind is how the Kitty image got its
	// position (GhosttyRenderStateKittyPlacementKind). Kitty images only.
	RenderStateImageDataKittyKind RenderStateImageData = C.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND
)

// RenderStateImageInfo holds everything needed to draw one image from a
// render state. Read it with [RenderStateImageIterator.Info].
//
// It describes a single rectangle. Copy the source rectangle (SourceX,
// SourceY, SourceWidth and SourceHeight) out of the image and scale it to
// PixelWidth by PixelHeight. Place its top-left corner at this position,
// in pixels from the top-left corner of the viewport:
//
//	x := info.ViewportCol*cellWidth + int32(info.CellOffsetX)
//	y := info.ViewportRow*cellHeight + int32(info.CellOffsetY)
//
// Here cellWidth and cellHeight are the cell size passed to
// [Terminal.Resize].
type RenderStateImageInfo struct {
	// Protocol is the protocol the image came from. Skip images whose
	// protocol you don't support.
	Protocol ImageProtocol

	// ImageID identifies the image. IDs are only unique within one
	// protocol, so look the image up in that protocol's storage. For
	// Kitty images, use [KittyGraphics.Image].
	ImageID uint32

	// ImageGeneration changes whenever the image's pixels change, such as
	// when an animation moves to its next frame. Together with Protocol
	// and ImageID, it makes a good key for caching textures. For Kitty
	// images, it equals [KittyGraphicsImage.Generation].
	ImageGeneration uint64

	// Layer says whether to draw the image behind or in front of the text.
	Layer RenderStateImageLayer

	// Z orders images within the same layer. Draw lower values first.
	// Kitty virtual placements always use -1.
	Z int32

	// ViewportCol is the column of the image's top-left cell. Column 0 is
	// the left edge of the viewport. It is negative when the image starts
	// to the left of the viewport.
	ViewportCol int32

	// ViewportRow is the row of the image's top-left cell. Row 0 is the
	// top of the viewport. It is negative when the image starts above the
	// viewport, including in an overscan row.
	ViewportRow int32

	// GridCols is the number of columns the image covers.
	GridCols uint32

	// GridRows is the number of rows the image covers.
	GridRows uint32

	// CellOffsetX is the distance in pixels from the left edge of the
	// top-left cell to the image. An image narrower than its cells is
	// centered, so the offset can be wider than one cell.
	CellOffsetX uint32

	// CellOffsetY is the distance in pixels from the top edge of the
	// top-left cell to the image. An image shorter than its cells is
	// centered, so the offset can be taller than one cell.
	CellOffsetY uint32

	// PixelWidth is the width to draw the image at, in pixels.
	PixelWidth uint32

	// PixelHeight is the height to draw the image at, in pixels.
	PixelHeight uint32

	// SourceX is the left edge of the part of the image to draw, in image
	// pixels.
	SourceX uint32

	// SourceY is the top edge of the part of the image to draw, in image
	// pixels.
	SourceY uint32

	// SourceWidth is the width of the part of the image to draw, in image
	// pixels.
	SourceWidth uint32

	// SourceHeight is the height of the part of the image to draw, in
	// image pixels.
	SourceHeight uint32

	// KittyPlacementID is the Kitty placement ID, or 0 if there isn't one.
	// It is only valid when Protocol is ImageProtocolKitty. See
	// [RenderStateImageIterator.KittyPlacementID] for details.
	KittyPlacementID uint32

	// KittyKind is how the Kitty image got its position. It is only valid
	// when Protocol is ImageProtocolKitty.
	KittyKind RenderStateKittyPlacementKind
}

// RenderStateImageIterator lists the images to draw for a render state.
//
// Create one iterator with [NewRenderStateImageIterator] and keep it for
// as long as you render. After each [RenderState.Update], fill it with
// [RenderState.ImageIterator]. Then call [RenderStateImageIterator.Next]
// in a loop and read each image with [RenderStateImageIterator.Info].
//
// Images are listed in no particular order. Sort them by layer and then by
// z-index before drawing.
//
// For now, the list only includes Kitty graphics images that a program
// placed with Unicode placeholder characters (U+10EEEE). These images move
// with the text, so the render state works out where they are during each
// update. Kitty images placed at a fixed position are not included yet.
// Read those with a [KittyGraphicsPlacementIterator] and sort both lists
// together. Later versions may add more protocols and kinds of images.
//
// The list is empty unless the terminal has a cell size in pixels, which
// you set with [Terminal.Resize].
//
// The render state does not copy image pixels. Read them from the terminal
// right after the update, while you still have exclusive access to it. An
// image only needs to be read again when its generation changes.
//
// The values in an iterator stay valid until the next update of the render
// state that filled it. Values returned by its methods are copies and can
// be kept longer.
//
// C: GhosttyRenderStateImageIterator
type RenderStateImageIterator struct {
	ptr C.GhosttyRenderStateImageIterator
}

// NewRenderStateImageIterator returns a new, empty image iterator. Fill it
// with [RenderState.ImageIterator]. Call [RenderStateImageIterator.Close]
// when it is no longer needed.
func NewRenderStateImageIterator() (*RenderStateImageIterator, error) {
	var ptr C.GhosttyRenderStateImageIterator
	if err := resultError(C.ghostty_render_state_image_iterator_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &RenderStateImageIterator{ptr: ptr}, nil
}

// Close frees the iterator. The iterator must not be used afterward.
// Calling Close more than once is safe.
func (it *RenderStateImageIterator) Close() {
	if it == nil || it.ptr == nil {
		return
	}
	C.ghostty_render_state_image_iterator_free(it.ptr)
	it.ptr = nil
}

// Next moves to the next image and reports whether there is one. Call it
// before reading the first image.
//
// Only images with something to draw are listed. For example, an image
// that the terminal has not fully received yet is skipped.
func (it *RenderStateImageIterator) Next() bool {
	return bool(C.ghostty_render_state_image_iterator_next(it.ptr))
}

// Info returns every value of the current image. It is faster than
// calling the individual getters, because it reads all of the values in
// a single call into libghostty.
func (it *RenderStateImageIterator) Info() (*RenderStateImageInfo, error) {
	v := C.render_image_info_get(it.ptr)
	if err := resultError(v.result); err != nil {
		return nil, err
	}
	return &RenderStateImageInfo{
		Protocol:         ImageProtocol(v.protocol),
		ImageID:          uint32(v.image_id),
		ImageGeneration:  uint64(v.image_generation),
		Layer:            RenderStateImageLayer(v.layer),
		Z:                int32(v.z),
		ViewportCol:      int32(v.viewport_col),
		ViewportRow:      int32(v.viewport_row),
		GridCols:         uint32(v.grid_cols),
		GridRows:         uint32(v.grid_rows),
		CellOffsetX:      uint32(v.cell_offset_x),
		CellOffsetY:      uint32(v.cell_offset_y),
		PixelWidth:       uint32(v.pixel_width),
		PixelHeight:      uint32(v.pixel_height),
		SourceX:          uint32(v.source_x),
		SourceY:          uint32(v.source_y),
		SourceWidth:      uint32(v.source_width),
		SourceHeight:     uint32(v.source_height),
		KittyPlacementID: uint32(v.kitty_placement_id),
		KittyKind:        RenderStateKittyPlacementKind(v.kitty_kind),
	}, nil
}

// GetMulti reads several values of the current image in one call. Most
// callers should use [RenderStateImageIterator.Info] instead. GetMulti is
// for callers that need only a few values and want to avoid the cost of
// reading the rest.
//
// values[i] receives the value for keys[i]. Each element of values must
// point to a variable with the same size as the C type listed for its key
// on [RenderStateImageData]. For example, use uint32 for uint32_t, uint64
// for uint64_t, and int32 for an enum.
//
// GetMulti returns an error if keys and values have different lengths or
// if a key can't be read. Values for the keys before the one that failed
// may already be written.
//
// C: ghostty_render_state_image_get_multi
func (it *RenderStateImageIterator) GetMulti(keys []RenderStateImageData, values []unsafe.Pointer) error {
	if len(keys) != len(values) {
		return errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil
	}
	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[C.GhosttyRenderStateImageData](&args, keys, values)
	if err != nil {
		return err
	}
	defer args.free()
	return resultError(C.ghostty_render_state_image_get_multi(
		it.ptr,
		C.size_t(len(keys)),
		cKeys,
		args.values,
		nil,
	))
}

// get reads one value of the current image into out, which must point to
// the C type listed for data on [RenderStateImageData].
func (it *RenderStateImageIterator) get(data C.GhosttyRenderStateImageData, out unsafe.Pointer) error {
	return resultError(C.ghostty_render_state_image_get(it.ptr, data, out))
}

// Protocol returns the protocol the current image came from. See
// [RenderStateImageInfo.Protocol].
func (it *RenderStateImageIterator) Protocol() (ImageProtocol, error) {
	var v C.GhosttyImageProtocol
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PROTOCOL, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return ImageProtocol(v), nil
}

// ImageID returns the ID of the current image. See
// [RenderStateImageInfo.ImageID].
func (it *RenderStateImageIterator) ImageID() (uint32, error) {
	var v C.uint32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_ID, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// ImageGeneration returns the generation of the current image's pixels.
// See [RenderStateImageInfo.ImageGeneration].
func (it *RenderStateImageIterator) ImageGeneration() (uint64, error) {
	var v C.uint64_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_IMAGE_GENERATION, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return uint64(v), nil
}

// Layer returns the layer to draw the current image in.
func (it *RenderStateImageIterator) Layer() (RenderStateImageLayer, error) {
	var v C.GhosttyRenderStateImageLayer
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_LAYER, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return RenderStateImageLayer(v), nil
}

// Z returns the z-index of the current image within its layer. Draw lower
// values first.
func (it *RenderStateImageIterator) Z() (int32, error) {
	var v C.int32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_Z, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return int32(v), nil
}

// ViewportPos returns the column and row of the current image's top-left
// cell, counted from the top-left corner of the viewport. Either value is
// negative when the image starts outside the viewport.
func (it *RenderStateImageIterator) ViewportPos() (col, row int32, err error) {
	var c, r C.int32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_COL, unsafe.Pointer(&c)); err != nil {
		return 0, 0, err
	}
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_VIEWPORT_ROW, unsafe.Pointer(&r)); err != nil {
		return 0, 0, err
	}
	return int32(c), int32(r), nil
}

// GridSize returns the number of columns and rows the current image
// covers.
func (it *RenderStateImageIterator) GridSize() (cols, rows uint32, err error) {
	var c, r C.uint32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_COLS, unsafe.Pointer(&c)); err != nil {
		return 0, 0, err
	}
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_GRID_ROWS, unsafe.Pointer(&r)); err != nil {
		return 0, 0, err
	}
	return uint32(c), uint32(r), nil
}

// CellOffset returns the distance in pixels from the top-left corner of
// the current image's top-left cell to the image itself. See
// [RenderStateImageInfo.CellOffsetX].
func (it *RenderStateImageIterator) CellOffset() (x, y uint32, err error) {
	var cx, cy C.uint32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_X, unsafe.Pointer(&cx)); err != nil {
		return 0, 0, err
	}
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_CELL_OFFSET_Y, unsafe.Pointer(&cy)); err != nil {
		return 0, 0, err
	}
	return uint32(cx), uint32(cy), nil
}

// PixelSize returns the size to draw the current image at, in pixels.
func (it *RenderStateImageIterator) PixelSize() (width, height uint32, err error) {
	var w, h C.uint32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_WIDTH, unsafe.Pointer(&w)); err != nil {
		return 0, 0, err
	}
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_PIXEL_HEIGHT, unsafe.Pointer(&h)); err != nil {
		return 0, 0, err
	}
	return uint32(w), uint32(h), nil
}

// SourceRect returns the part of the current image to draw, in image
// pixels. Scale it to the size returned by
// [RenderStateImageIterator.PixelSize].
func (it *RenderStateImageIterator) SourceRect() (x, y, width, height uint32, err error) {
	var sx, sy, sw, sh C.uint32_t
	keys := [...]C.GhosttyRenderStateImageData{
		C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_X,
		C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_Y,
		C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_WIDTH,
		C.GHOSTTY_RENDER_STATE_IMAGE_DATA_SOURCE_HEIGHT,
	}
	outs := [...]unsafe.Pointer{
		unsafe.Pointer(&sx),
		unsafe.Pointer(&sy),
		unsafe.Pointer(&sw),
		unsafe.Pointer(&sh),
	}
	for i, key := range keys {
		if err := it.get(key, outs[i]); err != nil {
			return 0, 0, 0, 0, err
		}
	}
	return uint32(sx), uint32(sy), uint32(sw), uint32(sh), nil
}

// KittyPlacementID returns the Kitty placement ID of the current image, or
// 0 if there isn't one.
//
// For a virtual placement, this is the placement ID that the placeholder
// cells ask for. Cells that don't ask for a specific placement report 0.
// For a relative placement, this is the ID the program gave it, or 0 if
// it didn't give one.
//
// KittyPlacementID returns an error matching [ErrNoValue] when the image
// is not a Kitty image.
func (it *RenderStateImageIterator) KittyPlacementID() (uint32, error) {
	var v C.uint32_t
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_PLACEMENT_ID, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// KittyKind returns how the current Kitty image got its position. It
// returns an error matching [ErrNoValue] when the image is not a Kitty
// image.
func (it *RenderStateImageIterator) KittyKind() (RenderStateKittyPlacementKind, error) {
	var v C.GhosttyRenderStateKittyPlacementKind
	if err := it.get(C.GHOSTTY_RENDER_STATE_IMAGE_DATA_KITTY_KIND, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return RenderStateKittyPlacementKind(v), nil
}
