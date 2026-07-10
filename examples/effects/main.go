// Example program demonstrating terminal effect callbacks.
//
// It registers write_pty, bell, title_changed, and clipboard_write effect
// handlers, then feeds VT sequences that trigger each one. Output shows how
// the callbacks fire and how terminal state can be queried from within them.
package main

import (
	"fmt"
	"log"

	ghostty "go.mitchellh.com/libghostty"
)

func main() {
	// Bell counter, captured by the bell handler closure.
	bellCount := 0

	term, err := ghostty.NewTerminal(
		ghostty.WithSize(80, 24),

		// write_pty: called when the terminal writes data back (e.g. query responses).
		ghostty.WithWritePty(func(_ *ghostty.Terminal, data []byte) {
			fmt.Printf("write_pty: %d bytes: %q\n", len(data), data)
		}),

		// bell: called on BEL (0x07).
		ghostty.WithBell(func(_ *ghostty.Terminal) {
			bellCount++
			fmt.Printf("bell: count=%d\n", bellCount)
		}),

		// clipboard_write: called with normalized, decoded clipboard contents.
		ghostty.WithClipboardWrite(func(_ *ghostty.Terminal, write ghostty.ClipboardWrite) ghostty.ClipboardWriteResult {
			fmt.Printf("clipboard_write: location=%d contents=%d\n", write.Location, len(write.Contents))
			for _, content := range write.Contents {
				fmt.Printf("  %s: %q\n", content.MIME, content.Data)
			}
			return ghostty.ClipboardWriteSuccess
		}),

		// title_changed: called when the terminal title changes via OSC 0/2.
		// The terminal is passed directly as a parameter.
		ghostty.WithTitleChanged(func(t *ghostty.Terminal) {
			x, err := t.CursorX()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("title_changed: cursor_x=%d\n", x)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// BEL → triggers bell handler.
	term.VTWrite([]byte{0x07})

	// OSC 2 (set title) → triggers title_changed handler.
	term.VTWrite([]byte("\x1b]2;hello\x1b\\"))

	// DECRQM query → triggers write_pty with the response.
	term.VTWrite([]byte("\x1b[?7$p"))

	// OSC 52 (set clipboard) → triggers clipboard_write with decoded data.
	term.VTWrite([]byte("\x1b]52;c;SGVsbG8gY2xpcGJvYXJk\x1b\\"))

	// Another BEL → triggers bell handler again.
	term.VTWrite([]byte{0x07})

	fmt.Printf("total bell count: %d\n", bellCount)
}
