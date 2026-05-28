package libghostty

import "testing"

func TestRenderStateRowCells(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
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

	// Advance to first row.
	if !ri.Next() {
		t.Fatal("expected at least one row")
	}

	rc, err := NewRenderStateRowCells()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}

	// Iterate cells and count them.
	count := 0
	for rc.Next() {
		count++
	}
	if count != 80 {
		t.Fatalf("expected 80 cells, got %d", count)
	}
}

func TestRenderStateRowCellsSelect(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("ABCDE"))

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

	rc, err := NewRenderStateRowCells()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}

	// Select column 2 (should be 'C').
	if err := rc.Select(2); err != nil {
		t.Fatal(err)
	}

	graphemes, err := rc.Graphemes()
	if err != nil {
		t.Fatal(err)
	}
	if len(graphemes) != 1 || graphemes[0] != 'C' {
		t.Fatalf("expected ['C'], got %v", graphemes)
	}
}

func TestRenderStateRowCellsGraphemes(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("AB"))

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

	rc, err := NewRenderStateRowCells()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}

	// First cell should be 'A'.
	if !rc.Next() {
		t.Fatal("expected at least one cell")
	}
	graphemes, err := rc.Graphemes()
	if err != nil {
		t.Fatal(err)
	}
	if len(graphemes) != 1 || graphemes[0] != 'A' {
		t.Fatalf("expected ['A'], got %v", graphemes)
	}

	// Second cell should be 'B'.
	if !rc.Next() {
		t.Fatal("expected second cell")
	}
	graphemes, err = rc.Graphemes()
	if err != nil {
		t.Fatal(err)
	}
	if len(graphemes) != 1 || graphemes[0] != 'B' {
		t.Fatalf("expected ['B'], got %v", graphemes)
	}

	// Empty cell should return nil.
	if !rc.Next() {
		t.Fatal("expected third cell")
	}
	graphemes, err = rc.Graphemes()
	if err != nil {
		t.Fatal(err)
	}
	if graphemes != nil {
		t.Fatalf("expected nil graphemes for empty cell, got %v", graphemes)
	}
}

func TestRenderStateRowCellsStyle(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write one bold cell and one default-styled cell.
	term.VTWrite([]byte("\x1b[1mX\x1b[0mY"))

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

	style, err := rc.Style()
	if err != nil {
		t.Fatal(err)
	}
	if !style.Bold() {
		t.Fatal("expected bold style")
	}

	hasStyling, err := rc.HasStyling()
	if err != nil {
		t.Fatal(err)
	}
	if !hasStyling {
		t.Fatal("expected first cell to have styling")
	}

	if !rc.Next() {
		t.Fatal("expected second cell")
	}
	hasStyling, err = rc.HasStyling()
	if err != nil {
		t.Fatal(err)
	}
	if hasStyling {
		t.Fatal("expected second cell to have default styling")
	}
}

func TestRenderStateRowCellsSelected(t *testing.T) {
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

	rc, err := NewRenderStateRowCells()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}

	for x := uint16(0); x < 5; x++ {
		if err := rc.Select(x); err != nil {
			t.Fatal(err)
		}
		selected, err := rc.Selected()
		if err != nil {
			t.Fatal(err)
		}
		expected := x >= 1 && x <= 3
		if selected != expected {
			t.Fatalf("cell %d selected = %t, expected %t", x, selected, expected)
		}
	}
}

func TestRenderStateRowCellsColors(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write text with explicit fg/bg colors (SGR 38;2;R;G;B and 48;2;R;G;B).
	term.VTWrite([]byte("\x1b[38;2;255;0;0;48;2;0;0;255mX"))

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

	fg, err := rc.FgColor()
	if err != nil {
		t.Fatal(err)
	}
	if fg == nil {
		t.Fatal("expected non-nil fg color")
	}
	if *fg != (ColorRGB{R: 255, G: 0, B: 0}) {
		t.Fatalf("expected red fg, got %+v", *fg)
	}

	bg, err := rc.BgColor()
	if err != nil {
		t.Fatal(err)
	}
	if bg == nil {
		t.Fatal("expected non-nil bg color")
	}
	if *bg != (ColorRGB{R: 0, G: 0, B: 255}) {
		t.Fatalf("expected blue bg, got %+v", *bg)
	}

	// Empty cell should have nil colors (use default).
	if !rc.Next() {
		t.Fatal("expected second cell")
	}
	fg, err = rc.FgColor()
	if err != nil {
		t.Fatal(err)
	}
	if fg != nil {
		t.Fatalf("expected nil fg for unstyled cell, got %+v", *fg)
	}
}
