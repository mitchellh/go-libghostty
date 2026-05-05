package libghostty

import "testing"

func TestPasteIsSafe(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{name: "empty", data: "", want: true},
		{name: "plain", data: "hello", want: true},
		{name: "newline", data: "hello\nworld", want: false},
		{name: "bracketed paste end", data: "hello\x1b[201~world", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PasteIsSafe([]byte(tt.data)); got != tt.want {
				t.Fatalf("PasteIsSafe(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestPasteEncodeLegacy(t *testing.T) {
	in := []byte("hello\nworld\x1b!")

	out, err := PasteEncode(in, false)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != "hello\rworld !" {
		t.Fatalf("PasteEncode legacy = %q, want %q", out, "hello\rworld !")
	}
	if string(in) != "hello\nworld\x1b!" {
		t.Fatalf("PasteEncode mutated input to %q", in)
	}
}

func TestPasteEncodeBracketed(t *testing.T) {
	out, err := PasteEncode([]byte("hello\nworld"), true)
	if err != nil {
		t.Fatal(err)
	}

	want := "\x1b[200~hello\nworld\x1b[201~"
	if string(out) != want {
		t.Fatalf("PasteEncode bracketed = %q, want %q", out, want)
	}
}

func TestPasteEncodeEmpty(t *testing.T) {
	out, err := PasteEncode(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatalf("PasteEncode empty = %q, want nil", out)
	}
}
