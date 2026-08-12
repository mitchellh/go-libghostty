package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// Terminal wraps a Ghostty VT terminal handle.
// It is stateful, not safe for concurrent use, and not reentrant.
// Serialize all calls that touch a terminal, including getters,
// setters, [Terminal.VTWrite], [Terminal.VTWriteUntilGround],
// [Terminal.Resize], [Terminal.Close],
// and any borrowed handles derived from it. Effect callbacks run
// synchronously during VT writes; they must not call [Terminal.VTWrite]
// or [Terminal.VTWriteUntilGround] on the same terminal and should avoid
// blocking for long periods.
// C: GhosttyTerminal
type Terminal struct {
	ptr C.GhosttyTerminal

	// handle is a cgo.Handle pointing back to this Terminal. It is
	// stored as the C-side userdata (GHOSTTY_TERMINAL_OPT_USERDATA)
	// so that C effect trampolines can recover the *Terminal and
	// dispatch to the appropriate Go effect handler.
	handle cgo.Handle

	onWritePty            WritePtyFn
	onBell                BellFn
	onClipboardWrite      ClipboardWriteFn
	onDesktopNotification DesktopNotificationFn
	onTitleChanged        TitleChangedFn
	onPwdChanged          PwdChangedFn
	onProgressReport      ProgressReportFn
	onEnquiry             EnquiryFn
	onXtversion           XtversionFn
	onSize                SizeFn
	onColorScheme         ColorSchemeFn
	onDeviceAttributes    DeviceAttributesFn
	onUnknownSequence     UnknownSequenceFn

	// effectBuf holds C-allocated memory for the most recent response
	// returned by an effect trampoline (e.g. enquiry, xtversion).
	// libghostty copies the data immediately, so a single buffer
	// shared across effects is sufficient.
	effectBuf    unsafe.Pointer
	effectBufLen uintptr
}

// TerminalOption is a functional option for configuring a Terminal.
type TerminalOption func(*TerminalConfig)

// TerminalConfig holds the configuration for creating a Terminal.
// It is populated by functional options such as [WithSize],
// [WithMaxScrollbackBytes], and [WithMaxScrollbackLines].
type TerminalConfig struct {
	// Cols is the terminal width in cells. Must be greater than zero.
	Cols uint16

	// Rows is the terminal height in cells. Must be greater than zero.
	Rows uint16

	// MaxScrollbackBytes is the optional approximate maximum scrollback
	// allocation in bytes. Nil retains libghostty's default.
	MaxScrollbackBytes *uint

	// MaxScrollbackLines is the optional maximum number of physical lines
	// retained in scrollback. Nil retains libghostty's default.
	MaxScrollbackLines *uint

	// ContinuationMaxBytes is the optional maximum number of replay-safe VT
	// continuation bytes retained by the terminal. Nil leaves tracking
	// disabled; zero also explicitly disables tracking.
	ContinuationMaxBytes *uint

	// TitleReport optionally enables replies to CSI 21 t title queries. Nil
	// retains the secure disabled default.
	TitleReport *bool

	// ModeDefaults configure terminal mode reset defaults. Each entry also
	// immediately replaces the current value of that mode.
	ModeDefaults []TerminalModeConfig

	// UnknownMaxBytes is the optional maximum number of bytes retained for an
	// unsupported terminal sequence. Nil leaves capture disabled.
	UnknownMaxBytes *uint

	// TerminfoName is the optional terminfo entry name reported for XTGETTCAP
	// TN queries. Nil leaves the query unanswered; an empty string clears it.
	TerminfoName *string

	// Effect handlers applied after terminal creation.
	onWritePty            WritePtyFn
	onBell                BellFn
	onClipboardWrite      ClipboardWriteFn
	onDesktopNotification DesktopNotificationFn
	onTitleChanged        TitleChangedFn
	onPwdChanged          PwdChangedFn
	onProgressReport      ProgressReportFn
	onEnquiry             EnquiryFn
	onXtversion           XtversionFn
	onSize                SizeFn
	onColorScheme         ColorSchemeFn
	onDeviceAttributes    DeviceAttributesFn
	onUnknownSequence     UnknownSequenceFn
}

// WritePtyFn is called when the terminal writes data back to the pty
// (e.g. query responses). The first parameter is the terminal that
// triggered the effect. The data is only valid for the call duration.
// C: GhosttyTerminalWritePtyFn
type WritePtyFn func(t *Terminal, data []byte)

// BellFn is called when the terminal receives a BEL character (0x07).
// The parameter is the terminal that triggered the effect.
// C: GhosttyTerminalBellFn
type BellFn func(t *Terminal)

