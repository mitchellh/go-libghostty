package libghostty

import "testing"

// TestCloseIdempotent verifies that every handle type tolerates a repeated
// Close. Each C free must run exactly once; a second Close must be a no-op
// rather than a double free. The test runs under the Go race detector and
// ASan-less builds, so a double free would typically crash the process
// rather than fail an assertion, which is why simply calling Close twice
// is a sufficient check.
func TestCloseIdempotent(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}

	// Handles that depend on the terminal must be created before the
	// terminal is closed, and closing them twice must not touch the
	// (still open) terminal more than once.
	formatter, err := NewFormatter(term)
	if err != nil {
		t.Fatalf("NewFormatter: %v", err)
	}
	search, err := NewSearch(term)
	if err != nil {
		t.Fatalf("NewSearch: %v", err)
	}
	tracked, err := term.TrackGridRef(Point{Tag: PointTagActive})
	if err != nil {
		t.Fatalf("TrackGridRef: %v", err)
	}

	closers := []struct {
		name string
		make func() (interface{ Close() }, error)
	}{
		{"RenderState", func() (interface{ Close() }, error) { return NewRenderState() }},
		{"RenderStateRowIterator", func() (interface{ Close() }, error) { return NewRenderStateRowIterator() }},
		{"RenderStateRowCells", func() (interface{ Close() }, error) { return NewRenderStateRowCells() }},
		{"KeyEncoder", func() (interface{ Close() }, error) { return NewKeyEncoder() }},
		{"KeyEvent", func() (interface{ Close() }, error) { return NewKeyEvent() }},
		{"MouseEncoder", func() (interface{ Close() }, error) { return NewMouseEncoder() }},
		{"MouseEvent", func() (interface{ Close() }, error) { return NewMouseEvent() }},
		{"OSCParser", func() (interface{ Close() }, error) { return NewOSCParser() }},
		{"SGRParser", func() (interface{ Close() }, error) { return NewSGRParser() }},
		{"KittyGraphicsPlacementIterator", func() (interface{ Close() }, error) { return NewKittyGraphicsPlacementIterator() }},
		{"SelectionGestureEvent", func() (interface{ Close() }, error) {
			return NewSelectionGestureEvent(SelectionGestureEventTypePress)
		}},
		{"Formatter", func() (interface{ Close() }, error) { return formatter, nil }},
		{"Search", func() (interface{ Close() }, error) { return search, nil }},
		{"TrackedGridRef", func() (interface{ Close() }, error) { return tracked, nil }},
		{"Terminal", func() (interface{ Close() }, error) { return term, nil }},
	}

	for _, tc := range closers {
		t.Run(tc.name, func(t *testing.T) {
			c, err := tc.make()
			if err != nil {
				t.Fatalf("construct: %v", err)
			}
			c.Close()
			c.Close()
		})
	}
}

// TestCloseNilReceiver verifies that Close on a nil pointer is a no-op so
// deferred cleanup in error paths does not need a nil check.
func TestCloseNilReceiver(t *testing.T) {
	(*Terminal)(nil).Close()
	(*Formatter)(nil).Close()
	(*Search)(nil).Close()
	(*RenderState)(nil).Close()
	(*RenderStateRowIterator)(nil).Close()
	(*RenderStateRowCells)(nil).Close()
	(*KeyEncoder)(nil).Close()
	(*KeyEvent)(nil).Close()
	(*MouseEncoder)(nil).Close()
	(*MouseEvent)(nil).Close()
	(*OSCParser)(nil).Close()
	(*SGRParser)(nil).Close()
	(*KittyGraphicsPlacementIterator)(nil).Close()
	(*SelectionGestureEvent)(nil).Close()
	(*TrackedGridRef)(nil).Close()
	(*SnapshotDecoder)(nil).Close()
}
