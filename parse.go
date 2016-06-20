//line parse.y:6
package smt

import __yyfmt__ "fmt"

//line parse.y:7
import (
	"strconv"
)

//line parse.y:17
type smtSymType struct {
	yys   int
	sexps []Sexp
	sexp  Sexp

	tok tok
}

const yINT = 57346
const yHEX = 57347
const ySTRING = 57348
const ySYMBOL = 57349
const yKEYWORD = 57350

var smtToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"yINT",
	"yHEX",
	"ySTRING",
	"ySYMBOL",
	"yKEYWORD",
	"'('",
	"')'",
}
var smtStatenames = [...]string{}

const smtEofCode = 1
const smtErrCode = 2
const smtInitialStackSize = 16

//line yacctab:1
var smtExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const smtNprod = 10
const smtPrivate = 57344

var smtTokenNames []string
var smtStates []string

const smtLast = 18

var smtAct = [...]int{

	3, 1, 4, 5, 6, 7, 10, 3, 2, 4,
	5, 6, 7, 8, 0, 0, 0, 9,
}
var smtPact = [...]int{

	-1000, 3, -1000, -1000, -1000, -1000, -1000, -1000, -4, -1000,
	-1000,
}
var smtPgo = [...]int{

	0, 13, 8, 1,
}
var smtR1 = [...]int{

	0, 3, 3, 1, 1, 2, 2, 2, 2, 2,
}
var smtR2 = [...]int{

	0, 0, 2, 0, 2, 1, 1, 1, 1, 3,
}
var smtChk = [...]int{

	-1000, -3, -2, 4, 6, 7, 8, 9, -1, -2,
	10,
}
var smtDef = [...]int{

	1, -2, 2, 5, 6, 7, 8, 3, 0, 4,
	9,
}
var smtTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	9, 10,
}
var smtTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8,
}
var smtTok3 = [...]int{
	0,
}

var smtErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	smtDebug        = 0
	smtErrorVerbose = false
)

type smtLexer interface {
	Lex(lval *smtSymType) int
	Error(s string)
}

type smtParser interface {
	Parse(smtLexer) int
	Lookahead() int
}

type smtParserImpl struct {
	lval  smtSymType
	stack [smtInitialStackSize]smtSymType
	char  int
}

func (p *smtParserImpl) Lookahead() int {
	return p.char
}

func smtNewParser() smtParser {
	return &smtParserImpl{}
}

const smtFlag = -1000