// ClipboardLocation identifies the normalized destination for a clipboard
// write. Protocol-specific selectors are converted to one of these values
// before the callback runs.
// C: GhosttyClipboardLocation
type ClipboardLocation int

const (
	// ClipboardLocationStandard identifies the standard system clipboard.
	ClipboardLocationStandard ClipboardLocation = C.GHOSTTY_CLIPBOARD_LOCATION_STANDARD

	// ClipboardLocationSelection identifies the selection clipboard.
	ClipboardLocationSelection ClipboardLocation = C.GHOSTTY_CLIPBOARD_LOCATION_SELECTION

	// ClipboardLocationPrimary identifies the primary selection clipboard.
	ClipboardLocationPrimary ClipboardLocation = C.GHOSTTY_CLIPBOARD_LOCATION_PRIMARY
)

// ClipboardContent is one MIME representation in a clipboard write. Data is
// decoded from its protocol-level encoding and is binary-safe. Empty Data is
// an explicit empty representation; only an empty [ClipboardWrite.Contents]
// requests that the destination be cleared.
// C: GhosttyClipboardContent
type ClipboardContent struct {
	// MIME is the MIME type of this representation.
	MIME string

	// Data is the decoded, binary-safe representation data.
	Data []byte
}

// ClipboardWrite is a semantic, atomic clipboard write. Every entry in
// Contents represents the same logical value and should be committed
// atomically. An empty Contents slice requests that Location be cleared.
// C: GhosttyClipboardWrite
type ClipboardWrite struct {
	// Location is the normalized clipboard destination.
	Location ClipboardLocation

	// Contents contains all MIME representations of the logical value.
	Contents []ClipboardContent
}

// ClipboardWriteResult reports the outcome of a clipboard write callback.
// Protocols without write acknowledgements, including OSC 52 and iTerm2
// OSC 1337 Copy, ignore this result.
// C: GhosttyClipboardWriteResult
type ClipboardWriteResult int

const (
	// ClipboardWriteSuccess means the clipboard write completed successfully.
	ClipboardWriteSuccess ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_SUCCESS

	// ClipboardWriteDenied means policy or the user denied the write.
	ClipboardWriteDenied ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_DENIED

	// ClipboardWriteUnsupported means the destination or a representation is unsupported.
	ClipboardWriteUnsupported ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_UNSUPPORTED

	// ClipboardWriteBusy means the clipboard is temporarily unavailable.
	ClipboardWriteBusy ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_BUSY

	// ClipboardWriteInvalidData means one or more representations contain invalid data.
	ClipboardWriteInvalidData ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_INVALID_DATA

	// ClipboardWriteIOError means the clipboard write failed due to an I/O error.
	ClipboardWriteIOError ClipboardWriteResult = C.GHOSTTY_CLIPBOARD_WRITE_RESULT_IO_ERROR
)

// ClipboardWriteFn is called synchronously for a complete logical clipboard
// write. Protocol details such as selectors, encodings, chunks, and aliases
// have already been normalized. The write and all of its content are copied
// into Go-owned memory before the callback runs and may be retained. Return
// the result of attempting the write.
// C: GhosttyTerminalClipboardWriteFn
type ClipboardWriteFn func(t *Terminal, write ClipboardWrite) ClipboardWriteResult

// TerminalDesktopNotification is a request from the running program to show a
// desktop notification. Title is empty for protocols such as OSC 9 that
// provide only a body. Both strings are copied into Go-owned memory before
// the callback runs and may be retained.
// C: GhosttyTerminalDesktopNotification
type TerminalDesktopNotification struct {
	// Title is the notification title, or empty when the protocol omits it.
	Title string

	// Body is the notification body.
	Body string
}

// DesktopNotificationFn is called synchronously when the terminal receives
// a desktop notification request via OSC 9 or OSC 777.
// C: GhosttyTerminalDesktopNotificationFn
type DesktopNotificationFn func(t *Terminal, notification TerminalDesktopNotification)

// TitleChangedFn is called when the terminal title changes via OSC 0/2.
// The parameter is the terminal that triggered the effect.
// C: GhosttyTerminalTitleChangedFn
type TitleChangedFn func(t *Terminal)

// PwdChangedFn is called when the terminal working directory changes via
// OSC 7, OSC 9, or OSC 1337 CurrentDir. Query [Terminal.Pwd] in the callback
// to read the new value.
// C: GhosttyTerminalPwdChangedFn
type PwdChangedFn func(t *Terminal)

// TerminalProgressState identifies the state of a progress report emitted by
// the running program.
// C: GhosttyTerminalProgressState
type TerminalProgressState int

