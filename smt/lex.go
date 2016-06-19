// Copyright 2016 Bobby Powers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smt

import (
	"bytes"
	"fmt"
	"go/token"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:generate go tool yacc -o parse.go -p smt parse.y

const eof = 0

type iType int

const (
	iEOF iType = iota
	iInt
	iHex
	iSymbol
	iString
	iKeyword
	iLParen
	iRParen
)

type tok struct {
	pos    token.Pos
	val    string
	ival   int64
	kind   iType
	yyKind int
}

type stateFn func(*smtLex) stateFn

type smtLex struct {
	f     *token.File
	s     string // the string being scanned
	pos   int    // current position in the input
	start int    // start of this token
	width int    // width of the last rune
	last  tok
	items chan tok // channel of scanned items
	state stateFn

	result *[]Sexp
}

func (l *smtLex) Lex(lval *smtSymType) int {
	for {
		select {
		case item := <-l.items:
			lval.tok = item
			return item.yyKind
		default:
			l.state = l.state(l)
		}
	}
	panic("unreachable")
}

func newSmtLex(input string, file *token.File, result *[]Sexp) *smtLex {
	return &smtLex{
		f:      file,
		s:      input,
		items:  make(chan tok, 2),
		state:  lexStatement,
		result: result,
	}
}

func (l *smtLex) getLine(pos token.Position) string {
	p := pos.Offset - pos.Column
	if p < 0 || p >= len(l.s) {
		return fmt.Sprintf("getLine: o%d c%d, len%d",
			pos.Offset, pos.Column, len(l.s))
	}
	result := l.s[pos.Offset-pos.Column:]
	if newline := strings.IndexRune(result, '\n'); newline != -1 {
		result = result[:newline]
	}
	return result
}

func (l *smtLex) Error(s string) {
	pos := l.f.Position(l.last.pos)
	line := l.getLine(pos)
	// we want the number of spaces (taking into account tabs)
	// before the problematic token
	prefixLen := pos.Column + strings.Count(line[:pos.Column], "\t")*7 - 1
	prefix := strings.Repeat(" ", prefixLen)

	line = strings.Replace(line, "\t", "        ", -1)

	fmt.Printf("%s:%d:%d: error: %s\n", pos.Filename,
		pos.Line, pos.Column, s)
	fmt.Printf("%s\n", line)
	fmt.Printf("%s^\n", prefix)
}

func (l *smtLex) next() rune {
	if l.pos >= len(l.s) {
		return 0
	}
	r, width := utf8.DecodeRuneInString(l.s[l.pos:])
	l.pos += width
	l.width = width

	if r == '\n' {
		l.f.AddLine(l.pos + 1)
	}
	return r
}

func (l *smtLex) backup() {
	l.pos -= l.width
}

func (l *smtLex) peek() rune {
	peek := l.next()
	l.backup()
	return peek
}

func (l *smtLex) ignore() {
	l.start = l.pos
}

func (l *smtLex) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *smtLex) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *smtLex) emit(yyTy rune, ty iType) {
	t := tok{
		pos:    l.f.Pos(l.pos),
		val:    l.s[l.start:l.pos],
		yyKind: int(yyTy),
		kind:   ty,
	}
	//log.Printf("t: %#v\n", t)
	l.last = t
	l.items <- t
	l.ignore()
}

func (l *smtLex) errorf(format string, args ...interface{}) stateFn {
	log.Printf(format, args...)
	l.emit(eof, iEOF)
	return nil
}

func lexStatement(l *smtLex) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(eof, iEOF)
	case r == '/':
		if l.peek() == '/' {
			l.next()
			return lexComment
		}
		if l.peek() == '*' {
			l.next()
			return lexMultiComment
		}
		return lexSymbol
	case unicode.IsSpace(r):
		//	log.Print("1 ignoring:", l.s[l.start:l.pos])
		l.ignore()
	case unicode.IsDigit(r):
		l.backup()
		return lexInteger
	case isKeywordStart(r):
		l.backup()
		return lexKeyword
	case r == '"':
		l.backup()
		return lexString
	}

	return lexSymbol
}

func lexOperator(l *smtLex) stateFn {
	var ty iType
	r := l.next()
	switch {
	case r == '(':
		ty = iLParen
	case r == ')':
		ty = iRParen
	default:
		panic("unknown operator type")
	}
	l.emit(r, ty)
	return lexStatement
}

func lexComment(l *smtLex) stateFn {
	// skip everything until the end of the line, or the end of
	// the file, whichever is first
	for r := l.next(); r != '\n' && r != eof; r = l.next() {
	}
	l.backup()
	//	log.Print("2 ignoring:", l.s[l.start:l.pos])
	l.ignore()
	return lexStatement
}

func lexMultiComment(l *smtLex) stateFn {
	// skip everything until the end of the line, or the end of
	// the file, whichever is first
	for r := l.next(); ; r = l.next() {
		if r == eof {
			l.backup()
			break
		}
		if r != '*' {
			continue
		}
		if l.peek() == '/' {
			l.next()
			break
		}
	}
	//	log.Print("2 ignoring:", l.s[l.start:l.pos])
	l.ignore()
	return lexStatement
}

func lexInteger(l *smtLex) stateFn {
	l.acceptRun("0123456789")
	l.emit(YINT, iInt)
	return lexStatement
}

func lexString(l *smtLex) stateFn {
	delim := l.next()
	l.ignore()
	for r := l.next(); r != delim && r != eof; r = l.next() {
	}
	l.backup()

	if l.peek() != delim {
		return l.errorf("unexpected EOF")
	}
	l.emit(YSTRING, iString)
	l.next()
	l.ignore()
	return lexStatement
}

func lexKeyword(l *smtLex) stateFn {
	_ = l.next() // ignore leading ':'
	l.ignore()
	for !unicode.IsSpace(l.next()) {
	}
	l.backup()
	l.emit(YKEYWORD, iKeyword)
	return lexStatement
}

func lexSymbol(l *smtLex) stateFn {
	for !unicode.IsSpace(l.next()) {
	}
	l.backup()
	l.emit(YSYMBOL, iSymbol)
	return lexStatement
}

func isStringStart(r rune) bool {
	return r == '"'
}

func isKeywordStart(r rune) bool {
	return r == ':'
}

func isOperator(r rune) bool {
	return bytes.IndexRune([]byte("()"), r) > -1
}
