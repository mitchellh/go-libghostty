package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

// TypeJSON returns a process-lifetime JSON description of every C API struct
// layout for the current target.
func TypeJSON() string {
	return C.GoString(C.ghostty_type_json())
}

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
