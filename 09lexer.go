package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

var keywords map[string]bool = map[string]bool{"select": true, "from": true, "where": true, "and": true, "insert": true, "into": true, "values": true, "delete": true, "update": true, "set": true, "create": true, "table": true, "varchar": true, "int": true, "view": true, "as": true, "index": true, "on": true}

type Lexer struct {
	keywords map[string]bool // might need map
	tok      scanner.Scanner
	currTok  rune // maybe unnecessary tbh
}

func makeLexer(s string) Lexer {
	var tok scanner.Scanner
	tok.Init(strings.NewReader(s))

	currTok := tok.Scan()

	return Lexer{keywords, tok, currTok}
}

func (l *Lexer) matchDelim(d rune) bool {
	return l.tok.TokenText() == string(d) // honestly no idea if that works
}
func (l *Lexer) matchIntConstant() bool {
	//fmt.Println("Checking ", l.tok.TokenText(), " for integer")
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
	//fmt.Println("checking ", l.tok.TokenText(), " for Id")
	_, ok := l.keywords[l.tok.TokenText()]
	if l.matchIntConstant() {
		return false
	}
	// todo check that it's actually a word
	return !ok
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
func lexerTest() {
	x := ""
	y := 0
	ok := true

	src := "A = 1\n3 = B"

	sc := bufio.NewScanner(strings.NewReader(src))

	for sc.Scan() {
		lex := makeLexer(sc.Text())
		fmt.Println("looking at line ", sc.Text())
		if lex.matchId() {
			x, ok = lex.eatId()
			if !ok {
				fmt.Println("thinks first part is Id but does not get it")
			}
			lex.eatDelim('=')
			y, ok = lex.eatIntConstant()
			if !ok {
				fmt.Println("thinks second part is int constant but does not get it")
			}
		} else {
			y, ok = lex.eatIntConstant()
			if !ok {
				fmt.Println("thinks first part is int constant but does not get it")
			}
			lex.eatDelim('=')
			x, ok = lex.eatId()
			if !ok {
				fmt.Println("thinks second part is Id but does not get it")
			}
		}
		fmt.Println(x, " equals ", y)
	}

	// a := "A = 1"
	// delim := '='

	// var tok scanner.Scanner
	// tok.Init(strings.NewReader(a))
	// tok.Scan()

	// // print and eat id (A)
	// fmt.Println("first token:", tok.TokenText())
	// tok.Scan()

	// tt := tok.TokenText()
	// fmt.Println("second token:", tok.TokenText())
	// fmt.Println("with peek:", tok.Peek())
	// fmt.Println("with next:", tok.Next())

	// fmt.Println("delimiter:", delim)
	// fmt.Println("delimiter as string:", string(delim))

	// fmt.Println("token contains delim rune?", strings.ContainsRune(tt, delim))
	// fmt.Println("token equal to string(delim)?", tt == string(delim))

}