const (
	// TerminalProgressStateRemove requests removal of any progress indication.
	TerminalProgressStateRemove TerminalProgressState = C.GHOSTTY_TERMINAL_PROGRESS_STATE_REMOVE

	// TerminalProgressStateSet reports determinate progress.
	TerminalProgressStateSet TerminalProgressState = C.GHOSTTY_TERMINAL_PROGRESS_STATE_SET

	// TerminalProgressStateError reports failed progress.
	TerminalProgressStateError TerminalProgressState = C.GHOSTTY_TERMINAL_PROGRESS_STATE_ERROR

	// TerminalProgressStateIndeterminate reports progress without a percentage.
	TerminalProgressStateIndeterminate TerminalProgressState = C.GHOSTTY_TERMINAL_PROGRESS_STATE_INDETERMINATE

	// TerminalProgressStatePause reports paused progress.
	TerminalProgressStatePause TerminalProgressState = C.GHOSTTY_TERMINAL_PROGRESS_STATE_PAUSE
)

// TerminalProgressReport is a progress update emitted by the running program
// via OSC 9;4.
// C: GhosttyTerminalProgressReport
type TerminalProgressReport struct {
	// State is the literal progress state reported by the program.
	State TerminalProgressState

	// Progress is a percentage from 0 through 100, or -1 when omitted.
	Progress int8
}

// ProgressReportFn is called synchronously when the terminal receives a
// progress report via OSC 9;4.
// C: GhosttyTerminalProgressReportFn
type ProgressReportFn func(t *Terminal, report TerminalProgressReport)

// TerminalUnknownSequenceTag identifies the kind of unsupported terminal
// sequence reported by [UnknownSequenceFn]. Additional tags may be added by
// future libghostty versions.
// C: GhosttyTerminalUnknownSequenceTag
type TerminalUnknownSequenceTag int

const (
	// TerminalUnknownSequenceAPC identifies an unsupported Application Program
	// Command sequence.
	TerminalUnknownSequenceAPC TerminalUnknownSequenceTag = C.GHOSTTY_TERMINAL_UNKNOWN_SEQUENCE_APC
)

// TerminalUnknownStringSequence contains one unsupported string sequence.
// Content is copied into Go-owned memory before the callback runs and may be
// retained after the callback returns.
// C: GhosttyTerminalUnknownStringSequence
type TerminalUnknownStringSequence struct {
	// Truncated reports whether Content was shortened by the configured byte
	// limit or by an allocation failure inside libghostty.
	Truncated bool

	// Content contains the bytes between the sequence introducer and
	// terminator. It is binary-safe and does not include either delimiter.
	Content []byte
}

// TerminalUnknownSequence is an unsupported terminal sequence. APC is
// populated when Tag is [TerminalUnknownSequenceAPC]. For tags introduced by
// newer libghostty versions, fields unknown to this binding remain zero.
// C: GhosttyTerminalUnknownSequence
type TerminalUnknownSequence struct {
	// Tag identifies which sequence value is populated.
	Tag TerminalUnknownSequenceTag

	// APC contains the unsupported APC value when Tag is
	// TerminalUnknownSequenceAPC.
	APC TerminalUnknownStringSequence
}

// UnknownSequenceFn is called synchronously for normally terminated terminal
// sequences whose identifier libghostty does not support. Capture must also be
// enabled with [WithUnknownMaxBytes] or [Terminal.SetUnknownMaxBytes].
// C: GhosttyTerminalUnknownSequenceFn
type UnknownSequenceFn func(t *Terminal, sequence TerminalUnknownSequence)

// EnquiryFn is called when the terminal receives ENQ (0x05).
// The first parameter is the terminal that triggered the effect.
// Return the response bytes; nil or empty means no response.
// C: GhosttyTerminalEnquiryFn
type EnquiryFn func(t *Terminal) []byte

// XtversionFn is called for XTVERSION queries (CSI > q).
// The first parameter is the terminal that triggered the effect.
// Return the version string; empty uses the default "libghostty".
// C: GhosttyTerminalXtversionFn
type XtversionFn func(t *Terminal) string

// SizeFn is called for XTWINOPS size queries (CSI 14/16/18 t).
// The first parameter is the terminal that triggered the effect.
// Return the size and true, or zero value and false to ignore the query.
// C: GhosttyTerminalSizeFn
type SizeFn func(t *Terminal) (SizeReportSize, bool)

// ColorSchemeFn is called for color scheme queries (CSI ? 996 n).
// The first parameter is the terminal that triggered the effect.
// Return the scheme and true, or zero value and false to ignore the query.
// C: GhosttyTerminalColorSchemeFn
type ColorSchemeFn func(t *Terminal) (ColorScheme, bool)

