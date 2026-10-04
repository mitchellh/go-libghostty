package libghostty

// This file holds the shared code behind every GetMulti method.
//
// A libghostty get_multi function reads several values in one call. It
// takes an array of keys and an array of output pointers, one per key,
// and writes each value through its pointer. The output pointers
// usually point at Go memory, such as a local variable or a struct
// field.
//
// cgo does not let Go pass C an array of Go pointers, so the output
// pointers are copied into C memory instead. Keeping Go pointers in C
// memory is safe only if three rules are followed. getMultiArgs follows
// all of them, so a GetMulti method cannot forget one.
//
//  1. Pin each value before storing it. The cgo rules require any Go
//     memory that C memory points to to be pinned with runtime.Pinner
//     for as long as the pointer is stored. Today's garbage collector
//     does not move objects, but cgo does not promise that.
//
//  2. Zero the C memory before using it and again before freeing it.
//     When Go stores a pointer, the garbage collector may look at the
//     value being overwritten. Leftover data that looks like a Go
//     pointer can make it stop the program with "found bad pointer in
//     Go heap". Zeroing on allocation protects this call (see
//     allocZeroed). Zeroing before freeing protects whoever is handed
//     the memory next.
//
//  3. Unpin last, after the C memory no longer holds the pointers.
//
// The keys are copied into C memory too. The Go key types, such as
// CellData, are ints, which are wider than the C enums, so C cannot
// read a Go key slice directly. Converted keys could live in a Go
// slice, but cgo moves any Go memory passed to C onto the heap, which
// would cost an allocation on every call. Renderers make these calls
// for every row and cell, so the keys share the C block with the
// output pointers instead.

import (
	"runtime"
	"unsafe"
)

// getMultiArgs holds the arguments of one get_multi call in a single
// block of C memory: the output pointers, optionally followed by the
// keys. It also pins the Go memory the output pointers refer to.
//
// Declare it as a local variable, set it up with alloc or
// allocWithKeys, and free it once the C call returns:
//
//	var args getMultiArgs
//	if err := args.alloc(values); err != nil {
//		return err
//	}
//	defer args.free()
//	C.ghostty_example_get_multi(obj, n, keys, args.values, nil)
//
// A getMultiArgs must not be copied after it has been set up, because
// it contains a runtime.Pinner. Setting it up in place, rather than
// returning it from a constructor, also keeps it on the caller's stack
// so that it costs no heap allocation.
type getMultiArgs struct {
	// values is the start of the C block, which begins with the output
	// pointers. It is nil until the block is allocated and after it is
	// freed.
	values *unsafe.Pointer

	// size is the size of the C block in bytes.
	size uintptr

	// pinner pins the Go memory that the output pointers refer to.
	pinner runtime.Pinner
}

// testHookGetMultiFree, if not nil, is called by getMultiArgs.free
// after the block has been zeroed and before it is freed. Tests use it
// to check the zeroing without reading freed memory.
var testHookGetMultiFree func(block unsafe.Pointer, size uintptr)

// alloc pins each non-nil value, then allocates a zeroed C block and
// copies the values into it. If values is empty, alloc does nothing.
//
// A value may point at Go heap memory, a Go global or variable, or C
// memory. Pinning only affects Go heap memory, so the other kinds are
// stored unchanged. A nil value is stored as nil.
//
// If the allocation fails, alloc returns an error with
// ResultOutOfMemory and leaves nothing pinned.
func (a *getMultiArgs) alloc(values []unsafe.Pointer) error {
	return a.allocBlock(values, 0)
}

// allocBlock does the work of alloc. It reserves keyBytes extra bytes
// after the output pointers for allocWithKeys to store keys in.
func (a *getMultiArgs) allocBlock(values []unsafe.Pointer, keyBytes uintptr) error {
	if len(values) == 0 {
		return nil
	}

	// Pin before storing, so the block never holds a pointer to
	// unpinned memory.
	for _, v := range values {
		if v != nil {
			a.pinner.Pin(v)
		}
	}

	// The block holds pointers, so it must start out zeroed.
	size := uintptr(len(values))*unsafe.Sizeof(unsafe.Pointer(nil)) + keyBytes
	block := (*unsafe.Pointer)(allocZeroed(size))
	if block == nil {
		a.pinner.Unpin()
		return &Error{Result: ResultOutOfMemory}
	}

	copy(unsafe.Slice(block, len(values)), values)
	a.values = block
	a.size = size
	return nil
}

// allocWithKeys does the same as a.alloc(values) and also stores keys
// in the block, after the output pointers, converted to the C enum type
// CKey. It returns a pointer to the first key, which stays valid until
// a.free is called. keys and values must have the same length. If they
// are empty, allocWithKeys does nothing and returns nil.
//
// allocWithKeys is a function because Go methods cannot have type
// parameters. CKey comes first so that callers can name the C type and
// let the compiler infer the Go one:
//
//	cKeys, err := allocWithKeys[C.GhosttyCellData](&args, keys, values)
//
// CKey is limited to 32-bit integer types because every libghostty enum
// is a C int. The constraint lets the compiler check each caller's C
// type.
func allocWithKeys[CKey ~int32 | ~uint32, GoKey ~int](
	a *getMultiArgs,
	keys []GoKey,
	values []unsafe.Pointer,
) (*CKey, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	keySize := unsafe.Sizeof(CKey(0))
	if err := a.allocBlock(values, uintptr(len(keys))*keySize); err != nil {
		return nil, err
	}

	// The keys start right after the output pointers. Pointers are
	// 4 or 8 bytes and keys are 4, so the keys are always aligned.
	first := (*CKey)(unsafe.Add(
		unsafe.Pointer(a.values),
		uintptr(len(values))*unsafe.Sizeof(unsafe.Pointer(nil)),
	))
	cKeys := unsafe.Slice(first, len(keys))
	for i, key := range keys {
		cKeys[i] = CKey(key)
	}
	return first, nil
}

// free zeroes the C block, frees it, and then unpins the values. It is
// safe to call when nothing was allocated, and to call more than once.
func (a *getMultiArgs) free() {
	if a.values != nil {
		// Zero the block one byte at a time. Writing nil through
		// *unsafe.Pointer would be a pointer store, and the garbage
		// collector would look at the old values while they are being
		// removed. See allocZeroed.
		clear(unsafe.Slice((*byte)(unsafe.Pointer(a.values)), a.size))
		if testHookGetMultiFree != nil {
			testHookGetMultiFree(unsafe.Pointer(a.values), a.size)
		}
		Free(unsafe.Pointer(a.values), a.size)
		a.values = nil
		a.size = 0
	}

	// Unpin last, now that no C memory holds the pointers.
	a.pinner.Unpin()
}
