package libghostty

// Kitty graphics protocol bindings wrapping the C API from kitty_graphics.h.
// Provides access to images and placements stored via the Kitty graphics
// protocol.

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttySelection (sized struct).
static inline GhosttySelection init_selection() {
	GhosttySelection s = GHOSTTY_INIT_SIZED(GhosttySelection);
	return s;
}

// Helper to create a properly initialized GhosttyKittyGraphicsImageInfo (sized struct).
static inline GhosttyKittyGraphicsImageInfo init_kitty_image_info() {
	GhosttyKittyGraphicsImageInfo info = GHOSTTY_INIT_SIZED(GhosttyKittyGraphicsImageInfo);
	return info;
}

// Helper to create a properly initialized GhosttyKittyGraphicsPlacementInfo (sized struct).
static inline GhosttyKittyGraphicsPlacementInfo init_kitty_placement_info() {
	GhosttyKittyGraphicsPlacementInfo info = GHOSTTY_INIT_SIZED(GhosttyKittyGraphicsPlacementInfo);
	return info;
}

// Helper to create a properly initialized GhosttyKittyGraphicsPlacementRenderInfo (sized struct).
static inline GhosttyKittyGraphicsPlacementRenderInfo init_kitty_placement_render_info() {
	GhosttyKittyGraphicsPlacementRenderInfo info = GHOSTTY_INIT_SIZED(GhosttyKittyGraphicsPlacementRenderInfo);
	return info;
}
*/
import "C"

import "unsafe"

// KittyGraphics is a handle to the Kitty graphics image storage
// associated with a terminal's active screen. It is borrowed from
// the terminal and remains valid until the next mutating terminal
// call (e.g. VTWrite or Reset).
//
// C: GhosttyKittyGraphics
type KittyGraphics struct {
	ptr C.GhosttyKittyGraphics
}

// KittyGraphicsImage is a handle to a single Kitty graphics image.
// It is borrowed from the storage and remains valid until the next
// mutating terminal call.
//
// C: GhosttyKittyGraphicsImage
type KittyGraphicsImage struct {
	ptr C.GhosttyKittyGraphicsImage
}

// KittyGraphicsPlacementIterator iterates over placements in the
// Kitty graphics storage. It is independently owned and must be
// freed by calling Close, but the data it yields is only valid
// while the underlying terminal is not mutated.
//
// C: GhosttyKittyGraphicsPlacementIterator
type KittyGraphicsPlacementIterator struct {
	ptr C.GhosttyKittyGraphicsPlacementIterator
}

// KittyPlacementLayer classifies z-layer for kitty graphics placements.
// Based on the kitty protocol z-index conventions.
//
// C: GhosttyKittyPlacementLayer
type KittyPlacementLayer int

const (
	// KittyPlacementLayerAll disables layer filtering (all placements).
	KittyPlacementLayerAll KittyPlacementLayer = C.GHOSTTY_KITTY_PLACEMENT_LAYER_ALL

	// KittyPlacementLayerBelowBG matches placements below cell background
	// (z < INT32_MIN/2).
	KittyPlacementLayerBelowBG KittyPlacementLayer = C.GHOSTTY_KITTY_PLACEMENT_LAYER_BELOW_BG

	// KittyPlacementLayerBelowText matches placements above background but
	// below text (INT32_MIN/2 <= z < 0).
	KittyPlacementLayerBelowText KittyPlacementLayer = C.GHOSTTY_KITTY_PLACEMENT_LAYER_BELOW_TEXT

	// KittyPlacementLayerAboveText matches placements above text (z >= 0).
	KittyPlacementLayerAboveText KittyPlacementLayer = C.GHOSTTY_KITTY_PLACEMENT_LAYER_ABOVE_TEXT
)

// KittyImageFormat describes the pixel format of a Kitty graphics image.
//
// C: GhosttyKittyImageFormat
type KittyImageFormat int

