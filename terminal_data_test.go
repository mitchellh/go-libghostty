package libghostty

import "testing"

func TestTerminalColsRows(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	cols, err := term.Cols()
	if err != nil {
		t.Fatal(err)
	}
	if cols != 80 {
		t.Fatalf("expected 80 cols, got %d", cols)
	}

	rows, err := term.Rows()
	if err != nil {
		t.Fatal(err)
	}
	if rows != 24 {
		t.Fatalf("expected 24 rows, got %d", rows)
	}

	// Resize and verify.
	if err := term.Resize(120, 40, 8, 16); err != nil {
		t.Fatal(err)
	}
	cols, _ = term.Cols()
	rows, _ = term.Rows()
	if cols != 120 || rows != 40 {
		t.Fatalf("expected 120x40 after resize, got %dx%d", cols, rows)
	}
}

func TestTerminalCursorPosition(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Cursor starts at 0,0.
	x, err := term.CursorX()
	if err != nil {
		t.Fatal(err)
	}
	y, err := term.CursorY()
	if err != nil {
		t.Fatal(err)
	}
	if x != 0 || y != 0 {
		t.Fatalf("expected cursor at 0,0, got %d,%d", x, y)
	}

	// Move cursor to col 5, row 3 (1-indexed in VT: CSI 4;6H).
	term.VTWrite([]byte("\x1b[4;6H"))
	x, _ = term.CursorX()
	y, _ = term.CursorY()
	if x != 5 || y != 3 {
		t.Fatalf("expected cursor at 5,3, got %d,%d", x, y)
	}
}

func TestTerminalTitle(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Set title via SetTitle and read it back.
	if err := term.SetTitle("hello"); err != nil {
		t.Fatal(err)
	}
	title, err := term.Title()
	if err != nil {
		t.Fatal(err)
	}
	if title != "hello" {
		t.Fatalf("expected title %q, got %q", "hello", title)
	}

	// Set title via VT escape and read back.
	term.VTWrite([]byte("\x1b]2;world\x07"))
	title, _ = term.Title()
	if title != "world" {
		t.Fatalf("expected title %q, got %q", "world", title)
	}
}

func TestTerminalPwd(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetPwd("/home/user"); err != nil {
		t.Fatal(err)
	}
	pwd, err := term.Pwd()
	if err != nil {
		t.Fatal(err)
	}
	if pwd != "/home/user" {
		t.Fatalf("expected pwd %q, got %q", "/home/user", pwd)
	}
}

func TestTerminalSelection(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// No selection is active initially.
	sel, err := term.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel != nil {
		t.Fatal("expected nil selection before setting")
	}

	start, err := term.GridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	end, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}

	if err := term.SetSelection(&Selection{Start: *start, End: *end}); err != nil {
		t.Fatal(err)
	}
	sel, err = term.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel == nil {
		t.Fatal("expected non-nil selection after setting")
	}

	if err := term.SetSelection(nil); err != nil {
		t.Fatal(err)
	}
	sel, err = term.Selection()
	if err != nil {
		t.Fatal(err)
	}
	if sel != nil {
		t.Fatal("expected nil selection after clearing")
	}
}

func TestTerminalTotalScrollbackRows(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24), WithMaxScrollbackLines(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	total, err := term.TotalRows()
	if err != nil {
		t.Fatal(err)
	}
	if total < 24 {
		t.Fatalf("expected at least 24 total rows, got %d", total)
	}

	sb, err := term.ScrollbackRows()
	if err != nil {
		t.Fatal(err)
	}
	// Fresh terminal has no scrollback content.
	if sb != 0 {
		t.Fatalf("expected 0 scrollback rows, got %d", sb)
	}
}

func TestTerminalPixelDimensions(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Before resize with pixel sizes, dimensions are 0.
	if err := term.Resize(80, 24, 10, 20); err != nil {
		t.Fatal(err)
	}
	w, err := term.WidthPx()
	if err != nil {
		t.Fatal(err)
	}
	h, err := term.HeightPx()
	if err != nil {
		t.Fatal(err)
	}
	if w != 800 {
		t.Fatalf("expected width 800px, got %d", w)
	}
	if h != 480 {
		t.Fatalf("expected height 480px, got %d", h)
	}
}

