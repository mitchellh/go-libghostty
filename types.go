package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// ghosttySizeToInt converts a C size to a Go slice length without allowing
// an overflowing conversion to produce an invalid unsafe.Slice length.
func ghosttySizeToInt(size C.size_t) (int, bool) {
	if uint64(size) > uint64(^uint(0)>>1) {
		return 0, false
	}
	return int(size), true
}

// copyGhosttyString copies a borrowed, binary-safe GhosttyString into Go
// memory. For zero-length strings the pointer is intentionally ignored
// because libghostty does not require it to be valid.
func copyGhosttyString(value C.GhosttyString) ([]byte, bool) {
	length, ok := ghosttySizeToInt(value.len)
	if !ok {
		return nil, false
	}
	if length == 0 {
		return []byte{}, true
	}
	if value.ptr == nil {
		return nil, false
	}

	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(value.ptr)), length)...), true
}

// copyGhosttyStringArray copies a borrowed GhosttyString array into Go-owned
// strings. A nonzero count requires a non-NULL array pointer.
func copyGhosttyStringArray(values *C.GhosttyString, count C.size_t) ([]string, bool) {
	length, ok := ghosttySizeToInt(count)
	if !ok || (length > 0 && values == nil) {
		return nil, false
	}

	result := make([]string, length)
	if length == 0 {
		return result, true
	}
	for i, value := range unsafe.Slice(values, length) {
		data, valid := copyGhosttyString(value)
		if !valid {
			return nil, false
		}
		result[i] = string(data)
	}
	return result, true
}

// cGhosttyStringArray owns a C-allocated array of GhosttyString descriptors
// and the C-allocated byte buffers referenced by those descriptors. Keeping
// every pointer in C memory satisfies cgo's nested-pointer rules for reply
// structs passed synchronously to libghostty.
type cGhosttyStringArray struct {
	ptr            *C.GhosttyString
	descriptorPtr  unsafe.Pointer
	descriptorSize uintptr
	buffers        []cGhosttyStringBuffer
}

// cGhosttyStringBuffer records one allocation so it can be released with the
// exact size required by ghostty_free.
type cGhosttyStringBuffer struct {
	ptr  unsafe.Pointer
	size uintptr
}

// newCGhosttyStringArray copies values into a C-owned GhosttyString array.
// The caller must close the returned owner after the synchronous C call.
func newCGhosttyStringArray(values [][]byte) (*cGhosttyStringArray, error) {
	a := &cGhosttyStringArray{}
	if len(values) == 0 {
		return a, nil
	}

	elementSize := uintptr(C.sizeof_GhosttyString)
	if uintptr(len(values)) > ^uintptr(0)/elementSize {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	a.descriptorSize = uintptr(len(values)) * elementSize
	a.descriptorPtr = Alloc(a.descriptorSize)
	if a.descriptorPtr == nil {
		return nil, &Error{Result: ResultOutOfMemory}
	}
	a.ptr = (*C.GhosttyString)(a.descriptorPtr)
	a.buffers = make([]cGhosttyStringBuffer, 0, len(values))

	descriptors := unsafe.Slice(a.ptr, len(values))
	for i, value := range values {
		if len(value) == 0 {
			descriptors[i] = C.GhosttyString{}
			continue
		}

		size := uintptr(len(value))
		ptr := Alloc(size)
		if ptr == nil {
			a.close()
			return nil, &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(ptr), len(value)), value)
		a.buffers = append(a.buffers, cGhosttyStringBuffer{ptr: ptr, size: size})
		descriptors[i] = C.GhosttyString{
			ptr: (*C.uint8_t)(ptr),
			len: C.size_t(len(value)),
		}
	}

	return a, nil
}

// close releases every descriptor and data allocation owned by a.
func (a *cGhosttyStringArray) close() {
	if a == nil {
		return
	}
	for _, buffer := range a.buffers {
		Free(buffer.ptr, buffer.size)
	}
	Free(a.descriptorPtr, a.descriptorSize)
	a.ptr = nil
	a.descriptorPtr = nil
	a.descriptorSize = 0
	a.buffers = nil
}

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
