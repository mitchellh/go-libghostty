package libghostty

/*
#include <ghostty/vt.h>
#include "go_io.h"

// Returning small result structs by value keeps cgo from forcing every C out
// parameter onto the Go heap. These are binding-only adapters; the underlying
// libghostty functions and ownership rules remain unchanged.
typedef struct {
	GhosttyResult result;
	uint8_t* ptr;
	size_t len;
} ghostty_go_snapshot_encode_alloc_result;

static inline ghostty_go_snapshot_encode_alloc_result
ghostty_go_snapshot_encode_alloc(GhosttyTerminal terminal) {
	ghostty_go_snapshot_encode_alloc_result out = {
		.result = GHOSTTY_SUCCESS,
		.ptr = NULL,
		.len = 0,
	};
	out.result = ghostty_snapshot_encode_alloc(
		terminal,
		NULL,
		&out.ptr,
		&out.len);
	return out;
}

typedef struct {
	GhosttyResult result;
	size_t written;
} ghostty_go_snapshot_encode_buf_result;

static inline ghostty_go_snapshot_encode_buf_result
ghostty_go_snapshot_encode_buf(
	GhosttyTerminal terminal,
	uint8_t* buf,
	size_t len
) {
	ghostty_go_snapshot_encode_buf_result out = {
		.result = GHOSTTY_SUCCESS,
		.written = 0,
	};
	out.result = ghostty_snapshot_encode_buf(
		terminal,
		buf,
		len,
		&out.written);
	return out;
}

typedef struct {
	GhosttyResult result;
	GhosttySnapshotDecoder decoder;
} ghostty_go_snapshot_decoder_result;

static inline ghostty_go_snapshot_decoder_result
ghostty_go_snapshot_decoder_new(uintptr_t userdata) {
	ghostty_go_snapshot_decoder_result out = {
		.result = GHOSTTY_SUCCESS,
		.decoder = NULL,
	};
	GhosttyReader reader = {
		.read = ghostty_go_reader_trampoline,
		.userdata = (void*)userdata,
	};
	out.result = ghostty_snapshot_decoder_new(NULL, &out.decoder, reader);
	return out;
}

static inline GhosttyResult ghostty_go_snapshot_encode(
	GhosttyTerminal terminal,
	uintptr_t userdata
) {
	return ghostty_snapshot_encode(terminal, ghostty_go_writer(userdata));
}

static inline ghostty_go_snapshot_decoder_result
ghostty_go_snapshot_decoder_new_buf(const uint8_t* ptr, size_t len) {
	ghostty_go_snapshot_decoder_result out = {
		.result = GHOSTTY_SUCCESS,
		.decoder = NULL,
	};
	out.result = ghostty_snapshot_decoder_new_buf(
		NULL,
		&out.decoder,
		ptr,
		len);
	return out;
}

typedef struct {
	GhosttyResult result;
	GhosttyTerminal terminal;
} ghostty_go_snapshot_decode_result;

static inline ghostty_go_snapshot_decode_result
ghostty_go_snapshot_decoder_ready(GhosttySnapshotDecoder decoder) {
	ghostty_go_snapshot_decode_result out = {
		.result = GHOSTTY_SUCCESS,
		.terminal = NULL,
	};
	out.result = ghostty_snapshot_decoder_ready(decoder, &out.terminal);
	return out;
}

static inline ghostty_go_snapshot_decode_result
ghostty_go_snapshot_decoder_decode(GhosttySnapshotDecoder decoder) {
	ghostty_go_snapshot_decode_result out = {
		.result = GHOSTTY_SUCCESS,
		.terminal = NULL,
	};
	out.result = ghostty_snapshot_decoder_decode(decoder, &out.terminal);
	return out;
}

// ghostty_go_snapshot_decoder_progress reads every progress value that
// applies to the most recently decoded record in one cgo call. Which values
// exist depends on the progress kind, so a get_multi call with a fixed list
// of keys would fail for one kind or the other.
typedef struct {
	GhosttyResult result;
	GhosttySnapshotDecoderProgressKind kind;
	GhosttyTerminalScreen screen;
	size_t rows;
	uint32_t remaining;
	uint32_t image_id;
	uint32_t image_frame;
	size_t image_bytes;
} ghostty_go_snapshot_decoder_progress_result;

static inline ghostty_go_snapshot_decoder_progress_result
ghostty_go_snapshot_decoder_progress(GhosttySnapshotDecoder decoder) {
	ghostty_go_snapshot_decoder_progress_result out = {0};
	out.result = ghostty_snapshot_decoder_get(
		decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_KIND, &out.kind);
	if (out.result != GHOSTTY_SUCCESS) return out;
	out.result = ghostty_snapshot_decoder_get(
		decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_SCREEN, &out.screen);
	if (out.result != GHOSTTY_SUCCESS) return out;

	switch (out.kind) {
	case GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_PAGE:
		out.result = ghostty_snapshot_decoder_get(
			decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_ROWS, &out.rows);
		if (out.result != GHOSTTY_SUCCESS) return out;
		out.result = ghostty_snapshot_decoder_get(
			decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_REMAINING, &out.remaining);
		return out;

	case GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_IMAGE:
		out.result = ghostty_snapshot_decoder_get(
			decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_ID, &out.image_id);
		if (out.result != GHOSTTY_SUCCESS) return out;
		out.result = ghostty_snapshot_decoder_get(
			decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_FRAME, &out.image_frame);
		if (out.result != GHOSTTY_SUCCESS) return out;
		out.result = ghostty_snapshot_decoder_get(
			decoder, GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_BYTES, &out.image_bytes);
		return out;

	default:
		// A newer library may add more progress kinds. Report the kind
		// and screen, and leave the kind-specific values zero.
		return out;
	}
}
*/
import "C"

