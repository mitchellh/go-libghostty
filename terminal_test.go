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

// vtWriteUntilExpect is the expected outcome of one VTWriteUntil call.
type vtWriteUntilExpect struct {
	reached bool
	stop    TerminalWriteStop
}

// assertVTWriteUntil runs one VTWriteUntil call and compares every field
// of the result, so a misread field in the C-to-Go conversion fails.
func assertVTWriteUntil(
	t *testing.T,
	term *Terminal,
	data []byte,
	want TerminalBoundary,
	expect vtWriteUntilExpect,
) {
	t.Helper()

	stop, err := term.VTWriteUntil(data, want)
	if err != nil {
		t.Fatal(err)
	}
	if stop != expect.stop {
		t.Fatalf("VTWriteUntil(%q, %#x) = %+v, want %+v", data, want, stop, expect.stop)
	}

	// A stop reports a reached boundary exactly when one was expected.
	if (stop.Reached != 0) != expect.reached {
		t.Fatalf("VTWriteUntil(%q, %#x) reached = %v, want %v", data, want, stop.Reached != 0, expect.reached)
	}
}

func TestTerminalVTWriteUntilGround(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// A terminal already at ground leaves the input untouched.
	assertVTWriteUntil(t, term, []byte("untouched"), TerminalBoundaryGround, vtWriteUntilExpect{
		reached: true,
		stop:    TerminalWriteStop{Reached: TerminalBoundaryGround},
	})

	// Complete a split CSI sequence, but stop before processing the printable
	// suffix because the parser reaches ground after the final byte.
	term.VTWrite([]byte("\x1b[31"))
	input := []byte("mABC")
	assertVTWriteUntil(t, term, input, TerminalBoundaryGround, vtWriteUntilExpect{
		reached: true,
		stop:    TerminalWriteStop{Consumed: 1, Reached: TerminalBoundaryGround},
	})
	x, err := term.CursorX()
	if err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("expected suffix to remain unprocessed, cursor is at %d", x)
	}

	term.VTWrite(input[1:])
	x, err = term.CursorX()
	if err != nil {
		t.Fatal(err)
	}
	if x != 3 {
		t.Fatalf("expected suffix write to advance cursor to 3, got %d", x)
	}

	// An empty slice while unfinished consumes nothing but cannot find a
	// ground boundary.
	term.VTWrite([]byte("\x1b["))
	assertVTWriteUntil(t, term, nil, TerminalBoundaryGround, vtWriteUntilExpect{})

	// All input consumed without reaching ground.
	input = []byte("123")
	assertVTWriteUntil(t, term, input, TerminalBoundaryGround, vtWriteUntilExpect{
		stop: TerminalWriteStop{Consumed: len(input)},
	})
}

func TestTerminalVTWriteUntilAPC(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	want := TerminalBoundaryAPCStart | TerminalBoundaryAPCEnd
	data := []byte("ab\x1b_Gq=1\x1b\\cd")

	// The start stop lands on '_' so the content begins with 'G'.
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed:   4,
			Reached:    TerminalBoundaryAPCStart,
			Kind:       TerminalStringKindAPC,
			Introducer: TerminalStringIntroducerESC,
			Byte:       '_',
		},
	})
	data = data[4:]
	if data[0] != 'G' {
		t.Fatalf("expected APC content to start at G, got %q", data[0])
	}

	// The end stop consumes the ESC of ESC \ but not the backslash.
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed: 5,
			Reached:  TerminalBoundaryAPCEnd,
			Kind:     TerminalStringKindAPC,
			End:      TerminalStringEndESC,
			Byte:     0x1b,
		},
	})
	data = data[5:]
	if data[0] != '\\' {
		t.Fatalf("expected backslash to remain unconsumed, got %q", data[0])
	}

	// The rest has no boundary and is fully consumed.
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		stop: TerminalWriteStop{Consumed: len(data)},
	})
	x, err := term.CursorX()
	if err != nil {
		t.Fatal(err)
	}
	if x != 4 {
		t.Fatalf("expected cursor at 4 after printable bytes, got %d", x)
	}
}

func TestTerminalVTWriteUntilOSCAndGround(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Start an OSC without asking for ground, then ask for both. BEL ends
	// the OSC and returns to ground in one byte, so both flags are set.
	want := TerminalBoundaryOSCStart | TerminalBoundaryOSCEnd
	data := []byte("\x1b]2;title\x07")
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed:   2,
			Reached:    TerminalBoundaryOSCStart,
			Kind:       TerminalStringKindOSC,
			Introducer: TerminalStringIntroducerESC,
			Byte:       ']',
		},
	})
	data = data[2:]
	assertVTWriteUntil(t, term, data, want|TerminalBoundaryGround, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed: len(data),
			Reached:  TerminalBoundaryOSCEnd | TerminalBoundaryGround,
			Kind:     TerminalStringKindOSC,
			End:      TerminalStringEndBEL,
			Byte:     0x07,
		},
	})

	title, err := term.Title()
	if err != nil {
		t.Fatal(err)
	}
	if title != "title" {
		t.Fatalf("expected OSC to be processed, title is %q", title)
	}
}

