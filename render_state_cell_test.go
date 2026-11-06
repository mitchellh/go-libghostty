package libghostty

import (
	"slices"
	"testing"
)

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

func TestRenderStateRowCellsGraphemesInto(t *testing.T) {
	rc := testRenderStateRowCells(t, "ABC")

	// Select column 1 (should be 'B') and append it into caller-owned scratch
	// that already has content. This verifies append semantics and lets callers
	// reuse the same backing array by passing scratch[:0].
	if err := rc.Select(1); err != nil {
		t.Fatal(err)
	}
	scratch := make([]uint32, 1, 4)
	scratch[0] = 'A'
	base := &scratch[0]
	graphemes, err := rc.GraphemesInto(scratch)
	if err != nil {
		t.Fatal(err)
	}
	if len(graphemes) != 2 || graphemes[0] != 'A' || graphemes[1] != 'B' {
		t.Fatalf("expected ['A', 'B'], got %v", graphemes)
	}
	if &graphemes[0] != base {
		t.Fatal("expected GraphemesInto to reuse caller-provided capacity")
	}

	// Empty cells append nothing and return the same slice unchanged.
	if err := rc.Select(3); err != nil {
		t.Fatal(err)
	}
	graphemes, err = rc.GraphemesInto(graphemes)
	if err != nil {
		t.Fatal(err)
	}
	if len(graphemes) != 2 || graphemes[0] != 'A' || graphemes[1] != 'B' {
		t.Fatalf("expected empty cell to leave graphemes unchanged, got %v", graphemes)
	}
}

func TestRenderStateRowCellsAppendGraphemes(t *testing.T) {
	rc := testRenderStateRowCells(t, "Aé")

	var text []byte
	for x := uint16(0); x < 2; x++ {
		if err := rc.Select(x); err != nil {
			t.Fatal(err)
		}
		var err error
		text, err = rc.AppendGraphemes(text)
		if err != nil {
			t.Fatal(err)
		}
	}
	if string(text) != "Aé" {
		t.Fatalf("expected %q, got %q", "Aé", string(text))
	}

	// Empty cells append nothing and preserve the existing byte buffer content.
	if err := rc.Select(2); err != nil {
		t.Fatal(err)
	}
	text, err := rc.AppendGraphemes(text)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "Aé" {
		t.Fatalf("expected empty cell to leave text unchanged, got %q", string(text))
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

func TestRenderStateRowCellsStyleInto(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write one heavily styled cell and one default-styled cell. The styled
	// cell exercises the resolved RGB color path as well as the style flags Rex
	// needs to produce ANSI output without allocating per-cell Style/Color values.
	term.VTWrite([]byte("\x1b[1;2;3;4;7;9;38;2;255;0;0;48;2;0;0;255mX\x1b[0mY"))

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

	var style RenderCellStyle
	if err := rc.StyleInto(&style); err != nil {
		t.Fatal(err)
	}
	if !style.HasStyling {
		t.Fatal("expected first cell to have styling")
	}
	if !style.HasForeground || style.Foreground != (ColorRGB{R: 255, G: 0, B: 0}) {
		t.Fatalf("expected red foreground, got %#v", style)
	}
	if !style.HasBackground || style.Background != (ColorRGB{R: 0, G: 0, B: 255}) {
		t.Fatalf("expected blue background, got %#v", style)
	}
	if !style.Bold || !style.Faint || !style.Italic || !style.Underline || !style.Strikethrough || !style.Inverse {
		t.Fatalf("expected all style flags to be set, got %+v", style)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := rc.StyleInto(&style); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected StyleInto to allocate zero times, got %v", allocs)
	}

	if !rc.Next() {
		t.Fatal("expected second cell")
	}
	if err := rc.StyleInto(&style); err != nil {
		t.Fatal(err)
	}
	if style.HasStyling {
		t.Fatal("expected second cell to have default styling")
	}
	if style.HasForeground || style.HasBackground {
		t.Fatalf("expected default style to have no colors, got %+v", style)
	}
	if style.Bold || style.Faint || style.Italic || style.Underline || style.Strikethrough || style.Inverse {
		t.Fatalf("expected all style flags to be cleared, got %+v", style)
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

func testRenderStateRowCells(t *testing.T, text string) *RenderStateRowCells {
	t.Helper()

	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(term.Close)

	term.VTWrite([]byte(text))

	rs, err := NewRenderState()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rs.Close)

	if err := rs.Update(term); err != nil {
		t.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ri.Close)

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
	t.Cleanup(rc.Close)

	if err := ri.Cells(rc); err != nil {
		t.Fatal(err)
	}

	return rc
}

// TestRenderStateRowCellsGraphemesIntoCapacity covers each outcome of filling a
// GhosttyCodepointBuffer: enough room, no room, and too little room for a
// multi-codepoint cluster. Going through GhosttyCodepointBuffer is what a raw
// pointer would corrupt, so the codepoints are checked exactly, along with what
// happens to the caller's existing elements.
func TestRenderStateRowCellsGraphemesIntoCapacity(t *testing.T) {
	// "e" with two combining marks forms one cell of three codepoints.
	cluster := []uint32{'e', 0x0301, 0x0323}
	rc := testRenderStateRowCells(t, "ẹ́")
	if err := rc.Select(0); err != nil {
		t.Fatal(err)
	}

	prefix := []uint32{'A', 'B'}
	want := append(append([]uint32(nil), prefix...), cluster...)

	tests := []struct {
		name   string
		extra  int
		reused bool
	}{
		{name: "no spare capacity", extra: 0},
		{name: "too little capacity", extra: 2},
		{name: "exact capacity", extra: 3, reused: true},
		{name: "more than enough capacity", extra: 8, reused: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]uint32, len(prefix), len(prefix)+tt.extra)
			copy(dst, prefix)
			base := &dst[0]

			got, err := rc.GraphemesInto(dst)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("got %x, want %x", got, want)
			}
			if reused := &got[0] == base; reused != tt.reused {
				t.Fatalf("reused backing array = %v, want %v", reused, tt.reused)
			}
		})
	}

	// A nil slice has no room at all, which still has to size and fill.
	got, err := rc.GraphemesInto(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, cluster) {
		t.Fatalf("got %x, want %x", got, cluster)
	}
}

// TestRenderStateRowCellsStyleIntoAttributes sets one attribute at a time.
// StyleInto reads the packed flags in C, so this checks each one against the
// right bit. A cell with every attribute set can't tell them apart.
func TestRenderStateRowCellsStyleIntoAttributes(t *testing.T) {
	tests := []struct {
		sgr  string
		want RenderCellStyle
	}{
		{"\x1b[1m", RenderCellStyle{Bold: true}},
		{"\x1b[2m", RenderCellStyle{Faint: true}},
		{"\x1b[3m", RenderCellStyle{Italic: true}},
		{"\x1b[4m", RenderCellStyle{Underline: true}},
		{"\x1b[4:3m", RenderCellStyle{Underline: true}},
		{"\x1b[7m", RenderCellStyle{Inverse: true}},
		{"\x1b[9m", RenderCellStyle{Strikethrough: true}},
	}

	for _, tt := range tests {
		rc := testRenderStateRowCells(t, tt.sgr+"X")
		if err := rc.Select(0); err != nil {
			t.Fatal(err)
		}

		var got RenderCellStyle
		if err := rc.StyleInto(&got); err != nil {
			t.Fatal(err)
		}
		want := tt.want
		want.HasStyling = true
		if got != want {
			t.Errorf("%q: got %+v, want %+v", tt.sgr, got, want)
		}
	}
}