import (
	"errors"
	"io"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

// SnapshotDecoderOption identifies a configurable snapshot decoder option.
// C: GhosttySnapshotDecoderOption
type SnapshotDecoderOption int

const (
	// SnapshotDecoderOptionMaxContinuationBytes is the largest non-ground VT
	// continuation the decoder will accept (size_t). When continuation
	// retention is enabled, this is also the tracking limit on the returned
	// terminal.
	SnapshotDecoderOptionMaxContinuationBytes SnapshotDecoderOption = C.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES

	// SnapshotDecoderOptionRetainContinuation controls whether the returned
	// terminal retains the decoded continuation (bool).
	SnapshotDecoderOptionRetainContinuation SnapshotDecoderOption = C.GHOSTTY_SNAPSHOT_DECODER_OPT_RETAIN_CONTINUATION

	// SnapshotDecoderOptionCompressHistory controls whether the decoder
	// compresses scrollback as it restores it (bool). See
	// [SnapshotDecoder.SetCompressHistory].
	SnapshotDecoderOptionCompressHistory SnapshotDecoderOption = C.GHOSTTY_SNAPSHOT_DECODER_OPT_COMPRESS_HISTORY

	// SnapshotDecoderOptionKittyImageStorageLimit is the maximum number of
	// bytes of Kitty graphics images kept for each restored screen
	// (uint64_t). See [SnapshotDecoder.SetKittyImageStorageLimit].
	SnapshotDecoderOptionKittyImageStorageLimit SnapshotDecoderOption = C.GHOSTTY_SNAPSHOT_DECODER_OPT_KITTY_IMAGE_STORAGE_LIMIT
)

// SnapshotDecoderProgressKind is the kind of record that the most recent
// successful call to [SnapshotDecoder.Next] decoded.
// C: GhosttySnapshotDecoderProgressKind
type SnapshotDecoderProgressKind int

const (
	// SnapshotDecoderProgressKindPage is a page of scrollback history,
	// along with the Kitty graphics placements on that page. The Rows and
	// Remaining fields of [SnapshotDecoderProgress] describe it.
	SnapshotDecoderProgressKindPage SnapshotDecoderProgressKind = C.GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_PAGE

	// SnapshotDecoderProgressKindImage is the pixel data of one Kitty
	// graphics image or animation frame. The ImageID, ImageFrame, and
	// ImageBytes fields of [SnapshotDecoderProgress] describe it.
	SnapshotDecoderProgressKindImage SnapshotDecoderProgressKind = C.GHOSTTY_SNAPSHOT_DECODER_PROGRESS_KIND_IMAGE
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

	// SnapshotDecoderDataProgressScreen is the screen that the most recently
	// decoded history page or image belongs to (GhosttyTerminalScreen).
	SnapshotDecoderDataProgressScreen SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_SCREEN

	// SnapshotDecoderDataProgressRows is the number of rows prepended by the
	// most recently decoded history page (size_t). It is only available
	// for [SnapshotDecoderProgressKindPage].
	SnapshotDecoderDataProgressRows SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_ROWS

	// SnapshotDecoderDataProgressRemaining is the number of page records
	// remaining in the current screen's history sequence (uint32_t). It is
	// only available for [SnapshotDecoderProgressKindPage].
	SnapshotDecoderDataProgressRemaining SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_REMAINING

	// SnapshotDecoderDataRetainContinuation reports whether continuation
	// tracking is retained on returned terminals (bool).
	SnapshotDecoderDataRetainContinuation SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_RETAIN_CONTINUATION

	// SnapshotDecoderDataCompressHistory reports whether the decoder
	// compresses scrollback as it restores it (bool). See
	// [SnapshotDecoder.SetCompressHistory].
	SnapshotDecoderDataCompressHistory SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_COMPRESS_HISTORY

	// SnapshotDecoderDataKittyImageStorageLimit is the maximum number of
	// bytes of Kitty graphics images kept for each restored screen
	// (uint64_t). See [SnapshotDecoder.SetKittyImageStorageLimit].
	SnapshotDecoderDataKittyImageStorageLimit SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGE_STORAGE_LIMIT

	// SnapshotDecoderDataProgressKind is the kind of record that the most
	// recent successful call to Next decoded
	// (GhosttySnapshotDecoderProgressKind). It is available whenever
	// [SnapshotDecoderDataProgressScreen] is.
	SnapshotDecoderDataProgressKind SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_KIND

	// SnapshotDecoderDataProgressImageID is the ID of the Kitty graphics
	// image whose pixel data the most recent call to Next decoded
	// (uint32_t). It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageID SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_ID

	// SnapshotDecoderDataProgressImageFrame is the frame number of that
	// pixel data (uint32_t). Frame 1 is the image itself, and frames 2 and
	// up are animation frames. It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageFrame SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_FRAME

	// SnapshotDecoderDataProgressImageBytes is the number of bytes of pixel
	// data that the most recent call to Next added to its image (size_t).
	// Zero means the data was read but discarded. It is only available for
	// [SnapshotDecoderProgressKindImage].
	SnapshotDecoderDataProgressImageBytes SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_PROGRESS_IMAGE_BYTES

	// SnapshotDecoderDataKittyImagesPrimary is the number of Kitty graphics
	// images restored on the primary screen (uint32_t). See
	// [SnapshotDecoder.KittyImageCount].
	SnapshotDecoderDataKittyImagesPrimary SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGES_PRIMARY

	// SnapshotDecoderDataKittyImagesAlternate is the number of Kitty
	// graphics images restored on the alternate screen (uint32_t). See
	// [SnapshotDecoder.KittyImageCount].
	SnapshotDecoderDataKittyImagesAlternate SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_IMAGES_ALTERNATE

	// SnapshotDecoderDataKittyPendingBytesPrimary is the number of bytes of
	// Kitty graphics pixel data still to come for the primary screen when
	// Ready returned (uint64_t). See
	// [SnapshotDecoder.KittyImagePendingBytes].
	SnapshotDecoderDataKittyPendingBytesPrimary SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_PENDING_BYTES_PRIMARY

	// SnapshotDecoderDataKittyPendingBytesAlternate is the number of bytes
	// of Kitty graphics pixel data still to come for the alternate screen
	// when Ready returned (uint64_t). See
	// [SnapshotDecoder.KittyImagePendingBytes].
	SnapshotDecoderDataKittyPendingBytesAlternate SnapshotDecoderData = C.GHOSTTY_SNAPSHOT_DECODER_DATA_KITTY_PENDING_BYTES_ALTERNATE
)

// SnapshotDecoder incrementally decodes and validates one terminal
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

	// sourceBytes and sourcePin keep the zero-copy input used by
	// NewSnapshotDecoderBytes alive, immovable, and therefore legal for C to
	// retain across calls. The pin is released at FINISH or decoder destruction.
	sourceBytes  []byte
	sourcePin    runtime.Pinner
	sourcePinned bool

	// sourcePtr is the C-owned input used by NewSnapshotDecoderBytesCopy. C
	// retains this borrowed pointer until FINISH or decoder destruction.
	sourcePtr unsafe.Pointer
	sourceLen uintptr
}

