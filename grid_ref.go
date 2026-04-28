package libghostty

/*
#include <ghostty/vt.h>

// Helper to create a properly initialized GhosttyGridRef (sized struct).
static inline GhosttyGridRef init_grid_ref() {
	GhosttyGridRef ref = GHOSTTY_INIT_SIZED(GhosttyGridRef);
	return ref;
}
*/
import "C"

import "unsafe"

// initCGridRef returns a zero-initialized C GhosttyGridRef with its
// size field set (GHOSTTY_INIT_SIZED). Used by terminal.go to pass
// a grid ref to C APIs.
func initCGridRef() C.GhosttyGridRef {
	return C.init_grid_ref()
}

// GridRef is a resolved reference to a specific cell position in the
// terminal's internal page structure. Obtain a GridRef from
// [Terminal.GridRef], then extract cell or row data from it.
//
// A GridRef is a borrowed view into terminal internals, so callers
// must use it under the same serialized access that protects the
// owning terminal. Any later terminal operation may invalidate the
// GridRef, even if it looks unrelated, so read and cache what you
// need immediately. Values returned by its getter methods are copied
// snapshots and may be retained after the GridRef itself becomes
// invalid.
// C: GhosttyGridRef
type GridRef struct {
	ref C.GhosttyGridRef
}

// Cell returns the cell at the grid reference's position.
func (g *GridRef) Cell() (*Cell, error) {
	var cell C.GhosttyCell
	if err := resultError(C.ghostty_grid_ref_cell(&g.ref, &cell)); err != nil {
		return nil, err
	}
	return &Cell{c: cell}, nil
}

// Row returns the row at the grid reference's position.
func (g *GridRef) Row() (*Row, error) {
	var row C.GhosttyRow
	if err := resultError(C.ghostty_grid_ref_row(&g.ref, &row)); err != nil {
		return nil, err
	}
	return &Row{c: row}, nil
}

// Graphemes returns the full grapheme cluster codepoints for the cell
// at the grid reference's position. Returns nil if the cell has no text.
func (g *GridRef) Graphemes() ([]uint32, error) {
	// First call to get the required length.
	var outLen C.size_t
	err := resultError(C.ghostty_grid_ref_graphemes(&g.ref, nil, 0, &outLen))
	if err != nil {
		// OUT_OF_SPACE means we need a bigger buffer; outLen has the size.
		ge, ok := err.(*Error)
		if !ok || ge.Result != ResultOutOfSpace {
			return nil, err
		}
	}

	if outLen == 0 {
		return nil, nil
	}

	buf := make([]uint32, uint(outLen))
	if err := resultError(C.ghostty_grid_ref_graphemes(
		&g.ref,
		(*C.uint32_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&outLen,
	)); err != nil {
		return nil, err
	}

	return buf[:uint(outLen)], nil
}

// Style returns the style of the cell at the grid reference's position.
func (g *GridRef) Style() (*Style, error) {
	cs := initCStyle()
	if err := resultError(C.ghostty_grid_ref_style(&g.ref, &cs)); err != nil {
		return nil, err
	}
	return &Style{c: cs}, nil
}
