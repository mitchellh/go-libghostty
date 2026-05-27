package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// PointCoordinate is an untagged coordinate in the terminal grid.
//
// It is used by APIs that already define which coordinate system applies,
// such as selection gesture autoscroll viewport positions. Use [Point] when
// an API needs the coordinate system tag alongside X and Y.
// C: GhosttyPointCoordinate
type PointCoordinate struct {
	// X is the column (0-indexed).
	X uint16

	// Y is the row (0-indexed). May exceed page size for screen/history tags.
	Y uint32
}

// toC converts a Go PointCoordinate to a C GhosttyPointCoordinate.
func (p PointCoordinate) toC() C.GhosttyPointCoordinate {
	return C.GhosttyPointCoordinate{
		x: C.uint16_t(p.X),
		y: C.uint32_t(p.Y),
	}
}

// PointTag determines which coordinate system a point uses.
// C: GhosttyPointTag
type PointTag int

const (
	// PointTagActive references the active area where the cursor can move.
	PointTagActive PointTag = C.GHOSTTY_POINT_TAG_ACTIVE

	// PointTagViewport references the visible viewport (changes when scrolled).
	PointTagViewport PointTag = C.GHOSTTY_POINT_TAG_VIEWPORT

	// PointTagScreen references the full screen including scrollback.
	PointTagScreen PointTag = C.GHOSTTY_POINT_TAG_SCREEN

	// PointTagHistory references scrollback history only (before active area).
	PointTagHistory PointTag = C.GHOSTTY_POINT_TAG_HISTORY
)

// Point is a tagged position in the terminal grid. The Tag determines
// which coordinate system X and Y refer to.
// C: GhosttyPoint
type Point struct {
	// Tag determines the coordinate system.
	Tag PointTag

	// X is the column (0-indexed).
	X uint16

	// Y is the row (0-indexed). May exceed page size for screen/history tags.
	Y uint32
}

// toC converts a Go Point to a C GhosttyPoint.
func (p Point) toC() C.GhosttyPoint {
	var cp C.GhosttyPoint
	cp.tag = C.GhosttyPointTag(p.Tag)
	// Set the coordinate in the value union.
	coord := (*C.GhosttyPointCoordinate)(unsafe.Pointer(&cp.value[0]))
	coord.x = C.uint16_t(p.X)
	coord.y = C.uint32_t(p.Y)
	return cp
}

// pointFromC converts a C coordinate plus its requested tag into a Go
// Point. The C APIs that resolve refs back to points return only the
// coordinate because the caller already selected the coordinate system.
func pointFromC(tag PointTag, coord C.GhosttyPointCoordinate) Point {
	return Point{
		Tag: tag,
		X:   uint16(coord.x),
		Y:   uint32(coord.y),
	}
}