const (
	// KittyImageFormatRGB is 24-bit RGB (3 bytes per pixel).
	KittyImageFormatRGB KittyImageFormat = C.GHOSTTY_KITTY_IMAGE_FORMAT_RGB

	// KittyImageFormatRGBA is 32-bit RGBA (4 bytes per pixel).
	KittyImageFormatRGBA KittyImageFormat = C.GHOSTTY_KITTY_IMAGE_FORMAT_RGBA

	// KittyImageFormatPNG is compressed PNG data.
	KittyImageFormatPNG KittyImageFormat = C.GHOSTTY_KITTY_IMAGE_FORMAT_PNG

	// KittyImageFormatGrayAlpha is 16-bit gray+alpha (2 bytes per pixel).
	KittyImageFormatGrayAlpha KittyImageFormat = C.GHOSTTY_KITTY_IMAGE_FORMAT_GRAY_ALPHA

	// KittyImageFormatGray is 8-bit grayscale (1 byte per pixel).
	KittyImageFormatGray KittyImageFormat = C.GHOSTTY_KITTY_IMAGE_FORMAT_GRAY
)

// KittyImageCompression describes the compression of a Kitty graphics image.
//
// C: GhosttyKittyImageCompression
type KittyImageCompression int

const (
	// KittyImageCompressionNone means no compression.
	KittyImageCompressionNone KittyImageCompression = C.GHOSTTY_KITTY_IMAGE_COMPRESSION_NONE

	// KittyImageCompressionZlibDeflate means zlib/deflate compression.
	KittyImageCompressionZlibDeflate KittyImageCompression = C.GHOSTTY_KITTY_IMAGE_COMPRESSION_ZLIB_DEFLATE
)

// Selection represents a grid selection range defined by two grid references.
//
// C: GhosttySelection
type Selection struct {
	// Start is the start of the selection range (inclusive).
	Start GridRef

	// End is the end of the selection range (inclusive).
	End GridRef

	// Rectangle indicates whether the selection is rectangular (block)
	// rather than linear.
	Rectangle bool
}

// selectionFromC converts a C GhosttySelection to a Go Selection.
func selectionFromC(cs C.GhosttySelection) Selection {
	return Selection{
		Start:     GridRef{ref: cs.start},
		End:       GridRef{ref: cs.end},
		Rectangle: bool(cs.rectangle),
	}
}

// KittyGraphicsImageInfo contains all image metadata in a single struct.
// This is more efficient than querying each field individually since it
// requires only one cgo call.
//
// C: GhosttyKittyGraphicsImageInfo
type KittyGraphicsImageInfo struct {
	// ID is the image ID.
	ID uint32

	// Number is the image number.
	Number uint32

	// Width is the image width in pixels.
	Width uint32

	// Height is the image height in pixels.
	Height uint32

	// Format is the pixel format of the image.
	Format KittyImageFormat

	// Compression is the compression of the image.
	Compression KittyImageCompression

	// Data is a borrowed slice of the raw pixel data. Only valid
	// until the next mutating terminal call.
	Data []byte
}

// KittyGraphicsPlacementInfo contains all placement metadata in a single
// struct. This is more efficient than querying each field individually
// since it requires only one cgo call.
//
// C: GhosttyKittyGraphicsPlacementInfo
type KittyGraphicsPlacementInfo struct {
	// ImageID is the image ID this placement belongs to.
	ImageID uint32

	// PlacementID is the placement ID.
	PlacementID uint32

	// IsVirtual indicates whether this is a virtual placement (unicode placeholder).
	IsVirtual bool

	// XOffset is the pixel offset from the left edge of the cell.
	XOffset uint32

	// YOffset is the pixel offset from the top edge of the cell.
	YOffset uint32

	// SourceX is the source rectangle x origin in pixels.
	SourceX uint32

	// SourceY is the source rectangle y origin in pixels.
	SourceY uint32

	// SourceWidth is the source rectangle width in pixels (0 = full image width).
	SourceWidth uint32

	// SourceHeight is the source rectangle height in pixels (0 = full image height).
	SourceHeight uint32

	// Columns is the number of columns this placement occupies.
	Columns uint32

	// Rows is the number of rows this placement occupies.
	Rows uint32

	// Z is the z-index for this placement.
	Z int32
}

