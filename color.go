package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// ColorRGB represents an RGB color value.
// C: GhosttyColorRgb
type ColorRGB struct {
	// R is the red component.
	R uint8

	// G is the green component.
	G uint8

	// B is the blue component.
	B uint8
}

// PaletteSize is the number of entries in a terminal color palette.
const PaletteSize = 256

// Palette is a 256-color palette.
type Palette [PaletteSize]ColorRGB

// ColorPaletteMask identifies palette entries that palette generation should
// preserve from its base palette.
// C: GhosttyColorPaletteMask
type ColorPaletteMask struct {
	// Bits stores the 256-bit mask as four consecutive 64-bit words.
	Bits [4]uint64
}

// Set marks index as preserved during palette generation.
func (m *ColorPaletteMask) Set(index uint8) {
	m.Bits[index>>6] |= uint64(1) << (index & 63)
}

// Unset clears index from the set preserved during palette generation.
func (m *ColorPaletteMask) Unset(index uint8) {
	m.Bits[index>>6] &^= uint64(1) << (index & 63)
}

// IsSet reports whether index is preserved during palette generation.
func (m *ColorPaletteMask) IsSet(index uint8) bool {
	return m.Bits[index>>6]&(uint64(1)<<(index&63)) != 0
}

// X11Color is one entry in Ghostty's embedded X11 rgb.txt table.
// C: GhosttyColorX11Entry
type X11Color struct {
	// Name is the exact spelling stored in the X11 color table.
	Name string

	// Color is the RGB value associated with Name.
	Color ColorRGB
}

// Named color palette indices.
// C: GHOSTTY_COLOR_NAMED_*
const (
	ColorNamedBlack         = C.GHOSTTY_COLOR_NAMED_BLACK
	ColorNamedRed           = C.GHOSTTY_COLOR_NAMED_RED
	ColorNamedGreen         = C.GHOSTTY_COLOR_NAMED_GREEN
	ColorNamedYellow        = C.GHOSTTY_COLOR_NAMED_YELLOW
	ColorNamedBlue          = C.GHOSTTY_COLOR_NAMED_BLUE
	ColorNamedMagenta       = C.GHOSTTY_COLOR_NAMED_MAGENTA
	ColorNamedCyan          = C.GHOSTTY_COLOR_NAMED_CYAN
	ColorNamedWhite         = C.GHOSTTY_COLOR_NAMED_WHITE
	ColorNamedBrightBlack   = C.GHOSTTY_COLOR_NAMED_BRIGHT_BLACK
	ColorNamedBrightRed     = C.GHOSTTY_COLOR_NAMED_BRIGHT_RED
	ColorNamedBrightGreen   = C.GHOSTTY_COLOR_NAMED_BRIGHT_GREEN
	ColorNamedBrightYellow  = C.GHOSTTY_COLOR_NAMED_BRIGHT_YELLOW
	ColorNamedBrightBlue    = C.GHOSTTY_COLOR_NAMED_BRIGHT_BLUE
	ColorNamedBrightMagenta = C.GHOSTTY_COLOR_NAMED_BRIGHT_MAGENTA
	ColorNamedBrightCyan    = C.GHOSTTY_COLOR_NAMED_BRIGHT_CYAN
	ColorNamedBrightWhite   = C.GHOSTTY_COLOR_NAMED_BRIGHT_WHITE
)

// Components returns the red, green, and blue components through
// ghostty_color_rgb_get. The fields on ColorRGB expose the same values
// directly; this method is useful for consumers mirroring the C API.
func (c ColorRGB) Components() (r, g, b uint8) {
	cc := c.toC()
	var cr, cg, cb C.uint8_t
	C.ghostty_color_rgb_get(&cc, &cr, &cg, &cb)
	return uint8(cr), uint8(cg), uint8(cb)
}

// ParseColor parses a color using Ghostty's config and theme syntax.
func ParseColor(value string) (ColorRGB, error) {
	var out C.GhosttyColorRgb
	if err := resultError(C.ghostty_color_parse(
		(*C.char)(unsafe.Pointer(unsafe.StringData(value))),
		C.size_t(len(value)),
		&out,
	)); err != nil {
		return ColorRGB{}, err
	}
	return colorRGBFromC(out), nil
}

