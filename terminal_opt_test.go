package libghostty

import (
	"bytes"
	"testing"
)

func TestTerminalSetTitle(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetTitle("my terminal"); err != nil {
		t.Fatal(err)
	}

	// Clear the title.
	if err := term.SetTitle(""); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetPwd(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if err := term.SetPwd("/tmp"); err != nil {
		t.Fatal(err)
	}

	if err := term.SetPwd(""); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetColors(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	white := &ColorRGB{R: 255, G: 255, B: 255}
	black := &ColorRGB{R: 0, G: 0, B: 0}

	if err := term.SetColorForeground(white); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorBackground(black); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorCursor(white); err != nil {
		t.Fatal(err)
	}

	// Clear colors.
	if err := term.SetColorForeground(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorBackground(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetColorCursor(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetAPCMaxBytes(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	limit := uint(1024)
	if err := term.SetAPCMaxBytes(&limit); err != nil {
		t.Fatal(err)
	}
	if err := term.SetAPCMaxBytes(nil); err != nil {
		t.Fatal(err)
	}

	kittyLimit := uint(512)
	if err := term.SetAPCMaxBytesKitty(&kittyLimit); err != nil {
		t.Fatal(err)
	}
	if err := term.SetAPCMaxBytesKitty(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalKittyClipboardWriteMaxBytes(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithKittyClipboardWriteMaxBytes(4),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	limit, err := term.KittyClipboardWriteMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	if limit != 4 {
		t.Fatalf("expected constructor limit 4, got %d", limit)
	}

	if err := term.SetKittyClipboardWriteMaxBytes(nil); err != nil {
		t.Fatal(err)
	}
	limit, err = term.KittyClipboardWriteMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	const defaultLimit = 64 * 1024 * 1024
	if limit != defaultLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLimit, limit)
	}
}

func TestTerminalWithBell(t *testing.T) {
	var bellCount int
	term, err := NewTerminal(WithSize(80, 24), WithBell(func(_ *Terminal) {
		bellCount++
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// BEL character should trigger the callback.
	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected 1 bell, got %d", bellCount)
	}

	// Multiple BELs.
	term.VTWrite([]byte("\x07\x07"))
	if bellCount != 3 {
		t.Fatalf("expected 3 bells, got %d", bellCount)
	}
}

func TestTerminalSetEffectBell(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	var bellCount int
	term.SetEffectBell(func(_ *Terminal) {
		bellCount++
	})

	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected 1 bell, got %d", bellCount)
	}

	// Clear the callback; bell should no longer fire.
	term.SetEffectBell(nil)
	term.VTWrite([]byte("\x07"))
	if bellCount != 1 {
		t.Fatalf("expected still 1 bell after clearing, got %d", bellCount)
	}
}

func TestTerminalWithClipboardWrite(t *testing.T) {
	var writes []ClipboardWrite
	term, err := NewTerminal(
		WithSize(80, 24),
		WithClipboardWrite(func(_ *Terminal, write ClipboardWrite) ClipboardWriteReply {
			writes = append(writes, write)
			return ClipboardWriteReply{Result: ClipboardWriteDenied}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 52 is decoded before the effect runs, including binary data and
	// sequences split across multiple writes.
	term.VTWrite([]byte("\x1b]52;c;aGVs"))
	term.VTWrite([]byte("bG8Ad29ybGQ=\x1b\\"))
	if len(writes) != 1 {
		t.Fatalf("expected 1 clipboard write, got %d", len(writes))
	}
	if got := writes[0].Location; got != ClipboardLocationStandard {
		t.Fatalf("expected standard clipboard, got %d", got)
	}
	if got := len(writes[0].Contents); got != 1 {
		t.Fatalf("expected 1 clipboard representation, got %d", got)
	}
	content := writes[0].Contents[0]
	if got := content.MIME; got != "text/plain" {
		t.Fatalf("expected text/plain MIME type, got %q", got)
	}
	if want := []byte("hello\x00world"); !bytes.Equal(content.Data, want) {
		t.Fatalf("expected clipboard data %q, got %q", want, content.Data)
	}

	// An empty content list is a clear, while the normalized destination is
	// still preserved.
	term.VTWrite([]byte("\x1b]52;s;\x1b\\"))
	if len(writes) != 2 {
		t.Fatalf("expected 2 clipboard writes, got %d", len(writes))
	}
	if got := writes[1].Location; got != ClipboardLocationSelection {
		t.Fatalf("expected selection clipboard, got %d", got)
	}
	if got := len(writes[1].Contents); got != 0 {
		t.Fatalf("expected a clipboard clear, got %d representations", got)
	}

	// iTerm2 Copy uses the same protocol-neutral callback shape.
	term.VTWrite([]byte("\x1b]1337;Copy=:aVRlcm0=\x1b\\"))
	if len(writes) != 3 {
		t.Fatalf("expected 3 clipboard writes, got %d", len(writes))
	}
	if got := writes[2].Contents[0].Data; !bytes.Equal(got, []byte("iTerm")) {
		t.Fatalf("expected decoded iTerm2 data, got %q", got)
	}

	// Clipboard reads and malformed payloads are intentionally ignored.
	term.VTWrite([]byte("\x1b]52;c;?\x1b\\"))
	term.VTWrite([]byte("\x1b]52;c;%%%\x1b\\"))
	if len(writes) != 3 {
		t.Fatalf("expected ignored reads and malformed data, got %d writes", len(writes))
	}

	// Callback data is copied into Go memory and remains valid after later
	// terminal writes have reused libghostty's borrowed descriptors.
	if want := []byte("hello\x00world"); !bytes.Equal(writes[0].Contents[0].Data, want) {
		t.Fatalf("expected retained clipboard data %q, got %q", want, writes[0].Contents[0].Data)
	}
}

func TestTerminalSetEffectClipboardWrite(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	var writes int
	term.SetEffectClipboardWrite(func(_ *Terminal, write ClipboardWrite) ClipboardWriteReply {
		writes++
		if write.Location != ClipboardLocationPrimary {
			t.Errorf("expected primary clipboard, got %d", write.Location)
		}
		return ClipboardWriteReply{Result: ClipboardWriteSuccess}
	})

	term.VTWrite([]byte("\x1b]52;p;eA==\x1b\\"))
	if writes != 1 {
		t.Fatalf("expected 1 clipboard write, got %d", writes)
	}

	// Clearing the callback takes effect immediately.
	term.SetEffectClipboardWrite(nil)
	term.VTWrite([]byte("\x1b]52;p;eA==\x1b\\"))
	if writes != 1 {
		t.Fatalf("expected still 1 clipboard write after clearing, got %d", writes)
	}
}

func TestTerminalClipboardWriteKittyReplyAndGrant(t *testing.T) {
	var output []byte
	var writes []ClipboardWrite
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output = append(output, data...)
		}),
		WithClipboardWrite(func(_ *Terminal, write ClipboardWrite) ClipboardWriteReply {
			writes = append(writes, write)
			return ClipboardWriteReply{
				Result:   ClipboardWriteSuccess,
				Remember: true,
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// A password makes the request rememberable. Replying successfully with
	// Remember records a grant for the next matching request.
	term.VTWrite([]byte("\x1b]5522;type=write:id=g1:pw=c2VjcmV0:name=YXBw\x1b\\"))
	term.VTWrite([]byte("\x1b]5522;type=wdata\x1b\\"))
	term.VTWrite([]byte("\x1b]5522;type=write:id=g2:pw=c2VjcmV0:name=YXBw\x1b\\"))
	term.VTWrite([]byte("\x1b]5522;type=wdata\x1b\\"))

	if len(writes) != 2 {
		t.Fatalf("expected 2 clipboard writes, got %d", len(writes))
	}
	if got := writes[0].Name; got != "app" {
		t.Fatalf("expected program name app, got %q", got)
	}
	if writes[0].Granted {
		t.Fatal("expected first request not to be granted")
	}
	if !writes[0].CanRemember {
		t.Fatal("expected password request to be rememberable")
	}
	if !writes[1].Granted {
		t.Fatal("expected remembered request to be granted")
	}
	want := "\x1b]5522;type=write:status=DONE:id=g1\x1b\\" +
		"\x1b]5522;type=write:status=DONE:id=g2\x1b\\"
	if got := string(output); got != want {
		t.Fatalf("expected Kitty write replies %q, got %q", want, got)
	}
}

func TestTerminalClipboardReadReplies(t *testing.T) {
	var output []byte
	var reads []ClipboardRead
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			output = append(output, data...)
		}),
		WithClipboardRead(func(_ *Terminal, read ClipboardRead) ClipboardReadReply {
			reads = append(reads, read)
			if read.List {
				return ClipboardReadReply{
					Result:    ClipboardReadSuccess,
					Available: []string{"text/plain", "image/png"},
				}
			}
			return ClipboardReadReply{
				Result: ClipboardReadSuccess,
				Contents: []ClipboardContent{{
					MIME: "text/plain",
					Data: []byte("hello"),
				}},
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 52 asks for one text representation and preserves the request's BEL
	// terminator in the encoded reply.
	term.VTWrite([]byte("\x1b]52;p;?\x07"))
	if len(reads) != 1 {
		t.Fatalf("expected 1 clipboard read, got %d", len(reads))
	}
	read := reads[0]
	if read.Location != ClipboardLocationPrimary {
		t.Fatalf("expected primary clipboard, got %d", read.Location)
	}
	if len(read.MIMEs) != 1 || read.MIMEs[0] != "text/plain" {
		t.Fatalf("expected text/plain request, got %q", read.MIMEs)
	}
	if read.List || read.Name != "" || read.Granted || read.CanRemember {
		t.Fatalf("unexpected OSC 52 metadata: %+v", read)
	}
	if want := "\x1b]52;p;aGVsbG8=\x07"; string(output) != want {
		t.Fatalf("expected OSC 52 reply %q, got %q", want, output)
	}

	// Kitty's special "." MIME asks for the available-target listing. This
	// exercises the separate Available array in ClipboardReadReply.
	output = output[:0]
	term.VTWrite([]byte("\x1b]5522;type=read:id=list:name=YXBw;Lg==\x1b\\"))
	if len(reads) != 2 {
		t.Fatalf("expected 2 clipboard reads, got %d", len(reads))
	}
	read = reads[1]
	if !read.List || len(read.MIMEs) != 0 || read.Name != "app" {
		t.Fatalf("unexpected Kitty list request: %+v", read)
	}
	if !bytes.Contains(output, []byte(";dGV4dC9wbGFpbiBpbWFnZS9wbmcK\x1b\\")) {
		t.Fatalf("expected encoded MIME listing, got %q", output)
	}

	term.SetEffectClipboardRead(nil)
	output = output[:0]
	term.VTWrite([]byte("\x1b]52;c;?\x1b\\"))
	if len(reads) != 2 || len(output) != 0 {
		t.Fatalf("expected cleared callback to ignore OSC 52 read, reads=%d output=%q", len(reads), output)
	}
}

func TestTerminalDesktopNotificationEffect(t *testing.T) {
	var notifications []TerminalDesktopNotification
	term, err := NewTerminal(
		WithSize(80, 24),
		WithDesktopNotification(func(_ *Terminal, notification TerminalDesktopNotification) {
			notifications = append(notifications, notification)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 777 preserves its separate title and body, including when the
	// sequence arrives across multiple VT writes.
	term.VTWrite([]byte("\x1b]777;notify;Codex;"))
	if len(notifications) != 0 {
		t.Fatalf("expected incomplete notification to remain buffered, got %d", len(notifications))
	}
	term.VTWrite([]byte("Needs attention\x1b\\"))
	if len(notifications) != 1 {
		t.Fatalf("expected 1 desktop notification, got %d", len(notifications))
	}
	if got := notifications[0]; got.Title != "Codex" || got.Body != "Needs attention" {
		t.Fatalf("unexpected desktop notification: %+v", got)
	}

	// OSC 9 has no title and preserves its body.
	term.VTWrite([]byte("\x1b]9;Build complete\x07"))
	if len(notifications) != 2 {
		t.Fatalf("expected 2 desktop notifications, got %d", len(notifications))
	}
	if got := notifications[1]; got.Title != "" || got.Body != "Build complete" {
		t.Fatalf("unexpected OSC 9 desktop notification: %+v", got)
	}

	// Callback strings are copied and remain valid after the parser reuses its
	// borrowed input buffer for later terminal writes.
	if got := notifications[0]; got.Title != "Codex" || got.Body != "Needs attention" {
		t.Fatalf("expected retained desktop notification, got %+v", got)
	}

	term.SetEffectDesktopNotification(nil)
	term.VTWrite([]byte("\x1b]9;Ignored\x07"))
	if len(notifications) != 2 {
		t.Fatalf("expected callback to remain cleared, got %d notifications", len(notifications))
	}
}

func TestTerminalProgressReportEffect(t *testing.T) {
	var reports []TerminalProgressReport
	term, err := NewTerminal(
		WithSize(80, 24),
		WithProgressReport(func(_ *Terminal, report TerminalProgressReport) {
			reports = append(reports, report)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	tests := []struct {
		sequence string
		state    TerminalProgressState
		progress int8
	}{
		{"\x1b]9;4;0;\x1b\\", TerminalProgressStateRemove, -1},
		{"\x1b]9;4;1;42\x07", TerminalProgressStateSet, 42},
		{"\x1b]9;4;2;7\x1b\\", TerminalProgressStateError, 7},
		{"\x1b]9;4;3\x1b\\", TerminalProgressStateIndeterminate, -1},
		{"\x1b]9;4;4;75\x1b\\", TerminalProgressStatePause, 75},
	}
	for i, test := range tests {
		midpoint := len(test.sequence) / 2
		term.VTWrite([]byte(test.sequence[:midpoint]))
		if len(reports) != i {
			t.Fatalf("expected split report %d to remain buffered, got %d reports", i, len(reports))
		}
		term.VTWrite([]byte(test.sequence[midpoint:]))
		if len(reports) != i+1 {
			t.Fatalf("expected report %d, got %d reports", i, len(reports))
		}
		if got := reports[i]; got.State != test.state || got.Progress != test.progress {
			t.Fatalf("unexpected progress report %d: %+v", i, got)
		}
	}

	term.SetEffectProgressReport(nil)
	term.VTWrite([]byte("\x1b]9;4;1;90\x1b\\"))
	if len(reports) != len(tests) {
		t.Fatalf("expected callback to remain cleared, got %d reports", len(reports))
	}
}

func TestTerminalWithWritePty(t *testing.T) {
	var received []byte
	term, err := NewTerminal(WithSize(80, 24), WithWritePty(func(_ *Terminal, data []byte) {
		received = append(received, data...)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// DA1 query should produce a response via write_pty.
	term.VTWrite([]byte("\x1b[c"))
	if len(received) == 0 {
		t.Fatal("expected write_pty data from DA1 query")
	}
}

func TestTerminalWritePtyDECRQM(t *testing.T) {
	var received []byte
	term, err := NewTerminal(WithSize(80, 24), WithWritePty(func(_ *Terminal, data []byte) {
		received = append(received, data...)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// DEC private mode query (DECAWM, set by default).
	term.VTWrite([]byte("\x1b[?7$p"))
	if want := []byte("\x1b[?7;1$y"); !bytes.Equal(received, want) {
		t.Fatalf("expected DEC mode report %q, got %q", want, received)
	}

	// ANSI mode query (IRM, after enabling it).
	received = nil
	term.VTWrite([]byte("\x1b[4h\x1b[4$p"))
	if want := []byte("\x1b[4;1$y"); !bytes.Equal(received, want) {
		t.Fatalf("expected ANSI mode report %q, got %q", want, received)
	}
}

func TestTerminalWithTitleChanged(t *testing.T) {
	var titleChanged int
	term, err := NewTerminal(WithSize(80, 24), WithTitleChanged(func(_ *Terminal) {
		titleChanged++
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// OSC 2 sets the title.
	term.VTWrite([]byte("\x1b]2;hello\x07"))
	if titleChanged != 1 {
		t.Fatalf("expected 1 title change, got %d", titleChanged)
	}
}

func TestTerminalWithEnquiry(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
		WithEnquiry(func(_ *Terminal) []byte {
			return []byte("hello")
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// ENQ character should trigger enquiry and write response via pty.
	term.VTWrite([]byte("\x05"))
	if string(received) != "hello" {
		t.Fatalf("expected enquiry response %q, got %q", "hello", string(received))
	}
}

func TestTerminalWithXtversion(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
		WithXtversion(func(_ *Terminal) string {
			return "myterm 1.0"
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// XTVERSION query: CSI > q
	term.VTWrite([]byte("\x1b[>q"))
	// Response should contain our version string in a DCS sequence.
	if len(received) == 0 {
		t.Fatal("expected xtversion response")
	}
	resp := string(received)
	if !contains(resp, "myterm 1.0") {
		t.Fatalf("expected response to contain %q, got %q", "myterm 1.0", resp)
	}
}

// contains reports whether s contains substr.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTerminalSetColorPalette(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Set a custom palette.
	var palette Palette
	for i := range palette {
		palette[i] = ColorRGB{R: uint8(i), G: uint8(i), B: uint8(i)}
	}
	if err := term.SetColorPalette(&palette); err != nil {
		t.Fatal(err)
	}

	// Reset to default.
	if err := term.SetColorPalette(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalSetScrollbackLimits(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithMaxScrollbackBytes(4096),
		WithMaxScrollbackLines(100),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	bytes, err := term.ScrollbackMaxBytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes == nil || *bytes != 4096 {
		t.Fatalf("expected 4096-byte limit, got %v", bytes)
	}

	lines, err := term.ScrollbackMaxLines()
	if err != nil {
		t.Fatal(err)
	}
	if lines == nil || *lines != 100 {
		t.Fatalf("expected 100-line limit, got %v", lines)
	}

	if err := term.SetScrollbackMaxBytes(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetScrollbackMaxLines(nil); err != nil {
		t.Fatal(err)
	}
	if bytes, err := term.ScrollbackMaxBytes(); err != nil {
		t.Fatal(err)
	} else if bytes != nil {
		t.Fatalf("expected unlimited byte limit, got %v", *bytes)
	}
	if lines, err := term.ScrollbackMaxLines(); err != nil {
		t.Fatal(err)
	} else if lines != nil {
		t.Fatalf("expected unlimited line limit, got %v", *lines)
	}
}

func TestTerminalAdditionalOptions(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	style := TerminalCursorStyleBar
	blink := true
	if err := term.SetDefaultCursorStyle(&style); err != nil {
		t.Fatal(err)
	}
	if err := term.SetDefaultCursorBlink(&blink); err != nil {
		t.Fatal(err)
	}
	if err := term.SetGlyphProtocol(true); err != nil {
		t.Fatal(err)
	}
	if err := term.SetDefaultCursorStyle(nil); err != nil {
		t.Fatal(err)
	}
	if err := term.SetDefaultCursorBlink(nil); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalPwdChangedEffect(t *testing.T) {
	var calls int
	var pwd string
	term, err := NewTerminal(
		WithSize(80, 24),
		WithPwdChanged(func(term *Terminal) {
			calls++
			pwd, _ = term.Pwd()
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("\x1b]7;file:///tmp\x07"))
	if calls != 1 {
		t.Fatalf("expected one pwd-changed callback, got %d", calls)
	}
	if pwd != "file:///tmp" {
		t.Fatalf("expected pwd %q, got %q", "file:///tmp", pwd)
	}

	term.SetEffectPwdChanged(nil)
	term.VTWrite([]byte("\x1b]7;file:///var\x07"))
	if calls != 1 {
		t.Fatalf("expected callback to remain cleared, got %d calls", calls)
	}
}

func TestTerminalCompression(t *testing.T) {
	term, err := NewTerminal(WithSize(80, 24), WithMaxScrollbackLines(100))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	if _, err := term.CompressionActivity(); err != nil {
		t.Fatal(err)
	}
	result, err := term.Compress(TerminalCompressionIncremental)
	if err != nil {
		t.Fatal(err)
	}
	switch result {
	case TerminalCompressionUnsupported,
		TerminalCompressionPending,
		TerminalCompressionComplete:
	default:
		t.Fatalf("unexpected compression result %d", result)
	}
}

func TestTerminalModeDefault(t *testing.T) {
	term, err := NewTerminal(
		WithSize(80, 24),
		WithModeDefault(ModeGraphemeCluster, true),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	value, err := term.Mode(ModeGraphemeCluster)
	if err != nil {
		t.Fatal(err)
	}
	if !value {
		t.Fatal("expected configured mode default to update the current value")
	}

	if err := term.SetMode(ModeGraphemeCluster, false); err != nil {
		t.Fatal(err)
	}
	term.Reset()
	value, err = term.Mode(ModeGraphemeCluster)
	if err != nil {
		t.Fatal(err)
	}
	if !value {
		t.Fatal("expected reset to restore the configured mode default")
	}

	if err := term.SetModeDefault(ModeGraphemeCluster, false); err != nil {
		t.Fatal(err)
	}
	value, err = term.Mode(ModeGraphemeCluster)
	if err != nil {
		t.Fatal(err)
	}
	if value {
		t.Fatal("expected SetModeDefault to update the current value")
	}

	if err := term.SetModeDefault(ModeAltScreen, true); err == nil {
		t.Fatal("expected a transition mode to reject default configuration")
	}
}

func TestTerminalTitleReport(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithTitleReport(true),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	term.VTWrite([]byte("\x1b]2;safe title\x1b\\"))
	term.VTWrite([]byte("\x1b[21t"))
	if want := []byte("\x1b]lsafe title\x1b\\"); !bytes.Equal(received, want) {
		t.Fatalf("expected title report %q, got %q", want, received)
	}

	received = nil
	if err := term.SetTitleReport(false); err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte("\x1b[21t"))
	if received != nil {
		t.Fatalf("expected disabled title report to be ignored, got %q", received)
	}
}

func TestTerminalTerminfoName(t *testing.T) {
	var received []byte
	term, err := NewTerminal(
		WithSize(80, 24),
		WithTerminfoName("xterm"),
		WithWritePty(func(_ *Terminal, data []byte) {
			received = append(received, data...)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	const query = "\x1bP+q544E\x1b\\"
	term.VTWrite([]byte(query))
	if want := []byte("\x1bP1+r544E=787465726D\x1b\\"); !bytes.Equal(received, want) {
		t.Fatalf("expected terminfo response %q, got %q", want, received)
	}

	received = nil
	if err := term.SetTerminfoName(""); err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte(query))
	if received != nil {
		t.Fatalf("expected cleared terminfo name to leave TN unanswered, got %q", received)
	}

	if err := term.SetTerminfoName(string(make([]byte, 129))); err == nil {
		t.Fatal("expected a terminfo name longer than 128 bytes to be rejected")
	}
}

func TestTerminalUnknownSequenceEffect(t *testing.T) {
	var sequences []TerminalUnknownSequence
	term, err := NewTerminal(
		WithSize(80, 24),
		WithUnknownMaxBytes(8),
		WithUnknownSequence(func(_ *Terminal, sequence TerminalUnknownSequence) {
			sequences = append(sequences, sequence)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	// Capture persists across writes and excludes the APC delimiters.
	term.VTWrite([]byte("\x1b_abc;"))
	if len(sequences) != 0 {
		t.Fatalf("expected split APC to remain buffered, got %d callbacks", len(sequences))
	}
	term.VTWrite([]byte("xy\x1b\\"))
	if len(sequences) != 1 {
		t.Fatalf("expected one unknown-sequence callback, got %d", len(sequences))
	}
	if got := sequences[0]; got.Tag != TerminalUnknownSequenceAPC ||
		got.APC.Truncated || !bytes.Equal(got.APC.Content, []byte("abc;xy")) {
		t.Fatalf("unexpected APC callback: %+v", got)
	}

	term.VTWrite([]byte("\x1b_abcdefghijkl\x1b\\"))
	if len(sequences) != 2 {
		t.Fatalf("expected two unknown-sequence callbacks, got %d", len(sequences))
	}
	if got := sequences[1]; !got.APC.Truncated ||
		!bytes.Equal(got.APC.Content, []byte("abcdefgh")) {
		t.Fatalf("unexpected truncated APC callback: %+v", got)
	}
	if !bytes.Equal(sequences[0].APC.Content, []byte("abc;xy")) {
		t.Fatalf("expected callback content to remain Go-owned, got %q", sequences[0].APC.Content)
	}

	term.SetEffectUnknownSequence(nil)
	term.VTWrite([]byte("\x1b_ignored\x1b\\"))
	if len(sequences) != 2 {
		t.Fatalf("expected cleared callback to remain inactive, got %d callbacks", len(sequences))
	}

	term.SetEffectUnknownSequence(func(_ *Terminal, sequence TerminalUnknownSequence) {
		sequences = append(sequences, sequence)
	})
	if err := term.SetUnknownMaxBytes(0); err != nil {
		t.Fatal(err)
	}
	term.VTWrite([]byte("\x1b_disabled\x1b\\"))
	if len(sequences) != 2 {
		t.Fatalf("expected zero limit to disable capture, got %d callbacks", len(sequences))
	}
}

func TestTerminalResizePullScrollback(t *testing.T) {
	// fill writes enough lines to push content into scrollback and leave
	// the cursor on the bottom row of a 5-row terminal.
	fill := func(term *Terminal) {
		for i := 0; i < 10; i++ {
			term.VTWrite([]byte("line\r\n"))
		}
	}

	cases := []struct {
		name  string
		opts  []TerminalOption
		wantY uint16
	}{
		// Growing rows pulls three scrollback rows into view, so the
		// cursor follows its content down the screen.
		{"default", nil, 7},
		{"enabled", []TerminalOption{WithResizePullScrollback(true)}, 7},
		// Growing rows appends blank rows at the bottom instead, so the
		// cursor stays on its original row.
		{"disabled", []TerminalOption{WithResizePullScrollback(false)}, 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := append([]TerminalOption{WithSize(10, 5)}, tc.opts...)
			term, err := NewTerminal(opts...)
			if err != nil {
				t.Fatal(err)
			}
			defer term.Close()

			fill(term)
			if err := term.Resize(10, 8, 8, 16); err != nil {
				t.Fatal(err)
			}
			y, err := term.CursorY()
			if err != nil {
				t.Fatal(err)
			}
			if y != tc.wantY {
				t.Fatalf("expected cursor y %d, got %d", tc.wantY, y)
			}
		})
	}

	// The setter can toggle the behavior on a live terminal.
	term, err := NewTerminal(WithSize(10, 5))
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	if err := term.SetResizePullScrollback(false); err != nil {
		t.Fatal(err)
	}
	fill(term)
	if err := term.Resize(10, 8, 8, 16); err != nil {
		t.Fatal(err)
	}
	if y, err := term.CursorY(); err != nil {
		t.Fatal(err)
	} else if y != 4 {
		t.Fatalf("expected cursor y 4, got %d", y)
	}
}
