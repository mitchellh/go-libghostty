package libghostty

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"testing"
)

// testDecodePng is a minimal SysDecodePngFn for tests. It avoids
// importing the syspng subpackage (which would create an import cycle
// in internal tests) by inlining the same logic.
func testDecodePng(data []byte) (*SysImage, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("png decode: %w", err)
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if nrgba, ok := img.(*image.NRGBA); ok {
		return &SysImage{Width: uint32(w), Height: uint32(h), Data: nrgba.Pix}, nil
	}
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}
	return &SysImage{Width: uint32(w), Height: uint32(h), Data: dst.Pix}, nil
}

// newKittyTerminal creates a terminal with Kitty graphics enabled
// (PNG decode callback, WritePty handler, storage limit, and cell
// pixel dimensions), ready for Kitty graphics protocol testing.
func newKittyTerminal(t *testing.T) *Terminal {
	t.Helper()

	// Install the PNG decoder.
	if err := SysSetDecodePng(testDecodePng); err != nil {
		t.Fatal(err)
	}

	term, err := NewTerminal(
		WithSize(80, 24),
		// Install a WritePty handler so the terminal can send
		// protocol responses (required for kitty graphics).
		WithWritePty(func(data []byte) {}),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Set cell pixel dimensions (required for image placement calculations).
	if err := term.Resize(80, 24, 8, 16); err != nil {
		t.Fatal(err)
	}

	// Enable Kitty graphics with a generous storage limit.
	limit := uint64(64 * 1024 * 1024)
	if err := term.SetKittyImageStorageLimit(&limit); err != nil {
		t.Fatal(err)
	}

	return term
}

// sendKittyImage sends a 1x1 PNG image to the terminal using the Kitty
// graphics protocol. Uses the same image as the upstream C example.
// The terminal auto-assigns the image ID.
func sendKittyImage(t *testing.T, term *Terminal) {
	t.Helper()

	// Kitty graphics protocol: transmit+display, PNG format (f=100),
	// direct transmission (t=d, implicit), request response (q=1).
	// Uses the same 1x1 red PNG as the upstream C example.
	cmd := "\x1b_Ga=T,f=100,q=1;" +
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAA" +
		"DUlEQVR4nGP4z8DwHwAFAAH/iZk9HQAAAABJRU5ErkJggg==" +
		"\x1b\\"
	term.VTWrite([]byte(cmd))
}

func TestKittyGraphicsStorageLimit(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Set a storage limit.
	limit := uint64(1024 * 1024)
	if err := term.SetKittyImageStorageLimit(&limit); err != nil {
		t.Fatal(err)
	}

	// Disable by passing nil.
	if err := term.SetKittyImageStorageLimit(nil); err != nil {
		t.Fatal(err)
	}
}

func TestKittyGraphicsMediumSetters(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetKittyImageMediumFile(true); err != nil {
		t.Fatal(err)
	}
	if err := term.SetKittyImageMediumFile(false); err != nil {
		t.Fatal(err)
	}
	if err := term.SetKittyImageMediumTempFile(true); err != nil {
		t.Fatal(err)
	}
	if err := term.SetKittyImageMediumSharedMem(true); err != nil {
		t.Fatal(err)
	}
}

func TestKittyGraphicsHandle(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}
	if kg == nil {
		t.Fatal("expected non-nil KittyGraphics handle")
	}
}

func TestKittyGraphicsPlacementIteratorEmpty(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	// No images sent; iterator should be empty.
	if iter.Next() {
		t.Fatal("expected no placements in empty terminal")
	}
}

func TestKittyGraphicsImageLookupMiss(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	// No image with ID 999 should exist.
	img := kg.Image(999)
	if img != nil {
		t.Fatal("expected nil for non-existent image ID")
	}
}

func TestKittyGraphicsImageSendAndLookup(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	// Find the image ID by iterating placements (terminal auto-assigns IDs).
	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement after sending image")
	}

	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}

	// Look up the image by its ID.
	img := kg.Image(imageID)
	if img == nil {
		t.Fatal("expected non-nil image for placement's image ID")
	}

	// Verify image properties.
	id, err := img.ID()
	if err != nil {
		t.Fatal(err)
	}
	if id != imageID {
		t.Fatalf("expected image ID %d, got %d", imageID, id)
	}

	w, err := img.Width()
	if err != nil {
		t.Fatal(err)
	}
	h, err := img.Height()
	if err != nil {
		t.Fatal(err)
	}
	if w != 1 || h != 1 {
		t.Fatalf("expected 1x1 image, got %dx%d", w, h)
	}

	// Check format — after PNG decoding it should be RGBA.
	format, err := img.Format()
	if err != nil {
		t.Fatal(err)
	}
	if format != KittyImageFormatRGBA {
		t.Fatalf("expected RGBA format, got %d", format)
	}

	// Check compression.
	compression, err := img.Compression()
	if err != nil {
		t.Fatal(err)
	}
	if compression != KittyImageCompressionNone {
		t.Fatalf("expected no compression after decode, got %d", compression)
	}

	// Check data is accessible.
	data, err := img.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty pixel data")
	}
	// 1x1 RGBA = 4 bytes.
	if len(data) != 4 {
		t.Fatalf("expected 4 bytes of pixel data, got %d", len(data))
	}
}

