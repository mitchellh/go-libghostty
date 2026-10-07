package libghostty

// OSC parser bindings from osc.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import "unsafe"

// OSCCommandType identifies a parsed Operating System Command.
// C: GhosttyOscCommandType
type OSCCommandType int

const (
	// OSCCommandInvalid identifies an invalid or unsupported command.
	OSCCommandInvalid OSCCommandType = C.GHOSTTY_OSC_COMMAND_INVALID

	// OSCCommandChangeWindowTitle changes the window title.
	OSCCommandChangeWindowTitle OSCCommandType = C.GHOSTTY_OSC_COMMAND_CHANGE_WINDOW_TITLE

	// OSCCommandChangeWindowIcon changes the window icon.
	OSCCommandChangeWindowIcon OSCCommandType = C.GHOSTTY_OSC_COMMAND_CHANGE_WINDOW_ICON

	// OSCCommandSemanticPrompt carries semantic prompt metadata.
	OSCCommandSemanticPrompt OSCCommandType = C.GHOSTTY_OSC_COMMAND_SEMANTIC_PROMPT

	// OSCCommandClipboardContents accesses clipboard contents.
	OSCCommandClipboardContents OSCCommandType = C.GHOSTTY_OSC_COMMAND_CLIPBOARD_CONTENTS

	// OSCCommandReportPwd reports the present working directory.
	OSCCommandReportPwd OSCCommandType = C.GHOSTTY_OSC_COMMAND_REPORT_PWD

	// OSCCommandMouseShape changes the pointer shape.
	OSCCommandMouseShape OSCCommandType = C.GHOSTTY_OSC_COMMAND_MOUSE_SHAPE

	// OSCCommandColorOperation performs a color query or update.
	OSCCommandColorOperation OSCCommandType = C.GHOSTTY_OSC_COMMAND_COLOR_OPERATION

	// OSCCommandKittyColorProtocol uses Kitty's color protocol.
	OSCCommandKittyColorProtocol OSCCommandType = C.GHOSTTY_OSC_COMMAND_KITTY_COLOR_PROTOCOL

	// OSCCommandShowDesktopNotification requests a desktop notification.
	OSCCommandShowDesktopNotification OSCCommandType = C.GHOSTTY_OSC_COMMAND_SHOW_DESKTOP_NOTIFICATION

	// OSCCommandHyperlinkStart begins an OSC 8 hyperlink.
	OSCCommandHyperlinkStart OSCCommandType = C.GHOSTTY_OSC_COMMAND_HYPERLINK_START

	// OSCCommandHyperlinkEnd ends an OSC 8 hyperlink.
	OSCCommandHyperlinkEnd OSCCommandType = C.GHOSTTY_OSC_COMMAND_HYPERLINK_END

	// OSCCommandConEmuSleep is a ConEmu sleep command.
	OSCCommandConEmuSleep OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_SLEEP

	// OSCCommandConEmuShowMessageBox is a ConEmu message-box command.
	OSCCommandConEmuShowMessageBox OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_SHOW_MESSAGE_BOX

	// OSCCommandConEmuChangeTabTitle is a ConEmu tab-title command.
	OSCCommandConEmuChangeTabTitle OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_CHANGE_TAB_TITLE

	// OSCCommandConEmuProgressReport is a ConEmu progress command.
	OSCCommandConEmuProgressReport OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_PROGRESS_REPORT

	// OSCCommandConEmuWaitInput is a ConEmu wait-input command.
	OSCCommandConEmuWaitInput OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_WAIT_INPUT

	// OSCCommandConEmuGuiMacro is a ConEmu GUI macro command.
	OSCCommandConEmuGuiMacro OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_GUIMACRO

	// OSCCommandConEmuRunProcess is a ConEmu process command.
	OSCCommandConEmuRunProcess OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_RUN_PROCESS

	// OSCCommandConEmuOutputEnvironmentVariable is a ConEmu environment
	// variable command.
	OSCCommandConEmuOutputEnvironmentVariable OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_OUTPUT_ENVIRONMENT_VARIABLE

	// OSCCommandConEmuXtermEmulation is a ConEmu xterm-emulation command.
	OSCCommandConEmuXtermEmulation OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_XTERM_EMULATION

	// OSCCommandConEmuComment is a ConEmu comment command.
	OSCCommandConEmuComment OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONEMU_COMMENT

	// OSCCommandKittyTextSizing uses Kitty's text-sizing protocol.
	OSCCommandKittyTextSizing OSCCommandType = C.GHOSTTY_OSC_COMMAND_KITTY_TEXT_SIZING

	// OSCCommandKittyClipboardProtocol uses Kitty's clipboard protocol.
	OSCCommandKittyClipboardProtocol OSCCommandType = C.GHOSTTY_OSC_COMMAND_KITTY_CLIPBOARD_PROTOCOL

	// OSCCommandKittyDNDProtocol uses Kitty's drag-and-drop protocol.
	OSCCommandKittyDNDProtocol OSCCommandType = C.GHOSTTY_OSC_COMMAND_KITTY_DND_PROTOCOL

	// OSCCommandContextSignal carries a terminal context signal.
	OSCCommandContextSignal OSCCommandType = C.GHOSTTY_OSC_COMMAND_CONTEXT_SIGNAL

	// OSCCommandKittyDesktopNotification uses Kitty's desktop notification protocol.
	OSCCommandKittyDesktopNotification OSCCommandType = C.GHOSTTY_OSC_COMMAND_KITTY_DESKTOP_NOTIFICATION

	// OSCCommandUnknown identifies an OSC command whose number the parser
	// does not recognize. Enable capture with [OSCParser.SetUnknownMaxBytes].
	OSCCommandUnknown OSCCommandType = C.GHOSTTY_OSC_COMMAND_UNKNOWN

	// OSCCommandProgramStatus identifies an OSC 7501 program status report,
	// or a program asking whether the terminal supports those reports. The
	// parser doesn't decode the report itself. To read reports, create a
	// [Terminal] with [WithProgramStatus].
	OSCCommandProgramStatus OSCCommandType = C.GHOSTTY_OSC_COMMAND_PROGRAM_STATUS
)

