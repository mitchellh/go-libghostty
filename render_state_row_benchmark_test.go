package libghostty

import (
	"strings"
	"testing"
)

// rowTextBenchmarkCase is a line of terminal content. The benchmark fills an
// 80x30 screen with it and reads back the text of every row, which is how a
// program watching a terminal would typically use AppendText.
type rowTextBenchmarkCase struct {
	name string
	line string
}

var rowTextBenchmarkCases = []rowTextBenchmarkCase{
	{
		name: "Plain",
		line: "plain ASCII terminal content with words, numbers 0123456789, and punctuation",
	},
	{
		name: "Sparse",
		line: "short line",
	},
	{
		name: "Mixed",
		line: "\x1b[1;38;5;33mstatus\x1b[0m plain 日本語 é 👩🏽‍💻 \x1b[4munderlined\x1b[0m",
	},
}

// appendRowTextPerCell produces the same text as AppendText using only the
// per-cell API from Go, which is what callers had to do before AppendText
// existed. It looks up the cell width only for empty cells, since those are
// the only cells where it changes the result.
//
// With checkWide false, it is the cheapest per-cell loop possible, with two
// calls into libghostty per cell (Next and AppendGraphemes). That version
// writes an extra space after each wide character, so it is only useful to
// show the lowest cost a per-cell loop can reach.
func appendRowTextPerCell(dst []byte, ri *RenderStateRowIterator, rc *RenderStateRowCells, checkWide bool) ([]byte, error) {
	if err := ri.Cells(rc); err != nil {
		return dst, err
	}

	blanks := 0
	for rc.Next() {
		// Write the pending spaces first. If this cell turns out to be
		// empty too, they are removed again below.
		mark := len(dst)
		for range blanks {
			dst = append(dst, ' ')
		}
		before := len(dst)

		var err error
		dst, err = rc.AppendGraphemes(dst)
		if err != nil {
			return dst, err
		}
		if len(dst) > before {
			blanks = 0
			continue
		}
		dst = dst[:mark]

		if checkWide {
			cell, err := rc.Raw()
			if err != nil {
				return dst, err
			}
			wide, err := cell.Wide()
			if err != nil {
				return dst, err
			}
			if wide == CellWideSpacerTail || wide == CellWideSpacerHead {
				continue
			}
		}
		blanks++
	}
	return dst, nil
}

// BenchmarkRenderStateRowText compares reading the text of a full screen with
// AppendText, which makes one call into libghostty per row, against the
// per-cell loops above, which make two or more calls per cell.
func BenchmarkRenderStateRowText(b *testing.B) {
	methods := []struct {
		name string
		fn   func([]byte, *RenderStateRowIterator, *RenderStateRowCells) ([]byte, error)
	}{
		{
			name: "AppendText",
			fn: func(dst []byte, ri *RenderStateRowIterator, rc *RenderStateRowCells) ([]byte, error) {
				return ri.AppendText(dst, rc)
			},
		},
		{
			name: "PerCell",
			fn: func(dst []byte, ri *RenderStateRowIterator, rc *RenderStateRowCells) ([]byte, error) {
				return appendRowTextPerCell(dst, ri, rc, true)
			},
		},
		{
			name: "PerCellMinimal",
			fn: func(dst []byte, ri *RenderStateRowIterator, rc *RenderStateRowCells) ([]byte, error) {
				return appendRowTextPerCell(dst, ri, rc, false)
			},
		},
	}

	for _, tc := range rowTextBenchmarkCases {
		for _, m := range methods {
			b.Run(tc.name+"/"+m.name, func(b *testing.B) {
				// Fill an 80x30 screen with the line on every row.
				term, err := NewTerminal(WithSize(80, 30), WithMaxScrollbackLines(0))
				if err != nil {
					b.Fatal(err)
				}
				defer term.Close()
				term.VTWrite([]byte(strings.Repeat(tc.line+"\r\n", 29) + tc.line))

				rs, err := NewRenderState()
				if err != nil {
					b.Fatal(err)
				}
				defer rs.Close()
				if err := rs.Update(term); err != nil {
					b.Fatal(err)
				}

				ri, err := NewRenderStateRowIterator()
				if err != nil {
					b.Fatal(err)
				}
				defer ri.Close()

				rc, err := NewRenderStateRowCells()
				if err != nil {
					b.Fatal(err)
				}
				defer rc.Close()

				// Make sure AppendText and the correct per-cell loop produce
				// the same text before timing anything.
				if err := rs.RowIterator(ri); err != nil {
					b.Fatal(err)
				}
				for ri.Next() {
					want, err := ri.AppendText(nil, rc)
					if err != nil {
						b.Fatal(err)
					}
					got, err := appendRowTextPerCell(nil, ri, rc, true)
					if err != nil {
						b.Fatal(err)
					}
					if string(want) != string(got) {
						b.Fatalf("methods disagree: %q vs %q", want, got)
					}
				}

				buf := make([]byte, 0, 4096)
				b.ReportAllocs()
				for b.Loop() {
					if err := rs.RowIterator(ri); err != nil {
						b.Fatal(err)
					}
					for ri.Next() {
						buf, err = m.fn(buf[:0], ri, rc)
						if err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}
