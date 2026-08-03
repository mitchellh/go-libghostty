package libghostty

import (
	"strconv"
	"testing"
)

type snapshotBenchmarkDecoder func([]byte) (*SnapshotDecoder, error)

// snapshotBenchmarkCase describes terminal state prepared outside the timed
// benchmark region. The cases deliberately vary both the number of retained
// rows and the cell representation: ASCII exercises the compact common path,
// while Unicode and mixed VT input exercise grapheme storage and styled cells.
type snapshotBenchmarkCase struct {
	name             string
	line             string
	lines            int
	maxScrollback    uint
	expectScrollback bool
}

var snapshotBenchmarkCases = []snapshotBenchmarkCase{
	{
		name:          "EmptyNoScrollback",
		maxScrollback: 0,
	},
	{
		name:          "PlainNoScrollback",
		line:          "plain ASCII terminal content with words, numbers 0123456789, and punctuation",
		lines:         12,
		maxScrollback: 0,
	},
	{
		name:          "UnicodeNoScrollback",
		line:          "日本語 العربية e\u0301moji 👩🏽‍💻 🚀 text across scripts",
		lines:         12,
		maxScrollback: 0,
	},
	{
		name:          "MixedNoScrollback",
		line:          "\x1b[1;38;5;33mstatus\x1b[0m plain 日本語 e\u0301 👩🏽‍💻 \x1b[4munderlined\x1b[0m",
		lines:         12,
		maxScrollback: 0,
	},
	{
		name:             "Plain10KScrollback",
		line:             "plain ASCII history with words, numbers 0123456789, and punctuation",
		lines:            10_024,
		maxScrollback:    10_000,
		expectScrollback: true,
	},
	{
		name:             "Mixed10KScrollback",
		line:             "\x1b[1;38;5;33mstatus\x1b[0m plain 日本語 e\u0301 👩🏽‍💻 \x1b[4munderlined\x1b[0m",
		lines:            10_024,
		maxScrollback:    10_000,
		expectScrollback: true,
	},
}

// BenchmarkSnapshotEncode measures both Go-owned snapshot allocation and the
// zero-allocation steady-state API for callers that retain a correctly sized
// buffer. Snapshot construction and buffer sizing are benchmark setup work.
func BenchmarkSnapshotEncode(b *testing.B) {
	for _, test := range snapshotBenchmarkCases {
		b.Run(test.name, func(b *testing.B) {
			term, snapshot := newSnapshotBenchmarkState(b, test)

			b.Run("Allocate", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(snapshot)))
				b.ReportMetric(float64(len(snapshot)), "snapshot-B")

				var encoded []byte
				for b.Loop() {
					var err error
					encoded, err = term.Snapshot()
					if err != nil {
						b.Fatal(err)
					}
				}
				snapshotBenchmarkBytes = encoded
			})

			b.Run("ReuseBuffer", func(b *testing.B) {
				buf := make([]byte, len(snapshot))
				b.ReportAllocs()
				b.SetBytes(int64(len(snapshot)))
				b.ReportMetric(float64(len(snapshot)), "snapshot-B")

				var written int
				for b.Loop() {
					var err error
					written, err = term.SnapshotBuf(buf)
					if err != nil {
						b.Fatal(err)
					}
				}
				snapshotBenchmarkInt = written
			})
		})
	}
}

// BenchmarkSnapshotDecode measures complete restoration and the READY-prefix
// latency exposed for applications that can show a terminal before its older
// history pages have been restored. Each mode compares the normal zero-copy
// byte source with the mutation-safe copying constructor so the input-copy
// cost remains visible as snapshots grow.
func BenchmarkSnapshotDecode(b *testing.B) {
	for _, test := range snapshotBenchmarkCases {
		b.Run(test.name, func(b *testing.B) {
			_, snapshot := newSnapshotBenchmarkState(b, test)

			for _, source := range []struct {
				name string
				new  snapshotBenchmarkDecoder
			}{
				{name: "ZeroCopy", new: NewSnapshotDecoderBytes},
				{name: "Copy", new: NewSnapshotDecoderBytesCopy},
			} {
				b.Run("Full/"+source.name, func(b *testing.B) {
					benchmarkSnapshotDecode(b, snapshot, source.new, false)
				})
				b.Run("Ready/"+source.name, func(b *testing.B) {
					benchmarkSnapshotDecode(b, snapshot, source.new, true)
				})
			}
		})
	}
}

