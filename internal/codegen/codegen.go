package codegen

import (
	"fmt"
	"io"
	"unicode"
	"orez/internal/ast"
	"orez/internal/token"
)

type Transpiler struct {
	w           io.Writer
	indentLevel int
	fset        *token.FileSet
	classes     map[string]bool
	structs     map[string]bool
	varTypes    map[string]string
	hasClass      bool
	inClassMethod bool
	err           error
}

func NewTranspiler(fset *token.FileSet) *Transpiler {
	return &Transpiler{
		fset:     fset,
		classes:  make(map[string]bool),
		structs:  make(map[string]bool),
		varTypes: make(map[string]string),
	}
}

func (t *Transpiler) Generate(w io.Writer, file *ast.File) error {
	t.w = w
	err := t.genFile(file)
	if err != nil {
		return err
	}
	return t.err
}

func (t *Transpiler) write(format string, args ...any) {
	fmt.Fprintf(t.w, format, args...)
}

func (t *Transpiler) indent() {
	for i := 0; i < t.indentLevel; i++ {
		t.write("\t")
	}
}

func (t *Transpiler) genFile(node *ast.File) error {
	t.write("package %s\n\n", node.Name.Name)

	// Pre-scan for class & struct declarations
	for _, decl := range node.Decls {
		if cd, ok := decl.(*ast.ClassDecl); ok {
			t.classes[cd.Name.Name] = true
			t.hasClass = true
		} else if td, ok := decl.(*ast.TypeDecl); ok {
			t.structs[td.Name.Name] = true
		}
	}
	
	var imports []*ast.RequireDecl
	var others []ast.Decl
	
	for _, decl := range node.Decls {
		if req, ok := decl.(*ast.RequireDecl); ok {
			imports = append(imports, req)
		} else {
			others = append(others, decl)
		}
	}
	
	if len(imports) > 0 {
		t.write("import (\n")
		for _, req := range imports {
			pos := t.fset.Position(req.Pos())
			if pos.IsValid() {
				t.write("//line %s:%d\n", pos.Filename, pos.Line)
			}
			for _, pkg := range req.Pkgs {
				t.write("\t")
				t.genExpr(pkg)
				t.write("\n")
			}
		}
		t.write(")\n\n")
	}
	
	for _, decl := range others {
		t.genDecl(decl)
	}
	return nil
}