// KittyGraphicsPlacementRenderInfo contains all rendering geometry for a
// placement in a single struct. Combines pixel size, grid size, viewport
// position, and source rectangle into one cgo call.
//
// C: GhosttyKittyGraphicsPlacementRenderInfo
type KittyGraphicsPlacementRenderInfo struct {
	// PixelWidth is the rendered width in pixels.
	PixelWidth uint32

	// PixelHeight is the rendered height in pixels.
	PixelHeight uint32

	// GridCols is the number of grid columns the placement occupies.
	GridCols uint32

	// GridRows is the number of grid rows the placement occupies.
	GridRows uint32

	// ViewportCol is the viewport-relative column (may be negative
	// for partially visible placements).
	ViewportCol int32

	// ViewportRow is the viewport-relative row (may be negative
	// for partially visible placements).
	ViewportRow int32

	// ViewportVisible is false when the placement is fully off-screen
	// or is a virtual placement. When false, ViewportCol and ViewportRow
	// may contain meaningless values.
	ViewportVisible bool

	// SourceX is the resolved source rectangle x origin in pixels.
	SourceX uint32

	// SourceY is the resolved source rectangle y origin in pixels.
	SourceY uint32

	// SourceWidth is the resolved source rectangle width in pixels.
	SourceWidth uint32

	// SourceHeight is the resolved source rectangle height in pixels.
	SourceHeight uint32
}

// PlacementIterator populates the given iterator with placement data
// from this storage. The iterator must have been created with
// NewKittyGraphicsPlacementIterator.
func (kg *KittyGraphics) PlacementIterator(iter *KittyGraphicsPlacementIterator) error {
	return resultError(C.ghostty_kitty_graphics_get(
		kg.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_DATA_PLACEMENT_ITERATOR,
		unsafe.Pointer(&iter.ptr),
	))
}

// Image looks up a Kitty graphics image by its image ID. Returns nil
// if no image with the given ID exists.
func (kg *KittyGraphics) Image(imageID uint32) *KittyGraphicsImage {
	ptr := C.ghostty_kitty_graphics_image(kg.ptr, C.uint32_t(imageID))
	if ptr == nil {
		return nil
	}
	return &KittyGraphicsImage{ptr: ptr}
}

// ID returns the image ID.
func (img *KittyGraphicsImage) ID() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_ID,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Number returns the image number.
func (img *KittyGraphicsImage) Number() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_NUMBER,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Width returns the image width in pixels.
func (img *KittyGraphicsImage) Width() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_WIDTH,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Height returns the image height in pixels.
func (img *KittyGraphicsImage) Height() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_HEIGHT,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Format returns the pixel format of the image.
func (img *KittyGraphicsImage) Format() (KittyImageFormat, error) {
	var v C.GhosttyKittyImageFormat
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_FORMAT,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return KittyImageFormat(v), nil
}

// Compression returns the compression of the image.
func (img *KittyGraphicsImage) Compression() (KittyImageCompression, error) {
	var v C.GhosttyKittyImageCompression
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_COMPRESSION,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return KittyImageCompression(v), nil
}

// Info returns all image metadata in a single call. This is more
// efficient than calling ID, Number, Width, Height, Format,
// Compression, and Data individually.
func (img *KittyGraphicsImage) Info() (*KittyGraphicsImageInfo, error) {
	ci := C.init_kitty_image_info()
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_INFO,
		unsafe.Pointer(&ci),
	)); err != nil {
		return nil, err
	}

	info := &KittyGraphicsImageInfo{
		ID:          uint32(ci.id),
		Number:      uint32(ci.number),
		Width:       uint32(ci.width),
		Height:      uint32(ci.height),
		Format:      KittyImageFormat(ci.format),
		Compression: KittyImageCompression(ci.compression),
	}

	if ci.data_ptr != nil && ci.data_len > 0 {
		info.Data = unsafe.Slice((*byte)(unsafe.Pointer(ci.data_ptr)), int(ci.data_len))
	}

	return info, nil
}

// Data returns a borrowed slice of the raw pixel data. The slice is
// only valid until the next mutating terminal call.
func (img *KittyGraphicsImage) Data() ([]byte, error) {
	var ptr *C.uint8_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_DATA_PTR,
		unsafe.Pointer(&ptr),
	)); err != nil {
		return nil, err
	}

	var length C.size_t
	if err := resultError(C.ghostty_kitty_graphics_image_get(
		img.ptr,
		C.GHOSTTY_KITTY_IMAGE_DATA_DATA_LEN,
		unsafe.Pointer(&length),
	)); err != nil {
		return nil, err
	}

	if ptr == nil || length == 0 {
		return nil, nil
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(length)), nil
}