// DeviceAttributesFn is called for device attributes queries
// (CSI c / CSI > c / CSI = c). The first parameter is the terminal
// that triggered the effect. Return the attributes and true,
// or zero value and false to ignore the query.
// C: GhosttyTerminalDeviceAttributesFn
type DeviceAttributesFn func(t *Terminal) (DeviceAttributes, bool)

// TerminalModeConfig pairs a terminal mode with a boolean value. It is used
// by the generic terminal option and data APIs for mode configuration and
// queries.
// C: GhosttyTerminalModeConfig
type TerminalModeConfig struct {
	// Mode identifies the terminal mode to configure or query.
	Mode Mode

	// Value is the value to set or the current value returned by a query.
	Value bool
}

// WithSize sets the terminal dimensions in cells.
// Both cols and rows must be greater than zero.
func WithSize(cols, rows uint16) TerminalOption {
	return func(c *TerminalConfig) {
		c.Cols = cols
		c.Rows = rows
	}
}

// WithMaxScrollbackBytes sets the maximum scrollback allocation in bytes.
// The limit is approximate because libghostty prunes at page granularity.
func WithMaxScrollbackBytes(bytes uint) TerminalOption {
	return func(c *TerminalConfig) {
		c.MaxScrollbackBytes = &bytes
	}
}

// WithMaxScrollbackLines sets the maximum number of physical scrollback
// lines. The limit is approximate because libghostty prunes at page
// granularity.
func WithMaxScrollbackLines(lines uint) TerminalOption {
	return func(c *TerminalConfig) {
		c.MaxScrollbackLines = &lines
	}
}

// WithContinuationMaxBytes enables replay-safe VT continuation tracking and
// sets the maximum retained byte count. A zero limit disables tracking.
func WithContinuationMaxBytes(limit uint) TerminalOption {
	return func(c *TerminalConfig) {
		c.ContinuationMaxBytes = &limit
	}
}

// WithTitleReport enables or disables replies to CSI 21 t window-title
// queries. Reporting is disabled by default because a program can otherwise
// query a title it previously set and inject that text into the pty stream.
func WithTitleReport(enabled bool) TerminalOption {
	return func(c *TerminalConfig) {
		c.TitleReport = &enabled
	}
}

// WithModeDefault sets the reset default for mode. Setting a default also
// immediately updates the mode's current value. Modes that represent
// transitions or mirror other terminal state are rejected by NewTerminal.
func WithModeDefault(mode Mode, value bool) TerminalOption {
	return func(c *TerminalConfig) {
		c.ModeDefaults = append(c.ModeDefaults, TerminalModeConfig{
			Mode:  mode,
			Value: value,
		})
	}
}

// WithUnknownMaxBytes sets the maximum content bytes retained for each
// unsupported terminal sequence. A zero limit disables capture.
func WithUnknownMaxBytes(limit uint) TerminalOption {
	return func(c *TerminalConfig) {
		c.UnknownMaxBytes = &limit
	}
}

// WithTerminfoName sets the terminfo entry name reported in response to an
// XTGETTCAP query for TN. The name may contain at most 128 bytes. An empty name
// leaves the query unanswered.
func WithTerminfoName(name string) TerminalOption {
	return func(c *TerminalConfig) {
		c.TerminfoName = &name
	}
}

// WithWritePty registers an effect handler invoked when the terminal
// writes data back to the pty (e.g. query responses). The data slice
// is only valid for the duration of the call.
func WithWritePty(fn WritePtyFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onWritePty = fn
	}
}

// WithBell registers an effect handler invoked when the terminal
// receives a BEL character (0x07).
func WithBell(fn BellFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onBell = fn
	}
}

// WithClipboardWrite registers an effect handler invoked for normalized,
// decoded clipboard writes. Clipboard read requests are never forwarded.
func WithClipboardWrite(fn ClipboardWriteFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onClipboardWrite = fn
	}
}

// WithDesktopNotification registers an effect handler invoked for desktop
// notification requests received via OSC 9 or OSC 777.
func WithDesktopNotification(fn DesktopNotificationFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onDesktopNotification = fn
	}
}

// WithTitleChanged registers an effect handler invoked when the
// terminal title changes via OSC 0 or OSC 2.
func WithTitleChanged(fn TitleChangedFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onTitleChanged = fn
	}
}

// WithPwdChanged registers an effect handler invoked when the terminal
// working directory changes via an OSC sequence.
func WithPwdChanged(fn PwdChangedFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onPwdChanged = fn
	}
}