// benchmarkSnapshotDecode owns the complete per-iteration lifecycle. READY
// terminals must outlive their decoders until the decoder is closed.
func benchmarkSnapshotDecode(
	b *testing.B,
	snapshot []byte,
	newDecoder snapshotBenchmarkDecoder,
	readyOnly bool,
) {
	b.Helper()
	b.ReportAllocs()
	processedBytes := len(snapshot)
	if readyOnly {
		processedBytes = snapshotBenchmarkReadyPrefixLen(b, snapshot)
	}
	b.SetBytes(int64(processedBytes))
	b.ReportMetric(float64(len(snapshot)), "snapshot-B")

	for b.Loop() {
		decoder, err := newDecoder(snapshot)
		if err != nil {
			b.Fatal(err)
		}
		var restored *Terminal
		if readyOnly {
			restored, err = decoder.Ready()
		} else {
			restored, err = decoder.Decode()
		}
		if err != nil {
			decoder.Close()
			b.Fatal(err)
		}
		decoder.Close()
		restored.Close()
	}
}

// snapshotBenchmarkReadyPrefixLen reports only the bytes consumed through the
// READY checkpoint. Using the complete snapshot size for SetBytes would make
// the READY throughput metric misleading for large-history cases.
func snapshotBenchmarkReadyPrefixLen(b *testing.B, snapshot []byte) int {
	b.Helper()

	decoder, err := NewSnapshotDecoderBytes(snapshot)
	if err != nil {
		b.Fatal(err)
	}
	restored, err := decoder.Ready()
	if err != nil {
		decoder.Close()
		b.Fatal(err)
	}
	offset, err := decoder.SourceOffset()
	decoder.Close()
	restored.Close()
	if err != nil {
		b.Fatal(err)
	}
	if offset > uint(len(snapshot)) {
		b.Fatalf("READY source offset %d exceeds snapshot length %d", offset, len(snapshot))
	}
	return int(offset)
}

// newSnapshotBenchmarkState builds and validates a scenario before benchmark
// timing begins. A varying decimal prefix prevents the large-history cases
// from collapsing to a single repeated row representation.
func newSnapshotBenchmarkState(b *testing.B, test snapshotBenchmarkCase) (*Terminal, []byte) {
	b.Helper()

	term, err := NewTerminal(
		WithSize(80, 24),
		// The byte limit otherwise becomes the effective bound before the
		// requested 10K-line limit for Unicode and densely styled rows.
		WithMaxScrollbackBytes(256<<20),
		WithMaxScrollbackLines(test.maxScrollback),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(term.Close)

	if test.lines > 0 {
		input := make([]byte, 0, test.lines*(len(test.line)+16))
		for i := range test.lines {
			input = strconv.AppendInt(input, int64(i), 10)
			input = append(input, ':', ' ')
			input = append(input, test.line...)
			if i+1 < test.lines {
				input = append(input, '\r', '\n')
			}
		}
		term.VTWrite(input)
	}

	encoded, err := term.Snapshot()
	if err != nil {
		b.Fatal(err)
	}
	if len(encoded) == 0 {
		b.Fatal("snapshot encoding returned no bytes")
	}
	validateSnapshotBenchmarkHistory(b, encoded, test.expectScrollback)
	return term, encoded
}

// validateSnapshotBenchmarkHistory catches fixture regressions that would
// accidentally turn a 10K-scrollback measurement into another screen-only
// measurement.
func validateSnapshotBenchmarkHistory(b *testing.B, encoded []byte, expected bool) {
	b.Helper()

	decoder, err := NewSnapshotDecoderBytes(encoded)
	if err != nil {
		b.Fatal(err)
	}
	restored, err := decoder.Ready()
	if err != nil {
		decoder.Close()
		b.Fatal(err)
	}
	historyRows, err := decoder.HistoryRows(ScreenPrimary)
	decoder.Close()
	restored.Close()
	if err != nil {
		b.Fatal(err)
	}
	if expected && historyRows < 9_000 {
		b.Fatalf("10K-scrollback fixture retained only %d logical rows", historyRows)
	}
	if !expected && historyRows > 24 {
		b.Fatalf("no-scrollback fixture unexpectedly retained %d logical rows", historyRows)
	}
}

var (
	snapshotBenchmarkBytes []byte
	snapshotBenchmarkInt   int
)
