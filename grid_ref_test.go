package libghostty

import (
	"errors"
	"testing"
)

func TestGridRefHyperlinkURIAndPointFromGridRef(t *testing.T) {
	term, err := NewTerminal(WithSize(10, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 8 starts a hyperlink, prints one cell, then ends the hyperlink.
	term.VTWrite([]byte("\x1b]8;;https://example.com\x1b\\A\x1b]8;;\x1b\\"))

	ref, err := term.GridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}

	uri, err := ref.HyperlinkURI()
	if err != nil {
		t.Fatal(err)
	}
	if uri != "https://example.com" {
		t.Fatalf("expected hyperlink URI %q, got %q", "https://example.com", uri)
	}

	point, err := term.PointFromGridRef(ref, PointTagActive)
	if err != nil {
		t.Fatal(err)
	}
	if point != (Point{Tag: PointTagActive, X: 0, Y: 0}) {
		t.Fatalf("expected active point 0,0, got %+v", point)
	}

	emptyRef, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	uri, err = emptyRef.HyperlinkURI()
	if err != nil {
		t.Fatal(err)
	}
	if uri != "" {
		t.Fatalf("expected empty hyperlink URI for non-hyperlinked cell, got %q", uri)
	}
}

func TestTerminalPointFromGridRefNoValue(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3), WithMaxScrollback(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("alpha\r\nbravo\r\ncharlie\r\ndelta"))

	ref, err := term.GridRef(Point{Tag: PointTagHistory, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	_, err = term.PointFromGridRef(ref, PointTagActive)
	assertResultError(t, err, ResultNoValue)
}

func TestTerminalTrackGridRefInvalidPoint(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	_, err = term.TrackGridRef(Point{Tag: PointTagActive, X: 0, Y: 99})
	assertResultError(t, err, ResultInvalidValue)
}

func TestTrackedGridRef(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3), WithMaxScrollback(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("alpha\r\nbravo\r\ncharlie"))

	tracked, err := term.TrackGridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer tracked.Close()

	// Force the original first row into scrollback. The tracked reference
	// should continue to point at the same cell.
	term.VTWrite([]byte("\r\ndelta"))

	if !tracked.HasValue() {
		t.Fatal("expected tracked ref to retain value after scroll")
	}

	snapshot, err := tracked.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	cell, err := snapshot.Cell()
	if err != nil {
		t.Fatal(err)
	}
	cp, err := cell.Codepoint()
	if err != nil {
		t.Fatal(err)
	}
	if cp != 'a' {
		t.Fatalf("expected tracked codepoint %q after scroll, got %q", 'a', rune(cp))
	}

	point, err := tracked.Point(PointTagScreen)
	if err != nil {
		t.Fatal(err)
	}
	if point != (Point{Tag: PointTagScreen, X: 0, Y: 0}) {
		t.Fatalf("expected tracked screen point 0,0, got %+v", point)
	}

	term.Reset()
	if tracked.HasValue() {
		t.Fatal("expected tracked ref to lose value after reset")
	}
	_, err = tracked.Snapshot()
	assertResultError(t, err, ResultNoValue)
	_, err = tracked.Point(PointTagScreen)
	assertResultError(t, err, ResultNoValue)

	term.VTWrite([]byte("echo"))
	if err := tracked.Set(term, Point{Tag: PointTagActive, X: 0, Y: 0}); err != nil {
		t.Fatal(err)
	}
	if !tracked.HasValue() {
		t.Fatal("expected tracked ref to have value after set")
	}

	snapshot, err = tracked.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	cell, err = snapshot.Cell()
	if err != nil {
		t.Fatal(err)
	}
	cp, err = cell.Codepoint()
	if err != nil {
		t.Fatal(err)
	}
	if cp != 'e' {
		t.Fatalf("expected tracked codepoint %q after set, got %q", 'e', rune(cp))
	}
}

func TestTrackedGridRefClose(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	tracked, err := term.TrackGridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if tracked.ptr == nil {
		t.Fatal("expected tracked ref pointer before close")
	}

	tracked.Close()
	if tracked.ptr != nil {
		t.Fatal("expected tracked ref pointer to be nil after close")
	}

	// Close sets the handle to nil, and libghostty explicitly permits freeing a
	// nil tracked reference. This verifies the Go wrapper is safe to defer after
	// an explicit close.
	tracked.Close()
}

func TestTrackedGridRefAfterTerminalClose(t *testing.T) {
	term, err := NewTerminal(WithSize(8, 3))
	if err != nil {
		t.Fatal(err)
	}

	tracked, err := term.TrackGridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		term.Close()
		t.Fatal(err)
	}
	defer tracked.Close()

	term.Close()

	if tracked.HasValue() {
		t.Fatal("expected tracked ref to lose value after terminal close")
	}
	_, err = tracked.Snapshot()
	assertResultError(t, err, ResultNoValue)
	_, err = tracked.Point(PointTagActive)
	assertResultError(t, err, ResultNoValue)
}

func assertResultError(t *testing.T, err error, result Result) {
	t.Helper()

	var ge *Error
	if !errors.As(err, &ge) || ge.Result != result {
		t.Fatalf("expected %v error, got %v", result, err)
	}
}