// WithProgressReport registers an effect handler invoked for progress reports
// received via OSC 9;4.
func WithProgressReport(fn ProgressReportFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onProgressReport = fn
	}
}

// WithUnknownSequence registers a handler for unsupported terminal sequence
// identifiers. Capture must also be enabled with [WithUnknownMaxBytes].
func WithUnknownSequence(fn UnknownSequenceFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onUnknownSequence = fn
	}
}

// WithEnquiry registers an effect handler invoked when the terminal
// receives an ENQ character (0x05). Return the response bytes; nil
// or empty means no response.
func WithEnquiry(fn EnquiryFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onEnquiry = fn
	}
}

// WithXtversion registers an effect handler invoked for XTVERSION
// queries (CSI > q). Return the version string; empty uses the
// default "libghostty" version.
func WithXtversion(fn XtversionFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onXtversion = fn
	}
}

// WithSizeReport registers an effect handler invoked for XTWINOPS
// size queries (CSI 14/16/18 t). Return the size and true, or
// zero value and false to silently ignore the query.
func WithSizeReport(fn SizeFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onSize = fn
	}
}

// WithColorScheme registers an effect handler invoked for color
// scheme queries (CSI ? 996 n). Return the scheme and true, or
// zero value and false to silently ignore the query.
func WithColorScheme(fn ColorSchemeFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onColorScheme = fn
	}
}

// WithDeviceAttributes registers an effect handler invoked for
// device attributes queries (CSI c / CSI > c / CSI = c). Return
// the attributes and true, or zero value and false to silently
// ignore the query.
func WithDeviceAttributes(fn DeviceAttributesFn) TerminalOption {
	return func(c *TerminalConfig) {
		c.onDeviceAttributes = fn
	}
}

// NewTerminal creates a new terminal with the given options.
// WithSize is required; cols and rows must both be greater than zero.
func NewTerminal(opts ...TerminalOption) (*Terminal, error) {
	// Apply defaults and user options.
	cfg := TerminalConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	var cterm C.GhosttyTerminal
	if err := resultError(C.ghostty_terminal_new(
		nil,
		&cterm,
		C.uint16_t(cfg.Cols),
		C.uint16_t(cfg.Rows),
	)); err != nil {
		return nil, err
	}

	// Apply explicitly configured scrollback limits. Omitting either option
	// retains libghostty's constructor default for that limit.
	if cfg.MaxScrollbackBytes != nil {
		maxBytes := C.size_t(*cfg.MaxScrollbackBytes)
		if err := resultError(C.ghostty_terminal_set(
			cterm,
			C.GHOSTTY_TERMINAL_OPT_SCROLLBACK_MAX_BYTES,
			unsafe.Pointer(&maxBytes),
		)); err != nil {
			C.ghostty_terminal_free(cterm)
			return nil, err
		}
	}
	if cfg.MaxScrollbackLines != nil {
		maxLines := C.size_t(*cfg.MaxScrollbackLines)
		if err := resultError(C.ghostty_terminal_set(
			cterm,
			C.GHOSTTY_TERMINAL_OPT_SCROLLBACK_MAX_LINES,
			unsafe.Pointer(&maxLines),
		)); err != nil {
			C.ghostty_terminal_free(cterm)
			return nil, err
		}
	}
	if cfg.ContinuationMaxBytes != nil {
		limit := C.size_t(*cfg.ContinuationMaxBytes)
		if err := resultError(C.ghostty_terminal_set(
			cterm,
			C.GHOSTTY_TERMINAL_OPT_CONTINUATION_MAX_BYTES,
			unsafe.Pointer(&limit),
		)); err != nil {
			C.ghostty_terminal_free(cterm)
			return nil, err
		}
	}

	t := terminalFromC(cterm, cfg)

	// Apply the remaining constructor options through the same typed setters
	// exposed for live terminals. The terminal is already wrapped so failures
	// can use the normal Close path to release both the C terminal and userdata
	// handle.
	if cfg.TitleReport != nil {
		if err := t.SetTitleReport(*cfg.TitleReport); err != nil {
			t.Close()
			return nil, err
		}
	}
	for _, mode := range cfg.ModeDefaults {
		if err := t.SetModeDefault(mode.Mode, mode.Value); err != nil {
			t.Close()
			return nil, err
		}
	}
	if cfg.UnknownMaxBytes != nil {
		if err := t.SetUnknownMaxBytes(*cfg.UnknownMaxBytes); err != nil {
			t.Close()
			return nil, err
		}
	}
	if cfg.TerminfoName != nil {
		if err := t.SetTerminfoName(*cfg.TerminfoName); err != nil {
			t.Close()
			return nil, err
		}
	}

	// Register any effects that were provided via options.
	t.syncEffects()

	return t, nil
}

