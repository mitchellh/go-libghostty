package libghostty

// Shared helpers for the get_multi pattern used by multiple types.
// These helpers solve the cgo pointer-passing rule: Go cannot pass
// a Go-allocated void** (array of pointers to Go memory) directly
// to C. Instead, we allocate the void** array via libghostty's
// allocator, copy the Go pointer values in, call the C function,
// then free.

import "unsafe"

// cValuesArray allocates a C-heap array of void* pointers via the
// libghostty allocator, copies the Go unsafe.Pointer values into it,
// and returns the C array pointer and allocation size. The caller must
// free the returned pointer with Free(ptr, size) when done.
func cValuesArray(values []unsafe.Pointer) (*unsafe.Pointer, uintptr) {
	n := len(values)
	size := uintptr(n) * unsafe.Sizeof(unsafe.Pointer(nil))
	// The array holds pointers, so it must start out zeroed. See
	// allocZeroed.
	cArr := (*unsafe.Pointer)(allocZeroed(size))
	dst := unsafe.Slice(cArr, n)
	copy(dst, values)
	return cArr, size
}
