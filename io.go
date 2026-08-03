package libghostty

// Synchronous Go I/O adapters for GhosttyReader and GhosttyWriter.
//
// libghostty invokes both callbacks only while the originating C API call is
// active. A cgo.Handle carries the Go reader or writer through C userdata
// without storing a Go pointer in C memory.

/*
#include <ghostty/vt.h>
#include <stdint.h>

extern bool goGhosttyReaderTrampoline(
	void* userdata,
	uint8_t* buffer,
	size_t capacity,
	size_t* out_read);

extern bool goGhosttyWriterTrampoline(
	void* userdata,
	uint8_t* data,
	size_t len);

static inline GhosttyReader ghostty_go_reader(uintptr_t userdata) {
	GhosttyReader reader = {
		.read = goGhosttyReaderTrampoline,
		.userdata = (void*)userdata,
	};
	return reader;
}

static inline GhosttyWriter ghostty_go_writer(uintptr_t userdata) {
	GhosttyWriter writer = {
		// cgo cannot export a const-qualified pointer parameter. The Go
		// trampoline treats data as read-only, so this cast is ABI-safe.
		.write = (GhosttyWriterFn)goGhosttyWriterTrampoline,
		.userdata = (void*)userdata,
	};
	return writer;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"runtime/cgo"
	"unsafe"
)

// ghosttyReaderBridge owns the cgo handle installed in a GhosttyReader. The
// decoder retains that C descriptor, so this bridge must live until the
// decoder is closed even though each individual read is synchronous.
type ghosttyReaderBridge struct {
	reader     io.Reader
	handle     cgo.Handle
	err        error
	pendingErr error
	eof        bool
}

// newGhosttyReader builds a C reader backed by r. The returned bridge must be
// closed after the C object that retains the reader has been freed.
func newGhosttyReader(r io.Reader) (*ghosttyReaderBridge, C.GhosttyReader, error) {
	if r == nil {
		return nil, C.GhosttyReader{}, &Error{Result: ResultInvalidValue}
	}

	b := &ghosttyReaderBridge{reader: r}
	b.handle = cgo.NewHandle(b)
	return b, C.ghostty_go_reader(C.uintptr_t(b.handle)), nil
}

// close releases the reader's cgo handle. It must be called only after C can
// no longer invoke the callback.
func (b *ghosttyReaderBridge) close() {
	if b == nil || b.handle == 0 {
		return
	}
	b.handle.Delete()
	b.handle = 0
}

// ghosttyWriterBridge owns the cgo handle used by one synchronous writer
// operation and records both the accepted byte count and the original Go
// error, if any.
type ghosttyWriterBridge struct {
	writer  io.Writer
	handle  cgo.Handle
	written int64
	err     error
}

// newGhosttyWriter builds a C writer backed by w. The caller must close the
// bridge after the synchronous libghostty operation returns.
func newGhosttyWriter(w io.Writer) (*ghosttyWriterBridge, C.GhosttyWriter, error) {
	if w == nil {
		return nil, C.GhosttyWriter{}, &Error{Result: ResultInvalidValue}
	}

	b := &ghosttyWriterBridge{writer: w}
	b.handle = cgo.NewHandle(b)
	return b, C.ghostty_go_writer(C.uintptr_t(b.handle)), nil
}

// close releases the writer's cgo handle.
func (b *ghosttyWriterBridge) close() {
	if b == nil || b.handle == 0 {
		return
	}
	b.handle.Delete()
	b.handle = 0
}

// resultErrorWithCallback preserves libghostty's I/O result while also
// exposing the original Go reader or writer error through errors.Is/As.
func resultErrorWithCallback(result C.GhosttyResult, callbackErr error) error {
	err := resultError(result)
	if result != C.GHOSTTY_IO_ERROR || callbackErr == nil {
		return err
	}
	return errors.Join(err, callbackErr)
}

//export goGhosttyReaderTrampoline
func goGhosttyReaderTrampoline(
	userdata unsafe.Pointer,
	buffer *C.uint8_t,
	capacity C.size_t,
	outRead *C.size_t,
) (ok C.bool) {
	b := cgo.Handle(userdata).Value().(*ghosttyReaderBridge)
	*outRead = 0

	// A panic must never unwind across the C boundary. Convert it to the same
	// fatal read failure that an ordinary Go reader error would produce.
	defer func() {
		if recovered := recover(); recovered != nil {
			b.err = fmt.Errorf("ghostty reader panic: %v", recovered)
			ok = C.bool(false)
		}
	}()

	if b.pendingErr != nil {
		b.err = b.pendingErr
		b.pendingErr = nil
		return C.bool(false)
	}
	if b.eof {
		return C.bool(true)
	}

	length, valid := ghosttySizeToInt(capacity)
	if !valid || length <= 0 || buffer == nil {
		b.err = &Error{Result: ResultLimitExceeded}
		return C.bool(false)
	}

	n, err := b.reader.Read(unsafe.Slice((*byte)(unsafe.Pointer(buffer)), length))
	if n < 0 || n > length {
		b.err = fmt.Errorf("ghostty reader returned invalid byte count %d for capacity %d", n, length)
		return C.bool(false)
	}
	*outRead = C.size_t(n)

	if err == nil {
		// GhosttyReader defines a successful zero-byte read as permanent EOF.
		b.eof = n == 0
		return C.bool(true)
	}
	if errors.Is(err, io.EOF) {
		// Deliver bytes returned with io.EOF first, then report definitive EOF
		// without consulting the Go reader again.
		b.eof = true
		return C.bool(true)
	}
	if n > 0 {
		// Go permits data and an error in the same read. Deliver the data now
		// and report the fatal error on libghostty's next callback.
		b.pendingErr = err
		return C.bool(true)
	}

	b.err = err
	return C.bool(false)
}

//export goGhosttyWriterTrampoline
func goGhosttyWriterTrampoline(
	userdata unsafe.Pointer,
	data *C.uint8_t,
	length C.size_t,
) (ok C.bool) {
	b := cgo.Handle(userdata).Value().(*ghosttyWriterBridge)

	// A panic must never unwind across the C boundary. Preserve it as a Go
	// error and make libghostty return GHOSTTY_IO_ERROR.
	defer func() {
		if recovered := recover(); recovered != nil {
			b.err = fmt.Errorf("ghostty writer panic: %v", recovered)
			ok = C.bool(false)
		}
	}()

	count, valid := ghosttySizeToInt(length)
	if !valid || count <= 0 || data == nil {
		b.err = &Error{Result: ResultLimitExceeded}
		return C.bool(false)
	}
	remaining := unsafe.Slice((*byte)(unsafe.Pointer(data)), count)

	for len(remaining) > 0 {
		n, err := b.writer.Write(remaining)
		if n < 0 || n > len(remaining) {
			b.err = fmt.Errorf("ghostty writer returned invalid byte count %d for length %d", n, len(remaining))
			return C.bool(false)
		}

		const maxInt64 = int64(^uint64(0) >> 1)
		if b.written > maxInt64-int64(n) {
			b.err = &Error{Result: ResultLimitExceeded}
			return C.bool(false)
		}
		b.written += int64(n)
		remaining = remaining[n:]

		if err != nil {
			b.err = err
			return C.bool(false)
		}
		if n == 0 {
			b.err = io.ErrShortWrite
			return C.bool(false)
		}
	}

	return C.bool(true)
}