// terminalFromC wraps a caller-owned C terminal and installs the Go userdata
// handle required by all terminal effect callbacks. Snapshot decoding uses
// this path so restored terminals have the same lifecycle and effect support
// as terminals created by NewTerminal.
func terminalFromC(cterm C.GhosttyTerminal, cfg TerminalConfig) *Terminal {
	t := &Terminal{
		ptr:                   cterm,
		onWritePty:            cfg.onWritePty,
		onBell:                cfg.onBell,
		onClipboardWrite:      cfg.onClipboardWrite,
		onDesktopNotification: cfg.onDesktopNotification,
		onTitleChanged:        cfg.onTitleChanged,
		onPwdChanged:          cfg.onPwdChanged,
		onProgressReport:      cfg.onProgressReport,
		onEnquiry:             cfg.onEnquiry,
		onXtversion:           cfg.onXtversion,
		onSize:                cfg.onSize,
		onColorScheme:         cfg.onColorScheme,
		onDeviceAttributes:    cfg.onDeviceAttributes,
		onUnknownSequence:     cfg.onUnknownSequence,
	}

	// Always set userdata to our handle so trampolines can find us.
	t.handle = cgo.NewHandle(t)
	C.ghostty_terminal_set(
		t.ptr,
		C.GHOSTTY_TERMINAL_OPT_USERDATA,
		handleToPointer(t.handle),
	)
	return t
}

// Close frees the underlying terminal handle and releases the cgo.Handle.
// After this call, the terminal must not be used.
func (t *Terminal) Close() {
	t.handle.Delete()
	C.ghostty_terminal_free(t.ptr)
	if t.effectBuf != nil {
		Free(t.effectBuf, t.effectBufLen)
	}
}

// Reset performs a full terminal reset (RIS).
// All state is reset to initial configuration (modes, scrollback,
// scrolling region, screen contents). Dimensions are preserved.
func (t *Terminal) Reset() {
	C.ghostty_terminal_reset(t.ptr)
}

// Resize changes the terminal dimensions.
// Both cols and rows must be greater than zero. cellWidthPx and
// cellHeightPx specify the pixel dimensions of a single cell, used
// for image protocols and size reports.
func (t *Terminal) Resize(cols, rows uint16, cellWidthPx, cellHeightPx uint32) error {
	return resultError(C.ghostty_terminal_resize(
		t.ptr,
		C.uint16_t(cols),
		C.uint16_t(rows),
		C.uint32_t(cellWidthPx),
		C.uint32_t(cellHeightPx),
	))
}

// TerminalCompressionMode controls how much scrollback compression work a
// call to [Terminal.Compress] performs.
// C: GhosttyTerminalCompressionMode
type TerminalCompressionMode int

const (
	// TerminalCompressionIncremental performs one bounded step suitable for
	// an idle callback.
	TerminalCompressionIncremental TerminalCompressionMode = C.GHOSTTY_TERMINAL_COMPRESSION_MODE_INCREMENTAL

	// TerminalCompressionFull synchronously scans all eligible pages.
	TerminalCompressionFull TerminalCompressionMode = C.GHOSTTY_TERMINAL_COMPRESSION_MODE_FULL
)

// TerminalCompressionResult describes whether a compression pass needs more
// work.
// C: GhosttyTerminalCompressionResult
type TerminalCompressionResult int

const (
	// TerminalCompressionUnsupported means retained-mapping reclamation is
	// unavailable on this target.
	TerminalCompressionUnsupported TerminalCompressionResult = C.GHOSTTY_TERMINAL_COMPRESSION_RESULT_UNSUPPORTED

	// TerminalCompressionPending means another incremental step should run
	// while the terminal remains idle.
	TerminalCompressionPending TerminalCompressionResult = C.GHOSTTY_TERMINAL_COMPRESSION_RESULT_PENDING

	// TerminalCompressionComplete means no continuation is needed until the
	// activity token changes.
	TerminalCompressionComplete TerminalCompressionResult = C.GHOSTTY_TERMINAL_COMPRESSION_RESULT_COMPLETE
)

// TerminalCursorStyle identifies the visual style used when DECSCUSR resets
// the cursor.
// C: GhosttyTerminalCursorStyle
type TerminalCursorStyle int