func TestKittyGraphicsPlacementIteration(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	// Verify we can read placement fields.
	_, err = iter.PlacementID()
	if err != nil {
		t.Fatal(err)
	}

	isVirtual, err := iter.IsVirtual()
	if err != nil {
		t.Fatal(err)
	}
	if isVirtual {
		t.Fatal("expected non-virtual placement for direct display")
	}

	_, err = iter.Z()
	if err != nil {
		t.Fatal(err)
	}

	// Look up the image for rendering helpers.
	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}
	img := kg.Image(imageID)
	if img == nil {
		t.Fatal("expected image lookup to succeed")
	}

	// PixelSize should return valid dimensions.
	pw, ph, err := iter.PixelSize(img, term)
	if err != nil {
		t.Fatal(err)
	}
	if pw == 0 || ph == 0 {
		t.Fatalf("expected non-zero pixel size, got %dx%d", pw, ph)
	}

	// GridSize should return valid dimensions.
	gc, gr, err := iter.GridSize(img, term)
	if err != nil {
		t.Fatal(err)
	}
	if gc == 0 || gr == 0 {
		t.Fatalf("expected non-zero grid size, got %dx%d", gc, gr)
	}

	// SourceRect should succeed.
	_, _, sw, sh, err := iter.SourceRect(img)
	if err != nil {
		t.Fatal(err)
	}
	if sw == 0 || sh == 0 {
		t.Fatalf("expected non-zero source rect size, got %dx%d", sw, sh)
	}
}

func TestKittyGraphicsImageInfo(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}

	img := kg.Image(imageID)
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	info, err := img.Info()
	if err != nil {
		t.Fatal(err)
	}

	if info.ID != imageID {
		t.Fatalf("expected image ID %d, got %d", imageID, info.ID)
	}
	if info.Width != 1 || info.Height != 1 {
		t.Fatalf("expected 1x1 image, got %dx%d", info.Width, info.Height)
	}
	if info.Format != KittyImageFormatRGBA {
		t.Fatalf("expected RGBA format, got %d", info.Format)
	}
	if info.Compression != KittyImageCompressionNone {
		t.Fatalf("expected no compression, got %d", info.Compression)
	}
	if len(info.Data) != 4 {
		t.Fatalf("expected 4 bytes of pixel data, got %d", len(info.Data))
	}
}

func TestKittyGraphicsPlacementInfo(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	info, err := iter.Info()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the info matches individual getters.
	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}
	if info.ImageID != imageID {
		t.Fatalf("expected image ID %d, got %d", imageID, info.ImageID)
	}

	isVirtual, err := iter.IsVirtual()
	if err != nil {
		t.Fatal(err)
	}
	if info.IsVirtual != isVirtual {
		t.Fatalf("expected IsVirtual=%v, got %v", isVirtual, info.IsVirtual)
	}

	z, err := iter.Z()
	if err != nil {
		t.Fatal(err)
	}
	if info.Z != z {
		t.Fatalf("expected Z=%d, got %d", z, info.Z)
	}
}

func TestKittyGraphicsPlacementRenderInfo(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}
	img := kg.Image(imageID)
	if img == nil {
		t.Fatal("expected image lookup to succeed")
	}

	ri, err := iter.RenderInfo(img, term)
	if err != nil {
		t.Fatal(err)
	}

	// Verify render info matches individual calls.
	pw, ph, err := iter.PixelSize(img, term)
	if err != nil {
		t.Fatal(err)
	}
	if ri.PixelWidth != pw || ri.PixelHeight != ph {
		t.Fatalf("pixel size mismatch: RenderInfo=%dx%d, PixelSize=%dx%d",
			ri.PixelWidth, ri.PixelHeight, pw, ph)
	}

	gc, gr, err := iter.GridSize(img, term)
	if err != nil {
		t.Fatal(err)
	}
	if ri.GridCols != gc || ri.GridRows != gr {
		t.Fatalf("grid size mismatch: RenderInfo=%dx%d, GridSize=%dx%d",
			ri.GridCols, ri.GridRows, gc, gr)
	}

	_, _, sw, sh, err := iter.SourceRect(img)
	if err != nil {
		t.Fatal(err)
	}
	if ri.SourceWidth != sw || ri.SourceHeight != sh {
		t.Fatalf("source rect size mismatch: RenderInfo=%dx%d, SourceRect=%dx%d",
			ri.SourceWidth, ri.SourceHeight, sw, sh)
	}

	// A freshly placed image should be viewport-visible.
	if !ri.ViewportVisible {
		t.Fatal("expected placement to be viewport-visible")
	}
}

func TestKittyGraphicsPlacementLayerFilter(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	// Set layer filter to ABOVE_TEXT (default z=0 should match).
	if err := iter.SetLayer(KittyPlacementLayerAboveText); err != nil {
		t.Fatal(err)
	}

	// Should still find the placement (z=0 is above text).
	if !iter.Next() {
		t.Fatal("expected placement with ABOVE_TEXT layer filter")
	}
}

func TestKittyGraphicsPlacementViewportPos(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()

	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}

	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}

	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	imageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}
	img := kg.Image(imageID)
	if img == nil {
		t.Fatal("expected image lookup to succeed")
	}

	// The image was just placed, so it should be visible in the viewport.
	col, row, err := iter.ViewportPos(img, term)
	if err != nil {
		t.Fatal(err)
	}
	// Position should be non-negative for a freshly placed image.
	if col < 0 || row < 0 {
		t.Fatalf("expected non-negative viewport position, got col=%d row=%d", col, row)
	}
}
