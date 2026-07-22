package parser

import (
	"fmt"
	"zerouge/internal/ast"
	"zerouge/internal/scanner"
	"zerouge/internal/token"
)

type prefixParseFn func() ast.Expr
type infixParseFn func(ast.Expr) ast.Expr

type Parser struct {
	fset    *token.FileSet
	scanner scanner.Scanner
	pos     token.Pos
	tok     token.Token
	lit     string
	errors  []error

	prefixParseFns map[token.Token]prefixParseFn
	infixParseFns  map[token.Token]infixParseFn
}

func (p *Parser) error(pos token.Pos, msg string) {
	position := p.fset.Position(pos)
	p.errors = append(p.errors, fmt.Errorf("%s: %s", position.String(), msg))
}

func (p *Parser) registerPrefix(tokenType token.Token, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Token, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) init(fset *token.FileSet, filename string, src []byte) {
	p.fset = fset
	file := fset.AddFile(filename, fset.Base(), len(src))
	p.scanner.Init(file, src, func(pos token.Position, msg string) {
		p.errors = append(p.errors, fmt.Errorf("%s: %s", pos.String(), msg))
	}, 0)
	
	p.prefixParseFns = make(map[token.Token]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdent)
	p.registerPrefix(token.CHANNEL, p.parseIdent) // Treat Channel keyword as Ident in expressions
	p.registerPrefix(token.INT, p.parseBasicLit)
	p.registerPrefix(token.STRING, p.parseBasicLit)
	p.registerPrefix(token.SUB, p.parsePrefixExpr)
	p.registerPrefix(token.NOT, p.parsePrefixExpr)
	p.registerPrefix(token.ARROW, p.parsePrefixExpr)
	p.registerPrefix(token.MUL, p.parsePrefixExpr)
	p.registerPrefix(token.AND, p.parsePrefixExpr)
	p.registerPrefix(token.LBRACE, p.parsePrefixCompositeLit)
	p.registerPrefix(token.AT, p.parseAtExpr)

	p.infixParseFns = make(map[token.Token]infixParseFn)
	p.registerInfix(token.DOUBLE_COLON, p.parseSelectorExpr)
	p.registerInfix(token.ADD, p.parseInfixExpr)
	p.registerInfix(token.SUB, p.parseInfixExpr)
	p.registerInfix(token.MUL, p.parseInfixExpr)
	p.registerInfix(token.QUO, p.parseInfixExpr)
	p.registerInfix(token.NEQ, p.parseInfixExpr)
	p.registerInfix(token.EQL, p.parseInfixExpr)
	p.registerInfix(token.LSS, p.parseInfixExpr)
	p.registerInfix(token.GTR, p.parseInfixExpr)
	p.registerInfix(token.ARROW, p.parseInfixExpr)
	p.registerInfix(token.LPAREN, p.parseCallExpr)
	p.registerInfix(token.PERIOD, p.parseSelectorExpr)
	p.registerInfix(token.LBRACE, p.parseCompositeLit)
	p.registerInfix(token.COLON, p.parseKeyValueExpr)
	p.registerInfix(token.LBRACK, p.parseIndexExpr)

	p.next()
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	pos := p.pos
	p.next() // consume 'def'

	var recv *ast.FieldList
	if p.tok == token.LPAREN {
		recv = p.parseFieldList()
	}

	if p.tok != token.IDENT {
		return nil
	}
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next() // consume name

	var params *ast.FieldList
	if p.tok == token.LPAREN {
		params = p.parseFieldList()
	}

	var results *ast.FieldList
	if p.tok == token.IDENT {
		resType := &ast.Ident{NamePos: p.pos, Name: p.lit}
		p.next()
		results = &ast.FieldList{
			List: []*ast.Field{{Type: resType}},
		}
	}

	if p.tok == token.SEMICOLON { // ASI semicolon setelah baris def myfunc()
		p.next()
	}

	body := p.parseBlockStmt()

	return &ast.FuncDecl{
		Def:  pos,
		Recv: recv,
		Name: name,
		Type: &ast.FuncType{Func: pos, Params: params, Results: results},
		Body: body,
	}
}

