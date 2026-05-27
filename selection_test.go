package libghostty

import (
	"errors"
	"testing"
	"unsafe"
)

func TestTerminalSelectionAPIs(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 5))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("hello world\nsecond"))

	wordRef, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	word, err := term.SelectWord(SelectWordOptions{Ref: wordRef})
	if err != nil {
		t.Fatal(err)
	}
	if word == nil {
		t.Fatal("expected word selection")
	}

	formatted, err := term.SelectionFormatString(WithSelection(word))
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "hello" {
		t.Fatalf("expected formatted word %q, got %q", "hello", formatted)
	}

	required, err := term.SelectionFormatBuf(nil, WithSelection(word))
	if err == nil {
		t.Fatal("expected out-of-space error from size query")
	}
	var ge *Error
	if !errors.As(err, &ge) || ge.Result != ResultOutOfSpace {
		t.Fatalf("expected out-of-space error, got %v", err)
	}
	if required != len("hello") {
		t.Fatalf("expected required size %d, got %d", len("hello"), required)
	}

	buf := make([]byte, required)
	n, err := term.SelectionFormatBuf(buf, WithSelection(word))
	if err != nil {
		t.Fatal(err)
	}
	if n != required || string(buf[:n]) != "hello" {
		t.Fatalf("expected buffered word %q, got %q (%d bytes)", "hello", string(buf[:n]), n)
	}

	lineRef, err := term.GridRef(Point{Tag: PointTagActive, X: 2, Y: 1})
	if err != nil {
		t.Fatal(err)
	}
	line, err := term.SelectLine(SelectLineOptions{Ref: lineRef})
	if err != nil {
		t.Fatal(err)
	}
	if line == nil {
		t.Fatal("expected line selection")
	}
	formatted, err = term.SelectionFormatString(WithSelection(line), WithSelectionTrim(true))
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "second" {
		t.Fatalf("expected formatted line %q, got %q", "second", formatted)
	}

	all, err := term.SelectAll()
	if err != nil {
		t.Fatal(err)
	}
	if all == nil {
		t.Fatal("expected select-all selection")
	}

	order, err := term.SelectionOrder(word)
	if err != nil {
		t.Fatal(err)
	}
	if order != SelectionOrderForward {
		t.Fatalf("expected forward selection order, got %d", order)
	}

	ordered, err := term.SelectionOrdered(word, SelectionOrderForward)
	if err != nil {
		t.Fatal(err)
	}
	equal, err := term.SelectionEqual(word, ordered)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("expected ordered selection to equal original")
	}

	contains, err := term.SelectionContains(word, Point{Tag: PointTagActive, X: 2, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !contains {
		t.Fatal("expected word selection to contain point")
	}
	contains, err = term.SelectionContains(word, Point{Tag: PointTagActive, X: 8, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if contains {
		t.Fatal("expected word selection not to contain point")
	}

	adjusted := *word
	if err := term.SelectionAdjust(&adjusted, SelectionAdjustRight); err != nil {
		t.Fatal(err)
	}

	betweenStart, err := term.GridRef(Point{Tag: PointTagActive, X: 6, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	betweenEnd, err := term.GridRef(Point{Tag: PointTagActive, X: 9, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	between, err := term.SelectWordBetween(SelectWordBetweenOptions{
		Start: betweenStart,
		End:   betweenEnd,
	})
	if err != nil {
		t.Fatal(err)
	}
	if between == nil {
		t.Fatal("expected word-between selection")
	}
	formatted, err = term.SelectionFormatString(WithSelection(between))
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "world" {
		t.Fatalf("expected formatted word-between %q, got %q", "world", formatted)
	}

	output, err := term.SelectOutput(SelectOutputOptions{Ref: wordRef})
	if err != nil {
		t.Fatal(err)
	}
	if output != nil {
		formatted, err = term.SelectionFormatString(WithSelection(output), WithSelectionTrim(true))
		if err != nil {
			t.Fatal(err)
		}
		if formatted == "" {
			t.Fatal("expected non-empty command-output formatting")
		}
	}
}

func TestTerminalActiveSelectionFormat(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 5))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	formatted, err := term.SelectionFormatString()
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "" {
		t.Fatalf("expected no active selection to format empty string, got %q", formatted)
	}

	term.VTWrite([]byte("hello"))
	ref, err := term.GridRef(Point{Tag: PointTagActive, X: 1, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	word, err := term.SelectWord(SelectWordOptions{Ref: ref})
	if err != nil {
		t.Fatal(err)
	}
	if err := term.SetSelection(word); err != nil {
		t.Fatal(err)
	}

	formatted, err = term.SelectionFormatString()
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "hello" {
		t.Fatalf("expected active selection format %q, got %q", "hello", formatted)
	}
}

func TestSelectionGestureAPIs(t *testing.T) {
	term, err := NewTerminal(WithSize(20, 4), WithMaxScrollback(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("hello world\r\nsecond line"))

	gesture, err := NewSelectionGesture()
	if err != nil {
		t.Fatal(err)
	}
	defer gesture.Close(term)

	press, err := NewSelectionGestureEvent(SelectionGestureEventTypePress)
	if err != nil {
		t.Fatal(err)
	}
	defer press.Close()

	drag, err := NewSelectionGestureEvent(SelectionGestureEventTypeDrag)
	if err != nil {
		t.Fatal(err)
	}
	defer drag.Close()

	release, err := NewSelectionGestureEvent(SelectionGestureEventTypeRelease)
	if err != nil {
		t.Fatal(err)
	}
	defer release.Close()

	deepPress, err := NewSelectionGestureEvent(SelectionGestureEventTypeDeepPress)
	if err != nil {
		t.Fatal(err)
	}
	defer deepPress.Close()

	geometry := SelectionGestureGeometry{
		Columns:      20,
		CellWidth:    10,
		ScreenHeight: 40,
	}

	pressRef, err := term.GridRef(Point{Tag: PointTagActive, X: 0, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := press.SetRef(pressRef); err != nil {
		t.Fatal(err)
	}
	if err := press.SetPosition(SurfacePosition{X: 2, Y: 8}); err != nil {
		t.Fatal(err)
	}
	if err := press.SetTimeNS(1); err != nil {
		t.Fatal(err)
	}
	if err := press.SetRepeatDistance(4); err != nil {
		t.Fatal(err)
	}
	if err := press.SetRepeatIntervalNS(500); err != nil {
		t.Fatal(err)
	}

	selection, err := gesture.Event(term, press)
	if err != nil {
		t.Fatal(err)
	}
	if selection != nil {
		t.Fatal("expected press not to produce a selection")
	}

	anchor, err := gesture.Anchor(term)
	if err != nil {
		t.Fatal(err)
	}
	if anchor == nil {
		t.Fatal("expected press to record an anchor")
	}

	dragRef, err := term.GridRef(Point{Tag: PointTagActive, X: 4, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := drag.SetRef(dragRef); err != nil {
		t.Fatal(err)
	}
	if err := drag.SetPosition(SurfacePosition{X: 46, Y: 8}); err != nil {
		t.Fatal(err)
	}
	if err := drag.SetGeometry(geometry); err != nil {
		t.Fatal(err)
	}
	if err := drag.SetRectangle(false); err != nil {
		t.Fatal(err)
	}

	selection, err = gesture.Event(term, drag)
	if err != nil {
		t.Fatal(err)
	}
	if selection == nil {
		t.Fatal("expected drag selection")
	}
	formatted, err := term.SelectionFormatString(WithSelection(selection))
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "hello" {
		t.Fatalf("expected drag selection %q, got %q", "hello", formatted)
	}

	dragged, err := gesture.Dragged(term)
	if err != nil {
		t.Fatal(err)
	}
	if !dragged {
		t.Fatal("expected dragged state")
	}
	clickCount, err := gesture.ClickCount(term)
	if err != nil {
		t.Fatal(err)
	}
	if clickCount != 1 {
		t.Fatalf("expected click count 1, got %d", clickCount)
	}
	behavior, err := gesture.Behavior(term)
	if err != nil {
		t.Fatal(err)
	}
	if behavior != SelectionGestureBehaviorCell {
		t.Fatalf("expected cell behavior, got %d", behavior)
	}

	var multiClickCount uint8
	var multiDragged bool
	written, err := gesture.GetMulti(
		term,
		[]SelectionGestureData{SelectionGestureDataClickCount, SelectionGestureDataDragged},
		[]unsafe.Pointer{unsafe.Pointer(&multiClickCount), unsafe.Pointer(&multiDragged)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if written != 2 || multiClickCount != 1 || !multiDragged {
		t.Fatalf("unexpected get_multi result: written=%d click_count=%d dragged=%t", written, multiClickCount, multiDragged)
	}

	if err := release.SetRef(dragRef); err != nil {
		t.Fatal(err)
	}
	selection, err = gesture.Event(term, release)
	if err != nil {
		t.Fatal(err)
	}
	if selection != nil {
		t.Fatal("expected release not to produce a selection")
	}

	gesture.Reset(term)
	worldRef, err := term.GridRef(Point{Tag: PointTagActive, X: 6, Y: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := press.SetRef(worldRef); err != nil {
		t.Fatal(err)
	}
	if err := press.ClearTimeNS(); err != nil {
		t.Fatal(err)
	}
	selection, err = gesture.Event(term, press)
	if err != nil {
		t.Fatal(err)
	}
	if selection != nil {
		t.Fatal("expected second press not to produce a selection")
	}

	selection, err = gesture.Event(term, deepPress)
	if err != nil {
		t.Fatal(err)
	}
	if selection == nil {
		t.Fatal("expected deep press selection")
	}
	formatted, err = term.SelectionFormatString(WithSelection(selection))
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "world" {
		t.Fatalf("expected deep press selection %q, got %q", "world", formatted)
	}
}
