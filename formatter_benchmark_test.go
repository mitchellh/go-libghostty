package libghostty

import (
	"io"
	"testing"
)

// formatterBenchmarkCase describes an active-screen shape and cell mix.
type formatterBenchmarkCase struct {
	name string
	cols uint16
	rows uint16
	line string
}

var formatterBenchmarkCases = []formatterBenchmarkCase{
	{
		name: "Empty80x24",
		cols: 80,
		rows: 24,
	},
	{
		name: "Plain80x24",
		cols: 80,
		rows: 24,
		line: "plain ASCII terminal content with words, numbers 0123456789, and punctuation",
	},
	{
		name: "Mixed80x24",
		cols: 80,
		rows: 24,
		line: "\x1b[1;38;5;33mstatus\x1b[0m plain 日本語 e\u0301 👩🏽‍💻 \x1b[4munderlined\x1b[0m",
	},
	{
		name: "Mixed240x80",
		cols: 240,
		rows: 80,
		line: "\x1b[1;38;5;33mstatus\x1b[0m plain 日本語 e\u0301 👩🏽‍💻 \x1b[4munderlined\x1b[0m",
	},
}

// formatterBenchmarkProbe counts bytes and calls outside the timed loop.
type formatterBenchmarkProbe struct {
	bytes int64
	calls int64
}

func (w *formatterBenchmarkProbe) Write(p []byte) (int, error) {
	w.bytes += int64(len(p))
	w.calls++
	return len(p), nil
}

// BenchmarkFormatterVTWriteTo measures VT formatting through WriteTo. It
// excludes terminal setup, output allocation, and sink-side copying.
func BenchmarkFormatterVTWriteTo(b *testing.B) {
	for _, test := range formatterBenchmarkCases {
		b.Run(test.name, func(b *testing.B) {
			formatter := newFormatterBenchmark(b, test)

			var probe formatterBenchmarkProbe
			written, err := formatter.WriteTo(&probe)
			if err != nil {
				b.Fatal(err)
			}
			if written != probe.bytes {
				b.Fatalf("WriteTo reported %d bytes after sink accepted %d", written, probe.bytes)
			}

			b.ReportAllocs()
			b.SetBytes(probe.bytes)
			b.ResetTimer()

			for b.Loop() {
				written, err = formatter.WriteTo(io.Discard)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(probe.bytes), "output-B")
			b.ReportMetric(float64(probe.calls), "callbacks/op")
			formatterBenchmarkWritten = written
		})
	}
}

// newFormatterBenchmark constructs a full active screen without scrollback.
// The last row has no newline so the first row remains on screen.
func newFormatterBenchmark(b *testing.B, test formatterBenchmarkCase) *Formatter {
	b.Helper()

	term, err := NewTerminal(
		WithSize(test.cols, test.rows),
		WithMaxScrollbackLines(0),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(term.Close)

	if test.line != "" {
		input := make([]byte, 0, int(test.rows)*(len(test.line)+2))
		for row := range int(test.rows) {
			input = append(input, test.line...)
			if row+1 < int(test.rows) {
				input = append(input, '\r', '\n')
			}
		}
		term.VTWrite(input)
	}

	formatter, err := NewFormatter(term, WithFormatterFormat(FormatterFormatVT))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(formatter.Close)
	return formatter
}

var formatterBenchmarkWritten int64
