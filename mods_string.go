package libghostty

import (
	"encoding"
	"fmt"
	"strings"
)

// Compile-time assertions that Mods implements the standard text
// marshaling interfaces.
var (
	_ encoding.TextMarshaler   = Mods(0)
	_ encoding.TextUnmarshaler = (*Mods)(nil)
)

// Human-friendly string conversion for Mods bitmasks. Each set bit
// is rendered using a snake_case name; multiple modifiers are joined
// with "+". Parsing accepts "+" or "," as separators and supports
// common aliases (cmd/command for super, opt/option for alt, control
// for ctrl), matching the upstream Zig source's modifier alias list.

// modBitNames is the canonical ordered list of single-bit Mods values
// and their snake_case names. The ordering controls the output of
// String, which always renders bits in this fixed order so that
// equivalent bitmasks produce identical strings.
var modBitNames = []struct {
	bit  Mods
	name string
}{
	{ModShift, "shift"},
	{ModCtrl, "ctrl"},
	{ModAlt, "alt"},
	{ModSuper, "super"},
	{ModCapsLock, "caps_lock"},
	{ModNumLock, "num_lock"},
	{ModShiftSide, "shift_side"},
	{ModCtrlSide, "ctrl_side"},
	{ModAltSide, "alt_side"},
	{ModSuperSide, "super_side"},
}

// modAliases lists alternate names accepted by FromString. The
// canonical name for each modifier is in modBitNames.
var modAliases = map[string]Mods{
	"cmd":     ModSuper,
	"command": ModSuper,
	"opt":     ModAlt,
	"option":  ModAlt,
	"control": ModCtrl,
}

// String returns a human-friendly representation of the modifier
// bitmask: each set bit is rendered as its snake_case name and bits
// are joined with "+" in a stable canonical order. Returns "" if no
// bits are set.
//
// Examples:
//
//	Mods(0).String()                   == ""
//	(ModShift | ModCtrl).String()      == "shift+ctrl"
//	(ModShift | ModShiftSide).String() == "shift+shift_side"
func (m Mods) String() string {
	if m == 0 {
		return ""
	}
	var parts []string
	for _, e := range modBitNames {
		if m&e.bit != 0 {
			parts = append(parts, e.name)
		}
	}
	return strings.Join(parts, "+")
}

// MarshalText implements encoding.TextMarshaler. The output is the
// same as String() so the type integrates with encoding/json and
// other text-based encoders. An empty bitmask marshals to an empty
// byte slice (i.e. JSON `""`).
func (m Mods) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It parses a
// "+" or "," separated list of modifier names and stores the
// resulting bitmask in the receiver. Whitespace around tokens and
// empty tokens are ignored. Recognized names are the canonical
// snake_case names returned by String plus the upstream aliases
// (cmd/command, opt/option, control). An empty input is valid and
// produces a zero bitmask. Returns an error if any token is not
// recognized.
//
// The receiver is overwritten, not OR'd into.
func (m *Mods) UnmarshalText(text []byte) error {
	var out Mods
	if len(text) > 0 {
		// Normalize "," separators to "+" so we can split once.
		s := strings.ReplaceAll(string(text), ",", "+")
		for _, raw := range strings.Split(s, "+") {
			tok := strings.TrimSpace(raw)
			if tok == "" {
				continue
			}
			if alias, ok := modAliases[tok]; ok {
				out |= alias
				continue
			}
			matched := false
			for _, e := range modBitNames {
				if e.name == tok {
					out |= e.bit
					matched = true
					break
				}
			}
			if !matched {
				return fmt.Errorf("libghostty: unknown modifier %q", tok)
			}
		}
	}
	*m = out
	return nil
}

// ParseMods returns the Mods value parsed from the given "+" or ","
// separated list of modifier names. See Mods.UnmarshalText for the
// accepted syntax and aliases.
func ParseMods(s string) (Mods, error) {
	var m Mods
	if err := m.UnmarshalText([]byte(s)); err != nil {
		return 0, err
	}
	return m, nil
}
