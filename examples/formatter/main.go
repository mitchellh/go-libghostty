// Example program demonstrating the Formatter API from libghostty.
// It creates a terminal, writes various VT sequences to it, then
// formats the terminal contents as plain text with trimming enabled.
package main

import (
	"fmt"
	"log"

	lg "go.mitchellh.com/libghostty"
)

func main() {
	// Create an 80x24 terminal.
	term, err := lg.NewTerminal(lg.WithSize(80, 24))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// Write some content with VT formatting.
	fmt.Fprintf(term, "Line 1: Hello World!\r\n")
	fmt.Fprintf(term, "Line 2: \033[1mBold\033[0m and \033[4mUnderline\033[0m\r\n")
	fmt.Fprintf(term, "Line 3: placeholder\r\n")

	// Move to row 3, col 1 and overwrite line 3.
	fmt.Fprintf(term, "\033[3;1H") // CUP row 3 col 1
	fmt.Fprintf(term, "\033[2K")   // Erase entire line
	fmt.Fprintf(term, "Line 3: Overwritten!\r\n")

	// Place text at specific positions.
	fmt.Fprintf(term, "\033[5;10H") // CUP row 5 col 10
	fmt.Fprintf(term, "Placed at (5,10)")
	fmt.Fprintf(term, "\033[1;72H") // CUP row 1 col 72
	fmt.Fprintf(term, "RIGHT->")

	// Create a plain-text formatter with trimming enabled.
	f, err := lg.NewFormatter(term,
		lg.WithFormatterFormat(lg.FormatterFormatPlain),
		lg.WithFormatterTrim(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Format and print the output.
	output, err := f.FormatString()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", output)
	fmt.Printf("(%d bytes)\n", len(output))
}
