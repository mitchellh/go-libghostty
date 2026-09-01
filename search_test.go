package libghostty

import (
	"testing"
	"unsafe"
)

func TestSearchLifecycleAndResults(t *testing.T) {
	term, err := NewTerminal(
		WithSize(10, 4),
		WithMaxScrollbackLines(100),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("Fizz\r\nBuzz\r\nFIZZ"))

	search, err := NewSearch(term)
	if err != nil {
		t.Fatal(err)
	}
	defer search.Close()

	status, err := search.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status != SearchStatusComplete {
		t.Fatalf("expected idle search to be complete, got %d", status)
	}
	needle, err := search.Needle()
	if err != nil {
		t.Fatal(err)
	}
	if needle != nil {
		t.Fatalf("expected no idle needle, got %q", needle)
	}

	if err := search.SetNeedle([]byte("fizz")); err != nil {
		t.Fatal(err)
	}

	status, err = search.Tick()
	if err != nil {
		t.Fatal(err)
	}
	if status != SearchStatusFeedRequired {
		t.Fatalf("expected a feed-required search, got %d", status)
	}
	if err := search.Run(); err != nil {
		t.Fatal(err)
	}

	matchCount, err := search.MatchCount()
	if err != nil {
		t.Fatal(err)
	}
	if matchCount != 2 {
		t.Fatalf("expected 2 matches, got %d", matchCount)
	}
	if index, err := search.SelectedIndex(); err != nil {
		t.Fatal(err)
	} else if index != nil {
		t.Fatalf("expected no initial selection, got index %d", *index)
	}
	matches, err := search.Matches()
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 match snapshots, got %d", len(matches))
	}
	newestPoint, err := term.PointFromGridRef(&matches[0].Start, PointTagActive)
	if err != nil {
		t.Fatal(err)
	}
	oldestPoint, err := term.PointFromGridRef(&matches[1].Start, PointTagActive)
	if err != nil {
		t.Fatal(err)
	}
	if newestPoint.Y != 2 || oldestPoint.Y != 0 {
		t.Fatalf("expected newest-to-oldest rows 2,0; got %d,%d", newestPoint.Y, oldestPoint.Y)
	}

	viewportMatches, err := search.ViewportMatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(viewportMatches) != 2 {
		t.Fatalf("expected 2 viewport matches, got %d", len(viewportMatches))
	}

	if err := search.SelectNext(); err != nil {
		t.Fatal(err)
	}
	index, err := search.SelectedIndex()
	if err != nil {
		t.Fatal(err)
	}
	if index == nil || *index != 0 {
		t.Fatalf("expected selected index 0, got %v", index)
	}
	selectedMatch, err := search.SelectedMatch()
	if err != nil {
		t.Fatal(err)
	}
	if selectedMatch == nil {
		t.Fatal("expected a selected match")
	}

	var rawMatchCount, rawSelectedIndex uint
	written, err := search.GetMulti(
		[]SearchData{SearchDataMatchCount, SearchDataSelectedIndex},
		[]unsafe.Pointer{unsafe.Pointer(&rawMatchCount), unsafe.Pointer(&rawSelectedIndex)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if written != 2 || rawMatchCount != 2 || rawSelectedIndex != 0 {
		t.Fatalf(
			"unexpected batched values: written=%d matches=%d index=%d",
			written,
			rawMatchCount,
			rawSelectedIndex,
		)
	}

	if err := search.SelectNext(); err != nil {
		t.Fatal(err)
	}
	index, err = search.SelectedIndex()
	if err != nil {
		t.Fatal(err)
	}
	if index == nil || *index != 1 {
		t.Fatalf("expected selected index 1, got %v", index)
	}
	if err := search.SelectPrevious(); err != nil {
		t.Fatal(err)
	}
	index, err = search.SelectedIndex()
	if err != nil {
		t.Fatal(err)
	}
	if index == nil || *index != 0 {
		t.Fatalf("expected selected index 0 after previous, got %v", index)
	}

	if policy, err := search.ScrollPolicy(); err != nil {
		t.Fatal(err)
	} else if policy != SearchScrollPolicyIfNeeded {
		t.Fatalf("expected default scroll policy, got %d", policy)
	}
	if err := search.SetScrollPolicy(SearchScrollPolicyNever); err != nil {
		t.Fatal(err)
	}
	if policy, err := search.ScrollPolicy(); err != nil {
		t.Fatal(err)
	} else if policy != SearchScrollPolicyNever {
		t.Fatalf("expected no-scroll policy, got %d", policy)
	}
	if err := search.SetScrollPolicy(SearchScrollPolicyIfNeeded); err != nil {
		t.Fatal(err)
	}
	if policy, err := search.ScrollPolicy(); err != nil {
		t.Fatal(err)
	} else if policy != SearchScrollPolicyIfNeeded {
		t.Fatalf("expected restored scroll policy, got %d", policy)
	}

	if err := search.SetNeedle(nil); err != nil {
		t.Fatal(err)
	}
	if needle, err := search.Needle(); err != nil {
		t.Fatal(err)
	} else if needle != nil {
		t.Fatalf("expected cleared needle, got %q", needle)
	}
	if err := search.SelectNext(); err == nil {
		t.Fatal("expected selecting without matches to fail")
	} else {
		assertResultError(t, err, ResultNoValue)
	}
}

func TestSearchTerminalMayCloseFirst(t *testing.T) {
	term, err := NewTerminal(WithSize(10, 4))
	if err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte("searchable"))

	search, err := NewSearch(term)
	if err != nil {
		t.Fatal(err)
	}
	if err := search.SetNeedle([]byte("search")); err != nil {
		t.Fatal(err)
	}
	if err := search.Run(); err != nil {
		t.Fatal(err)
	}

	term.Close()
	assertResultError(t, search.Feed(), ResultInvalidValue)
	if needle, err := search.Needle(); err != nil {
		t.Fatal(err)
	} else if string(needle) != "search" {
		t.Fatalf("expected retained needle after terminal close, got %q", needle)
	}
	search.Close()
}
