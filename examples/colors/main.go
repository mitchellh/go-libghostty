// Command colors demonstrates the libghostty color APIs: setting and
// querying foreground, background, cursor, and palette colors, as well
// as the distinction between "effective" (OSC-overridden) and "default"
// values.
package main

import (
	"fmt"
	"log"

	ghostty "go.mitchellh.com/libghostty"
)

func main() {
	// Step 1: Create an 80×24 terminal with no scrollback.
	t, err := ghostty.NewTerminal(
		ghostty.WithSize(80, 24),
		ghostty.WithMaxScrollback(0),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()

	// Step 2: Print colors before any configuration — everything is unset.
	fmt.Println("=== Before setting colors ===")
	printColors(t)

	// Step 3: Apply a Catppuccin-inspired dark theme via the config API.
	if err := t.SetColorForeground(&ghostty.ColorRGB{R: 205, G: 214, B: 244}); err != nil {
		log.Fatal(err)
	}
	if err := t.SetColorBackground(&ghostty.ColorRGB{R: 30, G: 30, B: 46}); err != nil {
		log.Fatal(err)
	}
	if err := t.SetColorCursor(&ghostty.ColorRGB{R: 245, G: 224, B: 220}); err != nil {
		log.Fatal(err)
	}

	// Override the first 8 palette entries with Catppuccin colors.
	palette, err := t.ColorPalette()
	if err != nil {
		log.Fatal(err)
	}
	palette[ghostty.ColorNamedBlack] = ghostty.ColorRGB{R: 69, G: 71, B: 90}
	palette[ghostty.ColorNamedRed] = ghostty.ColorRGB{R: 243, G: 139, B: 168}
	palette[ghostty.ColorNamedGreen] = ghostty.ColorRGB{R: 166, G: 227, B: 161}
	palette[ghostty.ColorNamedYellow] = ghostty.ColorRGB{R: 249, G: 226, B: 175}
	palette[ghostty.ColorNamedBlue] = ghostty.ColorRGB{R: 137, G: 180, B: 250}
	palette[ghostty.ColorNamedMagenta] = ghostty.ColorRGB{R: 245, G: 194, B: 231}
	palette[ghostty.ColorNamedCyan] = ghostty.ColorRGB{R: 148, G: 226, B: 213}
	palette[ghostty.ColorNamedWhite] = ghostty.ColorRGB{R: 186, G: 194, B: 222}
	if err := t.SetColorPalette(palette); err != nil {
		log.Fatal(err)
	}

	// Step 4: Print colors after applying the theme.
	fmt.Println("\n=== After setting Catppuccin theme ===")
	printColors(t)

	// Step 5: Use OSC 10 to override the foreground color to red via VT
	// input. This changes the "effective" color but leaves the "default"
	// unchanged.
	t.VTWrite([]byte("\x1b]10;rgb:ff/00/00\x1b\\"))

	fmt.Println("\n=== After OSC 10 override (fg → red) ===")
	printColors(t)

	// Step 7: Clear the default foreground by passing nil.
	if err := t.SetColorForeground(nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== After clearing default foreground ===")
	printColors(t)
}

// printColors prints the effective and default values for foreground,
// background, cursor, and palette[0].
func printColors(t *ghostty.Terminal) {
	type colorPair struct {
		label    string
		eff, def func() (*ghostty.ColorRGB, error)
	}

	pairs := []colorPair{
		{"Foreground", t.ColorForeground, t.ColorForegroundDefault},
		{"Background", t.ColorBackground, t.ColorBackgroundDefault},
		{"Cursor", t.ColorCursor, t.ColorCursorDefault},
	}

	for _, p := range pairs {
		eff, err := p.eff()
		if err != nil {
			log.Fatal(err)
		}
		def, err := p.def()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %-12s effective=%-12s default=%s\n",
			p.label, formatColor(eff), formatColor(def))
	}

	// Print palette entry 0 (black).
	palette, err := t.ColorPalette()
	if err != nil {
		log.Fatal(err)
	}
	paletteDefault, err := t.ColorPaletteDefault()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %-12s effective=%-12s default=%s\n",
		"Palette[0]",
		formatColor(&palette[0]),
		formatColor(&paletteDefault[0]))
}

// formatColor formats a *ColorRGB as "#RRGGBB" or "(not set)" if nil.
func formatColor(c *ghostty.ColorRGB) string {
	if c == nil {
		return "(not set)"
	}
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}