// snapshotReaderDecoder stores a decoder and its reader bridge in one
// allocation. The bridge remains separately addressable for use as C callback
// data.
type snapshotReaderDecoder struct {
	decoder SnapshotDecoder
	reader  ghosttyReaderBridge
}

// Snapshot encodes a complete terminal snapshot and returns a Go-owned copy.
// Calls must be serialized with every other operation on t. Callers taking
// repeated snapshots can use [Terminal.SnapshotBuf] with a reusable buffer to
// avoid the result allocation and copy.
// C: ghostty_snapshot_encode_alloc
func (t *Terminal) Snapshot() ([]byte, error) {
	result := C.ghostty_go_snapshot_encode_alloc(t.ptr)
	if err := resultError(result.result); err != nil {
		return nil, err
	}
	defer C.ghostty_free(nil, result.ptr, result.len)

	length, valid := ghosttySizeToInt(result.len)
	if !valid {
		return nil, &Error{Result: ResultLimitExceeded}
	}
	if length == 0 {
		return nil, nil
	}

	out := make([]byte, length)
	copy(out, unsafe.Slice((*byte)(unsafe.Pointer(result.ptr)), length))
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

	result := C.ghostty_go_snapshot_encode_buf(
		t.ptr,
		ptr,
		C.size_t(len(buf)),
	)
	length, valid := ghosttySizeToInt(result.written)
	if !valid {
		return 0, &Error{Result: ResultLimitExceeded}
	}
	return length, resultError(result.result)
}

