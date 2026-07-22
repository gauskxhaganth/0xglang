package scanner

import (
	"testing"
	"orez/internal/token"
)

func TestScannerAdvancedNumbers(t *testing.T) {
	src := []byte(`0xBADFACE 0o755 0b101010 1.23e4 3.14i`)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	expected := []struct {
		tok token.Token
		lit string
	}{
		{token.INT, "0xBADFACE"},
		{token.INT, "0o755"},
		{token.INT, "0b101010"},
		{token.FLOAT, "1.23e4"},
		{token.IMAG, "3.14i"},
	}

	for i, exp := range expected {
		_, tok, lit := s.Scan()
		if tok != exp.tok || lit != exp.lit {
			t.Errorf("Index %d: Expected %s (%s), got %s (%s)", i, exp.tok, exp.lit, tok, lit)
		}
	}
}
