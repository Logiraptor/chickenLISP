package main

import (
	"strings"
	"testing"
)

func fact(n int) int {
	if n == 0 {
		return 1
	} else {
		return n * fact(n-1)
	}
}

var lispFact = `
(define factorial
  (lambda (n)
    (if (= n 0) 1
        (* n (factorial (- n 1))))))
(factorial 10)`

var use Atom

func init() {
	f, err := parser.Parse(strings.NewReader(lispFact))
	if err != nil {
		panic(err)
	}

	fun := transform(f.Children[0])
	eval(fun, global_env)
	use = transform(f.Children[1])
}

func TestFactsAgree(t *testing.T) {
	a := fact(10)
	b := eval(use, global_env).(Number)
	if a != int(b) {
		t.Errorf("Got: %v Exp: %d", b, a)
	}
}

func BenchmarkGoFact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fact(100)
	}
}

func BenchmarkCLISPFact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		eval(use, global_env)
	}
}
