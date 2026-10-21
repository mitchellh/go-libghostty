package libghostty_test

import (
	"fmt"
	"log"

	"go.mitchellh.com/libghostty"
)

// This example writes program output to a terminal and prints the content
// of every APC string in it, such as Kitty graphics commands.
func ExampleTerminal_VTWriteUntil() {
	term, err := libghostty.NewTerminal(libghostty.WithSize(80, 24))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	output := []byte("hello \x1b_Gf=24,s=1,v=1;AAAA\x1b\\world \x1b_Ga=d\x1b\\")
	want := libghostty.TerminalBoundaryAPCStart | libghostty.TerminalBoundaryAPCEnd

	// A string can span several calls, so collect its content as it arrives.
	var content []byte
	inside := false

	data := output
	for len(data) > 0 {
		stop, err := term.VTWriteUntil(data, want)
		if err != nil {
			log.Fatal(err)
		}
		consumed := data[:stop.Consumed]
		data = data[stop.Consumed:]

		switch {
		case stop.Reached&libghostty.TerminalBoundaryAPCStart != 0:
			// The bytes before the start are ordinary output, and the
			// content begins with the next byte.
			inside = true

		case stop.Reached&libghostty.TerminalBoundaryAPCEnd != 0:
			// The last consumed byte ended the string, so it is not part
			// of the content.
			content = append(content, consumed[:len(consumed)-1]...)
			fmt.Printf("APC: %q\n", content)
			content = content[:0]
			inside = false

		case inside:
			// No boundary in the rest of the input, but the string is
			// still open, so all of it is content.
			content = append(content, consumed...)
		}
	}

	// Output:
	// APC: "Gf=24,s=1,v=1;AAAA"
	// APC: "Ga=d"
}
