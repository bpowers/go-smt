package smt

import (
	"errors"
	"fmt"
	"io"
)

var ParserEOF = errors.New("End-of-Input")

type Parser struct {
	sexps chan Sexp
	errs  chan error
}

func (p *Parser) Read() (Sexp, error) {
	select {
	case s := <-p.sexps:
		return s, nil
	case err := <-p.errs:
		return nil, err
	}
}

func (p *Parser) emit(s Sexp) {
	p.sexps <- s
}

func NewParser(r io.Reader) *Parser {
	p := &Parser{
		sexps: make(chan Sexp),
		errs:  make(chan error),
	}

	go p.streamingParse(r)

	return p
}

func (p *Parser) streamingParse(r io.Reader) {
	// this is weird, but without passing in a reference to this
	// parser object through the lexer, there isn't another good
	// way to keep the parser and lexer reentrant.
	err := smtParse(newSmtLex(r, p))
	if err != 0 {
		p.errs <- fmt.Errorf("%d parse errors", err)
	} else {
		p.errs <- ParserEOF
	}
}
