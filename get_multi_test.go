package libghostty

import (
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// The tests in this file cover getMultiArgs and every public GetMulti
// that is not already tested elsewhere. Each GetMulti test reads at
// least two keys, and the second key never has its zero value. That way
// the test fails if any key after the first is dropped or misread.
//
// These GetMulti callers are tested elsewhere:
//   - KittyGraphicsImage.Info: TestKittyGraphicsImageInfo
//   - KittyGraphicsPlacementIterator.Info: TestKittyGraphicsPlacementInfo
//   - SelectionGesture.GetMulti: TestSelectionGestureAPIs
//   - Search.GetMulti: TestSearchLifecycleAndResults
//   - SnapshotDecoder.GetMulti, through Progress: TestSnapshotIncrementalDecode

// getMultiGlobal gives TestGetMultiArgsValueKinds a pointer to a Go
// global.
var getMultiGlobal uint64

// TestGetMultiArgsValueKinds checks that alloc accepts every kind of
// pointer a caller may pass and stores each one unchanged. Pinning only
// applies to Go heap memory, so the other kinds must pass through.
func TestGetMultiArgsValueKinds(t *testing.T) {
	heap := new(uint64)
	cmem := Alloc(8)
	if cmem == nil {
		t.Fatal("allocation failed")
	}
	defer Free(cmem, 8)

	values := []unsafe.Pointer{
		unsafe.Pointer(heap),
		unsafe.Pointer(&getMultiGlobal),
		cmem,
		nil,
	}

	var args getMultiArgs
	if err := args.alloc(values); err != nil {
		t.Fatal(err)
	}
	defer args.free()

	got := unsafe.Slice(args.values, len(values))
	for i := range values {
		if got[i] != values[i] {
			t.Fatalf("value %d: got %p, want %p", i, got[i], values[i])
		}
	}
}

// TestGetMultiArgsEmpty checks that alloc with no values allocates
// nothing and that free then does nothing.
func TestGetMultiArgsEmpty(t *testing.T) {
	var args getMultiArgs
	if err := args.alloc(nil); err != nil {
		t.Fatal(err)
	}
	if args.values != nil {
		t.Fatal("expected no allocation for empty values")
	}
	args.free()
}

// TestAllocWithKeys checks that allocWithKeys converts the keys to the
// C type and stores them right after the output pointers.
func TestAllocWithKeys(t *testing.T) {
	var a, b uint16
	keys := []TerminalData{TerminalDataCols, TerminalDataRows}
	values := []unsafe.Pointer{unsafe.Pointer(&a), unsafe.Pointer(&b)}

	var args getMultiArgs
	cKeys, err := allocWithKeys[int32](&args, keys, values)
	if err != nil {
		t.Fatal(err)
	}
	defer args.free()

	ptrSize := unsafe.Sizeof(unsafe.Pointer(nil))
	if want := uintptr(len(keys)) * (ptrSize + 4); args.size != want {
		t.Fatalf("expected block size %d, got %d", want, args.size)
	}
	if want := unsafe.Add(unsafe.Pointer(args.values), 2*ptrSize); unsafe.Pointer(cKeys) != want {
		t.Fatal("expected keys to follow the output pointers")
	}
	for i, key := range unsafe.Slice(cKeys, len(keys)) {
		if int(key) != int(keys[i]) {
			t.Fatalf("key %d: got %d, want %d", i, key, keys[i])
		}
	}
}

// TestGetMultiArgsFreeZeroes checks that free zeroes the whole block,
// keys included, before freeing it. The test hook runs after the
// zeroing and before the free, so the test never reads freed memory.
func TestGetMultiArgsFreeZeroes(t *testing.T) {
	var calls int
	testHookGetMultiFree = func(block unsafe.Pointer, size uintptr) {
		calls++
		for i, b := range unsafe.Slice((*byte)(block), size) {
			if b != 0 {
				t.Errorf("byte %d of %d not zeroed before free: %#x", i, size, b)
				return
			}
		}
	}
	defer func() { testHookGetMultiFree = nil }()

	var a, b uint16
	var args getMultiArgs
	if _, err := allocWithKeys[int32](
		&args,
		[]TerminalData{TerminalDataCols, TerminalDataRows},
		[]unsafe.Pointer{unsafe.Pointer(&a), unsafe.Pointer(&b)},
	); err != nil {
		t.Fatal(err)
	}
	args.free()
	if calls != 1 {
		t.Fatalf("expected the free hook to run once, ran %d times", calls)
	}
	if args.values != nil || args.size != 0 {
		t.Fatal("expected free to reset the block")
	}

	// A second free has nothing left to zero or free.
	args.free()
	if calls != 1 {
		t.Fatalf("expected a second free to do nothing, hook ran %d times", calls)
	}
}

// TestGetMultiArgsKeepsValuesUntilFree checks that a value stays alive
// while getMultiArgs holds it, even with no other Go reference to it,
// and can be collected after free.
//
// runtime.Pinner cannot report whether an object is pinned, so this is
// the closest the test can get. It cannot tell pinning apart from the
// Pinner simply holding a reference, because both keep the object
// alive. To check the pinning itself, run the tests with
// GOEXPERIMENT=cgocheck2. The runtime then stops the program if Go
// stores an unpinned pointer in C memory.
func TestGetMultiArgsKeepsValuesUntilFree(t *testing.T) {
	var args getMultiArgs
	collected := make(chan struct{})
	func() {
		// The object is only reachable from this function. Once it
		// returns, only args and the C block refer to it, and the
		// garbage collector does not look inside C memory.
		obj := new([64]byte)
		runtime.AddCleanup(obj, func(ch chan struct{}) { close(ch) }, collected)
		if err := args.alloc([]unsafe.Pointer{unsafe.Pointer(obj)}); err != nil {
			t.Fatal(err)
		}
	}()

	for range 5 {
		runtime.GC()
	}
	select {
	case <-collected:
		t.Fatal("value was collected while getMultiArgs held it")
	case <-time.After(50 * time.Millisecond):
	}

	args.free()

	deadline := time.After(5 * time.Second)
	for {
		runtime.GC()
		select {
		case <-collected:
			return
		case <-deadline:
			t.Fatal("value was not collected after free")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// TestTerminalGetMulti checks Terminal.GetMulti.
func TestTerminalGetMulti(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	var cols, rows uint16
	if err := term.GetMulti(
		[]TerminalData{TerminalDataCols, TerminalDataRows},
		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
	); err != nil {
		t.Fatal(err)
	}
	if cols != 80 || rows != 24 {
		t.Fatalf("expected 80x24, got %dx%d", cols, rows)
	}
}

// TestCellAndRowGetMulti checks Cell.GetMulti and Row.GetMulti.
func TestCellAndRowGetMulti(t *testing.T) {
	term, err := NewTerminal(WithSize(10, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Twelve bold characters on a ten column screen: the first row is
	// styled and soft wraps into the second.
	term.VTWrite([]byte("\x1b[1mABCDEFGHIJKL"))

	ref, err := term.GridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}

	cell, err := ref.Cell()
	if err != nil {
		t.Fatal(err)
	}
	var codepoint uint32
	var hasText, hasStyling bool
	if err := cell.GetMulti(
		[]CellData{CellDataCodepoint, CellDataHasText, CellDataHasStyling},
		[]unsafe.Pointer{
			unsafe.Pointer(&codepoint),
			unsafe.Pointer(&hasText),
			unsafe.Pointer(&hasStyling),
		},
	); err != nil {
		t.Fatal(err)
	}
	if codepoint != 'A' || !hasText || !hasStyling {
		t.Fatalf(
			"unexpected cell values: codepoint=%U has_text=%t has_styling=%t",
			codepoint, hasText, hasStyling,
		)
	}

	row, err := ref.Row()
	if err != nil {
		t.Fatal(err)
	}
	var wrap, styled bool
	if err := row.GetMulti(
		[]RowData{RowDataWrap, RowDataStyled},
		[]unsafe.Pointer{unsafe.Pointer(&wrap), unsafe.Pointer(&styled)},
	); err != nil {
		t.Fatal(err)
	}
	if !wrap || !styled {
		t.Fatalf("unexpected row values: wrap=%t styled=%t", wrap, styled)
	}
}

// TestRenderStateGetMulti checks RenderState.GetMulti,
// RenderStateRowIterator.GetMulti and RenderStateRowCells.GetMulti.
func TestRenderStateGetMulti(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("first\r\n\x1b[1mhello"))

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	var cols, rows uint16
	if err := rs.GetMulti(
		[]RenderStateData{RenderStateDataCols, RenderStateDataRows},
		[]unsafe.Pointer{unsafe.Pointer(&cols), unsafe.Pointer(&rows)},
	); err != nil {
		t.Fatal(err)
	}
	if cols != 80 || rows != 24 {
		t.Fatalf("expected 80x24, got %dx%d", cols, rows)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()
	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	// Move to the second row, so its viewport position is not zero.
	if !ri.Next() || !ri.Next() {
		t.Fatal("expected at least two rows")
	}
	wantDirty, err := ri.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	var dirty bool
	var viewportY int32
	if err := ri.GetMulti(
		[]RenderStateRowData{RenderStateRowDataDirty, RenderStateRowDataViewportY},
		[]unsafe.Pointer{unsafe.Pointer(&dirty), unsafe.Pointer(&viewportY)},
	); err != nil {
		t.Fatal(err)
	}
	if dirty != wantDirty || viewportY != 1 {
		t.Fatalf(
			"unexpected row values: dirty=%t (want %t) viewport_y=%d (want 1)",
			dirty, wantDirty, viewportY,
		)
	}

	rc, err := NewRenderStateRowCells()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}
	if !rc.Next() {
		t.Fatal("expected at least one cell")
	}
	var graphemesLen uint32
	var hasStyling bool
	if err := rc.GetMulti(
		[]RenderStateRowCellsData{
			RenderStateRowCellsDataGraphemesLen,
			RenderStateRowCellsDataHasStyling,
		},
		[]unsafe.Pointer{unsafe.Pointer(&graphemesLen), unsafe.Pointer(&hasStyling)},
	); err != nil {
		t.Fatal(err)
	}
	if graphemesLen != 1 || !hasStyling {
		t.Fatalf(
			"unexpected cell values: graphemes_len=%d has_styling=%t",
			graphemesLen, hasStyling,
		)
	}
}

// TestKittyGraphicsGetMulti checks KittyGraphicsImage.GetMulti and
// KittyGraphicsPlacementIterator.GetMulti.
func TestKittyGraphicsGetMulti(t *testing.T) {
	term := newKittyTerminal(t)
	defer term.Close()
	sendKittyImage(t, term)

	kg, err := term.KittyGraphics()
	if err != nil {
		t.Fatal(err)
	}
	iter, err := NewKittyGraphicsPlacementIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()
	if err := kg.PlacementIterator(iter); err != nil {
		t.Fatal(err)
	}
	if !iter.Next() {
		t.Fatal("expected at least one placement")
	}

	wantImageID, err := iter.ImageID()
	if err != nil {
		t.Fatal(err)
	}
	wantPlacementID, err := iter.PlacementID()
	if err != nil {
		t.Fatal(err)
	}
	// The image ID goes second because it is never zero.
	var placementID, imageID uint32
	if err := iter.GetMulti(
		[]KittyGraphicsPlacementData{
			KittyGraphicsPlacementDataPlacementID,
			KittyGraphicsPlacementDataImageID,
		},
		[]unsafe.Pointer{unsafe.Pointer(&placementID), unsafe.Pointer(&imageID)},
	); err != nil {
		t.Fatal(err)
	}
	if imageID != wantImageID || placementID != wantPlacementID {
		t.Fatalf(
			"unexpected placement values: image_id=%d (want %d) placement_id=%d (want %d)",
			imageID, wantImageID, placementID, wantPlacementID,
		)
	}

	img := kg.Image(wantImageID)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	var id, width, height uint32
	if err := img.GetMulti(
		[]KittyGraphicsImageData{
			KittyGraphicsImageDataID,
			KittyGraphicsImageDataWidth,
			KittyGraphicsImageDataHeight,
		},
		[]unsafe.Pointer{
			unsafe.Pointer(&id),
			unsafe.Pointer(&width),
			unsafe.Pointer(&height),
		},
	); err != nil {
		t.Fatal(err)
	}
	if id != wantImageID || width != 1 || height != 1 {
		t.Fatalf(
			"unexpected image values: id=%d (want %d) size=%dx%d (want 1x1)",
			id, wantImageID, width, height,
		)
	}
}
