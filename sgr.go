package libghostty

// SGR parser bindings from sgr.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// SGRAttributeTag identifies the active value in an [SGRAttribute].
// C: GhosttySgrAttributeTag
type SGRAttributeTag int

const (
	// SGRAttrUnset is the zero-value attribute tag.
	SGRAttrUnset SGRAttributeTag = C.GHOSTTY_SGR_ATTR_UNSET

	// SGRAttrUnknown identifies an unknown or malformed parameter sequence.
	SGRAttrUnknown SGRAttributeTag = C.GHOSTTY_SGR_ATTR_UNKNOWN

	// SGRAttrBold enables bold text.
	SGRAttrBold SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BOLD

	// SGRAttrResetBold disables bold text.
	SGRAttrResetBold SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_BOLD

	// SGRAttrItalic enables italic text.
	SGRAttrItalic SGRAttributeTag = C.GHOSTTY_SGR_ATTR_ITALIC

	// SGRAttrResetItalic disables italic text.
	SGRAttrResetItalic SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_ITALIC

	// SGRAttrFaint enables faint text.
	SGRAttrFaint SGRAttributeTag = C.GHOSTTY_SGR_ATTR_FAINT

	// SGRAttrUnderline sets the underline style in Underline.
	SGRAttrUnderline SGRAttributeTag = C.GHOSTTY_SGR_ATTR_UNDERLINE

	// SGRAttrUnderlineColor sets a direct RGB underline color in Color.
	SGRAttrUnderlineColor SGRAttributeTag = C.GHOSTTY_SGR_ATTR_UNDERLINE_COLOR

	// SGRAttrUnderlineColor256 sets a palette underline color in PaletteIndex.
	SGRAttrUnderlineColor256 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_UNDERLINE_COLOR_256

	// SGRAttrResetUnderlineColor resets the underline color.
	SGRAttrResetUnderlineColor SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_UNDERLINE_COLOR

	// SGRAttrOverline enables overline.
	SGRAttrOverline SGRAttributeTag = C.GHOSTTY_SGR_ATTR_OVERLINE

	// SGRAttrResetOverline disables overline.
	SGRAttrResetOverline SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_OVERLINE

	// SGRAttrBlink enables blinking text.
	SGRAttrBlink SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BLINK

	// SGRAttrResetBlink disables blinking text.
	SGRAttrResetBlink SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_BLINK

	// SGRAttrInverse enables inverse video.
	SGRAttrInverse SGRAttributeTag = C.GHOSTTY_SGR_ATTR_INVERSE

	// SGRAttrResetInverse disables inverse video.
	SGRAttrResetInverse SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_INVERSE

	// SGRAttrInvisible enables invisible text.
	SGRAttrInvisible SGRAttributeTag = C.GHOSTTY_SGR_ATTR_INVISIBLE

	// SGRAttrResetInvisible disables invisible text.
	SGRAttrResetInvisible SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_INVISIBLE

	// SGRAttrStrikethrough enables strikethrough.
	SGRAttrStrikethrough SGRAttributeTag = C.GHOSTTY_SGR_ATTR_STRIKETHROUGH

	// SGRAttrResetStrikethrough disables strikethrough.
	SGRAttrResetStrikethrough SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_STRIKETHROUGH

	// SGRAttrDirectColorFG sets a direct RGB foreground color in Color.
	SGRAttrDirectColorFG SGRAttributeTag = C.GHOSTTY_SGR_ATTR_DIRECT_COLOR_FG

	// SGRAttrDirectColorBG sets a direct RGB background color in Color.
	SGRAttrDirectColorBG SGRAttributeTag = C.GHOSTTY_SGR_ATTR_DIRECT_COLOR_BG

	// SGRAttrBG8 sets an eight-color background in PaletteIndex.
	SGRAttrBG8 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BG_8

	// SGRAttrFG8 sets an eight-color foreground in PaletteIndex.
	SGRAttrFG8 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_FG_8

	// SGRAttrResetFG resets the foreground color.
	SGRAttrResetFG SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_FG

	// SGRAttrResetBG resets the background color.
	SGRAttrResetBG SGRAttributeTag = C.GHOSTTY_SGR_ATTR_RESET_BG

	// SGRAttrBrightBG8 sets a bright eight-color background in PaletteIndex.
	SGRAttrBrightBG8 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BRIGHT_BG_8

	// SGRAttrBrightFG8 sets a bright eight-color foreground in PaletteIndex.
	SGRAttrBrightFG8 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BRIGHT_FG_8

	// SGRAttrBG256 sets a 256-color background in PaletteIndex.
	SGRAttrBG256 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_BG_256

	// SGRAttrFG256 sets a 256-color foreground in PaletteIndex.
	SGRAttrFG256 SGRAttributeTag = C.GHOSTTY_SGR_ATTR_FG_256
)

// SGRUnderline identifies an SGR underline style.
// C: GhosttySgrUnderline
type SGRUnderline int

