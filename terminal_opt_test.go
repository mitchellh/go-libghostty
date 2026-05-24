package libghostty

import "testing"

func TestTerminalSetTitle(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetTitle("my terminal"); err != nil {
		t.Fatal(err)
	}

	// Clear the title.
	if err := term.SetTitle(""); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetPwd(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetPwd("/tmp"); err != nil {
		t.Fatal(err)
	}

	if err := term.SetPwd(""); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetColors(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	white := &ColorRGB{R: 255, G: 255, B: 255}
	black := &ColorRGB{R: 0, G: 0, B: 0}

	if err := term.SetColorForeground(white); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorBackground(black); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorCursor(white); err != nil {
		t.Fatal(err)
	}

	// Clear colors.
	if err := term.SetColorForeground(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorBackground(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorCursor(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetAPCMaxBytes(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	limit := uint(1024)
	if err := term.SetAPCMaxBytes(&limit); err != nil {
		t.Fatal(err)
	}
	if err := term.SetAPCMaxBytes(nil); err != nil {
		t.Fatal(err)
	}

	kittyLimit := uint(512)
	if err := term.SetAPCMaxBytesKitty(&kittyLimit); err != nil {
		t.Fatal(err)
	}
	if err := term.SetAPCMaxBytesKitty(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalWithBell(t *testing.T) {
	var bellCount int
	term, err := NewTerminal(WithSize(80, 24), WithBell(func(_ *Terminal) {
		bellCount++
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// BEL character should trigger the callback.
	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected 1 bell, got %d", bellCount)
	}

	// Multiple BELs.
	term.VTWrite([]byte("\x07\x07"))
	if bellCount != 3 {
		t.Fatalf("expected 3 bells, got %d", bellCount)
	}
}

func TestTerminalSetEffectBell(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	var bellCount int
	term.SetEffectBell(func(_ *Terminal) {
		bellCount++
	})

	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected 1 bell, got %d", bellCount)
	}

	// Clear the callback; bell should no longer fire.
	term.SetEffectBell(nil)
	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected still 1 bell after clearing, got %d", bellCount)
	}
}

func TestTerminalWithWritePty(t *testing.T) {
	var received []byte
	term, err := NewTerminal(WithSize(80, 24), WithWritePty(func(_ *Terminal, data []byte) {
		received = append(received, data...)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// DA1 query should produce a response via write_pty.
	term.VTWrite([]byte("\x1b[c"))
	if len(received) == 0 {
		t.Fatal("expected write_pty data from DA1 query")
	}
}

func TestTerminalWithTitleChanged(t *testing.T) {
	var titleChanged int
	term, err := NewTerminal(WithSize(80, 24), WithTitleChanged(func(_ *Terminal) {
		titleChanged++
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 2 sets the title.
	term.VTWrite([]byte("\x1b]2;hello\x07"))
	if titleChanged != 1 {
		t.Fatalf("expected 1 title change, got %d", titleChanged)
	}
}

func TestTerminalWithEnquiry(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
		WithEnquiry(func(_ *Terminal) []byte {
			return []byte("hello")
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// ENQ character should trigger enquiry and write response via pty.
	term.VTWrite([]byte("\x05"))
	if string(received) != "hello" {
		t.Fatalf("expected enquiry response %q, got %q", "hello", string(received))
	}
}

func TestTerminalWithXtversion(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
		WithXtversion(func(_ *Terminal) string {
			return "myterm 1.0"
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// XTVERSION query: CSI > q
	term.VTWrite([]byte("\x1b[>q"))
	// Response should contain our version string in a DCS sequence.
	if len(received) == 0 {
		t.Fatal("expected xtversion response")
	}
	resp := string(received)
	if !contains(resp, "myterm 1.0") {
		t.Fatalf("expected response to contain %q, got %q", "myterm 1.0", resp)
	}
}

// contains reports whether s contains substr.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTerminalSetColorPalette(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Set a custom palette.
	var palette Palette
	for i := range palette {
		palette[i] = ColorRGB{R: uint8(i), G: uint8(i), B: uint8(i)}
	}
	if err := term.SetColorPalette(&palette); err != nil {
		t.Fatal(err)
	}

	// Reset to default.
	if err := term.SetColorPalette(nil); err != nil {
		t.Fatal(err)
	}
}
