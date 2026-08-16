package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

// TypeJSON returns the process-lifetime, versioned libghostty-vt C type
// manifest for the current target. The manifest describes all public types,
// including layouts, enum values, union fields, and packed bit layouts.
// Consumers should reject unknown schema versions and use this metadata rather
// than hardcoding ABI details such as [Cell] bit positions.
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