// SnapshotWriteTo encodes a complete terminal snapshot directly to w. The
// writer is called synchronously and must not call methods on t. The returned
// count includes bytes accepted before an error.
// C: ghostty_snapshot_encode
func (t *Terminal) SnapshotWriteTo(w io.Writer) (int64, error) {
	bridge, err := newGhosttyWriter(w)
	if err != nil {
		return 0, err
	}
	defer bridge.close()

	result := C.ghostty_go_snapshot_encode(t.ptr, C.uintptr_t(bridge.handle))
	return bridge.written, resultErrorWithCallback(result, bridge.err)
}

// NewSnapshotDecoder creates a snapshot decoder backed by r. Reads are
// synchronous and occur only during Ready, Next, or Decode. A successful
// zero-byte read is permanent EOF; nonblocking readers must wait internally.
// C: ghostty_snapshot_decoder_new
func NewSnapshotDecoder(r io.Reader) (*SnapshotDecoder, error) {
	if r == nil {
		return nil, &Error{Result: ResultInvalidValue}
	}
	holder := &snapshotReaderDecoder{}
	holder.reader.reader = r
	holder.reader.handle = cgo.NewHandle(&holder.reader)

	result := C.ghostty_go_snapshot_decoder_new(C.uintptr_t(holder.reader.handle))
	if err := resultError(result.result); err != nil {
		holder.reader.close()
		return nil, err
	}

	holder.decoder.ptr = result.decoder
	holder.decoder.reader = &holder.reader
	return &holder.decoder, nil
}

// NewSnapshotDecoderBytes creates a zero-copy snapshot decoder over data. The
// decoder pins the byte slice because libghostty retains its source pointer
// across calls. The caller must not modify data until the decoder reaches
// FINISH through [SnapshotDecoder.Decode] or [SnapshotDecoder.Next], or until
// the decoder is closed.
// C: ghostty_snapshot_decoder_new_buf
func NewSnapshotDecoderBytes(data []byte) (*SnapshotDecoder, error) {
	d := &SnapshotDecoder{sourceBytes: data}
	var source *C.uint8_t
	if len(data) > 0 {
		d.sourcePin.Pin(&data[0])
		d.sourcePinned = true
		source = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}

	result := C.ghostty_go_snapshot_decoder_new_buf(source, C.size_t(len(data)))
	runtime.KeepAlive(data)
	if err := resultError(result.result); err != nil {
		d.releaseSource()
		return nil, err
	}

	d.ptr = result.decoder
	return d, nil
}

