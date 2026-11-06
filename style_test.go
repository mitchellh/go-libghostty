package libghostty

import "testing"

func TestDefaultStyle(t *testing.T) {
	style := DefaultStyle()
	if !style.IsDefault() {
		t.Fatal("expected default style")
	}
	if style.Underline() != UnderlineNone {
		t.Fatalf("expected no underline, got %d", style.Underline())
	}
}

// styleAtCursor writes sgr followed by a character and returns the style of
// that character, so each case exercises the packed C flags end to end.
func styleAtCursor(t *testing.T, sgr string) *Style {
	t.Helper()

	term, err := NewTerminal(WithSize(10, 2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(term.Close)

	term.VTWrite([]byte(sgr + "x"))
	ref, err := term.GridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	style, err := ref.Style()
	if err != nil {
		t.Fatal(err)
	}
	return style
}

// TestStyleFlags checks that each attribute reads its own bit. Every case sets
// exactly one attribute and requires the others to stay clear, so a getter
// wired to the wrong flag fails here.
func TestStyleFlags(t *testing.T) {
	getters := map[string]func(*Style) bool{
		"bold":          (*Style).Bold,
		"italic":        (*Style).Italic,
		"faint":         (*Style).Faint,
		"blink":         (*Style).Blink,
		"inverse":       (*Style).Inverse,
		"invisible":     (*Style).Invisible,
		"strikethrough": (*Style).Strikethrough,
		"overline":      (*Style).Overline,
	}
	sequences := map[string]string{
		"bold":          "\x1b[1m",
		"faint":         "\x1b[2m",
		"italic":        "\x1b[3m",
		"blink":         "\x1b[5m",
		"inverse":       "\x1b[7m",
		"invisible":     "\x1b[8m",
		"strikethrough": "\x1b[9m",
		"overline":      "\x1b[53m",
	}

	for name, sgr := range sequences {
		t.Run(name, func(t *testing.T) {
			style := styleAtCursor(t, sgr)
			for other, get := range getters {
				if got, want := get(style), other == name; got != want {
					t.Errorf("%s = %v, want %v", other, got, want)
				}
			}
			if style.Underline() != UnderlineNone {
				t.Errorf("unexpected underline %d", style.Underline())
			}
		})
	}
}

// TestStyleUnderline checks that every underline style survives the shift out
// of the packed flags, and that it doesn't disturb the on/off attributes that
// share the same 16 bits.
func TestStyleUnderline(t *testing.T) {
	tests := []struct {
		sgr  string
		want SGRUnderline
	}{
		{"\x1b[4:1m", UnderlineSingle},
		{"\x1b[4:2m", UnderlineDouble},
		{"\x1b[4:3m", UnderlineCurly},
		{"\x1b[4:4m", UnderlineDotted},
		{"\x1b[4:5m", UnderlineDashed},
	}

	for _, tt := range tests {
		style := styleAtCursor(t, tt.sgr)
		if got := style.Underline(); got != tt.want {
			t.Errorf("%q: underline = %d, want %d", tt.sgr, got, tt.want)
		}
		if style.Bold() || style.Overline() {
			t.Errorf("%q: underline leaked into the attribute flags", tt.sgr)
		}
	}

	// The highest flag and the underline bits sit next to each other.
	style := styleAtCursor(t, "\x1b[53;4:5m")
	if !style.Overline() || style.Underline() != UnderlineDashed {
		t.Fatalf("overline = %v, underline = %d", style.Overline(), style.Underline())
	}
}

// TestStyleColors checks the three-byte color value of the frozen style layout.
func TestStyleColors(t *testing.T) {
	style := styleAtCursor(t, "\x1b[38;5;42;48;2;10;20;30;58;2;1;2;3m")

	if fg := style.FgColor(); fg.Tag != StyleColorPalette || fg.Palette != 42 {
		t.Fatalf("unexpected foreground: %+v", fg)
	}
	want := ColorRGB{R: 10, G: 20, B: 30}
	if bg := style.BgColor(); bg.Tag != StyleColorRGB || bg.RGB != want {
		t.Fatalf("unexpected background: %+v", bg)
	}
	want = ColorRGB{R: 1, G: 2, B: 3}
	if ul := style.UnderlineColor(); ul.Tag != StyleColorRGB || ul.RGB != want {
		t.Fatalf("unexpected underline color: %+v", ul)
	}
	if style.IsDefault() {
		t.Fatal("expected a non-default style")
	}
}