// NewKittyGraphicsPlacementIterator creates a new placement iterator.
// Call KittyGraphics.PlacementIterator to populate it with data, then
// iterate with Next and read fields with the getter methods.
// The iterator must be freed by calling Close.
func NewKittyGraphicsPlacementIterator() (*KittyGraphicsPlacementIterator, error) {
	var ptr C.GhosttyKittyGraphicsPlacementIterator
	if err := resultError(C.ghostty_kitty_graphics_placement_iterator_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &KittyGraphicsPlacementIterator{ptr: ptr}, nil
}

// Close frees the placement iterator. After this call, the iterator
// must not be used.
func (it *KittyGraphicsPlacementIterator) Close() {
	C.ghostty_kitty_graphics_placement_iterator_free(it.ptr)
}

// SetLayer sets the z-layer filter for the iterator. Only placements
// matching the given layer will be returned by Next. The default is
// KittyPlacementLayerAll (no filtering).
func (it *KittyGraphicsPlacementIterator) SetLayer(layer KittyPlacementLayer) error {
	v := C.GhosttyKittyPlacementLayer(layer)
	return resultError(C.ghostty_kitty_graphics_placement_iterator_set(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_ITERATOR_OPTION_LAYER,
		unsafe.Pointer(&v),
	))
}

// Next advances the iterator to the next placement. Returns true if
// a placement is available, false when iteration is complete.
func (it *KittyGraphicsPlacementIterator) Next() bool {
	return bool(C.ghostty_kitty_graphics_placement_next(it.ptr))
}

// ImageID returns the image ID of the current placement.
func (it *KittyGraphicsPlacementIterator) ImageID() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IMAGE_ID,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// PlacementID returns the placement ID of the current placement.
func (it *KittyGraphicsPlacementIterator) PlacementID() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_PLACEMENT_ID,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// IsVirtual reports whether the current placement is a virtual
// (unicode placeholder) placement.
func (it *KittyGraphicsPlacementIterator) IsVirtual() (bool, error) {
	var v C.bool
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_IS_VIRTUAL,
		unsafe.Pointer(&v),
	)); err != nil {
		return false, err
	}
	return bool(v), nil
}

// XOffset returns the pixel offset from the left edge of the cell.
func (it *KittyGraphicsPlacementIterator) XOffset() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_X_OFFSET,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// YOffset returns the pixel offset from the top edge of the cell.
func (it *KittyGraphicsPlacementIterator) YOffset() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Y_OFFSET,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// SourceX returns the source rectangle x origin in pixels.
func (it *KittyGraphicsPlacementIterator) SourceX() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_X,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// SourceY returns the source rectangle y origin in pixels.
func (it *KittyGraphicsPlacementIterator) SourceY() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_Y,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// SourceWidth returns the source rectangle width in pixels
// (0 = full image width).
func (it *KittyGraphicsPlacementIterator) SourceWidth() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_WIDTH,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// SourceHeight returns the source rectangle height in pixels
// (0 = full image height).
func (it *KittyGraphicsPlacementIterator) SourceHeight() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_SOURCE_HEIGHT,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Columns returns the number of columns this placement occupies.
func (it *KittyGraphicsPlacementIterator) Columns() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_COLUMNS,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Rows returns the number of rows this placement occupies.
func (it *KittyGraphicsPlacementIterator) Rows() (uint32, error) {
	var v C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_ROWS,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// Z returns the z-index of the current placement.
func (it *KittyGraphicsPlacementIterator) Z() (int32, error) {
	var v C.int32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_Z,
		unsafe.Pointer(&v),
	)); err != nil {
		return 0, err
	}
	return int32(v), nil
}

// Info returns all placement metadata in a single call. This is more
// efficient than calling ImageID, PlacementID, IsVirtual, XOffset,
// YOffset, SourceX, SourceY, SourceWidth, SourceHeight, Columns,
// Rows, and Z individually.
func (it *KittyGraphicsPlacementIterator) Info() (*KittyGraphicsPlacementInfo, error) {
	ci := C.init_kitty_placement_info()
	if err := resultError(C.ghostty_kitty_graphics_placement_get(
		it.ptr,
		C.GHOSTTY_KITTY_GRAPHICS_PLACEMENT_DATA_INFO,
		unsafe.Pointer(&ci),
	)); err != nil {
		return nil, err
	}

	return &KittyGraphicsPlacementInfo{
		ImageID:      uint32(ci.image_id),
		PlacementID:  uint32(ci.placement_id),
		IsVirtual:    bool(ci.is_virtual),
		XOffset:      uint32(ci.x_offset),
		YOffset:      uint32(ci.y_offset),
		SourceX:      uint32(ci.source_x),
		SourceY:      uint32(ci.source_y),
		SourceWidth:  uint32(ci.source_width),
		SourceHeight: uint32(ci.source_height),
		Columns:      uint32(ci.columns),
		Rows:         uint32(ci.rows),
		Z:            int32(ci.z),
	}, nil
}

