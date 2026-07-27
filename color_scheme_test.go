package libghostty

import "testing"

func TestColorSchemeReportEncode(t *testing.T) {
	got, err := ColorSchemeReportEncode(ColorSchemeDark)
	if err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[?997;1n"; string(got) != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