func (p *Parser) parseFieldList() *ast.FieldList {
	pos := p.pos
	p.next() // consume '('
	var list []*ast.Field
	for p.tok != token.RPAREN && p.tok != token.EOF {
		if p.tok == token.IDENT {
			name := &ast.Ident{NamePos: p.pos, Name: p.lit}
			p.next()
			
			typ := p.parseExpr(token.LowestPrec)
			list = append(list, &ast.Field{Name: name, Type: typ})
			
			if p.tok == token.COMMA {
				p.next()
			}
		} else {
			p.next() // error recovery
		}
	}
	endPos := p.pos
	if p.tok == token.RPAREN {
		p.next()
	}
	return &ast.FieldList{Opening: pos, List: list, Closing: endPos}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	var list []ast.Stmt
	for p.tok != token.END && p.tok != token.EOF {
		if stmt := p.parseStmt(); stmt != nil {
			list = append(list, stmt)
		}
	}
	endPos := p.pos
	if p.tok == token.END {
		p.next()
	}
	return &ast.BlockStmt{List: list, End: endPos}
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.tok {
	case token.LET:
		return p.parseLetStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.LOOP:
		return p.parseLoopStmt()
	case token.FOREACH:
		return p.parseForeachStmt()
	case token.CASE:
		return p.parseSwitchStmt()
	case token.SELECT:
		return p.parseSelectStmt()
	case token.GO:
		return p.parseGoStmt()
	case token.DEFER:
		return p.parseDeferStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.SEMICOLON:
		p.next()
		return nil
	case token.LBRACE:
		p.error(p.pos, "FATAL: curly braces '{' are forbidden for blocks. Remove the '{' brace since 0xg uses block-keywords and 'end'.")
		p.next()
		return nil
	case token.RBRACE:
		p.error(p.pos, "FATAL: curly braces '}' are forbidden for blocks. Use 'end' instead of '}'.")
		p.next()
		return nil
	default:
		// Try to parse as Expression (could be AssignStmt or ExprStmt)
		if expr := p.parseExpr(token.LowestPrec); expr != nil {
			if p.tok == token.ASSIGN || p.tok == token.DEFINE {
				tokPos := p.pos
				tok := p.tok
				p.next() // consume '=' or ':='
				right := p.parseExpr(token.LowestPrec)
				if p.tok == token.SEMICOLON {
					p.next()
				}
				return &ast.AssignStmt{Left: expr, TokPos: tokPos, Tok: tok, Right: right}
			}
			
			if p.tok == token.SEMICOLON {
				p.next()
			}
			return &ast.ExprStmt{X: expr}
		}

		// Avoid infinite loop if still mismatched
		if p.tok != token.END && p.tok != token.EOF {
			p.next()
		}
		return nil
	}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	pos := p.pos
	p.next() // consume 'if'
	cond := p.parseExpr(token.LowestPrec)

	// Blood Lock: Cegah 'if err != nil'
	if infix, ok := cond.(*ast.InfixExpr); ok && infix.Operator == "!=" {
		if left, ok := infix.Left.(*ast.Ident); ok && left.Name == "err" {
			if right, ok := infix.Right.(*ast.Ident); ok && right.Name == "nil" {
				p.error(pos, "FATAL: BLOOD LOCK: 'if err != nil' is forbidden in 0xg. Use 'if err'.")
			}
		}
	}

	body := p.parseBlockStmt()
	return &ast.IfStmt{If: pos, Cond: cond, Body: body}
}

func (p *Parser) parseWhileStmt() ast.Stmt {
	pos := p.pos
	p.next() // consume 'while'
	cond := p.parseExpr(token.LowestPrec)
	body := p.parseBlockStmt()
	return &ast.WhileStmt{While: pos, Cond: cond, Body: body}
}

func (p *Parser) parseLoopStmt() *ast.LoopStmt {
	pos := p.pos
	p.next() // consume 'loop'
	body := p.parseBlockStmt()
	return &ast.LoopStmt{Loop: pos, Body: body}
}

func (p *Parser) parseForeachStmt() *ast.ForeachStmt {
	pos := p.pos
	p.next() // consume 'foreach'
	
	var key, value *ast.Ident
	if p.tok != token.IDENT {
		return nil
	}
	first := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()
	
	if p.tok == token.COMMA {
		p.next() // consume ','
		if p.tok != token.IDENT {
			return nil
		}
		second := &ast.Ident{NamePos: p.pos, Name: p.lit}
		p.next()
		
		key = first
		value = second
	} else {
		value = first
	}
	
	if p.tok != token.IN {
		return nil
	}
	inPos := p.pos
	p.next() // consume 'in'
	
	x := p.parseExpr(token.LowestPrec)
	body := p.parseBlockStmt()
	
	return &ast.ForeachStmt{Foreach: pos, Key: key, Value: value, In: inPos, X: x, Body: body}
}

