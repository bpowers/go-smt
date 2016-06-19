package smt

import (
	"encoding/json"
	"fmt"
	"go/token"
	"log"
	"reflect"
	"testing"
)

type sexpRTTest struct {
	input string
	sexp  Sexp
}

var sexpRTData = []sexpRTTest{
	{"3", &SInt{3}},
	{"()", &SList{[]Sexp{}}},
	{"(=)", &SList{[]Sexp{&SSymbol{"="}}}},
	{"(= a 3)", &SList{[]Sexp{&SSymbol{"="}, &SSymbol{"a"}, &SInt{3}}}},
	{"?", &SSymbol{"?"}},
	{":kw", &SKeyword{"kw"}},
	{"symbol", &SSymbol{"symbol"}},
	{`"string"`, &SString{"string"}},
	{`"!string!"`, &SString{"!string!"}},
}

func TestSexpRT(t *testing.T) {
	for i, test := range sexpRTData {
		fs := token.NewFileSet()
		f := fs.AddFile(fmt.Sprintf("<Test %d>", i), -1, len(test.input))
		sexps, err := Parse(f, test.input)
		if err != nil {
			t.Fatalf("Parse('%s'): %s", test.input, err)
		}
		if len(sexps) != 1 {
			t.Fatalf("len(%#v) != 1", sexps)
		}

		if !reflect.DeepEqual(sexps[0], test.sexp) {
			buf, err := json.Marshal(sexps[0])
			if err != nil {
				log.Printf("couldn't encode")
			}
			t.Fatalf("expected %s == %#v", string(buf), test.sexp)
		}
	}
}
