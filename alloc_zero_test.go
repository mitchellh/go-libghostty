package libghostty

import (
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"
)

// The tests in this file check that the package never stores a pointer
// in C memory that has not been zeroed. See allocZeroed for why that
// matters.
//
// Each test recreates the conditions for the crash on purpose. It fills
// freed C blocks with a value the garbage collector treats as a bad
// pointer, keeps a collection running, and then calls the code under
// test. If that code skips the zeroing step, the Go runtime stops the
// whole test binary with "found bad pointer in Go heap". A regression
// therefore shows up as a crash, not as an ordinary test failure.

// Test inputs are sized so the arrays under test span several pages.
// Some C allocators zero small blocks when they are freed, and only
// larger blocks reliably come back with their old contents.
const (
	// uninitElems is the number of elements in each array under test.
	uninitElems = 1024

	// uninitRounds is how many times each test repeats its check.
	// Without zeroing, the crash appears within the first few rounds.
	uninitRounds = 200
)

// freedHeapAddr returns an address inside Go heap memory that has
// already been released. The garbage collector reports a fatal error if
// it is ever asked to follow this address.
func freedHeapAddr() uintptr {
	// An allocation this large gets its own region of the heap, and the
	// whole region is released once the slice is collected.
	buf := make([]byte, 4<<20)
	addr := uintptr(unsafe.Pointer(&buf[0])) + 4096
	buf = nil
	runtime.GC()
	runtime.GC()
	return addr
}

// fillFreedBlocks allocates several C blocks of the given size, fills
// them with the word v, and frees them. The next allocation of the same
// size is then very likely to return a block that still holds v.
func fillFreedBlocks(size, v uintptr) {
	var blocks [8]unsafe.Pointer
	for i := range blocks {
		blocks[i] = Alloc(size)

		// uintptr is not a pointer type, so these stores are safe
		// even though the block has not been zeroed.
		words := unsafe.Slice(
			(*uintptr)(blocks[i]),
			size/unsafe.Sizeof(uintptr(0)),
		)
		for j := range words {
			words[j] = v
		}
	}
	for _, block := range blocks {
		Free(block, size)
	}
}

// runGC runs garbage collections back to back in the background until
// the test ends. Pointer stores are only checked while a collection is
// running, so this keeps the check active for most of the test.
func runGC(t *testing.T) {
	t.Helper()

	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		for !stop.Load() {
			runtime.GC()
		}
	}()
	t.Cleanup(func() {
		stop.Store(true)
		<-done
	})
}

func TestAllocZeroed(t *testing.T) {
	const size = 64 * 1024
	fillFreedBlocks(size, ^uintptr(0))

	ptr := allocZeroed(size)
	if ptr == nil {
		t.Fatal("allocZeroed returned nil")
	}
	defer Free(ptr, size)
	for i, b := range unsafe.Slice((*byte)(ptr), size) {
		if b != 0 {
			t.Fatalf("byte %d = %#x, want 0", i, b)
		}
	}
}

func TestStringArrayUninitializedMemory(t *testing.T) {
	bad := freedHeapAddr()
	runGC(t)

	// Use both empty and non-empty values. Each kind takes a different
	// path when it is written to the array.
	values := make([][]byte, uninitElems)
	for i := range values {
		if i%3 != 0 {
			values[i] = []byte("text/plain")
		}
	}

	// Test files cannot import "C", so take the element size from the
	// array's own pointer type.
	var a cGhosttyStringArray
	size := uintptr(len(values)) * unsafe.Sizeof(*a.ptr)

	for range uninitRounds {
		fillFreedBlocks(size, bad)

		arr, err := newCGhosttyStringArray(values)
		if err != nil {
			t.Fatal(err)
		}
		arr.close()
	}
}

func TestClipboardContentsUninitializedMemory(t *testing.T) {
	bad := freedHeapAddr()
	runGC(t)

	values := make([]ClipboardContent, uninitElems)
	for i := range values {
		values[i] = ClipboardContent{
			MIME: "text/plain",
			Data: []byte("hello"),
		}
	}

	// Test files cannot import "C", so take the element size from the
	// array's own pointer type.
	var c cGhosttyClipboardContents
	size := uintptr(len(values)) * unsafe.Sizeof(*c.ptr)

	for range uninitRounds {
		// The clipboard array and the string array behind it have
		// the same size in bytes, so one fill covers both.
		fillFreedBlocks(size, bad)

		contents, err := newCGhosttyClipboardContents(values)
		if err != nil {
			t.Fatal(err)
		}
		contents.close()
	}
}
