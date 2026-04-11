// Package png provides a ready-to-use PNG decoder for libghostty
// using Go's standard [image/png] package.
//
// This package is separate from the root libghostty package so that
// importing libghostty does not unconditionally pull in image/png
// (and its init-time image format registration). Import this package
// only when you need PNG decoding:
//
//	import (
//		libghostty "github.com/mitchellh/go-libghostty"
//		syspng "github.com/mitchellh/go-libghostty/sys/png"
//	)
//
//	libghostty.SysSetDecodePng(syspng.Decode)
package png

import (
	"bytes"
	"fmt"
	"image"
	goimg "image/png"

	libghostty "github.com/mitchellh/go-libghostty"
)

// Decode is a ready-to-use [libghostty.SysDecodePngFn] implementation
// that decodes PNG data using Go's standard [image/png] package. It
// converts any decoded image format to NRGBA (non-premultiplied alpha)
// before returning the raw pixel bytes.
//
// Usage:
//
//	libghostty.SysSetDecodePng(syspng.Decode)
func Decode(data []byte) (*libghostty.SysImage, error) {
	img, err := goimg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("png decode: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Fast path: if the image is already NRGBA we can use the pixels
	// directly without a per-pixel conversion.
	if nrgba, ok := img.(*image.NRGBA); ok {
		return &libghostty.SysImage{
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

	return &libghostty.SysImage{
		Width:  uint32(w),
		Height: uint32(h),
		Data:   dst.Pix,
	}, nil
}
