// Example: input demonstrates the key and mouse encoders from libghostty.
//
// A terminal emulator host receives key presses and mouse events from
// the windowing system and must translate them into the byte sequences
// that a program running in the terminal expects to read from the pty.
// The KeyEncoder and MouseEncoder perform that translation. Both are
// configured from the terminal's current modes with SetOptFromTerminal,
// so the encoding automatically follows whatever the running program
// has requested (application cursor keys, mouse tracking, the Kitty
// keyboard protocol, and so on).
//
// This program never talks to a real pty. It feeds mode-changing escape
// sequences into a Terminal, then shows how the same key or mouse
// event encodes differently before and after.
package main

import (
	"fmt"
	"log"

	ghostty "go.mitchellh.com/libghostty"
)

func main() {
	term, err := ghostty.NewTerminal(ghostty.WithSize(80, 24))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	keys(term)
	mouse(term)
}

// keys encodes a few key presses, first with the encoder's defaults and
// then after the terminal has switched modes.
func keys(term *ghostty.Terminal) {
	enc, err := ghostty.NewKeyEncoder()
	if err != nil {
		log.Fatal(err)
	}
	defer enc.Close()

	// A KeyEvent is a reusable, mutable description of one key press.
	// Each field is set independently; anything not set keeps its
	// previous value, so we reset the relevant fields before each use.
	ev, err := ghostty.NewKeyEvent()
	if err != nil {
		log.Fatal(err)
	}
	defer ev.Close()

	fmt.Println("== keys (legacy encoding)")

	// A plain letter. Key identifies the physical key; UTF8 carries the
	// text the platform produced for it, which is what actually gets
	// sent for printable keys. UnshiftedCodepoint is the character the
	// key produces with no modifiers held. Legacy encoding ignores it,
	// but the Kitty protocol below reports it, so a host should always
	// supply it when it is known.
	ev.SetAction(ghostty.KeyActionPress)
	ev.SetKey(ghostty.KeyA)
	ev.SetMods(0)
	ev.SetUTF8("a")
	ev.SetUnshiftedCodepoint('a')
	encodeKey(enc, ev, "a")

	// Ctrl+C. No UTF-8 text is produced for control combinations, so
	// the encoder derives the control byte from the key and modifiers.
	ev.SetKey(ghostty.KeyC)
	ev.SetMods(ghostty.ModCtrl)
	ev.SetUTF8("")
	ev.SetUnshiftedCodepoint('c')
	encodeKey(enc, ev, "ctrl+c")

	// Up arrow in normal cursor key mode encodes as CSI A. Arrow keys
	// have no codepoint of their own.
	ev.SetKey(ghostty.KeyArrowUp)
	ev.SetMods(0)
	ev.SetUnshiftedCodepoint(0)
	encodeKey(enc, ev, "up")

	// Now have the "program" enable application cursor keys (DECCKM,
	// DEC private mode 1) by writing the escape sequence to the
	// terminal, then sync the encoder with the terminal's modes. The
	// same up arrow now encodes as SS3 A.
	fmt.Println("== keys (after ESC [ ? 1 h, application cursor keys)")
	term.VTWrite([]byte("\x1b[?1h"))
	enc.SetOptFromTerminal(term)
	encodeKey(enc, ev, "up")

	// Enable the Kitty keyboard protocol with the "disambiguate" flag
	// (CSI > 1 u). Ctrl+C is no longer a bare control byte; it becomes
	// a CSI u sequence carrying the codepoint and modifier so the
	// program can tell it apart from other keys that share the byte.
	fmt.Println("== keys (after ESC [ > 1 u, Kitty disambiguate)")
	term.VTWrite([]byte("\x1b[>1u"))
	enc.SetOptFromTerminal(term)
	ev.SetKey(ghostty.KeyC)
	ev.SetMods(ghostty.ModCtrl)
	ev.SetUnshiftedCodepoint('c')
	encodeKey(enc, ev, "ctrl+c")

	// Key releases produce nothing under legacy encoding, and only
	// produce output under Kitty when the report-events flag is set.
	// Encode returns a nil slice and nil error for "nothing to send".
	ev.SetAction(ghostty.KeyActionRelease)
	encodeKey(enc, ev, "ctrl+c release")
}

// mouse encodes a click, first while the terminal has mouse reporting
// disabled and then after it has been enabled in SGR format.
func mouse(term *ghostty.Terminal) {
	enc, err := ghostty.NewMouseEncoder()
	if err != nil {
		log.Fatal(err)
	}
	defer enc.Close()

	// The encoder converts pixel positions to cells, so it needs to know
	// the screen and cell geometry. This is host state, not terminal
	// state, so it is not covered by SetOptFromTerminal.
	enc.SetOptSize(ghostty.MouseEncoderSize{
		ScreenWidth:  640,
		ScreenHeight: 384,
		CellWidth:    8,
		CellHeight:   16,
	})

	ev, err := ghostty.NewMouseEvent()
	if err != nil {
		log.Fatal(err)
	}
	defer ev.Close()

	// Left button press over the third column of the second row. Mouse
	// positions are in pixels relative to the top-left of the grid.
	ev.SetAction(ghostty.MouseActionPress)
	ev.SetButton(ghostty.MouseButtonLeft)
	ev.SetPosition(ghostty.MousePosition{X: 20, Y: 20})

	// With mouse tracking off, which is the default, the click produces
	// no output at all and the host would handle it locally instead
	// (for example by starting a selection).
	fmt.Println("== mouse (tracking off)")
	enc.SetOptFromTerminal(term)
	encodeMouse(enc, ev, "left press")

	// Enable normal mouse tracking (mode 1000) with SGR encoding (mode
	// 1006), the combination most modern programs request, and sync the
	// encoder. The click now encodes as CSI < 0 ; 3 ; 2 M.
	fmt.Println("== mouse (after ESC [ ? 1000 h and ESC [ ? 1006 h)")
	term.VTWrite([]byte("\x1b[?1000h\x1b[?1006h"))
	enc.SetOptFromTerminal(term)
	encodeMouse(enc, ev, "left press")

	// The matching release. SGR format distinguishes releases with a
	// trailing 'm' instead of 'M'.
	ev.SetAction(ghostty.MouseActionRelease)
	encodeMouse(enc, ev, "left release")

	// Motion without a button is only reported in "any event" tracking
	// (mode 1003). Under normal tracking it is dropped.
	ev.SetAction(ghostty.MouseActionMotion)
	ev.ClearButton()
	ev.SetPosition(ghostty.MousePosition{X: 100, Y: 20})
	encodeMouse(enc, ev, "motion")
}

// encodeKey prints the escape sequence for a key event, or a note when
// the event produces no output.
func encodeKey(enc *ghostty.KeyEncoder, ev *ghostty.KeyEvent, label string) {
	out, err := enc.Encode(ev)
	if err != nil {
		log.Fatalf("%s: %v", label, err)
	}
	report(label, out)
}

// encodeMouse prints the escape sequence for a mouse event, or a note
// when the event produces no output.
func encodeMouse(enc *ghostty.MouseEncoder, ev *ghostty.MouseEvent, label string) {
	out, err := enc.Encode(ev)
	if err != nil {
		log.Fatalf("%s: %v", label, err)
	}
	report(label, out)
}

// report prints an encoded sequence using Go's quoted form so escape
// bytes are visible (e.g. "\x1b[A").
func report(label string, out []byte) {
	if out == nil {
		fmt.Printf("%-16s -> (no output)\n", label)
		return
	}
	fmt.Printf("%-16s -> %q\n", label, out)
}
