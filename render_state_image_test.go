package libghostty

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"testing"
	"unsafe"
)

// This example draws the images in a render state. A program shows a Kitty
// image with Unicode placeholder characters, and the render state works
// out where to draw it.
func ExampleRenderStateImageIterator() {
	term, err := NewTerminal(
		WithSize(10, 5),
		WithWritePty(func(*Terminal, []byte) {}),
	)
	if err != nil {
		panic(err)
	}
	defer term.Close()

	// Images need a cell size in pixels. Each cell here is 10x10.
	const cellWidth, cellHeight = 10, 10
	if err := term.Resize(10, 5, cellWidth, cellHeight); err != nil {
		panic(err)
	}
	limit := uint64(1 << 20)
	if err := term.SetKittyImageStorageLimit(&limit); err != nil {
		panic(err)
	}

	// Send a 1x2 pixel image with a virtual placement two cells wide, then
	// print two placeholder cells for it on the second row.
	term.VTWrite([]byte("\x1b_Ga=T,t=d,f=24,i=1,s=1,v=2,U=1,c=2,r=1,q=2;////////\x1b\\"))
	term.VTWrite([]byte("\x1b[2;4H\x1b[38;5;1m\U0010EEEE\U0010EEEE"))

	rs, err := NewRenderState()
	if err != nil {
		panic(err)
	}
	defer rs.Close()

	// A renderer keeps one iterator and fills it again after every update.
	images, err := NewRenderStateImageIterator()
	if err != nil {
		panic(err)
	}
	defer images.Close()

	if err := rs.Update(term); err != nil {
		panic(err)
	}
	if err := rs.ImageIterator(images); err != nil {
		panic(err)
	}

	var draws []*RenderStateImageInfo
	for images.Next() {
		info, err := images.Info()
		if err != nil {
			panic(err)
		}
		if info.Protocol != ImageProtocolKitty {
			continue // A protocol this renderer doesn't support.
		}
		draws = append(draws, info)
	}

	// Images are listed in no particular order, so sort them into drawing
	// order: by layer, then by z-index.
	slices.SortFunc(draws, func(a, b *RenderStateImageInfo) int {
		return cmp.Or(cmp.Compare(a.Layer, b.Layer), cmp.Compare(a.Z, b.Z))
	})

	for _, info := range draws {
		x := info.ViewportCol*cellWidth + int32(info.CellOffsetX)
		y := info.ViewportRow*cellHeight + int32(info.CellOffsetY)
		fmt.Printf("image %d at (%d, %d), size %dx%d\n",
			info.ImageID, x, y, info.PixelWidth, info.PixelHeight)
	}

	// Output:
	// image 1 at (38, 10), size 5x10
}

func TestRenderStateImageIteratorEmpty(t *testing.T) {
	it, err := NewRenderStateImageIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()

	// A new iterator is empty until it is filled from a render state.
	if it.Next() {
		t.Fatal("expected empty iterator")
	}
	if _, err := it.ImageID(); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}

	// Close is idempotent.
	it.Close()
	it.Close()
}

// newImageRenderState writes a Kitty virtual placement and two placeholder
// cells for it, then returns an updated render state and a filled image
// iterator. It mirrors the upstream C API test.
func newImageRenderState(t *testing.T) (*Terminal, *RenderState, *RenderStateImageIterator) {
	t.Helper()

	term := newKittyTerminal(t)
	t.Cleanup(term.Close)
	if err := term.Resize(10, 5, 10, 10); err != nil {
		t.Fatal(err)
	}

	// Transmit a 1x2 RGB image with a 2x1 virtual placement (U=1), then
	// write two placeholder cells (U+10EEEE) for it at row 1, column 3.
	// The foreground color encodes the image ID and the underline color
	// encodes the placement ID.
	term.VTWrite([]byte("\x1b_Ga=T,t=d,f=24,i=1,p=1,s=1,v=2,U=1,c=2,r=1,q=2;////////\x1b\\"))
	term.VTWrite([]byte("\x1b[2;4H\x1b[38;5;1;58;5;1m\U0010EEEE\U0010EEEE"))

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rs.Close)
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	it, err := NewRenderStateImageIterator()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(it.Close)
	if err := rs.ImageIterator(it); err != nil {
		t.Fatal(err)
	}

	return term, rs, it
}

