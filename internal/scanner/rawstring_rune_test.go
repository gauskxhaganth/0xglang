package scanner

import (
	"testing"
	"zerouge/internal/token"
)

func TestScannerRawStringsAndRunes(t *testing.T) {
	src := []byte("`raw \n string` '\\n' 'a' '\\x41' '\\u00E4'")
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	expected := []struct {
		tok token.Token
		lit string
	}{
		{token.STRING, "`raw \n string`"},
		{token.CHAR, "'\\n'"},
		{token.CHAR, "'a'"},
		{token.CHAR, "'\\x41'"},
		{token.CHAR, "'\\u00E4'"},
	}

	for i, exp := range expected {
		_, tok, lit := s.Scan()
		if tok != exp.tok || lit != exp.lit {
			t.Errorf("Index %d: Expected %s (%s), got %s (%s)", i, exp.tok, exp.lit, tok, lit)
		}
	}
}
