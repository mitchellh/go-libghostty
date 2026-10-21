package libghostty

import (
	"fmt"
	"strings"
	"testing"
)

// The benchmarks in this file measure RenderState.Update, which a renderer
// calls once per frame. The render state works out the position of every
// Kitty image on screen during each update, so the scenes below cover
// terminals without images, with many fixed placements spread through the
// scrollback, and with virtual placements shown by placeholder cells.

// renderStateBenchmarkScene describes the terminal contents a benchmark
// starts from.
type renderStateBenchmarkScene struct {
	name string

	// kitty enables Kitty graphics on the terminal.
	kitty bool

	// setup writes the scene's contents to the terminal.
	setup func(term *Terminal)
}

// renderStateBenchmarkLines is the number of lines of output each scene
// writes, so that the screen is full and the scrollback has history to
// scroll through.
const renderStateBenchmarkLines = 200

// writeBenchmarkText writes renderStateBenchmarkLines lines of text. If
// placeEvery is non-zero, it also places image 1 at the start of every
// placeEvery'th line, as a 4x2 cell fixed placement. That spreads the
// placements through the viewport and the scrollback, like the output of a
// program that prints images inline.
func writeBenchmarkText(term *Terminal, placeEvery int) {
	var sb strings.Builder
	if placeEvery > 0 {
		// Transmit a 1x1 RGB image once, then place it many times.
		sb.WriteString("\x1b_Ga=t,t=d,f=24,i=1,s=1,v=1,q=2;////\x1b\\")
	}
	for i := range renderStateBenchmarkLines {
		if placeEvery > 0 && i%placeEvery == 0 {
			// C=1 keeps the cursor in place so the text overwrites the
			// cells under the image, as it would in a real program.
			fmt.Fprintf(&sb, "\x1b_Ga=p,i=1,p=%d,c=4,r=2,C=1,q=2\x1b\\", i+1)
		}
		fmt.Fprintf(&sb, "line %d: plain terminal output with some words and numbers 0123456789\r\n", i)
	}
	term.VTWrite([]byte(sb.String()))
}

// renderStateBenchmarkRow rewrites one row of the screen. It is a package
// variable so that converting it to bytes isn't counted in each frame.
var renderStateBenchmarkRow = []byte("\x1b[15;1H\x1b[2Kstatus: frame")

var renderStateBenchmarkScenes = []renderStateBenchmarkScene{
	{
		name:  "Text",
		setup: func(term *Terminal) { writeBenchmarkText(term, 0) },
	},
	{
		name:  "TextKitty",
		kitty: true,
		setup: func(term *Terminal) { writeBenchmarkText(term, 0) },
	},
	{
		// About 10 placements in the viewport at a time and about 60 in
		// the scrollback.
		name:  "Fixed",
		kitty: true,
		setup: func(term *Terminal) { writeBenchmarkText(term, 3) },
	},
	{
		// Every line has an image, so every row of the viewport has one.
		name:  "FixedDense",
		kitty: true,
		setup: func(term *Terminal) { writeBenchmarkText(term, 1) },
	},
	{
		// A virtual placement shown by ten rows of ten placeholder cells
		// at the top of the screen, as tmux would show an image. Each row
		// becomes one image in the render state.
		name:  "Virtual",
		kitty: true,
		setup: func(term *Terminal) {
			writeBenchmarkText(term, 0)

			// The Kitty diacritics for rows and columns 0 through 9. The
			// first cell of each row names its row and column 0. The
			// cells after it continue the same row in the next column.
			diacritics := []rune{
				'\u0305', '\u030D', '\u030E', '\u0310', '\u0312',
				'\u033D', '\u033E', '\u033F', '\u0346', '\u034A',
			}

			var sb strings.Builder
			sb.WriteString("\x1b_Ga=T,t=d,f=24,i=2,s=1,v=1,U=1,c=10,r=10,q=2;////\x1b\\")
			sb.WriteString("\x1b[38;5;2m")
			for row := range 10 {
				fmt.Fprintf(&sb, "\x1b[%d;1H\U0010EEEE%c%c%s",
					row+1, diacritics[row], diacritics[0],
					strings.Repeat("\U0010EEEE", 9))
			}
			sb.WriteString("\x1b[0m")
			term.VTWrite([]byte(sb.String()))
		},
	},
}

// newRenderStateBenchmarkTerminal returns an 80x30 terminal with 8x16
// pixel cells and the scene written to it, and a render state already
// updated from it once.
func newRenderStateBenchmarkTerminal(b *testing.B, scene renderStateBenchmarkScene) (*Terminal, *RenderState) {
	b.Helper()

	term, err := NewTerminal(
		WithSize(80, 30),
		WithMaxScrollbackLines(1000),
		WithWritePty(func(*Terminal, []byte) {}),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(term.Close)
	if err := term.Resize(80, 30, 8, 16); err != nil {
		b.Fatal(err)
	}
	if scene.kitty {
		limit := uint64(64 << 20)
		if err := term.SetKittyImageStorageLimit(&limit); err != nil {
			b.Fatal(err)
		}
	}
	scene.setup(term)

	rs, err := NewRenderState()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(rs.Close)
	if err := rs.Update(term); err != nil {
		b.Fatal(err)
	}
	return term, rs
}

// BenchmarkRenderStateUpdate measures one frame's RenderState.Update after
// a typical change:
//
//   - Clean: nothing changed since the last update.
//   - Row: a program rewrote one row of text, the most common frame.
//   - Scroll: the user scrolled the viewport by one row, which moves every
//     row and every image on screen.
func BenchmarkRenderStateUpdate(b *testing.B) {
	frames := []struct {
		name string
		fn   func(term *Terminal, i int)
	}{
		{
			name: "Clean",
			fn:   func(*Terminal, int) {},
		},
		{
			name: "Row",
			fn: func(term *Terminal, i int) {
				term.VTWrite(renderStateBenchmarkRow)
			},
		},
		{
			name: "Scroll",
			fn: func(term *Terminal, i int) {
				// Alternate up and down so the viewport stays in the same
				// region of the scrollback for the whole benchmark.
				if i%2 == 0 {
					term.ScrollViewportDelta(-1)
				} else {
					term.ScrollViewportDelta(1)
				}
			},
		},
	}

	for _, scene := range renderStateBenchmarkScenes {
		for _, frame := range frames {
			b.Run(scene.name+"/"+frame.name, func(b *testing.B) {
				term, rs := newRenderStateBenchmarkTerminal(b, scene)

				// Scroll up a little first, so Scroll has history above and
				// below the viewport to move through.
				if frame.name == "Scroll" {
					term.ScrollViewportDelta(-50)
				}

				b.ReportAllocs()
				i := 0
				for b.Loop() {
					frame.fn(term, i)
					if err := rs.Update(term); err != nil {
						b.Fatal(err)
					}
					i++
				}
			})
		}
	}
}