func TestTerminalMouseTracking(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	tracking, err := term.MouseTracking()
	if err != nil {
		t.Fatal(err)
	}
	if tracking {
		t.Fatal("expected no mouse tracking by default")
	}

	// Enable normal mouse tracking.
	term.VTWrite([]byte("\x1b[?1000h"))
	tracking, _ = term.MouseTracking()
	if !tracking {
		t.Fatal("expected mouse tracking after enabling")
	}
}

func TestTerminalViewportActive(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3), WithMaxScrollbackLines(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	active, err := term.ViewportActive()
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("expected fresh terminal viewport to be active")
	}

	// Fill more rows than the viewport can show so that scrolling to the
	// top moves the viewport into history rather than leaving it pinned to
	// the active area.
	term.VTWrite([]byte("alpha\r\nbravo\r\ncharlie\r\ndelta"))

	active, err = term.ViewportActive()
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("expected viewport to remain active after writing")
	}

	term.ScrollViewportTop()
	active, err = term.ViewportActive()
	if err != nil {
		t.Fatal(err)
	}
	if active {
		t.Fatal("expected viewport to be inactive after scrolling into history")
	}

	term.ScrollViewportBottom()
	active, err = term.ViewportActive()
	if err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("expected viewport to be active after scrolling to bottom")
	}
}

func TestTerminalVTProcessingError(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	failed, err := term.VTProcessingError()
	if err != nil {
		t.Fatal(err)
	}
	if failed {
		t.Fatal("expected a fresh terminal to have no VT processing error")
	}
}

func TestTerminalVTGround(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	ground, err := term.VTGround()
	if err != nil {
		t.Fatal(err)
	}
	if !ground {
		t.Fatal("expected a fresh terminal to be at VT ground")
	}

	term.VTWrite([]byte("\x1b[31"))
	ground, err = term.VTGround()
	if err != nil {
		t.Fatal(err)
	}
	if ground {
		t.Fatal("expected an unfinished CSI sequence to be above VT ground")
	}

	term.VTWrite([]byte("m"))
	ground, err = term.VTGround()
	if err != nil {
		t.Fatal(err)
	}
	if !ground {
		t.Fatal("expected a completed CSI sequence to return to VT ground")
	}
}

func TestTerminalColorRoundTrip(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// No color set initially.
	fg, err := term.ColorForeground()
	if err != nil {
		t.Fatal(err)
	}
	if fg != nil {
		t.Fatal("expected nil foreground before setting")
	}

	// Set and read back.
	red := &ColorRGB{R: 255, G: 0, B: 0}
	if err := term.SetColorForeground(red); err != nil {
		t.Fatal(err)
	}
	fg, err = term.ColorForeground()
	if err != nil {
		t.Fatal(err)
	}
	if fg == nil || *fg != *red {
		t.Fatalf("expected %+v, got %+v", red, fg)
	}

	// Default getter should also return the value.
	fgDef, err := term.ColorForegroundDefault()
	if err != nil {
		t.Fatal(err)
	}
	if fgDef == nil || *fgDef != *red {
		t.Fatalf("expected default %+v, got %+v", red, fgDef)
	}

	// Clear and verify nil again.
	if err := term.SetColorForeground(nil); err != nil {
		t.Fatal(err)
	}
	fg, _ = term.ColorForeground()
	if fg != nil {
		t.Fatal("expected nil foreground after clearing")
	}
}

func TestTerminalPaletteRoundTrip(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Read default palette.
	p, err := term.ColorPalette()
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("expected non-nil palette")
	}

	// Set a custom palette and read it back.
	var custom Palette
	for i := range custom {
		custom[i] = ColorRGB{R: uint8(i), G: 0, B: 0}
	}
	if err := term.SetColorPalette(&custom); err != nil {
		t.Fatal(err)
	}
	p, _ = term.ColorPalette()
	if p[0] != (ColorRGB{R: 0, G: 0, B: 0}) {
		t.Fatalf("expected palette[0] = {0,0,0}, got %+v", p[0])
	}
	if p[128] != (ColorRGB{R: 128, G: 0, B: 0}) {
		t.Fatalf("expected palette[128] = {128,0,0}, got %+v", p[128])
	}
}
