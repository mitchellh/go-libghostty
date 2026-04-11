package libghostty

// Shared helpers for the get_multi pattern used by multiple types.
// These helpers solve the cgo pointer-passing rule: Go cannot pass
// a Go-allocated void** (array of pointers to Go memory) directly
// to C. Instead, we allocate the void** array in C heap memory,
// copy the Go pointer values in, call the C function, then free.

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

// cValuesArray allocates a C-heap array of void* pointers, copies the
// Go unsafe.Pointer values into it, and returns the C array pointer.
// The caller must free the returned pointer with C.free when done.
func cValuesArray(values []unsafe.Pointer) *unsafe.Pointer {
	n := len(values)
	cArr := (*unsafe.Pointer)(C.malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)))))
	dst := unsafe.Slice(cArr, n)
	copy(dst, values)
	return cArr
}