func (t *Transpiler) genDecl(node ast.Decl) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		t.write("func ")
		if d.Recv != nil && len(d.Recv.List) > 0 {
			t.write("(")
			for i, field := range d.Recv.List {
				if i > 0 {
					t.write(", ")
				}
				if field.Name != nil {
					t.write("%s ", field.Name.Name)
				}
				t.genExpr(field.Type)
			}
			t.write(") ")
		}
		funcName := d.Name.Name
		if d.Recv != nil {
			funcName = capitalize(funcName)
		}
		t.write("%s(", funcName)
		if d.Type != nil && d.Type.Params != nil {
			for i, field := range d.Type.Params.List {
				if i > 0 {
					t.write(", ")
				}
				t.genExpr(field.Name)
				t.write(" ")
				
				// Auto-pointer for testing.T
				isTestingT := false
				if sel, ok := field.Type.(*ast.SelectorExpr); ok {
					if xIdent, ok2 := sel.X.(*ast.Ident); ok2 && xIdent.Name == "testing" && sel.Sel.Name == "T" {
						isTestingT = true
					}
				}
				if isTestingT {
					t.write("*")
				}
				
				t.genExpr(field.Type)
			}
		}
		t.write(")")
		
		if d.Type != nil && d.Type.Results != nil {
			t.write(" ")
			if len(d.Type.Results.List) > 1 {
				t.write("(")
			}
			for i, res := range d.Type.Results.List {
				if i > 0 {
					t.write(", ")
				}
				t.genExpr(res.Type)
			}
			if len(d.Type.Results.List) > 1 {
				t.write(")")
			}
		}
		
		if d.Body != nil {
			t.write(" ")
			t.genBlockStmt(d.Body)
			t.write("\n\n")
		} else {
			t.write("{}\n\n")
		}
	case *ast.ClassDecl:
		cName := capitalize(d.Name.Name)
		t.write("type %s struct {\n", cName)
		if d.Parent != nil {
			t.write("\t%s\n", capitalize(d.Parent.Name))
		}
		if d.Fields != nil {
			for _, field := range d.Fields.List {
				t.write("\t")
				if field.Name != nil {
					t.write("%s ", capitalize(field.Name.Name))
					t.genExpr(field.Type)
				} else {
					// Anonymous field
					t.genExpr(field.Type)
				}
				t.write("\n")
			}
		}
		t.write("}\n\n")

		// Flatten methods inside class to receiver methods outside
		for _, m := range d.Methods {
			mName := capitalize(m.Name.Name)
			t.write("func (self *%s) %s(", cName, mName)
			if m.Type != nil && m.Type.Params != nil {
				for i, field := range m.Type.Params.List {
					if i > 0 {
						t.write(", ")
					}
					t.genExpr(field.Name)
					t.write(" ")
					t.genExpr(field.Type)
				}
			}
			t.write(")")

			if m.Type != nil && m.Type.Results != nil {
				t.write(" ")
				if len(m.Type.Results.List) > 1 {
					t.write("(")
				}
				for i, res := range m.Type.Results.List {
					if i > 0 {
						t.write(", ")
					}
					t.genExpr(res.Type)
				}
				if len(m.Type.Results.List) > 1 {
					t.write(")")
				}
			}

			if m.Body != nil {
				t.write(" ")
				t.inClassMethod = true
				t.genBlockStmt(m.Body)
				t.inClassMethod = false
				t.write("\n\n")
			} else {
				t.write("{}\n\n")
			}
		}
	case *ast.TypeDecl:
		t.write("type %s ", d.Name.Name)
		t.genExpr(d.TypeExpr)
		t.write("\n\n")
	case *ast.LetStmt:
		pos := t.fset.Position(d.Pos())
		if pos.IsValid() {
			t.write("//line %s:%d\n", pos.Filename, pos.Line)
		}
		if d.Type != nil {
			t.write("var %s %s", d.Name.Name, typeTranslator(d.Type.Name))
			if d.Value != nil {
				t.write(" = ")
				t.genExpr(d.Value)
			}
		} else {
			t.write("var %s = ", d.Name.Name)
			if d.Value != nil {
				t.genExpr(d.Value)
			} else {
				t.write("nil") // Fallback
			}
		}
		t.write("\n\n")
	}
}

func (t *Transpiler) genBlockStmt(node *ast.BlockStmt) {
	t.write("{\n")
	t.indentLevel++
	for _, stmt := range node.List {
		t.genStmt(stmt)
		t.write("\n")
	}
	t.indentLevel--
	t.indent()
	t.write("}")
}

