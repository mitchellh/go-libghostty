package libghostty

import (
	"slices"
	"testing"
)

func TestSGRParser(t *testing.T) {
	parser, err := NewSGRParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	if err := parser.SetParams([]uint16{1, 31}, nil); err != nil {
		t.Fatal(err)
	}
	attr, ok := parser.Next()
	if !ok || attr.Tag != SGRAttrBold {
		t.Fatalf("expected bold attribute, got %#v (ok=%v)", attr, ok)
	}
	attr, ok = parser.Next()
	if !ok || attr.Tag != SGRAttrFG8 || attr.PaletteIndex != 1 {
		t.Fatalf("expected red foreground attribute, got %#v (ok=%v)", attr, ok)
	}
	if _, ok := parser.Next(); ok {
		t.Fatal("expected parser to be exhausted")
	}

	parser.Reset()
	attr, ok = parser.Next()
	if !ok || attr.Tag != SGRAttrBold {
		t.Fatalf("expected reset iteration to return bold, got %#v", attr)
	}
}

func TestSGRParserValuesAndUnknown(t *testing.T) {
	parser, err := NewSGRParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	if err := parser.SetParams([]uint16{38, 2, 10, 20, 30}, nil); err != nil {
		t.Fatal(err)
	}
	attr, ok := parser.Next()
	if !ok || attr.Tag != SGRAttrDirectColorFG {
		t.Fatalf("expected direct foreground, got %#v", attr)
	}
	if want := (ColorRGB{R: 10, G: 20, B: 30}); attr.Color != want {
		t.Fatalf("expected %#v, got %#v", want, attr.Color)
	}

	if err := parser.SetParams([]uint16{999}, nil); err != nil {
		t.Fatal(err)
	}
	attr, ok = parser.Next()
	if !ok || attr.Tag != SGRAttrUnknown {
		t.Fatalf("expected unknown attribute, got %#v", attr)
	}
	if !slices.Equal(attr.Unknown.Full, []uint16{999}) {
		t.Fatalf("unexpected full unknown params: %v", attr.Unknown.Full)
	}

	if err := parser.SetParams([]uint16{1}, []byte{';', ':'}); err == nil {
		t.Fatal("expected mismatched separators to fail")
	}
}
