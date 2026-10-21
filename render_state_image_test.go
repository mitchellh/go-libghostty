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

// This example reads the image list only on frames where it changed. Most
// frames only change text, so a renderer can keep drawing the images it
// read last time.
func ExampleRenderState_ImagePlacementsGeneration() {
	term, err := NewTerminal(
		WithSize(10, 5),
		WithWritePty(func(*Terminal, []byte) {}),
	)
	if err != nil {
		panic(err)
	}
	defer term.Close()
	if err := term.Resize(10, 5, 10, 10); err != nil {
		panic(err)
	}
	limit := uint64(1 << 20)
	if err := term.SetKittyImageStorageLimit(&limit); err != nil {
		panic(err)
	}

	rs, err := NewRenderState()
	if err != nil {
		panic(err)
	}
	defer rs.Close()

	images, err := NewRenderStateImageIterator()
	if err != nil {
		panic(err)
	}
	defer images.Close()

	// The renderer keeps these between frames. The queue holds the images
	// to draw, and is reused so that reading the list doesn't allocate.
	var (
		queue      []RenderStateImageInfo
		lastGen    uint64
		haveImages bool
	)

	// frame updates the render state and reads the image list only if it
	// changed since the last read.
	frame := func(name string) {
		if err := rs.Update(term); err != nil {
			panic(err)
		}
		gen, err := rs.ImagePlacementsGeneration()
		if err != nil {
			panic(err)
		}
		if haveImages && gen == lastGen {
			fmt.Printf("%s: reuse %d images\n", name, len(queue))
			return
		}

		if err := rs.ImageIterator(images); err != nil {
			panic(err)
		}
		queue = queue[:0]
		for images.Next() {
			queue = append(queue, RenderStateImageInfo{})
			if err := images.InfoInto(&queue[len(queue)-1]); err != nil {
				panic(err)
			}
		}
		lastGen, haveImages = gen, true
		fmt.Printf("%s: read %d images\n", name, len(queue))
	}

	frame("empty screen")

	// Place a 1x1 pixel image at the second row. C=1 keeps the cursor
	// where it is.
	term.VTWrite([]byte("\x1b[2;1H\x1b_Ga=T,t=d,f=24,i=1,s=1,v=1,c=1,r=1,C=1,q=2;////\x1b\\"))
	frame("image placed")

	// Text on another row doesn't change any image.
	term.VTWrite([]byte("\x1b[4;1Hhello"))
	frame("text written")

	// Scrolling moves the image.
	term.VTWrite([]byte("\x1b[5;1H\n"))
	frame("screen scrolled")

	// Output:
	// empty screen: read 0 images
	// image placed: read 1 images
	// text written: reuse 1 images
	// screen scrolled: read 1 images
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

func TestRenderStateImageIteratorInfoInto(t *testing.T) {
	_, _, it := newImageRenderState(t)

	// Before Next, there is no current image and dst is left unchanged.
	dst := RenderStateImageInfo{ImageID: 42}
	if err := it.InfoInto(&dst); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
	if dst.ImageID != 42 {
		t.Fatalf("expected dst unchanged, got %+v", dst)
	}

	if !it.Next() {
		t.Fatal("expected a placement")
	}
	if err := it.InfoInto(nil); err == nil {
		t.Fatal("expected error for nil dst")
	}

	// InfoInto overwrites every field with the same values Info returns.
	if err := it.InfoInto(&dst); err != nil {
		t.Fatal(err)
	}
	info, err := it.Info()
	if err != nil {
		t.Fatal(err)
	}
	if dst != *info {
		t.Fatalf("InfoInto %+v differs from Info %+v", dst, *info)
	}

	// Reading into a caller-owned struct doesn't allocate.
	if allocs := testing.AllocsPerRun(100, func() {
		if err := it.InfoInto(&dst); err != nil {
			t.Fatal(err)
		}
	}); allocs != 0 {
		t.Fatalf("expected no allocations, got %v", allocs)
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

// newFixedImageRenderState returns a 10x5 terminal with 10x10 pixel cells
// and Kitty graphics enabled, plus a render state and image iterator for
// it. Nothing is drawn yet.
func newFixedImageRenderState(t *testing.T) (*Terminal, *RenderState, *RenderStateImageIterator) {
	t.Helper()

	term := newKittyTerminal(t)
	t.Cleanup(term.Close)
	if err := term.Resize(10, 5, 10, 10); err != nil {
		t.Fatal(err)
	}

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rs.Close)

	it, err := NewRenderStateImageIterator()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(it.Close)

	return term, rs, it
}

// renderStateImages updates rs from term and returns every image in the
// refilled iterator.
func renderStateImages(t *testing.T, term *Terminal, rs *RenderState, it *RenderStateImageIterator) []*RenderStateImageInfo {
	t.Helper()
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}
	if err := rs.ImageIterator(it); err != nil {
		t.Fatal(err)
	}
	var infos []*RenderStateImageInfo
	for it.Next() {
		info, err := it.Info()
		if err != nil {
			t.Fatal(err)
		}
		infos = append(infos, info)
	}
	return infos
}

func TestRenderStateImageIteratorKittyFixed(t *testing.T) {
	term, rs, it := newFixedImageRenderState(t)

	// Place a 1x1 RGB image as placement 5, two columns by one row, at
	// row 2, column 3. C=1 keeps the cursor where it is.
	term.VTWrite([]byte("\x1b[3;4H\x1b_Ga=T,t=d,f=24,i=2,p=5,s=1,v=1,c=2,r=1,C=1,q=2;////\x1b\\"))

	infos := renderStateImages(t, term, rs, it)
	if len(infos) != 1 {
		t.Fatalf("expected one placement, got %d", len(infos))
	}
	info := infos[0]
	if info.KittyKind != RenderStateKittyPlacementKindFixed {
		t.Fatalf("expected fixed kind, got %d", info.KittyKind)
	}
	if info.ImageID != 2 || info.KittyPlacementID != 5 {
		t.Fatalf("expected image 2 placement 5, got %d %d", info.ImageID, info.KittyPlacementID)
	}
	if info.ViewportCol != 3 || info.ViewportRow != 2 {
		t.Fatalf("expected position (3, 2), got (%d, %d)", info.ViewportCol, info.ViewportRow)
	}
	if info.GridCols != 2 || info.GridRows != 1 {
		t.Fatalf("expected grid 2x1, got %dx%d", info.GridCols, info.GridRows)
	}

	// A z-index of 0 draws in front of the text.
	if info.Z != 0 || info.Layer != RenderStateImageLayerAboveText {
		t.Fatalf("expected z 0 above text, got %d %d", info.Z, info.Layer)
	}

	// The render state gives the same geometry as the storage does.
	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}
	pi, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer pi.Close()
	if err := kg.PlacementIterator(pi); err != nil {
		t.Fatal(err)
	}
	if !pi.Next() {
		t.Fatal("expected a storage placement")
	}
	ri, err := pi.RenderInfo(kg.Image(2), term)
	if err != nil {
		t.Fatal(err)
	}
	if !ri.ViewportVisible ||
		ri.ViewportCol != info.ViewportCol || ri.ViewportRow != info.ViewportRow ||
		ri.PixelWidth != info.PixelWidth || ri.PixelHeight != info.PixelHeight ||
		ri.SourceWidth != info.SourceWidth || ri.SourceHeight != info.SourceHeight {
		t.Fatalf("storage %+v disagrees with render state %+v", ri, info)
	}
}