func (t *Transpiler) genStmt(node ast.Stmt) {
	if node == nil {
		return
	}
	pos := t.fset.Position(node.Pos())
	if pos.IsValid() {
		t.write("//line %s:%d\n", pos.Filename, pos.Line)
	}
	t.indent()
	switch s := node.(type) {
	case *ast.LetStmt:
		if s.Value != nil {
			if comp, ok := s.Value.(*ast.CompositeLit); ok {
				if ident, ok := comp.Type.(*ast.Ident); ok {
					t.varTypes[s.Name.Name] = ident.Name
				}
			}
		}
		if s.Type != nil {
			t.write("var %s %s", s.Name.Name, typeTranslator(s.Type.Name))
			if s.Value != nil {
				t.write(" = ")
				t.genExpr(s.Value)
			}
		} else {
			t.write("%s := ", s.Name.Name)
			t.genExpr(s.Value)
		}
	case *ast.AssignStmt:
		t.genExpr(s.Left)
		t.write(" %s ", s.Tok.String())
		t.genExpr(s.Right)
	case *ast.IfStmt:
		t.write("if ")
		// Intersep "if err" (Blood Lock Metaprogramming)
		if ident, ok := s.Cond.(*ast.Ident); ok && ident.Name == "err" {
			t.write("err != nil ")
		} else if prefix, ok := s.Cond.(*ast.PrefixExpr); ok && prefix.Operator == "!" {
			if pIdent, ok := prefix.Right.(*ast.Ident); ok && pIdent.Name == "err" {
				t.write("err == nil ")
			} else {
				t.genExpr(s.Cond)
				t.write(" ")
			}
		} else {
			t.genExpr(s.Cond)
			t.write(" ")
		}
		t.genBlockStmt(s.Body)
	case *ast.WhileStmt:
		t.write("for ") // 0xg 'while' menjadi Go 'for' murni
		if ident, ok := s.Cond.(*ast.Ident); ok && ident.Name == "err" {
			t.write("err != nil ")
		} else if prefix, ok := s.Cond.(*ast.PrefixExpr); ok && prefix.Operator == "!" {
			if pIdent, ok := prefix.Right.(*ast.Ident); ok && pIdent.Name == "err" {
				t.write("err == nil ")
			} else {
				t.genExpr(s.Cond)
				t.write(" ")
			}
		} else {
			t.genExpr(s.Cond)
			t.write(" ")
		}
		t.genBlockStmt(s.Body)
	case *ast.LoopStmt:
		t.write("for ")
		t.genBlockStmt(s.Body)
	case *ast.ForeachStmt:
		t.write("for ")
		if s.Key != nil {
			t.write("%s, %s := range ", s.Key.Name, s.Value.Name)
		} else {
			t.write("_, %s := range ", s.Value.Name)
		}
		t.genExpr(s.X)
		t.write(" ")
		t.genBlockStmt(s.Body)
	case *ast.SwitchStmt:
		t.write("switch ")
		if s.Tag != nil {
			t.genExpr(s.Tag)
			t.write(" ")
		}
		t.write("{\n")
		t.indentLevel++
		for _, stmt := range s.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				t.indent()
				if len(clause.List) > 0 {
					t.write("case ")
					for i, expr := range clause.List {
						if i > 0 {
							t.write(", ")
						}
						t.genExpr(expr)
					}
					t.write(":\n")
				} else {
					t.write("default:\n")
				}
				
				t.indentLevel++
				for _, bStmt := range clause.Body {
					t.indent()
					t.genStmt(bStmt)
					t.write("\n")
				}
				t.indentLevel--
			}
		}
		t.indentLevel--
		t.indent()
		t.write("}")
	case *ast.SelectStmt:
		t.write("select {\n")
		t.indentLevel++
		for _, stmt := range s.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				t.indent()
				if len(clause.List) > 0 {
					t.write("case ")
					t.genExpr(clause.List[0])
					t.write(":\n")
				} else {
					t.write("default:\n")
				}
				
				t.indentLevel++
				for _, bStmt := range clause.Body {
					t.indent()
					t.genStmt(bStmt)
					t.write("\n")
				}
				t.indentLevel--
			}
		}
		t.indentLevel--
		t.indent()
		t.write("}")
	case *ast.GoStmt:
		t.write("go ")
		t.genExpr(s.Call)
	case *ast.DeferStmt:
		t.write("defer ")
		t.genExpr(s.Call)
	case *ast.ExprStmt:
		t.genExpr(s.X)
	case *ast.ReturnStmt:
		t.write("return")
		if len(s.Results) > 0 {
			t.write(" ")
			for i, res := range s.Results {
				if i > 0 {
					t.write(", ")
				}
				t.genExpr(res)
			}
		}
	}
}

