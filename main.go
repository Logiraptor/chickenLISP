package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Logiraptor/chickenVM/peg"
)

var QUOTE Symbol = "quote"
var IF Symbol = "if"
var SET Symbol = "set"
var DEFINE Symbol = "define"
var LAMBDA Symbol = "lambda"
var BEGIN Symbol = "begin"

type Env struct {
	mapping map[Symbol]Atom
	outer   *Env
}

func NewEnv() *Env {
	return &Env{
		mapping: make(map[Symbol]Atom),
	}
}

func NewEnvFrom(vars, args List, outer *Env) *Env {
	e := NewEnv()
	e.outer = outer
	for i := range vars {
		e.mapping[vars[i].(Symbol)] = args[i]
	}
	return e
}

func (e *Env) find(s Symbol) map[Symbol]Atom {
	if _, ok := e.mapping[s]; ok {
		return e.mapping
	} else if e.outer != nil {
		return e.outer.find(s)
	} else {
		return nil
	}
}

type Atom interface{}

type List []Atom
type Number float64
type Symbol string

type FuncLit func(List) Atom

func (f FuncLit) Call(l List) Atom {
	return f(l)
}

type Procedure struct {
	params List
	body   Atom
	env    *Env
}

func (p *Procedure) Call(l List) Atom {
	return nil
}

type Func interface {
	Call(List) Atom
}

var source = `
(+ 1 2 (/ 3 4) (- 5 6) (* -1 4.45))
(define foo (lambda (a b) (+ a b)))
(foo 4 5)
(define factorial
  (lambda (n)
    (if (= n 0) 1
        (* n (factorial (- n 1))))))
(factorial 100)
(car (list 1 2 3))
(cdr (list 1 2 3))
`

var global_env = NewEnv()

func init() {
	global_env.mapping["+"] = FuncLit(func(l List) Atom {
		sum := Number(0)
		for _, x := range l {
			sum += x.(Number)
		}
		return sum
	})
	global_env.mapping["*"] = FuncLit(func(l List) Atom {
		product := Number(1)
		for _, x := range l {
			product *= x.(Number)
		}
		return product
	})
	global_env.mapping["/"] = FuncLit(func(l List) Atom {
		quotient := l[0].(Number)
		for _, x := range l[1:] {
			quotient /= x.(Number)
		}
		return quotient
	})
	global_env.mapping["-"] = FuncLit(func(l List) Atom {
		difference := l[0].(Number)
		for _, x := range l[1:] {
			difference -= x.(Number)
		}
		return difference
	})
	global_env.mapping["="] = FuncLit(func(l List) Atom {
		if l[0].(Number) == l[1].(Number) {
			return true
		} else {
			return nil
		}
	})
	global_env.mapping["car"] = FuncLit(func(l List) Atom {
		return l[0].(List)[0]
	})
	global_env.mapping["cdr"] = FuncLit(func(l List) Atom {
		return l[0].(List)[1:]
	})
	global_env.mapping["list"] = FuncLit(func(l List) Atom {
		return l
	})
	input, err := os.OpenFile("list.peg", os.O_RDONLY, 0660)
	if err != nil {
		fmt.Println(err)
		return
	}

	parser, err = peg.NewParser(input)
	if err != nil {
		fmt.Println(err)
		return
	}
}

var parser *peg.Language

func main() {
	tree, err := parser.Parse(strings.NewReader(source))
	if err != nil {
		fmt.Println(err)
		return
	}

	lists := make(chan List)
	go transformPRGM(tree, lists)
	for a := range lists {
		v := eval(a, global_env)
		if v != nil {
			fmt.Println(v)
		}
	}
}

func eval(a Atom, e *Env) Atom {
	for {
		switch a.(type) {
		case Number:
			return a
		case Symbol:
			s := a.(Symbol)
			return e.find(s)[s]
		case List:
			l := a.(List)
			switch l[0] {
			case QUOTE:
				return l[1]
			case IF:
				test, conseq, alt := l[1], l[2], l[3]
				if eval(test, e) != nil {
					a = conseq
				} else {
					a = alt
				}
			case SET:
				name, exp := l[1].(Symbol), l[2]
				e.find(name)[name] = eval(exp, e)
				return nil
			case DEFINE:
				name, exp := l[1].(Symbol), l[2]
				e.mapping[name] = eval(exp, e)
				return nil
			case LAMBDA:
				vars, exp := l[1].(List), l[2]
				return &Procedure{
					vars, exp, e,
				}
			case BEGIN:
				for _, exp := range l[1 : len(l)-1] {
					eval(exp, e)
				}
				a = l[len(l)-1]
			default:
				exps := make([]Atom, len(l))
				for i, exp := range l {
					exps[i] = eval(exp, e)
				}
				proc := exps[0].(Func)
				if p, ok := proc.(*Procedure); ok {
					a = p.body
					e = NewEnvFrom(p.params, exps[1:], p.env)
				} else {
					return proc.Call(exps[1:])
				}
			}
		default:
			panic(fmt.Sprintf("Cannot eval %v", a))
		}
		// return nil
	}
}

func transformPRGM(p *peg.ParseTree, out chan List) {
	for _, child := range p.Children {
		out <- transform(child).(List)
	}
	close(out)
}

func transform(p *peg.ParseTree) Atom {
	switch p.Type {
	case "number":
		val, err := strconv.ParseFloat(string(p.Data), 64)
		if err != nil {
			panic(err)
		}
		return Number(val)
	case "name", "op":
		return Symbol(string(p.Data))
	case "list":
		var resp = make(List, len(p.Children[1].Children))
		for i, child := range p.Children[1].Children {
			resp[i] = transform(child)
		}
		return resp
	default:
		fmt.Println(p)
		panic(fmt.Sprintf("Cannot transform %s", p.Type))
	}
}
