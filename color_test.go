package libghostty

import (
	"math"
	"testing"
)

func TestColorParseAndComponents(t *testing.T) {
	color, err := ParseColor("#abc")
	if err != nil {
		t.Fatal(err)
	}
	if want := (ColorRGB{R: 0xaa, G: 0xbb, B: 0xcc}); color != want {
		t.Fatalf("expected %#v, got %#v", want, color)
	}

	r, g, b := color.Components()
	if r != color.R || g != color.G || b != color.B {
		t.Fatalf("components returned %d,%d,%d for %#v", r, g, b, color)
	}

	x11, err := ParseX11Color("ForestGreen")
	if err != nil {
		t.Fatal(err)
	}
	if want := (ColorRGB{R: 34, G: 139, B: 34}); x11 != want {
		t.Fatalf("expected %#v, got %#v", want, x11)
	}

	if _, err := ParseColor("not a real color"); err == nil {
		t.Fatal("expected invalid color to fail")
	}
}

func TestColorPaletteUtilities(t *testing.T) {
	index, color, err := ParsePaletteEntry("0x10=#282c34")
	if err != nil {
		t.Fatal(err)
	}
	if index != 16 || color != (ColorRGB{R: 0x28, G: 0x2c, B: 0x34}) {
		t.Fatalf("unexpected palette entry %d=%#v", index, color)
	}

	base := DefaultPalette()
	base[20] = ColorRGB{R: 1, G: 2, B: 3}
	var skip ColorPaletteMask
	skip.Set(20)
	if !skip.IsSet(20) {
		t.Fatal("expected palette mask bit to be set")
	}

	generated := GeneratePalette(
		&base,
		&skip,
		ColorRGB{R: 20, G: 20, B: 20},
		ColorRGB{R: 230, G: 230, B: 230},
		true,
	)
	if generated[20] != base[20] {
		t.Fatalf("expected skipped entry %#v, got %#v", base[20], generated[20])
	}

	skip.Unset(20)
	if skip.IsSet(20) {
		t.Fatal("expected palette mask bit to be unset")
	}
}

func TestColorMathAndX11Names(t *testing.T) {
	black := ColorRGB{}
	white := ColorRGB{R: 255, G: 255, B: 255}
	if got := black.Luminance(); got != 0 {
		t.Fatalf("expected black luminance 0, got %f", got)
	}
	if got := white.Luminance(); math.Abs(got-1) > 1e-12 {
		t.Fatalf("expected white luminance 1, got %f", got)
	}
	if got := black.Contrast(white); math.Abs(got-21) > 1e-12 {
		t.Fatalf("expected black/white contrast 21, got %f", got)
	}
	if black.PerceivedLuminance() >= white.PerceivedLuminance() {
		t.Fatal("expected white to have greater perceived luminance")
	}

	names := X11Colors()
	if len(names) == 0 {
		t.Fatal("expected X11 color names")
	}
	if names[0].Name == "" {
		t.Fatal("expected first X11 color name to be non-empty")
	}
}
