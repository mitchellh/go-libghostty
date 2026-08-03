package libghostty

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// chunkWriter accepts at most max bytes per Write call. It verifies that the
// GhosttyWriter bridge retries ordinary short writes until the full callback
// slice has been accepted.
type chunkWriter struct {
	bytes.Buffer
	max int
}

func (w *chunkWriter) Write(p []byte) (int, error) {
	if len(p) > w.max {
		p = p[:w.max]
	}
	return w.Buffer.Write(p)
}

// failingWriter accepts limit total bytes and then returns err. Returning data
// and an error together exercises Go's complete io.Writer contract.
type failingWriter struct {
	limit   int
	written int
	err     error
}

func (w *failingWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.written
	if remaining <= 0 {
		return 0, w.err
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	w.written += len(p)
	if w.written >= w.limit {
		return len(p), w.err
	}
	return len(p), nil
}

func TestTerminalContinuationDisabled(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	limit, err := term.ContinuationMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	if limit != 0 {
		t.Fatalf("expected continuation tracking disabled, got limit %d", limit)
	}

	_, err = term.Continuation()
	assertResultError(t, err, ResultInvalidValue)
}

func TestTerminalContinuationForms(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithContinuationMaxBytes(1024),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	limit, err := term.ContinuationMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	if limit != 1024 {
		t.Fatalf("expected continuation limit 1024, got %d", limit)
	}

	ground, err := term.Continuation()
	if err != nil {
		t.Fatal(err)
	}
	if len(ground) != 0 {
		t.Fatalf("expected empty ground continuation, got %q", ground)
	}

	const unfinished = "\x1b[31"
	term.VTWrite([]byte(unfinished))

	continuation, err := term.Continuation()
	if err != nil {
		t.Fatal(err)
	}
	if string(continuation) != unfinished {
		t.Fatalf("expected continuation %q, got %q", unfinished, continuation)
	}

	required, err := term.ContinuationBuf(nil)
	assertResultError(t, err, ResultOutOfSpace)
	if required != len(unfinished) {
		t.Fatalf("expected required size %d, got %d", len(unfinished), required)
	}

	buf := make([]byte, required)
	written, err := term.ContinuationBuf(buf)
	if err != nil {
		t.Fatal(err)
	}
	if written != len(unfinished) || string(buf[:written]) != unfinished {
		t.Fatalf("unexpected buffer continuation: written=%d data=%q", written, buf[:written])
	}

	w := &chunkWriter{max: 2}
	written64, err := term.ContinuationWriteTo(w)
	if err != nil {
		t.Fatal(err)
	}
	if written64 != int64(len(unfinished)) || w.String() != unfinished {
		t.Fatalf("unexpected writer continuation: written=%d data=%q", written64, w.String())
	}
}

func TestTerminalContinuationWriterError(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithContinuationMaxBytes(1024),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("\x1b[31"))

	wantErr := errors.New("continuation sink failed")
	w := &failingWriter{limit: 2, err: wantErr}
	written, err := term.ContinuationWriteTo(w)
	if written != 2 {
		t.Fatalf("expected 2 accepted bytes, got %d", written)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original writer error, got %v", err)
	}
	assertResultError(t, err, ResultIOError)
}

func TestTerminalContinuationSetAndComplete(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetContinuationMaxBytes(16); err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte{0xF0, 0x9F})
	continuation, err := term.Continuation()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(continuation, []byte{0xF0, 0x9F}) {
		t.Fatalf("unexpected UTF-8 continuation %v", continuation)
	}

	// Complete the four-byte scalar; the stream returns to ground.
	term.VTWrite([]byte{0x98, 0x80})
	continuation, err = term.Continuation()
	if err != nil {
		t.Fatal(err)
	}
	if len(continuation) != 0 {
		t.Fatalf("expected completed stream at ground, got %v", continuation)
	}

	if err := term.SetContinuationMaxBytes(0); err != nil {
		t.Fatal(err)
	}
	_, err = term.ContinuationWriteTo(io.Discard)
	assertResultError(t, err, ResultInvalidValue)
}
