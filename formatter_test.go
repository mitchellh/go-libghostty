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

// shortFormatterWriter accepts one byte per call to verify that the shared
// GhosttyWriter bridge retries short writes until it consumes each callback.
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
	if out1 != "" {
		t.Fatalf("expected empty terminal to format empty string, got %q", out1)
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
		WithFormatterExtras(FormatterExtraCursor),
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

func TestFormatterWriteToLargerThanBridgeBuffer(t *testing.T) {
	const (
		cols = 240
		rows = 80
	)
	term, err := NewTerminal(WithSize(cols, rows), WithMaxScrollbackLines(0))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	line := bytes.Repeat([]byte{'x'}, cols-1)
	input := make([]byte, 0, rows*(len(line)+2))
	for row := range rows {
		input = append(input, line...)
		if row+1 < rows {
			input = append(input, '\r', '\n')
		}
	}
	term.VTWrite(input)

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatVT))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	want, err := f.Format()
	if err != nil {
		t.Fatal(err)
	}
	if len(want) <= 16<<10 {
		t.Fatalf("test output is only %d bytes; expected multiple bridge-buffer chunks", len(want))
	}

	var got bytes.Buffer
	written, err := f.WriteTo(&got)
	if err != nil {
		t.Fatal(err)
	}
	if written != int64(len(want)) {
		t.Fatalf("WriteTo returned %d bytes, want %d", written, len(want))
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Fatal("buffered WriteTo output differs from Format output")
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
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("expected four accepted bytes, got %d", n)
	}
	if output.String() != "abcd" {
		t.Fatalf("expected accepted output %q, got %q", "abcd", output.String())
	}
}

func TestFormatterWriteToError(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("formatter writer error"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatPlain))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	wantErr := errors.New("formatter sink failed")
	w := &failingWriter{limit: 3, err: wantErr}
	written, err := f.WriteTo(w)
	if written != 3 {
		t.Fatalf("expected three accepted bytes, got %d", written)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original writer error, got %v", err)
	}
	assertResultError(t, err, ResultIOError)

	// A later call must not retain the previous writer error.
	var recovered bytes.Buffer
	written, err = f.WriteTo(&recovered)
	if err != nil {
		t.Fatal(err)
	}
	if written != int64(recovered.Len()) || !strings.Contains(recovered.String(), "formatter writer error") {
		t.Fatalf("unexpected output after writer recovery: written=%d output=%q", written, recovered.String())
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

// TestFormatterDefaults pins the options a new formatter starts with. These are
// libghostty's defaults, and NewFormatter doesn't change them.
func TestFormatterDefaults(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	f, err := NewFormatter(term)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if got, err := f.Emit(); err != nil || got != FormatterFormatPlain {
		t.Fatalf("Emit() = %v, %v", got, err)
	}
	if got, err := f.Unwrap(); err != nil || got {
		t.Fatalf("Unwrap() = %v, %v", got, err)
	}
	if got, err := f.Trim(); err != nil || !got {
		t.Fatalf("Trim() = %v, %v", got, err)
	}
	if got, err := f.Extras(); err != nil || got != FormatterExtrasDefault {
		t.Fatalf("Extras() = %#x, %v", got, err)
	}
	if got, err := f.Selection(); err != nil || got != nil {
		t.Fatalf("Selection() = %v, %v", got, err)
	}
}

// TestFormatterSetters checks that each option reads back what was set and that
// a change applies to the next format call on the same formatter.
func TestFormatterSetters(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("\x1b[1mbold\x1b[0m   "))

	f, err := NewFormatter(term)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	plain, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if plain != "bold" {
		t.Fatalf("expected trimmed plain output, got %q", plain)
	}

	if err := f.SetEmit(FormatterFormatVT); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Emit(); err != nil || got != FormatterFormatVT {
		t.Fatalf("Emit() = %v, %v", got, err)
	}
	styled, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(styled, "\x1b[") {
		t.Fatalf("expected VT output after SetEmit, got %q", styled)
	}

	if err := f.SetEmit(FormatterFormatPlain); err != nil {
		t.Fatal(err)
	}
	if err := f.SetTrim(false); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Trim(); err != nil || got {
		t.Fatalf("Trim() = %v, %v", got, err)
	}
	untrimmed, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(untrimmed, "bold   ") {
		t.Fatalf("expected untrimmed output, got %q", untrimmed)
	}

	if err := f.SetUnwrap(true); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Unwrap(); err != nil || !got {
		t.Fatalf("Unwrap() = %v, %v", got, err)
	}
}

