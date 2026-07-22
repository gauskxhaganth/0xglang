package ast

import "orez/internal/token"

// Node is the base interface for all AST nodes
type Node interface {
	Pos() token.Pos
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Decl interface {
	Node
	declNode()
}

// Ident represents an identifier (e.g., variable, function, or package name)
type Ident struct {
	NamePos token.Pos
	Name    string
}

func (x *Ident) Pos() token.Pos { return x.NamePos }
func (x *Ident) exprNode()      {}

// FieldList merepresentasikan daftar parameter
type FieldList struct {
	Opening token.Pos
	List    []*Field
	Closing token.Pos
}
func (f *FieldList) Pos() token.Pos { return f.Opening }

// Field merepresentasikan satu parameter
type Field struct {
	Name *Ident
	Type Expr
}
func (f *Field) Pos() token.Pos { return f.Name.Pos() }

// FuncType contains function type (parameters and returns)
type FuncType struct {
	Func    token.Pos
	Params  *FieldList
	Results *FieldList
}
func (f *FuncType) Pos() token.Pos { return f.Func }
func (f *FuncType) exprNode()      {}

// BlockStmt merepresentasikan blok kode yang diakhiri 'end'
type BlockStmt struct {
	List []Stmt
	End  token.Pos // posisi 'end'
}
func (b *BlockStmt) Pos() token.Pos {
	if len(b.List) > 0 {
		return b.List[0].Pos()
	}
	return b.End
}
func (b *BlockStmt) stmtNode() {}

// LetStmt merepresentasikan deklarasi variabel (let x = 5)
type LetStmt struct {
	Let   token.Pos
	Name  *Ident
	Type  *Ident // Opsional tipe data (misal: let user Pengguna)
	Value Expr   // Optional if only a type declaration
}
func (l *LetStmt) Pos() token.Pos { return l.Let }
func (l *LetStmt) stmtNode()      {}
func (l *LetStmt) declNode()      {}

// BasicLit merepresentasikan literal (angka, string, dll)
type BasicLit struct {
	ValuePos token.Pos
	Kind     token.Token
	Value    string
}
func (b *BasicLit) Pos() token.Pos { return b.ValuePos }
func (b *BasicLit) exprNode()      {}

// PrefixExpr (e.g., -x, !y)
type PrefixExpr struct {
	OpPos    token.Pos
	Operator string
	Right    Expr
}
func (p *PrefixExpr) Pos() token.Pos { return p.OpPos }
func (p *PrefixExpr) exprNode()      {}

// InfixExpr (e.g., x + y)
type InfixExpr struct {
	Left     Expr
	OpPos    token.Pos
	Operator string
	Right    Expr
}
func (i *InfixExpr) Pos() token.Pos { return i.Left.Pos() }
func (i *InfixExpr) exprNode()      {}

// IfStmt (e.g., if x > 0 ... end)
type IfStmt struct {
	If   token.Pos
	Cond Expr
	Body *BlockStmt
}
func (i *IfStmt) Pos() token.Pos { return i.If }
func (i *IfStmt) stmtNode()      {}

// WhileStmt (e.g., while x < 10 ... end)
type WhileStmt struct {
	While token.Pos
	Cond  Expr
	Body  *BlockStmt
}
func (w *WhileStmt) Pos() token.Pos { return w.While }
func (w *WhileStmt) stmtNode()      {}

// FuncDecl represents a 'def' function declaration
type FuncDecl struct {
	Def  token.Pos  // posisi keyword 'def'
	Recv *FieldList // receiver (e.g., u User) if any
	Name *Ident     // function name
	Type *FuncType  // function type (parameters)
	Body *BlockStmt // function body
}

func (d *FuncDecl) Pos() token.Pos { return d.Def }
func (d *FuncDecl) declNode()      {}

// File merepresentasikan satu unit file source code 0xg
type File struct {
	Cabinet token.Pos // posisi keyword 'cabinet'
	Name    *Ident    // nama cabinet (package)
	Decls   []Decl    // daftar deklarasi level atas (seperti def)
}

func (f *File) Pos() token.Pos { return f.Cabinet }

// ExprStmt represents an expression evaluated as a statement (e.g., independent function call)
type ExprStmt struct {
	X Expr // Ekspresi itu sendiri
}
func (e *ExprStmt) Pos() token.Pos { return e.X.Pos() }
func (e *ExprStmt) stmtNode()      {}

// CallExpr represents a function call
type CallExpr struct {
	Fun    Expr      // Function name or expression
	Lparen token.Pos // Posisi '('
	Args   []Expr    // Daftar argumen
	Rparen token.Pos // Posisi ')'
}
func (c *CallExpr) Pos() token.Pos { return c.Fun.Pos() }
func (c *CallExpr) exprNode()      {}

// RequireDecl merepresentasikan impor paket (require "fmt" retain)
type RequireDecl struct {
	Require token.Pos   // Posisi 'require'
	Pkgs    []*BasicLit // Packages (e.g., "fmt", "os")
	Retain  token.Pos // Posisi 'retain'
}
func (r *RequireDecl) Pos() token.Pos { return r.Require }
func (r *RequireDecl) declNode()      {}

// SelectorExpr represents attribute access or package function call (e.g., fmt.Println)
type SelectorExpr struct {
	X             Expr
	Sel           *Ident
	IsDoubleColon bool
	ColonPos      token.Pos
}
func (s *SelectorExpr) Pos() token.Pos { return s.X.Pos() }
func (s *SelectorExpr) exprNode()      {}

// ReturnStmt merepresentasikan pengembalian nilai (return)
type ReturnStmt struct {
	Return  token.Pos
	Results []Expr
}
func (r *ReturnStmt) Pos() token.Pos { return r.Return }
func (r *ReturnStmt) stmtNode()      {}

// TypeDecl merepresentasikan deklarasi tipe data (seperti struct)
type TypeDecl struct {
	Type     token.Pos
	Name     *Ident
	TypeExpr Expr
}
func (t *TypeDecl) Pos() token.Pos { return t.Type }
func (t *TypeDecl) declNode()      {}

// StructType merepresentasikan bentuk tipe struct
type StructType struct {
	Struct token.Pos
	Fields *FieldList
}
func (s *StructType) Pos() token.Pos { return s.Struct }
func (s *StructType) exprNode()      {}

// ClassDecl merepresentasikan deklarasi kelas ala Ruby/Crystal (Heap / Pointer Type)
type ClassDecl struct {
	Class   token.Pos
	Name    *Ident
	Parent  *Ident // null if there is no inheritance (< Parent)
	Fields  *FieldList
	Methods []*FuncDecl
}
func (c *ClassDecl) Pos() token.Pos { return c.Class }
func (c *ClassDecl) declNode()      {}

// AtExpr merepresentasikan ekspresi @ident (instance variable)
type AtExpr struct {
	At   token.Pos
	Name *Ident
}
func (a *AtExpr) Pos() token.Pos { return a.At }
func (a *AtExpr) exprNode()      {}

// AssignStmt merepresentasikan penugasan ulang (mutasi murni tanpa let)
type AssignStmt struct {
	Left   Expr
	TokPos token.Pos
	Tok    token.Token
	Right  Expr
}
func (a *AssignStmt) Pos() token.Pos { return a.Left.Pos() }
func (a *AssignStmt) stmtNode()      {}

// CompositeLit merepresentasikan inisialisasi nilai komposit
type CompositeLit struct {
	Type   Expr
	Lbrace token.Pos
	Elts   []Expr
	Rbrace token.Pos
}
func (c *CompositeLit) Pos() token.Pos {
	if c.Type != nil {
		return c.Type.Pos()
	}
	return c.Lbrace
}
func (c *CompositeLit) exprNode()      {}

// KeyValueExpr merepresentasikan pasangan kunci: nilai
type KeyValueExpr struct {
	Key   Expr
	Colon token.Pos
	Value Expr
}
func (k *KeyValueExpr) Pos() token.Pos { return k.Key.Pos() }
func (k *KeyValueExpr) exprNode()      {}

// IndexExpr merepresentasikan akses indeks
type IndexExpr struct {
	X      Expr
	Lbrack token.Pos
	Index  Expr
	Rbrack token.Pos
}
func (i *IndexExpr) Pos() token.Pos { return i.X.Pos() }
func (i *IndexExpr) exprNode()      {}

// LoopStmt merepresentasikan perulangan tanpa henti (loop ... end)
type LoopStmt struct {
	Loop token.Pos
	Body *BlockStmt
}
func (l *LoopStmt) Pos() token.Pos { return l.Loop }
func (l *LoopStmt) stmtNode()      {}

// ForeachStmt merepresentasikan perulangan rentang data (foreach k, v in data ... end)
type ForeachStmt struct {
	Foreach token.Pos
	Key     *Ident // Optional, can be nil
	Value   *Ident
	In      token.Pos
	X       Expr
	Body    *BlockStmt
}
func (f *ForeachStmt) Pos() token.Pos { return f.Foreach }
func (f *ForeachStmt) stmtNode()      {}

// SwitchStmt merepresentasikan blok percabangan case ... when ... end
type SwitchStmt struct {
	Case  token.Pos
	Tag   Expr          // Tag can be nil
	Body  *BlockStmt    // Body menyimpan kumpulan CaseClause
}
func (s *SwitchStmt) Pos() token.Pos { return s.Case }
func (s *SwitchStmt) stmtNode()      {}

// CaseClause merepresentasikan klausa when (kondisi) ...
type CaseClause struct {
	When  token.Pos
	List  []Expr        // Condition list (empty for else/default)
	Body  []Stmt
}
func (c *CaseClause) Pos() token.Pos { return c.When }
func (c *CaseClause) stmtNode()      {}

// SelectStmt merepresentasikan blok multiplexer select ... when ... end
type SelectStmt struct {
	Select token.Pos
	Body   *BlockStmt // Body menyimpan kumpulan CaseClause
}
func (s *SelectStmt) Pos() token.Pos { return s.Select }
func (s *SelectStmt) stmtNode()      {}

// GoStmt merepresentasikan eksekusi asinkron goroutine
type GoStmt struct {
	Go   token.Pos
	Call *CallExpr
}
func (g *GoStmt) Pos() token.Pos { return g.Go }
func (g *GoStmt) stmtNode()      {}

// DeferStmt merepresentasikan penundaan eksekusi
type DeferStmt struct {
	Defer token.Pos
	Call  *CallExpr
}
func (d *DeferStmt) Pos() token.Pos { return d.Defer }
func (d *DeferStmt) stmtNode()      {}
