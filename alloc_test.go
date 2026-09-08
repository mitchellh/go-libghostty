package libghostty

import "testing"

func TestAllocZeroLength(t *testing.T) {
	// Empty allocations must cross cgo as nil, never as Zig's empty-slice
	// sentinel, which is not a valid pointer for Go's stack scanner.
	ptr := Alloc(0)
	defer Free(ptr, 0)
	if ptr != nil {
		t.Fatal("expected zero-length allocation to return nil")
	}
}