// TestFormatterExtras checks that extras combine as a bitmask and take effect:
// the cursor extra is the one that adds a CUP sequence to VT output.
func TestFormatterExtras(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("hi"))

	f, err := NewFormatter(
		term,
		WithFormatterFormat(FormatterFormatVT),
		WithFormatterExtras(FormatterExtrasNone),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	without, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(without, "H") {
		t.Fatalf("expected no cursor sequence, got %q", without)
	}

	extras := FormatterExtraCursor | FormatterExtraPalette
	if err := f.SetExtras(extras); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Extras(); err != nil || got != extras {
		t.Fatalf("Extras() = %#x, %v", got, err)
	}
	with, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(with, "\x1b[1;3H") {
		t.Fatalf("expected trailing cursor position, got %q", with)
	}

	// A bit with no meaning is rejected and keeps the previous value.
	if err := f.SetExtras(FormatterExtrasAll + 1); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if got, err := f.Extras(); err != nil || got != extras {
		t.Fatalf("Extras() after rejected set = %#x, %v", got, err)
	}
}

// TestFormatterSelectionOption checks setting, reading back, and clearing a
// selection on an existing formatter.
func TestFormatterSelectionOption(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("hello world"))

	ref, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	word, err := term.SelectWord(SelectWordOptions{Ref: ref})
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewFormatter(term)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := f.SetSelection(word); err != nil {
		t.Fatal(err)
	}
	if got, err := f.FormatString(); err != nil || got != "hello" {
		t.Fatalf("FormatString() = %q, %v", got, err)
	}
	sel, err := f.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel == nil {
		t.Fatal("expected the selection to read back")
	}
	equal, err := term.SelectionEqual(sel, word)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("expected the selection to read back unchanged")
	}

	if err := f.SetSelection(nil); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Selection(); err != nil || got != nil {
		t.Fatalf("Selection() after clear = %v, %v", got, err)
	}
	if got, err := f.FormatString(); err != nil || got != "hello world" {
		t.Fatalf("FormatString() after clear = %q, %v", got, err)
	}
}

// TestNewFormatterOptionError checks that an option libghostty rejects fails
// construction rather than leaving a half-configured formatter behind.
func TestNewFormatterOptionError(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormat(99)))
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if f != nil {
		t.Fatal("expected no formatter")
	}
}

// TestFormatterHTMLPaletteExtra pins what the palette extra means for HTML:
// palette colors are CSS variables, and the extra is what defines them.
func TestFormatterHTMLPaletteExtra(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	term.VTWrite([]byte("\x1b[31mred\x1b[0m"))

	f, err := NewFormatter(term, WithFormatterFormat(FormatterFormatHTML))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	with, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(with, "--vt-palette-1:") {
		t.Fatalf("expected the palette to be defined, got %q", with)
	}

	if err := f.SetExtras(FormatterExtrasNone); err != nil {
		t.Fatal(err)
	}
	without, err := f.FormatString()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(without, "--vt-palette-1:") {
		t.Fatalf("expected no palette definitions, got %q", without)
	}
	if !strings.Contains(without, "var(--vt-palette-1)") {
		t.Fatalf("expected the cell to refer to the palette, got %q", without)
	}
}
