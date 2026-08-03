package libghostty

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"unsafe"
)

// chunkReader deliberately returns short successful reads to verify that the
// GhosttyReader bridge and snapshot decoder continue until a checkpoint is
// complete.
type chunkReader struct {
	reader *bytes.Reader
	max    int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(p) > r.max {
		p = p[:r.max]
	}
	return r.reader.Read(p)
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func newSnapshotTerminal(t *testing.T) *Terminal {
	t.Helper()

	term, err := NewTerminal(
		WithSize(80, 24),
		WithContinuationMaxBytes(1024),
		WithMaxScrollbackLines(1000),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := term.SetTitle("snapshot title"); err != nil {
		term.Close()
		t.Fatal(err)
	}
	term.VTWrite([]byte("hello "))
	term.VTWrite([]byte("\x1b[31"))
	return term
}

func TestSnapshotEncodeFormsAndDecode(t *testing.T) {
	term := newSnapshotTerminal(t)
	defer term.Close()

	snapshot, err := term.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(snapshot, []byte("GHOSTSNP")) {
		t.Fatalf("snapshot has invalid envelope prefix %q", snapshot[:min(len(snapshot), 8)])
	}

	required, err := term.SnapshotBuf(nil)
	assertResultError(t, err, ResultOutOfSpace)
	if required != len(snapshot) {
		t.Fatalf("size query returned %d, allocated form has %d", required, len(snapshot))
	}

	buf := make([]byte, required)
	written, err := term.SnapshotBuf(buf)
	if err != nil {
		t.Fatal(err)
	}
	if written != required || !bytes.Equal(buf[:written], snapshot) {
		t.Fatal("buffer snapshot differs from allocated snapshot")
	}

	w := &chunkWriter{max: 17}
	written64, err := term.SnapshotWriteTo(w)
	if err != nil {
		t.Fatal(err)
	}
	if written64 != int64(len(snapshot)) || !bytes.Equal(w.Bytes(), snapshot) {
		t.Fatal("writer snapshot differs from allocated snapshot")
	}

	// The byte-slice constructor must copy because C retains the source. Clear
	// the caller's slice after construction to prove the decoder is independent.
	callerBytes := append([]byte(nil), snapshot...)
	decoder, err := NewSnapshotDecoderBytes(callerBytes)
	if err != nil {
		t.Fatal(err)
	}
	defer decoder.Close()
	clear(callerBytes)

	if err := decoder.SetMaxContinuationBytes(4096); err != nil {
		t.Fatal(err)
	}
	limit, err := decoder.MaxContinuationBytes()
	if err != nil {
		t.Fatal(err)
	}
	if limit != 4096 {
		t.Fatalf("expected decoder continuation limit 4096, got %d", limit)
	}
	offset, err := decoder.SourceOffset()
	if err != nil {
		t.Fatal(err)
	}
	if offset != 0 {
		t.Fatalf("expected initial source offset 0, got %d", offset)
	}

	restored, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	offset, err = decoder.SourceOffset()
	if err != nil {
		t.Fatal(err)
	}
	if offset != uint(len(snapshot)) {
		t.Fatalf("expected source offset %d, got %d", len(snapshot), offset)
	}
	continuationLimit, err := restored.ContinuationMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	if continuationLimit != 0 {
		t.Fatalf("decoded terminal should have continuation tracking disabled, got %d", continuationLimit)
	}
	title, err := restored.Title()
	if err != nil {
		t.Fatal(err)
	}
	if title != "snapshot title" {
		t.Fatalf("expected restored title, got %q", title)
	}

	// Finish the SGR sequence that was incomplete when the snapshot was taken.
	// Correctly rendering the following text proves parser continuation state
	// was restored, not merely the visible grid.
	restored.VTWrite([]byte("mred"))
	formatter, err := NewFormatter(restored)
	if err != nil {
		t.Fatal(err)
	}
	defer formatter.Close()
	formatted, err := formatter.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted, "hello red") {
		t.Fatalf("expected restored and resumed content, got %q", formatted)
	}
}

func TestSnapshotIncrementalDecode(t *testing.T) {
	term, err := NewTerminal(
		WithSize(215, 2),
		WithContinuationMaxBytes(1024),
		WithMaxScrollbackLines(5000),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	for range 1000 {
		term.VTWrite([]byte("snapshot history line\r\n"))
	}
	snapshot, err := term.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	decoder, err := NewSnapshotDecoderBytes(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer decoder.Close()

	restored, err := decoder.Ready()
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	historyRows, err := decoder.HistoryRows(ScreenPrimary)
	if err != nil {
		t.Fatal(err)
	}
	if historyRows == 0 {
		t.Fatal("expected primary history in incremental snapshot")
	}
	_, err = decoder.HistoryRows(TerminalScreen(99))
	assertResultError(t, err, ResultInvalidValue)
	_, err = decoder.Progress()
	assertResultError(t, err, ResultNoValue)
	if err := decoder.SetMaxContinuationBytes(1); err == nil {
		t.Fatal("expected lifecycle error when setting decoder option after READY")
	} else {
		assertResultError(t, err, ResultInvalidValue)
	}

	pages := 0
	restoredRows := uint(0)
	for {
		advanced, err := decoder.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !advanced {
			break
		}

		progress, err := decoder.Progress()
		if err != nil {
			t.Fatal(err)
		}
		if progress.Screen != ScreenPrimary && progress.Screen != ScreenAlternate {
			t.Fatalf("invalid progress screen %d", progress.Screen)
		}
		restoredRows += progress.Rows
		if pages == 0 {
			screen, err := decoder.ProgressScreen()
			if err != nil {
				t.Fatal(err)
			}
			rows, err := decoder.ProgressRows()
			if err != nil {
				t.Fatal(err)
			}
			remaining, err := decoder.ProgressRemaining()
			if err != nil {
				t.Fatal(err)
			}
			if screen != progress.Screen || rows != progress.Rows || remaining != progress.Remaining {
				t.Fatalf("individual progress getters differ from combined result: %+v", progress)
			}
		}
		pages++
	}
	if pages == 0 {
		t.Fatal("expected at least one incremental history page")
	}
	if restoredRows == 0 {
		t.Fatal("expected incremental history pages to be applied")
	}

	advanced, err := decoder.Next()
	if err != nil {
		t.Fatal(err)
	}
	if advanced {
		t.Fatal("expected repeated Next after FINISH to remain complete")
	}
	_, err = decoder.Progress()
	assertResultError(t, err, ResultNoValue)

	offset, err := decoder.SourceOffset()
	if err != nil {
		t.Fatal(err)
	}
	if offset != uint(len(snapshot)) {
		t.Fatalf("expected FINISH offset %d, got %d", len(snapshot), offset)
	}
}

func TestSnapshotReaderAndTrailingBytes(t *testing.T) {
	term := newSnapshotTerminal(t)
	defer term.Close()
	snapshot, err := term.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	withTrailing := append(append([]byte(nil), snapshot...), []byte("transport trailer")...)
	reader := &chunkReader{reader: bytes.NewReader(withTrailing), max: 7}
	decoder, err := NewSnapshotDecoder(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer decoder.Close()
	restored, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	offset, err := decoder.SourceOffset()
	if err != nil {
		t.Fatal(err)
	}
	if offset != uint(len(snapshot)) {
		t.Fatalf("expected trailing-byte boundary %d, got %d", len(snapshot), offset)
	}
}

func TestSnapshotReaderAndWriterErrors(t *testing.T) {
	readerErr := errors.New("snapshot source failed")
	decoder, err := NewSnapshotDecoder(errorReader{err: readerErr})
	if err != nil {
		t.Fatal(err)
	}
	_, err = decoder.Decode()
	if !errors.Is(err, readerErr) {
		t.Fatalf("expected original reader error, got %v", err)
	}
	assertResultError(t, err, ResultIOError)
	decoder.Close()

	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	writerErr := errors.New("snapshot sink failed")
	w := &failingWriter{limit: 12, err: writerErr}
	written, err := term.SnapshotWriteTo(w)
	if written != 12 {
		t.Fatalf("expected 12 accepted bytes, got %d", written)
	}
	if !errors.Is(err, writerErr) {
		t.Fatalf("expected original writer error, got %v", err)
	}
	assertResultError(t, err, ResultIOError)
}

func TestSnapshotValidationErrors(t *testing.T) {
	term := newSnapshotTerminal(t)
	defer term.Close()
	snapshot, err := term.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	truncated, err := NewSnapshotDecoderBytes(snapshot[:len(snapshot)-1])
	if err != nil {
		t.Fatal(err)
	}
	_, err = truncated.Decode()
	assertResultError(t, err, ResultInvalidValue)
	truncated.Close()

	limited, err := NewSnapshotDecoderBytes(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := limited.SetMaxContinuationBytes(0); err != nil {
		t.Fatal(err)
	}
	_, err = limited.Decode()
	assertResultError(t, err, ResultLimitExceeded)
	limited.Close()

	empty, err := NewSnapshotDecoderBytes(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	_, err = empty.GetMulti(
		[]SnapshotDecoderData{SnapshotDecoderDataSourceOffset},
		nil,
	)
	if err == nil {
		t.Fatal("expected mismatched keys and values error")
	}
	written, err := empty.GetMulti(nil, nil)
	if err != nil || written != 0 {
		t.Fatalf("empty GetMulti returned written=%d err=%v", written, err)
	}

	// Exercise the raw Get API with its documented C output type.
	var offset uintptr
	if err := empty.Get(
		SnapshotDecoderDataSourceOffset,
		unsafe.Pointer(&offset),
	); err != nil {
		t.Fatal(err)
	}
	if offset != 0 {
		t.Fatalf("expected raw source offset 0, got %d", offset)
	}
	empty.Close()
}

func TestNewSnapshotDecoderNilReader(t *testing.T) {
	_, err := NewSnapshotDecoder(nil)
	assertResultError(t, err, ResultInvalidValue)
}

var _ io.Reader = (*chunkReader)(nil)
var _ io.Writer = (*chunkWriter)(nil)
