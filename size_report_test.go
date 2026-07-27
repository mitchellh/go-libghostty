package libghostty

import "testing"

func TestSizeReportEncode(t *testing.T) {
	report, err := SizeReportEncode(
		SizeReportCSI18T,
		SizeReportSize{Rows: 24, Columns: 80, CellWidth: 8, CellHeight: 16},
	)
	if err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[8;24;80t"; string(report) != want {
		t.Fatalf("expected %q, got %q", want, report)
	}
}
