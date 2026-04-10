package libghostty

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestSysDecodePng(t *testing.T) {
	// Encode a small 2x2 NRGBA PNG in-memory.
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(0, 1, color.NRGBA{B: 255, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	img, err := SysDecodePng(buf.Bytes())
	if err != nil {
		t.Fatalf("SysDecodePng: %v", err)
	}

	if img.Width != 2 || img.Height != 2 {
		t.Fatalf("dimensions = %dx%d, want 2x2", img.Width, img.Height)
	}

	// 2x2 NRGBA = 16 bytes (4 bytes per pixel).
	if len(img.Data) != 16 {
		t.Fatalf("len(Data) = %d, want 16", len(img.Data))
	}

	// Spot-check top-left pixel: red, fully opaque.
	if img.Data[0] != 255 || img.Data[1] != 0 || img.Data[2] != 0 || img.Data[3] != 255 {
		t.Errorf("pixel(0,0) = %v, want [255 0 0 255]", img.Data[0:4])
	}
}

func TestSysDecodePngInvalid(t *testing.T) {
	_, err := SysDecodePng([]byte("not a png"))
	if err == nil {
		t.Fatal("SysDecodePng(invalid) = nil error, want error")
	}
}

func TestSysDecodePngRGBA(t *testing.T) {
	// Use an RGBA image (premultiplied alpha) to exercise the slow path
	// conversion to NRGBA.
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 128, G: 0, B: 0, A: 128})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	img, err := SysDecodePng(buf.Bytes())
	if err != nil {
		t.Fatalf("SysDecodePng: %v", err)
	}

	if img.Width != 1 || img.Height != 1 {
		t.Fatalf("dimensions = %dx%d, want 1x1", img.Width, img.Height)
	}

	// The premultiplied (128,0,0,128) should convert to non-premultiplied
	// (255,0,0,128).
	if img.Data[0] != 255 || img.Data[1] != 0 || img.Data[2] != 0 || img.Data[3] != 128 {
		t.Errorf("pixel(0,0) = %v, want [255 0 0 128]", img.Data[0:4])
	}
}
