package smt

import (
	"fmt"
)

type Identifier string
type Satisfiable int

const (
	Sat Satisfiable = iota
	Unsat
	Unknown
)

var (
	IntSort  = &SortName{"int"}
	BoolSort = &SortName{"bool"}
)

type Solver interface {
	Close()
	DeclareConst(id string, sort Sort) error
	Assert(t Term) error
	CheckSat() (Satisfiable, error)
	GetModel() (map[string]Term, error)
	Push()
	Pop() error
}

type Sort interface {
	sort()
}

type SortName struct {
	Id Identifier
}

type SortApp struct {
	Id   Identifier
	Args []Sort
}

type BitVecSort struct {
	Width int
}

func (*SortName) sort()   {}
func (*SortApp) sort()    {}
func (*BitVecSort) sort() {}

type Term interface {
	term()
}

type String struct {
	String string
}

type Int struct {
	Int int64
}

type BitVec struct {
	Value int64
	Width int64
}

type Const struct {
	Id Identifier
}

type App struct {
	Id   Identifier
	Args []Term
}

type Let struct {
	Id    Identifier
	Value Term
	In    Term
}

func (*String) term() {}
func (*Int) term()    {}
func (*BitVec) term() {}
func (*Const) term()  {}
func (*App) term()    {}
func (*Let) term()    {}

func NewInt(i int) Term {
	return &Int{int64(i)}
}

func NewBool(b bool) Term {
	if b {
		return &Const{"true"}
	} else {
		return &Const{"false"}
	}
}

func NewBitVec(n, id int64) Term {
	return &BitVec{n, id}
}

func NewConst(s string) Term {
	return &Const{Identifier(s)}
}

func NewApp(x string, args ...Term) Term {
	return &App{Identifier(x), args}
}

func Equals(a, b Term) Term {
	return NewApp("=", a, b)
}

func matchApp(t Term, id Identifier) *App {
	if app, ok := t.(*App); ok && app.Id == id {
		return app
	}
	return nil
}

func logical(op string, a, b Term) Term {
	opA := matchApp(a, Identifier(op))
	opB := matchApp(b, Identifier(op))

	var args []Term
	switch {
	case opA != nil && opB != nil:
		args = append(opA.Args, opB.Args...)
	case opA != nil:
		args = append(opA.Args, b)
	case opB != nil:
		args = append([]Term{b}, opB.Args...)
	default:
		args = []Term{a, b}
	}

	return NewApp(op, args...)
}

func And(a, b Term) Term {
	return logical("and", a, b)
}

func Or(a, b Term) Term {
	return logical("or", a, b)
}

func IfThenElse(e1, e2, e3 Term) Term {
	return NewApp("ite", e1, e2, e3)
}

func Implies(a, b Term) Term {
	return NewApp("=>", a, b)
}

func Add(a, b Term) Term {
	return NewApp("+", a, b)
}

func Sub(a, b Term) Term {
	return NewApp("-", a, b)
}

func Mul(a, b Term) Term {
	return NewApp("*", a, b)
}

func LT(a, b Term) Term {
	return NewApp("<", a, b)
}

func GT(a, b Term) Term {
	return NewApp(">", a, b)
}

func LTE(a, b Term) Term {
	return NewApp("<=", a, b)
}

func GTE(a, b Term) Term {
	return NewApp(">=", a, b)
}

func BVAdd(a, b Term) Term {
	return NewApp("bvadd", a, b)
}

func BVSub(a, b Term) Term {
	return NewApp("bvsub", a, b)
}

func BVMul(a, b Term) Term {
	return NewApp("bvmul", a, b)
}

func BVURem(a, b Term) Term {
	return NewApp("bvurem", a, b)
}

func BVSRem(a, b Term) Term {
	return NewApp("bvsrem", a, b)
}

func BVSMod(a, b Term) Term {
	return NewApp("bvsmod", a, b)
}

func BVShl(a, b Term) Term {
	return NewApp("bvshl", a, b)
}

func BVLShr(a, b Term) Term {
	return NewApp("bvlshr", a, b)
}

func BVAShr(a, b Term) Term {
	return NewApp("bvashr", a, b)
}

func BVOr(a, b Term) Term {
	return NewApp("bvor", a, b)
}

func BVAnd(a, b Term) Term {
	return NewApp("bvand", a, b)
}

func BVNand(a, b Term) Term {
	return NewApp("bvnand", a, b)
}

func BVNor(a, b Term) Term {
	return NewApp("bvnor", a, b)
}

func BVXNor(a, b Term) Term {
	return NewApp("bvxnor", a, b)
}

func BVUDiv(a, b Term) Term {
	return NewApp("bvudiv", a, b)
}

func BVSDiv(a, b Term) Term {
	return NewApp("bvsdiv", a, b)
}

func BVNeg(a Term) Term {
	return NewApp("bvneg", a)
}

func BVNot(a Term) Term {
	return NewApp("bvnot", a)
}

type Sexp interface {
	sexp()
	String() string
}

type SList struct {
	List []Sexp
}

type SSymbol struct {
	Symbol string
}

type SString struct {
	Str string
}

type SKeyword struct {
	Keyword string
}

type SInt struct {
	Int int
}

type SBitVec struct {
	Value int64
	Width int64
}

func (*SList) sexp()    {}
func (*SSymbol) sexp()  {}
func (*SString) sexp()  {}
func (*SKeyword) sexp() {}
func (*SInt) sexp()     {}
func (*SBitVec) sexp()  {}

func (s *SList) String() string {
	r := "("
	for _, child := range s.List {
		r += child.String()
		r += " "
	}
	r += ")"
	return r
}
func (s *SSymbol) String() string  { return s.Symbol }
func (s *SString) String() string  { return fmt.Sprintf(`"%s"`, s.Str) }
func (s *SKeyword) String() string { return fmt.Sprintf(":%s", s.Keyword) }
func (s *SInt) String() string     { return fmt.Sprintf("%v", s.Int) }
func (s *SBitVec) String() string  { return fmt.Sprintf("(_ bv%d %d)", s.Value, s.Width) }
