package libghostty

import (
	"errors"
	"testing"
)

func TestMouseEventNewClose(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()
}

func TestMouseEventAction(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetAction(MouseActionPress)
	if got := ev.Action(); got != MouseActionPress {
		t.Fatalf("expected MouseActionPress, got %d", got)
	}

	ev.SetAction(MouseActionRelease)
	if got := ev.Action(); got != MouseActionRelease {
		t.Fatalf("expected MouseActionRelease, got %d", got)
	}

	ev.SetAction(MouseActionMotion)
	if got := ev.Action(); got != MouseActionMotion {
		t.Fatalf("expected MouseActionMotion, got %d", got)
	}
}

func TestMouseEventButton(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetButton(MouseButtonLeft)
	btn, ok := ev.Button()
	if !ok {
		t.Fatal("expected button to be set")
	}
	if btn != MouseButtonLeft {
		t.Fatalf("expected MouseButtonLeft, got %d", btn)
	}

	ev.ClearButton()
	_, ok = ev.Button()
	if ok {
		t.Fatal("expected button to be cleared")
	}
}

func TestMouseEventMods(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetMods(ModCtrl | ModShift)
	if got := ev.Mods(); got != ModCtrl|ModShift {
		t.Fatalf("expected Ctrl|Shift, got %d", got)
	}
}

func TestMouseEventPosition(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	pos := MousePosition{X: 10.5, Y: 20.75}
	ev.SetPosition(pos)
	got := ev.Position()
	if got.X != pos.X || got.Y != pos.Y {
		t.Fatalf("expected %+v, got %+v", pos, got)
	}
}

func TestMouseEncoderNewClose(t *testing.T) {
	enc, err := NewMouseEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
}

func TestMouseEncoderEncodeNormalSGR(t *testing.T) {
	enc, err := NewMouseEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	// Configure for normal tracking with SGR format.
	enc.SetTrackingMode(MouseTrackingNormal)
	enc.SetFormat(MouseFormatSGR)
	enc.SetSize(MouseEncoderSize{
		ScreenWidth:  640,
		ScreenHeight: 480,
		CellWidth:    8,
		CellHeight:   16,
	})

	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	// Left button press at cell (1,1) — position (0,0) in pixels.
	ev.SetAction(MouseActionPress)
	ev.SetButton(MouseButtonLeft)
	ev.SetPosition(MousePosition{X: 0, Y: 0})

	out, err := enc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty output for mouse press")
	}
	// SGR press format: \x1b[<0;1;1M
	expected := "\x1b[<0;1;1M"
	if string(out) != expected {
		t.Fatalf("expected %q, got %q", expected, string(out))
	}
}

func TestMouseEncoderReset(t *testing.T) {
	enc, err := NewMouseEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	// Should not panic.
	enc.Reset()
}

func TestMouseEncoderSetFromTerminal(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	enc, err := NewMouseEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	// Should not panic.
	enc.SetFromTerminal(term)
}

// TestMouseEventSetterErrors checks that setters report what libghostty rejects
// and leave the event unchanged.
func TestMouseEventSetterErrors(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	if err := ev.SetAction(MouseActionRelease); err != nil {
		t.Fatal(err)
	}
	if err := ev.SetAction(MouseAction(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if got := ev.Action(); got != MouseActionRelease {
		t.Fatalf("expected rejected action to be ignored, got %d", got)
	}

	if err := ev.SetButton(MouseButton(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
}

// TestMouseEventButtonNoValue checks that having no button is reported as
// ok=false rather than an error. C reports it with GHOSTTY_NO_VALUE.
func TestMouseEventButtonNoValue(t *testing.T) {
	ev, err := NewMouseEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	if _, ok := ev.Button(); ok {
		t.Fatal("expected a new event to have no button")
	}
	if err := ev.SetButton(MouseButtonRight); err != nil {
		t.Fatal(err)
	}
	if got, ok := ev.Button(); !ok || got != MouseButtonRight {
		t.Fatalf("Button() = %d, %v", got, ok)
	}
	if err := ev.ClearButton(); err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.Button(); ok {
		t.Fatal("expected ClearButton to remove the button")
	}
}

// TestMouseEncoderSetterErrors checks that encoder options report invalid
// values.
func TestMouseEncoderSetterErrors(t *testing.T) {
	enc, err := NewMouseEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	if err := enc.SetTrackingMode(MouseTrackingMode(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if err := enc.SetFormat(MouseFormat(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if err := enc.SetAnyButtonPressed(true); err != nil {
		t.Fatal(err)
	}
	if err := enc.SetTrackLastCell(true); err != nil {
		t.Fatal(err)
	}
	if err := enc.Reset(); err != nil {
		t.Fatal(err)
	}
}
