package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

// SurfacePosition is an x/y position in rendered surface pixel space.
//
// This is not a terminal grid coordinate. The origin is the top-left of the
// rendered terminal surface.
// C: GhosttySurfacePosition
type SurfacePosition struct {
	// X is the horizontal position in surface pixels.
	X float64

	// Y is the vertical position in surface pixels.
	Y float64
}

// toC converts a Go SurfacePosition to a C GhosttySurfacePosition.
func (p SurfacePosition) toC() C.GhosttySurfacePosition {
	return C.GhosttySurfacePosition{
		x: C.double(p.X),
		y: C.double(p.Y),
	}
}
