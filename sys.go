package libghostty

// System-level configuration wrapping ghostty_sys_set().
// These are process-global settings that must be configured at startup.

/*
#include <ghostty/vt.h>
#include <ghostty/vt/sys.h>

// Forward declaration for the Go log trampoline so we can take its
// address on the C side. Uses compatible types (no const, int for enum)
// to match what cgo generates for the //export function.
extern void goSysLogTrampoline(
	void* userdata,
	int level,
	uint8_t* scope,
	size_t scope_len,
	uint8_t* message,
	size_t message_len);

// Forward declaration for the Go decode-PNG trampoline.
// Uses compatible types (no const) to match what cgo generates for
// the //export function.
extern _Bool goSysDecodePngTrampoline(
	void* userdata,
	GhosttyAllocator* allocator,
	uint8_t* data,
	size_t data_len,
	GhosttySysImage* out);

// Helper to install the Go log trampoline via ghostty_sys_set.
// We need this because cgo cannot take the address of a Go-exported
// function directly as a C function pointer.
static inline GhosttyResult sys_set_log_go(void) {
	return ghostty_sys_set(GHOSTTY_SYS_OPT_LOG, (const void*)goSysLogTrampoline);
}

// Helper to install the built-in stderr log callback.
static inline GhosttyResult sys_set_log_stderr(void) {
	return ghostty_sys_set(GHOSTTY_SYS_OPT_LOG, (const void*)ghostty_sys_log_stderr);
}

// Helper to clear the log callback.
static inline GhosttyResult sys_clear_log(void) {
	return ghostty_sys_set(GHOSTTY_SYS_OPT_LOG, NULL);
}

// Helper to install the Go decode-PNG trampoline via ghostty_sys_set.
static inline GhosttyResult sys_set_decode_png_go(void) {
	return ghostty_sys_set(GHOSTTY_SYS_OPT_DECODE_PNG, (const void*)goSysDecodePngTrampoline);
}

// Helper to clear the decode-PNG callback.
static inline GhosttyResult sys_clear_decode_png(void) {
	return ghostty_sys_set(GHOSTTY_SYS_OPT_DECODE_PNG, NULL);
}
*/
import "C"

import "unsafe"

// SysImage holds the result of decoding an image (e.g. PNG) into raw
// RGBA pixel data. Returned by the user-supplied decode callback.
// C: GhosttySysImage
type SysImage struct {
	// Width of the decoded image in pixels.
	Width uint32

	// Height of the decoded image in pixels.
	Height uint32

	// Data is the decoded RGBA pixel data (4 bytes per pixel).
	Data []byte
}

// SysDecodePngFn is the Go callback type for PNG decoding. It receives
// raw PNG data and must return a decoded SysImage. The returned pixel
// data will be copied into library-managed memory; the caller does not
// need to keep the slice alive after returning.
//
// Return a non-nil error to indicate decode failure.
// C: GhosttySysDecodePngFn
type SysDecodePngFn func(data []byte) (*SysImage, error)

// sysDecodePngFn is the currently installed Go decode-PNG callback.
var sysDecodePngFn SysDecodePngFn

// SysSetDecodePng installs a Go callback that decodes PNG image data
// into RGBA pixels. This enables PNG support in the Kitty Graphics
// Protocol. Pass nil to clear the callback and disable PNG decoding.
//
// This function is not safe for concurrent use. Callers must ensure
// that decode configuration is not modified while terminals may
// process image data (e.g. configure at startup before creating
// terminals).
func SysSetDecodePng(fn SysDecodePngFn) error {
	sysDecodePngFn = fn
	if fn == nil {
		return resultError(C.sys_clear_decode_png())
	}
	return resultError(C.sys_set_decode_png_go())
}

// SysLogLevel represents the severity level of a log message from the
// library. Maps directly to the C enum values.
// C: GhosttySysLogLevel
type SysLogLevel int

