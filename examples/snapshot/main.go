// Example snapshot demonstrates saving and restoring a terminal.
//
// It shows both restoration styles:
//   - Decode restores the entire snapshot in one call.
//   - Ready restores a usable terminal first, then Next adds older history
//     one page at a time.
package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"go.mitchellh.com/libghostty"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Continuation tracking lets a snapshot preserve an escape sequence or
	// UTF-8 codepoint that is only partially received. Enable it before writing
	// terminal input. The limit bounds how many unfinished bytes are retained.
	source, err := libghostty.NewTerminal(
		// A wide, shallow terminal fills backing pages quickly, which makes the
		// incremental history portion of the example easy to observe.
		libghostty.WithSize(215, 2),
		libghostty.WithMaxScrollbackLines(2_000),
		libghostty.WithContinuationMaxBytes(1_024),
	)
	if err != nil {
		return fmt.Errorf("create source terminal: %w", err)
	}
	defer source.Close()

	if err := source.SetTitle("snapshot demo"); err != nil {
		return fmt.Errorf("set terminal title: %w", err)
	}

	// Generate enough lines to create multiple scrollback pages. A real
	// program would receive this data from a PTY instead.
	for line := 1; line <= 1_000; line++ {
		fmt.Fprintf(source, "history line %03d\r\n", line)
	}
	fmt.Fprint(source, "prompt> ")

	// Deliberately stop in the middle of an SGR color sequence. The snapshot
	// will preserve these four bytes, and restored input can finish it later.
	source.VTWrite([]byte("\x1b[32"))
	continuation, err := source.Continuation()
	if err != nil {
		return fmt.Errorf("read source continuation: %w", err)
	}
	fmt.Printf("unfinished VT continuation: %q\n", continuation)

	// Snapshot returns a Go-owned byte slice containing the complete terminal
	// state: visible cells, modes, title, scrollback, and parser continuation.
	snapshot, err := source.Snapshot()
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	fmt.Printf("encoded snapshot: %d bytes\n", len(snapshot))

	if err := restoreAllAtOnce(snapshot); err != nil {
		return err
	}
	if err := restoreIncrementally(snapshot); err != nil {
		return err
	}
	return nil
}

// restoreAllAtOnce is the simplest restoration path. Use it when the caller
// can wait for all scrollback to be decoded before using the terminal.
func restoreAllAtOnce(snapshot []byte) error {
	decoder, err := libghostty.NewSnapshotDecoderBytes(snapshot)
	if err != nil {
		return fmt.Errorf("create one-shot decoder: %w", err)
	}
	defer decoder.Close()

	restored, err := decoder.Decode()
	if err != nil {
		return fmt.Errorf("decode complete snapshot: %w", err)
	}
	defer restored.Close()

	// Finish the SGR sequence that was incomplete in the snapshot. If parser
	// continuation was restored correctly, this text becomes green internally.
	restored.VTWrite([]byte("mcontinued after restore\x1b[0m"))

	title, err := restored.Title()
	if err != nil {
		return fmt.Errorf("read restored title: %w", err)
	}
	visible, err := visibleText(restored)
	if err != nil {
		return err
	}

	fmt.Printf("\none-shot restore: title=%q\n", title)
	printIndented(visible)
	return nil
}

// restoreIncrementally is useful for interactive applications. Ready returns
// as soon as the visible terminal is authenticated; Next can then restore old
// scrollback during idle time while the terminal is already usable.
func restoreIncrementally(snapshot []byte) error {
	// The reader constructor accepts any io.Reader. bytes.Reader keeps this
	// example self-contained; a file or network stream works the same way.
	decoder, err := libghostty.NewSnapshotDecoder(bytes.NewReader(snapshot))
	if err != nil {
		return fmt.Errorf("create incremental decoder: %w", err)
	}

	restored, err := decoder.Ready()
	if err != nil {
		decoder.Close()
		return fmt.Errorf("decode snapshot through READY: %w", err)
	}
	// The decoder borrows restored until FINISH, so close the decoder first if
	// this function returns early.
	defer func() {
		decoder.Close()
		restored.Close()
	}()

	historyRows, err := decoder.HistoryRows(libghostty.ScreenPrimary)
	if err != nil {
		return fmt.Errorf("read advertised history size: %w", err)
	}
	// READY already carries the active screen and some recent resident history.
	// Next supplies any older pages that remain in the snapshot.
	fmt.Printf("\nincremental READY: terminal is usable; snapshot contains %d total history rows\n", historyRows)

	pages := 0
	restoredRows := uint(0)
	for {
		advanced, err := decoder.Next()
		if err != nil {
			return fmt.Errorf("decode history page: %w", err)
		}
		if !advanced {
			break // FINISH was authenticated.
		}

		progress, err := decoder.Progress()
		if err != nil {
			return fmt.Errorf("read history progress: %w", err)
		}
		pages++
		restoredRows += progress.Rows
		fmt.Printf(
			"  page %d: restored %d rows (%d pages remain in this screen)\n",
			pages,
			progress.Rows,
			progress.Remaining,
		)
	}

	pageLabel := "pages"
	if pages == 1 {
		pageLabel = "page"
	}
	fmt.Printf("incremental FINISH: added %d older rows from %d %s\n", restoredRows, pages, pageLabel)
	return nil
}

// visibleText formats the terminal as plain text and returns only the final
// active rows. The formatter also includes scrollback, so the helper trims
// that prefix to keep this example's output short. Snapshotting itself is
// independent of the Formatter API.
func visibleText(term *libghostty.Terminal) (string, error) {
	formatter, err := libghostty.NewFormatter(
		term,
		libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
		libghostty.WithFormatterTrim(true),
	)
	if err != nil {
		return "", fmt.Errorf("create formatter: %w", err)
	}
	defer formatter.Close()

	text, err := formatter.FormatString()
	if err != nil {
		return "", fmt.Errorf("format visible terminal: %w", err)
	}
	rows, err := term.Rows()
	if err != nil {
		return "", fmt.Errorf("read terminal rows: %w", err)
	}
	lines := strings.Split(text, "\n")
	if len(lines) > int(rows) {
		lines = lines[len(lines)-int(rows):]
	}
	return strings.Join(lines, "\n"), nil
}

func printIndented(text string) {
	for _, line := range strings.Split(text, "\n") {
		fmt.Printf("  %s\n", line)
	}
}
