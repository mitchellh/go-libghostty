package libghostty

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyStyle (sized struct).
static inline GhosttyStyle init_style() {
	GhosttyStyle style = GHOSTTY_INIT_SIZED(GhosttyStyle);
	return style;
}
*/
import "C"

import "unsafe"

// initCStyle returns a zero-initialized C GhosttyStyle with its size
// field set (GHOSTTY_INIT_SIZED). Used by other files that need to
// pass a style to C APIs.
func initCStyle() C.GhosttyStyle {
	return C.init_style()
}

// StyleColorTag identifies the type of color in a style color.
// C: GhosttyStyleColorTag
type StyleColorTag int

const (
	// StyleColorNone means no color is set.
	StyleColorNone StyleColorTag = C.GHOSTTY_STYLE_COLOR_NONE

	// StyleColorPalette means the color is a palette index.
	StyleColorPalette StyleColorTag = C.GHOSTTY_STYLE_COLOR_PALETTE

	// StyleColorRGB means the color is a direct RGB value.
	StyleColorRGB StyleColorTag = C.GHOSTTY_STYLE_COLOR_RGB
)

// StyleColor is a tagged union representing a color in a style attribute.
// Check Tag to determine which field is valid.
// C: GhosttyStyleColor
type StyleColor struct {
	// Tag identifies the type of color.
	Tag StyleColorTag

	// Palette is the palette index (valid when Tag is StyleColorPalette).
	Palette uint8

	// RGB is the direct RGB color (valid when Tag is StyleColorRGB).
	RGB ColorRGB
}

// Underline style constants.
// C: GhosttySgrUnderline
const (
	UnderlineNone   = C.GHOSTTY_SGR_UNDERLINE_NONE
	UnderlineSingle = C.GHOSTTY_SGR_UNDERLINE_SINGLE
	UnderlineDouble = C.GHOSTTY_SGR_UNDERLINE_DOUBLE
	UnderlineCurly  = C.GHOSTTY_SGR_UNDERLINE_CURLY
	UnderlineDotted = C.GHOSTTY_SGR_UNDERLINE_DOTTED
	UnderlineDashed = C.GHOSTTY_SGR_UNDERLINE_DASHED
)

// Style is a thin wrapper around the copied C GhosttyStyle value. It
// provides getter methods to access individual style attributes
// without copying the entire struct upfront. A Style is a value
// snapshot and may be retained after the terminal, [GridRef], or
// render-state iterator that produced it becomes invalid.
// C: GhosttyStyle
type Style struct {
	c C.GhosttyStyle
}

// IsDefault reports whether the style is the default style
// (no colors, no flags).
func (s *Style) IsDefault() bool {
	return bool(C.ghostty_style_is_default(&s.c))
}

// FgColor returns the foreground color.
func (s *Style) FgColor() StyleColor {
	return styleColorFromC(s.c.fg_color)
}

// BgColor returns the background color.
func (s *Style) BgColor() StyleColor {
	return styleColorFromC(s.c.bg_color)
}

// UnderlineColor returns the underline color.
func (s *Style) UnderlineColor() StyleColor {
	return styleColorFromC(s.c.underline_color)
}

// Bold reports whether bold is set.
func (s *Style) Bold() bool {
	return bool(s.c.bold)
}

// Italic reports whether italic is set.
func (s *Style) Italic() bool {
	return bool(s.c.italic)
}

// Faint reports whether faint (dim) is set.
func (s *Style) Faint() bool {
	return bool(s.c.faint)
}

// Blink reports whether blink is set.
func (s *Style) Blink() bool {
	return bool(s.c.blink)
}

// Inverse reports whether inverse video is set.
func (s *Style) Inverse() bool {
	return bool(s.c.inverse)
}

// Invisible reports whether invisible is set.
func (s *Style) Invisible() bool {
	return bool(s.c.invisible)
}

// Strikethrough reports whether strikethrough is set.
func (s *Style) Strikethrough() bool {
	return bool(s.c.strikethrough)
}

// Overline reports whether overline is set.
func (s *Style) Overline() bool {
	return bool(s.c.overline)
}

// Underline returns the underline style (one of the Underline* constants).
func (s *Style) Underline() int {
	return int(s.c.underline)
}

// styleColorFromC converts a C GhosttyStyleColor to a Go StyleColor.
func styleColorFromC(c C.GhosttyStyleColor) StyleColor {
	sc := StyleColor{Tag: StyleColorTag(c.tag)}
	switch sc.Tag {
	case StyleColorPalette:
		sc.Palette = uint8(*(*C.GhosttyColorPaletteIndex)(unsafe.Pointer(&c.value[0])))
	case StyleColorRGB:
		rgb := *(*C.GhosttyColorRgb)(unsafe.Pointer(&c.value[0]))
		sc.RGB = ColorRGB{R: uint8(rgb.r), G: uint8(rgb.g), B: uint8(rgb.b)}
	}
	return sc
}
