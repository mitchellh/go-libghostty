package libghostty

import "testing"

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
		OSCCommandContextSignal,
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
}
