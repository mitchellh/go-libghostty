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
// C: ghostty_alloc
func Alloc(len uintptr) unsafe.Pointer {
	return unsafe.Pointer(C.ghostty_alloc(nil, C.size_t(len)))
}

// Free frees memory allocated by Alloc (or returned by a libghostty
// function) using the default libghostty allocator (NULL allocator).
// The len must match the original allocation size. It is safe to pass
// nil.
// C: ghostty_free
func Free(ptr unsafe.Pointer, len uintptr) {
	C.ghostty_free(nil, (*C.uint8_t)(ptr), C.size_t(len))
}