func TestTerminalVTWriteUntilDCS(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	want := TerminalBoundaryDCSStart | TerminalBoundaryDCSEnd
	data := []byte("\x1bP1;2|data\x18")
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed:   2,
			Reached:    TerminalBoundaryDCSStart,
			Kind:       TerminalStringKindDCS,
			Introducer: TerminalStringIntroducerESC,
			Byte:       'P',
		},
	})
	data = data[2:]
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed: len(data),
			Reached:  TerminalBoundaryDCSEnd,
			Kind:     TerminalStringKindDCS,
			End:      TerminalStringEndCanceled,
			Byte:     0x18,
		},
	})
}

func TestTerminalVTWriteUntilAbortedStartsNewString(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// 0x9D inside an APC ends it and starts an OSC on the same byte.
	want := TerminalBoundaryAPCEnd | TerminalBoundaryOSCStart
	data := []byte("\x1b_a\x9d2;t\x07")
	assertVTWriteUntil(t, term, data, want, vtWriteUntilExpect{
		reached: true,
		stop: TerminalWriteStop{
			Consumed:   4,
			Reached:    TerminalBoundaryAPCEnd | TerminalBoundaryOSCStart,
			Kind:       TerminalStringKindAPC,
			Introducer: TerminalStringIntroducerC1,
			End:        TerminalStringEndAborted,
			Byte:       0x9d,
		},
	})
}

func TestTerminalVTWriteUntilInvalid(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	for _, want := range []TerminalBoundary{0, 1 << 31} {
		stop, err := term.VTWriteUntil([]byte("abc"), want)
		assertResultError(t, err, ResultInvalidValue)
		if stop != (TerminalWriteStop{}) {
			t.Fatalf("expected zero stop on error, got %+v", stop)
		}
	}

	// Nothing was consumed by the failed calls.
	x, err := term.CursorX()
	if err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("expected invalid calls to consume nothing, cursor at %d", x)
	}
}

// TestTerminalVTWriteUntilNoAllocs guards the hot path: neither a stop
// nor a full consume may allocate.
func TestTerminalVTWriteUntilNoAllocs(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	data := []byte("text \x1b]2;t\x07 more text")
	want := TerminalBoundaryOSCStart | TerminalBoundaryOSCEnd
	allocs := testing.AllocsPerRun(100, func() {
		rest := data
		for len(rest) > 0 {
			stop, err := term.VTWriteUntil(rest, want)
			if err != nil {
				t.Fatal(err)
			}
			rest = rest[stop.Consumed:]
		}
	})
	if allocs != 0 {
		t.Fatalf("expected zero allocations, got %v", allocs)
	}
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

func TestTerminalScrollViewportPrompt(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 5), WithMaxScrollbackLines(1000))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Write three prompts, each followed by enough output to push the
	// earlier prompts into scrollback. OSC 133;A marks each prompt start.
	for i := 0; i < 3; i++ {
		term.VTWrite([]byte("\x1b]133;A\x07$ cmd\r\n"))
		for j := 0; j < 10; j++ {
			term.VTWrite([]byte("output\r\n"))
		}
	}

	offset := func() uint64 {
		t.Helper()
		sb, err := term.Scrollbar()
		if err != nil {
			t.Fatal(err)
		}
		return sb.Offset
	}

	bottom := offset()
	if bottom == 0 {
		t.Fatal("expected content in scrollback")
	}

	// Zero is a no-op.
	term.ScrollViewportPrompt(0)
	if got := offset(); got != bottom {
		t.Fatalf("delta 0 moved viewport: %d, want %d", got, bottom)
	}

	// Each prompt is 11 rows apart, starting at row 0. Moving back one
	// prompt from the active area lands on the last prompt (row 22).
	term.ScrollViewportPrompt(-1)
	if got := offset(); got != 22 {
		t.Fatalf("delta -1: offset %d, want 22", got)
	}

	// Moving back past the first prompt clamps to the first prompt.
	term.ScrollViewportPrompt(-10)
	if got := offset(); got != 0 {
		t.Fatalf("delta -10: offset %d, want 0", got)
	}

	// Moving forward one prompt lands on the second prompt.
	term.ScrollViewportPrompt(1)
	if got := offset(); got != 11 {
		t.Fatalf("delta 1: offset %d, want 11", got)
	}
}
