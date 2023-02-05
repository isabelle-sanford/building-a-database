package main

var keywords []string = []string{"select", "from", "where", "and", "insert", "into", "values", "delete",
	"update", "set", "create", "table", "varchar", "int", "view", "as", "index", "on"}

type Lexer struct {
	keywords []string // might need map
	// ! streamtokenizer
}

type PredParser struct {
	lex Lexer
}

type Parser struct {
	lex Lexer
}

func makeLexer(s string) Lexer

func (l *Lexer) matchDelim(d int) bool // int should be char
func (l *Lexer) matchIntConstant() bool
func (l *Lexer) matchStringConstant() bool
func (l *Lexer) matchKeyword(w string) bool
func (l *Lexer) matchId() bool

func (l *Lexer) eatDelim(d int) // 'int' is char
func (l *Lexer) eatIntConstant() int
func (l *Lexer) eatStringConstant() string
func (l *Lexer) eatKeyword(w string)
func (l *Lexer) eatId() string

func (l *Lexer) nextToken()

// func initKeywords() []string {
// 	keywords := []string{"select","from","where","and","insert","into","values","delete",
// 		"update","set","create","table","varchar","int","view","as","index","on"}
// 		return keywords
// }

// PRED PARSER

func (pp *PredParser) field() {
	pp.lex.eatId()
}

func (pp *PredParser) constant() {
	if pp.lex.matchStringConstant() {
		pp.lex.matchIntConstant()
	} else {
		pp.lex.eatIntConstant()
	}
}

func (pp *PredParser) expression() {
	if pp.lex.matchId() {
		pp.field()
	} else {
		pp.constant()
	}
}

func (pp *PredParser) term() {
	pp.expression()
	pp.lex.eatDelim('=')
	pp.expression()
}

func (pp *PredParser) predicate() {
	pp.term()
	if pp.lex.matchKeyword("and") {
		pp.lex.keyword("and") // ???
		pp.predicate()
	}
}

// PARSER

func makeParser(s string) Parser {
	p := Parser{makeLexer(s)}
	return p
}

func (p *Parser) field() string {
	return p.lex.eatId()
}

func (p *Parser) constant() Constant {
	if p.lex.matchStringConstant() {
		return makeConstString(p.lex.eatStringConstant())
	} else {
		return makeConstInt(p.lex.eatIntConstant())
	}
}

func (p *Parser) expression() Expression {
	if p.lex.matchId() {
		return makeExprFld(p.field())
	} else {
		return makeExprConst(p.constant())
	}
}

func (p *Parser) term() Term {
	lhs := p.expression()
	p.lex.eatDelim('=')
	rhs := p.expression()
	return Term{lhs, rhs}
}

func (p *Parser) predicate() *Predicate {
	pred := makePredwTerm(p.term())
	if p.lex.matchKeyword("and") {
		p.lex.eatKeyword("and")
		pred.conjoinWith(p.predicate())
	}
	return &pred
}

func (p *Parser) query() QueryData {
	p.lex.eatKeyword("select")
	fields := p.selectList()
	p.lex.eatKeyword("from")
	tables := p.tableList()
	var pred *Predicate
	if p.lex.matchKeyword("where") {
		p.lex.eatKeyword("where")
		pred = p.predicate()
	}
	return QueryData{fields, tables, pred}
}

func (p *Parser) selectList() []string {
	L := []string{p.field()}
	if p.lex.matchDelim(',') {
		p.lex.eatDelim(',')
		L = append(L, p.selectList()...)
	}
	return L
}

func (p *Parser) tableList() []string {
	L := []string{p.lex.eatId()}
	if p.lex.matchDelim(',') {
		p.lex.eatDelim(',')
		L = append(L, p.tableList()...)
	}
	return L
}

func (p *Parser) updateCmd() { // returns something ???
	if p.lex.matchKeyword("insert") {
		return p.insert()
	} else if p.lex.matchKeyword("delete") {
		return p.delete()
	} else if p.lex.matchKeyword("update") {
		return p.modify()
	} else {
		return p.create()
	}
}

func (p *Parser) create() { // returns ??
	p.lex.eatKeyword("create")
	if p.lex.matchKeyword("table") {
		return createTable()
	} else if p.lex.matchKeyword("view") {
		return createView() // might not be valid
	} else {
		return createIndex() // might not be valid
	}
}

func (p *Parser) delete() DeleteData {
	p.lex.eatKeyword("delete")
	p.lex.eatKeyword("from")
	tblname := p.lex.eatId()
	var pred Predicate
	if p.lex.matchKeyword("where") {
		lex.eatKeyword("where")
		pred = *p.predicate()
	}
	return DeleteData{tblname, pred}
}

func (p *Parser) insert() InsertData {
	p.lex.eatKeyword("insert")
	p.lex.eatKeyword("into")
	tblname := p.lex.eatId()
	p.lex.eatDelim('(')
	fields := p.fieldList()
	p.lex.eatDelim(')')
	p.lex.eatKeyword("values")
	p.lex.eatDelim('(')
	vals := p.constList()
	p.lex.eatDelim(')')
	return InsertData{tblname, fields, vals}
}

func (p *Parser) fieldList() []string {
	L := []string{p.field()}
	if p.lex.matchDelim(',') {
		p.lex.eatDelim(',')
		L = append(L, p.fieldList()...)
	}
	return L
}

func (p *Parser) constList() []Constant {
	L := []Constant{p.constant()}
	if p.lex.matchDelim(',') {
		p.lex.eatDelim(',')
		L = append(L, p.constList()...)
	}
	return L
}

func (p *Parser) modify() ModifyData {
	p.lex.eatKeyword("update")
	tblname := p.lex.eatId()
	p.lex.eatKeyword("set")
	fldname := p.field()
	p.lex.eatDelim('=')
	newval := p.expression()
	var pred Predicate
	if p.lex.matchKeyword("where") {
		p.lex.eatKeyword("where")
		pred := p.predicate()
	}
	return ModifyData{tblname, fldname, newval, pred}
}

func (p *Parser) createTable() CreateTableData {
	p.lex.eatKeyword("table")
	tblname := p.lex.eatId()
	p.lex.eatDelim('(')
	sch := p.fieldDefs()
	p.lex.eatDelim(')')
	return CreateTableData{tblname, sch}
}

func (p *Parser) fieldDefs() Schema {
	schema := p.fieldDef()
	if p.lex.matchDelim(',') {
		p.lex.eatDelim(',')
		sch2 := p.fieldDefs()
		schema.addAll(sch2)
	}
	return schema
}

func (p *Parser) fieldDef() Schema {
	fldname := p.field()
	return p.fieldType(fldname)
}

func (p *Parser) fieldType(fldname string) Schema {
	schema := makeSchema()
	if p.lex.matchKeyword("int") {
		p.lex.eatKeyword("int")
		schema.addIntField(fldname)
	} else {
		p.lex.eatKeyword("varchar")
		p.lex.eatDelim('(')
		strlen := p.lex.eatIntConstant()
		p.lex.eatDelim(')')
		schema.addStringField(fldname, strlen)
	}
	return schema
}

func (p *Parser) createView() CreateViewData
func (p *Parser) createIndex() CreateIndexData
