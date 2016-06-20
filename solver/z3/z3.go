package z3

import (
	"fmt"
	"go/token"
	"io"
	"os/exec"

	"github.com/bpowers/go-smt"
)

func isSuccess(sexp smt.Sexp) bool {
	return smt.IsSymbol(sexp, "success")
}

func NewSolverAt(path string) (smt.Solver, error) {
	cmd := exec.Command(path, "-in", "-smt2")
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
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
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
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
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

	buf := make([]byte, 8092)
	n, err = s.stdout.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("stdout.Read: %s", err)
	}

	fs := token.NewFileSet()
	f := fs.AddFile("<z3out>", -1, n)

	r := string(buf[:n])
	sexps, err := smt.Parse(f, r)
	if err != nil {
		return nil, fmt.Errorf("Parse('%s'): %s", r, err)
	}

	if len(sexps) != 1 {
		return nil, fmt.Errorf("Parse('%s'): expected 1 sexp not %d", r, len(sexps))
	}

	return sexps[0], nil
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

func (s *solver) GetModel() (map[string]smt.Term, error) {
	r, err := s.Command(&smt.SList{[]smt.Sexp{
		&smt.SSymbol{"get-model"}}})
	if err != nil {
		return nil, fmt.Errorf("Command: %s", err)
	}
	terms := make(map[string]smt.Term)

	fmt.Printf("get-model: %s\n", r)

	return terms, nil
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