func (t *Transpiler) genExpr(node ast.Expr) {
	if node == nil {
		return
	}
	switch e := node.(type) {
	case *ast.BasicLit:
		t.write("%s", e.Value)
	case *ast.Ident:
		if e != nil {
			t.write("%s", typeTranslator(e.Name))
		}
	case *ast.PrefixExpr:
		t.write("%s", e.Operator)
		t.genExpr(e.Right)
	case *ast.InfixExpr:
		t.genExpr(e.Left)
		t.write(" %s ", e.Operator)
		t.genExpr(e.Right)
	case *ast.CallExpr:
		if ident, ok := e.Fun.(*ast.Ident); ok {
			if ident.Name == "Array" && len(e.Args) == 1 {
				t.write("[]")
				t.genExpr(e.Args[0])
				return
			} else if ident.Name == "Rigid" && len(e.Args) == 2 {
				t.write("[")
				t.genExpr(e.Args[0])
				t.write("]")
				t.genExpr(e.Args[1])
				return
			} else if ident.Name == "Map" && len(e.Args) == 2 {
				t.write("map[")
				t.genExpr(e.Args[0])
				t.write("]")
				t.genExpr(e.Args[1])
				return
			} else if ident.Name == "Channel" && len(e.Args) == 1 {
				t.write("chan ")
				t.genExpr(e.Args[0])
				return
			}
		}

		t.genExpr(e.Fun)
		t.write("(")
		for i, arg := range e.Args {
			if i > 0 {
				t.write(", ")
			}
			t.genExpr(arg)
		}
		t.write(")")
	case *ast.SelectorExpr:
		if e.IsDoubleColon {
			if !t.hasClass {
				pos := t.fset.Position(e.ColonPos)
				t.err = fmt.Errorf("%s: FATAL: Blood Lock - '::' is forbidden in struct-style files. Use '.' instead.", pos.String())
				return
			}
			if xIdent, ok := e.X.(*ast.Ident); ok {
				typeName := t.varTypes[xIdent.Name]
				if t.structs[xIdent.Name] || t.structs[typeName] {
					pos := t.fset.Position(e.ColonPos)
					t.err = fmt.Errorf("%s: FATAL: Blood Lock - '::' is forbidden on 'struct'. Use '.' instead.", pos.String())
					return
				}
			}
		}
		t.genExpr(e.X)
		t.write(".")
		t.write("%s", capitalize(e.Sel.Name))
	case *ast.StructType:
		t.write("struct {\n")
		t.indentLevel++
		for _, field := range e.Fields.List {
			t.indent()
			if field.Name != nil {
				t.write("%s ", capitalize(field.Name.Name))
			}
			t.genExpr(field.Type)
			t.write("\n")
		}
		t.indentLevel--
		t.indent()
		t.write("}")
	case *ast.AtExpr:
		if !t.inClassMethod {
			pos := t.fset.Position(e.At)
			t.err = fmt.Errorf("%s: FATAL: '@' (instance variable) is only allowed inside class methods.", pos.String())
			return
		}
		t.write("self.%s", capitalize(e.Name.Name))
	case *ast.CompositeLit:

		t.genExpr(e.Type)
		t.write("{\n")
		t.indentLevel++
		for i, elt := range e.Elts {
			if i > 0 {
				t.write(",\n")
			}
			t.indent()
			t.genExpr(elt)
		}
		if len(e.Elts) > 0 {
			t.write(",\n")
		}
		t.indentLevel--
		t.indent()
		t.write("}")
	case *ast.KeyValueExpr:
		if kIdent, ok := e.Key.(*ast.Ident); ok {
			t.write("%s", capitalize(kIdent.Name))
		} else {
			t.genExpr(e.Key)
		}
		t.write(": ")
		t.genExpr(e.Value)
	case *ast.IndexExpr:
		t.genExpr(e.X)
		t.write("[")
		t.genExpr(e.Index)
		t.write("]")
	}
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func typeTranslator(name string) string {
	switch name {
	case "String":
		return "string"
	case "Int":
		return "int"
	case "Int8":
		return "int8"
	case "Int16":
		return "int16"
	case "Int32":
		return "int32"
	case "Int64":
		return "int64"
	case "Float32":
		return "float32"
	case "Float64":
		return "float64"
	case "Bool":
		return "bool"
	default:
		return name
	}
}
