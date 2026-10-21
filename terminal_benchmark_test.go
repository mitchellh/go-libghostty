package libghostty

import (
	"bytes"
	"strings"
	"testing"
)

// vtBenchmarkCase is one VT input workload. Each workload is written in
// full every benchmark iteration, so ns/op and MB/s compare directly
// between [Terminal.VTWrite] and [Terminal.VTWriteUntil].
type vtBenchmarkCase struct {
	name string
	data []byte
}

// vtBenchmarkCases cover the shapes VTWriteUntil callers see: plain output
// with no boundaries at all, output dense with short OSC strings (shell
// integration prompts and titles), and large APC payloads (Kitty
// graphics) where the boundaries are far apart.
var vtBenchmarkCases = []vtBenchmarkCase{
	{
		name: "Plain",
		data: bytes.Repeat(
			[]byte("plain ASCII output with words, numbers 0123456789, and punctuation\r\n"),
			1024,
		),
	},
	{
		name: "OSCDense",
		data: bytes.Repeat(
			[]byte("\x1b]133;A\x07$ \x1b]133;B\x07ls\r\n\x1b]133;C\x07file\r\n\x1b]133;D;0\x07\x1b]2;title\x1b\\"),
			1024,
		),
	},
	{
		name: "APCLarge",
		data: bytes.Repeat(
			[]byte("\x1b_Ga=t,q=2,m=1;"+strings.Repeat("QUFBQUFBQUFB", 512)+"\x1b\\text\r\n"),
			16,
		),
	},
}

// vtBenchmarkBoundaries are the boundary sets benchmarked for each case.
var vtBenchmarkBoundaries = []struct {
	name string
	want TerminalBoundary
}{
	{"APC", TerminalBoundaryAPCStart | TerminalBoundaryAPCEnd},
	{"OSC", TerminalBoundaryOSCStart | TerminalBoundaryOSCEnd},
	{"All", TerminalBoundaryAPCStart | TerminalBoundaryAPCEnd |
		TerminalBoundaryDCSStart | TerminalBoundaryDCSEnd |
		TerminalBoundaryOSCStart | TerminalBoundaryOSCEnd},
}

// newVTBenchmarkTerminal creates a terminal without scrollback so that
// repeated iterations don't grow memory and skew later iterations.
func newVTBenchmarkTerminal(b *testing.B) *Terminal {
	b.Helper()

	term, err := NewTerminal(WithSize(120, 40), WithMaxScrollbackLines(0))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(term.Close)
	return term
}

// BenchmarkVTWrite is the baseline for BenchmarkVTWriteUntil: the same
// workloads parsed in one call with no boundary tracking.
func BenchmarkVTWrite(b *testing.B) {
	for _, tc := range vtBenchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			term := newVTBenchmarkTerminal(b)
			b.SetBytes(int64(len(tc.data)))
			b.ReportAllocs()
			for b.Loop() {
				term.VTWrite(tc.data)
			}
		})
	}
}

// BenchmarkVTWriteUntil feeds each workload through VTWriteUntil, calling
// it again after every stop until the workload is fully consumed. The
// stops/op metric shows how many cgo calls each iteration made, so the
// difference from BenchmarkVTWrite divided by stops/op is the per-stop
// overhead.
func BenchmarkVTWriteUntil(b *testing.B) {
	for _, tc := range vtBenchmarkCases {
		for _, bc := range vtBenchmarkBoundaries {
			b.Run(tc.name+"/"+bc.name, func(b *testing.B) {
				term := newVTBenchmarkTerminal(b)
				b.SetBytes(int64(len(tc.data)))
				b.ReportAllocs()

				stops := 0
				for b.Loop() {
					data := tc.data
					for len(data) > 0 {
						stop, err := term.VTWriteUntil(data, bc.want)
						if err != nil {
							b.Fatal(err)
						}
						data = data[stop.Consumed:]
						if stop.Reached != 0 {
							stops++
						}
					}
				}
				b.ReportMetric(float64(stops)/float64(b.N), "stops/op")
			})
		}
	}
}

// BenchmarkVTWriteUntilCall measures the fixed cost of one VTWriteUntil
// call. The terminal stays at ground and the call asks for ground, so it
// returns immediately without parsing anything.
func BenchmarkVTWriteUntilCall(b *testing.B) {
	term := newVTBenchmarkTerminal(b)
	data := []byte("x")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := term.VTWriteUntil(data, TerminalBoundaryGround); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVTWriteCall measures the fixed cost of one VTWrite call with a
// single printable byte, for comparison with BenchmarkVTWriteUntilCall.
func BenchmarkVTWriteCall(b *testing.B) {
	term := newVTBenchmarkTerminal(b)
	data := []byte("x")
	b.ReportAllocs()
	for b.Loop() {
		term.VTWrite(data)
	}
}
