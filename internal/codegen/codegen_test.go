package codegen

import (
	"bytes"
	"strings"
	"testing"
	"zerouge/internal/parser"
	"zerouge/internal/token"
)

func TestTranspileBasic(t *testing.T) {
	src := []byte("cabinet main\n\ndef myfunc()\nlet x = 1\nend")
	fset := token.NewFileSet()
	
	file, err := parser.ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	
	var buf bytes.Buffer
	transpiler := NewTranspiler(fset)
	err = transpiler.Generate(&buf, file)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	
	expected := "package main\n\nfunc Myfunc() {\n//line :4\n\tx := 1\n}\n"
	result := buf.String()
	
	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("FASE MERAH: Transpilasi tidak sesuai.\nHarapan:\n%s\n--- \nHasil:\n%s", expected, result)
	}
}

// Layer 6 Test: Transpile If and While Blocks (0xg loops convert to native Go 'for')
func TestTranspileControlBlocks(t *testing.T) {
	src := []byte("cabinet main\n\ndef logic()\nif x > 0\nlet y = 1\nend\nwhile x < 10\nlet x = 2\nend\nend")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	var buf bytes.Buffer
	transpiler := NewTranspiler(fset)
	err = transpiler.Generate(&buf, file)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	expectedGo := "package main\n\nfunc Logic() {\n//line :4\n\tif x > 0 {\n//line :5\n\t\ty := 1\n\t}\n//line :7\n\tfor x < 10 {\n//line :8\n\t\tx := 2\n\t}\n}\n"
	result := buf.String()

	if strings.TrimSpace(result) != strings.TrimSpace(expectedGo) {
		t.Errorf("FASE MERAH: Transpilasi If/While tidak sesuai.\nHarapan:\n%s\n---\nHasil:\n%s", expectedGo, result)
	}
}

// Layer 6 Test: Transpile Math Operations (Pratt Parser Output)
func TestTranspileMathExpr(t *testing.T) {
	src := []byte("cabinet main\n\ndef math()\nlet x = -5 + 10 * 2\nend")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	var buf bytes.Buffer
	transpiler := NewTranspiler(fset)
	err = transpiler.Generate(&buf, file)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	expectedGo := "package main\n\nfunc Math() {\n//line :4\n\tx := -5 + 10 * 2\n}\n"
	result := buf.String()

	if strings.TrimSpace(result) != strings.TrimSpace(expectedGo) {
		t.Errorf("FASE MERAH: Transpilasi Math Expr tidak sesuai.\nHarapan:\n%s\n---\nHasil:\n%s", expectedGo, result)
	}
}
