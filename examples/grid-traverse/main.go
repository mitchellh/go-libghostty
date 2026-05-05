// Example grid-traverse demonstrates walking the terminal grid
// cell-by-cell using the GridRef API to inspect content and style.
package main

import (
	"fmt"
	"log"

	"go.mitchellh.com/libghostty"
)

func main() {
	term, err := libghostty.NewTerminal(libghostty.WithSize(10, 3))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// Write some content: two plain lines and one bold line.
	term.VTWrite([]byte("Hello!\r\n"))
	term.VTWrite([]byte("World\r\n"))
	term.VTWrite([]byte("\033[1mBold"))

	cols, err := term.Cols()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := term.Rows()
	if err != nil {
		log.Fatal(err)
	}

	for row := range rows {
		fmt.Printf("Row %d: ", row)

		for col := range cols {
			ref, err := term.GridRef(libghostty.Point{
				Tag: libghostty.PointTagActive,
				X:   col,
				Y:   uint32(row),
			})
			if err != nil {
				log.Fatal(err)
			}

			cell, err := ref.Cell()
			if err != nil {
				log.Fatal(err)
			}

			hasText, err := cell.HasText()
			if err != nil {
				log.Fatal(err)
			}

			if hasText {
				cp, err := cell.Codepoint()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%c", rune(cp))
			} else {
				fmt.Print(".")
			}
		}

		// Print wrap and bold state for the first cell in the row.
		ref, err := term.GridRef(libghostty.Point{
			Tag: libghostty.PointTagActive,
			X:   0,
			Y:   uint32(row),
		})
		if err != nil {
			log.Fatal(err)
		}

		rowData, err := ref.Row()
		if err != nil {
			log.Fatal(err)
		}
		wrap, err := rowData.Wrap()
		if err != nil {
			log.Fatal(err)
		}

		style, err := ref.Style()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(" (wrap=%t, bold=%t)\n", wrap, style.Bold())
	}
}