// OSCCommandData identifies typed data extractable from an OSC command.
// C: GhosttyOscCommandData
type OSCCommandData int

const (
	// OSCDataInvalid is an invalid data query.
	OSCDataInvalid OSCCommandData = C.GHOSTTY_OSC_DATA_INVALID

	// OSCDataChangeWindowTitleString extracts a null-terminated title string.
	OSCDataChangeWindowTitleString OSCCommandData = C.GHOSTTY_OSC_DATA_CHANGE_WINDOW_TITLE_STR

	// OSCDataUnknownContent extracts an unknown command's body as a
	// GhosttyString. The bytes belong to the parser and remain valid only
	// until the next operation on it. Use [OSCCommand.UnknownContent] for a copy.
	OSCDataUnknownContent OSCCommandData = C.GHOSTTY_OSC_DATA_UNKNOWN_CONTENT

	// OSCDataUnknownTruncated extracts whether an unknown command's content
	// was shortened by the capture limit or an allocation failure (bool).
	OSCDataUnknownTruncated OSCCommandData = C.GHOSTTY_OSC_DATA_UNKNOWN_TRUNCATED

	// OSCDataUnknownTerminator extracts how an unknown command ended
	// (GhosttyOscTerminator).
	OSCDataUnknownTerminator OSCCommandData = C.GHOSTTY_OSC_DATA_UNKNOWN_TERMINATOR
)

// OSCTerminator identifies how an OSC sequence ended. Replies should use the
// same terminator as the request.
//
// These constants identify terminators but are not byte values. [OSCParser.End]
// takes the terminating byte instead.
//
// C: GhosttyOscTerminator
type OSCTerminator int

const (
	// OSCTerminatorST is the string terminator: ESC (0x1b) followed by a backslash (0x5c).
	OSCTerminatorST OSCTerminator = C.GHOSTTY_OSC_TERMINATOR_ST

	// OSCTerminatorBEL is the bell byte (0x07).
	OSCTerminatorBEL OSCTerminator = C.GHOSTTY_OSC_TERMINATOR_BEL
)

// OSCParser incrementally parses the bytes inside an OSC sequence.
// C: GhosttyOscParser
type OSCParser struct {
	ptr C.GhosttyOscParser
}

// OSCCommand is a borrowed command produced by [OSCParser.End]. It remains
// valid until the next parser operation other than command introspection.
// C: GhosttyOscCommand
type OSCCommand struct {
	ptr C.GhosttyOscCommand
}