// RenderInfo returns all rendering geometry for the current placement
// in a single call. This combines the results of PixelSize, GridSize,
// ViewportPos, and SourceRect into one cgo call.
//
// When ViewportVisible is false, the placement is fully off-screen or
// is a virtual placement; ViewportCol and ViewportRow may contain
// meaningless values in that case.
func (it *KittyGraphicsPlacementIterator) RenderInfo(img *KittyGraphicsImage, t *Terminal) (*KittyGraphicsPlacementRenderInfo, error) {
	ci := C.init_kitty_placement_render_info()
	if err := resultError(C.ghostty_kitty_graphics_placement_render_info(
		it.ptr,
		img.ptr,
		t.ptr,
		&ci,
	)); err != nil {
		return nil, err
	}

	return &KittyGraphicsPlacementRenderInfo{
		PixelWidth:      uint32(ci.pixel_width),
		PixelHeight:     uint32(ci.pixel_height),
		GridCols:        uint32(ci.grid_cols),
		GridRows:        uint32(ci.grid_rows),
		ViewportCol:     int32(ci.viewport_col),
		ViewportRow:     int32(ci.viewport_row),
		ViewportVisible: bool(ci.viewport_visible),
		SourceX:         uint32(ci.source_x),
		SourceY:         uint32(ci.source_y),
		SourceWidth:     uint32(ci.source_width),
		SourceHeight:    uint32(ci.source_height),
	}, nil
}

// Rect computes the grid rectangle occupied by the current placement.
// Virtual placements (unicode placeholders) return an error with
// ResultNoValue.
func (it *KittyGraphicsPlacementIterator) Rect(img *KittyGraphicsImage, t *Terminal) (*Selection, error) {
	cs := C.init_selection()
	if err := resultError(C.ghostty_kitty_graphics_placement_rect(
		it.ptr,
		img.ptr,
		t.ptr,
		&cs,
	)); err != nil {
		return nil, err
	}
	sel := selectionFromC(cs)
	return &sel, nil
}

// PixelSize computes the rendered pixel dimensions of the current
// placement, accounting for the source rectangle, specified
// columns/rows, and aspect ratio.
func (it *KittyGraphicsPlacementIterator) PixelSize(img *KittyGraphicsImage, t *Terminal) (width, height uint32, err error) {
	var w, h C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_pixel_size(
		it.ptr,
		img.ptr,
		t.ptr,
		&w,
		&h,
	)); err != nil {
		return 0, 0, err
	}
	return uint32(w), uint32(h), nil
}

// GridSize computes the number of grid columns and rows the current
// placement occupies.
func (it *KittyGraphicsPlacementIterator) GridSize(img *KittyGraphicsImage, t *Terminal) (cols, rows uint32, err error) {
	var c, r C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_grid_size(
		it.ptr,
		img.ptr,
		t.ptr,
		&c,
		&r,
	)); err != nil {
		return 0, 0, err
	}
	return uint32(c), uint32(r), nil
}

// ViewportPos returns the viewport-relative grid position of the
// current placement. The row can be negative for partially visible
// placements. Returns an error with ResultNoValue when fully
// off-screen or for virtual placements.
func (it *KittyGraphicsPlacementIterator) ViewportPos(img *KittyGraphicsImage, t *Terminal) (col, row int32, err error) {
	var c, r C.int32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_viewport_pos(
		it.ptr,
		img.ptr,
		t.ptr,
		&c,
		&r,
	)); err != nil {
		return 0, 0, err
	}
	return int32(c), int32(r), nil
}

// SourceRect returns the resolved source rectangle for the current
// placement in pixels, clamped to the actual image bounds. A width
// or height of 0 in the placement means "use the full image dimension".
func (it *KittyGraphicsPlacementIterator) SourceRect(img *KittyGraphicsImage) (x, y, width, height uint32, err error) {
	var cx, cy, cw, ch C.uint32_t
	if err := resultError(C.ghostty_kitty_graphics_placement_source_rect(
		it.ptr,
		img.ptr,
		&cx,
		&cy,
		&cw,
		&ch,
	)); err != nil {
		return 0, 0, 0, 0, err
	}
	return uint32(cx), uint32(cy), uint32(cw), uint32(ch), nil
}
