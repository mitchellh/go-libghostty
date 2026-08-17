package libghostty

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// Verify interface satisfaction at compile time.
var _ io.WriterTo = (*Formatter)(nil)

// shortFormatterWriter simulates a broken io.Writer that accepts only one
// byte without reporting the required error.
type shortFormatterWriter struct {
	bytes.Buffer
}

func (w *shortFormatterWriter) Write(p []byte) (int, error) {
	return w.Buffer.Write(p[:1])
}

func TestFormatterPlainText(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("hello world"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected output to contain 'hello world', got %q", out)
	}
}

func TestFormatterVT(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write red text.
	term.VTWrite([]byte("\x1b[31mred text\x1b[0m"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatVT))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "red text") {
		t.Fatalf("expected output to contain 'red text', got %q", out)
	}
	// VT format should contain escape sequences.
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected VT output to contain escape sequences, got %q", out)
	}
}

func TestFormatterHTML(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("\x1b[1mbold text\x1b[0m"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatHTML))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "bold text") {
		t.Fatalf("expected output to contain 'bold text', got %q", out)
	}
}

func TestFormatterFormat(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("bytes test"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := f.Format()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "bytes test") {
		t.Fatalf("expected output to contain 'bytes test', got %q", string(out))
	}
}

func TestFormatterReflectsCurrentState(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Format before writing anything.
	out1, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}

	// Write some text and format again.
	term.VTWrite([]byte("after write"))
	out2, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out2, "after write") {
		t.Fatalf("expected second format to contain 'after write', got %q", out2)
	}
	if strings.Contains(out1, "after write") {
		t.Fatal("first format should not contain text written afterward")
	}
}

func TestFormatterSelection(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("first\r\nsecond\r\nthird"))

	secondRef, err := term.GridRef(Point{Tag: PointTagActive, X: 2, Y: 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := term.SelectLine(SelectLineOptions{Ref: secondRef})
	if err != nil {
		t.Fatal(err)
	}
	if second == nil {
		t.Fatal("expected second-line selection")
	}

	f, err := NewFormatter(
		term,
		WithFormatterFormat(FormatterFormatVT),
		WithFormatterTrim(true),
		WithFormatterSelection(second),
		WithFormatterExtraCursor(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// The native formatter copies the selection during construction. Changing
	// the caller's Go value afterward must not change the formatter's range.
	firstRef, err := term.GridRef(Point{Tag: PointTagActive, X: 2, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	first, err := term.SelectLine(SelectLineOptions{Ref: firstRef})
	if err != nil {
		t.Fatal(err)
	}
	if first == nil {
		t.Fatal("expected first-line selection")
	}
	*second = *first

	out, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "second") {
		t.Fatalf("expected selected line in formatter output, got %q", out)
	}
	if strings.Contains(out, "first") || strings.Contains(out, "third") {
		t.Fatalf("expected output to be restricted to the selected line, got %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected VT terminal extras alongside selected output, got %q", out)
	}
}

func TestFormatterWriteTo(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("writeto test"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Formatter satisfies io.WriterTo.
	var wt io.WriterTo = f
	var buf bytes.Buffer
	n, err := wt.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(buf.Len()) {
		t.Fatalf("WriteTo returned %d but buffer has %d bytes", n, buf.Len())
	}
	if !strings.Contains(buf.String(), "writeto test") {
		t.Fatalf("expected output to contain 'writeto test', got %q", buf.String())
	}
}

func TestFormatterWriteToShortWrite(t *testing.T) {
	term, err := NewTerminal(WithSize(4, 2))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("abcd"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var output shortFormatterWriter
	n, err := f.WriteTo(&output)
	if n != 1 {
		t.Fatalf("expected one accepted byte, got %d", n)
	}
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected io.ErrShortWrite, got %v", err)
	}
	if output.String() != "a" {
		t.Fatalf("expected accepted output %q, got %q", "a", output.String())
	}
}

func TestFormatterFormatBuf(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("buffer format"))

	f, err := NewFormatter(term)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	required, err := f.FormatBuf(nil)
	var resultErr *Error
	if !errors.As(err, &resultErr) || resultErr.Result != ResultOutOfSpace {
		t.Fatalf("expected out-of-space size query, got required=%d err=%v", required, err)
	}

	buf := make([]byte, required)
	written, err := f.FormatBuf(buf)
	if err != nil {
		t.Fatal(err)
	}
	if written != required {
		t.Fatalf("expected %d bytes written, got %d", required, written)
	}
	if !strings.Contains(string(buf[:written]), "buffer format") {
		t.Fatalf("expected formatted contents, got %q", buf[:written])
	}
}
