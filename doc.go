// Package libghostty provides Go bindings for libghostty-vt, a
// virtual terminal emulator library from the Ghostty project.
//
// # Getting Started
//
// Create a terminal with [NewTerminal], feed it input with
// [Terminal.VTWrite] (or [Terminal.Write] for an [io.Writer]), and
// inspect state through data getters such as [Terminal.CursorX],
// [Terminal.Title], and [Terminal.ActiveScreen]. When finished, call
// [Terminal.Close] to release resources.
//
//	term, err := libghostty.NewTerminal(
//		libghostty.WithSize(80, 24),
//		libghostty.WithMaxScrollback(1000),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer term.Close()
//
//	term.VTWrite([]byte("Hello, world!\r\n"))
//
// # Concurrency
//
// Unless documented otherwise, exported handle types in this package are
// not safe for concurrent use. Keep each [Terminal], [Formatter],
// [KeyEncoder], [MouseEncoder], [KeyEvent], [MouseEvent], and borrowed
// view confined to one goroutine at a time or protect it with your own
// synchronization.
//
// [RenderState] is the main exception. Hold exclusive access to the
// terminal while calling [RenderState.Update]. After Update returns, the
// render state can be read without touching the terminal until the next
// Update. Do not call Update concurrently with reads from the same
// render state.
//
// Borrowed views such as [GridRef], [KittyGraphics], [KittyGraphicsImage],
// and [Selection], plus raw pixel slices returned by Kitty graphics
// accessors, are only valid until the next mutating terminal call. Read
// and copy what you need before mutating the terminal again.
//
// Plain copied values such as [Cell], [Row], [Style], [ColorRGB], and
// [Palette] are regular Go values and may be retained after the call
// that produced them.
//
// # Effects
//
// The terminal communicates side-effects back to the host through
// effect callbacks. Register them at creation time with functional
// options like [WithWritePty], [WithBell], and [WithEnquiry], or
// on a live terminal with [Terminal.SetEffectWritePty] and friends.
//
// Effect callbacks run synchronously during [Terminal.VTWrite]. They
// must not call [Terminal.VTWrite] on the same terminal and should avoid
// blocking for long periods.
//
// [WithWritePty] is the most common effect — it delivers data that
// the terminal wants to send back to the pty (e.g. query responses):
//
//	term, _ := libghostty.NewTerminal(
//		libghostty.WithSize(80, 24),
//		libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
//			os.Stdout.Write(data)
//		}),
//	)
//
// # Terminal Options
//
// Terminal properties can be changed after creation with setter
// methods such as [Terminal.SetColorForeground],
// [Terminal.SetColorBackground], [Terminal.SetColorPalette],
// [Terminal.SetTitle], and [Terminal.SetPwd].
//
// # Linking
//
// This is a cgo package. By default it links the shared library via
// pkg-config. Build with "-tags static" to link statically instead.
package libghostty
