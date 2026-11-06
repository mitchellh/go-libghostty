package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

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

// SGR underline styles.
// C: GhosttySgrUnderline
const (
	// UnderlineNone disables underlining.
	UnderlineNone SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_NONE

	// UnderlineSingle draws a single underline.
	UnderlineSingle SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_SINGLE

	// UnderlineDouble draws a double underline.
	UnderlineDouble SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_DOUBLE

	// UnderlineCurly draws a curly underline.
	UnderlineCurly SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_CURLY

	// UnderlineDotted draws a dotted underline.
	UnderlineDotted SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_DOTTED

	// UnderlineDashed draws a dashed underline.
	UnderlineDashed SGRUnderline = C.GHOSTTY_SGR_UNDERLINE_DASHED
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

// DefaultStyle returns a new style initialized to Ghostty's default values.
func DefaultStyle() *Style {
	var c C.GhosttyStyle
	C.ghostty_style_default(&c)
	return &Style{c: c}
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

// hasFlag reports whether the on/off attribute flag is set in the style's
// packed flags.
func (s *Style) hasFlag(flag C.GhosttyStyleFlags) bool {
	return s.c.flags&flag != 0
}

// Bold reports whether bold is set.
func (s *Style) Bold() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_BOLD)
}

// Italic reports whether italic is set.
func (s *Style) Italic() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_ITALIC)
}

// Faint reports whether faint (dim) is set.
func (s *Style) Faint() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_FAINT)
}

// Blink reports whether blink is set.
func (s *Style) Blink() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_BLINK)
}

// Inverse reports whether inverse video is set.
func (s *Style) Inverse() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_INVERSE)
}

// Invisible reports whether invisible is set.
func (s *Style) Invisible() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_INVISIBLE)
}

// Strikethrough reports whether strikethrough is set.
func (s *Style) Strikethrough() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_STRIKETHROUGH)
}

// Overline reports whether overline is set.
func (s *Style) Overline() bool {
	return s.hasFlag(C.GHOSTTY_STYLE_FLAG_OVERLINE)
}

// Underline returns the underline style (one of the Underline* constants).
func (s *Style) Underline() SGRUnderline {
	// GHOSTTY_STYLE_UNDERLINE is a function-like macro, which cgo can't call,
	// so the masking it does is spelled out here.
	return SGRUnderline((s.c.flags & C.GHOSTTY_STYLE_UNDERLINE_MASK) >> C.GHOSTTY_STYLE_UNDERLINE_SHIFT)
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
