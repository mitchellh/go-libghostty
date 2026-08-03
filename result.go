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
)

// Error holds a non-success Ghostty result.
type Error struct {
	Result Result
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
