// Example render demonstrates the RenderState API by creating a terminal,
// writing styled VT content, and iterating over the resulting rows and
// cells to produce ANSI-colored output.
package main

import (
	"fmt"
	"log"

	"go.mitchellh.com/libghostty"
)

// resolveColor converts a StyleColor to a concrete ColorRGB using the
// render state's palette and a fallback for unset colors.
func resolveColor(sc libghostty.StyleColor, colors *libghostty.RenderStateColors, fallback libghostty.ColorRGB) libghostty.ColorRGB {
	switch sc.Tag {
	case libghostty.StyleColorRGB:
		return sc.RGB
	case libghostty.StyleColorPalette:
		return colors.Palette[sc.Palette]
	default:
		return fallback
	}
}

// cursorStyleName returns a human-readable name for a cursor visual style.
func cursorStyleName(s libghostty.CursorVisualStyle) string {
	switch s {
	case libghostty.CursorVisualStyleBar:
		return "bar"
	case libghostty.CursorVisualStyleBlock:
		return "block"
	case libghostty.CursorVisualStyleUnderline:
		return "underline"
	case libghostty.CursorVisualStyleBlockHollow:
		return "block_hollow"
	default:
		return "unknown"
	}
}

func main() {
	// 1. Create terminal 40x5 with scrollback 10000.
	term, err := libghostty.NewTerminal(
		libghostty.WithSize(40, 5),
		libghostty.WithMaxScrollback(10000),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// 2. Create render state.
	rs, err := libghostty.NewRenderState()
	if err != nil {
		log.Fatal(err)
	}
	defer rs.Close()

	// 3. Write styled VT content.
	term.VTWrite([]byte("Hello, \033[1;32mworld\033[0m!\r\n"))
	term.VTWrite([]byte("\033[4munderlined\033[0m text\r\n"))
	term.VTWrite([]byte("\033[38;2;255;128;0morange\033[0m\r\n"))

	// 4. Update render state from terminal.
	if err := rs.Update(term); err != nil {
		log.Fatal(err)
	}

	// 5. Check and print dirty state.
	dirty, err := rs.Dirty()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("dirty: %d\n", dirty)

	// 6. Get and print colors.
	colors, err := rs.Colors()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bg: #%02x%02x%02x\n", colors.Background.R, colors.Background.G, colors.Background.B)
	fmt.Printf("fg: #%02x%02x%02x\n", colors.Foreground.R, colors.Foreground.G, colors.Foreground.B)

	// 7. Cursor information.
	cursorVisible, err := rs.CursorVisible()
	if err != nil {
		log.Fatal(err)
	}
	cursorHasValue, err := rs.CursorViewportHasValue()
	if err != nil {
		log.Fatal(err)
	}
	if cursorVisible && cursorHasValue {
		cx, err := rs.CursorViewportX()
		if err != nil {
			log.Fatal(err)
		}
		cy, err := rs.CursorViewportY()
		if err != nil {
			log.Fatal(err)
		}
		style, err := rs.CursorVisualStyle()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("cursor: x=%d y=%d style=%s\n", cx, cy, cursorStyleName(style))
	} else {
		fmt.Printf("cursor: not visible\n")
	}

	// 8. Iterate rows and cells.
	ri, err := libghostty.NewRenderStateRowIterator()
	if err != nil {
		log.Fatal(err)
	}
	defer ri.Close()

	rc, err := libghostty.NewRenderStateRowCells()
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	if err := rs.RowIterator(ri); err != nil {
		log.Fatal(err)
	}

	for ri.Next() {
		rowDirty, err := ri.Dirty()
		if err != nil {
			log.Fatal(err)
		}
		_ = rowDirty

		if err := ri.Cells(rc); err != nil {
			log.Fatal(err)
		}

		for rc.Next() {
			graphemes, err := rc.Graphemes()
			if err != nil {
				log.Fatal(err)
			}
			if len(graphemes) == 0 {
				continue
			}

			style, err := rc.Style()
			if err != nil {
				log.Fatal(err)
			}

			// Resolve foreground color.
			fg := resolveColor(style.FgColor(), colors, colors.Foreground)

			// Emit ANSI true-color escape for foreground.
			fmt.Printf("\033[38;2;%d;%d;%dm", fg.R, fg.G, fg.B)

			// Bold marker.
			if style.Bold() {
				fmt.Printf("\033[1m")
			}

			// Underline marker.
			if style.Underline() != libghostty.UnderlineNone {
				fmt.Printf("\033[4m")
			}

			// Print codepoints.
			for _, cp := range graphemes {
				fmt.Printf("%c", rune(cp))
			}

			// Reset style after each cell.
			fmt.Printf("\033[0m")
		}

		// Clear row dirty flag.
		if err := ri.SetDirty(false); err != nil {
			log.Fatal(err)
		}

		fmt.Println()
	}

	// 9. Reset global dirty state.
	if err := rs.SetDirty(libghostty.RenderStateDirtyFalse); err != nil {
		log.Fatal(err)
	}
}
