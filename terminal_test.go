package libghostty

import (
	"fmt"
	"io"
	"testing"
)

func TestNewTerminalClose(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	term.Close()
}

func TestNewTerminalZeroDimensions(t *testing.T) {
	_, err := NewTerminal(WithSize(0, 24))
	if err == nil {
		t.Fatal("expected error for zero cols")
	}

	_, err = NewTerminal(WithSize(80, 0))
	if err == nil {
		t.Fatal("expected error for zero rows")
	}
}

func TestNewTerminalNoOptions(t *testing.T) {
	_, err := NewTerminal()
	if err == nil {
		t.Fatal("expected error when no size is specified")
	}
}

func TestNewTerminalWithScrollback(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24), WithMaxScrollbackLines(1000))
	if err != nil {
		t.Fatal(err)
	}
	term.Close()
}

func TestTerminalReset(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write some data then reset; should not panic or error.
	term.VTWrite([]byte("hello"))
	term.Reset()
}

func TestTerminalResize(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.Resize(120, 40, 8, 16); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalResizeZero(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.Resize(0, 24, 8, 16); err == nil {
		t.Fatal("expected error for zero cols")
	}
}

func TestTerminalVTWrite(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write plain text and escape sequences; should not panic.
	term.VTWrite([]byte("hello world"))
	term.VTWrite([]byte("\x1b[2J")) // clear screen
	term.VTWrite(nil)               // empty write
}

func TestTerminalIOWriter(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Terminal satisfies io.Writer.
	var w io.Writer = term
	n, err := fmt.Fprintf(w, "hello %s", "world")
	if err != nil {
		t.Fatal(err)
	}
	if n != len("hello world") {
		t.Fatalf("expected %d bytes written, got %d", len("hello world"), n)
	}
}

func TestTerminalMode(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Test several modes: set via SetMode, read back via Mode.
	tests := []struct {
		name       string
		mode       Mode
		defaultVal bool
	}{
		{"CursorVisible", ModeCursorVisible, true},
		{"Wraparound", ModeWraparound, true},
		{"BackarrowKeyMode", ModeBackarrowKeyMode, false},
		{"BracketedPaste", ModeBracketedPaste, false},
		{"FocusEvent", ModeFocusEvent, false},
		{"AltScreen", ModeAltScreen, false},
		{"Origin", ModeOrigin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify default value.
			val, err := term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if val != tt.defaultVal {
				t.Fatalf("expected default %v, got %v", tt.defaultVal, val)
			}

			// Toggle the mode.
			if err := term.SetMode(tt.mode, !tt.defaultVal); err != nil {
				t.Fatal(err)
			}
			val, err = term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if val != !tt.defaultVal {
				t.Fatalf("expected %v after set, got %v", !tt.defaultVal, val)
			}

			// Toggle back.
			if err := term.SetMode(tt.mode, tt.defaultVal); err != nil {
				t.Fatal(err)
			}
			val, err = term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if val != tt.defaultVal {
				t.Fatalf("expected %v after restore, got %v", tt.defaultVal, val)
			}
		})
	}
}

func TestTerminalModeVTWrite(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Test that VT escape sequences correctly set and reset modes
	// and that Mode reads them back.
	// DEC private modes use CSI ? <n> h (set) / CSI ? <n> l (reset).
	// ANSI modes use CSI <n> h (set) / CSI <n> l (reset).
	tests := []struct {
		name       string
		mode       Mode
		setSeq     string
		resetSeq   string
		defaultVal bool
	}{
		{"BracketedPaste", ModeBracketedPaste, "\x1b[?2004h", "\x1b[?2004l", false},
		{"BackarrowKeyMode", ModeBackarrowKeyMode, "\x1b[?67h", "\x1b[?67l", false},
		{"CursorVisible", ModeCursorVisible, "\x1b[?25h", "\x1b[?25l", true},
		{"FocusEvent", ModeFocusEvent, "\x1b[?1004h", "\x1b[?1004l", false},
		{"NormalMouse", ModeNormalMouse, "\x1b[?1000h", "\x1b[?1000l", false},
		{"SGRMouse", ModeSGRMouse, "\x1b[?1006h", "\x1b[?1006l", false},
		{"Insert", ModeInsert, "\x1b[4h", "\x1b[4l", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify default.
			val, err := term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if val != tt.defaultVal {
				t.Fatalf("expected default %v, got %v", tt.defaultVal, val)
			}

			// Set via VT escape sequence.
			term.VTWrite([]byte(tt.setSeq))
			val, err = term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if !val {
				t.Fatal("expected mode set after VT write set sequence")
			}

			// Reset via VT escape sequence.
			term.VTWrite([]byte(tt.resetSeq))
			val, err = term.Mode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if val {
				t.Fatal("expected mode reset after VT write reset sequence")
			}

			// Restore to default for next subtest.
			if tt.defaultVal {
				term.VTWrite([]byte(tt.setSeq))
			}
		})
	}
}