func (p *Parser) parseSwitchStmt() *ast.SwitchStmt {
	pos := p.pos
	p.next() // consume 'case'
	
	var tag ast.Expr
	if p.tok != token.WHEN && p.tok != token.ELSE && p.tok != token.END {
		tag = p.parseExpr(token.LowestPrec)
	}
	
	body := &ast.BlockStmt{}
	
	for p.tok != token.END && p.tok != token.EOF {
		if p.tok == token.WHEN || p.tok == token.ELSE {
			body.List = append(body.List, p.parseCaseClause())
		} else {
			// Check for stray tokens before when
			p.next()
		}
	}
	
	body.End = p.pos
	if p.tok == token.END {
		p.next() // consume 'end'
	}
	
	return &ast.SwitchStmt{Case: pos, Tag: tag, Body: body}
}

func (p *Parser) parseCaseClause() *ast.CaseClause {
	pos := p.pos
	isElse := p.tok == token.ELSE
	p.next() // consume 'when' or 'else'
	
	var list []ast.Expr
	if !isElse {
		for p.tok != token.EOF {
			list = append(list, p.parseExpr(token.LowestPrec))
			if p.tok == token.COMMA {
				p.next() // consume ','
			} else {
				break
			}
		}
	}
	
	var body []ast.Stmt
	for p.tok != token.WHEN && p.tok != token.ELSE && p.tok != token.END && p.tok != token.EOF {
		if stmt := p.parseStmt(); stmt != nil {
			body = append(body, stmt)
		}
	}
	
	return &ast.CaseClause{When: pos, List: list, Body: body}
}

func (p *Parser) parseSelectStmt() *ast.SelectStmt {
	pos := p.pos
	p.next() // consume 'select'
	
	body := &ast.BlockStmt{}
	for p.tok != token.END && p.tok != token.EOF {
		if p.tok == token.WHEN || p.tok == token.ELSE {
			body.List = append(body.List, p.parseSelectCaseClause())
		} else {
			p.next() // error recovery
		}
	}
	body.End = p.pos
	if p.tok == token.END {
		p.next() // consume 'end'
	}
	
	return &ast.SelectStmt{Select: pos, Body: body}
}

func (p *Parser) parseSelectCaseClause() *ast.CaseClause {
	pos := p.pos
	isDefault := p.tok == token.ELSE
	p.next() // consume 'when' or 'else'

	var list []ast.Expr
	if !isDefault {
		// Parsing statement pengiriman/penerimaan channel
		// Example: msg1 := <-c1, or <-time.After(...)
		// Because := is an assignment token, parse as statement
		stmt := p.parseStmt()
		
		// Convert statement back to expression to store in CaseClause.List
		// (This is a temporary hack to reuse ast.CaseClause)
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			list = append(list, exprStmt.X)
		} else if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			// Create a mock InfixExpr to represent x := y
			list = append(list, &ast.InfixExpr{
				Left:     assignStmt.Left,
				Operator: assignStmt.Tok.String(),
				Right:    assignStmt.Right,
			})
		}
	}

	var body []ast.Stmt
	for p.tok != token.WHEN && p.tok != token.ELSE && p.tok != token.END && p.tok != token.EOF {
		stmt := p.parseStmt()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	
	return &ast.CaseClause{When: pos, List: list, Body: body}
}

func (p *Parser) parseGoStmt() *ast.GoStmt {
	pos := p.pos
	p.next() // consume 'go'
	
	expr := p.parseExpr(token.LowestPrec)
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		p.error(pos, "FATAL: Execution of 'go' must be followed by a function call. Add a valid function call (e.g., 'go myFunc()') after the keyword.")
		return nil
	}
	return &ast.GoStmt{Go: pos, Call: call}
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	pos := p.pos
	p.next() // consume 'defer'
	
	expr := p.parseExpr(token.LowestPrec)
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		p.error(pos, "FATAL: Execution of 'defer' must be followed by a function call. Add a valid function call after the keyword.")
		return nil
	}
	return &ast.DeferStmt{Defer: pos, Call: call}
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	pos := p.pos
	p.next() // consume 'return'

	var results []ast.Expr
	if p.tok != token.SEMICOLON && p.tok != token.END && p.tok != token.EOF {
		results = append(results, p.parseExpr(token.LowestPrec))
		for p.tok == token.COMMA {
			p.next()
			results = append(results, p.parseExpr(token.LowestPrec))
		}
	}

	if p.tok == token.SEMICOLON {
		p.next()
	}

	return &ast.ReturnStmt{Return: pos, Results: results}
}

