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
