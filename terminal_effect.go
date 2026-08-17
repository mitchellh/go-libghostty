package libghostty

// C trampolines for terminal effects.
//
// Each exported Go function here is passed to the C side as a function
// pointer. The C library calls it with a userdata void*, which is a
// cgo.Handle pointing back to the owning *Terminal. The trampoline
// recovers the Terminal and dispatches to the user-supplied Go effect handler.

/*
#include <ghostty/vt.h>

// Forward declarations for the Go trampolines so we can take their
// addresses on the C side.
extern void goWritePtyTrampoline(GhosttyTerminal, void*, uint8_t*, size_t);
extern void goBellTrampoline(GhosttyTerminal, void*);
extern GhosttyClipboardWriteResult goClipboardWriteTrampoline(GhosttyTerminal, void*, GhosttyClipboardWrite*);
extern void goDesktopNotificationTrampoline(GhosttyTerminal, void*, GhosttyTerminalDesktopNotification*);
extern void goTitleChangedTrampoline(GhosttyTerminal, void*);
extern void goPwdChangedTrampoline(GhosttyTerminal, void*);
extern void goProgressReportTrampoline(GhosttyTerminal, void*, GhosttyTerminalProgressReport*);
extern GhosttyString goEnquiryTrampoline(GhosttyTerminal, void*);
extern GhosttyString goXtversionTrampoline(GhosttyTerminal, void*);
extern bool goSizeTrampoline(GhosttyTerminal, void*, GhosttySizeReportSize*);
extern bool goColorSchemeTrampoline(GhosttyTerminal, void*, GhosttyColorScheme*);
extern bool goDeviceAttributesTrampoline(GhosttyTerminal, void*, GhosttyDeviceAttributes*);
extern void goUnknownSequenceTrampoline(GhosttyTerminal, void*, GhosttyTerminalUnknownSequence*);

// Helpers to set each effect via ghostty_terminal_set.
// We need these because cgo cannot take the address of a Go-exported
// function directly as a C function pointer.
static inline GhosttyResult set_write_pty(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_WRITE_PTY, (const void*)goWritePtyTrampoline);
}
static inline GhosttyResult set_bell(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_BELL, (const void*)goBellTrampoline);
}
static inline GhosttyResult set_clipboard_write(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_CLIPBOARD_WRITE, (const void*)goClipboardWriteTrampoline);
}
static inline GhosttyResult set_desktop_notification(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_DESKTOP_NOTIFICATION, (const void*)goDesktopNotificationTrampoline);
}
static inline GhosttyResult set_title_changed(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_TITLE_CHANGED, (const void*)goTitleChangedTrampoline);
}
static inline GhosttyResult set_pwd_changed(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_PWD_CHANGED, (const void*)goPwdChangedTrampoline);
}
static inline GhosttyResult set_progress_report(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_PROGRESS_REPORT, (const void*)goProgressReportTrampoline);
}
static inline GhosttyResult set_enquiry(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_ENQUIRY, (const void*)goEnquiryTrampoline);
}
static inline GhosttyResult set_xtversion(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_XTVERSION, (const void*)goXtversionTrampoline);
}
static inline GhosttyResult set_size(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_SIZE, (const void*)goSizeTrampoline);
}
static inline GhosttyResult set_color_scheme(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_COLOR_SCHEME, (const void*)goColorSchemeTrampoline);
}
static inline GhosttyResult set_device_attributes(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_DEVICE_ATTRIBUTES, (const void*)goDeviceAttributesTrampoline);
}
static inline GhosttyResult set_unknown_sequence(GhosttyTerminal t) {
	return ghostty_terminal_set(t, GHOSTTY_TERMINAL_OPT_UNKNOWN_SEQUENCE, (const void*)goUnknownSequenceTrampoline);
}

// Return the APC member of the tagged union without making Go depend on
// cgo's platform-specific representation of C unions.
static inline const GhosttyTerminalUnknownStringSequence* unknown_sequence_apc(
		const GhosttyTerminalUnknownSequence* sequence) {
	return &sequence->value.apc;
}

// Helper to clear an effect by setting it to NULL.
static inline GhosttyResult clear_effect(GhosttyTerminal t, GhosttyTerminalOption opt) {
	return ghostty_terminal_set(t, opt, NULL);
}
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// syncEffects registers or clears each C effect based on whether
// the corresponding Go effect handler is set.
func (t *Terminal) syncEffects() {
	// Install userdata before registering the first callback. Terminals without
	// callbacks do not need a handle.
	if t.handle == 0 && t.hasEffects() {
		t.handle = cgo.NewHandle(t)
		C.ghostty_terminal_set(
			t.ptr,
			C.GHOSTTY_TERMINAL_OPT_USERDATA,
			handleToPointer(t.handle),
		)
	}

	if t.onWritePty != nil {
		C.set_write_pty(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_WRITE_PTY)
	}
	if t.onBell != nil {
		C.set_bell(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_BELL)
	}
	if t.onClipboardWrite != nil {
		C.set_clipboard_write(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_CLIPBOARD_WRITE)
	}
	if t.onDesktopNotification != nil {
		C.set_desktop_notification(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_DESKTOP_NOTIFICATION)
	}
	if t.onTitleChanged != nil {
		C.set_title_changed(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_TITLE_CHANGED)
	}
	if t.onPwdChanged != nil {
		C.set_pwd_changed(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_PWD_CHANGED)
	}
	if t.onProgressReport != nil {
		C.set_progress_report(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_PROGRESS_REPORT)
	}
	if t.onEnquiry != nil {
		C.set_enquiry(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_ENQUIRY)
	}
	if t.onXtversion != nil {
		C.set_xtversion(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_XTVERSION)
	}
	if t.onSize != nil {
		C.set_size(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_SIZE)
	}
	if t.onColorScheme != nil {
		C.set_color_scheme(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_COLOR_SCHEME)
	}
	if t.onDeviceAttributes != nil {
		C.set_device_attributes(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_DEVICE_ATTRIBUTES)
	}
	if t.onUnknownSequence != nil {
		C.set_unknown_sequence(t.ptr)
	} else {
		C.clear_effect(t.ptr, C.GHOSTTY_TERMINAL_OPT_UNKNOWN_SEQUENCE)
	}
}

// hasEffects reports whether any native effect trampoline needs to recover
// this Terminal through userdata.
func (t *Terminal) hasEffects() bool {
	return t.onWritePty != nil ||
		t.onBell != nil ||
		t.onClipboardWrite != nil ||
		t.onDesktopNotification != nil ||
		t.onTitleChanged != nil ||
		t.onPwdChanged != nil ||
		t.onProgressReport != nil ||
		t.onEnquiry != nil ||
		t.onXtversion != nil ||
		t.onSize != nil ||
		t.onColorScheme != nil ||
		t.onDeviceAttributes != nil ||
		t.onUnknownSequence != nil
}

// terminalFromUserdata recovers a *Terminal from the C userdata pointer.
func terminalFromUserdata(userdata unsafe.Pointer) *Terminal {
	return cgo.Handle(userdata).Value().(*Terminal)
}

//export goWritePtyTrampoline
func goWritePtyTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer, data *C.uint8_t, length C.size_t) {
	t := terminalFromUserdata(userdata)
	if t.onWritePty != nil {
		t.onWritePty(t, C.GoBytes(unsafe.Pointer(data), C.int(length)))
	}
}

//export goBellTrampoline
func goBellTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer) {
	t := terminalFromUserdata(userdata)
	if t.onBell != nil {
		t.onBell(t)
	}
}

//export goClipboardWriteTrampoline
func goClipboardWriteTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer, write *C.GhosttyClipboardWrite) C.GhosttyClipboardWriteResult {
	t := terminalFromUserdata(userdata)
	if t.onClipboardWrite == nil {
		return C.GHOSTTY_CLIPBOARD_WRITE_RESULT_UNSUPPORTED
	}

	// GhosttyClipboardWrite is a sized struct so newer libghostty versions
	// can extend it without breaking existing callbacks. Only read the fields
	// this binding knows about when the descriptor contains the full current
	// layout.
	if write == nil || write.size < C.size_t(C.sizeof_GhosttyClipboardWrite) {
		return C.GHOSTTY_CLIPBOARD_WRITE_RESULT_INVALID_DATA
	}

	count, ok := ghosttySizeToInt(write.contents_len)
	if !ok || (count > 0 && write.contents == nil) {
		return C.GHOSTTY_CLIPBOARD_WRITE_RESULT_INVALID_DATA
	}

	contents := make([]ClipboardContent, count)
	if count > 0 {
		cContents := unsafe.Slice(write.contents, count)
		for i, content := range cContents {
			mime, valid := copyGhosttyString(content.mime)
			if !valid {
				return C.GHOSTTY_CLIPBOARD_WRITE_RESULT_INVALID_DATA
			}
			data, valid := copyGhosttyString(content.data)
			if !valid {
				return C.GHOSTTY_CLIPBOARD_WRITE_RESULT_INVALID_DATA
			}

			contents[i] = ClipboardContent{
				MIME: string(mime),
				Data: data,
			}
		}
	}

	result := t.onClipboardWrite(t, ClipboardWrite{
		Location: ClipboardLocation(write.location),
		Contents: contents,
	})
	return C.GhosttyClipboardWriteResult(result)
}

// ghosttySizeToInt converts a C size to a Go slice length without allowing
// an overflowing conversion to produce an invalid unsafe.Slice length.
func ghosttySizeToInt(size C.size_t) (int, bool) {
	if uint64(size) > uint64(^uint(0)>>1) {
		return 0, false
	}
	return int(size), true
}

// copyGhosttyString copies a borrowed, binary-safe GhosttyString into Go
// memory. For zero-length strings the pointer is intentionally ignored
// because libghostty does not require it to be valid.
func copyGhosttyString(value C.GhosttyString) ([]byte, bool) {
	length, ok := ghosttySizeToInt(value.len)
	if !ok {
		return nil, false
	}
	if length == 0 {
		return []byte{}, true
	}
	if value.ptr == nil {
		return nil, false
	}

	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(value.ptr)), length)...), true
}

//export goDesktopNotificationTrampoline
func goDesktopNotificationTrampoline(
	_ C.GhosttyTerminal,
	userdata unsafe.Pointer,
	notification *C.GhosttyTerminalDesktopNotification,
) {
	t := terminalFromUserdata(userdata)
	if t.onDesktopNotification == nil {
		return
	}

	// GhosttyTerminalDesktopNotification is a sized struct so newer
	// libghostty versions can extend it without breaking existing callbacks.
	// Ignore descriptors that do not contain the full layout this binding
	// knows how to read.
	if notification == nil ||
		notification.size < C.size_t(C.sizeof_GhosttyTerminalDesktopNotification) {
		return
	}

	title, ok := copyGhosttyString(notification.title)
	if !ok {
		return
	}
	body, ok := copyGhosttyString(notification.body)
	if !ok {
		return
	}

	t.onDesktopNotification(t, TerminalDesktopNotification{
		Title: string(title),
		Body:  string(body),
	})
}

//export goTitleChangedTrampoline
func goTitleChangedTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer) {
	t := terminalFromUserdata(userdata)
	if t.onTitleChanged != nil {
		t.onTitleChanged(t)
	}
}

//export goPwdChangedTrampoline
func goPwdChangedTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer) {
	t := terminalFromUserdata(userdata)
	if t.onPwdChanged != nil {
		t.onPwdChanged(t)
	}
}

//export goProgressReportTrampoline
func goProgressReportTrampoline(
	_ C.GhosttyTerminal,
	userdata unsafe.Pointer,
	report *C.GhosttyTerminalProgressReport,
) {
	t := terminalFromUserdata(userdata)
	if t.onProgressReport == nil {
		return
	}

	// GhosttyTerminalProgressReport is also a sized struct. Only read the
	// fields when the caller supplied the complete layout known here.
	if report == nil ||
		report.size < C.size_t(C.sizeof_GhosttyTerminalProgressReport) {
		return
	}

	t.onProgressReport(t, TerminalProgressReport{
		State:    TerminalProgressState(report.state),
		Progress: int8(report.progress),
	})
}

//export goUnknownSequenceTrampoline
func goUnknownSequenceTrampoline(
	_ C.GhosttyTerminal,
	userdata unsafe.Pointer,
	sequence *C.GhosttyTerminalUnknownSequence,
) {
	t := terminalFromUserdata(userdata)
	if t.onUnknownSequence == nil || sequence == nil {
		return
	}

	value := TerminalUnknownSequence{
		Tag: TerminalUnknownSequenceTag(sequence.tag),
	}
	if value.Tag == TerminalUnknownSequenceAPC {
		apc := C.unknown_sequence_apc(sequence)
		content, ok := copyGhosttyString(apc.content)
		if !ok {
			return
		}
		value.APC = TerminalUnknownStringSequence{
			Truncated: bool(apc.truncated),
			Content:   content,
		}
	}

	t.onUnknownSequence(t, value)
}

//export goEnquiryTrampoline
func goEnquiryTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer) C.GhosttyString {
	t := terminalFromUserdata(userdata)
	if t.onEnquiry == nil {
		return C.GhosttyString{}
	}
	return t.effectString(t.onEnquiry(t))
}

//export goXtversionTrampoline
func goXtversionTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer) C.GhosttyString {
	t := terminalFromUserdata(userdata)
	if t.onXtversion == nil {
		return C.GhosttyString{}
	}
	return t.effectString([]byte(t.onXtversion(t)))
}

//export goSizeTrampoline
func goSizeTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer, outSize *C.GhosttySizeReportSize) C.bool {
	t := terminalFromUserdata(userdata)
	if t.onSize == nil {
		return C.bool(false)
	}
	size, ok := t.onSize(t)
	if !ok {
		return C.bool(false)
	}
	outSize.rows = C.uint16_t(size.Rows)
	outSize.columns = C.uint16_t(size.Columns)
	outSize.cell_width = C.uint32_t(size.CellWidth)
	outSize.cell_height = C.uint32_t(size.CellHeight)
	return C.bool(true)
}

//export goColorSchemeTrampoline
func goColorSchemeTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer, outScheme *C.GhosttyColorScheme) C.bool {
	t := terminalFromUserdata(userdata)
	if t.onColorScheme == nil {
		return C.bool(false)
	}
	scheme, ok := t.onColorScheme(t)
	if !ok {
		return C.bool(false)
	}
	*outScheme = C.GhosttyColorScheme(scheme)
	return C.bool(true)
}

//export goDeviceAttributesTrampoline
func goDeviceAttributesTrampoline(_ C.GhosttyTerminal, userdata unsafe.Pointer, outAttrs *C.GhosttyDeviceAttributes) C.bool {
	t := terminalFromUserdata(userdata)
	if t.onDeviceAttributes == nil {
		return C.bool(false)
	}
	attrs, ok := t.onDeviceAttributes(t)
	if !ok {
		return C.bool(false)
	}

	// Primary (DA1).
	outAttrs.primary.conformance_level = C.uint16_t(attrs.Primary.ConformanceLevel)
	outAttrs.primary.num_features = C.size_t(attrs.Primary.NumFeatures)
	for i := 0; i < attrs.Primary.NumFeatures && i < 64; i++ {
		outAttrs.primary.features[i] = C.uint16_t(attrs.Primary.Features[i])
	}

	// Secondary (DA2).
	outAttrs.secondary.device_type = C.uint16_t(attrs.Secondary.DeviceType)
	outAttrs.secondary.firmware_version = C.uint16_t(attrs.Secondary.FirmwareVersion)
	outAttrs.secondary.rom_cartridge = C.uint16_t(attrs.Secondary.ROMCartridge)

	// Tertiary (DA3).
	outAttrs.tertiary.unit_id = C.uint32_t(attrs.Tertiary.UnitID)

	return C.bool(true)
}

// effectString copies data into C memory allocated via the libghostty
// allocator, updates effectBuf/effectBufLen, and returns a
// GhosttyString pointing to it. The previous effectBuf is freed.
// Returns a zero-length GhosttyString if data is empty.
func (t *Terminal) effectString(data []byte) C.GhosttyString {
	if t.effectBuf != nil {
		Free(t.effectBuf, t.effectBufLen)
		t.effectBuf = nil
		t.effectBufLen = 0
	}

	if len(data) == 0 {
		return C.GhosttyString{}
	}

	n := uintptr(len(data))
	cmem := Alloc(n)
	copy(unsafe.Slice((*byte)(cmem), n), data)
	t.effectBuf = cmem
	t.effectBufLen = n
	return C.GhosttyString{
		ptr: (*C.uint8_t)(cmem),
		len: C.size_t(n),
	}
}