// NewSnapshotDecoderBytesCopy creates a snapshot decoder over a private copy
// of data. Unlike [NewSnapshotDecoderBytes], the caller may modify or release
// data immediately after this function returns. Prefer the zero-copy form for
// immutable snapshot input, especially for large histories.
// C: ghostty_snapshot_decoder_new_buf
func NewSnapshotDecoderBytesCopy(data []byte) (*SnapshotDecoder, error) {
	length := uintptr(len(data))
	var source unsafe.Pointer
	if length > 0 {
		source = Alloc(length)
		if source == nil {
			return nil, &Error{Result: ResultOutOfMemory}
		}
		copy(unsafe.Slice((*byte)(source), len(data)), data)
	}

	result := C.ghostty_go_snapshot_decoder_new_buf(
		(*C.uint8_t)(source),
		C.size_t(length),
	)
	if err := resultError(result.result); err != nil {
		Free(source, length)
		return nil, err
	}

	return &SnapshotDecoder{
		ptr:       result.decoder,
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
	d.releaseSource()
}

// releaseSource ends either form of source ownership after libghostty can no
// longer read the input. It is idempotent because FINISH releases the source
// before the decoder itself is closed.
func (d *SnapshotDecoder) releaseSource() {
	if d.sourcePinned {
		d.sourcePin.Unpin()
		d.sourcePinned = false
	}
	d.sourceBytes = nil
	if d.sourcePtr != nil {
		Free(d.sourcePtr, d.sourceLen)
		d.sourcePtr = nil
		d.sourceLen = 0
	}
}

// SetMaxContinuationBytes sets the largest non-ground VT continuation the
// decoder will accept. When [SnapshotDecoder.SetRetainContinuation] is true,
// this also becomes the continuation tracking limit on the returned terminal.
// It may only be called before decoding begins. Zero accepts only snapshots
// whose parser and UTF-8 decoder are at ground and leaves tracking disabled.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetMaxContinuationBytes(limit uint) error {
	v := C.size_t(limit)
	return resultError(C.ghostty_snapshot_decoder_set(
		d.ptr,
		C.GHOSTTY_SNAPSHOT_DECODER_OPT_MAX_CONTINUATION_BYTES,
		unsafe.Pointer(&v),
	))
}

// SetRetainContinuation controls whether the returned terminal retains the
// exact continuation decoded from the snapshot. Retention is disabled by
// default and requires a nonzero maximum continuation size to enable tracking
// on the returned terminal. It may only be called before decoding begins.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetRetainContinuation(retain bool) error {
	v := C.bool(retain)
	return resultError(C.ghostty_snapshot_decoder_set(
		d.ptr,
		C.GHOSTTY_SNAPSHOT_DECODER_OPT_RETAIN_CONTINUATION,
		unsafe.Pointer(&v),
	))
}

// SetCompressHistory controls whether the decoder compresses scrollback as
// it restores it. It is off by default and may only be called before
// decoding begins.
//
// Without this option, a restored terminal keeps all of its scrollback
// uncompressed, even if the terminal that produced the snapshot had
// compressed it. The scrollback stays that size until the application calls
// [Terminal.Compress]. For a terminal with a lot of scrollback, that can
// use many times more memory than the original terminal did.
//
// With this option, the decoder compresses each page of scrollback right
// after restoring it, so at most one uncompressed page is held at a time.
// The result is the same as decoding normally and then calling
// [Terminal.Compress] with [TerminalCompressionFull], but without the
// temporary jump in memory use. A page that is on screen when it is
// restored is left uncompressed. Compressed scrollback is decompressed
// automatically when it is needed, for example when the user scrolls or
// searches.
//
// This option only affects how the restored terminal stores its scrollback
// in memory. It works with any snapshot. On platforms that do not support
// scrollback compression, it is accepted and has no effect.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetCompressHistory(compress bool) error {
	v := C.bool(compress)
	return resultError(C.ghostty_snapshot_decoder_set(
		d.ptr,
		C.GHOSTTY_SNAPSHOT_DECODER_OPT_COMPRESS_HISTORY,
		unsafe.Pointer(&v),
	))
}

