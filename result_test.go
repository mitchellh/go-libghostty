package libghostty

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		result Result
		want   string
	}{
		{ResultOutOfMemory, "ghostty: out of memory"},
		{ResultInvalidValue, "ghostty: invalid value"},
		{ResultOutOfSpace, "ghostty: out of space"},
		{ResultNoValue, "ghostty: no value"},
		{ResultIOError, "ghostty: I/O error"},
		{ResultLimitExceeded, "ghostty: limit exceeded"},
		{ResultRejected, "ghostty: rejected"},
		{Result(9999), "ghostty: result=9999"},
	}

	for _, tt := range tests {
		e := &Error{Result: tt.result}
		if got := e.Error(); got != tt.want {
			t.Errorf("Error{%d}.Error() = %q, want %q", tt.result, got, tt.want)
		}
	}
}

func TestErrorAs(t *testing.T) {
	var err error = &Error{Result: ResultOutOfMemory}

	var target *Error
	if !errors.As(err, &target) {
		t.Fatal("expected errors.As to succeed")
	}
	if target.Result != ResultOutOfMemory {
		t.Fatalf("expected ResultOutOfMemory, got %d", target.Result)
	}
}

func TestErrorAsType(t *testing.T) {
	var err error = &Error{Result: ResultInvalidValue}

	target, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatal("expected errors.AsType to succeed")
	}
	if target.Result != ResultInvalidValue {
		t.Fatalf("expected ResultInvalidValue, got %d", target.Result)
	}
}

func TestErrorIsSentinel(t *testing.T) {
	tests := []struct {
		result   Result
		sentinel error
	}{
		{ResultOutOfMemory, ErrOutOfMemory},
		{ResultInvalidValue, ErrInvalidValue},
		{ResultOutOfSpace, ErrOutOfSpace},
		{ResultNoValue, ErrNoValue},
		{ResultIOError, ErrIO},
		{ResultLimitExceeded, ErrLimitExceeded},
		{ResultRejected, ErrRejected},
	}

	for _, tt := range tests {
		// A freshly allocated error must match its sentinel, and only
		// its sentinel.
		var err error = &Error{Result: tt.result}
		if !errors.Is(err, tt.sentinel) {
			t.Errorf("errors.Is(Error{%d}, sentinel) = false, want true", tt.result)
		}
		for _, other := range tests {
			if other.result != tt.result && errors.Is(err, other.sentinel) {
				t.Errorf("Error{%d} matched sentinel for %d", tt.result, other.result)
			}
		}
	}
}

func TestErrorIsWrapped(t *testing.T) {
	// Sentinels must match through fmt.Errorf %w wrapping, which is how
	// callers add context to errors from this package.
	err := fmt.Errorf("reading cell: %w", &Error{Result: ResultNoValue})
	if !errors.Is(err, ErrNoValue) {
		t.Fatal("expected wrapped error to match ErrNoValue")
	}
	if errors.Is(err, ErrInvalidValue) {
		t.Fatal("wrapped error matched the wrong sentinel")
	}
}

func TestErrorIsNonError(t *testing.T) {
	// Comparing against an unrelated error type must not panic or match.
	var err error = &Error{Result: ResultNoValue}
	if errors.Is(err, errors.New("no value")) {
		t.Fatal("matched an unrelated error")
	}
	if errors.Is(err, nil) {
		t.Fatal("matched nil")
	}
}

func TestResultErrorSentinel(t *testing.T) {
	// End to end: a real failing call must satisfy errors.Is.
	term, err := NewTerminal(WithSize(8, 3))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	_, err = term.TrackGridRef(Point{Tag: PointTagActive, X: 100, Y: 100})
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}
