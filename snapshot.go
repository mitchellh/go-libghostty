package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"errors"
	"io"
	"unsafe"
)

// SnapshotDecoderOption identifies a configurable snapshot decoder option.
// C: GhosttySnapshotDecoderOption
type SnapshotDecoderOption int

const (
	// SnapshotDecoderOptionMaxContinuationBytes is the largest non-ground VT
	// continuation the decoder will accept (size_t).
	SnapshotDecoderOptionMaxContinuationBytes SnapshotDecoderOption = C.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES
)

// SnapshotDecoderData identifies a data field available from a
// [SnapshotDecoder].
// C: GhosttySnapshotDecoderData
type SnapshotDecoderData int

const (
	// SnapshotDecoderDataInvalid is an invalid or sentinel query.
	SnapshotDecoderDataInvalid SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_INVALID

	// SnapshotDecoderDataMaxContinuationBytes is the decoder's current
	// continuation input limit (size_t).
	SnapshotDecoderDataMaxContinuationBytes SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_MAX_CONTINUATION_BYTES

	// SnapshotDecoderDataSourceOffset is the number of source bytes consumed
	// so far (size_t).
	SnapshotDecoderDataSourceOffset SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_SOURCE_OFFSET

	// SnapshotDecoderDataHistoryRowsPrimary is the complete logical primary
	// screen history extent (uint64_t).
	SnapshotDecoderDataHistoryRowsPrimary SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_HISTORY_ROWS_PRIMARY

	// SnapshotDecoderDataHistoryRowsAlternate is the complete logical
	// alternate screen history extent (uint64_t).
	SnapshotDecoderDataHistoryRowsAlternate SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_HISTORY_ROWS_ALTERNATE

	// SnapshotDecoderDataProgressScreen is the screen associated with the most
	// recently decoded history page (GhosttyTerminalScreen).
	SnapshotDecoderDataProgressScreen SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_SCREEN

	// SnapshotDecoderDataProgressRows is the number of rows prepended by the
	// most recently decoded history page (size_t).
	SnapshotDecoderDataProgressRows SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_ROWS

	// SnapshotDecoderDataProgressRemaining is the number of page records
	// remaining in the current screen's history sequence (uint32_t).
	SnapshotDecoderDataProgressRemaining SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_REMAINING
)

// SnapshotDecoder incrementally decodes and authenticates one terminal
// snapshot. It is stateful, not safe for concurrent use, and must be closed.
//
// A terminal returned by [SnapshotDecoder.Ready] must remain alive until the
// decoder reaches FINISH or is closed. Closing the decoder does not close any
// returned terminal.
// C: GhosttySnapshotDecoder
type SnapshotDecoder struct {
	ptr C.GhosttySnapshotDecoder

	// reader is retained for the lifetime of a callback-backed decoder because
	// the C decoder stores its GhosttyReader descriptor by value.
	reader *ghosttyReaderBridge

	// sourcePtr is a C-owned copy used by NewSnapshotDecoderBytes. C retains
	// this borrowed pointer until FINISH or decoder destruction.
	sourcePtr unsafe.Pointer
	sourceLen uintptr
}

