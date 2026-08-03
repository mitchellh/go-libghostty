package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"io"
	"unsafe"
)

// Continuation returns a Go-owned copy of the replay-safe VT byte suffix
// needed to reconstruct the terminal's unfinished parser or UTF-8 decoder
// state. It returns an empty slice when the stream is at ground.
//
// Continuation tracking must have been enabled with
// [WithContinuationMaxBytes] or [Terminal.SetContinuationMaxBytes] before the
// input that produced an unfinished state was written.
// C: ghostty_terminal_continuation_alloc
func (t *Terminal) Continuation() ([]byte, error) {
	var outPtr *C.uint8_t
	var outLen C.size_t
	if err := resultError(C.ghostty_terminal_continuation_alloc(
		t.ptr,
		nil,
		&outPtr,
		&outLen,
	)); err != nil {
		return nil, err
	}
	defer C.ghostty_free(nil, outPtr, outLen)

	length, valid := ghosttySizeToInt(outLen)
	if !valid {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	if length == 0 {
		return nil, nil
	}

	out := make([]byte, length)
	copy(out, unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), length))
	return out, nil
}

// ContinuationBuf copies the replay-safe VT continuation into buf and returns
// the number of bytes written. If buf is nil or too small, the returned count
// is the required size and the error has result [ResultOutOfSpace].
// C: ghostty_terminal_continuation_buf
func (t *Terminal) ContinuationBuf(buf []byte) (int, error) {
	var ptr *C.uint8_t
	if len(buf) > 0 {
		ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
	}

	var written C.size_t
	result := C.ghostty_terminal_continuation_buf(
		t.ptr,
		ptr,
		C.size_t(len(buf)),
		&written,
	)
	length, valid := ghosttySizeToInt(written)
	if !valid {
		return 0, &Error{Result: ResultLimitExceeded}
	}
	return length, resultError(result)
}

// ContinuationWriteTo writes the replay-safe VT continuation directly to w.
// The writer is called synchronously and must not call methods on t. The
// returned count includes bytes accepted before an error.
// C: ghostty_terminal_continuation_write
func (t *Terminal) ContinuationWriteTo(w io.Writer) (int64, error) {
	bridge, writer, err := newGhosttyWriter(w)
	if err != nil {
		return 0, err
	}
	defer bridge.close()

	result := C.ghostty_terminal_continuation_write(t.ptr, writer)
	return bridge.written, resultErrorWithCallback(result, bridge.err)
}
