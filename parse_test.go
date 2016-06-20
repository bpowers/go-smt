package smt

import (
	"bytes"
	"encoding/json"
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
	for _, test := range sexpRTData {
		r := bytes.NewReader([]byte(test.input))
		p := NewParser(r)
		sexp, err := p.Read()
		if err != nil {
			t.Fatalf("Parse('%s'): %s", test.input, err)
		}
		if !reflect.DeepEqual(sexp, test.sexp) {
			buf, err := json.Marshal(sexp)
			if err != nil {
				t.Fatalf("couldn't encode %#v", sexp)
			}
			t.Fatalf("expected %s == %#v", string(buf), test.sexp)
		}

		sRT := sexp.String()
		sExpected := test.sexp.String()
		if sRT != sExpected {
			t.Fatalf("serialized versions differ: %s != %s", sRT, sExpected)
		}

		if _, err = p.Read(); err != ParserEOF {
			t.Fatalf("expected EOF")
		}
	}
}