// Snapshot encodes a complete terminal snapshot and returns a Go-owned copy.
// Calls must be serialized with every other operation on t.
// C: ghostty_snapshot_encode_alloc
func (t *Terminal) Snapshot() ([]byte, error) {
	var outPtr *C.uint8_t
	var outLen C.size_t
	if err := resultError(C.ghostty_snapshot_encode_alloc(
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

// SnapshotBuf encodes a complete terminal snapshot into buf and returns the
// number of bytes written. If buf is nil or too small, the returned count is
// the required capacity and the error has result [ResultOutOfSpace].
// C: ghostty_snapshot_encode_buf
func (t *Terminal) SnapshotBuf(buf []byte) (int, error) {
	var ptr *C.uint8_t
	if len(buf) > 0 {
		ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
	}

	var written C.size_t
	result := C.ghostty_snapshot_encode_buf(
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

// SnapshotWriteTo encodes a complete terminal snapshot directly to w. The
// writer is called synchronously and must not call methods on t. The returned
// count includes bytes accepted before an error.
// C: ghostty_snapshot_encode
func (t *Terminal) SnapshotWriteTo(w io.Writer) (int64, error) {
	bridge, writer, err := newGhosttyWriter(w)
	if err != nil {
		return 0, err
	}
	defer bridge.close()

	result := C.ghostty_snapshot_encode(t.ptr, writer)
	return bridge.written, resultErrorWithCallback(result, bridge.err)
}

// NewSnapshotDecoder creates a snapshot decoder backed by r. Reads are
// synchronous and occur only during Ready, Next, or Decode. A successful
// zero-byte read is permanent EOF; nonblocking readers must wait internally.
// C: ghostty_snapshot_decoder_new
func NewSnapshotDecoder(r io.Reader) (*SnapshotDecoder, error) {
	bridge, reader, err := newGhosttyReader(r)
	if err != nil {
		return nil, err
	}

	var ptr C.GhosttySnapshotDecoder
	if err := resultError(C.ghostty_snapshot_decoder_new(nil, &ptr, reader)); err != nil {
		bridge.close()
		return nil, err
	}

	return &SnapshotDecoder{ptr: ptr, reader: bridge}, nil
}

// NewSnapshotDecoderBytes creates a snapshot decoder over a C-owned copy of
// data. Copying is required because libghostty retains the source pointer
// across calls and C must not retain a pointer into Go memory.
// C: ghostty_snapshot_decoder_new_buf
func NewSnapshotDecoderBytes(data []byte) (*SnapshotDecoder, error) {
	length := uintptr(len(data))
	var source unsafe.Pointer
	if length > 0 {
		source = Alloc(length)
		if source == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(source), len(data)), data)
	}

	var ptr C.GhosttySnapshotDecoder
	if err := resultError(C.ghostty_snapshot_decoder_new_buf(
		nil,
		&ptr,
		(*C.uint8_t)(source),
		C.size_t(length),
	)); err != nil {
		Free(source, length)
		return nil, err
	}

	return &SnapshotDecoder{
		ptr:       ptr,
		sourcePtr: source,
		sourceLen: length,
	}, nil
}

// Close frees the decoder and releases any retained reader handle or copied
// source buffer. It does not close a terminal returned by Ready or Decode.
// C: ghostty_snapshot_decoder_free
func (d *SnapshotDecoder) Close() {
	if d == nil || d.ptr == nil {
		return
	}

	C.ghostty_snapshot_decoder_free(d.ptr)
	d.ptr = nil
	d.reader.close()
	d.reader = nil
	if d.sourcePtr != nil {
		Free(d.sourcePtr, d.sourceLen)
		d.sourcePtr = nil
		d.sourceLen = 0
	}
}

// SetMaxContinuationBytes sets the largest non-ground VT continuation the
// decoder will accept. It may only be called before decoding begins. Zero
// accepts only snapshots whose parser and UTF-8 decoder are at ground.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetMaxContinuationBytes(limit uint) error {
	v := C.size_t(limit)
	return resultError(C.ghostty_snapshot_decoder_set(
		d.ptr,
		C.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES,
		unsafe.Pointer(&v),
	))
}

// Ready decodes and authenticates the renderable snapshot prefix. The
// returned terminal is caller-owned and immediately usable; older history can
// then be restored one page at a time with [SnapshotDecoder.Next].
// C: ghostty_snapshot_decoder_ready
func (d *SnapshotDecoder) Ready() (*Terminal, error) {
	var terminal C.GhosttyTerminal
	result := C.ghostty_snapshot_decoder_ready(d.ptr, &terminal)
	if err := d.decodeError(result); err != nil {
		return nil, err
	}
	return terminalFromC(terminal, TerminalConfig{}), nil
}

// Next decodes and authenticates one history page. It returns true after a
// page is consumed. It returns false with no error after FINISH is validated,
// including on repeated calls after FINISH.
// C: ghostty_snapshot_decoder_next
func (d *SnapshotDecoder) Next() (bool, error) {
	result := C.ghostty_snapshot_decoder_next(d.ptr)
	switch result {
	case C.GHOSTTY_SUCCESS:
		return true, nil
	case C.GHOSTTY_NO_VALUE:
		return false, nil
	default:
		return false, d.decodeError(result)
	}
}

// Decode decodes and authenticates one complete snapshot in a single call.
// The returned terminal is caller-owned. Bytes following the FINISH record
// remain outside the snapshot and can be located with SourceOffset.
// C: ghostty_snapshot_decoder_decode
func (d *SnapshotDecoder) Decode() (*Terminal, error) {
	var terminal C.GhosttyTerminal
	result := C.ghostty_snapshot_decoder_decode(d.ptr, &terminal)
	if err := d.decodeError(result); err != nil {
		return nil, err
	}
	return terminalFromC(terminal, TerminalConfig{}), nil
}

// Get reads a raw decoder data field into value. The pointer type must match
// the documented type for data; prefer the typed helpers when possible.
// C: ghostty_snapshot_decoder_get
func (d *SnapshotDecoder) Get(data SnapshotDecoderData, value unsafe.Pointer) error {
	return resultError(C.ghostty_snapshot_decoder_get(
		d.ptr,
		C.GhosttySnapshotDecoderData(data),
		value,
	))
}

// GetMulti reads several raw decoder data fields in one C call. Each value
// must point to storage with the type documented by its corresponding key.
// On partial failure, written is the failing index.
// C: ghostty_snapshot_decoder_get_multi
func (d *SnapshotDecoder) GetMulti(keys []SnapshotDecoderData, values []unsafe.Pointer) (written int, err error) {
	if len(keys) != len(values) {
		return 0, errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	cKeys := make([]C.GhosttySnapshotDecoderData, len(keys))
	for i, key := range keys {
		cKeys[i] = C.GhosttySnapshotDecoderData(key)
	}
	cValues, cValuesSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cValues), cValuesSize)
	var cWritten C.size_t
	err = resultError(C.ghostty_snapshot_decoder_get_multi(
		d.ptr,
		C.size_t(len(keys)),
		&cKeys[0],
		cValues,
		&cWritten,
	))
	return int(cWritten), err
}

// MaxContinuationBytes returns the decoder's current continuation input limit.
func (d *SnapshotDecoder) MaxContinuationBytes() (uint, error) {
	var value C.size_t
	if err := d.Get(
		SnapshotDecoderDataMaxContinuationBytes,
		unsafe.Pointer(&value),
	); err != nil {
		return 0, err
	}
	return uint(value), nil
}

// SourceOffset returns the number of snapshot source bytes consumed so far.
// At FINISH it identifies the first byte after the snapshot.
func (d *SnapshotDecoder) SourceOffset() (uint, error) {
	var value C.size_t
	if err := d.Get(SnapshotDecoderDataSourceOffset, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint(value), nil
}

// HistoryRows returns the snapshot's complete logical history extent for
// screen. It becomes available after Ready validates. The alternate screen
// returns [ResultNoValue] when the snapshot does not declare one.
func (d *SnapshotDecoder) HistoryRows(screen TerminalScreen) (uint64, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataHistoryRowsPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataHistoryRowsAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	var value C.uint64_t
	if err := d.Get(data, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint64(value), nil
}

// ProgressScreen returns the screen associated with the most recently
// decoded history page.
func (d *SnapshotDecoder) ProgressScreen() (TerminalScreen, error) {
	var value C.GhosttyTerminalScreen
	if err := d.Get(SnapshotDecoderDataProgressScreen, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return TerminalScreen(value), nil
}

// ProgressRows returns the number of rows prepended by the most recently
// decoded history page. Zero means the page could not be applied safely.
func (d *SnapshotDecoder) ProgressRows() (uint, error) {
	var value C.size_t
	if err := d.Get(SnapshotDecoderDataProgressRows, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint(value), nil
}

// ProgressRemaining returns the number of page records remaining in the same
// screen's history sequence.
func (d *SnapshotDecoder) ProgressRemaining() (uint32, error) {
	var value C.uint32_t
	if err := d.Get(SnapshotDecoderDataProgressRemaining, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// SnapshotDecoderProgress describes the most recently decoded history page.
type SnapshotDecoderProgress struct {
	// Screen is the screen whose history page was consumed.
	Screen TerminalScreen

	// Rows is the number of rows prepended to the live terminal.
	Rows uint

	// Remaining is the number of page records left in this screen's history.
	Remaining uint32
}

// Progress returns all progress fields for the most recently decoded history
// page in one C call.
func (d *SnapshotDecoder) Progress() (SnapshotDecoderProgress, error) {
	var screen C.GhosttyTerminalScreen
	var rows C.size_t
	var remaining C.uint32_t
	written, err := d.GetMulti(
		[]SnapshotDecoderData{
			SnapshotDecoderDataProgressScreen,
			SnapshotDecoderDataProgressRows,
			SnapshotDecoderDataProgressRemaining,
		},
		[]unsafe.Pointer{
			unsafe.Pointer(&screen),
			unsafe.Pointer(&rows),
			unsafe.Pointer(&remaining),
		},
	)
	if err != nil {
		return SnapshotDecoderProgress{}, err
	}
	if written != 3 {
		return SnapshotDecoderProgress{}, &Error{Result: ResultInvalidValue}
	}
	return SnapshotDecoderProgress{
		Screen:    TerminalScreen(screen),
		Rows:      uint(rows),
		Remaining: uint32(remaining),
	}, nil
}

// decodeError combines GHOSTTY_IO_ERROR with the original Go reader error so
// callers can inspect both through errors.Is/As.
func (d *SnapshotDecoder) decodeError(result C.GhosttyResult) error {
	var callbackErr error
	if d.reader != nil {
		callbackErr = d.reader.err
	}
	return resultErrorWithCallback(result, callbackErr)
}
