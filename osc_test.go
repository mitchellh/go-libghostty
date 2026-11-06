package libghostty

import (
	"errors"
	"testing"
	"unsafe"
)

func TestOSCParserWindowTitle(t *testing.T) {
	parser, err := NewOSCParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	for _, b := range []byte("0;hello") {
		parser.Next(b)
	}
	command := parser.End('\a')
	if command.Type() != OSCCommandChangeWindowTitle {
		t.Fatalf("expected title command, got %d", command.Type())
	}
	title, ok := command.WindowTitle()
	if !ok || title != "hello" {
		t.Fatalf("expected title %q, got %q (ok=%v)", "hello", title, ok)
	}

	parser.Reset()
	for _, b := range []byte("999999;unsupported") {
		parser.Next(b)
	}
	if got := parser.End('\a').Type(); got != OSCCommandInvalid {
		t.Fatalf("expected invalid command, got %d", got)
	}
}

func TestOSCCommandTypesLatestProtocols(t *testing.T) {
	types := []OSCCommandType{
		OSCCommandKittyClipboardProtocol,
		OSCCommandKittyDNDProtocol,
		OSCCommandHierarchicalContext,
		OSCCommandKittyDesktopNotification,
	}
	seen := make(map[OSCCommandType]struct{}, len(types))
	for _, commandType := range types {
		if commandType == OSCCommandInvalid {
			t.Fatal("expected latest OSC command type to be valid")
		}
		if _, ok := seen[commandType]; ok {
			t.Fatalf("duplicate OSC command type value %d", commandType)
		}
		seen[commandType] = struct{}{}
	}

	// The deprecated name must keep matching the command it was renamed to.
	if OSCCommandContextSignal != OSCCommandHierarchicalContext {
		t.Fatal("expected OSCCommandContextSignal to alias OSCCommandHierarchicalContext")
	}
}

func TestOSCParserUnknown(t *testing.T) {
	p, err := NewOSCParser()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	parse := func(body string, end byte) OSCCommand {
		p.Reset()
		for _, b := range []byte(body) {
			p.Next(b)
		}
		return p.End(end)
	}
	if got := parse("7400;hello", 7).Type(); got != OSCCommandInvalid {
		t.Fatalf("capture disabled: %v", got)
	}
	if err := p.SetUnknownMaxBytes(8); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		end  byte
		want OSCTerminator
	}{{7, OSCTerminatorBEL}, {'\\', OSCTerminatorST}} {
		c := parse("7400;hello", tc.end)
		if c.Type() != OSCCommandUnknown {
			t.Fatalf("type: %v", c.Type())
		}
		body, ok := c.UnknownContent()
		if !ok || string(body) != "7400;hel" {
			t.Fatalf("body: %q, %v", body, ok)
		}
		if truncated, ok := c.UnknownTruncated(); !ok || !truncated {
			t.Fatalf("truncation: %v, %v", truncated, ok)
		}
		if end, ok := c.UnknownTerminator(); !ok || end != tc.want {
			t.Fatalf("terminator: %v, %v", end, ok)
		}
		parse("2;title", 7)
		if string(body) != "7400;hel" {
			t.Fatal("captured content was not copied")
		}
	}
	// Cancelled sequences are discarded whether or not they contain a
	// recognized command.
	for _, end := range []byte{0x18, 0x1a} {
		for _, body := range []string{"7400;hello", "2;title"} {
			if got := parse(body, end).Type(); got != OSCCommandInvalid {
				t.Fatalf("cancelled %q: %v", body, got)
			}
		}
	}
	// A recognized but malformed command must not turn into an extension.
	if got := parse("22;not-a-shape", 7).Type(); got == OSCCommandUnknown {
		t.Fatal("recognized command reported as unknown")
	}
	c := parse("2;title", 7)
	if _, ok := c.UnknownContent(); ok {
		t.Fatal("title has unknown content")
	}
	if _, ok := c.UnknownTruncated(); ok {
		t.Fatal("title has unknown truncation")
	}
	if _, ok := c.UnknownTerminator(); ok {
		t.Fatal("title has unknown terminator")
	}
	if err := p.SetUnknownMaxBytes(4096); err != nil {
		t.Fatal(err)
	}
	body := "7400;" + string(make([]byte, 3000))
	c = parse(body, 7)
	if got, ok := c.UnknownContent(); !ok || string(got) != body {
		t.Fatal("allocated capture did not preserve binary content")
	}
	if truncated, ok := c.UnknownTruncated(); !ok || truncated {
		t.Fatal("unexpected truncation")
	}
	if err := p.SetUnknownMaxBytes(0); err != nil {
		t.Fatal(err)
	}
	if got := parse("7400;hello", 7).Type(); got != OSCCommandInvalid {
		t.Fatalf("disabled again: %v", got)
	}
}

func TestOSCParserProgramStatus(t *testing.T) {
	parser, err := NewOSCParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	// The parser only identifies OSC 7501. Report contents are delivered
	// through a terminal's ProgramStatusFunc instead.
	for _, body := range []string{"7501;state=working", "7501;?"} {
		parser.Reset()
		for _, b := range []byte(body) {
			parser.Next(b)
		}
		if got := parser.End('\a').Type(); got != OSCCommandProgramStatus {
			t.Fatalf("%q: expected program status command, got %d", body, got)
		}
	}
}

// TestOSCCommandWindowTitle reads a title through the GhosttyString output of
// GHOSTTY_OSC_DATA_CHANGE_WINDOW_TITLE_STR.
func TestOSCCommandWindowTitle(t *testing.T) {
	parser, err := NewOSCParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	parse := func(body string) OSCCommand {
		parser.Reset()
		for _, b := range []byte(body) {
			parser.Next(b)
		}
		return parser.End('\a')
	}

	for _, title := range []string{"hello world", "tïtle ✓", "", "nul\x00byte"} {
		command := parse("2;" + title)
		if got := command.Type(); got != OSCCommandChangeWindowTitle {
			t.Fatalf("%q: expected change-window-title, got %d", title, got)
		}
		got, ok := command.WindowTitle()
		if !ok || got != title {
			t.Fatalf("WindowTitle() = %q, %v; want %q", got, ok, title)
		}
	}

	// Another command has no title, which is GHOSTTY_NO_VALUE rather than a
	// failure of the call itself.
	command := parse("1;icon")
	if _, ok := command.WindowTitle(); ok {
		t.Fatal("expected a non-title command to have no title")
	}
	// Test files can't use cgo, so this mirrors the layout of GhosttyString.
	var title struct {
		ptr unsafe.Pointer
		len uintptr
	}
	err = command.Data(OSCDataChangeWindowTitleString, unsafe.Pointer(&title))
	if !errors.Is(err, ErrNoValue) {
		t.Fatalf("expected no value, got %v", err)
	}
}
