#ifndef GHOSTTY_GO_IO_H
#define GHOSTTY_GO_IO_H

#include <ghostty/vt.h>
#include <stdint.h>

// The exported Go callbacks accept userdata as an integer so a cgo.Handle is
// never represented as a pointer in a Go stack frame. The native callback
// signatures still use void* as required by libghostty; these trampolines
// perform the conversion entirely on the C side.
extern bool goGhosttyReaderTrampoline(
    uintptr_t userdata,
    uint8_t* buffer,
    size_t capacity,
    size_t* out_read);

extern bool goGhosttyWriterTrampoline(
    uintptr_t userdata,
    uint8_t* data,
    size_t len);

extern bool goGhosttyMIMEReaderTrampoline(
    uintptr_t userdata,
    GhosttyString mime,
    GhosttyWriter writer);

static inline bool ghostty_go_reader_trampoline(
    void* userdata,
    uint8_t* buffer,
    size_t capacity,
    size_t* out_read
) {
    return goGhosttyReaderTrampoline(
        (uintptr_t)userdata,
        buffer,
        capacity,
        out_read);
}

static inline bool ghostty_go_writer_trampoline(
    void* userdata,
    const uint8_t* data,
    size_t len
) {
    // cgo cannot export a const-qualified pointer parameter. The Go callback
    // treats data as read-only, so removing const here is ABI-safe.
    return goGhosttyWriterTrampoline(
        (uintptr_t)userdata,
        (uint8_t*)data,
        len);
}

static inline bool ghostty_go_mime_reader_trampoline(
    void* userdata,
    GhosttyString mime,
    GhosttyWriter writer
) {
    return goGhosttyMIMEReaderTrampoline(
        (uintptr_t)userdata,
        mime,
        writer);
}

// This factory must only be called by C helpers. Returning the descriptor to
// Go would place the integer handle in a pointer-typed field, allowing Go's
// stack copier or garbage collector to mistake it for a pointer.
static inline GhosttyWriter ghostty_go_writer(uintptr_t userdata) {
    GhosttyWriter writer = {
        .write = ghostty_go_writer_trampoline,
        .userdata = (void*)userdata,
    };
    return writer;
}

#endif