func smtTokname(c int) string {
	if c >= 1 && c-1 < len(smtToknames) {
		if smtToknames[c-1] != "" {
			return smtToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func smtStatname(s int) string {
	if s >= 0 && s < len(smtStatenames) {
		if smtStatenames[s] != "" {
			return smtStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func smtErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !smtErrorVerbose {
		return "syntax error"
	}

	for _, e := range smtErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + smtTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := smtPact[state]
	for tok := TOKSTART; tok-1 < len(smtToknames); tok++ {
		if n := base + tok; n >= 0 && n < smtLast && smtChk[smtAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if smtDef[state] == -2 {
		i := 0
		for smtExca[i] != -1 || smtExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; smtExca[i] >= 0; i += 2 {
			tok := smtExca[i]
			if tok < TOKSTART || smtExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if smtExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += smtTokname(tok)
	}
	return res
}

func smtlex1(lex smtLexer, lval *smtSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = smtTok1[0]
		goto out
	}
	if char < len(smtTok1) {
		token = smtTok1[char]
		goto out
	}
	if char >= smtPrivate {
		if char < smtPrivate+len(smtTok2) {
			token = smtTok2[char-smtPrivate]
			goto out
		}
	}
	for i := 0; i < len(smtTok3); i += 2 {
		token = smtTok3[i+0]
		if token == char {
			token = smtTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = smtTok2[1] /* unknown char */
	}
	if smtDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", smtTokname(token), uint(char))
	}
	return char, token
}

func smtParse(smtlex smtLexer) int {
	return smtNewParser().Parse(smtlex)
}

func (smtrcvr *smtParserImpl) Parse(smtlex smtLexer) int {
	var smtn int
	var smtVAL smtSymType
	var smtDollar []smtSymType
	_ = smtDollar // silence set and not used
	smtS := smtrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	smtstate := 0
	smtrcvr.char = -1
	smttoken := -1 // smtrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		smtstate = -1
		smtrcvr.char = -1
		smttoken = -1
	}()
	smtp := -1
	goto smtstack

ret0:
	return 0

ret1:
	return 1

smtstack:
	/* put a state and value onto the stack */
	if smtDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", smtTokname(smttoken), smtStatname(smtstate))
	}

	smtp++
	if smtp >= len(smtS) {
		nyys := make([]smtSymType, len(smtS)*2)
		copy(nyys, smtS)
		smtS = nyys
	}
	smtS[smtp] = smtVAL
	smtS[smtp].yys = smtstate

smtnewstate:
	smtn = smtPact[smtstate]
	if smtn <= smtFlag {
		goto smtdefault /* simple state */
	}
	if smtrcvr.char < 0 {
		smtrcvr.char, smttoken = smtlex1(smtlex, &smtrcvr.lval)
	}
	smtn += smttoken
	if smtn < 0 || smtn >= smtLast {
		goto smtdefault
	}
	smtn = smtAct[smtn]
	if smtChk[smtn] == smttoken { /* valid shift */
		smtrcvr.char = -1
		smttoken = -1
		smtVAL = smtrcvr.lval
		smtstate = smtn
		if Errflag > 0 {
			Errflag--
		}
		goto smtstack
	}

smtdefault:
	/* default state action */
	smtn = smtDef[smtstate]
	if smtn == -2 {
		if smtrcvr.char < 0 {
			smtrcvr.char, smttoken = smtlex1(smtlex, &smtrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if smtExca[xi+0] == -1 && smtExca[xi+1] == smtstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			smtn = smtExca[xi+0]
			if smtn < 0 || smtn == smttoken {
				break
			}
		}
		smtn = smtExca[xi+1]
		if smtn < 0 {
			goto ret0
		}
	}
	if smtn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			smtlex.Error(smtErrorMessage(smtstate, smttoken))
			Nerrs++
			if smtDebug >= 1 {
				__yyfmt__.Printf("%s", smtStatname(smtstate))
				__yyfmt__.Printf(" saw %s\n", smtTokname(smttoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for smtp >= 0 {
				smtn = smtPact[smtS[smtp].yys] + smtErrCode
				if smtn >= 0 && smtn < smtLast {
					smtstate = smtAct[smtn] /* simulate a shift of "error" */
					if smtChk[smtstate] == smtErrCode {
						goto smtstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if smtDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", smtS[smtp].yys)
				}
				smtp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if smtDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", smtTokname(smttoken))
			}
			if smttoken == smtEofCode {
				goto ret1
			}
			smtrcvr.char = -1
			smttoken = -1
			goto smtnewstate /* try again in the same state */
		}
	}

	/* reduction by production smtn */
	if smtDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", smtn, smtStatname(smtstate))
	}

	smtnt := smtn
	smtpt := smtp
	_ = smtpt // guard against "declared and not used"

	smtp -= smtR2[smtn]
	// smtp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if smtp+1 >= len(smtS) {
		nyys := make([]smtSymType, len(smtS)*2)
		copy(nyys, smtS)
		smtS = nyys
	}
	smtVAL = smtS[smtp+1]

	/* consult goto table to find next state */
	smtn = smtR1[smtn]
	smtg := smtPgo[smtn]
	smtj := smtg + smtS[smtp].yys + 1

	if smtj >= smtLast {
		smtstate = smtAct[smtg]
	} else {
		smtstate = smtAct[smtj]
		if smtChk[smtstate] != -smtn {
			smtstate = smtAct[smtg]
		}
	}
	// dummy call; replaced with literal code
	switch smtnt {

	case 1:
		smtDollar = smtS[smtpt-0 : smtpt+1]
		//line parse.y:34
		{
		}
	case 2:
		smtDollar = smtS[smtpt-2 : smtpt+1]
		//line parse.y:37
		{
			smtlex.(*smtLex).parser.emit(smtDollar[2].sexp)
		}
	case 3:
		smtDollar = smtS[smtpt-0 : smtpt+1]
		//line parse.y:43
		{
			smtVAL.sexps = []Sexp{}
		}
	case 4:
		smtDollar = smtS[smtpt-2 : smtpt+1]
		//line parse.y:47
		{
			smtVAL.sexps = append(smtDollar[1].sexps, smtDollar[2].sexp)
		}
	case 5:
		smtDollar = smtS[smtpt-1 : smtpt+1]
		//line parse.y:53
		{
			i, _ := strconv.ParseInt(smtDollar[1].tok.val, 10, 0)
			smtVAL.sexp = &SInt{i}
		}
	case 6:
		smtDollar = smtS[smtpt-1 : smtpt+1]
		//line parse.y:58
		{
			smtVAL.sexp = &SString{smtDollar[1].tok.val}
		}
	case 7:
		smtDollar = smtS[smtpt-1 : smtpt+1]
		//line parse.y:62
		{
			smtVAL.sexp = &SSymbol{smtDollar[1].tok.val}
		}
	case 8:
		smtDollar = smtS[smtpt-1 : smtpt+1]
		//line parse.y:66
		{
			smtVAL.sexp = &SKeyword{smtDollar[1].tok.val}
		}
	case 9:
		smtDollar = smtS[smtpt-3 : smtpt+1]
		//line parse.y:70
		{
			smtVAL.sexp = &SList{smtDollar[2].sexps}
		}
	}
	goto smtstack /* stack new state and value */
}