// SetKittyImageStorageLimit sets the maximum number of bytes of Kitty
// graphics images to keep for each restored screen. It must be called before
// decoding begins. The default is the limit that a new terminal uses, and
// zero turns off Kitty graphics on the restored terminal.
//
// Snapshots do not record the limit of the terminal that created them, so
// the decoder always uses this one. Images are restored oldest first. If an
// image does not fit, older images are removed to make room, just as they
// would be on a live terminal. To change the limit after decoding, call
// [Terminal.SetKittyImageStorageLimit] on the restored terminal.
// C: ghostty_snapshot_decoder_set
func (d *SnapshotDecoder) SetKittyImageStorageLimit(limit uint64) error {
	v := C.uint64_t(limit)
	return resultError(C.ghostty_snapshot_decoder_set(
		d.ptr,
		C.GHOSTTY_SNAPSHOT_DECODER_OPT_KITTY_IMAGE_STORAGE_LIMIT,
		unsafe.Pointer(&v),
	))
}

// Ready decodes and validates the renderable snapshot prefix. The
// returned terminal belongs to the caller and is ready to use right away.
// Older history and Kitty graphics pixel data can then be restored one
// record at a time with [SnapshotDecoder.Next].
//
// Ready restores Kitty graphics images without their pixel data. You can
// place and inspect these images right away, but [KittyGraphicsImage.Data]
// has nothing to return until Next delivers the pixels. When the pixels
// arrive, [KittyGraphics.Generation] changes, so a renderer that watches it
// draws the image on its next frame.
// C: ghostty_snapshot_decoder_ready
func (d *SnapshotDecoder) Ready() (*Terminal, error) {
	result := C.ghostty_go_snapshot_decoder_ready(d.ptr)
	if err := d.decodeError(result.result); err != nil {
		return nil, err
	}
	return terminalFromC(result.terminal, TerminalConfig{}), nil
}

// Next decodes and validates one record. A record is either a page of
// history, or the pixel data of one Kitty graphics image or animation frame.
// Pixel data can arrive at any point, including between history pages.
//
// Next returns true after it decodes a record. Call [SnapshotDecoder.Progress]
// to see what kind of record it was. Next returns false with a nil error once
// the whole snapshot has been decoded, and keeps returning false after that.
//
// You can render, resize, and write live input to the terminal between
// calls. If that makes a record impossible to apply, Next still reads and
// validates it, and Progress reports zero rows or zero bytes. For example,
// this happens when live input deletes the image that pixel data belongs to.
//
// If [SnapshotDecoder.SetCompressHistory] is on, Next compresses the page
// before returning, unless the page is on screen.
// C: ghostty_snapshot_decoder_next
func (d *SnapshotDecoder) Next() (bool, error) {
	result := C.ghostty_snapshot_decoder_next(d.ptr)
	switch result {
	case C.GHOSTTY_SUCCESS:
		return true, nil
	case C.GHOSTTY_NO_VALUE:
		d.releaseSource()
		return false, nil
	default:
		return false, d.decodeError(result)
	}
}

// Decode decodes and validates one complete snapshot in a single call.
// The returned terminal is caller-owned. Bytes following the FINISH record
// remain outside the snapshot and can be located with SourceOffset.
//
// The returned terminal has no effect callbacks. Modes are restored, so a
// snapshot taken during synchronized output yields a terminal that is
// already in a render hold; a [RenderHoldFunc] attached afterwards is not
// told about it. See [RenderHoldFunc].
// C: ghostty_snapshot_decoder_decode
func (d *SnapshotDecoder) Decode() (*Terminal, error) {
	result := C.ghostty_go_snapshot_decoder_decode(d.ptr)
	if err := d.decodeError(result.result); err != nil {
		return nil, err
	}
	d.releaseSource()
	return terminalFromC(result.terminal, TerminalConfig{}), nil
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

	// Copy the keys and output pointers into C memory. See get_multi.go.
	var args getMultiArgs
	cKeys, err := allocWithKeys[C.GhosttySnapshotDecoderData](&args, keys, values)
	if err != nil {
		return 0, err
	}
	defer args.free()
	var cWritten C.size_t
	err = resultError(C.ghostty_snapshot_decoder_get_multi(
		d.ptr,
		C.size_t(len(keys)),
		cKeys,
		args.values,
		&cWritten,
	))
	return int(cWritten), err
}