const (
	// TerminalCursorStyleBar is a vertical bar cursor.
	TerminalCursorStyleBar TerminalCursorStyle = C.GHOSTTY_TERMINAL_CURSOR_STYLE_BAR

	// TerminalCursorStyleBlock is a filled block cursor.
	TerminalCursorStyleBlock TerminalCursorStyle = C.GHOSTTY_TERMINAL_CURSOR_STYLE_BLOCK

	// TerminalCursorStyleUnderline is an underline cursor.
	TerminalCursorStyleUnderline TerminalCursorStyle = C.GHOSTTY_TERMINAL_CURSOR_STYLE_UNDERLINE

	// TerminalCursorStyleBlockHollow is a hollow block cursor.
	TerminalCursorStyleBlockHollow TerminalCursorStyle = C.GHOSTTY_TERMINAL_CURSOR_STYLE_BLOCK_HOLLOW
)

// CompressionActivity returns the opaque scrollback-compression activity
// token. Only equality comparisons between tokens are meaningful.
func (t *Terminal) CompressionActivity() (uint64, error) {
	var activity C.uint64_t
	if err := resultError(C.ghostty_terminal_compression_activity(
		t.ptr,
		&activity,
	)); err != nil {
		return 0, err
	}
	return uint64(activity), nil
}

// Compress performs caller-driven scrollback compression. Calls must be
// serialized with all other access to the terminal.
func (t *Terminal) Compress(mode TerminalCompressionMode) (TerminalCompressionResult, error) {
	var result C.GhosttyTerminalCompressionResult
	if err := resultError(C.ghostty_terminal_compress(
		t.ptr,
		C.GhosttyTerminalCompressionMode(mode),
		&result,
	)); err != nil {
		return 0, err
	}
	return TerminalCompressionResult(result), nil
}

// VTWrite feeds raw VT-encoded bytes through the terminal's parser,
// updating terminal state. Malformed input is handled gracefully and
// will not cause an error. Effect callbacks run synchronously before
// this call returns; they must not call [Terminal.VTWrite] or
// [Terminal.VTWriteUntilGround] on the same terminal.
func (t *Terminal) VTWrite(data []byte) {
	if len(data) == 0 {
		return
	}
	C.ghostty_terminal_vt_write(t.ptr, (*C.uint8_t)(&data[0]), C.size_t(len(data)))
}

// VTWriteUntilGround feeds only the shortest prefix of data needed for the
// terminal's VT parser to return to its ground state. Ground is the stateless
// point between UTF-8 codepoints and VT sequences where callers can safely
// insert out-of-band VT data.
//
// If the parser is already at ground, the method consumes zero bytes and
// leaves data untouched. If all of data is consumed without reaching ground,
// consumed is len(data) and the returned error has [ResultNoValue]. Effect
// callbacks run synchronously for the consumed prefix only and must not call
// either VT write method on the same terminal.
// C: ghostty_terminal_vt_write_until_ground
func (t *Terminal) VTWriteUntilGround(data []byte) (consumed int, err error) {
	var ptr *C.uint8_t
	if len(data) > 0 {
		ptr = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}

	var out C.size_t
	result := C.ghostty_terminal_vt_write_until_ground(
		t.ptr,
		ptr,
		C.size_t(len(data)),
		&out,
	)
	return int(out), resultError(result)
}

// Write implements io.Writer by feeding data through the terminal's
// VT parser. It always consumes all bytes and never returns an error.
func (t *Terminal) Write(p []byte) (int, error) {
	t.VTWrite(p)
	return len(p), nil
}

// ScrollViewportTag describes the scroll behavior.
// C: GhosttyTerminalScrollViewportTag
type ScrollViewportTag int

const (
	// ScrollViewportTop scrolls to the top of scrollback.
	ScrollViewportTop ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_TOP

	// ScrollViewportBottom scrolls to the bottom (active area).
	ScrollViewportBottom ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_BOTTOM

	// ScrollViewportDelta scrolls by a delta amount (up is negative).
	ScrollViewportDelta ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_DELTA

	// ScrollViewportRow scrolls to an absolute row offset in the same
	// coordinate space as [Scrollbar.Offset].
	ScrollViewportRow ScrollViewportTag = C.GHOSTTY_SCROLL_VIEWPORT_ROW
)