func (p *Parser) parseLetStmt() *ast.LetStmt {
	pos := p.pos
	p.next() // consume 'let'
	
	if p.tok != token.IDENT {
		return nil
	}
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()
	
	var typ *ast.Ident
	if p.tok == token.IDENT {
		typ = &ast.Ident{NamePos: p.pos, Name: p.lit}
		p.next()
	}
	
	var val ast.Expr
	if p.tok == token.ASSIGN {
		p.next()
		val = p.parseExpr(token.LowestPrec)
	}
	
	return &ast.LetStmt{Let: pos, Name: name, Type: typ, Value: val}
}

func (p *Parser) parseIdent() ast.Expr {
	ident := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()
	return ident
}

func (p *Parser) parseBasicLit() ast.Expr {
	lit := &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
	p.next()
	return lit
}

func (p *Parser) parsePrefixExpr() ast.Expr {
	expr := &ast.PrefixExpr{OpPos: p.pos, Operator: p.tok.String()}
	p.next()
	expr.Right = p.parseExpr(token.UnaryPrec)
	return expr
}

func (p *Parser) parseInfixExpr(left ast.Expr) ast.Expr {
	expr := &ast.InfixExpr{Left: left, OpPos: p.pos, Operator: p.tok.String()}
	precedence := p.tok.Precedence()
	p.next()
	expr.Right = p.parseExpr(precedence)
	return expr
}

