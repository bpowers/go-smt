package solver

import (
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/bpowers/go-smt"
)

func isSuccess(sexp smt.Sexp) bool {
	return smt.IsSymbol(sexp, "success")
}

func NewPipedSolver(exe string, args ...string) (smt.Solver, error) {
	cmd := exec.Command(exe, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("StdinPipe: %s", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("StdoutPipe: %s", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("Run: %s", err)
	}

	s := &solver{
		cmd:          cmd,
		stdin:        stdin,
		stdoutCloser: stdout,
		results:      smt.NewParser(stdout),
	}

	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"set-option"},
		&smt.SKeyword{"print-success"},
		&smt.SSymbol{"true"}}})
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("print-success: %s", err)
	}
	if !isSuccess(r) {
		s.Close()
		return nil, fmt.Errorf("print-success(%#v): %s", r, err)
	}

	return s, nil
}

type solver struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdoutCloser io.Closer
	results      *smt.Parser
}

func (s *solver) Command(sexp smt.Sexp) (smt.Sexp, error) {

	str := sexp.String()
	n, err := s.stdin.Write([]byte(str))

	if n != len(str) {
		return nil, fmt.Errorf("stdin.Write: short (%d < %d)", n, len(str))
	}
	if err != nil {
		return nil, fmt.Errorf("stdin.Write: %s", err)
	}

	result, err := s.results.Read()
	if err != nil {
		return nil, fmt.Errorf("Parser.Read: %s", err)
	}

	return result, nil
}

func (s *solver) Close() {

}

func (s *solver) DeclareConst(id string, sort smt.Sort) error {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"declare-const"},
		&smt.SSymbol{id},
		smt.SortToSexp(sort)}})
	if err != nil {
		return fmt.Errorf("Command: %s", err)
	}
	if !isSuccess(r) {
		return fmt.Errorf("Command not success: %s", r)
	}
	return nil
}

func (s *solver) Assert(t smt.Term) error {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"assert"},
		smt.TermToSexp(t)}})
	if err != nil {
		return fmt.Errorf("Command: %s", err)
	}
	if !isSuccess(r) {
		return fmt.Errorf("Command not success: %s", r)
	}
	return nil
}

func (s *solver) CheckSat() (smt.Satisfiable, error) {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"check-sat"}}})
	if err != nil {
		return smt.Unknown, fmt.Errorf("Command: %s", err)
	}
	switch {
	case smt.IsSymbol(r, "sat"):
		return smt.Sat, nil
	case smt.IsSymbol(r, "unsat"):
		return smt.Unsat, nil
	case smt.IsSymbol(r, "unknown"):
		return smt.Unknown, nil
	default:
		return smt.Unknown, fmt.Errorf("unexpected (check-sat): %s", r)
	}
}

func readModel(sexps []smt.Sexp) (map[string]smt.Term, error) {
	terms := make(map[string]smt.Term)

	for _, sexp := range sexps {
		switch app := sexp.(type) {
		case *smt.SList:
			if len(app.List) < 5 {
				log.Printf("readModel: unexpected short sexp")
				continue
			}
			if !smt.IsSymbol(app.List[0], "define-fun") {
				log.Printf("readModel: unexpected symbol: %s", app.List[0])
				continue
			}
			name, ok := app.List[1].(*smt.SSymbol)
			if !ok {
				log.Printf("readModel: var name not a symbol: %s", app.List[1])
				continue
			}

			if slist, ok := app.List[2].(*smt.SList); !ok || len(slist.List) != 0 {
				log.Printf("readModel: expected empty list: %s", app.List[2])
				continue
			}

			if !smt.IsSymbol(app.List[3], "Int") {
				log.Printf("readModel: expected Int sort, not %s", app.List[3])
			}
			t, err := smt.SexpToTerm(app.List[4])
			if err != nil {
				log.Printf("couldn't parse sexp: %s", err)
			}
			terms[name.Symbol] = t
		default:
			return nil, fmt.Errorf("expected model list, got %s", app)
		}
	}

	return terms, nil

}

func (s *solver) GetModel() (map[string]smt.Term, error) {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"get-model"}}})
	if err != nil {
		return nil, fmt.Errorf("Command: %s", err)
	}

	switch app := r.(type) {
	case *smt.SList:
		if len(app.List) == 0 {
			return nil, fmt.Errorf("expected model, got empty slist")
		}
		if !smt.IsSymbol(app.List[0], "model") {
			return nil, fmt.Errorf("expected model, got2 %s", app)
		}
		return readModel(app.List[1:])
	default:
		return nil, fmt.Errorf("expected model, got %s", r)
	}
}

func (s *solver) Push() {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"push"}}})
	if err != nil {
		panic(fmt.Sprintf("Command: %s", err))
	}
	if !isSuccess(r) {
		panic(fmt.Sprintf("Command not success: %s", r))
	}
}

func (s *solver) Pop() error {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"pop"}}})
	if err != nil {
		return fmt.Errorf("Command: %s", err)
	}
	if !isSuccess(r) {
		return fmt.Errorf("Command not success: %s", r)
	}
	return nil
}