// ParseX11Color parses an X11 color name using Ghostty's embedded rgb.txt
// table. Matching is ASCII case-insensitive.
func ParseX11Color(name string) (ColorRGB, error) {
	var out C.GhosttyColorRgb
	if err := resultError(C.ghostty_color_parse_x11(
		(*C.char)(unsafe.Pointer(unsafe.StringData(name))),
		C.size_t(len(name)),
		&out,
	)); err != nil {
		return ColorRGB{}, err
	}
	return colorRGBFromC(out), nil
}

// ParsePaletteEntry parses a Ghostty palette override in INDEX=COLOR form.
func ParsePaletteEntry(value string) (uint8, ColorRGB, error) {
	var index C.uint8_t
	var color C.GhosttyColorRgb
	if err := resultError(C.ghostty_color_parse_palette_entry(
		(*C.char)(unsafe.Pointer(unsafe.StringData(value))),
		C.size_t(len(value)),
		&index,
		&color,
	)); err != nil {
		return 0, ColorRGB{}, err
	}
	return uint8(index), colorRGBFromC(color), nil
}

// DefaultPalette returns Ghostty's built-in 256-color palette.
func DefaultPalette() Palette {
	var out [PaletteSize]C.GhosttyColorRgb
	C.ghostty_color_palette_default(&out[0])
	return paletteFromC(&out)
}

// GeneratePalette derives the 216-color cube and grayscale ramp from a base
// palette, background, and foreground. A nil base uses Ghostty's default
// palette. A nil skip mask preserves no additional indices.
func GeneratePalette(
	base *Palette,
	skip *ColorPaletteMask,
	background ColorRGB,
	foreground ColorRGB,
	harmonious bool,
) Palette {
	var cbase [PaletteSize]C.GhosttyColorRgb
	var basePtr *C.GhosttyColorRgb
	if base != nil {
		paletteToC(base, &cbase)
		basePtr = &cbase[0]
	}

	var cskip C.GhosttyColorPaletteMask
	var skipPtr *C.GhosttyColorPaletteMask
	if skip != nil {
		for i, bits := range skip.Bits {
			cskip.bits[i] = C.uint64_t(bits)
		}
		skipPtr = &cskip
	}

	bg := background.toC()
	fg := foreground.toC()
	var out [PaletteSize]C.GhosttyColorRgb
	C.ghostty_color_palette_generate(
		basePtr,
		skipPtr,
		&bg,
		&fg,
		C.bool(harmonious),
		&out[0],
	)
	return paletteFromC(&out)
}

// Luminance returns the W3C relative luminance in the range 0 to 1.
func (c ColorRGB) Luminance() float64 {
	cc := c.toC()
	return float64(C.ghostty_color_luminance(&cc))
}

// PerceivedLuminance returns Ghostty's perceived luminance in the range 0
// to 1. Ghostty treats backgrounds above 0.5 as light.
func (c ColorRGB) PerceivedLuminance() float64 {
	cc := c.toC()
	return float64(C.ghostty_color_perceived_luminance(&cc))
}

// Contrast returns the WCAG contrast ratio between c and other.
func (c ColorRGB) Contrast(other ColorRGB) float64 {
	a := c.toC()
	b := other.toC()
	return float64(C.ghostty_color_contrast(&a, &b))
}

// X11Colors returns a Go-owned copy of Ghostty's X11 color name table in
// rgb.txt order.
func X11Colors() []X11Color {
	count := int(C.ghostty_color_x11_name_count())
	entries := unsafe.Slice(C.ghostty_color_x11_names(), count)
	result := make([]X11Color, count)
	for i, entry := range entries {
		result[i] = X11Color{
			Name:  C.GoString(entry.name),
			Color: colorRGBFromC(entry.color),
		}
	}
	return result
}

// toC converts a Go RGB value to GhosttyColorRgb.
func (c ColorRGB) toC() C.GhosttyColorRgb {
	return C.GhosttyColorRgb{
		r: C.uint8_t(c.R),
		g: C.uint8_t(c.G),
		b: C.uint8_t(c.B),
	}
}

// colorRGBFromC converts GhosttyColorRgb to its Go representation.
func colorRGBFromC(c C.GhosttyColorRgb) ColorRGB {
	return ColorRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}
}

// paletteToC copies a Go palette into C-compatible storage.
func paletteToC(src *Palette, dst *[PaletteSize]C.GhosttyColorRgb) {
	for i, color := range src {
		dst[i] = color.toC()
	}
}

// paletteFromC copies C-compatible palette storage into a Go palette.
func paletteFromC(src *[PaletteSize]C.GhosttyColorRgb) Palette {
	var result Palette
	for i, color := range src {
		result[i] = colorRGBFromC(color)
	}
	return result
}
