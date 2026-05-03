package libghostty

import "testing"

// Tests for String / FromString conversions across enum types.

func TestKeyString(t *testing.T) {
	cases := []struct {
		key  Key
		name string
	}{
		{KeyA, "key_a"},
		{KeyZ, "key_z"},
		{KeyArrowUp, "arrow_up"},
		{KeyDigit5, "digit_5"},
		{KeyFn, "fn"},
		{KeyF13, "f13"},
		{KeyNumpadMemorySub, "numpad_memory_subtract"},
		{KeyUnidentified, "unidentified"},
	}
	for _, c := range cases {
		if got := c.key.String(); got != c.name {
			t.Errorf("Key(%d).String() = %q, want %q", c.key, got, c.name)
		}
	}
}

func TestKeyFromString(t *testing.T) {
	var k Key
	if err := k.FromString("arrow_left"); err != nil {
		t.Fatal(err)
	}
	if k != KeyArrowLeft {
		t.Fatalf("expected KeyArrowLeft, got %d", k)
	}
	if err := k.FromString("not_a_real_key"); err == nil {
		t.Fatal("expected error for unknown key name")
	}
}

func TestKeyRoundtrip(t *testing.T) {
	for _, e := range keyNames {
		var k Key
		if err := k.FromString(e.name); err != nil {
			t.Fatalf("FromString(%q) failed: %v", e.name, err)
		}
		if k != e.key {
			t.Fatalf("FromString(%q) = %d, want %d", e.name, k, e.key)
		}
		if got := k.String(); got != e.name {
			t.Fatalf("String() roundtrip mismatch: %q -> %q", e.name, got)
		}
	}
}

func TestModsString(t *testing.T) {
	cases := []struct {
		mods Mods
		want string
	}{
		{0, ""},
		{ModShift, "shift"},
		{ModCtrl | ModShift, "shift+ctrl"},
		{ModSuper | ModCapsLock, "super+caps_lock"},
		{ModShift | ModShiftSide, "shift+shift_side"},
		{ModShift | ModCtrl | ModAlt | ModSuper, "shift+ctrl+alt+super"},
	}
	for _, c := range cases {
		if got := c.mods.String(); got != c.want {
			t.Errorf("Mods(%d).String() = %q, want %q", c.mods, got, c.want)
		}
	}
}

func TestModsFromString(t *testing.T) {
	cases := []struct {
		in   string
		want Mods
	}{
		{"", 0},
		{"shift", ModShift},
		{"shift+ctrl", ModShift | ModCtrl},
		{"ctrl,shift", ModShift | ModCtrl},
		{" shift + ctrl ", ModShift | ModCtrl},
		// Aliases
		{"cmd+opt", ModSuper | ModAlt},
		{"command+option+control", ModSuper | ModAlt | ModCtrl},
		// Empty tokens are skipped
		{"shift++ctrl", ModShift | ModCtrl},
	}
	for _, c := range cases {
		var m Mods
		if err := m.FromString(c.in); err != nil {
			t.Fatalf("FromString(%q) failed: %v", c.in, err)
		}
		if m != c.want {
			t.Errorf("FromString(%q) = %d, want %d", c.in, m, c.want)
		}
	}

	var m Mods
	if err := m.FromString("nope"); err == nil {
		t.Fatal("expected error for unknown modifier")
	}
}

func TestModsRoundtrip(t *testing.T) {
	all := ModShift | ModCtrl | ModAlt | ModSuper | ModCapsLock |
		ModNumLock | ModShiftSide | ModCtrlSide | ModAltSide | ModSuperSide
	var m Mods
	if err := m.FromString(all.String()); err != nil {
		t.Fatal(err)
	}
	if m != all {
		t.Fatalf("roundtrip failed: %d -> %q -> %d", all, all.String(), m)
	}
}

func TestMouseButtonString(t *testing.T) {
	cases := []struct {
		button MouseButton
		want   string
	}{
		{MouseButtonUnknown, "unknown"},
		{MouseButtonLeft, "left"},
		{MouseButtonRight, "right"},
		{MouseButtonMiddle, "middle"},
		{MouseButtonFour, "four"},
		{MouseButtonEleven, "eleven"},
	}
	for _, c := range cases {
		if got := c.button.String(); got != c.want {
			t.Errorf("MouseButton(%d).String() = %q, want %q", c.button, got, c.want)
		}
	}
}

func TestMouseButtonFromString(t *testing.T) {
	for _, e := range mouseButtonNames {
		var b MouseButton
		if err := b.FromString(e.name); err != nil {
			t.Fatalf("FromString(%q) failed: %v", e.name, err)
		}
		if b != e.button {
			t.Fatalf("FromString(%q) = %d, want %d", e.name, b, e.button)
		}
	}

	var b MouseButton
	if err := b.FromString("nope"); err == nil {
		t.Fatal("expected error for unknown mouse button")
	}
}

func TestFocusEventString(t *testing.T) {
	if FocusGained.String() != "gained" {
		t.Errorf("FocusGained.String() = %q, want %q", FocusGained.String(), "gained")
	}
	if FocusLost.String() != "lost" {
		t.Errorf("FocusLost.String() = %q, want %q", FocusLost.String(), "lost")
	}
}

func TestFocusEventFromString(t *testing.T) {
	var f FocusEvent
	if err := f.FromString("gained"); err != nil || f != FocusGained {
		t.Fatalf("FromString(gained) = %d, %v", f, err)
	}
	if err := f.FromString("lost"); err != nil || f != FocusLost {
		t.Fatalf("FromString(lost) = %d, %v", f, err)
	}
	if err := f.FromString("nope"); err == nil {
		t.Fatal("expected error for unknown focus event")
	}
}

func TestNewFromStringConstructors(t *testing.T) {
	if k, err := NewKeyFromString("arrow_up"); err != nil || k != KeyArrowUp {
		t.Fatalf("NewKeyFromString(arrow_up) = %d, %v", k, err)
	}
	if _, err := NewKeyFromString("nope"); err == nil {
		t.Fatal("expected error from NewKeyFromString")
	}

	if m, err := NewModsFromString("shift+ctrl"); err != nil || m != ModShift|ModCtrl {
		t.Fatalf("NewModsFromString = %d, %v", m, err)
	}
	if _, err := NewModsFromString("nope"); err == nil {
		t.Fatal("expected error from NewModsFromString")
	}

	if b, err := NewMouseButtonFromString("middle"); err != nil || b != MouseButtonMiddle {
		t.Fatalf("NewMouseButtonFromString = %d, %v", b, err)
	}
	if _, err := NewMouseButtonFromString("nope"); err == nil {
		t.Fatal("expected error from NewMouseButtonFromString")
	}

	if f, err := NewFocusEventFromString("gained"); err != nil || f != FocusGained {
		t.Fatalf("NewFocusEventFromString = %d, %v", f, err)
	}
	if _, err := NewFocusEventFromString("nope"); err == nil {
		t.Fatal("expected error from NewFocusEventFromString")
	}
}
