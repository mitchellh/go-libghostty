package libghostty

import "testing"

// TestStableModeNames checks that the 1.0 C macro renames preserve the existing
// Go API and each mode's numeric value and ANSI/DEC namespace.
func TestStableModeNames(t *testing.T) {
	tests := []struct {
		name  string
		mode  Mode
		value uint16
		ansi  bool
	}{
		{"KAM", ModeKAM, 2, true},
		{"SRM", ModeSRM, 12, true},
		{"DECCKM", ModeDECCKM, 1, false},
		{"X10Mouse", ModeX10Mouse, 9, false},
		{"LeftRightMargin", ModeLeftRightMargin, 69, false},
		{"NormalMouse", ModeNormalMouse, 1000, false},
		{"ButtonMouse", ModeButtonMouse, 1002, false},
		{"AnyMouse", ModeAnyMouse, 1003, false},
		{"UTF8Mouse", ModeUTF8Mouse, 1005, false},
		{"SGRMouse", ModeSGRMouse, 1006, false},
		{"AltScroll", ModeAltScroll, 1007, false},
		{"URxvtMouse", ModeURxvtMouse, 1015, false},
		{"SGRPixelsMouse", ModeSGRPixelsMouse, 1016, false},
		{"NumlockKeypad", ModeNumlockKeypad, 1035, false},
		{"AltSendsEsc", ModeAltSendsEsc, 1039, false},
		{"ReverseWrapExt", ModeReverseWrapExt, 1045, false},
		{"AltScreenSave", ModeAltScreenSave, 1049, false},
		{"SyncOutput", ModeSyncOutput, 2026, false},
		{"ColorSchemeReport", ModeColorSchemeReport, 2031, false},
		{"VisibilityReport", ModeVisibilityReport, 2033, false},
		{"InBandResize", ModeInBandResize, 2048, false},
		{"PasteEvents", ModePasteEvents, 5522, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.Value(); got != tt.value {
				t.Fatalf("value = %d, want %d", got, tt.value)
			}
			if got := tt.mode.ANSI(); got != tt.ansi {
				t.Fatalf("ANSI = %t, want %t", got, tt.ansi)
			}
		})
	}
}

func TestModeValueANSI(t *testing.T) {
	m := ModeInsert
	if m.Value() != 4 {
		t.Fatalf("expected value 4, got %d", m.Value())
	}
	if !m.ANSI() {
		t.Fatal("expected ANSI mode")
	}
}

func TestModeValueDEC(t *testing.T) {
	m := ModeCursorVisible
	if m.Value() != 25 {
		t.Fatalf("expected value 25, got %d", m.Value())
	}
	if m.ANSI() {
		t.Fatal("expected DEC private mode")
	}
}

func TestModeVisibilityReport(t *testing.T) {
	if got := ModeVisibilityReport.Value(); got != 2033 {
		t.Fatalf("expected value 2033, got %d", got)
	}
	if ModeVisibilityReport.ANSI() {
		t.Fatal("expected DEC private mode")
	}
}

func TestModePasteEvents(t *testing.T) {
	if got := ModePasteEvents.Value(); got != 5522 {
		t.Fatalf("expected value 5522, got %d", got)
	}
	if ModePasteEvents.ANSI() {
		t.Fatal("expected DEC private mode")
	}
}

func TestNewModeAndReportEncode(t *testing.T) {
	mode := NewMode(25, false)
	if mode != ModeCursorVisible {
		t.Fatalf("expected cursor-visible mode, got %d", mode)
	}

	report, err := ModeReportEncode(mode, ModeReportSet)
	if err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[?25;1$y"; string(report) != want {
		t.Fatalf("expected %q, got %q", want, report)
	}
}
