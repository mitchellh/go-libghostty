package libghostty

import (
	"encoding/json"
	"testing"
)

// Tests for String, MarshalText, UnmarshalText, and Parse* helpers
// across enum types. The compile-time interface assertions live
// alongside each type's implementation.

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

func TestKeyUnmarshalText(t *testing.T) {
	var k Key
	if err := k.UnmarshalText([]byte("arrow_left")); err != nil {
		t.Fatal(err)
	}
	if k != KeyArrowLeft {
		t.Fatalf("expected KeyArrowLeft, got %d", k)
	}
	if err := k.UnmarshalText([]byte("not_a_real_key")); err == nil {
		t.Fatal("expected error for unknown key name")
	}
}

func TestKeyRoundtrip(t *testing.T) {
	for _, e := range keyNames {
		var k Key
		if err := k.UnmarshalText([]byte(e.name)); err != nil {
			t.Fatalf("UnmarshalText(%q) failed: %v", e.name, err)
		}
		if k != e.key {
			t.Fatalf("UnmarshalText(%q) = %d, want %d", e.name, k, e.key)
		}
		got, err := k.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText failed: %v", err)
		}
		if string(got) != e.name {
			t.Fatalf("MarshalText roundtrip mismatch: %q -> %q", e.name, string(got))
		}
	}
}

func TestKeyJSON(t *testing.T) {
	// Round-trip through encoding/json to confirm TextMarshaler
	// integration.
	type wrap struct {
		K Key `json:"k"`
	}
	in := wrap{K: KeyArrowUp}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"k":"arrow_up"}` {
		t.Fatalf("json.Marshal = %s, want %s", data, `{"k":"arrow_up"}`)
	}
	var out wrap
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.K != KeyArrowUp {
		t.Fatalf("json.Unmarshal = %d, want %d", out.K, KeyArrowUp)
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

func TestModsMarshalEmpty(t *testing.T) {
	got, err := Mods(0).MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "" {
		t.Fatalf("Mods(0).MarshalText() = %q, want empty", string(got))
	}
}

func TestModsUnmarshalText(t *testing.T) {
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
		if err := m.UnmarshalText([]byte(c.in)); err != nil {
			t.Fatalf("UnmarshalText(%q) failed: %v", c.in, err)
		}
		if m != c.want {
			t.Errorf("UnmarshalText(%q) = %d, want %d", c.in, m, c.want)
		}
	}

	var m Mods
	if err := m.UnmarshalText([]byte("nope")); err == nil {
		t.Fatal("expected error for unknown modifier")
	}
}

func TestModsRoundtrip(t *testing.T) {
	all := ModShift | ModCtrl | ModAlt | ModSuper | ModCapsLock |
		ModNumLock | ModShiftSide | ModCtrlSide | ModAltSide | ModSuperSide
	text, err := all.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var m Mods
	if err := m.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}
	if m != all {
		t.Fatalf("roundtrip failed: %d -> %q -> %d", all, string(text), m)
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

func TestMouseButtonUnmarshalText(t *testing.T) {
	for _, e := range mouseButtonNames {
		var b MouseButton
		if err := b.UnmarshalText([]byte(e.name)); err != nil {
			t.Fatalf("UnmarshalText(%q) failed: %v", e.name, err)
		}
		if b != e.button {
			t.Fatalf("UnmarshalText(%q) = %d, want %d", e.name, b, e.button)
		}
	}

	var b MouseButton
	if err := b.UnmarshalText([]byte("nope")); err == nil {
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

func TestFocusEventUnmarshalText(t *testing.T) {
	var f FocusEvent
	if err := f.UnmarshalText([]byte("gained")); err != nil || f != FocusGained {
		t.Fatalf("UnmarshalText(gained) = %d, %v", f, err)
	}
	if err := f.UnmarshalText([]byte("lost")); err != nil || f != FocusLost {
		t.Fatalf("UnmarshalText(lost) = %d, %v", f, err)
	}
	if err := f.UnmarshalText([]byte("nope")); err == nil {
		t.Fatal("expected error for unknown focus event")
	}
}

func TestParseHelpers(t *testing.T) {
	if k, err := ParseKey("arrow_up"); err != nil || k != KeyArrowUp {
		t.Fatalf("ParseKey(arrow_up) = %d, %v", k, err)
	}
	if _, err := ParseKey("nope"); err == nil {
		t.Fatal("expected error from ParseKey")
	}

	if m, err := ParseMods("shift+ctrl"); err != nil || m != ModShift|ModCtrl {
		t.Fatalf("ParseMods = %d, %v", m, err)
	}
	if _, err := ParseMods("nope"); err == nil {
		t.Fatal("expected error from ParseMods")
	}

	if b, err := ParseMouseButton("middle"); err != nil || b != MouseButtonMiddle {
		t.Fatalf("ParseMouseButton = %d, %v", b, err)
	}
	if _, err := ParseMouseButton("nope"); err == nil {
		t.Fatal("expected error from ParseMouseButton")
	}

	if f, err := ParseFocusEvent("gained"); err != nil || f != FocusGained {
		t.Fatalf("ParseFocusEvent = %d, %v", f, err)
	}
	if _, err := ParseFocusEvent("nope"); err == nil {
		t.Fatal("expected error from ParseFocusEvent")
	}
}
