package libghostty

// Built-in implementations for system callbacks using Go standard library
// packages. These are optional convenience functions that can be passed
// directly to their corresponding SysSet* installers.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

// SysDecodePng is a ready-to-use [SysDecodePngFn] implementation that
// decodes PNG data using Go's standard [image/png] package. It converts
// any decoded image format to NRGBA (non-premultiplied alpha) before
// returning the raw pixel bytes.
//
// Usage:
//
//	libghostty.SysSetDecodePng(libghostty.SysDecodePng)
func SysDecodePng(data []byte) (*SysImage, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("png decode: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Fast path: if the image is already NRGBA we can use the pixels
	// directly without a per-pixel conversion.
	if nrgba, ok := img.(*image.NRGBA); ok {
		return &SysImage{
			Width:  uint32(w),
			Height: uint32(h),
			Data:   nrgba.Pix,
		}, nil
	}

	// Slow path: convert arbitrary image types to NRGBA.
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}

	return &SysImage{
		Width:  uint32(w),
		Height: uint32(h),
		Data:   dst.Pix,
	}, nil
}
