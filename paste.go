package libghostty

// Paste bindings for terminal-aware pasting plus terminal-free validation and
// encoding. Wraps the C APIs from paste.h.

/*
#include <ghostty/vt.h>
#include "go_io.h"

static inline GhosttyResult ghostty_go_terminal_paste(
	GhosttyTerminal terminal,
	const GhosttyPaste* request,
	uintptr_t userdata,
	bool* out_written
) {
	GhosttyPaste copy = *request;
	if (userdata != 0) {
		copy.reader = (GhosttyMimeReader){
			.read = ghostty_go_mime_reader_trampoline,
			.userdata = (void*)userdata,
		};
	}
	return ghostty_terminal_paste(terminal, &copy, out_written);
}
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// PasteSource identifies why paste content is being inserted.
//
// C: GhosttyPasteSource
type PasteSource int

const (
	// PasteSourceClipboard identifies a user-initiated clipboard paste. When
	// Kitty paste events are enabled and a clipboard-read callback is
	// installed, the terminal sends an event instead of writing text directly.
	PasteSourceClipboard PasteSource = C.GHOSTTY_PASTE_SOURCE_CLIPBOARD

	// PasteSourceText identifies text inserted by another mechanism such as an
	// IME commit, drag and drop, or scripted input. It is always written as text.
	PasteSourceText PasteSource = C.GHOSTTY_PASTE_SOURCE_TEXT
)

// Paste describes content to paste into a terminal. MIMEs lists every
// available representation in preferred order. Reader is called at most once
// for the first text representation selected by libghostty and is not called
// when a Kitty paste event is emitted.
//
// C: GhosttyPaste
type Paste struct {
	// Location identifies the clipboard that supplied the content.
	Location ClipboardLocation

	// Source identifies why the paste happened.
	Source PasteSource

	// MIMEs lists the available representation types in preferred order.
	MIMEs []string

	// Reader streams a requested representation. It is required when MIMEs is
	// non-empty.
	Reader MIMEReaderFn

	// AllowUnsafe permits text that could inject commands. Call Paste with this
	// false first, confirm with the user on ResultRejected, then retry with it
	// true.
	AllowUnsafe bool
}

// Paste applies the terminal's current paste modes and writes either encoded
// text or a Kitty clipboard paste event through the terminal's write-pty
// callback. The returned boolean reports whether anything was written.
//
// Paste refuses text that could run commands without the user meaning to.
// Normally that is text with a newline or carriage return, because the shell
// treats either one like the user pressing Enter. When the program has turned
// on bracketed paste mode, line breaks are safe, and only the sequence that
// ends a bracketed paste is refused. Refused text returns an error matching
// [ResultRejected] and nothing is written. Set Paste.AllowUnsafe to paste it
// anyway.
//
// If the reader fails, the returned error matches both [ResultIOError] and
// the error the reader returned.
//
// C: ghostty_terminal_paste
func (t *Terminal) Paste(paste Paste) (bool, error) {
	mimeValues := make([][]byte, len(paste.MIMEs))
	for i, mime := range paste.MIMEs {
		mimeValues[i] = []byte(mime)
	}
	mimes, err := newCGhosttyStringArray(mimeValues)
	if err != nil {
		return false, err
	}
	defer mimes.close()

	bridge := &ghosttyMIMEReaderBridge{}
	if len(paste.MIMEs) > 0 {
		if paste.Reader == nil {
			return false, &Error{Result: ResultInvalidValue}
		}
		bridge.reader = paste.Reader
		bridge.handle = cgo.NewHandle(bridge)
		defer bridge.close()
	}

	request := C.GhosttyPaste{
		size:         C.size_t(C.sizeof_GhosttyPaste),
		location:     C.GhosttyClipboardLocation(paste.Location),
		source:       C.GhosttyPasteSource(paste.Source),
		mimes:        mimes.ptr,
		mimes_len:    C.size_t(len(paste.MIMEs)),
		allow_unsafe: C.bool(paste.AllowUnsafe),
	}
	var userdata C.uintptr_t
	if bridge.handle != 0 {
		userdata = C.uintptr_t(bridge.handle)
	}
	var written C.bool
	result := C.ghostty_go_terminal_paste(t.ptr, &request, userdata, &written)
	if err := resultErrorWithCallback(result, bridge.err); err != nil {
		return false, err
	}
	return bool(written), nil
}

// PasteIsSafe reports whether data is safe to paste into a terminal.
//
// Data is unsafe if it contains a newline or carriage return, because the
// shell treats either one like the user pressing Enter and may run a
// command. Data is also unsafe if it contains the sequence that ends a
// bracketed paste (ESC [201~), because it lets the pasted text end the paste
// early and run whatever follows as a command. Empty data is safe.
//
// PasteIsSafe is stricter than [Terminal.Paste]. It doesn't know whether the
// terminal is in bracketed paste mode, so it treats line breaks as unsafe
// even when Paste would allow them. Use it to warn the user about every
// paste that contains a line break, whatever mode the terminal is in.
func PasteIsSafe(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	return bool(C.ghostty_paste_is_safe(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	))
}

// PasteEncode prepares data for writing to a terminal pty.
//
// The encoder applies Ghostty's terminal-input paste rules: unsafe control
// bytes are replaced with spaces, bracketed paste markers are added when
// bracketed is true, and newlines become carriage returns when bracketed is
// false. The input slice is copied before calling into Ghostty because the C
// encoder mutates its data buffer in place.
func PasteEncode(data []byte, bracketed bool) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// The current encoder only grows output by adding bracketed-paste markers,
	// but intentionally treat buffer sizing as an implementation detail of the C
	// API. Starting with modest slack keeps the common path allocation-light
	// while still honoring GHOSTTY_OUT_OF_SPACE below if the encoder changes.
	in := append([]byte(nil), data...)
	buf := make([]byte, len(data)+32)
	out, required, err := pasteEncodeInto(in, bracketed, buf)
	if err == nil {
		return out, nil
	}

	ge, ok := err.(*Error)
	if !ok || ge.Result != ResultOutOfSpace || required <= 0 {
		return nil, err
	}

	in = append([]byte(nil), data...)
	buf = make([]byte, required)
	out, _, err = pasteEncodeInto(in, bracketed, buf)
	return out, err
}

func pasteEncodeInto(
	data []byte,
	bracketed bool,
	buf []byte,
) ([]byte, int, error) {
	var written C.size_t
	result := C.ghostty_paste_encode(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		C.bool(bracketed),
		(*C.char)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&written,
	)
	if result == C.GHOSTTY_SUCCESS {
		return buf[:int(written)], int(written), nil
	}

	return nil, int(written), &Error{Result: Result(result)}
}