func TestRenderStateImageIteratorKittyVirtual(t *testing.T) {
	term, _, it := newImageRenderState(t)

	if !it.Next() {
		t.Fatal("expected a placement")
	}

	info, err := it.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Protocol != ImageProtocolKitty {
		t.Fatalf("expected kitty protocol, got %d", info.Protocol)
	}
	if info.Layer != RenderStateImageLayerBelowText {
		t.Fatalf("expected below-text layer, got %d", info.Layer)
	}
	if info.KittyKind != RenderStateKittyPlacementKindVirtual {
		t.Fatalf("expected virtual kind, got %d", info.KittyKind)
	}
	if info.ImageID != 1 || info.KittyPlacementID != 1 {
		t.Fatalf("expected image 1 placement 1, got %d %d", info.ImageID, info.KittyPlacementID)
	}
	if info.ViewportCol != 3 || info.ViewportRow != 1 {
		t.Fatalf("expected position (3, 1), got (%d, %d)", info.ViewportCol, info.ViewportRow)
	}
	if info.GridCols != 2 || info.GridRows != 1 {
		t.Fatalf("expected grid 2x1, got %dx%d", info.GridCols, info.GridRows)
	}
	if info.SourceHeight != 2 {
		t.Fatalf("expected source height 2, got %d", info.SourceHeight)
	}
	if info.Z != -1 {
		t.Fatalf("expected z -1, got %d", info.Z)
	}

	// The 1x2 pixel image is scaled to fit 2x1 cells of 10x10 pixels.
	// That makes it 5 pixels wide, centered in the 20 pixel width.
	if info.PixelWidth != 5 || info.CellOffsetX != 8 {
		t.Fatalf("expected width 5 offset 8, got %d %d", info.PixelWidth, info.CellOffsetX)
	}

	// The generation matches the image in the terminal's storage.
	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}
	gen, err := kg.Image(1).Generation()
	if err != nil {
		t.Fatal(err)
	}
	if info.ImageGeneration != gen {
		t.Fatalf("expected generation %d, got %d", gen, info.ImageGeneration)
	}

	// The typed getters agree with Info.
	if v, err := it.Protocol(); err != nil || v != info.Protocol {
		t.Fatalf("Protocol: %d %v", v, err)
	}
	if v, err := it.ImageID(); err != nil || v != info.ImageID {
		t.Fatalf("ImageID: %d %v", v, err)
	}
	if v, err := it.ImageGeneration(); err != nil || v != info.ImageGeneration {
		t.Fatalf("ImageGeneration: %d %v", v, err)
	}
	if v, err := it.Layer(); err != nil || v != info.Layer {
		t.Fatalf("Layer: %d %v", v, err)
	}
	if v, err := it.Z(); err != nil || v != info.Z {
		t.Fatalf("Z: %d %v", v, err)
	}
	if c, r, err := it.ViewportPos(); err != nil || c != info.ViewportCol || r != info.ViewportRow {
		t.Fatalf("ViewportPos: %d %d %v", c, r, err)
	}
	if c, r, err := it.GridSize(); err != nil || c != info.GridCols || r != info.GridRows {
		t.Fatalf("GridSize: %d %d %v", c, r, err)
	}
	if x, y, err := it.CellOffset(); err != nil || x != info.CellOffsetX || y != info.CellOffsetY {
		t.Fatalf("CellOffset: %d %d %v", x, y, err)
	}
	if w, h, err := it.PixelSize(); err != nil || w != info.PixelWidth || h != info.PixelHeight {
		t.Fatalf("PixelSize: %d %d %v", w, h, err)
	}
	if x, y, w, h, err := it.SourceRect(); err != nil ||
		x != info.SourceX || y != info.SourceY ||
		w != info.SourceWidth || h != info.SourceHeight {
		t.Fatalf("SourceRect: %d %d %d %d %v", x, y, w, h, err)
	}
	if v, err := it.KittyPlacementID(); err != nil || v != info.KittyPlacementID {
		t.Fatalf("KittyPlacementID: %d %v", v, err)
	}
	if v, err := it.KittyKind(); err != nil || v != info.KittyKind {
		t.Fatalf("KittyKind: %d %v", v, err)
	}

	if it.Next() {
		t.Fatal("expected one placement")
	}
	if _, err := it.Info(); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestRenderStateImageIteratorGetMulti(t *testing.T) {
	_, _, it := newImageRenderState(t)
	if !it.Next() {
		t.Fatal("expected a placement")
	}

	var protocol int32
	var imageID uint32
	var col, row int32
	err := it.GetMulti(
		[]RenderStateImageData{
			RenderStateImageDataProtocol,
			RenderStateImageDataImageID,
			RenderStateImageDataViewportCol,
			RenderStateImageDataViewportRow,
		},
		[]unsafe.Pointer{
			unsafe.Pointer(&protocol),
			unsafe.Pointer(&imageID),
			unsafe.Pointer(&col),
			unsafe.Pointer(&row),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if ImageProtocol(protocol) != ImageProtocolKitty || imageID != 1 || col != 3 || row != 1 {
		t.Fatalf("unexpected values: %d %d %d %d", protocol, imageID, col, row)
	}

	// Mismatched lengths are rejected before calling into C.
	if err := it.GetMulti([]RenderStateImageData{RenderStateImageDataZ}, nil); err == nil {
		t.Fatal("expected error for mismatched lengths")
	}
}

func TestRenderStateImageIteratorRefill(t *testing.T) {
	term, rs, it := newImageRenderState(t)

	// Erase the placeholder cells. After the next update and refill, the
	// iterator has nothing to draw.
	term.VTWrite([]byte("\x1b[2J"))
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}
	if err := rs.ImageIterator(it); err != nil {
		t.Fatal(err)
	}
	if it.Next() {
		t.Fatal("expected no placements after erasing placeholders")
	}
}
