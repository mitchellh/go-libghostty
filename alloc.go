package libghostty

// Memory allocation helpers wrapping the upstream ghostty_alloc() and
// ghostty_free() functions from allocator.h.
//
// These replace direct C.malloc/C.free calls so that all memory is
// allocated and freed through libghostty's allocator. This is critical
// on platforms where the library's internal allocator differs from the
// consumer's C runtime (e.g. Windows, where Zig's libc and MSVC's CRT
// maintain separate heaps).

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// Alloc allocates len bytes through the default libghostty allocator
// (NULL allocator). Returns a pointer to the allocated memory or nil
// if len is zero or the allocation failed.
//
// The returned memory must be freed with Free using the same length.
//
// The memory is not zeroed. It is safe to fill with bytes or with other
// values that contain no pointers. Before storing a pointer in it, or a
// struct that has a pointer field, zero the whole block through a byte
// slice:
//
//	clear(unsafe.Slice((*byte)(ptr), len))
//
// Skipping this step can crash the program with "found bad pointer in
// Go heap". When Go stores a pointer, the garbage collector may read
// the value being replaced, and leftover data in an unzeroed block can
// look like an invalid Go pointer.
// C: ghostty_alloc
func Alloc(len uintptr) unsafe.Pointer {
	return unsafe.Pointer(C.ghostty_alloc(nil, C.size_t(len)))
}

// allocZeroed allocates size bytes like Alloc and sets every byte to
// zero. It returns nil if size is zero or the allocation failed. The
// memory must be freed with Free using the same size.
//
// Use allocZeroed instead of Alloc for any block that Go code will
// store a pointer in. This includes C structs with pointer fields, such
// as C.GhosttyString. Blocks that only ever hold bytes or other
// pointer-free values can use Alloc.
//
// The reason is the garbage collector. While a collection is running,
// each pointer store also reports the value it replaces. Memory from
// Alloc still holds whatever was there before, so the replaced value is
// leftover data. If that data looks like a Go pointer, the collector
// follows it and stops the program with "found bad pointer in Go heap".
// Zeroed memory leaves the collector nothing to follow.
func allocZeroed(size uintptr) unsafe.Pointer {
	ptr := Alloc(size)
	if ptr != nil {
		// Zero the memory as plain bytes. Assigning zero values
		// through a type that contains pointers would not be safe,
		// because those assignments are pointer stores too.
		clear(unsafe.Slice((*byte)(ptr), size))
	}
	return ptr
}

// Free frees memory allocated by Alloc (or returned by a libghostty
// function) using the default libghostty allocator (NULL allocator).
// The len must match the original allocation size. It is safe to pass
// nil.
// C: ghostty_free
func Free(ptr unsafe.Pointer, len uintptr) {
	C.ghostty_free(nil, (*C.uint8_t)(ptr), C.size_t(len))
}
