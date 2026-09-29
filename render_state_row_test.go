package libghostty

import "testing"

func TestRenderStateRowIterator(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()

	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()

	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	// Iterate all rows and count them.
	count := 0
	for ri.Next() {
		count++
	}
	if count != 24 {
		t.Fatalf("expected 24 rows, got %d", count)
	}
}

func TestRenderStateRowIteratorDirty(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()

	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()

	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	// First row after initial update should be dirty.
	if !ri.Next() {
		t.Fatal("expected at least one row")
	}
	dirty, err := ri.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("expected first row to be dirty after initial update")
	}

	// Clear dirty and verify.
	if err := ri.SetDirty(false); err != nil {
		t.Fatal(err)
	}
	dirty, _ = ri.Dirty()
	if dirty {
		t.Fatal("expected row not dirty after clearing")
	}
}

func TestRenderStateRowIteratorNextDirty(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()
	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	for want := uint16(0); want < 24; want++ {
		got, ok := ri.NextDirty()
		if !ok {
			t.Fatalf("expected dirty row %d", want)
		}
		if got != want {
			t.Fatalf("expected dirty row %d, got %d", want, got)
		}
	}
	if y, ok := ri.NextDirty(); ok {
		t.Fatalf("expected dirty iterator exhaustion, got row %d", y)
	}

	if err := rs.Clean(); err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte("x"))
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}
	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}
	if got, ok := ri.NextDirty(); !ok || got != 0 {
		t.Fatalf("expected only row 0 after a one-row update, got row %d (ok=%v)", got, ok)
	}
	if y, ok := ri.NextDirty(); ok {
		t.Fatalf("expected one partial-dirty row, got extra row %d", y)
	}
}

func TestRenderStateRowIteratorRaw(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write some text so the first row has content.
	term.VTWrite([]byte("hello"))

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()

	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()

	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	if !ri.Next() {
		t.Fatal("expected at least one row")
	}

	row, err := ri.Raw()
	if err != nil {
		t.Fatal(err)
	}

	// The first row should not be wrapped.
	wrap, err := row.Wrap()
	if err != nil {
		t.Fatal(err)
	}
	if wrap {
		t.Fatal("expected first row not to be wrapped")
	}
}

func TestRenderStateRowIteratorCellsRaw(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 2))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("hello"))

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()
	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}
	if !ri.Next() {
		t.Fatal("expected first row")
	}

	view, err := ri.CellsRaw()
	if err != nil {
		t.Fatal(err)
	}
	if view.Len() != 8 {
		t.Fatalf("expected 8 cells, got %d", view.Len())
	}
	cell := view.Cell(0)
	codepoint, err := cell.Codepoint()
	if err != nil {
		t.Fatal(err)
	}
	if codepoint != 'h' {
		t.Fatalf("expected first codepoint %U, got %U", 'h', codepoint)
	}

	cloned := view.Clone()
	if len(cloned) != view.Len() {
		t.Fatalf("expected %d cloned cells, got %d", view.Len(), len(cloned))
	}
	if cloned[0].PackedValue() != cell.PackedValue() {
		t.Fatal("expected cloned packed value to match the borrowed view")
	}
}

func TestRenderStateRowIteratorSelection(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("hello"))
	start, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	end, err := term.GridRef(Point{Tag: PointTagActive, X: 3, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := term.SetSelection(&Selection{Start: *start, End: *end}); err != nil {
		t.Fatal(err)
	}

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()

	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()

	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}

	if !ri.Next() {
		t.Fatal("expected at least one row")
	}
	sel, err := ri.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel == nil {
		t.Fatal("expected row selection")
	}
	if sel.StartX != 1 || sel.EndX != 3 {
		t.Fatalf("expected row selection 1..3, got %d..%d", sel.StartX, sel.EndX)
	}

	if !ri.Next() {
		t.Fatal("expected second row")
	}
	sel, err = ri.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel != nil {
		t.Fatalf("expected no row selection on second row, got %+v", sel)
	}
}

func TestRenderStateOverscanAndRowIdentity(t *testing.T) {
	term, err := NewTerminal(WithSize(10, 3), WithMaxScrollbackLines(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("zero\r\none\r\ntwo\r\nthree\r\nfour\r\nfive\r\nsix"))
	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer ri.Close()
	request := RenderStateOverscan{Above: 1, Below: 1}
	if got, err := rs.OverscanRequest(); err != nil || got != (RenderStateOverscan{}) {
		t.Fatalf("default request: %+v, %v", got, err)
	}
	if err := rs.SetOverscan(request); err != nil {
		t.Fatal(err)
	}
	if got, err := rs.OverscanRequest(); err != nil || got != request {
		t.Fatalf("request: %+v, %v", got, err)
	}
	read := func(want RenderStateOverscan) map[int32]RenderStateRowID {
		t.Helper()
		if err := rs.Update(term); err != nil {
			t.Fatal(err)
		}
		if got, err := rs.Rows(); err != nil || got != 3 {
			t.Fatalf("viewport rows: %d, %v", got, err)
		}
		if got, err := rs.Overscan(); err != nil || got != want {
			t.Fatalf("captured: %+v, %v; want %+v", got, err, want)
		}
		if err := rs.RowIterator(ri); err != nil {
			t.Fatal(err)
		}
		ids := make(map[int32]RenderStateRowID)
		seen := make(map[RenderStateRowID]bool)
		for ri.Next() {
			y, err := ri.ViewportY()
			if err != nil {
				t.Fatal(err)
			}
			if wantY := int32(len(ids)) - int32(want.Above); y != wantY {
				t.Fatalf("y: %d, want %d", y, wantY)
			}
			id, err := ri.ID()
			if err != nil {
				t.Fatal(err)
			}
			if id == (RenderStateRowID{}) || seen[id] {
				t.Fatal("zero or duplicate ID")
			}
			seen[id] = true
			ids[y] = id
		}
		if len(ids) != 3+int(want.Above)+int(want.Below) {
			t.Fatalf("captured %d rows", len(ids))
		}
		return ids
	}
	read(RenderStateOverscan{Above: 1}) // No rows below the active viewport.
	term.ScrollViewportRow(1)
	before := read(request)
	term.ScrollViewportRow(2)
	after := read(request)
	if before[1] != after[0] || before[2] != after[1] {
		t.Fatal("row IDs changed across scrolling")
	}
	// Dirty iteration returns capture indices, while ViewportY includes overscan.
	if err := rs.RowIterator(ri); err != nil {
		t.Fatal(err)
	}
	for i := uint16(0); i < 5; i++ {
		index, ok := ri.NextDirty()
		if !ok || index != i {
			t.Fatalf("dirty index: %d, %v; want %d", index, ok, i)
		}
		if y, err := ri.ViewportY(); err != nil || y != int32(i)-1 {
			t.Fatalf("dirty y: %d, %v", y, err)
		}
	}
	if _, ok := ri.NextDirty(); ok {
		t.Fatal("unexpected extra dirty row")
	}
	term.ScrollViewportTop()
	read(RenderStateOverscan{Below: 1}) // No rows above the first scrollback row.
	if err := rs.SetOverscan(RenderStateOverscan{}); err != nil {
		t.Fatal(err)
	}
	if got, err := rs.Overscan(); err != nil || got != (RenderStateOverscan{Below: 1}) {
		t.Fatalf("setting request changed snapshot: %+v, %v", got, err)
	}
	read(RenderStateOverscan{})
}
