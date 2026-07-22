package scanner

import (
	"testing"
	"orez/internal/token"
)

func TestScannerKeywords(t *testing.T) {
	src := []byte(`
cabinet main

def test()
    let x = 123
    if err
        return err
    end
end
`)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	expected := []token.Token{
		token.CABINET, token.IDENT,
		token.DEF, token.IDENT, token.LPAREN, token.RPAREN,
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.IF, token.IDENT,
		token.RETURN, token.IDENT,
		token.END,
		token.END,
	}

	for i, exp := range expected {
		var tok token.Token
		var lit string
		for {
			_, tok, lit = s.Scan()
			if tok != token.SEMICOLON {
				break
			}
		}
		if tok != exp {
			t.Errorf("Index %d: Expected %s, got %s (lit: %s)", i, exp, tok, lit)
		}
	}
}

func TestScannerOperators(t *testing.T) {
	src := []byte(`= == != < <= << <<= > >= >> >>= + += ++ - -= -- * *= / /= % %= ^ ^= & &= && &^ &^= | |= || : . ...`)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	expected := []token.Token{
		token.ASSIGN, token.EQL, token.NEQ, token.LSS, token.LEQ, token.SHL, token.SHL_ASSIGN,
		token.GTR, token.GEQ, token.SHR, token.SHR_ASSIGN, token.ADD, token.ADD_ASSIGN, token.INC,
		token.SUB, token.SUB_ASSIGN, token.DEC, token.MUL, token.MUL_ASSIGN, token.QUO, token.QUO_ASSIGN,
		token.REM, token.REM_ASSIGN, token.XOR, token.XOR_ASSIGN, token.AND, token.AND_ASSIGN, token.LAND,
		token.AND_NOT, token.AND_NOT_ASSIGN, token.OR, token.OR_ASSIGN, token.LOR,
		token.COLON,
		token.PERIOD, token.ELLIPSIS,
	}

	for i, exp := range expected {
		_, tok, lit := s.Scan()
		if tok != exp {
			t.Errorf("Index %d: Expected %s, got %s (lit: %s)", i, exp, tok, lit)
		}
	}
}

func TestScannerStringsAndForbidden(t *testing.T) {
	src := []byte(`"halo dunia" "escape \" test" package func var`)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, 0)

	expected := []token.Token{
		token.STRING, token.STRING,
		token.ILLEGAL, // 'package' is no longer a keyword in 0xg, will fail in Parser
		token.ILLEGAL, // 'func' is no longer a keyword
		token.ILLEGAL, // 'var' is no longer a keyword
	}

	for i, exp := range expected {
		_, tok, lit := s.Scan()
		if tok != exp {
			t.Errorf("Index %d: Expected %s, got %s (lit: %s)", i, exp, tok, lit)
		}
	}
}
