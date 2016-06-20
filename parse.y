// Copyright 2016 Bobby Powers. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

%{

package smt

import (
	"strconv"
)

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union{
	sexps []Sexp
	sexp  Sexp

	tok    tok
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct
%type <sexps>  sexp_list
%type <sexp>   sexp top

// same for terminals
%token <tok> yINT yHEX ySTRING ySYMBOL yKEYWORD

%%

top:	{
	}
|	top sexp
	{
		smtlex.(*smtLex).parser.emit($2)
	}
;

sexp_list:
	{
		$$ = []Sexp{}
	}
|	sexp_list sexp
	{
		$$ = append($1, $2)
	}
;

sexp:	yINT
	{
		i, _ := strconv.ParseInt($1.val, 10, 0)
		$$ = &SInt{i}
	}
|	ySTRING
	{
		$$ = &SString{$1.val}
	}
|	ySYMBOL
	{
		$$ = &SSymbol{$1.val}
	}
|	yKEYWORD
	{
		$$ = &SKeyword{$1.val}
	}
|	'(' sexp_list ')'
	{
		$$ = &SList{$2}
	}
;