// ScrollViewport scrolls the terminal viewport to the top of scrollback.
func (t *Terminal) ScrollViewportTop() {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_TOP
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// ScrollViewportBottom scrolls the terminal viewport to the bottom
// (active area).
func (t *Terminal) ScrollViewportBottom() {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_BOTTOM
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// ScrollViewportDelta scrolls the terminal viewport by the given delta
// (negative for up, positive for down).
func (t *Terminal) ScrollViewportDelta(delta int) {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_DELTA
	// Set the delta in the value union. The delta field is at offset 0.
	*(*C.intptr_t)(unsafe.Pointer(&sv.value[0])) = C.intptr_t(delta)
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// ScrollViewportRow scrolls the viewport to an absolute row offset from the
// top of the scrollable area. The value is clamped to the available range.
func (t *Terminal) ScrollViewportRow(row uint) {
	var sv C.GhosttyTerminalScrollViewport
	sv.tag = C.GHOSTTY_SCROLL_VIEWPORT_ROW
	// Set the row in the value union. The row field is at offset 0.
	*(*C.size_t)(unsafe.Pointer(&sv.value[0])) = C.size_t(row)
	C.ghostty_terminal_scroll_viewport(t.ptr, sv)
}

// TerminalScreen identifies which screen buffer is active.
// C: GhosttyTerminalScreen
type TerminalScreen int

const (
	// ScreenPrimary is the primary (normal) screen.
	ScreenPrimary TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_PRIMARY

	// ScreenAlternate is the alternate screen.
	ScreenAlternate TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_ALTERNATE
)

// Scrollbar holds the scrollbar state for the terminal viewport.
// C: GhosttyTerminalScrollbar
type Scrollbar struct {
	// Total is the total size of the scrollable area in rows.
	Total uint64

	// Offset is the offset into the total area that the viewport is at.
	Offset uint64

	// Len is the length of the visible area in rows.
	Len uint64
}

// KittyGraphics returns the Kitty graphics image storage for the
// terminal's active screen. The returned handle is borrowed from the
// terminal and remains valid until the next mutating call (for example
// [Terminal.VTWrite] or [Terminal.Reset]). Serialize use of the
// returned handle with mutations of the terminal.
func (t *Terminal) KittyGraphics() (*KittyGraphics, error) {
	var ptr C.GhosttyKittyGraphics
	if err := resultError(C.ghostty_terminal_get(
		t.ptr,
		C.GHOSTTY_TERMINAL_DATA_KITTY_GRAPHICS,
		unsafe.Pointer(&ptr),
	)); err != nil {
		return nil, err
	}
	return &KittyGraphics{ptr: ptr}, nil
}

// GridRef resolves a point in the terminal grid to a grid reference.
// The returned GridRef is borrowed from terminal internals and should
// be used immediately; any later terminal operation may invalidate it.
//
// Lookups using PointTagActive and PointTagViewport are fast.
// PointTagScreen and PointTagHistory may be expensive for large
// scrollback buffers.
func (t *Terminal) GridRef(point Point) (*GridRef, error) {
	ref := initCGridRef()
	if err := resultError(C.ghostty_terminal_grid_ref(
		t.ptr,
		point.toC(),
		&ref,
	)); err != nil {
		return nil, err
	}
	return &GridRef{ref: ref}, nil
}

// TrackGridRef creates an owned tracked grid reference for a terminal
// point. The returned reference follows the referenced cell as normal
// screen operations update the terminal page list. Close the returned
// reference when finished.
//
// The tracked reference is attached to the terminal screen/page-list that
// is active when this method is called. If the terminal is closed first,
// the tracked handle remains valid only for tracked-grid-ref APIs: it
// reports no value and can still be closed.
func (t *Terminal) TrackGridRef(point Point) (*TrackedGridRef, error) {
	var ptr C.GhosttyTrackedGridRef
	if err := resultError(C.ghostty_terminal_grid_ref_track(
		t.ptr,
		point.toC(),
		&ptr,
	)); err != nil {
		return nil, err
	}
	return &TrackedGridRef{ptr: ptr}, nil
}

// PointFromGridRef converts a grid reference back to a point in the
// requested coordinate system. The grid reference must come from the same
// terminal and must still be valid. If the reference cannot be represented
// in the requested coordinate system, this returns an error with
// ResultNoValue.
func (t *Terminal) PointFromGridRef(ref *GridRef, tag PointTag) (Point, error) {
	var coord C.GhosttyPointCoordinate
	if err := resultError(C.ghostty_terminal_point_from_grid_ref(
		t.ptr,
		&ref.ref,
		C.GhosttyPointTag(tag),
		&coord,
	)); err != nil {
		return Point{}, err
	}
	return pointFromC(tag, coord), nil
}

// handleToPointer converts a cgo.Handle (uintptr) to unsafe.Pointer
// for passing as C userdata. The handle is an opaque integer, not a
// real Go pointer, so we suppress checkptr which would otherwise
// reject it under -race.
//
//go:nocheckptr
func handleToPointer(h cgo.Handle) unsafe.Pointer {
	return unsafe.Pointer(h)
}