func TestRenderStateImageIteratorKittyRelativeToFixed(t *testing.T) {
	term, rs, it := newFixedImageRenderState(t)

	// A fixed parent at row 1, column 2, and a child placed 3 columns
	// right of and 2 rows below it.
	term.VTWrite([]byte("\x1b[2;3H\x1b_Ga=T,t=d,f=24,i=1,p=1,s=1,v=1,c=1,r=1,C=1,q=2;////\x1b\\"))
	term.VTWrite([]byte("\x1b_Ga=T,t=d,f=24,i=2,p=1,s=1,v=1,c=1,r=1,P=1,Q=1,H=3,V=2,C=1,q=2;////\x1b\\"))

	infos := renderStateImages(t, term, rs, it)
	if len(infos) != 2 {
		t.Fatalf("expected two placements, got %d", len(infos))
	}
	var child *RenderStateImageInfo
	for _, info := range infos {
		if info.ImageID == 2 {
			child = info
		}
	}
	if child == nil {
		t.Fatal("expected the child placement")
	}
	if child.KittyKind != RenderStateKittyPlacementKindRelative {
		t.Fatalf("expected relative kind, got %d", child.KittyKind)
	}
	if child.ViewportCol != 5 || child.ViewportRow != 3 {
		t.Fatalf("expected position (5, 3), got (%d, %d)", child.ViewportCol, child.ViewportRow)
	}
}

