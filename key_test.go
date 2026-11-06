package libghostty

import (
	"errors"
	"testing"
	"unsafe"
)

func TestKeyEventNewClose(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()
}

func TestKeyEventAction(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetAction(KeyActionPress)
	if got := ev.Action(); got != KeyActionPress {
		t.Fatalf("expected KeyActionPress, got %d", got)
	}

	ev.SetAction(KeyActionRelease)
	if got := ev.Action(); got != KeyActionRelease {
		t.Fatalf("expected KeyActionRelease, got %d", got)
	}
}

func TestKeyEventKey(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetKey(KeyA)
	if got := ev.Key(); got != KeyA {
		t.Fatalf("expected KeyA, got %d", got)
	}
}

func TestKeyEventMods(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetMods(ModCtrl | ModShift)
	if got := ev.Mods(); got != ModCtrl|ModShift {
		t.Fatalf("expected Ctrl|Shift, got %d", got)
	}
}

func TestKeyEventConsumedMods(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetConsumedMods(ModAlt)
	if got := ev.ConsumedMods(); got != ModAlt {
		t.Fatalf("expected ModAlt, got %d", got)
	}
}

func TestKeyEventComposing(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetComposing(true)
	if !ev.Composing() {
		t.Fatal("expected composing to be true")
	}

	ev.SetComposing(false)
	if ev.Composing() {
		t.Fatal("expected composing to be false")
	}
}

func TestKeyEventUTF8(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	// Default should be empty.
	if got := ev.UTF8(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	text := "a"
	if err := ev.SetUTF8(text); err != nil {
		t.Fatal(err)
	}
	if got := ev.UTF8(); got != text {
		t.Fatalf("expected %q, got %q", text, got)
	}
}

func TestKeyEventUnshiftedCodepoint(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetUnshiftedCodepoint('a')
	if got := ev.UnshiftedCodepoint(); got != 'a' {
		t.Fatalf("expected 'a', got %c", got)
	}
}

func TestKeyEncoderNewClose(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
}

func TestKeyEncoderEncodeSimpleKey(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	// Encode 'a' key press with UTF-8 text.
	ev.SetAction(KeyActionPress)
	ev.SetKey(KeyA)
	ev.SetUTF8("a")

	out, err := enc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty output for 'a' key press")
	}
	if string(out) != "a" {
		t.Fatalf("expected 'a', got %q", string(out))
	}
}

func TestKeyEncoderEncodeArrowKey(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	// Arrow up should produce an escape sequence.
	ev.SetAction(KeyActionPress)
	ev.SetKey(KeyArrowUp)

	out, err := enc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty output for arrow up")
	}
	// Default mode: CSI A (\x1b[A)
	if string(out) != "\x1b[A" {
		t.Fatalf("expected \\x1b[A, got %q", string(out))
	}
}

func TestKeyEncoderSetFromTerminal(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	// Should not panic.
	enc.SetFromTerminal(term)

	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetAction(KeyActionPress)
	ev.SetKey(KeyA)
	ev.SetUTF8("a")

	out, err := enc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "a" {
		t.Fatalf("expected 'a', got %q", string(out))
	}
}

func TestKeyEncoderCursorKeyApplicationMode(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	// Enable cursor key application mode.
	enc.SetBool(KeyEncoderOptCursorKeyApplication, true)

	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	ev.SetAction(KeyActionPress)
	ev.SetKey(KeyArrowUp)

	out, err := enc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	// Application mode: SS3 A (\x1bOA)
	if string(out) != "\x1bOA" {
		t.Fatalf("expected \\x1bOA in application mode, got %q", string(out))
	}
}

// TestKeyEventSetterErrors checks that setters report what libghostty rejects
// and leave the event unchanged.
func TestKeyEventSetterErrors(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	if err := ev.SetAction(KeyActionRepeat); err != nil {
		t.Fatal(err)
	}
	if err := ev.SetAction(KeyAction(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if got := ev.Action(); got != KeyActionRepeat {
		t.Fatalf("expected rejected action to be ignored, got %d", got)
	}

	if err := ev.SetKey(Key(100000)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
}

// TestKeyEventUTF8Clear checks that an empty string clears earlier text. It
// reaches C as a zero-length GhosttyString.
func TestKeyEventUTF8Clear(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	text := "é"
	if err := ev.SetUTF8(text); err != nil {
		t.Fatal(err)
	}
	if got := ev.UTF8(); got != text {
		t.Fatalf("expected %q, got %q", text, got)
	}
	if err := ev.SetUTF8(""); err != nil {
		t.Fatal(err)
	}
	if got := ev.UTF8(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// TestKeyEncoderSetterErrors checks that encoder options report invalid values.
func TestKeyEncoderSetterErrors(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	if err := enc.SetBool(KeyEncoderOptCursorKeyApplication, true); err != nil {
		t.Fatal(err)
	}
	if err := enc.SetKittyFlags(KittyKeyAll); err != nil {
		t.Fatal(err)
	}
	if err := enc.SetOptionAsAlt(OptionAsAltLeft); err != nil {
		t.Fatal(err)
	}
	if err := enc.SetOptionAsAlt(OptionAsAlt(99)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
	if err := enc.SetBool(KeyEncoderOption(0), true); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value for the INVALID key, got %v", err)
	}
	if err := enc.SetBool(KeyEncoderOption(1000), true); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected unsupported for an unknown key, got %v", err)
	}
}

// TestKeyEventUTF8Copied checks that the event doesn't depend on the caller's
// string. libghostty keeps a pointer instead of copying, so a binding that
// handed it Go memory would see these bytes change, or be freed by the garbage
// collector. Mutating the backing array shows the same thing deterministically.
func TestKeyEventUTF8Copied(t *testing.T) {
	ev, err := NewKeyEvent()
	if err != nil {
		t.Fatal(err)
	}
	defer ev.Close()

	backing := []byte("hello")
	if err := ev.SetUTF8(unsafe.String(&backing[0], len(backing))); err != nil {
		t.Fatal(err)
	}
	backing[0] = 'X'
	if got := ev.UTF8(); got != "hello" {
		t.Fatalf("expected the event to keep its own copy, got %q", got)
	}

	// Replacing the text must release the old copy and use the new one.
	if err := ev.SetUTF8("second"); err != nil {
		t.Fatal(err)
	}
	if got := ev.UTF8(); got != "second" {
		t.Fatalf("expected replaced text, got %q", got)
	}
}

// TestKeyEncoderSetBoolWrongType checks that options wider than a bool are
// rejected rather than read past the end of the value.
func TestKeyEncoderSetBoolWrongType(t *testing.T) {
	enc, err := NewKeyEncoder()
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	for _, opt := range []KeyEncoderOption{KeyEncoderOptKittyFlags, KeyEncoderOptMacOSOptionAsAlt} {
		if err := enc.SetBool(opt, true); !errors.Is(err, ErrInvalidValue) {
			t.Errorf("option %d: expected invalid value, got %v", opt, err)
		}
	}
}
