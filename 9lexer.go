package main

import (
	"strconv"
	"strings"
	"text/scanner"
)

var keywords map[string]bool = map[string]bool{"select": true, "from": true, "where": true, "and": true, "insert": true, "into": true, "values": true, "delete": true, "update": true, "set": true, "create": true, "table": true, "varchar": true, "int": true, "view": true, "as": true, "index": true, "on": true}

type Lexer struct {
	keywords map[string]bool // might need map
	tok      scanner.Scanner
	currTok  rune
}

func makeLexer(s string) Lexer {
	var tok scanner.Scanner
	tok.Init(strings.NewReader(s))

	currTok := tok.Scan()

	return Lexer{keywords, tok, currTok}
}

func (l *Lexer) matchDelim(d rune) bool { // int should be char
	return l.tok.Peek() == d // honestly no idea if that works
}
func (l *Lexer) matchIntConstant() bool {
	_, err := strconv.Atoi(l.tok.TokenText())
	if err == nil {
		return true
	}
	return false // i guess
}
func (l *Lexer) matchStringConstant() bool {
	return strconv.QuoteRune(l.tok.Peek()) == "'"
}
func (l *Lexer) matchKeyword(w string) bool {
	return l.tok.TokenText() == w
}
func (l *Lexer) matchId() bool {
	_, ok := l.keywords[l.tok.TokenText()]
	// todo check that it's actually a word
	return ok
}

func (l *Lexer) eatDelim(d rune) (ok bool) {
	ok = true
	if !l.matchDelim(d) {
		ok = false
	} else {
		l.tok.Scan()
	}
	return
}
func (l *Lexer) eatIntConstant() (i int, ok bool) {
	ok = true
	if !l.matchIntConstant() {
		ok = false
		return
	} else {
		i, _ = strconv.Atoi(l.tok.TokenText())
		l.tok.Scan()
		return
	}
}
func (l *Lexer) eatStringConstant() (s string, ok bool) {
	ok = true
	if !l.matchStringConstant() {
		ok = false
		return
	} else {
		s = l.tok.TokenText()
		l.tok.Scan()
		return
	}
}
func (l *Lexer) eatKeyword(w string) (ok bool) {
	ok = true
	if !l.matchKeyword(w) {
		ok = false
	} else {
		l.tok.Scan()
	}
	return
}
func (l *Lexer) eatId() (id string, ok bool) {
	ok = true
	if !l.matchId() {
		ok = false
	} else {
		id = l.tok.TokenText()
		l.tok.Scan()
	}
	return
}

func (l *Lexer) nextToken() {
	if l.tok.Next() != scanner.EOF {
		l.tok.Scan()
	} else {
		//! error
	}
}

// page 243
func LexerTest() {
	x := ""
	y := 0

}

// func initKeywords() []string {
// 	keywords := []string{"select","from","where","and","insert","into","values","delete",
// 		"update","set","create","table","varchar","int","view","as","index","on"}
// 		return keywords
// }

// PRED PARSER