func TestRenderStateImageIteratorKittyFixedPartlyVisible(t *testing.T) {
	term, rs, it := newFixedImageRenderState(t)

	// Place an image 4 rows tall at the top row, then scroll the screen up
	// by 2 rows. The image is listed at its full size with its top 2 rows
	// above the viewport.
	term.VTWrite([]byte("\x1b[1;1H\x1b_Ga=T,t=d,f=24,i=1,s=1,v=1,c=1,r=4,C=1,q=2;////\x1b\\"))
	term.VTWrite([]byte("\x1b[5;1H\n\n"))

	infos := renderStateImages(t, term, rs, it)
	if len(infos) != 1 {
		t.Fatalf("expected one placement, got %d", len(infos))
	}
	if infos[0].ViewportRow != -2 || infos[0].GridRows != 4 {
		t.Fatalf("expected row -2 covering 4 rows, got %d %d", infos[0].ViewportRow, infos[0].GridRows)
	}

	// Once it scrolls fully out of the viewport, it is no longer listed.
	term.VTWrite([]byte("\n\n"))
	if infos := renderStateImages(t, term, rs, it); len(infos) != 0 {
		t.Fatalf("expected no placements, got %d", len(infos))
	}
}

func TestRenderStateImagePlacementsGeneration(t *testing.T) {
	term, rs, _ := newFixedImageRenderState(t)

	generation := func() uint64 {
		t.Helper()
		if err := rs.Update(term); err != nil {
			t.Fatal(err)
		}
		gen, err := rs.ImagePlacementsGeneration()
		if err != nil {
			t.Fatal(err)
		}
		return gen
	}

	// Adding an image changes the generation.
	empty := generation()
	term.VTWrite([]byte("\x1b[2;1H\x1b_Ga=T,t=d,f=24,i=1,s=1,v=1,c=1,r=1,C=1,q=2;////\x1b\\"))
	placed := generation()
	if placed == empty {
		t.Fatal("expected a new generation after placing an image")
	}

	// Writing text on another row and updating again doesn't.
	term.VTWrite([]byte("\x1b[4;1Hhello"))
	if gen := generation(); gen != placed {
		t.Fatalf("expected generation %d after writing text, got %d", placed, gen)
	}

	// Scrolling moves the image, which does.
	term.VTWrite([]byte("\x1b[5;1H\n"))
	if gen := generation(); gen == placed {
		t.Fatal("expected a new generation after scrolling")
	}

	// The generation agrees with GetMulti.
	var gen uint64
	if err := rs.GetMulti(
		[]RenderStateData{RenderStateDataImagePlacementsGeneration},
		[]unsafe.Pointer{unsafe.Pointer(&gen)},
	); err != nil {
		t.Fatal(err)
	}
	if want, _ := rs.ImagePlacementsGeneration(); gen != want {
		t.Fatalf("GetMulti returned %d, want %d", gen, want)
	}
}

func TestRenderStateImagePlacementsGenerationDisabled(t *testing.T) {
	term, err := NewTerminal(WithSize(10, 5))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	// Without Kitty graphics, the generation is always 0.
	gen, err := rs.ImagePlacementsGeneration()
	if err != nil {
		t.Fatal(err)
	}
	if gen != 0 {
		t.Fatalf("expected 0, got %d", gen)
	}

	// Renderers read the generation every frame, so it must not allocate.
	if allocs := testing.AllocsPerRun(100, func() {
		if _, err := rs.ImagePlacementsGeneration(); err != nil {
			t.Fatal(err)
		}
	}); allocs != 0 {
		t.Fatalf("expected no allocations, got %v", allocs)
	}
}
