package libghostty

import "testing"

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
