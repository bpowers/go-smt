package z3

import (
	"fmt"

	"github.com/bpowers/go-smt"
)

func NewSolverAt(path string) (smt.Solver, error) {
	// pipes for stdin + out

	// spawn $path -in -smt2

	// print success command

	return nil, fmt.Errorf("not implemented yet")
}