// SGRUnknown contains a copied unknown SGR parameter sequence.
// C: GhosttySgrUnknown
type SGRUnknown struct {
	// Full is the complete parameter list supplied to the parser.
	Full []uint16

	// Partial is the portion at which parsing encountered the unknown value.
	Partial []uint16
}

// SGRAttribute is one parsed SGR operation. Callers should inspect Tag and
// then use the corresponding field documented by that tag.
// C: GhosttySgrAttribute
type SGRAttribute struct {
	// Tag identifies the attribute operation and active value.
	Tag SGRAttributeTag

	// Unknown is populated when Tag is SGRAttrUnknown.
	Unknown SGRUnknown

	// Underline is populated when Tag is SGRAttrUnderline.
	Underline SGRUnderline

	// Color is populated for direct RGB foreground, background, or underline
	// color attributes.
	Color ColorRGB

	// PaletteIndex is populated for 8-, 16-, and 256-color attributes.
	PaletteIndex uint8
}

// SGRParser parses CSI SGR parameter lists into semantic attributes.
// C: GhosttySgrParser
type SGRParser struct {
	ptr C.GhosttySgrParser
}

// NewSGRParser creates a reusable SGR parser.
func NewSGRParser() (*SGRParser, error) {
	var ptr C.GhosttySgrParser
	if err := resultError(C.ghostty_sgr_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &SGRParser{ptr: ptr}, nil
}

// Close frees the parser.
func (p *SGRParser) Close() {
	C.ghostty_sgr_free(p.ptr)
}

// Reset restarts iteration at the beginning of the current parameter list.
func (p *SGRParser) Reset() {
	C.ghostty_sgr_reset(p.ptr)
}

// SetParams copies a CSI SGR parameter list into the parser. separators may
// be nil to treat every separator as a semicolon; otherwise it must contain
// one ';' or ':' byte for every parameter.
func (p *SGRParser) SetParams(params []uint16, separators []byte) error {
	if len(separators) != 0 && len(separators) != len(params) {
		return fmt.Errorf(
			"libghostty: SGR separators length %d does not match params length %d",
			len(separators),
			len(params),
		)
	}

	var paramsPtr *C.uint16_t
	if len(params) > 0 {
		paramsPtr = (*C.uint16_t)(unsafe.Pointer(&params[0]))
	}
	var separatorsPtr *C.char
	if len(separators) > 0 {
		separatorsPtr = (*C.char)(unsafe.Pointer(&separators[0]))
	}

	return resultError(C.ghostty_sgr_set_params(
		p.ptr,
		paramsPtr,
		separatorsPtr,
		C.size_t(len(params)),
	))
}

// Next returns the next parsed attribute. The boolean is false when the
// current parameter list is exhausted.
func (p *SGRParser) Next() (SGRAttribute, bool) {
	var raw C.GhosttySgrAttribute
	if !bool(C.ghostty_sgr_next(p.ptr, &raw)) {
		return SGRAttribute{}, false
	}

	attr := SGRAttribute{
		Tag: SGRAttributeTag(C.ghostty_sgr_attribute_tag(raw)),
	}
	value := C.ghostty_sgr_attribute_value(&raw)

	switch attr.Tag {
	case SGRAttrUnknown:
		unknown := *(*C.GhosttySgrUnknown)(unsafe.Pointer(value))
		attr.Unknown = sgrUnknownFromC(unknown)

	case SGRAttrUnderline:
		attr.Underline = SGRUnderline(
			*(*C.GhosttySgrUnderline)(unsafe.Pointer(value)),
		)

	case SGRAttrUnderlineColor, SGRAttrDirectColorFG, SGRAttrDirectColorBG:
		attr.Color = colorRGBFromC(
			*(*C.GhosttyColorRgb)(unsafe.Pointer(value)),
		)

	case SGRAttrUnderlineColor256,
		SGRAttrBG8,
		SGRAttrFG8,
		SGRAttrBrightBG8,
		SGRAttrBrightFG8,
		SGRAttrBG256,
		SGRAttrFG256:
		attr.PaletteIndex = uint8(
			*(*C.GhosttyColorPaletteIndex)(unsafe.Pointer(value)),
		)
	}

	return attr, true
}

// sgrUnknownFromC copies both borrowed parameter slices from an unknown
// attribute into Go-owned memory.
func sgrUnknownFromC(unknown C.GhosttySgrUnknown) SGRUnknown {
	var fullPtr *C.uint16_t
	fullLen := C.ghostty_sgr_unknown_full(unknown, &fullPtr)
	full := make([]uint16, int(fullLen))
	for i, value := range unsafe.Slice(fullPtr, int(fullLen)) {
		full[i] = uint16(value)
	}

	var partialPtr *C.uint16_t
	partialLen := C.ghostty_sgr_unknown_partial(unknown, &partialPtr)
	partial := make([]uint16, int(partialLen))
	for i, value := range unsafe.Slice(partialPtr, int(partialLen)) {
		partial[i] = uint16(value)
	}

	return SGRUnknown{
		Full:    full,
		Partial: partial,
	}
}
