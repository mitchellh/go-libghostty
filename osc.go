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
)

// OSCCommandData identifies typed data extractable from an OSC command.
// C: GhosttyOscCommandData
type OSCCommandData int

const (
	// OSCDataInvalid is an invalid data query.
	OSCDataInvalid OSCCommandData = C.GHOSTTY_OSC_DATA_INVALID

	// OSCDataChangeWindowTitleString extracts a null-terminated title string.
	OSCDataChangeWindowTitleString OSCCommandData = C.GHOSTTY_OSC_DATA_CHANGE_WINDOW_TITLE_STR
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

// Close frees the parser. Commands borrowed from it become invalid.
func (p *OSCParser) Close() {
	C.ghostty_osc_free(p.ptr)
}

// Reset clears partially parsed input and returns the parser to its initial
// state.
func (p *OSCParser) Reset() {
	C.ghostty_osc_reset(p.ptr)
}

// Next feeds one byte from the OSC sequence body into the parser.
func (p *OSCParser) Next(b byte) {
	C.ghostty_osc_next(p.ptr, C.uint8_t(b))
}

// End finalizes the current sequence. terminator is normally BEL (0x07) or
// the final backslash byte of ST (0x5c).
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