func (p *Parser) parseCallExpr(function ast.Expr) ast.Expr {
	exp := &ast.CallExpr{Fun: function, Lparen: p.pos}
	p.next() // consume '('
	exp.Args = p.parseCallArguments()
	exp.Rparen = p.pos
	if p.tok == token.RPAREN {
		p.next()
	}
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expr {
	var args []ast.Expr
	if p.tok == token.RPAREN {
		return args
	}
	
	args = append(args, p.parseExpr(token.LowestPrec))
	for p.tok == token.COMMA {
		p.next()
		args = append(args, p.parseExpr(token.LowestPrec))
	}
	return args
}

func (p *Parser) parseExpr(precedence int) ast.Expr {
	prefix := p.prefixParseFns[p.tok]
	if prefix == nil {
		p.next() // skip unknown
		return nil
	}
	leftExp := prefix()

	for p.tok != token.SEMICOLON && p.tok != token.EOF {
		tokPrec := p.tok.Precedence()
		if p.tok == token.LPAREN || p.tok == token.PERIOD || p.tok == token.DOUBLE_COLON || p.tok == token.LBRACE || p.tok == token.LBRACK {
			tokPrec = 7 // Call, Selector (including ::), CompositeLit, Index precedence tertinggi
		}
		if p.tok == token.COLON {
			tokPrec = 1 // KeyValue
		}
		
		if precedence >= tokPrec {
			break
		}
		
		infix := p.infixParseFns[p.tok]
		if infix == nil {
			return leftExp
		}
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseSelectorExpr(left ast.Expr) ast.Expr {
	isDoubleColon := (p.tok == token.DOUBLE_COLON)
	expr := &ast.SelectorExpr{X: left, IsDoubleColon: isDoubleColon, ColonPos: p.pos}
	p.next() // consume '.' or '::'
	if p.tok != token.IDENT {
		return nil
	}
	expr.Sel = &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()
	return expr
}

func (p *Parser) parseRequireDecl() *ast.RequireDecl {
	pos := p.pos
	reqLine := p.fset.Position(pos).Line
	p.next() // consume 'require'

	var pkgs []*ast.BasicLit

	// parse strings until we hit something else
	for p.tok == token.STRING || p.tok == token.SEMICOLON {
		if p.tok == token.SEMICOLON {
			p.next()
			continue
		}
		pkgs = append(pkgs, p.parseBasicLit().(*ast.BasicLit))
		
		// If the next token is semicolon or not string/retain, it might be a single inline require
		if p.tok == token.SEMICOLON {
			p.next()
			if p.tok != token.STRING && p.tok != token.RETAIN {
				break
			}
		}
	}

	if len(pkgs) == 0 {
		p.error(pos, "FATAL: 'require' block cannot be empty. Provide at least one package string.")
		return nil
	}

	var retainPos token.Pos
	if p.tok == token.RETAIN {
		retainPos = p.pos
		retLine := p.fset.Position(retainPos).Line
		p.next() // consume 'retain'
		if p.tok == token.SEMICOLON {
			p.next()
		}
		
		if reqLine == retLine {
			p.error(retainPos, "FATAL: Blood Lock - Inline 'require \"pkg\" retain' is STRICTLY FORBIDDEN. Either use 'require \"pkg\"' without retain, or use a multi-line block.")
		}
	} else if len(pkgs) > 1 {
		p.error(pos, "FATAL: Blood Lock - Multiple packages in a 'require' block is missing retain. Close the block with 'retain' on the last line.")
	}

	return &ast.RequireDecl{
		Require: pos,
		Pkgs:    pkgs,
		Retain:  retainPos,
	}
}

func (p *Parser) parseDecl() ast.Decl {
	switch p.tok {
	case token.DEF:
		if decl := p.parseFuncDecl(); decl != nil {
			return decl
		}
		return nil
	case token.REQUIRE:
		return p.parseRequireDecl()
	case token.STRUCT:
		return p.parseTypeDecl()
	case token.CLASS:
		return p.parseClassDecl()
	case token.LET:
		return p.parseLetStmt()
	case token.SEMICOLON:
		p.next()
		return nil
	default:
		p.next() // skip unknown
		return nil
	}
}

func (p *Parser) parseAtExpr() ast.Expr {
	pos := p.pos
	p.next() // consume '@'
	if p.tok != token.IDENT {
		p.error(pos, "FATAL: expected identifier after '@'. Provide a valid instance variable name after '@' (e.g., '@name').")
		return nil
	}
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()
	return &ast.AtExpr{At: pos, Name: name}
}

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	pos := p.pos
	p.next() // consume 'class'

	if p.tok != token.IDENT {
		p.error(pos, "FATAL: expected class name after 'class'. Provide a valid identifier name for the class.")
		return nil
	}
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next()

	var parent *ast.Ident
	if p.tok == token.LSS {
		p.next() // consume '<'
		if p.tok == token.IDENT {
			parent = &ast.Ident{NamePos: p.pos, Name: p.lit}
			p.next()
		}
	}

	if p.tok == token.SEMICOLON {
		p.next()
	}

	var fields []*ast.Field
	var methods []*ast.FuncDecl

	for p.tok != token.END && p.tok != token.EOF {
		if p.tok == token.DEF {
			if m := p.parseFuncDecl(); m != nil {
				methods = append(methods, m)
			}
		} else if p.tok == token.IDENT {
			fname := &ast.Ident{NamePos: p.pos, Name: p.lit}
			p.next()
			
			if p.tok == token.SEMICOLON || p.tok == token.END || p.tok == token.DEF || p.tok == token.EOF {
				// Anonymous embedded field
				fields = append(fields, &ast.Field{Type: fname})
			} else {
				ftype := p.parseExpr(token.LowestPrec)
				if ftype != nil {
					fields = append(fields, &ast.Field{Name: fname, Type: ftype})
				} else {
					// Fallback to anonymous field if parse failed
					fields = append(fields, &ast.Field{Type: fname})
				}
			}
			
			if p.tok == token.SEMICOLON {
				p.next()
			}
		} else if p.tok == token.SEMICOLON {
			p.next()
		} else {
			p.next()
		}
	}

	if p.tok == token.END {
		p.next() // consume 'end'
	}

	return &ast.ClassDecl{
		Class:   pos,
		Name:    name,
		Parent:  parent,
		Fields:  &ast.FieldList{List: fields},
		Methods: methods,
	}
}

func (p *Parser) parseFile() (*ast.File, error) {
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}
	if p.tok != token.CABINET {
		return nil, fmt.Errorf("expected 'cabinet', found '%s'", p.tok)
	}
	
	pos := p.pos
	p.next() // consume 'cabinet'
	
	if p.tok != token.IDENT {
		if len(p.errors) > 0 {
			return nil, p.errors[0]
		}
		return nil, fmt.Errorf("expected cabinet name, found '%s'", p.tok)
	}
	
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next() // consume name
	
	if p.tok == token.SEMICOLON {
		p.next()
	}
	
	var decls []ast.Decl
	for p.tok != token.EOF {
		if decl := p.parseDecl(); decl != nil {
			decls = append(decls, decl)
		}
	}
	
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return &ast.File{
		Cabinet: pos,
		Name:    name,
		Decls:   decls,
	}, nil
}

func (p *Parser) parseTypeDecl() *ast.TypeDecl {
	pos := p.pos
	p.next() // consume 'struct'
	
	if p.tok != token.IDENT {
		p.error(pos, fmt.Sprintf("FATAL: Expected struct name, found '%s'. Provide a valid identifier name for the struct.", p.lit))
		return nil
	}
	name := &ast.Ident{NamePos: p.pos, Name: p.lit}
	p.next() // consume name
	
	var fields []*ast.Field
	for p.tok != token.END && p.tok != token.EOF {
		if p.tok == token.SEMICOLON {
			p.next()
			continue
		}
		
		if p.tok == token.DEF {
			p.error(p.pos, "FATAL: methods cannot be declared inside a 'struct'. Use 'class' for internal methods, or declare methods outside the struct with a receiver.")
			p.next()
			continue
		}

		if p.tok == token.LSS {
			p.error(p.pos, "FATAL: inheritance '<' is forbidden for 'struct'. Use 'class B < A' for inheritance, or use anonymous field embedding inside 'struct'.")
			p.next()
			continue
		}
		
		if p.tok == token.IDENT {
			fieldName := &ast.Ident{NamePos: p.pos, Name: p.lit}
			p.next()
			
			if p.tok == token.SEMICOLON || p.tok == token.END || p.tok == token.EOF {
				// Anonymous embedded field
				fields = append(fields, &ast.Field{Type: fieldName})
			} else {
				fieldType := p.parseExpr(token.LowestPrec)
				if fieldType != nil {
					fields = append(fields, &ast.Field{Name: fieldName, Type: fieldType})
				} else {
					fields = append(fields, &ast.Field{Type: fieldName})
				}
			}
			
			if p.tok == token.SEMICOLON {
				p.next()
			}
		} else {
			p.error(p.pos, fmt.Sprintf("FATAL: expected field name or struct embedding, found '%s'. Provide a valid field name or type.", p.lit))
			p.next()
		}
	}
	
	if p.tok == token.END {
		p.next() // consume 'end'
	}
	if p.tok == token.SEMICOLON {
		p.next()
	}
	
	typeExpr := &ast.StructType{
		Struct: pos,
		Fields: &ast.FieldList{List: fields},
	}
	
	return &ast.TypeDecl{
		Type:     pos,
		Name:     name,
		TypeExpr: typeExpr,
	}
}

func (p *Parser) parseGroupedExpr() ast.Expr {
	p.next() // consume '('
	exp := p.parseExpr(token.LowestPrec)
	if p.tok == token.RPAREN {
		p.next()
	}
	return exp
}

func (p *Parser) parsePrefixCompositeLit() ast.Expr {
	expr := &ast.CompositeLit{Lbrace: p.pos}
	p.next() // consume '{'
	
	for p.tok != token.RBRACE && p.tok != token.EOF {
		expr.Elts = append(expr.Elts, p.parseExpr(token.LowestPrec))
		if p.tok == token.COMMA {
			p.next()
		}
	}
	
	expr.Rbrace = p.pos
	if p.tok == token.RBRACE {
		p.next() // consume '}'
	}
	return expr
}

func (p *Parser) parseCompositeLit(left ast.Expr) ast.Expr {
	expr := &ast.CompositeLit{Type: left, Lbrace: p.pos}
	p.next() // consume '{'
	
	for p.tok != token.RBRACE && p.tok != token.EOF {
		expr.Elts = append(expr.Elts, p.parseExpr(token.LowestPrec))
		if p.tok == token.COMMA {
			p.next()
		}
	}
	
	expr.Rbrace = p.pos
	if p.tok == token.RBRACE {
		p.next() // consume '}'
	}
	return expr
}

func (p *Parser) parseKeyValueExpr(left ast.Expr) ast.Expr {
	expr := &ast.KeyValueExpr{Key: left, Colon: p.pos}
	p.next() // consume ':'
	expr.Value = p.parseExpr(2)
	return expr
}

func (p *Parser) parseIndexExpr(left ast.Expr) ast.Expr {
	expr := &ast.IndexExpr{X: left, Lbrack: p.pos}
	p.next() // consume '['
	expr.Index = p.parseExpr(token.LowestPrec)
	expr.Rbrack = p.pos
	if p.tok == token.RBRACK {
		p.next()
	}
	return expr
}

func ParseFile(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	var p Parser
	p.init(fset, filename, src)
	return p.parseFile()
}
