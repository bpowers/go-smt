// Copyright 2016 Bobby Powers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smt

import (
	"bufio"
	"bytes"
	"fmt"
	"go/token"
	"io"
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
	pos    int
	val    string
	ival   int64
	kind   iType
	yyKind int
}

type stateFn func(*smtLex) stateFn

type smtLex struct {
	in    *bufio.Scanner
	line  string
	pos   int // current position in the input
	start int // start of this token
	width int // width of the last rune
	last  tok
	items chan tok // channel of scanned items
	state stateFn

	parser *Parser
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
}

func newSmtLex(r io.Reader, p *Parser) *smtLex {
	return &smtLex{
		in:     bufio.NewScanner(r),
		items:  make(chan tok, 2),
		state:  lexStatement,
		parser: p,
	}
}

func (l *smtLex) getLine(pos token.Position) string {
	col := pos.Column
	if col > pos.Offset {
		col = pos.Offset
	}
	p := pos.Offset - col
	if p >= len(l.line) {
		return fmt.Sprintf("getLine: o%d c%d, len%d",
			pos.Offset, col, len(l.line))
	}
	result := l.line[pos.Offset-col:]
	if newline := strings.IndexRune(result, '\n'); newline != -1 {
		result = result[:newline]
	}
	return result
}

func (l *smtLex) Error(s string) {
	line := l.line

	col := l.pos
	if col > len(line) {
		col = len(line)
	}
	// we want the number of spaces (taking into account tabs)
	// before the problematic token
	prefixLen := col + strings.Count(line[:col], "\t")*7 - 1
	prefix := strings.Repeat(" ", prefixLen)

	line = strings.Replace(line, "\t", "        ", -1)

	fmt.Printf("%d: error: %s\n", col, s)
	fmt.Printf("%s\n", line)
	fmt.Printf("%s^\n", prefix)
}

func (l *smtLex) next() rune {
	if l.pos >= len(l.line) {
		ok := l.in.Scan()
		if !ok {
			l.width = 0
			return 0
		}
		l.line = l.in.Text() + "\n"
		l.pos = 0
		l.start = 0
	}
	r, width := utf8.DecodeRuneInString(l.line[l.pos:])
	l.pos += width
	l.width = width

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
	if l.start == l.pos {
		log.Printf("ignore: start == pos (%d), error.", l.start)
	}
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
		pos:    l.pos,
		val:    l.line[l.start:l.pos],
		yyKind: int(yyTy),
		kind:   ty,
	}
	//log.Printf("t(%s): %#v\n", l.line[l.start:l.pos], t)
	l.last = t
	l.items <- t
	if ty != iEOF {
		l.ignore()
	}
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
	case isOperator(r):
		l.backup()
		return lexOperator
	case isKeywordStart(r):
		return lexKeyword
	case r == '"':
		l.backup()
		return lexString
	default:
		return lexSymbol
	}

	return lexStatement
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
	l.emit(yINT, iInt)
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
	l.emit(ySTRING, iString)
	l.next()
	l.ignore()
	return lexStatement
}

func lexKeyword(l *smtLex) stateFn {
	l.ignore()
	for r := l.next(); r != eof && !isOperator(r) && !unicode.IsSpace(r); r = l.next() {
	}
	l.backup()
	l.emit(yKEYWORD, iKeyword)
	return lexStatement
}

func lexSymbol(l *smtLex) stateFn {
	for r := l.next(); r != eof && !isOperator(r) && !unicode.IsSpace(r); r = l.next() {
	}
	l.backup()
	l.emit(ySYMBOL, iSymbol)
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
