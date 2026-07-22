package parser

import (
	"testing"
	"zerouge/internal/ast"
	"zerouge/internal/token"
)

// Uji Peluru Pelacak (Tracer Bullet): Memastikan Parser mampu membaca 'cabinet main'
func TestParseCabinet(t *testing.T) {
	src := []byte("cabinet main")
	fset := token.NewFileSet()

	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if file == nil {
		t.Fatalf("FASE MERAH: ParseFile mengembalikan nil AST")
	}

	if file.Name == nil || file.Name.Name != "main" {
		t.Errorf("Expected cabinet name 'main', got %v", file.Name)
	}
}

// Layer 2 Test: Ensure Parser assembles 'def' declarations into FuncDecl
func TestParseDef(t *testing.T) {
	src := []byte("cabinet main\n\ndef myfunc()\nend")
	fset := token.NewFileSet()

	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	if len(file.Decls) != 1 {
		t.Fatalf("FASE MERAH: Expected 1 declaration, got %d", len(file.Decls))
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("Expected *ast.FuncDecl, got %T", file.Decls[0])
	}

	if funcDecl.Name.Name != "myfunc" {
		t.Errorf("Expected func name 'myfunc', got '%s'", funcDecl.Name.Name)
	}
}

// Layer 3 Test: Function Parameters, Block Statement (end), and Let Statement
func TestParseFuncParams(t *testing.T) {
	src := []byte("cabinet main\n\ndef myfunc(a Int, b String)\nend")
	fset := token.NewFileSet()

	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	funcDecl := file.Decls[0].(*ast.FuncDecl)
	if funcDecl.Type == nil || funcDecl.Type.Params == nil {
		t.Fatalf("FASE MERAH: FuncDecl.Type atau Params nil")
	}

	if len(funcDecl.Type.Params.List) != 2 {
		t.Fatalf("Expected 2 parameters, got %d", len(funcDecl.Type.Params.List))
	}
}

func TestParseBlockAndLet(t *testing.T) {
	src := []byte("cabinet main\n\ndef myfunc()\nlet x = 1\nend")
	fset := token.NewFileSet()

	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	funcDecl := file.Decls[0].(*ast.FuncDecl)
	if funcDecl.Body == nil {
		t.Fatalf("FASE MERAH: FuncDecl.Body nil (BlockStmt belum ditangani)")
	}

	if len(funcDecl.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in body, got %d", len(funcDecl.Body.List))
	}

	letStmt, ok := funcDecl.Body.List[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("Expected *ast.LetStmt, got %T", funcDecl.Body.List[0])
	}

	if letStmt.Name.Name != "x" {
		t.Errorf("Expected let variable 'x', got '%s'", letStmt.Name.Name)
	}
}

// Uji Lapis 4: Pratt Parsing (Math/Logic)
func TestPrattParsing(t *testing.T) {
	src := []byte("cabinet main\n\ndef math()\nlet x = -5 + 10 * 2\nend")
	fset := token.NewFileSet()
	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	letStmt := funcDecl.Body.List[0].(*ast.LetStmt)

	infix, ok := letStmt.Value.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("FASE MERAH: Expected InfixExpr, got %T", letStmt.Value)
	}
	if infix.Operator != "+" {
		t.Errorf("Expected operator '+', got '%s'", infix.Operator)
	}

	prefix, ok := infix.Left.(*ast.PrefixExpr)
	if !ok {
		t.Fatalf("Expected PrefixExpr on left of '+', got %T", infix.Left)
	}
	if prefix.Operator != "-" {
		t.Errorf("Expected operator '-', got '%s'", prefix.Operator)
	}

	right, ok := infix.Right.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("Expected InfixExpr on right of '+', got %T", infix.Right)
	}
	if right.Operator != "*" {
		t.Errorf("Expected operator '*', got '%s'", right.Operator)
	}
}

// Uji Lapis 4: Blok Kendali (if, while)
func TestParseControlBlocks(t *testing.T) {
	src := []byte("cabinet main\n\ndef logic()\nif x > 0\nend\nwhile y < 10\nend\nend")
	fset := token.NewFileSet()
	file, err := ParseFile(fset, "", src)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	body := file.Decls[0].(*ast.FuncDecl).Body.List
	if len(body) != 2 {
		t.Fatalf("FASE MERAH: Expected 2 statements in body, got %d", len(body))
	}

	ifStmt, ok := body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("FASE MERAH: Expected IfStmt, got %T", body[0])
	}
	if ifStmt.Cond == nil {
		t.Fatalf("FASE MERAH: IfStmt.Cond nil")
	}

	whileStmt, ok := body[1].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("FASE MERAH: Expected WhileStmt, got %T", body[1])
	}
	if whileStmt.Cond == nil {
		t.Fatalf("FASE MERAH: WhileStmt.Cond nil")
	}
}

func TestBloodLockIfErr(t *testing.T) {
	src := []byte("cabinet main\ndef check()\nif err != nil\nend\nend")
	fset := token.NewFileSet()
	_, err := ParseFile(fset, "", src)
	if err == nil {
		t.Fatalf("FASE MERAH: Blood Lock gagal! Parser tidak memblokir 'if err != nil'")
	}
}
