package libghostty

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

// The benchmarks in this file measure the cost of one GetMulti call on
// the render state types. Renderers make these calls for every row and
// every cell of every frame, so any fixed cost per call, such as
// allocating, pinning, or zeroing memory, adds directly to frame time.

// newGetMultiBenchmarkState returns a render state for an 80x30 screen
// full of text, with a row iterator positioned on the first row and a
// cell iterator positioned on its first cell.
func newGetMultiBenchmarkState(b *testing.B) (*RenderStateRowIterator, *RenderStateRowCells) {
	b.Helper()

	term, err := NewTerminal(WithSize(80, 30), WithMaxScrollbackLines(0))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { term.Close() })
	line := strings.Repeat("x", 80)
	term.VTWrite([]byte(strings.Repeat(line+"\r\n", 29) + line))

	rs, err := NewRenderState()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { rs.Close() })
	if err := rs.Update(term); err != nil {
		b.Fatal(err)
	}

	ri, err := NewRenderStateRowIterator()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { ri.Close() })
	if err := rs.RowIterator(ri); err != nil {
		b.Fatal(err)
	}
	if !ri.Next() {
		b.Fatal("expected at least one row")
	}

	rc, err := NewRenderStateRowCells()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { rc.Close() })
	if err := ri.Cells(rc); err != nil {
		b.Fatal(err)
	}
	if !rc.Next() {
		b.Fatal("expected at least one cell")
	}

	return ri, rc
}

// BenchmarkRenderStateRowCellsGetMulti measures one GetMulti call that
// reads a cell's raw value and, with two keys, its grapheme length: the
// kind of call a renderer makes for every cell.
func BenchmarkRenderStateRowCellsGetMulti(b *testing.B) {
	_, rc := newGetMultiBenchmarkState(b)

	// The output variables and slices are set up once, so the benchmark
	// measures only the call.
	var raw uint64
	var graphemesLen uint32
	keys := []RenderStateRowCellsData{
		RenderStateRowCellsDataRaw,
		RenderStateRowCellsDataGraphemesLen,
	}
	values := []unsafe.Pointer{
		unsafe.Pointer(&raw),
		unsafe.Pointer(&graphemesLen),
	}

	for n := 1; n <= len(keys); n++ {
		b.Run(fmt.Sprintf("Keys%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := rc.GetMulti(keys[:n], values[:n]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRenderStateRowGetMulti measures one GetMulti call that reads
// a row's raw value and, with two keys, its dirty flag: the kind of
// call a renderer makes for every row.
func BenchmarkRenderStateRowGetMulti(b *testing.B) {
	ri, _ := newGetMultiBenchmarkState(b)

	// The output variables and slices are set up once, so the benchmark
	// measures only the call.
	var raw uint64
	var dirty bool
	keys := []RenderStateRowData{
		RenderStateRowDataRaw,
		RenderStateRowDataDirty,
	}
	values := []unsafe.Pointer{
		unsafe.Pointer(&raw),
		unsafe.Pointer(&dirty),
	}

	for n := 1; n <= len(keys); n++ {
		b.Run(fmt.Sprintf("Keys%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := ri.GetMulti(keys[:n], values[:n]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