// MaxContinuationBytes returns the decoder's current continuation input limit.
// When [SnapshotDecoder.RetainContinuation] is true, this is also the tracking
// limit applied to the returned terminal.
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

// CompressHistory reports whether the decoder compresses scrollback as it
// restores it. See [SnapshotDecoder.SetCompressHistory].
func (d *SnapshotDecoder) CompressHistory() (bool, error) {
	var value C.bool
	if err := d.Get(
		SnapshotDecoderDataCompressHistory,
		unsafe.Pointer(&value),
	); err != nil {
		return false, err
	}
	return value != C.bool(false), nil
}

// RetainContinuation reports whether decoded continuation tracking will be
// retained on terminals returned by Ready or Decode.
func (d *SnapshotDecoder) RetainContinuation() (bool, error) {
	var value C.bool
	if err := d.Get(
		SnapshotDecoderDataRetainContinuation,
		unsafe.Pointer(&value),
	); err != nil {
		return false, err
	}
	return value != C.bool(false), nil
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
// returns [ErrNoValue] if the snapshot has no alternate screen.
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

// KittyImageStorageLimit returns the maximum number of bytes of Kitty
// graphics images the decoder keeps for each restored screen. See
// [SnapshotDecoder.SetKittyImageStorageLimit].
func (d *SnapshotDecoder) KittyImageStorageLimit() (uint64, error) {
	var value C.uint64_t
	if err := d.Get(
		SnapshotDecoderDataKittyImageStorageLimit,
		unsafe.Pointer(&value),
	); err != nil {
		return 0, err
	}
	return uint64(value), nil
}

// KittyImageCount returns the number of Kitty graphics images that Ready
// restored on screen. Their pixel data arrives later through
// [SnapshotDecoder.Next].
//
// It is available after Ready returns. It is zero if the library was built
// without Kitty graphics support. For [ScreenAlternate], it returns
// [ErrNoValue] if the snapshot has no alternate screen.
func (d *SnapshotDecoder) KittyImageCount(screen TerminalScreen) (uint32, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataKittyImagesPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataKittyImagesAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	var value C.uint32_t
	if err := d.Get(data, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// KittyImagePendingBytes returns the number of bytes of Kitty graphics pixel
// data that were still to come for screen when Ready returned.
//
// Use it to show progress while the pixel data arrives. After each call to
// Next, add [SnapshotDecoderProgress.ImageBytes] to a running total for the
// screen. Discarded pixel data counts as zero bytes, so the total can end up
// below this value.
//
// It is available after Ready returns. For [ScreenAlternate], it returns
// [ErrNoValue] if the snapshot has no alternate screen.
func (d *SnapshotDecoder) KittyImagePendingBytes(screen TerminalScreen) (uint64, error) {
	var data SnapshotDecoderData
	switch screen {
	case ScreenPrimary:
		data = SnapshotDecoderDataKittyPendingBytesPrimary
	case ScreenAlternate:
		data = SnapshotDecoderDataKittyPendingBytesAlternate
	default:
		return 0, &Error{Result: ResultInvalidValue}
	}

	var value C.uint64_t
	if err := d.Get(data, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint64(value), nil
}

// ProgressKind returns the kind of record that the most recent successful
// call to Next decoded.
func (d *SnapshotDecoder) ProgressKind() (SnapshotDecoderProgressKind, error) {
	var value C.GhosttySnapshotDecoderProgressKind
	if err := d.Get(SnapshotDecoderDataProgressKind, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return SnapshotDecoderProgressKind(value), nil
}

// ProgressImageID returns the ID of the Kitty graphics image whose pixel
// data the most recent call to Next decoded. It returns [ErrNoValue] if that
// record was not [SnapshotDecoderProgressKindImage].
func (d *SnapshotDecoder) ProgressImageID() (uint32, error) {
	var value C.uint32_t
	if err := d.Get(SnapshotDecoderDataProgressImageID, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// ProgressImageFrame returns the frame number of the pixel data that the
// most recent call to Next decoded. Frame 1 is the image itself, and frames
// 2 and up are animation frames. It returns [ErrNoValue] if that record was
// not [SnapshotDecoderProgressKindImage].
func (d *SnapshotDecoder) ProgressImageFrame() (uint32, error) {
	var value C.uint32_t
	if err := d.Get(SnapshotDecoderDataProgressImageFrame, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// ProgressImageBytes returns the number of bytes of pixel data that the most
// recent call to Next added to its image. It returns [ErrNoValue] if that
// record was not [SnapshotDecoderProgressKindImage].
//
// Zero means the pixel data was read and validated but then discarded. That
// happens when live input deleted the image or replaced it with a new image
// that has the same ID, when the image did not fit the storage limit, or
// when its screen no longer exists.
func (d *SnapshotDecoder) ProgressImageBytes() (uint, error) {
	var value C.size_t
	if err := d.Get(SnapshotDecoderDataProgressImageBytes, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint(value), nil
}

// ProgressScreen returns the screen associated with the most recently
// decoded history page or image belongs to.
func (d *SnapshotDecoder) ProgressScreen() (TerminalScreen, error) {
	var value C.GhosttyTerminalScreen
	if err := d.Get(SnapshotDecoderDataProgressScreen, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return TerminalScreen(value), nil
}

// ProgressRows returns the number of rows prepended by the most recently
// decoded history page. Zero means the page could not be applied safely.
// It returns [ErrNoValue] if that record was not
// [SnapshotDecoderProgressKindPage].
func (d *SnapshotDecoder) ProgressRows() (uint, error) {
	var value C.size_t
	if err := d.Get(SnapshotDecoderDataProgressRows, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint(value), nil
}

// ProgressRemaining returns the number of page records remaining in the same
// screen's history sequence. It returns [ErrNoValue] if that record was not
// [SnapshotDecoderProgressKindPage].
func (d *SnapshotDecoder) ProgressRemaining() (uint32, error) {
	var value C.uint32_t
	if err := d.Get(SnapshotDecoderDataProgressRemaining, unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// SnapshotDecoderProgress describes the record that the most recent
// successful call to [SnapshotDecoder.Next] decoded. Kind says which fields
// apply. Fields that don't apply are zero.
type SnapshotDecoderProgress struct {
	// Kind is the kind of record: a history page or image pixel data.
	Kind SnapshotDecoderProgressKind

	// Screen is the screen that the record belongs to.
	Screen TerminalScreen

	// Rows is the number of rows added to the top of the scrollback. Zero
	// means the page could not be applied. It is set for
	// [SnapshotDecoderProgressKindPage].
	Rows uint

	// Remaining is the number of history pages still to come for this
	// screen. It is set for [SnapshotDecoderProgressKindPage].
	Remaining uint32

	// ImageID is the ID of the Kitty graphics image that the pixel data
	// belongs to. It is set for [SnapshotDecoderProgressKindImage].
	ImageID uint32

	// ImageFrame is the frame number of the pixel data. Frame 1 is the
	// image itself, and frames 2 and up are animation frames. It is set for
	// [SnapshotDecoderProgressKindImage].
	ImageFrame uint32

	// ImageBytes is the number of bytes of pixel data added to the image.
	// Zero means the data was discarded. See
	// [SnapshotDecoder.ProgressImageBytes] for when that happens. It is set
	// for [SnapshotDecoderProgressKindImage].
	ImageBytes uint
}

// Progress returns every progress value for the record that the most recent
// successful call to Next decoded. It reads them all in a single call and
// does not allocate, so prefer it to the individual Progress methods.
func (d *SnapshotDecoder) Progress() (SnapshotDecoderProgress, error) {
	out := C.ghostty_go_snapshot_decoder_progress(d.ptr)
	if err := resultError(out.result); err != nil {
		return SnapshotDecoderProgress{}, err
	}
	return SnapshotDecoderProgress{
		Kind:       SnapshotDecoderProgressKind(out.kind),
		Screen:     TerminalScreen(out.screen),
		Rows:       uint(out.rows),
		Remaining:  uint32(out.remaining),
		ImageID:    uint32(out.image_id),
		ImageFrame: uint32(out.image_frame),
		ImageBytes: uint(out.image_bytes),
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
