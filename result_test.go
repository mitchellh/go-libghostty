package libghostty

import (
	"errors"
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
