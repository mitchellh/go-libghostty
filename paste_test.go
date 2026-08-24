package libghostty

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestTerminalPasteText(t *testing.T) {
	var output bytes.Buffer
	var requested []string
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output.Write(data)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	written, err := term.Paste(Paste{
		Location: ClipboardLocationStandard,
		Source:   PasteSourceText,
		MIMEs:    []string{"image/png", "text/plain"},
		Reader: func(mime string, writer io.Writer) error {
			requested = append(requested, mime)
			_, err := io.WriteString(writer, "hello")
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected paste to write text")
	}
	if got := output.String(); got != "hello" {
		t.Fatalf("expected pasted text %q, got %q", "hello", got)
	}
	if len(requested) != 1 || requested[0] != "text/plain" {
		t.Fatalf("expected only text/plain to be requested, got %q", requested)
	}
}

func TestTerminalPasteRejectedAndRetry(t *testing.T) {
	var output bytes.Buffer
	var reads int
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output.Write(data)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	paste := Paste{
		Location: ClipboardLocationStandard,
		Source:   PasteSourceClipboard,
		MIMEs:    []string{"text/plain"},
		Reader: func(_ string, writer io.Writer) error {
			reads++
			_, err := io.WriteString(writer, "hello\nworld")
			return err
		},
	}
	written, err := term.Paste(paste)
	if written {
		t.Fatal("expected rejected paste not to write")
	}
	assertResultError(t, err, ResultRejected)
	if output.Len() != 0 {
		t.Fatalf("expected rejected paste to write nothing, got %q", output.Bytes())
	}

	paste.AllowUnsafe = true
	written, err = term.Paste(paste)
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected confirmed paste to write")
	}
	if got := output.String(); got != "hello\rworld" {
		t.Fatalf("expected legacy newline framing, got %q", got)
	}
	if reads != 2 {
		t.Fatalf("expected source to be read once per call, got %d reads", reads)
	}
}

func TestTerminalPasteBracketed(t *testing.T) {
	var output bytes.Buffer
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output.Write(data)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	if err := term.SetMode(ModeBracketedPaste, true); err != nil {
		t.Fatal(err)
	}

	written, err := term.Paste(Paste{
		Source: PasteSourceText,
		MIMEs:  []string{"text/plain"},
		Reader: func(_ string, writer io.Writer) error {
			_, err := io.WriteString(writer, "hello\nworld")
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected bracketed paste to write")
	}
	if want := "\x1b[200~hello\nworld\x1b[201~"; output.String() != want {
		t.Fatalf("expected bracketed paste %q, got %q", want, output.String())
	}
}

func TestTerminalPasteReaderError(t *testing.T) {
	sentinel := errors.New("clipboard unavailable")
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, _ []byte) {}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	written, err := term.Paste(Paste{
		MIMEs: []string{"text/plain"},
		Reader: func(_ string, _ io.Writer) error {
			return sentinel
		},
	})
	if written {
		t.Fatal("expected failed paste not to write")
	}
	assertResultError(t, err, ResultIOError)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected callback error to be preserved, got %v", err)
	}
}

func TestTerminalPasteEmpty(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, _ []byte) {}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	written, err := term.Paste(Paste{})
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("expected empty paste not to write")
	}
}

func TestTerminalPasteKittyEventUsesSecureRandom(t *testing.T) {
	var randomCalls int
	if err := SysSetRandomSecure(func(data []byte) error {
		randomCalls++
		for i := range data {
			data[i] = byte(i + 1)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := SysSetRandomSecure(nil); err != nil {
			t.Errorf("restore platform random source: %v", err)
		}
	}()

	var output strings.Builder
	var reads int
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output.Write(data)
		}),
		WithClipboardRead(func(_ *Terminal, _ ClipboardRead) ClipboardReadReply {
			return ClipboardReadReply{Result: ClipboardReadDenied}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	if err := term.SetMode(ModePasteEvents, true); err != nil {
		t.Fatal(err)
	}

	written, err := term.Paste(Paste{
		Location: ClipboardLocationPrimary,
		Source:   PasteSourceClipboard,
		MIMEs:    []string{"text/plain", "image/png"},
		Reader: func(_ string, _ io.Writer) error {
			reads++
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected Kitty paste event to be written")
	}
	if randomCalls == 0 {
		t.Fatal("expected paste event to request secure random bytes")
	}
	if reads != 0 {
		t.Fatalf("expected paste event not to read content, got %d reads", reads)
	}
	if got := output.String(); !strings.HasPrefix(got, "\x1b]5522;type=read:status=OK:loc=primary:pw=") {
		t.Fatalf("unexpected Kitty paste event %q", got)
	}
}

func TestPasteIsSafe(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{name: "empty", data: "", want: true},
		{name: "plain", data: "hello", want: true},
		{name: "newline", data: "hello\nworld", want: false},
		{name: "bracketed paste end", data: "hello\x1b[201~world", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PasteIsSafe([]byte(tt.data)); got != tt.want {
				t.Fatalf("PasteIsSafe(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestPasteEncodeLegacy(t *testing.T) {
	in := []byte("hello\nworld\x1b!")

	out, err := PasteEncode(in, false)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != "hello\rworld !" {
		t.Fatalf("PasteEncode legacy = %q, want %q", out, "hello\rworld !")
	}
	if string(in) != "hello\nworld\x1b!" {
		t.Fatalf("PasteEncode mutated input to %q", in)
	}
}

func TestPasteEncodeBracketed(t *testing.T) {
	out, err := PasteEncode([]byte("hello\nworld"), true)
	if err != nil {
		t.Fatal(err)
	}

	want := "\x1b[200~hello\nworld\x1b[201~"
	if string(out) != want {
		t.Fatalf("PasteEncode bracketed = %q, want %q", out, want)
	}
}

func TestPasteEncodeEmpty(t *testing.T) {
	out, err := PasteEncode(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatalf("PasteEncode empty = %q, want nil", out)
	}
}
