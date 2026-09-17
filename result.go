package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import "fmt"

// Result represents a Ghostty result code.
//
// C: GhosttyResult
type Result int

const (
	// ResultSuccess indicates that an operation completed successfully.
	ResultSuccess Result = C.GHOSTTY_SUCCESS

	// ResultOutOfMemory indicates that an allocation failed.
	ResultOutOfMemory Result = C.GHOSTTY_OUT_OF_MEMORY

	// ResultInvalidValue indicates that an argument or encoded value was invalid.
	ResultInvalidValue Result = C.GHOSTTY_INVALID_VALUE

	// ResultOutOfSpace indicates that a caller-provided buffer was too small.
	ResultOutOfSpace Result = C.GHOSTTY_OUT_OF_SPACE

	// ResultNoValue indicates that the requested value is unavailable.
	ResultNoValue Result = C.GHOSTTY_NO_VALUE

	// ResultIOError indicates that external reader or writer I/O failed.
	ResultIOError Result = C.GHOSTTY_IO_ERROR

	// ResultLimitExceeded indicates that encoded input exceeded a configured limit.
	ResultLimitExceeded Result = C.GHOSTTY_LIMIT_EXCEEDED

	// ResultRejected indicates that an operation was rejected by a safety check.
	// The caller may confirm with the user and retry with the operation's allow
	// flag set.
	ResultRejected Result = C.GHOSTTY_REJECTED
)

// Error holds a non-success Ghostty result.
//
// Use [errors.As] to inspect the Result, or [errors.Is] against one of
// the Err* sentinels to test for a specific result code:
//
//	if errors.Is(err, libghostty.ErrNoValue) {
//		// The requested value is unavailable.
//	}
type Error struct {
	Result Result
}

// Sentinel errors, one per non-success [Result]. Compare with
// [errors.Is]. They match any [*Error] carrying the same Result,
// including errors returned by this package's functions and errors
// wrapped by callers with fmt.Errorf("...: %w", err).
var (
	// ErrOutOfMemory matches [ResultOutOfMemory].
	ErrOutOfMemory error = &Error{Result: ResultOutOfMemory}

	// ErrInvalidValue matches [ResultInvalidValue].
	ErrInvalidValue error = &Error{Result: ResultInvalidValue}

	// ErrOutOfSpace matches [ResultOutOfSpace].
	ErrOutOfSpace error = &Error{Result: ResultOutOfSpace}

	// ErrNoValue matches [ResultNoValue].
	ErrNoValue error = &Error{Result: ResultNoValue}

	// ErrIO matches [ResultIOError].
	ErrIO error = &Error{Result: ResultIOError}

	// ErrLimitExceeded matches [ResultLimitExceeded].
	ErrLimitExceeded error = &Error{Result: ResultLimitExceeded}

	// ErrRejected matches [ResultRejected].
	ErrRejected error = &Error{Result: ResultRejected}
)

// Is reports whether target is an [*Error] with the same Result. This
// lets [errors.Is] match the Err* sentinels even though every failure
// returns a freshly allocated Error.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Result == e.Result
}

func (e *Error) Error() string {
	switch e.Result {
	case ResultOutOfMemory:
		return "ghostty: out of memory"
	case ResultInvalidValue:
		return "ghostty: invalid value"
	case ResultOutOfSpace:
		return "ghostty: out of space"
	case ResultNoValue:
		return "ghostty: no value"
	case ResultIOError:
		return "ghostty: I/O error"
	case ResultLimitExceeded:
		return "ghostty: limit exceeded"
	case ResultRejected:
		return "ghostty: rejected"
	default:
		return fmt.Sprintf("ghostty: result=%d", int(e.Result))
	}
}

// Convert a result code to an error, returning nil on success.
func resultError(result C.GhosttyResult) error {
	if result == C.GHOSTTY_SUCCESS {
		return nil
	}

	return &Error{Result: Result(result)}
}