const (
	// SysLogLevelError is the error log level.
	SysLogLevelError SysLogLevel = C.GHOSTTY_SYS_LOG_LEVEL_ERROR

	// SysLogLevelWarning is the warning log level.
	SysLogLevelWarning SysLogLevel = C.GHOSTTY_SYS_LOG_LEVEL_WARNING

	// SysLogLevelInfo is the info log level.
	SysLogLevelInfo SysLogLevel = C.GHOSTTY_SYS_LOG_LEVEL_INFO

	// SysLogLevelDebug is the debug log level.
	SysLogLevelDebug SysLogLevel = C.GHOSTTY_SYS_LOG_LEVEL_DEBUG
)

// String returns a human-readable name for the log level.
func (l SysLogLevel) String() string {
	switch l {
	case SysLogLevelError:
		return "error"
	case SysLogLevelWarning:
		return "warning"
	case SysLogLevelInfo:
		return "info"
	case SysLogLevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// SysLogFn is the Go callback type for log messages from the library.
// The scope identifies the subsystem (e.g. "osc", "kitty"); it is
// empty for unscoped (default) log messages. The message and scope
// are only valid for the duration of the call.
// C: GhosttySysLogFn
type SysLogFn func(level SysLogLevel, scope string, message string)

// sysLogFn is the currently installed Go log callback.
var sysLogFn SysLogFn

// SysSetLog installs a Go callback that receives internal library log
// messages. Pass nil to clear the callback and discard log messages.
//
// Which log levels are emitted depends on the build mode of the
// library. Debug builds emit all levels; release builds emit info
// and above.
//
// This function is not safe for concurrent use. Callers must ensure
// that log configuration is not modified while log messages may be
// delivered (e.g. configure at startup before creating terminals).
func SysSetLog(fn SysLogFn) error {
	sysLogFn = fn
	if fn == nil {
		return resultError(C.sys_clear_log())
	}
	return resultError(C.sys_set_log_go())
}

// SysSetLogStderr installs the built-in stderr log callback provided
// by libghostty. Each message is formatted as "[level](scope): message\n"
// and written to stderr in a thread-safe manner.
//
// This function is not safe for concurrent use. See [SysSetLog].
func SysSetLogStderr() error {
	sysLogFn = nil
	return resultError(C.sys_set_log_stderr())
}

//export goSysLogTrampoline
func goSysLogTrampoline(
	_ unsafe.Pointer,
	level C.int,
	scopePtr *C.uint8_t,
	scopeLen C.size_t,
	messagePtr *C.uint8_t,
	messageLen C.size_t,
) {
	fn := sysLogFn
	if fn == nil {
		return
	}

	var scope string
	if scopeLen > 0 {
		scope = C.GoStringN((*C.char)(unsafe.Pointer(scopePtr)), C.int(scopeLen))
	}

	var message string
	if messageLen > 0 {
		message = C.GoStringN((*C.char)(unsafe.Pointer(messagePtr)), C.int(messageLen))
	}

	fn(SysLogLevel(level), scope, message)
}

//export goSysDecodePngTrampoline
func goSysDecodePngTrampoline(
	_ unsafe.Pointer,
	allocator *C.GhosttyAllocator,
	dataPtr *C.uint8_t,
	dataLen C.size_t,
	out *C.GhosttySysImage,
) C.bool {
	fn := sysDecodePngFn
	if fn == nil {
		return false
	}

	// Build a Go slice over the input PNG data without copying.
	data := unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), int(dataLen))

	img, err := fn(data)
	if err != nil || img == nil {
		return false
	}

	// Allocate output pixel buffer through the library's allocator so
	// the library can free it later.
	pixelLen := C.size_t(len(img.Data))
	buf := C.ghostty_alloc(allocator, pixelLen)
	if buf == nil {
		return false
	}

	// Copy decoded pixels into the library-owned buffer.
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(pixelLen)), img.Data)

	out.width = C.uint32_t(img.Width)
	out.height = C.uint32_t(img.Height)
	out.data = buf
	out.data_len = pixelLen

	return true
}
