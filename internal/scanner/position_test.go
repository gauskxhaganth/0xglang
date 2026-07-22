package scanner

import (
	"testing"
	"zerouge/internal/token"
)

func TestScannerPositionTracking(t *testing.T) {
	src := []byte("cabinet main\n\ndef test()")
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	// Token 1: cabinet (Baris 1, Kolom 1)
	pos, tok, _ := s.Scan()
	p := fset.Position(pos)
	if tok != token.CABINET || p.Line != 1 || p.Column != 1 {
		t.Errorf("Expected cabinet at 1:1, got %s at %d:%d", tok, p.Line, p.Column)
	}

	// Token 2: main (Baris 1, Kolom 9)
	pos, tok, _ = s.Scan()
	p = fset.Position(pos)
	if tok != token.IDENT || p.Line != 1 || p.Column != 9 {
		t.Errorf("Expected IDENT at 1:9, got %s at %d:%d", tok, p.Line, p.Column)
	}

	// Token 3: SEMICOLON (ASI Injection due to newline after identifier)
	pos, tok, _ = s.Scan()
	if tok != token.SEMICOLON {
		t.Errorf("Expected SEMICOLON, got %s", tok)
	}

	// Token 4: def (Baris 3, Kolom 1)
	pos, tok, _ = s.Scan()
	p = fset.Position(pos)
	if tok != token.DEF || p.Line != 3 || p.Column != 1 {
		t.Errorf("Expected def at 3:1, got %s at %d:%d", tok, p.Line, p.Column)
	}
}