// NewOSCParser creates a reusable OSC parser.
func NewOSCParser() (*OSCParser, error) {
	var ptr C.GhosttyOscParser
	if err := resultError(C.ghostty_osc_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &OSCParser{ptr: ptr}, nil
}

// Close frees the parser. Commands borrowed from it become invalid. Calling
// Close again is a safe no-op.
func (p *OSCParser) Close() {
	if p == nil || p.ptr == nil {
		return
	}
	C.ghostty_osc_free(p.ptr)
	p.ptr = nil
}

// Reset clears the current sequence without changing parser options.
// Call Reset before parsing another sequence. Commands returned by earlier
// calls to [OSCParser.End] are no longer valid.
func (p *OSCParser) Reset() {
	C.ghostty_osc_reset(p.ptr)
}

// Next feeds one byte from the OSC sequence body into the parser.
func (p *OSCParser) Next(b byte) {
	C.ghostty_osc_next(p.ptr, C.uint8_t(b))
}

// End finishes parsing the current sequence and returns its command.
// The terminator is the byte that ended the sequence. This is usually BEL
// (0x07), or the backslash (0x5c) that ends an ESC \ string terminator.
// Commands that send a reply, such as color queries, end the reply the same
// way the request ended.
//
// A program can also cancel a sequence partway through by sending CAN (0x18)
// or SUB (0x1a). Pass that byte as the terminator. The sequence is then
// discarded, even if the bytes before the cancel form a complete command.
// This matches the behavior of xterm.
//
// If the sequence is invalid or was cancelled, the returned command has type
// [OSCCommandInvalid]. The command is valid until the next call to a method
// on p. Call [OSCParser.Reset] before parsing the next sequence.
func (p *OSCParser) End(terminator byte) OSCCommand {
	return OSCCommand{
		ptr: C.ghostty_osc_end(p.ptr, C.uint8_t(terminator)),
	}
}

// Type returns the parsed command type.
func (c OSCCommand) Type() OSCCommandType {
	return OSCCommandType(C.ghostty_osc_command_type(c.ptr))
}

// Data extracts a low-level typed value into out. The pointed-to Go value
// must match the output type documented for data in osc.h.
func (c OSCCommand) Data(data OSCCommandData, out unsafe.Pointer) bool {
	return bool(C.ghostty_osc_command_data(
		c.ptr,
		C.GhosttyOscCommandData(data),
		out,
	))
}

// WindowTitle returns the title carried by a change-window-title command.
// The string is copied into Go-owned memory.
func (c OSCCommand) WindowTitle() (string, bool) {
	var ptr *C.char
	if !c.Data(OSCDataChangeWindowTitleString, unsafe.Pointer(&ptr)) {
		return "", false
	}
	return C.GoString(ptr), true
}

// SetUnknownMaxBytes sets the maximum number of bytes retained for each OSC
// sequence whose command number the parser does not recognize. Zero, the
// default, disables capture. A positive limit enables [OSCCommandUnknown].
//
// Longer sequences are still reported, but contain only the first limit bytes.
// [OSCCommand.UnknownTruncated] reports whether content was lost. Recognized
// commands are never reported as unknown, even when their contents are invalid.
//
// The limit remains set across calls to [OSCParser.Reset]. Set it before feeding
// a sequence, since a sequence already in progress may keep the previous limit.
// Limits up to 2048 bytes use the parser's existing buffer. Larger limits may
// allocate memory for each unknown sequence.
//
// C: ghostty_osc_set, GHOSTTY_OSC_OPT_UNKNOWN_MAX_BYTES
func (p *OSCParser) SetUnknownMaxBytes(limit uint) error {
	v := C.size_t(limit)
	return resultError(C.ghostty_osc_set(p.ptr, C.GHOSTTY_OSC_OPT_UNKNOWN_MAX_BYTES, unsafe.Pointer(&v)))
}

// UnknownContent returns a copy of an unknown command's body, including its
// command number and excluding the sequence delimiters. For example, the OSC
// sequence "\x1b]7400;hello\a" has the body "7400;hello". The bytes may be kept
// or modified after the parser is reset or closed.
//
// It returns nil, false if c is not [OSCCommandUnknown] or the content cannot
// be represented as a Go byte slice.
func (c OSCCommand) UnknownContent() (content []byte, ok bool) {
	var v C.GhosttyString
	if !c.Data(OSCDataUnknownContent, unsafe.Pointer(&v)) {
		return nil, false
	}
	return copyGhosttyString(v)
}

// UnknownTruncated reports whether an unknown command's content was shortened
// by the capture limit or a memory allocation failure. When truncated is true,
// [OSCCommand.UnknownContent] returns only the beginning of the sequence.
// The ok result is false if c is not [OSCCommandUnknown].
func (c OSCCommand) UnknownTruncated() (truncated bool, ok bool) {
	var v C.bool
	ok = c.Data(OSCDataUnknownTruncated, unsafe.Pointer(&v))
	return bool(v), ok
}

// UnknownTerminator returns how an unknown command ended. Use the same
// terminator when replying to the command. The ok result is false if c is not
// [OSCCommandUnknown].
func (c OSCCommand) UnknownTerminator() (terminator OSCTerminator, ok bool) {
	var v C.GhosttyOscTerminator
	ok = c.Data(OSCDataUnknownTerminator, unsafe.Pointer(&v))
	return OSCTerminator(v), ok
}
